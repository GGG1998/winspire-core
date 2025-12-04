package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

// DomainEvent represents a domain event that occurred in the system
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	AggregateType() string
	BoundedContext() string
	Timestamp() time.Time
	Payload() interface{}
	Metadata() map[string]string
}

// BaseEvent provides common fields for all domain events
type BaseEvent struct {
	ID            string
	Type          string
	AggID         string
	AggType       string
	Context       string
	OccurredAt    time.Time
	EventPayload  interface{}
	EventMetadata map[string]string
}

func (e BaseEvent) EventID() string             { return e.ID }
func (e BaseEvent) EventType() string           { return e.Type }
func (e BaseEvent) AggregateID() string         { return e.AggID }
func (e BaseEvent) AggregateType() string       { return e.AggType }
func (e BaseEvent) BoundedContext() string      { return e.Context }
func (e BaseEvent) Timestamp() time.Time        { return e.OccurredAt }
func (e BaseEvent) Payload() interface{}        { return e.EventPayload }
func (e BaseEvent) Metadata() map[string]string { return e.EventMetadata }

// newBaseEvent creates a base event with common fields
func newBaseEvent(eventType, aggregateID, aggregateType string, payload interface{}, metadata map[string]string) BaseEvent {
	return BaseEvent{
		ID:            ulid.Make().String(),
		Type:          eventType,
		AggID:         aggregateID,
		AggType:       aggregateType,
		Context:       "Matchmaking",
		OccurredAt:    time.Now(),
		EventPayload:  payload,
		EventMetadata: metadata,
	}
}

// ============================================================================
// Bracket Events
// ============================================================================

// BracketGeneratedPayload contains bracket generation details
type BracketGeneratedPayload struct {
	BracketID    uuid.UUID   `json:"bracket_id"`
	TournamentID uuid.UUID   `json:"tournament_id"`
	TotalRounds  int         `json:"total_rounds"`
	TotalMatches int         `json:"total_matches"`
	ByesCount    int         `json:"byes_count"`
	Participants []uuid.UUID `json:"participants"`
}

// BracketGenerated event fired when bracket is successfully generated
type BracketGenerated struct {
	BaseEvent
}

func NewBracketGenerated(bracketID, tournamentID uuid.UUID, totalRounds, totalMatches, byesCount int, participants []uuid.UUID, metadata map[string]string) BracketGenerated {
	payload := BracketGeneratedPayload{
		BracketID:    bracketID,
		TournamentID: tournamentID,
		TotalRounds:  totalRounds,
		TotalMatches: totalMatches,
		ByesCount:    byesCount,
		Participants: participants,
	}
	return BracketGenerated{
		BaseEvent: newBaseEvent("BracketGenerated", bracketID.String(), "Bracket", payload, metadata),
	}
}

// ============================================================================
// Round Events
// ============================================================================

// RoundCreatedPayload contains round creation details
type RoundCreatedPayload struct {
	RoundID      uuid.UUID `json:"round_id"`
	BracketID    uuid.UUID `json:"bracket_id"`
	TournamentID uuid.UUID `json:"tournament_id"`
	RoundNumber  int       `json:"round_number"`
	RoundName    string    `json:"round_name"`
	MatchesCount int       `json:"matches_count"`
}

// RoundCreated event fired when a round is created
type RoundCreated struct {
	BaseEvent
}

func NewRoundCreated(roundID, bracketID, tournamentID uuid.UUID, roundNumber int, roundName string, matchesCount int, metadata map[string]string) RoundCreated {
	payload := RoundCreatedPayload{
		RoundID:      roundID,
		BracketID:    bracketID,
		TournamentID: tournamentID,
		RoundNumber:  roundNumber,
		RoundName:    roundName,
		MatchesCount: matchesCount,
	}
	return RoundCreated{
		BaseEvent: newBaseEvent("RoundCreated", roundID.String(), "Round", payload, metadata),
	}
}

// ============================================================================
// Match Events
// ============================================================================

// MatchCreatedPayload contains match creation details
type MatchCreatedPayload struct {
	MatchID        uuid.UUID  `json:"match_id"`
	TournamentID   uuid.UUID  `json:"tournament_id"`
	RoundID        uuid.UUID  `json:"round_id"`
	RoundNumber    int        `json:"round_number"`
	MatchNumber    int        `json:"match_number"`
	Participant1ID uuid.UUID  `json:"participant1_id"`
	Participant2ID *uuid.UUID `json:"participant2_id"` // Nil for bye
	NextMatchID    *uuid.UUID `json:"next_match_id"`
	IsBye          bool       `json:"is_bye"`
}

// MatchCreated event fired when a match is created
type MatchCreated struct {
	BaseEvent
}

func NewMatchCreated(matchID, tournamentID, roundID uuid.UUID, roundNumber, matchNumber int, participant1ID uuid.UUID, participant2ID, nextMatchID *uuid.UUID, isBye bool, metadata map[string]string) MatchCreated {
	payload := MatchCreatedPayload{
		MatchID:        matchID,
		TournamentID:   tournamentID,
		RoundID:        roundID,
		RoundNumber:    roundNumber,
		MatchNumber:    matchNumber,
		Participant1ID: participant1ID,
		Participant2ID: participant2ID,
		NextMatchID:    nextMatchID,
		IsBye:          isBye,
	}
	return MatchCreated{
		BaseEvent: newBaseEvent("MatchCreated", matchID.String(), "Match", payload, metadata),
	}
}

// MatchStartedPayload contains match start details
type MatchStartedPayload struct {
	MatchID        uuid.UUID `json:"match_id"`
	TournamentID   uuid.UUID `json:"tournament_id"`
	Participant1ID uuid.UUID `json:"participant1_id"`
	Participant2ID uuid.UUID `json:"participant2_id"`
	StartedAt      time.Time `json:"started_at"`
}

// MatchStarted event fired when match begins
type MatchStarted struct {
	BaseEvent
}

func NewMatchStarted(matchID, tournamentID, participant1ID, participant2ID uuid.UUID, startedAt time.Time, metadata map[string]string) MatchStarted {
	payload := MatchStartedPayload{
		MatchID:        matchID,
		TournamentID:   tournamentID,
		Participant1ID: participant1ID,
		Participant2ID: participant2ID,
		StartedAt:      startedAt,
	}
	return MatchStarted{
		BaseEvent: newBaseEvent("MatchStarted", matchID.String(), "Match", payload, metadata),
	}
}

// MatchCompletedPayload contains match completion details
type MatchCompletedPayload struct {
	MatchID      uuid.UUID    `json:"match_id"`
	TournamentID uuid.UUID    `json:"tournament_id"`
	WinnerID     uuid.UUID    `json:"winner_id"`
	LoserID      uuid.UUID    `json:"loser_id"`
	ScorePlayer1 int          `json:"score_player1"`
	ScorePlayer2 int          `json:"score_player2"`
	ResultSource ResultSource `json:"result_source"`
	CompletedAt  time.Time    `json:"completed_at"`
	NextMatchID  *uuid.UUID   `json:"next_match_id"`
}

// MatchCompleted event fired when match finishes
type MatchCompleted struct {
	BaseEvent
}

func NewMatchCompleted(matchID, tournamentID, winnerID, loserID uuid.UUID, scorePlayer1, scorePlayer2 int, source ResultSource, completedAt time.Time, nextMatchID *uuid.UUID, metadata map[string]string) MatchCompleted {
	payload := MatchCompletedPayload{
		MatchID:      matchID,
		TournamentID: tournamentID,
		WinnerID:     winnerID,
		LoserID:      loserID,
		ScorePlayer1: scorePlayer1,
		ScorePlayer2: scorePlayer2,
		ResultSource: source,
		CompletedAt:  completedAt,
		NextMatchID:  nextMatchID,
	}
	return MatchCompleted{
		BaseEvent: newBaseEvent("MatchCompleted", matchID.String(), "Match", payload, metadata),
	}
}

// ============================================================================
// Participant Events
// ============================================================================

// ParticipantAdvancedPayload contains advancement details
type ParticipantAdvancedPayload struct {
	PlayerID        uuid.UUID `json:"player_id"`
	TournamentID    uuid.UUID `json:"tournament_id"`
	FromMatchID     uuid.UUID `json:"from_match_id"`
	ToMatchID       uuid.UUID `json:"to_match_id"`
	FromRoundNumber int       `json:"from_round_number"`
	ToRoundNumber   int       `json:"to_round_number"`
}

// ParticipantAdvanced event fired when player advances to next round
type ParticipantAdvanced struct {
	BaseEvent
}

func NewParticipantAdvanced(playerID, tournamentID, fromMatchID, toMatchID uuid.UUID, fromRound, toRound int, metadata map[string]string) ParticipantAdvanced {
	payload := ParticipantAdvancedPayload{
		PlayerID:        playerID,
		TournamentID:    tournamentID,
		FromMatchID:     fromMatchID,
		ToMatchID:       toMatchID,
		FromRoundNumber: fromRound,
		ToRoundNumber:   toRound,
	}
	return ParticipantAdvanced{
		BaseEvent: newBaseEvent("ParticipantAdvanced", playerID.String(), "Participant", payload, metadata),
	}
}

// ParticipantEliminatedPayload contains elimination details
type ParticipantEliminatedPayload struct {
	PlayerID     uuid.UUID `json:"player_id"`
	TournamentID uuid.UUID `json:"tournament_id"`
	MatchID      uuid.UUID `json:"match_id"`
	RoundNumber  int       `json:"round_number"`
	EliminatedAt time.Time `json:"eliminated_at"`
}

// ParticipantEliminated event fired when player is eliminated
type ParticipantEliminated struct {
	BaseEvent
}

func NewParticipantEliminated(playerID, tournamentID, matchID uuid.UUID, roundNumber int, eliminatedAt time.Time, metadata map[string]string) ParticipantEliminated {
	payload := ParticipantEliminatedPayload{
		PlayerID:     playerID,
		TournamentID: tournamentID,
		MatchID:      matchID,
		RoundNumber:  roundNumber,
		EliminatedAt: eliminatedAt,
	}
	return ParticipantEliminated{
		BaseEvent: newBaseEvent("ParticipantEliminated", playerID.String(), "Participant", payload, metadata),
	}
}

// ============================================================================
// Walkover Events
// ============================================================================

// WalkoverGrantedPayload contains walkover details
type WalkoverGrantedPayload struct {
	MatchID        uuid.UUID `json:"match_id"`
	TournamentID   uuid.UUID `json:"tournament_id"`
	WinnerID       uuid.UUID `json:"winner_id"`
	NoShowPlayerID uuid.UUID `json:"no_show_player_id"`
	Reason         string    `json:"reason"`
	GrantedAt      time.Time `json:"granted_at"`
}

// WalkoverGranted event fired when walkover is awarded
type WalkoverGranted struct {
	BaseEvent
}

func NewWalkoverGranted(matchID, tournamentID, winnerID, noShowPlayerID uuid.UUID, reason string, grantedAt time.Time, metadata map[string]string) WalkoverGranted {
	payload := WalkoverGrantedPayload{
		MatchID:        matchID,
		TournamentID:   tournamentID,
		WinnerID:       winnerID,
		NoShowPlayerID: noShowPlayerID,
		Reason:         reason,
		GrantedAt:      grantedAt,
	}
	return WalkoverGranted{
		BaseEvent: newBaseEvent("WalkoverGranted", matchID.String(), "Match", payload, metadata),
	}
}

// ============================================================================
// Disconnect Events
// ============================================================================

// PlayerConnectionLostPayload contains disconnection details
type PlayerConnectionLostPayload struct {
	MatchID        uuid.UUID `json:"match_id"`
	TournamentID   uuid.UUID `json:"tournament_id"`
	PlayerID       uuid.UUID `json:"player_id"`
	DisconnectedAt time.Time `json:"disconnected_at"`
	PointAwarded   bool      `json:"point_awarded"`
}

// PlayerConnectionLost event fired when player disconnects
type PlayerConnectionLost struct {
	BaseEvent
}

func NewPlayerConnectionLost(matchID, tournamentID, playerID uuid.UUID, disconnectedAt time.Time, pointAwarded bool, metadata map[string]string) PlayerConnectionLost {
	payload := PlayerConnectionLostPayload{
		MatchID:        matchID,
		TournamentID:   tournamentID,
		PlayerID:       playerID,
		DisconnectedAt: disconnectedAt,
		PointAwarded:   pointAwarded,
	}
	return PlayerConnectionLost{
		BaseEvent: newBaseEvent("PlayerConnectionLost", matchID.String(), "Match", payload, metadata),
	}
}

// PlayerConnectionRestoredPayload contains reconnection details
type PlayerConnectionRestoredPayload struct {
	MatchID        uuid.UUID `json:"match_id"`
	TournamentID   uuid.UUID `json:"tournament_id"`
	PlayerID       uuid.UUID `json:"player_id"`
	ReconnectedAt  time.Time `json:"reconnected_at"`
	ElapsedSeconds int       `json:"elapsed_seconds"`
}

// PlayerConnectionRestored event fired when player reconnects
type PlayerConnectionRestored struct {
	BaseEvent
}

func NewPlayerConnectionRestored(matchID, tournamentID, playerID uuid.UUID, reconnectedAt time.Time, elapsedSeconds int, metadata map[string]string) PlayerConnectionRestored {
	payload := PlayerConnectionRestoredPayload{
		MatchID:        matchID,
		TournamentID:   tournamentID,
		PlayerID:       playerID,
		ReconnectedAt:  reconnectedAt,
		ElapsedSeconds: elapsedSeconds,
	}
	return PlayerConnectionRestored{
		BaseEvent: newBaseEvent("PlayerConnectionRestored", matchID.String(), "Match", payload, metadata),
	}
}

// ============================================================================
// Tournament Lifecycle Events (Published to Competition Service)
// ============================================================================

// TournamentCompletedPayload contains tournament completion details
type TournamentCompletedPayload struct {
	TournamentID uuid.UUID `json:"tournament_id"`
	ChampionID   uuid.UUID `json:"champion_id"`
	CompletedAt  time.Time `json:"completed_at"`
}

// TournamentCompleted event fired when tournament finishes (final match completed)
type TournamentCompleted struct {
	BaseEvent
}

func NewTournamentCompleted(tournamentID, championID uuid.UUID, completedAt time.Time, metadata map[string]string) TournamentCompleted {
	payload := TournamentCompletedPayload{
		TournamentID: tournamentID,
		ChampionID:   championID,
		CompletedAt:  completedAt,
	}
	return TournamentCompleted{
		BaseEvent: newBaseEvent("TournamentCompleted", tournamentID.String(), "Tournament", payload, metadata),
	}
}
