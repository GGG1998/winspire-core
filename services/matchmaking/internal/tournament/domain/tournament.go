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
	StatusStarting           = "starting" // Transient state during tournament start saga
	StatusStarted            = "started"
	StatusCompleted          = "completed"
	StatusCancelled          = "cancelled"
)

// GameSnapshot represents denormalized game data stored with tournament.
type GameSnapshot struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	LogoURL     string    `json:"logoUrl,omitempty"`
	Description string    `json:"description,omitempty"`
	StoragePath string    `json:"storagePath,omitempty"`
}

// ReadyWindow represents the check-in window configuration.
type ReadyWindow struct {
	StartsAt *time.Time `json:"startsAt,omitempty"`
	EndsAt   *time.Time `json:"endsAt,omitempty"`
}

// Prize represents the prize configuration.
type Prize struct {
	Type        string  `json:"type,omitempty"`        // custom, cash, points
	Description string  `json:"description,omitempty"`
	Value       float64 `json:"value,omitempty"`
	Currency    string  `json:"currency,omitempty"`
}

// Tournament represents a tournament (domain entity).
type Tournament struct {
	ID                   uuid.UUID
	HostID               uuid.UUID
	Name                 string
	Description          *string
	ExternalID           *string
	Status               string
	ScheduledStartTimeAt *time.Time
	RegistrationWindowOpenAt *time.Time
	ActualStartTimeAt    *time.Time
	CompletedAt          *time.Time
	CancelledAt          *time.Time
	MaximumTeamCount     *int32
	MinimumTeamCount     int32
	TeamSize             int32
	AutoForceReady       bool
	GameID               *uuid.UUID
	SpaceID              *uuid.UUID
	TemplateID           *uuid.UUID
	ReadyWindow          *ReadyWindow
	Prize                *Prize
	GameSnapshot         *GameSnapshot
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
// - Users can join up to 30 seconds after start time
func (t *Tournament) CanJoin() (time.Duration, error) {
	if t.ScheduledStartTimeAt == nil {
		return 0, fmt.Errorf("tournament has no scheduled start time")
	}

	timeUntilStart := time.Until(*t.ScheduledStartTimeAt)

	// Tournament already started (more than 30 seconds ago)
	if timeUntilStart < -30*time.Second {
		return timeUntilStart, fmt.Errorf("tournament has already started")
	}

	// Too early to join
	const joinWindow = 15 * time.Minute
	if timeUntilStart > joinWindow {
		return timeUntilStart, fmt.Errorf("too early to join, tournament starts in %.0f minutes", timeUntilStart.Minutes())
	}

	return timeUntilStart, nil
}

// CanStart checks if the tournament can be started.
// Business rules:
// - Tournament must be in "registration_open" or "registration_closed" status
func (t *Tournament) CanStart() error {
	if t.Status != StatusRegistrationOpen && t.Status != StatusRegistrationClosed {
		return fmt.Errorf("tournament cannot be started from status: %s", t.Status)
	}
	return nil
}

// RequestStart transitions the tournament to "starting" state.
// This is the first step of the start saga.
func (t *Tournament) RequestStart() error {
	if err := t.CanStart(); err != nil {
		return err
	}
	t.Status = StatusStarting
	t.UpdatedAt = time.Now()
	return nil
}

// ConfirmStart transitions the tournament to "started" state.
// This is called when the bracket generation is complete.
func (t *Tournament) ConfirmStart() {
	t.Status = StatusStarted
	now := time.Now()
	t.ActualStartTimeAt = &now
	t.UpdatedAt = now
}

// RollbackStart returns the tournament to registration_open if start fails.
func (t *Tournament) RollbackStart() {
	if t.Status == StatusStarting {
		t.Status = StatusRegistrationOpen
		t.UpdatedAt = time.Now()
	}
}

// Complete marks the tournament as completed.
func (t *Tournament) Complete() {
	t.Status = StatusCompleted
	now := time.Now()
	t.CompletedAt = &now
	t.UpdatedAt = now
}

// Cancel marks the tournament as cancelled.
func (t *Tournament) Cancel() {
	t.Status = StatusCancelled
	now := time.Now()
	t.CancelledAt = &now
	t.UpdatedAt = now
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

// Publish transitions the tournament from draft to scheduled/open state.
func (t *Tournament) Publish() error {
	if t.Status != StatusDraft {
		return fmt.Errorf("can only publish tournament from draft status, current: %s", t.Status)
	}
	t.Status = StatusScheduled
	t.UpdatedAt = time.Now()
	return nil
}

// OpenRegistration transitions the tournament to registration_open state.
func (t *Tournament) OpenRegistration() error {
	if t.Status != StatusScheduled && t.Status != StatusDraft {
		return fmt.Errorf("can only open registration from draft or scheduled status, current: %s", t.Status)
	}
	t.Status = StatusRegistrationOpen
	t.UpdatedAt = time.Now()
	return nil
}

// UpdateDetails updates the tournament's basic details.
func (t *Tournament) UpdateDetails(name, description *string, scheduledAt *time.Time, minCount, maxCount *int32) {
	if name != nil {
		t.Name = *name
	}
	if description != nil {
		t.Description = description
	}
	if scheduledAt != nil {
		t.ScheduledStartTimeAt = scheduledAt
	}
	if minCount != nil {
		t.MinimumTeamCount = *minCount
	}
	if maxCount != nil {
		t.MaximumTeamCount = maxCount
	}
	t.UpdatedAt = time.Now()
}
