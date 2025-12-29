// Package http provides HTTP handlers for the matchmaking service
package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/application"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/repository"
	"github.com/winspire-core/services/matchmaking/internal/websocket"
	"github.com/winspire/winspire-core/libs/go/httpx"
)

// matchService defines the subset of match service methods required by HTTP handlers
type matchService interface {
	OnPlayerReady(ctx context.Context, matchID, userID uuid.UUID) error
	OnGameLoaded(ctx context.Context, matchID, userID uuid.UUID) error
	GrantWalkover(ctx context.Context, matchID, userID, noShowPlayerID uuid.UUID, reason string) error
	GetCurrentMatchForUser(ctx context.Context, userID uuid.UUID) (*domain.Match, *domain.Round, uuid.UUID, error)
	CompleteMatch(ctx context.Context, matchID, winnerID uuid.UUID, scorePlayer1, scorePlayer2 int, source domain.ResultSource) error
	// TODO AdvanceWinner(ctx context.Context, winnerID, fromMatchID, toMatchID uuid.UUID) error
	// TODO HandleByeMatch(ctx context.Context, matchID uuid.UUID) error
}

// competitionClientInterface defines methods needed from CompetitionClient (tournament service client)
type competitionClientInterface interface {
	GetTournamentInfo(ctx context.Context, tournamentID uuid.UUID) (*application.TournamentInfo, error)
}

// MatchHandler handles match-related HTTP requests
type MatchHandler struct {
	matchRepo         repository.MatchRepository
	roundRepo         repository.RoundRepository
	bracketRepo       repository.BracketRepository
	preLobbyService   *application.PreLobbyService
	matchService      matchService
	hub               *websocket.Hub
	competitionClient competitionClientInterface
}

// NewMatchHandler creates a new match handler
func NewMatchHandler(
	matchRepo repository.MatchRepository,
	roundRepo repository.RoundRepository,
	bracketRepo repository.BracketRepository,
	preLobbyService *application.PreLobbyService,
	matchService matchService,
	hub *websocket.Hub,
	competitionClient competitionClientInterface,
) *MatchHandler {
	return &MatchHandler{
		matchRepo:         matchRepo,
		roundRepo:         roundRepo,
		bracketRepo:       bracketRepo,
		preLobbyService:   preLobbyService,
		matchService:      matchService,
		hub:               hub,
		competitionClient: competitionClient,
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

	// Fetch round to get tournament ID and round number
	round, err := h.roundRepo.GetByID(c.Request.Context(), match.RoundID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch round", "details": err.Error()})
		return
	}

	// Fetch bracket to get tournament info
	bracket, err := h.bracketRepo.GetByID(c.Request.Context(), round.BracketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bracket", "details": err.Error()})
		return
	}

	// Fetch tournament info from tournament service
	var tournamentName string
	var hostID uuid.UUID
	if h.competitionClient != nil {
		tournamentInfo, err := h.competitionClient.GetTournamentInfo(c.Request.Context(), bracket.TournamentID)
		if err != nil {
			log.Printf("WARN: Failed to fetch tournament info: %v", err)
			// Don't fail the request, continue with empty tournament info
		} else if tournamentInfo != nil {
			tournamentName = tournamentInfo.Name
			hostID = tournamentInfo.HostID
		}
	}

	// Get participant details from pre-lobby service
	participant1Details := h.preLobbyService.GetParticipantDetails(c.Request.Context(), bracket.TournamentID, match.Participant1ID)
	var participant2Details *application.PreLobbyParticipantInfo
	if match.Participant2ID != nil {
		details := h.preLobbyService.GetParticipantDetails(c.Request.Context(), bracket.TournamentID, *match.Participant2ID)
		participant2Details = &details
	}

	// Build match data
	matchData := map[string]interface{}{
		"id":                     match.ID,
		"match_number":           match.MatchNumber,
		"participant1_id":        match.Participant1ID,
		"participant2_id":        match.Participant2ID,
		"next_match_id":          match.NextMatchID,
		"status":                 match.Status,
		"participant1_ready":     match.Participant1Ready,
		"participant2_ready":     match.Participant2Ready,
		"winner_id":              match.WinnerID,
		"score_player1":          match.ScorePlayer1,
		"score_player2":          match.ScorePlayer2,
		"game_api_match_id":      match.GameAPIMatchID,
		"disconnected_player_id": match.DisconnectedPlayerID,
		"disconnected_at":        match.DisconnectedAt,
		"started_at":             match.StartedAt,
		"completed_at":           match.CompletedAt,
	}

	// Build participant1 data
	participant1Data := map[string]interface{}{
		"id":           participant1Details.UserID,
		"display_name": participant1Details.DisplayName,
		"avatar_url":   participant1Details.AvatarURL,
	}

	// Build participant2 data (can be null)
	var participant2Data interface{} = nil
	if participant2Details != nil {
		participant2Data = map[string]interface{}{
			"id":           participant2Details.UserID,
			"display_name": participant2Details.DisplayName,
			"avatar_url":   participant2Details.AvatarURL,
		}
	}

	// Build tournament info (for display in lobby without separate API call)
	tournamentInfo := map[string]interface{}{
		"id":   bracket.TournamentID,
		"name": tournamentName,
	}
	if hostID != uuid.Nil {
		tournamentInfo["host_id"] = hostID
	}

	// Build response with nested structure
	response := map[string]interface{}{
		"match":           matchData,
		"participant1":    participant1Data,
		"participant2":    participant2Data,
		"round_number":    round.RoundNumber,
		"game_snapshot":   bracket.GameSnapshot,
		"tournament_info": tournamentInfo,
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

	// Broadcast ready status via WebSocket (T059)
	if h.hub != nil {
		msg, err := websocket.NewMessage(websocket.MessageType("ready_updated"), map[string]interface{}{
			"playerId": userID,
			"ready":    true,
		})
		if err == nil {
			h.hub.BroadcastToMatch(matchID, msg)
		} else {
			log.Printf("WARN: Failed to create ready_updated message: %v", err)
		}
	}

	// Trigger ready state check and potential match start (T072, T073)
	if h.matchService != nil {
		err = h.matchService.OnPlayerReady(c.Request.Context(), matchID, userID)
		if err != nil {
			// Log error but don't fail the request - ready status is already saved
			log.Printf("WARN: Failed to process ready state: %v", err)
			// Don't return error - ready status was successfully saved and broadcast
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "ready status updated successfully",
		"match_id":  matchID,
		"player_id": userID,
		"ready":     true,
	})
}

// MarkGameLoaded marks that the player's game has loaded successfully
// POST /v1/matches/:id/game-loaded
func (h *MatchHandler) MarkGameLoaded(c *gin.Context) {
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
	// #region agent log
	func() {
		f, _ := os.OpenFile("/Users/gabrieldomanowski/programming/winspire-core/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			defer f.Close()
			json.NewEncoder(f).Encode(map[string]interface{}{"location": "match_handler.go:242", "message": "match fetched", "data": map[string]interface{}{"matchId": matchID.String(), "status": string(match.Status), "p1Loaded": match.Participant1GameLoaded, "p2Loaded": match.Participant2GameLoaded}, "timestamp": time.Now().UnixMilli(), "sessionId": "debug-session", "hypothesisId": "H3"})
		}
	}()
	// #endregion

	// Verify user is a participant
	isParticipant1 := match.Participant1ID == userID
	isParticipant2 := match.Participant2ID != nil && *match.Participant2ID == userID

	if !isParticipant1 && !isParticipant2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied: you are not a participant in this match"})
		return
	}

	// Call match service to handle game loaded
	err = h.matchService.OnGameLoaded(c.Request.Context(), matchID, userID)
	if err != nil {
		log.Printf("WARN: Failed to process game loaded: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark game loaded", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "game loaded successfully",
		"match_id":  matchID,
		"player_id": userID,
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

// CompleteMatchRequest represents the request body for completing a match
type CompleteMatchRequest struct {
	WinnerID     string `json:"winner_id" binding:"required"`
	LoserID      string `json:"loser_id" binding:"required"`
	ScoreWinner  int    `json:"score_winner"`
	ScoreLoser   int    `json:"score_loser"`
	ResultSource string `json:"result_source" binding:"required,oneof=game_api manual_host walkover"`
}

// CompleteMatch marks a match as completed with results
// POST /v1/matches/:id/complete
func (h *MatchHandler) CompleteMatch(c *gin.Context) {
	// Parse match ID
	matchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID", "details": err.Error()})
		return
	}

	// Parse request body
	var req CompleteMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	// Parse winner ID
	winnerID, err := uuid.Parse(req.WinnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid winner_id", "details": err.Error()})
		return
	}

	// Determine result source
	var source domain.ResultSource
	switch req.ResultSource {
	case "game_api":
		source = domain.ResultSourceGameAPI
	case "manual_host":
		source = domain.ResultSourceHostManual
	case "walkover":
		source = domain.ResultSourceWalkover
	default:
		source = domain.ResultSourceHostManual
	}

	// Fetch match to determine scores (winner is participant1 or participant2)
	match, err := h.matchRepo.GetByID(c.Request.Context(), matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found", "details": err.Error()})
		return
	}

	// Determine score order (score_player1, score_player2) based on who won
	var scorePlayer1, scorePlayer2 int
	if match.Participant1ID == winnerID {
		scorePlayer1 = req.ScoreWinner
		scorePlayer2 = req.ScoreLoser
	} else {
		scorePlayer1 = req.ScoreLoser
		scorePlayer2 = req.ScoreWinner
	}

	// Complete the match
	err = h.matchService.CompleteMatch(c.Request.Context(), matchID, winnerID, scorePlayer1, scorePlayer2, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete match", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "match completed successfully",
		"match_id":      matchID,
		"winner_id":     winnerID,
		"score_player1": scorePlayer1,
		"score_player2": scorePlayer2,
		"result_source": req.ResultSource,
	})
}

// GetMatchesForTournament retrieves all matches for a tournament with participant details
// GET /v1/tournaments/:id/matches
func (h *MatchHandler) GetMatchesForTournament(c *gin.Context) {
	// Parse tournament ID
	tournamentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tournament ID", "details": err.Error()})
		return
	}

	// Get bracket with rounds and matches
	bracket, rounds, matches, err := h.bracketRepo.GetWithRoundsAndMatches(c.Request.Context(), tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bracket not found for tournament", "details": err.Error()})
		return
	}

	// Create a map of round ID to round info for quick lookup
	roundMap := make(map[uuid.UUID]struct {
		RoundNumber int
		RoundName   string
	})
	for _, round := range rounds {
		roundMap[round.ID] = struct {
			RoundNumber int
			RoundName   string
		}{
			RoundNumber: round.RoundNumber,
			RoundName:   round.RoundName,
		}
	}

	// Build matches response with participant details
	matchesResponse := make([]interface{}, 0, len(matches))
	for _, match := range matches {
		// Get round info
		roundInfo := roundMap[match.RoundID]

		// Get participant details from pre-lobby service
		participant1Details := h.preLobbyService.GetParticipantDetails(c.Request.Context(), bracket.TournamentID, match.Participant1ID)

		var participant2Details *application.PreLobbyParticipantInfo
		if match.Participant2ID != nil {
			details := h.preLobbyService.GetParticipantDetails(c.Request.Context(), bracket.TournamentID, *match.Participant2ID)
			participant2Details = &details
		}

		// Build match data
		matchData := map[string]interface{}{
			"id":                 match.ID,
			"match_number":       match.MatchNumber,
			"participant1_id":    match.Participant1ID,
			"participant2_id":    match.Participant2ID,
			"next_match_id":      match.NextMatchID,
			"status":             match.Status,
			"participant1_ready": match.Participant1Ready,
			"participant2_ready": match.Participant2Ready,
			"winner_id":          match.WinnerID,
			"score_player1":      match.ScorePlayer1,
			"score_player2":      match.ScorePlayer2,
			"started_at":         match.StartedAt,
			"completed_at":       match.CompletedAt,
			"round_id":           match.RoundID,
		}

		// Build participant1 data
		participant1Data := map[string]interface{}{
			"id":           participant1Details.UserID,
			"display_name": participant1Details.DisplayName,
			"avatar_url":   participant1Details.AvatarURL,
		}

		// Build participant2 data (can be null)
		var participant2Data interface{} = nil
		if participant2Details != nil {
			participant2Data = map[string]interface{}{
				"id":           participant2Details.UserID,
				"display_name": participant2Details.DisplayName,
				"avatar_url":   participant2Details.AvatarURL,
			}
		}

		// Build complete match response
		matchResponse := map[string]interface{}{
			"match":        matchData,
			"participant1": participant1Data,
			"participant2": participant2Data,
			"roundNumber":  roundInfo.RoundNumber,
			"roundName":    roundInfo.RoundName,
		}

		matchesResponse = append(matchesResponse, matchResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": matchesResponse,
	})
}
