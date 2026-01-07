package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Participant status value objects
const (
	ParticipantStatusPending      = "pending"
	ParticipantStatusRegistered   = "registered"
	ParticipantStatusConfirmed    = "confirmed"
	ParticipantStatusCheckedIn    = "checked_in"
	ParticipantStatusWithdrawn    = "withdrawn"
	ParticipantStatusDisqualified = "disqualified"

	// Short aliases for convenience
	StatusPending      = ParticipantStatusPending
	StatusRegistered   = ParticipantStatusRegistered
	StatusConfirmed    = ParticipantStatusConfirmed
	StatusCheckedIn    = ParticipantStatusCheckedIn
	StatusWithdrawn    = ParticipantStatusWithdrawn
	StatusDisqualified = ParticipantStatusDisqualified
)

// TournamentParticipant represents a participant in a tournament (domain entity).
type TournamentParticipant struct {
	ID           uuid.UUID
	TournamentID uuid.UUID
	UserID       uuid.UUID
	TeamID       *uuid.UUID
	Status       string
	DisplayName  string  // Cached from user profile (nickname)
	AvatarURL    *string // Cached from user profile
	IsReady      bool
	RegisteredAt *time.Time
	ConfirmedAt  *time.Time
	CheckedInAt  *time.Time
	UpdatedAt    time.Time
}

// RegisterParticipation registers the participant for the tournament.
// Business rules:
// - Status must be "pending"
// - Sets status to "registered"
func (p *TournamentParticipant) RegisterParticipation() error {
	if p.Status != ParticipantStatusPending {
		return fmt.Errorf("cannot register with status: %s", p.Status)
	}

	p.Status = ParticipantStatusRegistered
	now := time.Now()
	p.RegisteredAt = &now
	p.UpdatedAt = now

	return nil
}

// ConfirmParticipation confirms the participant's participation in the tournament.
// Business rules:
// - Status must be "pending" or "registered"
// - Cannot confirm if already confirmed or checked in
func (p *TournamentParticipant) ConfirmParticipation() error {
	if p.Status == ParticipantStatusConfirmed || p.Status == ParticipantStatusCheckedIn {
		return fmt.Errorf("participation already confirmed")
	}

	if p.Status != ParticipantStatusPending && p.Status != ParticipantStatusRegistered {
		return fmt.Errorf("cannot confirm participation with status: %s", p.Status)
	}

	p.Status = ParticipantStatusConfirmed
	now := time.Now()
	p.ConfirmedAt = &now
	p.UpdatedAt = now

	return nil
}

// CheckIn checks in the participant for the tournament.
// Business rules:
// - Status must be "confirmed"
func (p *TournamentParticipant) CheckIn() error {
	if p.Status != ParticipantStatusConfirmed {
		return fmt.Errorf("cannot check in with status: %s", p.Status)
	}

	p.Status = ParticipantStatusCheckedIn
	now := time.Now()
	p.CheckedInAt = &now
	p.UpdatedAt = now

	return nil
}

// Withdraw withdraws the participant from the tournament.
func (p *TournamentParticipant) Withdraw() error {
	if p.Status == ParticipantStatusWithdrawn {
		return fmt.Errorf("already withdrawn")
	}

	if p.Status == ParticipantStatusDisqualified {
		return fmt.Errorf("cannot withdraw when disqualified")
	}

	p.Status = ParticipantStatusWithdrawn
	p.UpdatedAt = time.Now()

	return nil
}

// MarkReady marks the participant as ready.
func (p *TournamentParticipant) MarkReady() {
	p.IsReady = true
	p.UpdatedAt = time.Now()
}

// CanJoin checks if the participant can join the tournament.
// Business rules:
// - Status must be "registered", "confirmed", or "checked_in"
// - Time until start must be within join window
func (p *TournamentParticipant) CanJoin(timeUntilStart time.Duration) bool {
	// Must have registered or higher status
	if p.Status != ParticipantStatusRegistered &&
		p.Status != ParticipantStatusConfirmed &&
		p.Status != ParticipantStatusCheckedIn {
		return false
	}

	// Join window: 15 minutes before start, up to 30 seconds after
	const joinWindow = 15 * time.Minute

	// Can join if within the window (including 30s after start)
	return timeUntilStart <= joinWindow && timeUntilStart > -30*time.Second
}

// IsConfirmed returns true if the participant has confirmed participation.
func (p *TournamentParticipant) IsConfirmed() bool {
	return p.Status == ParticipantStatusConfirmed || p.Status == ParticipantStatusCheckedIn
}

// IsActive returns true if the participant is still active in the tournament.
func (p *TournamentParticipant) IsActive() bool {
	return p.Status != ParticipantStatusWithdrawn && p.Status != ParticipantStatusDisqualified
}
