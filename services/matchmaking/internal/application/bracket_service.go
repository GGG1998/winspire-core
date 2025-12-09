// Package application contains application services and business logic orchestration
package application

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
	"github.com/winspire-core/services/matchmaking/internal/repository"
)

// BracketService handles bracket generation and management
type BracketService struct {
	bracketRepo repository.BracketRepository
	roundRepo   repository.RoundRepository
	matchRepo   repository.MatchRepository
	publisher   *pubsub.EventPublisher
	metrics     *observability.MetricsEmitter
	logger      *observability.Logger
}

// NewBracketService creates a new bracket service
func NewBracketService(
	bracketRepo repository.BracketRepository,
	roundRepo repository.RoundRepository,
	matchRepo repository.MatchRepository,
	publisher *pubsub.EventPublisher,
	metrics *observability.MetricsEmitter,
	logger *observability.Logger,
) *BracketService {
	return &BracketService{
		bracketRepo: bracketRepo,
		roundRepo:   roundRepo,
		matchRepo:   matchRepo,
		publisher:   publisher,
		metrics:     metrics,
		logger:      logger,
	}
}

// GenerateBracket generates a complete single-elimination bracket
func (s *BracketService) GenerateBracket(ctx context.Context, tournamentID uuid.UUID, participants []uuid.UUID, gameSnapshot *domain.GameSnapshot) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		s.metrics.EmitBracketGenerationDuration(duration)
		s.logger.Info("Bracket generation completed", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"duration_ms":   duration.Milliseconds(),
			"participants":  len(participants),
		})
	}()

	s.logger.Info("Starting bracket generation", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"participants":  len(participants),
	})

	// Create bracket aggregate
	bracket, err := domain.NewBracket(tournamentID, participants)
	if err != nil {
		return fmt.Errorf("create bracket: %w", err)
	}

	// Attach game snapshot if provided
	bracket.GameSnapshot = gameSnapshot

	// Generate rounds
	rounds, err := s.generateRounds(bracket, participants)
	if err != nil {
		return fmt.Errorf("generate rounds: %w", err)
	}

	// Generate matches for all rounds
	matches, err := s.generateMatches(bracket, rounds, participants)
	if err != nil {
		return fmt.Errorf("generate matches: %w", err)
	}

	// Persist bracket, rounds, and matches in transaction
	err = s.bracketRepo.Create(ctx, bracket, rounds, matches)
	if err != nil {
		return fmt.Errorf("persist bracket: %w", err)
	}

	// Publish BracketGenerated event
	event := domain.NewBracketGenerated(
		bracket.ID,
		tournamentID,
		bracket.TotalRounds,
		bracket.TotalMatches,
		bracket.ByesCount,
		participants,
		map[string]string{
			"correlation_id": s.getCorrelationID(ctx),
		},
	)

	if err := s.publisher.Publish(ctx, event); err != nil {
		log.Printf("WARN: Failed to publish BracketGenerated event: %v", err)
	}

	// Publish BracketGenerationCompleted event (for tournament start saga coordination)
	completedEvent := domain.NewBracketGenerationCompleted(
		tournamentID,
		bracket.ID,
		bracket.TotalRounds,
		bracket.TotalMatches,
		map[string]string{
			"correlation_id": s.getCorrelationID(ctx),
		},
	)

	if err := s.publisher.Publish(ctx, completedEvent); err != nil {
		log.Printf("WARN: Failed to publish BracketGenerationCompleted event: %v", err)
	}

	// Publish RoundCreated events
	for _, round := range rounds {
		roundEvent := domain.NewRoundCreated(
			round.ID,
			bracket.ID,
			tournamentID,
			round.RoundNumber,
			round.RoundName,
			round.MatchesCount,
			map[string]string{
				"correlation_id": s.getCorrelationID(ctx),
			},
		)
		if err := s.publisher.Publish(ctx, roundEvent); err != nil {
			log.Printf("WARN: Failed to publish RoundCreated event: %v", err)
		}
	}

	// Publish MatchCreated events
	for _, match := range matches {
		matchEvent := domain.NewMatchCreated(
			match.ID,
			tournamentID,
			match.RoundID,
			s.getRoundNumber(rounds, match.RoundID),
			match.MatchNumber,
			match.Participant1ID,
			match.Participant2ID,
			match.NextMatchID,
			match.IsBye(),
			map[string]string{
				"correlation_id": s.getCorrelationID(ctx),
			},
		)
		if err := s.publisher.Publish(ctx, matchEvent); err != nil {
			log.Printf("WARN: Failed to publish MatchCreated event: %v", err)
		}
	}

	return nil
}

// generateRounds creates all rounds for the bracket
func (s *BracketService) generateRounds(bracket *domain.Bracket, participants []uuid.UUID) ([]domain.Round, error) {
	rounds := make([]domain.Round, bracket.TotalRounds)

	totalSlots := int(math.Pow(2, float64(bracket.TotalRounds)))

	for roundNum := 1; roundNum <= bracket.TotalRounds; roundNum++ {
		matchesInRound := totalSlots / int(math.Pow(2, float64(roundNum)))
		roundName := s.getRoundName(roundNum, bracket.TotalRounds)

		round := domain.Round{
			ID:           uuid.New(),
			BracketID:    bracket.ID,
			RoundNumber:  roundNum,
			RoundName:    roundName,
			MatchesCount: matchesInRound,
			Status:       domain.RoundStatusPending,
		}

		rounds[roundNum-1] = round
	}

	return rounds, nil
}

// generateMatches creates all matches for all rounds with proper linkage
func (s *BracketService) generateMatches(bracket *domain.Bracket, rounds []domain.Round, participants []uuid.UUID) ([]domain.Match, error) {
	// Shuffle participants for bye assignment
	shuffled := make([]uuid.UUID, len(participants))
	copy(shuffled, participants)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Pad with nil for byes
	totalSlots := int(math.Pow(2, float64(bracket.TotalRounds)))
	paddedParticipants := make([]*uuid.UUID, totalSlots)
	for i := 0; i < len(shuffled); i++ {
		paddedParticipants[i] = &shuffled[i]
	}

	var allMatches []domain.Match
	matchIDMap := make(map[string]uuid.UUID) // Map "roundNum-matchNum" to match ID

	// Generate matches round by round (reverse order for proper next_match_id linking)
	for roundIdx := len(rounds) - 1; roundIdx >= 0; roundIdx-- {
		round := rounds[roundIdx]
		roundMatches := []domain.Match{}

		for matchNum := 1; matchNum <= round.MatchesCount; matchNum++ {
			matchID := uuid.New()
			matchIDMap[fmt.Sprintf("%d-%d", round.RoundNumber, matchNum)] = matchID

			// Determine next match ID (for advancement)
			var nextMatchID *uuid.UUID
			if round.RoundNumber < bracket.TotalRounds {
				nextRound := round.RoundNumber + 1
				nextMatchNum := (matchNum + 1) / 2 // Integer division for pairing
				nextMatchKey := fmt.Sprintf("%d-%d", nextRound, nextMatchNum)
				if nextID, exists := matchIDMap[nextMatchKey]; exists {
					nextMatchID = &nextID
				}
			}

			// For first round, assign participants
			var participant1ID uuid.UUID
			var participant2ID *uuid.UUID

			if round.RoundNumber == 1 {
				// First round - assign from participant pool
				idx1 := (matchNum - 1) * 2
				idx2 := idx1 + 1

				if paddedParticipants[idx1] != nil {
					participant1ID = *paddedParticipants[idx1]
				}
				if idx2 < len(paddedParticipants) && paddedParticipants[idx2] != nil {
					participant2ID = paddedParticipants[idx2]
				}
			} else {
				// Later rounds - participants come from previous round winners (will be assigned later)
				participant1ID = uuid.Nil // Placeholder
			}

			match := domain.Match{
				ID:                matchID,
				RoundID:           round.ID,
				MatchNumber:       matchNum,
				NextMatchID:       nextMatchID,
				Participant1ID:    participant1ID,
				Participant2ID:    participant2ID,
				Status:            domain.MatchStatusPending,
				Participant1Ready: false,
				Participant2Ready: false,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}

			roundMatches = append(roundMatches, match)
		}

		allMatches = append(roundMatches, allMatches...)
	}

	return allMatches, nil
}

// getRoundName generates human-readable round names
func (s *BracketService) getRoundName(roundNumber, totalRounds int) string {
	remainingRounds := totalRounds - roundNumber

	switch remainingRounds {
	case 0:
		return "Final"
	case 1:
		return "Semi-finals"
	case 2:
		return "Quarter-finals"
	default:
		matchesInRound := int(math.Pow(2, float64(totalRounds-roundNumber)))
		return fmt.Sprintf("Round of %d", matchesInRound*2)
	}
}

// getRoundNumber finds the round number for a given round ID
func (s *BracketService) getRoundNumber(rounds []domain.Round, roundID uuid.UUID) int {
	for _, round := range rounds {
		if round.ID == roundID {
			return round.RoundNumber
		}
	}
	return 0
}

// getCorrelationID extracts correlation ID from context
func (s *BracketService) getCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		return correlationID
	}
	return uuid.New().String()
}

// GetMatchesByRound returns all matches for a specific round
func (s *BracketService) GetMatchesByRound(ctx context.Context, tournamentID uuid.UUID, roundNumber int) ([]domain.Match, error) {
	return s.matchRepo.GetByTournamentAndRound(ctx, tournamentID, roundNumber)
}
