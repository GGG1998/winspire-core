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
		Context:       "matchmaking",
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

type NextRoundStartPayload struct {
	TournamentID uuid.UUID `json:"tournament_id"`
	RoundNumber  int       `json:"round_number"`
}

type NextRoundStart struct {
	BaseEvent
}

func NewNextRoundStart(roundNumber int, tournamentID uuid.UUID) NextRoundStart {
	payload := NextRoundStartPayload{
		TournamentID: tournamentID,
		RoundNumber:  roundNumber,
	}
	return NextRoundStart{
		BaseEvent: newBaseEvent("NextRoundStartRequest", tournamentID.String(), "Tournament", payload, make(map[string]string, 0)),
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

// ParticipantJoinedLobby represents when a player joins a match lobby
type ParticipantJoinedLobby struct {
	BaseEvent
}

type ParticipantJoinedLobbyPayload struct {
	MatchID       uuid.UUID `json:"match_id"`
	ParticipantID uuid.UUID `json:"participant_id"`
	JoinedAt      time.Time `json:"joined_at"`
}

// NewParticipantJoinedLobby creates a new ParticipantJoinedLobby event
func NewParticipantJoinedLobby(eventID, matchID, participantID uuid.UUID, metadata map[string]string) ParticipantJoinedLobby {
	payload := ParticipantJoinedLobbyPayload{
		MatchID:       matchID,
		ParticipantID: participantID,
		JoinedAt:      time.Now(),
	}
	baseEvent := newBaseEvent(
		"ParticipantJoinedLobby",
		matchID.String(),
		"Match",
		payload,
		metadata,
	)
	return ParticipantJoinedLobby{BaseEvent: baseEvent}
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

// ScoreRetrieved represents when a score is retrieved from the Game API
type ScoreRetrieved struct {
	BaseEvent
}

type ScoreRetrievedPayload struct {
	MatchID      uuid.UUID `json:"match_id"`
	TournamentID uuid.UUID `json:"tournament_id"`
	WinnerID     uuid.UUID `json:"winner_id"`
	ScorePlayer1 int       `json:"score_player1"`
	ScorePlayer2 int       `json:"score_player2"`
	GameMatchID  string    `json:"game_match_id"`
	RetrievedAt  time.Time `json:"retrieved_at"`
}

// NewScoreRetrieved creates a new ScoreRetrieved event
func NewScoreRetrieved(matchID, tournamentID, winnerID uuid.UUID, scorePlayer1, scorePlayer2 int, gameMatchID string, retrievedAt time.Time, metadata map[string]string) ScoreRetrieved {
	payload := ScoreRetrievedPayload{
		MatchID:      matchID,
		TournamentID: tournamentID,
		WinnerID:     winnerID,
		ScorePlayer1: scorePlayer1,
		ScorePlayer2: scorePlayer2,
		GameMatchID:  gameMatchID,
		RetrievedAt:  retrievedAt,
	}
	return ScoreRetrieved{
		BaseEvent: newBaseEvent("ScoreRetrieved", matchID.String(), "Match", payload, metadata),
	}
}

// ============================================================================
// Pre-Lobby Events
// ============================================================================

// PreLobbyCreatedPayload contains pre-lobby creation details
type PreLobbyCreatedPayload struct {
	PreLobbyID      uuid.UUID `json:"prelobby_id"`
	TournamentID    uuid.UUID `json:"tournament_id"`
	MinParticipants int       `json:"min_participants"`
	CreatedAt       time.Time `json:"created_at"`
}

// PreLobbyCreated event fired when a pre-lobby is created
type PreLobbyCreated struct {
	BaseEvent
}

func NewPreLobbyCreated(preLobbyID, tournamentID uuid.UUID, minParticipants int, createdAt time.Time, metadata map[string]string) PreLobbyCreated {
	payload := PreLobbyCreatedPayload{
		PreLobbyID:      preLobbyID,
		TournamentID:    tournamentID,
		MinParticipants: minParticipants,
		CreatedAt:       createdAt,
	}
	return PreLobbyCreated{
		BaseEvent: newBaseEvent("PreLobbyCreated", preLobbyID.String(), "PreLobby", payload, metadata),
	}
}

// PreLobbyGracePeriodStartedPayload contains grace period start details
type PreLobbyGracePeriodStartedPayload struct {
	TournamentID     uuid.UUID `json:"tournament_id"`
	GracePeriodStart time.Time `json:"grace_period_start"`
	GracePeriodEnd   time.Time `json:"grace_period_end"`
	ParticipantCount int       `json:"participant_count"`
}

// PreLobbyGracePeriodStarted event fired when grace period begins
type PreLobbyGracePeriodStarted struct {
	BaseEvent
}

func NewPreLobbyGracePeriodStarted(tournamentID uuid.UUID, start, end time.Time, participantCount int, metadata map[string]string) PreLobbyGracePeriodStarted {
	payload := PreLobbyGracePeriodStartedPayload{
		TournamentID:     tournamentID,
		GracePeriodStart: start,
		GracePeriodEnd:   end,
		ParticipantCount: participantCount,
	}
	return PreLobbyGracePeriodStarted{
		BaseEvent: newBaseEvent("PreLobbyGracePeriodStarted", tournamentID.String(), "PreLobby", payload, metadata),
	}
}

// PreLobbyGracePeriodEndedPayload contains grace period end details
type PreLobbyGracePeriodEndedPayload struct {
	TournamentID          uuid.UUID `json:"tournament_id"`
	FinalParticipantCount int       `json:"final_participant_count"`
	Status                string    `json:"status"` // "generating_bracket" or "cancelled"
}

// PreLobbyGracePeriodEnded event fired when grace period ends
type PreLobbyGracePeriodEnded struct {
	BaseEvent
}

func NewPreLobbyGracePeriodEnded(tournamentID uuid.UUID, finalCount int, status string, metadata map[string]string) PreLobbyGracePeriodEnded {
	payload := PreLobbyGracePeriodEndedPayload{
		TournamentID:          tournamentID,
		FinalParticipantCount: finalCount,
		Status:                status,
	}
	return PreLobbyGracePeriodEnded{
		BaseEvent: newBaseEvent("PreLobbyGracePeriodEnded", tournamentID.String(), "PreLobby", payload, metadata),
	}
}

// PreLobbyParticipantSnapshotPayload contains participant snapshot details
type PreLobbyParticipantSnapshotPayload struct {
	TournamentID     uuid.UUID   `json:"tournament_id"`
	ParticipantIDs   []uuid.UUID `json:"participant_ids"`
	ParticipantCount int         `json:"participant_count"`
	CreatedAt        time.Time   `json:"created_at"`
}

// PreLobbyParticipantSnapshot event fired when participant snapshot is created
type PreLobbyParticipantSnapshot struct {
	BaseEvent
}

func NewPreLobbyParticipantSnapshot(tournamentID uuid.UUID, participantIDs []uuid.UUID, createdAt time.Time, metadata map[string]string) PreLobbyParticipantSnapshot {
	payload := PreLobbyParticipantSnapshotPayload{
		TournamentID:     tournamentID,
		ParticipantIDs:   participantIDs,
		ParticipantCount: len(participantIDs),
		CreatedAt:        createdAt,
	}
	return PreLobbyParticipantSnapshot{
		BaseEvent: newBaseEvent("PreLobbyParticipantSnapshot", tournamentID.String(), "PreLobby", payload, metadata),
	}
}

// PreLobbyCancelledPayload contains pre-lobby cancellation details
type PreLobbyCancelledPayload struct {
	TournamentID uuid.UUID `json:"tournament_id"`
	Reason       string    `json:"reason"`
	CancelledAt  time.Time `json:"cancelled_at"`
}

// PreLobbyCancelled event fired when pre-lobby is cancelled
type PreLobbyCancelled struct {
	BaseEvent
}

func NewPreLobbyCancelled(tournamentID uuid.UUID, reason string, cancelledAt time.Time, metadata map[string]string) PreLobbyCancelled {
	payload := PreLobbyCancelledPayload{
		TournamentID: tournamentID,
		Reason:       reason,
		CancelledAt:  cancelledAt,
	}
	return PreLobbyCancelled{
		BaseEvent: newBaseEvent("PreLobbyCancelled", tournamentID.String(), "PreLobby", payload, metadata),
	}
}

// ============================================================================
// Tournament Start Saga Events
// ============================================================================

// GracePeriodStartedPayload contains grace period start details for saga coordination
type GracePeriodStartedPayload struct {
	TournamentID uuid.UUID `json:"tournament_id"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

// GracePeriodStarted event fired when grace period successfully starts
// This confirms to tournament service that the saga can proceed
type GracePeriodStarted struct {
	BaseEvent
}

func NewGracePeriodStarted(tournamentID uuid.UUID, startTime, endTime time.Time, metadata map[string]string) GracePeriodStarted {
	payload := GracePeriodStartedPayload{
		TournamentID: tournamentID,
		StartTime:    startTime,
		EndTime:      endTime,
	}
	return GracePeriodStarted{
		BaseEvent: newBaseEvent("GracePeriodStarted", tournamentID.String(), "Tournament", payload, metadata),
	}
}

// BracketGenerationCompletedPayload contains bracket generation success details
type BracketGenerationCompletedPayload struct {
	TournamentID uuid.UUID `json:"tournament_id"`
	BracketID    uuid.UUID `json:"bracket_id"`
	TotalRounds  int       `json:"total_rounds"`
	TotalMatches int       `json:"total_matches"`
	CompletedAt  time.Time `json:"completed_at"`
}

// BracketGenerationCompleted event fired when bracket is successfully generated
type BracketGenerationCompleted struct {
	BaseEvent
}

func NewBracketGenerationCompleted(tournamentID, bracketID uuid.UUID, totalRounds, totalMatches int, metadata map[string]string) BracketGenerationCompleted {
	payload := BracketGenerationCompletedPayload{
		TournamentID: tournamentID,
		BracketID:    bracketID,
		TotalRounds:  totalRounds,
		TotalMatches: totalMatches,
		CompletedAt:  time.Now(),
	}
	return BracketGenerationCompleted{
		BaseEvent: newBaseEvent("BracketGenerationCompleted", tournamentID.String(), "Bracket", payload, metadata),
	}
}

// BracketGenerationFailedPayload contains bracket generation failure details
type BracketGenerationFailedPayload struct {
	TournamentID uuid.UUID `json:"tournament_id"`
	Error        string    `json:"error"`
	Reason       string    `json:"reason"`
	FailedAt     time.Time `json:"failed_at"`
}

// BracketGenerationFailed event fired when bracket generation fails
// This triggers compensating actions (status rollback in tournament service)
type BracketGenerationFailed struct {
	BaseEvent
}

func NewBracketGenerationFailed(tournamentID uuid.UUID, err error, reason string, metadata map[string]string) BracketGenerationFailed {
	payload := BracketGenerationFailedPayload{
		TournamentID: tournamentID,
		Error:        err.Error(),
		Reason:       reason,
		FailedAt:     time.Now(),
	}
	return BracketGenerationFailed{
		BaseEvent: newBaseEvent("BracketGenerationFailed", tournamentID.String(), "Bracket", payload, metadata),
	}
}

// ============================================================================
// Post-Match Navigation Events (WebSocket Messages)
// ============================================================================

// WebSocket message types for post-match navigation
const (
	// MsgTypeReturnToPreLobby - sent to winner, instructs them to go to pre-lobby for next round
	MsgTypeReturnToPreLobby = "return_to_prelobby"
	// MsgTypePlayerEliminatedNotify - sent to loser, informs them they are eliminated
	MsgTypePlayerEliminatedNotify = "player_eliminated_notify"
	// MsgTypeTournamentChampion - sent to tournament winner
	MsgTypeTournamentChampion = "tournament_champion"
)

// ReturnToPreLobbyPayload - sent to winner after match completion
// Instructs them to navigate to pre-lobby for the next round
type ReturnToPreLobbyPayload struct {
	TournamentID uuid.UUID `json:"tournament_id"`
	MatchID      uuid.UUID `json:"match_id"`
	NextRoundNum int       `json:"next_round_number"`
	Message      string    `json:"message"`
	PreLobbyURL  string    `json:"prelobby_url"`
}

// PlayerEliminatedNotificationPayload - sent to loser after match completion
// Informs them they have been eliminated from the tournament
type PlayerEliminatedNotificationPayload struct {
	TournamentID  uuid.UUID `json:"tournament_id"`
	MatchID       uuid.UUID `json:"match_id"`
	FinalPosition int       `json:"final_position"` // e.g., 5-8th place
	Message       string    `json:"message"`
}

// TournamentChampionPayload - sent to the tournament winner
// Congratulates them on winning the tournament
type TournamentChampionPayload struct {
	TournamentID uuid.UUID `json:"tournament_id"`
	ChampionID   uuid.UUID `json:"champion_id"`
	PrizeSummary string    `json:"prize_summary,omitempty"`
	Message      string    `json:"message"`
}
