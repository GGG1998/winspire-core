// Package application provides application services for matchmaking
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/repository"
)

// RoundTransitionData represents the persisted state for a round transition
type RoundTransitionData struct {
	TournamentID    uuid.UUID   `json:"tournament_id"`
	CompletedRound  int         `json:"completed_round"`
	NextRound       int         `json:"next_round"`
	ExpectedWinners []uuid.UUID `json:"expected_winners"`
	JoinedWinners   []uuid.UUID `json:"joined_winners"`
	GraceStarted    bool        `json:"grace_started"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// RoundTransitionStoreInterface defines the interface for round transition state management
type RoundTransitionStoreInterface interface {
	Get(ctx context.Context, tournamentID uuid.UUID) (*RoundTransitionData, error)
	Save(ctx context.Context, state *RoundTransitionData) error
	Delete(ctx context.Context, tournamentID uuid.UUID) error
	AddExpectedWinner(ctx context.Context, tournamentID uuid.UUID, completedRound int, winnerID uuid.UUID) error
	AddJoinedWinner(ctx context.Context, tournamentID uuid.UUID, winnerID uuid.UUID) (bool, error)
	MarkGraceStarted(ctx context.Context, tournamentID uuid.UUID) (bool, error)
	GetOrProject(ctx context.Context, tournamentID uuid.UUID) (*RoundTransitionData, error)
}

// RoundTransitionStore manages round transition state in Redis with DB projection
type RoundTransitionStore struct {
	redis       *redis.Client
	matchRepo   repository.MatchRepository
	roundRepo   repository.RoundRepository
	bracketRepo repository.BracketRepository
	logger      *observability.Logger
}

// Redis key patterns
const (
	keyRoundTransition = "matchmaking:round_transition:%s" // JSON data by tournament_id
	roundTransitionTTL = 24 * time.Hour                    // Expire after 24 hours
)

// NewRoundTransitionStore creates a new RoundTransitionStore
func NewRoundTransitionStore(
	redisClient *redis.Client,
	matchRepo repository.MatchRepository,
	roundRepo repository.RoundRepository,
	bracketRepo repository.BracketRepository,
	logger *observability.Logger,
) *RoundTransitionStore {
	return &RoundTransitionStore{
		redis:       redisClient,
		matchRepo:   matchRepo,
		roundRepo:   roundRepo,
		bracketRepo: bracketRepo,
		logger:      logger,
	}
}

// Get retrieves the round transition state for a tournament
func (s *RoundTransitionStore) Get(ctx context.Context, tournamentID uuid.UUID) (*RoundTransitionData, error) {
	key := fmt.Sprintf(keyRoundTransition, tournamentID.String())

	data, err := s.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("get round transition from redis: %w", err)
	}

	var state RoundTransitionData
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal round transition: %w", err)
	}

	return &state, nil
}

// Save persists the round transition state to Redis
func (s *RoundTransitionStore) Save(ctx context.Context, state *RoundTransitionData) error {
	state.UpdatedAt = time.Now()

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal round transition: %w", err)
	}

	key := fmt.Sprintf(keyRoundTransition, state.TournamentID.String())
	if err := s.redis.Set(ctx, key, data, roundTransitionTTL).Err(); err != nil {
		return fmt.Errorf("save round transition to redis: %w", err)
	}

	s.logger.Info("[RoundTransitionStore] Saved round transition", map[string]interface{}{
		"tournament_id":    state.TournamentID.String(),
		"completed_round":  state.CompletedRound,
		"expected_winners": len(state.ExpectedWinners),
		"joined_winners":   len(state.JoinedWinners),
	})

	return nil
}

// Delete removes the round transition state from Redis
func (s *RoundTransitionStore) Delete(ctx context.Context, tournamentID uuid.UUID) error {
	key := fmt.Sprintf(keyRoundTransition, tournamentID.String())
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete round transition from redis: %w", err)
	}
	return nil
}

// AddExpectedWinner adds a winner to the expected winners list
func (s *RoundTransitionStore) AddExpectedWinner(ctx context.Context, tournamentID uuid.UUID, completedRound int, winnerID uuid.UUID) error {
	state, err := s.Get(ctx, tournamentID)
	if err != nil {
		return err
	}

	if state == nil {
		// Create new state
		state = &RoundTransitionData{
			TournamentID:    tournamentID,
			CompletedRound:  completedRound,
			NextRound:       completedRound + 1,
			ExpectedWinners: []uuid.UUID{winnerID},
			JoinedWinners:   []uuid.UUID{},
			GraceStarted:    false,
			CreatedAt:       time.Now(),
		}
	} else {
		// Add winner if not already present
		if !containsUUID(state.ExpectedWinners, winnerID) {
			state.ExpectedWinners = append(state.ExpectedWinners, winnerID)
		}
	}

	return s.Save(ctx, state)
}

// AddJoinedWinner marks a winner as having joined the pre-lobby
func (s *RoundTransitionStore) AddJoinedWinner(ctx context.Context, tournamentID uuid.UUID, winnerID uuid.UUID) (allJoined bool, err error) {
	state, err := s.Get(ctx, tournamentID)
	if err != nil {
		return false, err
	}

	if state == nil {
		return false, nil // No transition in progress
	}

	// Check if this is an expected winner
	fmt.Println("[RoundTransitioStore] Debug state.ExpectedWinners", state.ExpectedWinners)
	if !containsUUID(state.ExpectedWinners, winnerID) {
		return false, nil // Not an expected winner
	}

	// Add to joined if not already present
	fmt.Println("[RoundTransitionStore] Debug state.JoinedWinners", state.JoinedWinners)
	if !containsUUID(state.JoinedWinners, winnerID) {
		state.JoinedWinners = append(state.JoinedWinners, winnerID)
	}

	if err := s.Save(ctx, state); err != nil {
		return false, err
	}

	// Check if all expected winners have joined
	allJoined = len(state.JoinedWinners) >= len(state.ExpectedWinners)
	return allJoined, nil
}

// MarkGraceStarted marks the grace period as started (prevents double-start)
func (s *RoundTransitionStore) MarkGraceStarted(ctx context.Context, tournamentID uuid.UUID) (bool, error) {
	state, err := s.Get(ctx, tournamentID)
	if err != nil {
		return false, err
	}

	if state == nil {
		return false, nil
	}

	if state.GraceStarted {
		return false, nil // Already started
	}

	state.GraceStarted = true
	if err := s.Save(ctx, state); err != nil {
		return false, err
	}

	return true, nil // Successfully marked as started
}

// ProjectFromDatabase rebuilds expected winners from completed matches in the database
// This is used on service startup to recover state
func (s *RoundTransitionStore) ProjectFromDatabase(ctx context.Context, tournamentID uuid.UUID) (*RoundTransitionData, error) {
	s.logger.Info("[RoundTransitionStore] Projecting round transition from database", map[string]interface{}{
		"tournament_id": tournamentID.String(),
	})

	// Get bracket for tournament
	bracket, err := s.bracketRepo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("get bracket: %w", err)
	}
	if bracket == nil {
		return nil, nil // No bracket exists
	}

	// Get all rounds for this bracket
	rounds, err := s.roundRepo.GetByBracketID(ctx, bracket.ID)
	if err != nil {
		return nil, fmt.Errorf("get rounds: %w", err)
	}

	if len(rounds) == 0 {
		return nil, nil
	}

	// Find the latest completed round (all matches completed)
	var completedRound *domain.Round
	for i := range rounds {
		round := &rounds[i]
		matches, err := s.matchRepo.GetByRoundID(ctx, round.ID)
		if err != nil {
			continue
		}

		allCompleted := true
		for _, match := range matches {
			if match.Status != domain.MatchStatusCompleted {
				allCompleted = false
				break
			}
		}

		if allCompleted && (completedRound == nil || round.RoundNumber > completedRound.RoundNumber) {
			completedRound = round
		}
	}

	if completedRound == nil {
		return nil, nil // No completed round
	}

	// Check if there's a next round (not final round)
	var hasNextRound bool
	for _, round := range rounds {
		if round.RoundNumber == completedRound.RoundNumber+1 {
			hasNextRound = true
			break
		}
	}

	if !hasNextRound {
		return nil, nil // Final round completed, no transition needed
	}

	// Get winners from completed round
	matches, err := s.matchRepo.GetByRoundID(ctx, completedRound.ID)
	if err != nil {
		return nil, fmt.Errorf("get matches for completed round: %w", err)
	}

	expectedWinners := make([]uuid.UUID, 0)
	for _, match := range matches {
		if match.WinnerID != nil && *match.WinnerID != uuid.Nil {
			expectedWinners = append(expectedWinners, *match.WinnerID)
		}
	}

	if len(expectedWinners) == 0 {
		return nil, nil
	}

	state := &RoundTransitionData{
		TournamentID:    tournamentID,
		CompletedRound:  completedRound.RoundNumber,
		NextRound:       completedRound.RoundNumber + 1,
		ExpectedWinners: expectedWinners,
		JoinedWinners:   []uuid.UUID{}, // Will be populated when winners join pre-lobby
		GraceStarted:    false,
		CreatedAt:       time.Now(),
	}

	s.logger.Info("[RoundTransitionStore] Projected round transition from database", map[string]interface{}{
		"tournament_id":    tournamentID.String(),
		"completed_round":  state.CompletedRound,
		"expected_winners": len(state.ExpectedWinners),
	})

	return state, nil
}

// GetOrProject gets from Redis or projects from database
func (s *RoundTransitionStore) GetOrProject(ctx context.Context, tournamentID uuid.UUID) (*RoundTransitionData, error) {
	// Try Redis first
	state, err := s.Get(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	if state != nil {
		return state, nil
	}

	// Project from database
	state, err = s.ProjectFromDatabase(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	if state != nil {
		// Save projected state to Redis
		if err := s.Save(ctx, state); err != nil {
			s.logger.Warn("[RoundTransitionStore] Failed to save projected state to Redis", map[string]interface{}{
				"tournament_id": tournamentID.String(),
				"error":         err.Error(),
			})
		}
	}

	return state, nil
}

// Helper function to check if a UUID slice contains a specific UUID
func containsUUID(slice []uuid.UUID, id uuid.UUID) bool {
	for _, item := range slice {
		if item == id {
			return true
		}
	}
	return false
}

// InMemoryRoundTransitionStore implements RoundTransitionStore for testing
type InMemoryRoundTransitionStore struct {
	states map[uuid.UUID]*RoundTransitionData
	logger *observability.Logger
}

// NewInMemoryRoundTransitionStore creates a new in-memory transition store for testing
func NewInMemoryRoundTransitionStore(logger *observability.Logger) *InMemoryRoundTransitionStore {
	return &InMemoryRoundTransitionStore{
		states: make(map[uuid.UUID]*RoundTransitionData),
		logger: logger,
	}
}

func (s *InMemoryRoundTransitionStore) Get(_ context.Context, tournamentID uuid.UUID) (*RoundTransitionData, error) {
	state, exists := s.states[tournamentID]
	if !exists {
		return nil, nil
	}
	return state, nil
}

func (s *InMemoryRoundTransitionStore) Save(_ context.Context, state *RoundTransitionData) error {
	state.UpdatedAt = time.Now()
	s.states[state.TournamentID] = state
	return nil
}

func (s *InMemoryRoundTransitionStore) Delete(_ context.Context, tournamentID uuid.UUID) error {
	delete(s.states, tournamentID)
	return nil
}

func (s *InMemoryRoundTransitionStore) AddExpectedWinner(ctx context.Context, tournamentID uuid.UUID, completedRound int, winnerID uuid.UUID) error {
	state, _ := s.Get(ctx, tournamentID)

	if state == nil {
		state = &RoundTransitionData{
			TournamentID:    tournamentID,
			CompletedRound:  completedRound,
			NextRound:       completedRound + 1,
			ExpectedWinners: []uuid.UUID{winnerID},
			JoinedWinners:   []uuid.UUID{},
			GraceStarted:    false,
			CreatedAt:       time.Now(),
		}
	} else {
		if !containsUUID(state.ExpectedWinners, winnerID) {
			state.ExpectedWinners = append(state.ExpectedWinners, winnerID)
		}
	}

	return s.Save(ctx, state)
}

func (s *InMemoryRoundTransitionStore) AddJoinedWinner(ctx context.Context, tournamentID uuid.UUID, winnerID uuid.UUID) (bool, error) {
	state, _ := s.Get(ctx, tournamentID)
	if state == nil {
		return false, nil
	}

	if !containsUUID(state.ExpectedWinners, winnerID) {
		return false, nil
	}

	if !containsUUID(state.JoinedWinners, winnerID) {
		state.JoinedWinners = append(state.JoinedWinners, winnerID)
	}

	if err := s.Save(ctx, state); err != nil {
		return false, err
	}

	return len(state.JoinedWinners) >= len(state.ExpectedWinners), nil
}

func (s *InMemoryRoundTransitionStore) MarkGraceStarted(ctx context.Context, tournamentID uuid.UUID) (bool, error) {
	state, _ := s.Get(ctx, tournamentID)
	if state == nil {
		return false, nil
	}

	if state.GraceStarted {
		return false, nil
	}

	state.GraceStarted = true
	if err := s.Save(ctx, state); err != nil {
		return false, err
	}

	return true, nil
}

func (s *InMemoryRoundTransitionStore) GetOrProject(ctx context.Context, tournamentID uuid.UUID) (*RoundTransitionData, error) {
	return s.Get(ctx, tournamentID)
}
