package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PreLobbyStatus represents the current state of a pre-lobby
type PreLobbyStatus string

const (
	PreLobbyStatusWaiting           PreLobbyStatus = "waiting"
	PreLobbyStatusGracePeriod       PreLobbyStatus = "grace_period"
	PreLobbyStatusGeneratingBracket PreLobbyStatus = "generating_bracket"
	PreLobbyStatusStarted           PreLobbyStatus = "started"
	PreLobbyStatusCancelled         PreLobbyStatus = "cancelled"
)

// PreLobby represents the waiting room state for a tournament before it starts.
// This is the aggregate root for pre-lobby functionality.
type PreLobby struct {
	ID               uuid.UUID
	TournamentID     uuid.UUID
	Status           PreLobbyStatus
	GracePeriodStart *time.Time
	GracePeriodEnd   *time.Time
	MinParticipants  int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewPreLobby creates a new pre-lobby for a tournament
func NewPreLobby(tournamentID uuid.UUID, minParticipants int) (*PreLobby, error) {
	if minParticipants < 2 {
		return nil, fmt.Errorf("minimum participants must be at least 2, got %d", minParticipants)
	}

	now := time.Now()
	return &PreLobby{
		ID:              uuid.New(),
		TournamentID:    tournamentID,
		Status:          PreLobbyStatusWaiting,
		MinParticipants: minParticipants,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// CanTransitionTo checks if a status transition is valid
func (s PreLobbyStatus) CanTransitionTo(target PreLobbyStatus) bool {
	transitions := map[PreLobbyStatus][]PreLobbyStatus{
		PreLobbyStatusWaiting:           {PreLobbyStatusGracePeriod, PreLobbyStatusCancelled},
		PreLobbyStatusGracePeriod:       {PreLobbyStatusGeneratingBracket, PreLobbyStatusCancelled},
		PreLobbyStatusGeneratingBracket: {PreLobbyStatusStarted, PreLobbyStatusCancelled},
		PreLobbyStatusStarted:           {}, // Terminal state
		PreLobbyStatusCancelled:         {}, // Terminal state
	}

	allowedTransitions, exists := transitions[s]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == target {
			return true
		}
	}
	return false
}

// StartGracePeriod transitions the pre-lobby to grace period status
func (p *PreLobby) StartGracePeriod() error {
	if !p.Status.CanTransitionTo(PreLobbyStatusGracePeriod) {
		return fmt.Errorf("cannot start grace period from status %s", p.Status)
	}

	now := time.Now()
	end := now.Add(30 * time.Second)

	p.Status = PreLobbyStatusGracePeriod
	p.GracePeriodStart = &now
	p.GracePeriodEnd = &end
	p.UpdatedAt = now

	return nil
}

// EndGracePeriod transitions to generating_bracket status
func (p *PreLobby) EndGracePeriod() error {
	if !p.Status.CanTransitionTo(PreLobbyStatusGeneratingBracket) {
		return fmt.Errorf("cannot end grace period from status %s", p.Status)
	}

	p.Status = PreLobbyStatusGeneratingBracket
	p.UpdatedAt = time.Now()

	return nil
}

// MarkStarted transitions to started status after bracket generation
func (p *PreLobby) MarkStarted() error {
	if !p.Status.CanTransitionTo(PreLobbyStatusStarted) {
		return fmt.Errorf("cannot mark started from status %s", p.Status)
	}

	p.Status = PreLobbyStatusStarted
	p.UpdatedAt = time.Now()

	return nil
}

// Cancel transitions to cancelled status
func (p *PreLobby) Cancel() error {
	if !p.Status.CanTransitionTo(PreLobbyStatusCancelled) {
		return fmt.Errorf("cannot cancel from status %s", p.Status)
	}

	p.Status = PreLobbyStatusCancelled
	p.UpdatedAt = time.Now()

	return nil
}

// IsInGracePeriod checks if the pre-lobby is currently in grace period
func (p *PreLobby) IsInGracePeriod() bool {
	if p.Status != PreLobbyStatusGracePeriod {
		return false
	}
	if p.GracePeriodEnd == nil {
		return false
	}
	return time.Now().Before(*p.GracePeriodEnd)
}

// GracePeriodRemaining returns the time remaining in the grace period
func (p *PreLobby) GracePeriodRemaining() time.Duration {
	if p.GracePeriodEnd == nil {
		return 0
	}
	remaining := time.Until(*p.GracePeriodEnd)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CanAcceptParticipants checks if the pre-lobby can accept new participants
func (p *PreLobby) CanAcceptParticipants() bool {
	return p.Status == PreLobbyStatusWaiting || p.Status == PreLobbyStatusGracePeriod
}

// PreLobbyParticipant represents a connected participant in the pre-lobby
type PreLobbyParticipant struct {
	UserID      uuid.UUID
	DisplayName string
	AvatarURL   *string
	JoinedAt    time.Time
}

// NewPreLobbyParticipant creates a new participant
func NewPreLobbyParticipant(userID uuid.UUID, displayName string, avatarURL *string) *PreLobbyParticipant {
	return &PreLobbyParticipant{
		UserID:      userID,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
		JoinedAt:    time.Now(),
	}
}

// ParticipantSnapshot represents the immutable list of participants when grace period ended
type ParticipantSnapshot struct {
	ID               uuid.UUID
	TournamentID     uuid.UUID
	Participants     []PreLobbyParticipant
	ParticipantCount int
	CreatedAt        time.Time
}

// NewParticipantSnapshot creates a new snapshot from current participants
func NewParticipantSnapshot(tournamentID uuid.UUID, participants []PreLobbyParticipant) *ParticipantSnapshot {
	return &ParticipantSnapshot{
		ID:               uuid.New(),
		TournamentID:     tournamentID,
		Participants:     participants,
		ParticipantCount: len(participants),
		CreatedAt:        time.Now(),
	}
}

// ActivityFeedEventType represents the type of activity feed event
type ActivityFeedEventType string

const (
	ActivityEventParticipantJoined   ActivityFeedEventType = "participant_joined"
	ActivityEventParticipantLeft     ActivityFeedEventType = "participant_left"
	ActivityEventGracePeriodStarted  ActivityFeedEventType = "grace_period_started"
	ActivityEventGracePeriodEnded    ActivityFeedEventType = "grace_period_ended"
	ActivityEventBracketGeneration   ActivityFeedEventType = "bracket_generation"
	ActivityEventTournamentCancelled ActivityFeedEventType = "tournament_cancelled"
	ActivityEventSystemMessage       ActivityFeedEventType = "system_message"
)

// ActivityFeedItem represents a single event in the pre-lobby history
type ActivityFeedItem struct {
	ID              uuid.UUID
	TournamentID    uuid.UUID
	EventType       ActivityFeedEventType
	Message         string
	ParticipantName *string
	CreatedAt       time.Time
}

// NewActivityFeedItem creates a new activity feed item
func NewActivityFeedItem(tournamentID uuid.UUID, eventType ActivityFeedEventType, message string, participantName *string) *ActivityFeedItem {
	return &ActivityFeedItem{
		ID:              uuid.New(),
		TournamentID:    tournamentID,
		EventType:       eventType,
		Message:         message,
		ParticipantName: participantName,
		CreatedAt:       time.Now(),
	}
}
