package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RoundStatus represents the state of a round
type RoundStatus string

const (
	RoundStatusPending    RoundStatus = "pending"
	RoundStatusInProgress RoundStatus = "in_progress"
	RoundStatusCompleted  RoundStatus = "completed"
)

// Round represents a round within a bracket
type Round struct {
	ID           uuid.UUID
	BracketID    uuid.UUID
	RoundNumber  int
	RoundName    string
	MatchesCount int
	Status       RoundStatus
	StartedAt    *time.Time
	CompletedAt  *time.Time
	Matches      []Match
}

// NewRound creates a new round
func NewRound(bracketID uuid.UUID, roundNumber int, roundName string, matchesCount int) *Round {
	return &Round{
		ID:           uuid.New(),
		BracketID:    bracketID,
		RoundNumber:  roundNumber,
		RoundName:    roundName,
		MatchesCount: matchesCount,
		Status:       RoundStatusPending,
		Matches:      []Match{},
	}
}

// Start marks the round as in progress
func (r *Round) Start() error {
	if r.Status != RoundStatusPending {
		return fmt.Errorf("round must be pending to start, current status: %s", r.Status)
	}

	now := time.Now()
	r.Status = RoundStatusInProgress
	r.StartedAt = &now
	return nil
}

// Complete marks the round as completed
func (r *Round) Complete() error {
	if r.Status != RoundStatusInProgress {
		return fmt.Errorf("round must be in progress to complete, current status: %s", r.Status)
	}

	// Verify all matches are completed
	for _, match := range r.Matches {
		if match.Status != MatchStatusCompleted && match.Status != MatchStatusCancelled {
			return fmt.Errorf("cannot complete round: match %s is not completed (status: %s)", match.ID, match.Status)
		}
	}

	now := time.Now()
	r.Status = RoundStatusCompleted
	r.CompletedAt = &now
	return nil
}

// IsComplete returns true if all matches in the round are completed
func (r *Round) IsComplete() bool {
	if len(r.Matches) == 0 {
		return false
	}

	for _, match := range r.Matches {
		if match.Status != MatchStatusCompleted && match.Status != MatchStatusCancelled {
			return false
		}
	}

	return true
}

// Validate checks round business invariants
func (r *Round) Validate() error {
	if r.RoundNumber < 1 {
		return fmt.Errorf("round number must be positive")
	}

	if r.MatchesCount < 1 {
		return fmt.Errorf("round must have at least 1 match")
	}

	if r.RoundName == "" {
		return fmt.Errorf("round name cannot be empty")
	}

	return nil
}



