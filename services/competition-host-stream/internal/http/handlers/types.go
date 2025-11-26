package handlers

import (
	"time"

	"github.com/google/uuid"
)

// StageStatus mirrors the OpenAPI schema for cup stage metadata.
type StageStatus struct {
	StageID      uuid.UUID `json:"stageId" binding:"required"`
	Status       string    `json:"status" binding:"required"`
	PlannedStart time.Time `json:"plannedStartAt"`
	PlannedEnd   time.Time `json:"plannedEndAt"`
}

// DependencyHealth mirrors supplier readiness overrides.
type DependencyHealth struct {
	Supplier string  `json:"supplier" binding:"required"`
	State    string  `json:"state" binding:"required"`
	Notes    *string `json:"notes"`
}

// CupAttendance represents host-facing attendance counts.
type CupAttendance struct {
	Total      int `json:"total"`
	Confirmed  int `json:"confirmed"`
	Waitlisted int `json:"waitlisted"`
}

// AttendanceTarget lets hosts document per-stage thresholds.
type AttendanceTarget struct {
	StageID          uuid.UUID `json:"stageId" binding:"required"`
	MinimumConfirmed int       `json:"minimumConfirmed" binding:"required"`
}

// CupPatchRequest defines PATCH payload for cup projections.
type CupPatchRequest struct {
	CompetitionContextID uuid.UUID          `json:"competitionContextId" binding:"required"`
	StageStatuses        []StageStatus      `json:"stageStatuses" binding:"required,dive"`
	DependencyOverrides  []DependencyHealth `json:"dependencyOverrides"`
	Attendance           *CupAttendance     `json:"attendance"`
	AttendanceTargets    []AttendanceTarget `json:"attendanceTargets"`
}

// LineupStatusEntry mirrors TournamentHostView lineup entries.
type LineupStatusEntry struct {
	LineupID       uuid.UUID `json:"lineupId" binding:"required"`
	Confirmed      bool      `json:"confirmed"`
	AllowedActions []string  `json:"allowedActions"`
}

// SeedingWindow describes window overrides.
type SeedingWindow struct {
	OpensAt  time.Time `json:"opensAt" binding:"required"`
	ClosesAt time.Time `json:"closesAt" binding:"required"`
}

// MatchGate describes tournament match readiness state.
type MatchGate struct {
	ReadyForMatch bool    `json:"readyForMatch"`
	BlockedReason *string `json:"blockedReason"`
}

// LineupDirective actions supported by Participation.
type LineupDirective struct {
	LineupID uuid.UUID `json:"lineupId" binding:"required"`
	Action   string    `json:"action" binding:"required,oneof=CONFIRM REVOKE FORCE_WITHDRAW"`
	Note     string    `json:"note"`
}

// AllowedActionOverride toggles a specific allowed action.
type AllowedActionOverride struct {
	LineupID uuid.UUID `json:"lineupId" binding:"required"`
	Action   string    `json:"action" binding:"required"`
	Enabled  bool      `json:"enabled"`
}

// TournamentPatchRequest models PATCH payload for tournaments.
type TournamentPatchRequest struct {
	CupID                   *uuid.UUID              `json:"cupId"`
	SettingsHash            string                  `json:"settingsHash" binding:"required"`
	LineupStatus            []LineupStatusEntry     `json:"lineupStatus"`
	SeedingWindow           *SeedingWindow          `json:"seedingWindow"`
	MatchGate               *MatchGate              `json:"matchGate"`
	LineupDirectives        []LineupDirective       `json:"lineupDirectives"`
	AllowedActionsOverrides []AllowedActionOverride `json:"allowedActionsOverrides"`
}

