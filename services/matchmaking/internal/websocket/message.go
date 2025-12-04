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

