// Package http provides HTTP handlers for the matchmaking service
package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gorillaws "github.com/gorilla/websocket"
	"github.com/winspire-core/services/matchmaking/internal/application"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/observability"
	"github.com/winspire-core/services/matchmaking/internal/websocket"
	authtypes "github.com/winspire/winspire-core/libs/go/auth/types"
)

// ============================================================================
// Response DTOs
// ============================================================================

// PreLobbyStateResponse is the REST API response for pre-lobby state
type PreLobbyStateResponse struct {
	TournamentID     string                        `json:"tournament_id"`
	TournamentName   string                        `json:"tournament_name"`
	Status           string                        `json:"status"`
	Participants     []PreLobbyParticipantResponse `json:"participants"`
	ParticipantCount int                           `json:"participant_count"`
	MinParticipants  int                           `json:"min_participants"`
	StartTime        time.Time                     `json:"start_time"`
	GracePeriod      *GracePeriodResponse          `json:"grace_period,omitempty"`
	ActivityFeed     []ActivityFeedItemResponse    `json:"activity_feed"`
}

// PreLobbyParticipantResponse represents a participant in the response
type PreLobbyParticipantResponse struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	JoinedAt    time.Time `json:"joined_at"`
}

// GracePeriodResponse represents grace period timing in the response
type GracePeriodResponse struct {
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	RemainingSeconds int       `json:"remaining_seconds"`
}

// ActivityFeedItemResponse represents an activity feed item in the response
type ActivityFeedItemResponse struct {
	ID              string    `json:"id"`
	EventType       string    `json:"event_type"`
	Message         string    `json:"message"`
	ParticipantName *string   `json:"participant_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// ============================================================================
// Handler
// ============================================================================

// PreLobbyHandler handles pre-lobby HTTP requests
type PreLobbyHandler struct {
	preLobbyService   *application.PreLobbyService
	competitionClient CompetitionClient
	logger            *observability.Logger
}

// CompetitionClient interface for fetching tournament info from competition service
type CompetitionClient interface {
	GetTournamentInfo(ctx context.Context, tournamentID uuid.UUID) (*application.TournamentInfo, error)
	IsUserRegistered(ctx context.Context, tournamentID, userID uuid.UUID) (bool, error)
}

// NewPreLobbyHandler creates a new pre-lobby handler
func NewPreLobbyHandler(
	preLobbyService *application.PreLobbyService,
	competitionClient CompetitionClient,
	logger *observability.Logger,
) *PreLobbyHandler {
	return &PreLobbyHandler{
		preLobbyService:   preLobbyService,
		competitionClient: competitionClient,
		logger:            logger,
	}
}

// GetPreLobbyState returns the current state of a tournament pre-lobby
// GET /v1/tournaments/:id/lobby
func (h *PreLobbyHandler) GetPreLobbyState(c *gin.Context) {
	// Parse tournament ID
	tournamentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_tournament_id",
			"message": "Tournament ID must be a valid UUID",
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	user, ok := authtypes.UserFromContext(c.Request.Context())
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User authentication required",
		})
		return
	}
	userID, err := uuid.Parse(string(user.ID))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_user_id",
			"message": "Invalid user ID format",
		})
		return
	}

	// Fetch tournament info from competition service
	tournamentInfo, err := h.competitionClient.GetTournamentInfo(c.Request.Context(), tournamentID)
	if err != nil {
		h.logger.Warn("failed to get tournament info", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "competition_service_unavailable",
			"message": "Unable to fetch tournament information",
		})
		return
	}

	if tournamentInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "tournament_not_found",
			"message": "Tournament does not exist",
		})
		return
	}

	// Validate tournament status
	if !isValidTournamentStatus(tournamentInfo.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_tournament_status",
			"message": "Tournament is not in a valid state for pre-lobby access",
			"status":  tournamentInfo.Status,
		})
		return
	}

	// Check if user is registered for the tournament
	isRegistered, err := h.competitionClient.IsUserRegistered(c.Request.Context(), tournamentID, userID)
	if err != nil {
		h.logger.Warn("failed to check user registration", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"user_id":       userID.String(),
			"error":         err.Error(),
		})
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "competition_service_unavailable",
			"message": "Unable to verify registration status",
		})
		return
	}

	if !isRegistered {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "not_registered",
			"message": "User is not registered for this tournament",
		})
		return
	}

	// Get pre-lobby state
	state, err := h.preLobbyService.GetState(c.Request.Context(), tournamentID, tournamentInfo)
	if err != nil {
		h.logger.Error("failed to get pre-lobby state", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get pre-lobby state",
		})
		return
	}

	// Build response
	response := h.buildStateResponse(state)
	c.JSON(http.StatusOK, response)
}

// buildStateResponse converts internal state to API response
func (h *PreLobbyHandler) buildStateResponse(state *application.PreLobbyState) PreLobbyStateResponse {
	// Convert participants
	participants := make([]PreLobbyParticipantResponse, len(state.Participants))
	for i, p := range state.Participants {
		participants[i] = PreLobbyParticipantResponse{
			UserID:      p.UserID.String(),
			DisplayName: p.DisplayName,
			AvatarURL:   p.AvatarURL,
			JoinedAt:    p.JoinedAt,
		}
	}

	// Convert activity feed
	activityFeed := make([]ActivityFeedItemResponse, len(state.ActivityFeed))
	for i, item := range state.ActivityFeed {
		activityFeed[i] = ActivityFeedItemResponse{
			ID:              item.ID.String(),
			EventType:       string(item.EventType),
			Message:         item.Message,
			ParticipantName: item.ParticipantName,
			CreatedAt:       item.CreatedAt,
		}
	}

	response := PreLobbyStateResponse{
		TournamentID:     state.TournamentID.String(),
		TournamentName:   state.TournamentName,
		Status:           string(state.Status),
		Participants:     participants,
		ParticipantCount: state.ParticipantCount,
		MinParticipants:  state.MinParticipants,
		StartTime:        state.StartTime,
		ActivityFeed:     activityFeed,
	}

	// Add grace period if present
	if state.GracePeriod != nil {
		response.GracePeriod = &GracePeriodResponse{
			StartTime:        state.GracePeriod.StartTime,
			EndTime:          state.GracePeriod.EndTime,
			RemainingSeconds: state.GracePeriod.RemainingSeconds,
		}
	}

	return response
}

// isValidTournamentStatus checks if tournament status allows pre-lobby access
func isValidTournamentStatus(status string) bool {
	validStatuses := map[string]bool{
		"registration_open":   true,
		"registration_closed": true,
		"started":             true, // Allow during grace period
	}
	return validStatuses[status]
}

// ============================================================================
// WebSocket Types (for later use in T017)
// ============================================================================

// WebSocketMessage represents a WebSocket message structure
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// toWebSocketParticipant converts internal participant info to WebSocket payload
func toWebSocketParticipant(p *application.PreLobbyParticipantInfo) map[string]interface{} {
	result := map[string]interface{}{
		"user_id":      p.UserID.String(),
		"display_name": p.DisplayName,
		"joined_at":    p.JoinedAt,
	}
	if p.AvatarURL != nil {
		result["avatar_url"] = *p.AvatarURL
	}
	return result
}

// toWebSocketActivityFeed converts activity feed items to WebSocket payload
func toWebSocketActivityFeed(items []*domain.ActivityFeedItem) []map[string]interface{} {
	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		entry := map[string]interface{}{
			"id":         item.ID.String(),
			"event_type": string(item.EventType),
			"message":    item.Message,
			"created_at": item.CreatedAt,
		}
		if item.ParticipantName != nil {
			entry["participant_name"] = *item.ParticipantName
		}
		result[i] = entry
	}
	return result
}

// ============================================================================
// Pre-Lobby WebSocket Handler
// ============================================================================

// PreLobbyWebSocketHandler handles WebSocket connections for tournament pre-lobbies
type PreLobbyWebSocketHandler struct {
	preLobbyService   *application.PreLobbyService
	competitionClient CompetitionClient
	hub               *websocket.Hub
	logger            *observability.Logger
}

// NewPreLobbyWebSocketHandler creates a new pre-lobby WebSocket handler
func NewPreLobbyWebSocketHandler(
	preLobbyService *application.PreLobbyService,
	competitionClient CompetitionClient,
	hub *websocket.Hub,
	logger *observability.Logger,
) *PreLobbyWebSocketHandler {
	return &PreLobbyWebSocketHandler{
		preLobbyService:   preLobbyService,
		competitionClient: competitionClient,
		hub:               hub,
		logger:            logger,
	}
}

// Pre-lobby WebSocket upgrader
var preLobbyUpgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking based on environment
		return true // Allow all origins for now (will be restricted in production)
	},
}

// UpgradePreLobbyConnection upgrades HTTP connection to WebSocket for pre-lobby
// GET /v1/tournaments/:id/lobby/ws
func (h *PreLobbyWebSocketHandler) UpgradePreLobbyConnection(c *gin.Context) {
	// Parse tournament ID
	tournamentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_tournament_id",
			"message": "Tournament ID must be a valid UUID",
		})
		return
	}

	// Get user from context
	user, ok := authtypes.UserFromContext(c.Request.Context())
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User authentication required",
		})
		return
	}
	userID, err := uuid.Parse(string(user.ID))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_user_id",
			"message": "Invalid user ID format",
		})
		return
	}

	// Fetch tournament info
	tournamentInfo, err := h.competitionClient.GetTournamentInfo(c.Request.Context(), tournamentID)
	if err != nil {
		h.logger.Warn("failed to get tournament info for WebSocket", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "competition_service_unavailable",
			"message": "Unable to fetch tournament information",
		})
		return
	}

	if tournamentInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "tournament_not_found",
			"message": "Tournament does not exist",
		})
		return
	}

	// Validate tournament status
	if !isValidTournamentStatus(tournamentInfo.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_tournament_status",
			"message": "Tournament is not in a valid state for pre-lobby access",
			"status":  tournamentInfo.Status,
		})
		return
	}

	// Check if user is registered
	isRegistered, err := h.competitionClient.IsUserRegistered(c.Request.Context(), tournamentID, userID)
	if err != nil {
		h.logger.Warn("failed to check user registration for WebSocket", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"user_id":       userID.String(),
			"error":         err.Error(),
		})
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "competition_service_unavailable",
			"message": "Unable to verify registration status",
		})
		return
	}

	if !isRegistered {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "not_registered",
			"message": "User is not registered for this tournament",
		})
		return
	}

	// Check if pre-lobby can accept participants
	canAccept, err := h.preLobbyService.CanAcceptParticipants(c.Request.Context(), tournamentID)
	if err != nil {
		h.logger.Error("failed to check pre-lobby status", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to check pre-lobby status",
		})
		return
	}

	if !canAccept {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "prelobby_closed",
			"message": "Pre-lobby is no longer accepting participants",
		})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := preLobbyUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[PreLobby WebSocket] Failed to upgrade connection: %v", err)
		return
	}

	// Get display name from user
	displayName := user.Nickname
	if displayName == "" {
		displayName = string(user.ID)
	}

	// Create WebSocket client
	client := websocket.NewClient(h.hub, conn, userID, tournamentID)

	// Register participant in pre-lobby service
	participantInfo := &application.PreLobbyParticipantInfo{
		UserID:      userID,
		DisplayName: displayName,
		AvatarURL:   nil, // TODO: Get from user profile
		JoinedAt:    time.Now(),
	}
	h.preLobbyService.AddParticipant(tournamentID, participantInfo)

	// Register in hub for tournament room
	h.hub.RegisterTournamentClient(tournamentID, userID, displayName, nil, client)

	h.logger.Info("WebSocket client connected to pre-lobby", map[string]interface{}{
		"tournament_id": tournamentID.String(),
		"user_id":       userID.String(),
		"display_name":  displayName,
	})

	// Record activity feed event
	go h.preLobbyService.RecordParticipantJoined(context.Background(), tournamentID, displayName)

	// Send initial pre-lobby state
	// Use context.Background() instead of c.Request.Context() because:
	// - This goroutine lives independently of the HTTP handler
	// - c.Request.Context() is cancelled immediately after handler returns (after WebSocket upgrade)
	// - Would cause "context canceled" error in GetState call
	go h.sendInitialState(context.Background(), tournamentID, userID, tournamentInfo, client)

	// Broadcast participant joined to all connected clients
	go h.preLobbyService.BroadcastParticipantJoined(tournamentID, participantInfo)

	// Start client goroutines
	go client.ReadPump()
	go client.WritePump()
}

// sendInitialState sends the initial pre-lobby state to a newly connected client
func (h *PreLobbyWebSocketHandler) sendInitialState(ctx context.Context, tournamentID, userID uuid.UUID, tournamentInfo *application.TournamentInfo, client *websocket.Client) {
	// Diagnostic: check if context is already cancelled
	if ctx.Err() != nil {
		h.logger.Error("sendInitialState called with cancelled context", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"user_id":       userID.String(),
			"ctx_error":     ctx.Err().Error(),
		})
		return
	}

	state, err := h.preLobbyService.GetState(ctx, tournamentID, tournamentInfo)
	if err != nil {
		h.logger.Error("failed to get pre-lobby state for initial send", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"user_id":       userID.String(),
			"error":         err.Error(),
			"ctx_error":     ctx.Err(),
		})
		return
	}

	// Build WebSocket payload
	payload := websocket.PreLobbyStatePayload{
		TournamentID:     tournamentID,
		TournamentName:   state.TournamentName,
		Status:           string(state.Status),
		ParticipantCount: state.ParticipantCount,
		MinParticipants:  state.MinParticipants,
		StartTime:        state.StartTime,
	}

	// Convert participants
	participants := make([]websocket.PreLobbyParticipantPayload, len(state.Participants))
	for i, p := range state.Participants {
		participants[i] = websocket.PreLobbyParticipantPayload{
			UserID:      p.UserID,
			DisplayName: p.DisplayName,
			AvatarURL:   p.AvatarURL,
			JoinedAt:    p.JoinedAt,
		}
	}
	payload.Participants = participants

	// Add grace period info if present
	if state.GracePeriod != nil {
		payload.GracePeriod = &websocket.GracePeriodInfoPayload{
			StartTime:        state.GracePeriod.StartTime,
			EndTime:          state.GracePeriod.EndTime,
			RemainingSeconds: state.GracePeriod.RemainingSeconds,
		}
	}

	// Convert activity feed
	activityFeed := make([]websocket.ActivityFeedItemPayload, len(state.ActivityFeed))
	for i, item := range state.ActivityFeed {
		activityFeed[i] = websocket.ActivityFeedItemPayload{
			ID:              item.ID,
			EventType:       string(item.EventType),
			Message:         item.Message,
			ParticipantName: item.ParticipantName,
			CreatedAt:       item.CreatedAt,
		}
	}
	payload.ActivityFeed = activityFeed

	// Create and send message
	msg, err := websocket.NewMessage(websocket.MessageTypePreLobbyState, payload)
	if err != nil {
		h.logger.Error("failed to create pre-lobby state message", map[string]interface{}{
			"tournament_id": tournamentID.String(),
			"error":         err.Error(),
		})
		return
	}

	h.hub.SendToTournamentParticipant(tournamentID, userID, msg)
}

// HandleTournamentDisconnect handles participant disconnection from pre-lobby
func (h *PreLobbyWebSocketHandler) HandleTournamentDisconnect(tournamentID, userID uuid.UUID, disconnectedAt time.Time) {
	// Remove from pre-lobby service
	participantInfo := h.preLobbyService.RemoveParticipant(tournamentID, userID)
	if participantInfo == nil {
		return
	}

	h.logger.Info("participant disconnected from pre-lobby", map[string]interface{}{
		"tournament_id":   tournamentID.String(),
		"user_id":         userID.String(),
		"display_name":    participantInfo.DisplayName,
		"disconnected_at": disconnectedAt,
	})

	// Broadcast participant left (with 5s delay for detection)
	go func() {
		time.Sleep(5 * time.Second)

		// Check if participant reconnected
		if h.preLobbyService.IsParticipantConnected(tournamentID, userID) {
			return // Reconnected, don't broadcast leave
		}

		// Record activity feed event
		h.preLobbyService.RecordParticipantLeft(context.Background(), tournamentID, participantInfo.DisplayName)

		h.preLobbyService.BroadcastParticipantLeft(tournamentID, participantInfo)
	}()
}
