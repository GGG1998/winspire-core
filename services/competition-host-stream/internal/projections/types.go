package projections

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

// CupHostView models the JSON payload we persist for host dashboards.
type CupHostView struct {
	CupID                uuid.UUID       `json:"cupId"`
	CompetitionContextID uuid.UUID       `json:"competitionContextId"`
	StageStatuses        json.RawMessage `json:"stageStatuses"`    // array of {stageId,status,...}
	AttendanceCounts     json.RawMessage `json:"attendanceCounts"` // {total,confirmed,waitlisted}
	DependencyHealth     json.RawMessage `json:"dependencyHealth"` // array of supplier states
}

// TournamentHostView captures downstream tournament + lineup state.
type TournamentHostView struct {
	TournamentID  uuid.UUID       `json:"tournamentId"`
	CupID         *uuid.UUID      `json:"cupId,omitempty"`
	SettingsHash  string          `json:"settingsHash"`
	LineupStatus  json.RawMessage `json:"lineupStatus"`
	SeedingWindow string          `json:"seedingWindow"` // textual (e.g., "[2025-01-01,2025-01-02]")
	MatchGate     json.RawMessage `json:"matchGate"`
}

// AttendanceSnapshot keeps running totals per cup/tournament.
type AttendanceSnapshot struct {
	ScopeType            string          `json:"scopeType"` // cup | tournament
	ScopeID              uuid.UUID       `json:"scopeId"`
	TotalJoined          int             `json:"totalJoined"`
	TotalConfirmed       int             `json:"totalConfirmed"`
	RestrictionsBreached json.RawMessage `json:"restrictionsBreached"`
	LastEventID          int64           `json:"lastEventId"`
}

// MatchLobbyView joins queue + lobby detail for operators.
type MatchLobbyView struct {
	MatchID          uuid.UUID       `json:"matchId"`
	TournamentID     uuid.UUID       `json:"tournamentId"`
	LobbyInformation json.RawMessage `json:"lobbyInformation"`
	QueueState       json.RawMessage `json:"queueState"`
}

func validateJSON(raw json.RawMessage) error {
	if len(raw) == 0 {
		return errors.New("empty json payload")
	}
	if !json.Valid(raw) {
		return errors.New("invalid json payload")
	}
	return nil
}
