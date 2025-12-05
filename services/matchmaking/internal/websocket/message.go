// Package websocket handles WebSocket connections for match lobbies
package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeHeartbeat        MessageType = "heartbeat"
	MessageTypeLobbyState       MessageType = "lobby_state"
	MessageTypePlayerReady      MessageType = "player_ready"
	MessageTypeMatchStart       MessageType = "match_start"
	MessageTypeMatchPause       MessageType = "match_pause"
	MessageTypeMatchResume      MessageType = "match_resume"
	MessageTypePlayerDisconnect MessageType = "player_disconnect"
	MessageTypePlayerReconnect  MessageType = "player_reconnect"
	MessageTypeScoreUpdate      MessageType = "score_update"
	MessageTypeMatchComplete    MessageType = "match_complete"
	MessageTypeError            MessageType = "error"

	// Pre-lobby message types
	MessageTypePreLobbyState       MessageType = "prelobby_state"
	MessageTypeParticipantJoined   MessageType = "participant_joined"
	MessageTypeParticipantLeft     MessageType = "participant_left"
	MessageTypeGracePeriodStarted  MessageType = "grace_period_started"
	MessageTypeRosterUpdated       MessageType = "roster_updated"
	MessageTypeGracePeriodEnded    MessageType = "grace_period_ended"
	MessageTypeMatchAssigned       MessageType = "match_assigned"
	MessageTypeTournamentCancelled MessageType = "tournament_cancelled"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// NewMessage creates a new WebSocket message
func NewMessage(msgType MessageType, payload interface{}) (*Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      msgType,
		Timestamp: time.Now(),
		Payload:   payloadBytes,
	}, nil
}

// ============================================================================
// Message Payloads
// ============================================================================

// HeartbeatPayload for heartbeat messages
type HeartbeatPayload struct {
	PlayerID uuid.UUID `json:"player_id"`
}

// LobbyStatePayload contains current lobby state
type LobbyStatePayload struct {
	MatchID           uuid.UUID  `json:"match_id"`
	Status            string     `json:"status"`
	Participant1ID    uuid.UUID  `json:"participant1_id"`
	Participant2ID    *uuid.UUID `json:"participant2_id"`
	Participant1Ready bool       `json:"participant1_ready"`
	Participant2Ready bool       `json:"participant2_ready"`
	BothReady         bool       `json:"both_ready"`
}

// PlayerReadyPayload for player ready notifications
type PlayerReadyPayload struct {
	PlayerID uuid.UUID `json:"player_id"`
	Ready    bool      `json:"ready"`
}

// MatchStartPayload for match start notifications
type MatchStartPayload struct {
	MatchID   uuid.UUID `json:"match_id"`
	StartedAt time.Time `json:"started_at"`
}

// MatchPausePayload for match pause notifications
type MatchPausePayload struct {
	MatchID              uuid.UUID `json:"match_id"`
	DisconnectedPlayerID uuid.UUID `json:"disconnected_player_id"`
	PausedAt             time.Time `json:"paused_at"`
	ReconnectWindow      int       `json:"reconnect_window_seconds"` // 30
}

// MatchResumePayload for match resume notifications
type MatchResumePayload struct {
	MatchID   uuid.UUID `json:"match_id"`
	ResumedAt time.Time `json:"resumed_at"`
}

// PlayerDisconnectPayload for disconnect notifications
type PlayerDisconnectPayload struct {
	PlayerID       uuid.UUID `json:"player_id"`
	DisconnectedAt time.Time `json:"disconnected_at"`
	PointAwarded   bool      `json:"point_awarded"`
}

// PlayerReconnectPayload for reconnect notifications
type PlayerReconnectPayload struct {
	PlayerID       uuid.UUID `json:"player_id"`
	ReconnectedAt  time.Time `json:"reconnected_at"`
	ElapsedSeconds int       `json:"elapsed_seconds"`
}

// ScoreUpdatePayload for score updates during match
type ScoreUpdatePayload struct {
	MatchID      uuid.UUID `json:"match_id"`
	ScorePlayer1 int       `json:"score_player1"`
	ScorePlayer2 int       `json:"score_player2"`
}

// MatchCompletePayload for match completion notifications
type MatchCompletePayload struct {
	MatchID      uuid.UUID `json:"match_id"`
	WinnerID     uuid.UUID `json:"winner_id"`
	LoserID      uuid.UUID `json:"loser_id"`
	ScorePlayer1 int       `json:"score_player1"`
	ScorePlayer2 int       `json:"score_player2"`
	ResultSource string    `json:"result_source"`
	CompletedAt  time.Time `json:"completed_at"`
}

// ErrorPayload for error messages
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ============================================================================
// Pre-Lobby Message Payloads
// ============================================================================

// PreLobbyParticipantPayload represents a participant in pre-lobby messages
type PreLobbyParticipantPayload struct {
	UserID      uuid.UUID  `json:"user_id"`
	DisplayName string     `json:"display_name"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	JoinedAt    time.Time  `json:"joined_at"`
}

// GracePeriodInfoPayload represents grace period timing info
type GracePeriodInfoPayload struct {
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	RemainingSeconds int       `json:"remaining_seconds"`
}

// ActivityFeedItemPayload represents an activity feed item
type ActivityFeedItemPayload struct {
	ID              uuid.UUID `json:"id"`
	EventType       string    `json:"event_type"`
	Message         string    `json:"message"`
	ParticipantName *string   `json:"participant_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// PreLobbyStatePayload contains complete pre-lobby state (sent on connection)
type PreLobbyStatePayload struct {
	TournamentID     uuid.UUID                    `json:"tournament_id"`
	TournamentName   string                       `json:"tournament_name"`
	Status           string                       `json:"status"`
	Participants     []PreLobbyParticipantPayload `json:"participants"`
	ParticipantCount int                          `json:"participant_count"`
	MinParticipants  int                          `json:"min_participants"`
	StartTime        time.Time                    `json:"start_time"`
	GracePeriod      *GracePeriodInfoPayload      `json:"grace_period,omitempty"`
	ActivityFeed     []ActivityFeedItemPayload    `json:"activity_feed"`
}

// ParticipantJoinedPayload for participant join notifications
type ParticipantJoinedPayload struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	JoinedAt    time.Time `json:"joined_at"`
}

// ParticipantLeftPayload for participant leave notifications
type ParticipantLeftPayload struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
}

// GracePeriodStartedPayload for grace period start notifications
type GracePeriodStartedPayload struct {
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	ParticipantCount int       `json:"participant_count"`
}

// RosterUpdatedPayload for participant count changes during grace period
type RosterUpdatedPayload struct {
	ParticipantCount int       `json:"participant_count"`
	Action           string    `json:"action"` // "joined" or "left"
	UserID           uuid.UUID `json:"user_id"`
	DisplayName      string    `json:"display_name"`
}

// GracePeriodEndedPayload for grace period end notifications
type GracePeriodEndedPayload struct {
	FinalParticipantCount int    `json:"final_participant_count"`
	Status                string `json:"status"` // "generating_bracket" or "cancelled"
}

// MatchAssignedPayload for match assignment notifications
type MatchAssignedPayload struct {
	MatchID     uuid.UUID                   `json:"match_id"`
	RoundNumber int                         `json:"round_number"`
	MatchNumber int                         `json:"match_number"`
	Opponent    *PreLobbyParticipantPayload `json:"opponent,omitempty"` // nil for bye
}

// TournamentCancelledPayload for tournament cancellation notifications
type TournamentCancelledPayload struct {
	Reason string `json:"reason"`
}

