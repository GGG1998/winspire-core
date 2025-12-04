package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MatchStatus represents the state of a match
type MatchStatus string

const (
	MatchStatusPending   MatchStatus = "pending"
	MatchStatusReady     MatchStatus = "ready"
	MatchStatusStarted   MatchStatus = "started"
	MatchStatusPaused    MatchStatus = "paused"
	MatchStatusCompleted MatchStatus = "completed"
	MatchStatusCancelled MatchStatus = "cancelled"
)

// ResultSource indicates how the match result was determined
type ResultSource string

const (
	ResultSourceGameAPI     ResultSource = "game_api"
	ResultSourceHostManual  ResultSource = "host_manual"
	ResultSourceWalkover    ResultSource = "walkover"
)

// Match represents an individual 1v1 match aggregate
type Match struct {
	ID                   uuid.UUID
	RoundID              uuid.UUID
	MatchNumber          int
	NextMatchID          *uuid.UUID
	Participant1ID       uuid.UUID
	Participant2ID       *uuid.UUID // Nil for bye matches
	Status               MatchStatus
	Participant1Ready    bool
	Participant2Ready    bool
	WinnerID             *uuid.UUID
	ScorePlayer1         *int
	ScorePlayer2         *int
	ResultSource         *ResultSource
	DisconnectedPlayerID *uuid.UUID
	DisconnectedAt       *time.Time
	GameAPIMatchID       *string
	GameAPIPollAttempts  int
	GameAPILastPoll      *time.Time
	CreatedAt            time.Time
	StartedAt            *time.Time
	CompletedAt          *time.Time
	UpdatedAt            time.Time
}

// NewMatch creates a new match
func NewMatch(roundID uuid.UUID, matchNumber int, participant1ID uuid.UUID, participant2ID *uuid.UUID, nextMatchID *uuid.UUID) *Match {
	now := time.Now()
	return &Match{
		ID:                uuid.New(),
		RoundID:           roundID,
		MatchNumber:       matchNumber,
		NextMatchID:       nextMatchID,
		Participant1ID:    participant1ID,
		Participant2ID:    participant2ID,
		Status:            MatchStatusPending,
		Participant1Ready: false,
		Participant2Ready: false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// IsBye returns true if this is a bye match (no opponent)
func (m *Match) IsBye() bool {
	return m.Participant2ID == nil || *m.Participant2ID == uuid.Nil
}

// MarkPlayerReady marks a player as ready
func (m *Match) MarkPlayerReady(playerID uuid.UUID) error {
	if m.Status != MatchStatusPending {
		return fmt.Errorf("can only mark ready in pending state, current status: %s", m.Status)
	}

	if playerID == m.Participant1ID {
		m.Participant1Ready = true
	} else if m.Participant2ID != nil && playerID == *m.Participant2ID {
		m.Participant2Ready = true
	} else {
		return fmt.Errorf("player %s is not a participant in this match", playerID)
	}

	m.UpdatedAt = time.Now()

	// Auto-transition to ready if both players are ready
	if m.BothPlayersReady() {
		m.Status = MatchStatusReady
	}

	return nil
}

// BothPlayersReady returns true if both players are ready
func (m *Match) BothPlayersReady() bool {
	// Bye matches are always ready
	if m.IsBye() {
		return m.Participant1Ready
	}

	return m.Participant1Ready && m.Participant2Ready
}

// Start transitions match to started state
func (m *Match) Start() error {
	if m.Status != MatchStatusReady && m.Status != MatchStatusPending {
		return fmt.Errorf("can only start match from ready or pending state, current status: %s", m.Status)
	}

	now := time.Now()
	m.Status = MatchStatusStarted
	m.StartedAt = &now
	m.UpdatedAt = now

	return nil
}

// Pause pauses the match (for disconnection)
func (m *Match) Pause(disconnectedPlayerID uuid.UUID) error {
	if m.Status != MatchStatusStarted {
		return fmt.Errorf("can only pause started match, current status: %s", m.Status)
	}

	if disconnectedPlayerID != m.Participant1ID && (m.Participant2ID == nil || disconnectedPlayerID != *m.Participant2ID) {
		return fmt.Errorf("disconnected player %s is not a participant", disconnectedPlayerID)
	}

	now := time.Now()
	m.Status = MatchStatusPaused
	m.DisconnectedPlayerID = &disconnectedPlayerID
	m.DisconnectedAt = &now
	m.UpdatedAt = now

	return nil
}

// Resume resumes the match after reconnection
func (m *Match) Resume() error {
	if m.Status != MatchStatusPaused {
		return fmt.Errorf("can only resume paused match, current status: %s", m.Status)
	}

	m.Status = MatchStatusStarted
	m.DisconnectedPlayerID = nil
	m.DisconnectedAt = nil
	m.UpdatedAt = time.Now()

	return nil
}

// Complete marks the match as completed with results
func (m *Match) Complete(winnerID uuid.UUID, scorePlayer1, scorePlayer2 int, source ResultSource) error {
	if m.Status != MatchStatusStarted && m.Status != MatchStatusPending && m.Status != MatchStatusPaused {
		return fmt.Errorf("can only complete started/pending/paused match, current status: %s", m.Status)
	}

	// Validate winner is a participant
	if winnerID != m.Participant1ID && (m.Participant2ID == nil || winnerID != *m.Participant2ID) {
		return fmt.Errorf("winner %s is not a participant", winnerID)
	}

	// Validate scores
	if scorePlayer1 < 0 || scorePlayer2 < 0 {
		return fmt.Errorf("scores cannot be negative")
	}

	now := time.Now()
	m.Status = MatchStatusCompleted
	m.WinnerID = &winnerID
	m.ScorePlayer1 = &scorePlayer1
	m.ScorePlayer2 = &scorePlayer2
	m.ResultSource = &source
	m.CompletedAt = &now
	m.UpdatedAt = now

	return nil
}

// Cancel cancels the match
func (m *Match) Cancel() error {
	if m.Status == MatchStatusCompleted || m.Status == MatchStatusCancelled {
		return fmt.Errorf("cannot cancel completed or already cancelled match")
	}

	now := time.Now()
	m.Status = MatchStatusCancelled
	m.CompletedAt = &now
	m.UpdatedAt = now

	return nil
}

// AwardPointToOpponent awards a point to the opponent of the disconnected player
func (m *Match) AwardPointToOpponent(disconnectedPlayerID uuid.UUID) error {
	if disconnectedPlayerID != m.Participant1ID && (m.Participant2ID == nil || disconnectedPlayerID != *m.Participant2ID) {
		return fmt.Errorf("disconnected player %s is not a participant", disconnectedPlayerID)
	}

	// Initialize scores if nil
	if m.ScorePlayer1 == nil {
		zero := 0
		m.ScorePlayer1 = &zero
	}
	if m.ScorePlayer2 == nil {
		zero := 0
		m.ScorePlayer2 = &zero
	}

	// Award point to opponent
	if disconnectedPlayerID == m.Participant1ID {
		*m.ScorePlayer2++
	} else {
		*m.ScorePlayer1++
	}

	m.UpdatedAt = time.Now()
	return nil
}

// GetOpponentID returns the opponent's ID for a given player
func (m *Match) GetOpponentID(playerID uuid.UUID) (*uuid.UUID, error) {
	if playerID == m.Participant1ID {
		return m.Participant2ID, nil
	} else if m.Participant2ID != nil && playerID == *m.Participant2ID {
		return &m.Participant1ID, nil
	}

	return nil, fmt.Errorf("player %s is not a participant", playerID)
}

// Validate checks match business invariants
func (m *Match) Validate() error {
	if m.MatchNumber < 1 {
		return fmt.Errorf("match number must be positive")
	}

	if m.Participant1ID == uuid.Nil {
		return fmt.Errorf("participant1 cannot be nil")
	}

	if m.Status == MatchStatusCompleted {
		if m.WinnerID == nil {
			return fmt.Errorf("completed match must have a winner")
		}
		if m.ScorePlayer1 == nil || m.ScorePlayer2 == nil {
			return fmt.Errorf("completed match must have scores")
		}
		if m.ResultSource == nil {
			return fmt.Errorf("completed match must have result source")
		}
	}

	return nil
}

