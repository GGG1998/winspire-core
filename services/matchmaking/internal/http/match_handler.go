// Package http provides HTTP handlers for the matchmaking service
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/application"
	"github.com/winspire-core/services/matchmaking/internal/repository"
	"github.com/winspire/winspire-core/libs/go/httpx"
)

// MatchHandler handles match-related HTTP requests
type MatchHandler struct {
	matchRepo    repository.MatchRepository
	matchService *application.MatchService
}

// NewMatchHandler creates a new match handler
func NewMatchHandler(matchRepo repository.MatchRepository, matchService *application.MatchService) *MatchHandler {
	return &MatchHandler{
		matchRepo:    matchRepo,
		matchService: matchService,
	}
}

// GetMatch retrieves a single match by ID with participant details
// GET /v1/matches/:id
func (h *MatchHandler) GetMatch(c *gin.Context) {
	// Parse match ID
	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID", "details": err.Error()})
		return
	}

	// Fetch match
	match, err := h.matchRepo.GetByID(c.Request.Context(), matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found", "details": err.Error()})
		return
	}

	// Build response with match details
	response := map[string]interface{}{
		"id":                     match.ID,
		"round_id":               match.RoundID,
		"match_number":           match.MatchNumber,
		"next_match_id":          match.NextMatchID,
		"participant1_id":        match.Participant1ID,
		"participant2_id":        match.Participant2ID,
		"status":                 match.Status,
		"participant1_ready":     match.Participant1Ready,
		"participant2_ready":     match.Participant2Ready,
		"winner_id":              match.WinnerID,
		"score_player1":          match.ScorePlayer1,
		"score_player2":          match.ScorePlayer2,
		"result_source":          match.ResultSource,
		"disconnected_player_id": match.DisconnectedPlayerID,
		"disconnected_at":        match.DisconnectedAt,
		"game_api_match_id":      match.GameAPIMatchID,
		"created_at":             match.CreatedAt,
		"started_at":             match.StartedAt,
		"completed_at":           match.CompletedAt,
		"updated_at":             match.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// MarkPlayerReady marks a player as ready for their match
// POST /v1/matches/:id/ready
func (h *MatchHandler) MarkPlayerReady(c *gin.Context) {
	// Parse match ID
	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID", "details": err.Error()})
		return
	}

	// Get authenticated user from JWT (set by auth middleware)
	user := httpx.MustGetUser(c)
	userID, err := uuid.Parse(string(user.ID))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID", "details": err.Error()})
		return
	}

	// Fetch match to verify user is a participant
	match, err := h.matchRepo.GetByID(c.Request.Context(), matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found", "details": err.Error()})
		return
	}

	// Verify user is a participant (FR-023: JWT authentication required)
	isParticipant1 := match.Participant1ID == userID
	isParticipant2 := match.Participant2ID != nil && *match.Participant2ID == userID

	if !isParticipant1 && !isParticipant2 {
		// FR-024: Deny access if user not in match participants
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied: you are not a participant in this match"})
		return
	}

	// Update ready status (FR-019: Server-side persistence)
	err = h.matchRepo.UpdateReady(c.Request.Context(), matchID, userID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update ready status", "details": err.Error()})
		return
	}

	// Trigger ready state check and potential match start (T072, T073)
	if h.matchService != nil {
		err = h.matchService.OnPlayerReady(c.Request.Context(), matchID, userID)
		if err != nil {
			// Log error but don't fail the request - ready status is already saved
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process ready state", "details": err.Error()})
			return
		}
	}

	// TODO: Broadcast ready status via WebSocket (T059)
	// This will be handled by the WebSocket hub when it's implemented

	c.JSON(http.StatusOK, gin.H{
		"message":   "ready status updated successfully",
		"match_id":  matchID,
		"player_id": userID,
		"ready":     true,
	})
}

// ClaimWalkover allows a player to claim a walkover when opponent doesn't show (T086)
// POST /v1/matches/:id/claim-walkover
func (h *MatchHandler) ClaimWalkover(c *gin.Context) {
	// Parse match ID
	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID", "details": err.Error()})
		return
	}

	// Get authenticated user
	user := httpx.MustGetUser(c)
	userID, err := uuid.Parse(string(user.ID))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID", "details": err.Error()})
		return
	}

	// Fetch match to verify user is a participant
	match, err := h.matchRepo.GetByID(c.Request.Context(), matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found", "details": err.Error()})
		return
	}

	// Verify user is a participant
	isParticipant1 := match.Participant1ID == userID
	isParticipant2 := match.Participant2ID != nil && *match.Participant2ID == userID

	if !isParticipant1 && !isParticipant2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied: you are not a participant in this match"})
		return
	}

	// Determine the no-show player (opponent)
	var noShowPlayerID uuid.UUID
	if isParticipant1 && match.Participant2ID != nil {
		noShowPlayerID = *match.Participant2ID
	} else if isParticipant2 {
		noShowPlayerID = match.Participant1ID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot claim walkover: no opponent assigned"})
		return
	}

	// Grant walkover (FR-016: 2-minute timeout rule enforced)
	err = h.matchService.GrantWalkover(c.Request.Context(), matchID, userID, noShowPlayerID, "opponent no-show")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to grant walkover", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "walkover granted successfully",
		"match_id":       matchID,
		"winner_id":      userID,
		"no_show_player": noShowPlayerID,
	})
}
