package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Participant status value objects
const (
	StatusPending    = "pending"
	StatusRegistered = "registered"
	StatusConfirmed  = "confirmed"
	StatusCheckedIn  = "checked_in"
)

// TournamentParticipant represents a participant in a tournament (domain entity).
type TournamentParticipant struct {
	ID           uuid.UUID
	TournamentID uuid.UUID
	UserID       uuid.UUID
	Status       string
	DisplayName  string  // Cached from user profile (nickname)
	AvatarURL    *string // Cached from user profile
	RegisteredAt *time.Time
	ConfirmedAt  *time.Time
	CheckedInAt  *time.Time
}

// RegisterParticipation registers the participant for the tournament.
// Business rules:
// - Status must be "pending"
// - Sets status to "registered"
func (p *TournamentParticipant) RegisterParticipation() error {
	if p.Status != StatusPending {
		return fmt.Errorf("cannot register with status: %s", p.Status)
	}

	p.Status = StatusRegistered
	now := time.Now()
	p.RegisteredAt = &now

	return nil
}

// ConfirmParticipation confirms the participant's participation in the tournament.
// Business rules:
// - Status must be "pending" or "registered"
// - Cannot confirm if already confirmed or checked in
func (p *TournamentParticipant) ConfirmParticipation() error {
	if p.Status == StatusConfirmed || p.Status == StatusCheckedIn {
		return fmt.Errorf("participation already confirmed")
	}

	if p.Status != StatusPending && p.Status != StatusRegistered {
		return fmt.Errorf("cannot confirm participation with status: %s", p.Status)
	}

	p.Status = StatusConfirmed
	now := time.Now()
	p.ConfirmedAt = &now

	return nil
}

// CanJoin checks if the participant can join the tournament.
// Business rules:
// - Status must be "registered", "confirmed", or "checked_in"
// - Time until start must be ≤ 15 minutes
// - Can join up to 30 seconds after start time
func (p *TournamentParticipant) CanJoin(timeUntilStart time.Duration) bool {
	// Must have registered or higher status
	if p.Status != StatusRegistered && p.Status != StatusConfirmed && p.Status != StatusCheckedIn {
		return false
	}

	// Join window: 15 minutes before start, up to 30 seconds after
	const joinWindow = 15 * time.Minute

	// Can join if within the window (including 30s after start)
	return timeUntilStart <= joinWindow && timeUntilStart > -30*time.Second
}

// IsConfirmed returns true if the participant has confirmed participation.
func (p *TournamentParticipant) IsConfirmed() bool {
	return p.Status == StatusConfirmed || p.Status == StatusCheckedIn
}
