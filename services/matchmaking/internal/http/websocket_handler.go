// Package http provides HTTP handlers for the matchmaking service
package http

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/winspire-core/services/matchmaking/internal/domain"
	"github.com/winspire-core/services/matchmaking/internal/pubsub"
	"github.com/winspire-core/services/matchmaking/internal/repository"
	wshub "github.com/winspire-core/services/matchmaking/internal/websocket"
	"github.com/winspire/winspire-core/libs/go/httpx"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking based on environment
		return true // Allow all origins for now (will be restricted in production)
	},
}

// WebSocketHandler handles WebSocket connections for match lobbies
type WebSocketHandler struct {
	hub       *wshub.Hub
	matchRepo repository.MatchRepository
	publisher *pubsub.EventPublisher
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	hub *wshub.Hub,
	matchRepo repository.MatchRepository,
	publisher *pubsub.EventPublisher,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:       hub,
		matchRepo: matchRepo,
		publisher: publisher,
	}
}

// UpgradeLobbyConnection upgrades HTTP connection to WebSocket for match lobby
// GET /v1/matches/:id/lobby
func (h *WebSocketHandler) UpgradeLobbyConnection(c *gin.Context) {
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

	// FR-023, FR-024: Verify user ID matches a participant in the match
	isParticipant1 := match.Participant1ID == userID
	isParticipant2 := match.Participant2ID != nil && *match.Participant2ID == userID

	if !isParticipant1 && !isParticipant2 {
		// FR-024, FR-025: Deny access with specific error message
		log.Printf("[WebSocket] Unauthorized lobby access attempt: user=%s, match=%s", userID, matchID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: you are not a participant in this match"})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocket] Failed to upgrade connection: %v", err)
		return
	}

	// Create client - constructor signature: NewClient(hub *Hub, conn *websocket.Conn, playerID, matchID uuid.UUID)
	client := wshub.NewClient(h.hub, conn, userID, matchID)

	// Publish ParticipantJoinedLobby event (T060)
	event := domain.NewParticipantJoinedLobby(
		uuid.New(), // event ID
		matchID,
		userID,
		map[string]string{
			"correlation_id": c.GetString("correlation_id"),
		},
	)
	if err := h.publisher.Publish(c.Request.Context(), event); err != nil {
		log.Printf("[WebSocket] Failed to publish ParticipantJoinedLobby event: %v", err)
	}

	// Client will automatically receive the lobby state on connection
	// The hub handles initial state transmission in the ReadPump/WritePump goroutines

	log.Printf("[WebSocket] Client connected: match=%s, user=%s", matchID, userID)

	// Start client goroutines for reading and writing
	go client.ReadPump()
	go client.WritePump()
}
