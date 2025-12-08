// Package domain contains domain models and business logic
package domain

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// Bracket represents a tournament bracket aggregate
type Bracket struct {
	ID           uuid.UUID
	TournamentID uuid.UUID
	TotalRounds  int
	TotalMatches int
	ByesCount    int
	GeneratedAt  time.Time
	Rounds       []Round
}

// NewBracket creates a new bracket for a tournament
func NewBracket(tournamentID uuid.UUID, participants []uuid.UUID) (*Bracket, error) {
	if len(participants) < 2 {
		return nil, fmt.Errorf("tournament must have at least 2 participants, got %d", len(participants))
	}

	// Calculate number of rounds needed (log2 ceiling)
	participantCount := len(participants)
	totalRounds := int(math.Ceil(math.Log2(float64(participantCount))))

	// Calculate total slots needed (next power of 2)
	totalSlots := int(math.Pow(2, float64(totalRounds)))

	// Calculate byes needed
	byesCount := totalSlots - participantCount

	// Calculate total matches (sum of matches per round)
	totalMatches := 0
	for i := 0; i < totalRounds; i++ {
		matchesInRound := totalSlots / int(math.Pow(2, float64(i+1)))
		totalMatches += matchesInRound
	}

	return &Bracket{
		ID:           uuid.New(),
		TournamentID: tournamentID,
		TotalRounds:  totalRounds,
		TotalMatches: totalMatches,
		ByesCount:    byesCount,
		GeneratedAt:  time.Now(),
		Rounds:       []Round{},
	}, nil
}

// GenerateRounds creates all rounds for the bracket
func (b *Bracket) GenerateRounds(participants []uuid.UUID) ([]Round, error) {
	if b.TotalRounds == 0 {
		return nil, fmt.Errorf("bracket must have at least 1 round")
	}

	rounds := make([]Round, b.TotalRounds)
	totalSlots := int(math.Pow(2, float64(b.TotalRounds)))

	// Assign byes randomly to top seeds
	participantPool := make([]uuid.UUID, totalSlots)
	copy(participantPool, participants)

	// Fill remaining slots with nil (byes)
	for i := len(participants); i < totalSlots; i++ {
		participantPool[i] = uuid.Nil
	}

	// Generate rounds from first to last
	for roundNum := 1; roundNum <= b.TotalRounds; roundNum++ {
		matchesInRound := totalSlots / int(math.Pow(2, float64(roundNum)))
		roundName := getRoundName(roundNum, b.TotalRounds)

		round := Round{
			ID:           uuid.New(),
			BracketID:    b.ID,
			RoundNumber:  roundNum,
			RoundName:    roundName,
			MatchesCount: matchesInRound,
			Status:       RoundStatusPending,
			Matches:      []Match{},
		}

		rounds[roundNum-1] = round
	}

	return rounds, nil
}

// getRoundName generates human-readable round names
func getRoundName(roundNumber, totalRounds int) string {
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

// Validate checks bracket business invariants
func (b *Bracket) Validate() error {
	if b.TotalRounds < 1 {
		return fmt.Errorf("bracket must have at least 1 round")
	}

	if b.TotalMatches < 1 {
		return fmt.Errorf("bracket must have at least 1 match")
	}

	if b.ByesCount < 0 {
		return fmt.Errorf("byes count cannot be negative")
	}

	expectedSlots := int(math.Pow(2, float64(b.TotalRounds)))
	if b.ByesCount >= expectedSlots {
		return fmt.Errorf("byes count (%d) must be less than total slots (%d)", b.ByesCount, expectedSlots)
	}

	return nil
}






