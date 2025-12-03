package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Tournament status constants
const (
	StatusDraft              = "draft"
	StatusScheduled          = "scheduled"
	StatusRegistrationOpen   = "registration_open"
	StatusRegistrationClosed = "registration_closed"
	StatusStarted            = "started"
	StatusCompleted          = "completed"
	StatusCancelled          = "cancelled"
)

// Tournament represents a tournament (domain entity).
type Tournament struct {
	ID                   uuid.UUID
	HostID               uuid.UUID
	Name                 string
	Description          *string
	Status               string
	ScheduledStartTimeAt *time.Time
	MaximumTeamCount     *int32
	MinimumTeamCount     int32
	TeamSize             int32
	AutoForceReady       bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// CanConfirmParticipation checks if the tournament allows participation confirmation.
// Business rules:
// - Tournament must be in "registration_open" or "scheduled" status
func (t *Tournament) CanConfirmParticipation() error {
	if t.Status != StatusRegistrationOpen && t.Status != StatusScheduled {
		return fmt.Errorf("tournament status must be 'registration_open' or 'scheduled', current status: %s", t.Status)
	}
	return nil
}

// CanJoin checks if users can join the tournament now.
// Business rules:
// - Tournament must have a scheduled start time
// - Current time must be within 15 minutes before start
func (t *Tournament) CanJoin() (time.Duration, error) {
	if t.ScheduledStartTimeAt == nil {
		return 0, fmt.Errorf("tournament has no scheduled start time")
	}

	timeUntilStart := time.Until(*t.ScheduledStartTimeAt)

	// Tournament already started
	if timeUntilStart < 0 {
		return timeUntilStart, fmt.Errorf("tournament has already started")
	}

	// Too early to join
	const joinWindow = 15 * time.Minute
	if timeUntilStart > joinWindow {
		return timeUntilStart, fmt.Errorf("too early to join, tournament starts in %.0f minutes", timeUntilStart.Minutes())
	}

	return timeUntilStart, nil
}

// ShouldCloseRegistration checks if registration should be closed.
// Business rules:
// - All places must be taken (participantCount >= maximumTeamCount)
// - Tournament must not be started or completed
func (t *Tournament) ShouldCloseRegistration(participantCount int64) bool {
	// No maximum count set, never close automatically
	if t.MaximumTeamCount == nil {
		return false
	}

	// Tournament already started or completed, don't change status
	if t.Status == StatusStarted || t.Status == StatusCompleted {
		return false
	}

	// All places taken
	return participantCount >= int64(*t.MaximumTeamCount)
}

// ShouldReopenRegistration checks if registration should be reopened.
// Business rules:
// - Tournament must be in "registration_closed" status
// - Must have free places (participantCount < maximumTeamCount)
// - Tournament must not be started or completed
func (t *Tournament) ShouldReopenRegistration(participantCount int64) bool {
	// Only reopen if currently closed
	if t.Status != StatusRegistrationClosed {
		return false
	}

	// Tournament already started or completed, don't reopen
	if t.Status == StatusStarted || t.Status == StatusCompleted {
		return false
	}

	// No maximum count set, reopen
	if t.MaximumTeamCount == nil {
		return true
	}

	// Have free places
	return participantCount < int64(*t.MaximumTeamCount)
}

// CloseRegistration closes the tournament registration.
func (t *Tournament) CloseRegistration() {
	if t.Status == StatusRegistrationOpen || t.Status == StatusScheduled {
		t.Status = StatusRegistrationClosed
		t.UpdatedAt = time.Now()
	}
}

// ReopenRegistration reopens the tournament registration.
func (t *Tournament) ReopenRegistration() {
	if t.Status == StatusRegistrationClosed {
		t.Status = StatusRegistrationOpen
		t.UpdatedAt = time.Now()
	}
}
