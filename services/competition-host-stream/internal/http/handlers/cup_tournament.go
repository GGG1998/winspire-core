package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/winspire/competition-host-stream/internal/projections"
	ssebroker "github.com/winspire/competition-host-stream/internal/sse"
)

// CupTournamentDeps bundles dependencies for cup/tournament routes.
type CupTournamentDeps struct {
	Reader              *projections.Reader
	CupProjector        *projections.CupProjector
	TournamentProjector *projections.TournamentProjector
	EventRouter         *ssebroker.EventRouter
}

// RegisterCupTournamentRoutes wires GET/PATCH handlers for cup + tournament projections.
func RegisterCupTournamentRoutes(group *gin.RouterGroup, deps CupTournamentDeps) {
	group.GET("/hosts/:hostId/cups/:cupId", func(c *gin.Context) {
		if _, ok := parseUUIDParam(c, "hostId"); !ok {
			return
		}
		cupID, ok := parseUUIDParam(c, "cupId")
		if !ok {
			return
		}
		payload, err := deps.Reader.GetCup(c.Request.Context(), cupID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, payload)
	})

	group.PATCH("/hosts/:hostId/cups/:cupId", func(c *gin.Context) {
		hostID, ok := parseUUIDParam(c, "hostId")
		if !ok {
			return
		}
		cupID, ok := parseUUIDParam(c, "cupId")
		if !ok {
			return
		}

		var req CupPatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			return
		}

		stagePayload, ok := marshalOrError(c, req.StageStatuses, []byte("[]"))
		if !ok {
			return
		}
		dependencyPayload, ok := marshalOrError(c, req.DependencyOverrides, []byte("[]"))
		if !ok {
			return
		}

		attendanceBody := req.Attendance
		if attendanceBody == nil {
			attendanceBody = &CupAttendance{}
		}
		withTargets := map[string]any{
			"total":      attendanceBody.Total,
			"confirmed":  attendanceBody.Confirmed,
			"waitlisted": attendanceBody.Waitlisted,
		}
		if len(req.AttendanceTargets) > 0 {
			withTargets["targets"] = req.AttendanceTargets
		}
		attendancePayload, ok := marshalOrError(c, withTargets, []byte(`{"total":0,"confirmed":0,"waitlisted":0}`))
		if !ok {
			return
		}

		view := projections.CupHostView{
			CupID:                cupID,
			CompetitionContextID: req.CompetitionContextID,
			StageStatuses:        stagePayload,
			AttendanceCounts:     attendancePayload,
			DependencyHealth:     dependencyPayload,
		}

		if err := deps.CupProjector.Upsert(c.Request.Context(), view); err != nil {
			c.Error(err)
			return
		}

		deps.EventRouter.CupUpdated(c.Request.Context(), cupID, map[string]any{
			"hostId": hostID,
			"cup":    view,
		})
		c.Status(http.StatusAccepted)
	})

	group.GET("/hosts/:hostId/tournaments/:tournamentId", func(c *gin.Context) {
		if _, ok := parseUUIDParam(c, "hostId"); !ok {
			return
		}
		tournamentID, ok := parseUUIDParam(c, "tournamentId")
		if !ok {
			return
		}

		payload, err := deps.Reader.GetTournament(c.Request.Context(), tournamentID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, payload)
	})

	group.PATCH("/hosts/:hostId/tournaments/:tournamentId", func(c *gin.Context) {
		hostID, ok := parseUUIDParam(c, "hostId")
		if !ok {
			return
		}
		tournamentID, ok := parseUUIDParam(c, "tournamentId")
		if !ok {
			return
		}

		var req TournamentPatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			return
		}

		lineup := inflateLineups(req.LineupStatus)
		applyLineupDirectives(lineup, req.LineupDirectives)
		applyAllowedActionOverrides(lineup, req.AllowedActionsOverrides)

		lineupPayload, ok := marshalOrError(c, lineupSlice(lineup), []byte("[]"))
		if !ok {
			return
		}
		if len(req.AllowedActionsOverrides) > 0 {
			overrides := make([]projections.AllowedActionOverride, 0, len(req.AllowedActionsOverrides))
			for _, override := range req.AllowedActionsOverrides {
				overrides = append(overrides, projections.AllowedActionOverride{
					LineupID: override.LineupID,
					Action:   override.Action,
					Enabled:  override.Enabled,
				})
			}
			var err error
			lineupPayload, err = projections.ApplyAllowedActionOverrides(lineupPayload, overrides)
			if err != nil {
				c.Error(err)
				return
			}
		}
		matchGatePayload, ok := marshalOrError(c, req.MatchGate, []byte(`{"readyForMatch":false}`))
		if !ok {
			return
		}

		view := projections.TournamentHostView{
			TournamentID:  tournamentID,
			CupID:         req.CupID,
			SettingsHash:  req.SettingsHash,
			LineupStatus:  lineupPayload,
			SeedingWindow: encodeSeedingWindow(req.SeedingWindow),
			MatchGate:     matchGatePayload,
		}

		if err := deps.TournamentProjector.Upsert(c.Request.Context(), view); err != nil {
			c.Error(err)
			return
		}

		deps.EventRouter.TournamentParticipationUpdate(c.Request.Context(), tournamentID, map[string]any{
			"hostId":     hostID,
			"tournament": view,
		})
		c.Status(http.StatusAccepted)
	})
}

type lineupMap map[uuid.UUID]*LineupStatusEntry

func inflateLineups(entries []LineupStatusEntry) lineupMap {
	out := make(lineupMap)
	for _, entry := range entries {
		item := entry
		item.AllowedActions = normalizeAllowedActions(item.AllowedActions)
		out[entry.LineupID] = &item
	}
	return out
}

func lineupSlice(entries lineupMap) []LineupStatusEntry {
	result := make([]LineupStatusEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, *entry)
	}
	return result
}

func applyLineupDirectives(entries lineupMap, directives []LineupDirective) {
	for _, directive := range directives {
		switch directive.Action {
		case "FORCE_WITHDRAW":
			delete(entries, directive.LineupID)
			continue
		case "CONFIRM", "REVOKE":
			entry := ensureLineupEntry(entries, directive.LineupID)
			entry.Confirmed = directive.Action == "CONFIRM"
		}
	}
}

func ensureLineupEntry(entries lineupMap, id uuid.UUID) *LineupStatusEntry {
	entry, ok := entries[id]
	if !ok {
		entry = &LineupStatusEntry{
			LineupID:       id,
			AllowedActions: []string{},
		}
		entries[id] = entry
	}
	return entry
}

func applyAllowedActionOverrides(entries lineupMap, overrides []AllowedActionOverride) {
	for _, override := range overrides {
		entry := ensureLineupEntry(entries, override.LineupID)
		if override.Enabled {
			// Add action if not already present
			found := false
			for _, action := range entry.AllowedActions {
				if action == override.Action {
					found = true
					break
				}
			}
			if !found {
				entry.AllowedActions = append(entry.AllowedActions, override.Action)
			}
		} else {
			// Remove action if present
			actions := make([]string, 0, len(entry.AllowedActions))
			for _, action := range entry.AllowedActions {
				if action != override.Action {
					actions = append(actions, action)
				}
			}
			entry.AllowedActions = actions
		}
	}
}
