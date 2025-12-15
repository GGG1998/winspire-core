// Package http provides HTTP handlers for the matchmaking service
package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/winspire-core/services/matchmaking/internal/application"
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
	hub               *wshub.Hub
	matchRepo         repository.MatchRepository
	publisher         *pubsub.EventPublisher
	sessionStore      *wshub.SessionStore
	disconnectService *application.DisconnectService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	hub *wshub.Hub,
	matchRepo repository.MatchRepository,
	publisher *pubsub.EventPublisher,
	sessionStore *wshub.SessionStore,
	disconnectService *application.DisconnectService,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:               hub,
		matchRepo:         matchRepo,
		publisher:         publisher,
		sessionStore:      sessionStore,
		disconnectService: disconnectService,
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
		fmt.Printf("[WS-H13] ERROR: Access denied - user=%s not participant in match=%s\n", userID, matchID)
		log.Printf("[WebSocket] Unauthorized lobby access attempt: user=%s, match=%s", userID, matchID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: you are not a participant in this match"})
		return
	}

	// Check if this is a reconnection (session was active in Redis)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	isReconnect := false
	if h.sessionStore != nil {
		wasActive, err := h.sessionStore.IsSessionActive(ctx, matchID, userID)
		if err != nil {
			log.Printf("[WebSocket] Failed to check session status: %v", err)
		} else if wasActive {
			isReconnect = true
			log.Printf("[WebSocket] Detected reconnection: match=%s, user=%s", matchID, userID)
		}
	}

	// Handle reconnection for paused matches
	if isReconnect && match.Status == domain.MatchStatusPaused {
		// Check if this user was the disconnected player
		if match.DisconnectedPlayerID != nil && *match.DisconnectedPlayerID == userID {
			log.Printf("[WebSocket] Reconnecting disconnected player: match=%s, user=%s", matchID, userID)

			// Call DisconnectService to handle reconnection logic
			if h.disconnectService != nil {
				if err := h.disconnectService.HandleReconnect(c.Request.Context(), matchID, userID); err != nil {
					log.Printf("[WebSocket] Failed to handle reconnect: %v", err)
					// Continue anyway - allow connection but log the error
				}
			}

			// Calculate elapsed time since disconnect
			elapsedSeconds := 0
			if match.DisconnectedAt != nil {
				elapsed := time.Since(*match.DisconnectedAt)
				elapsedSeconds = int(elapsed.Seconds())
			}

			// Publish PlayerReconnected event
			// Note: TournamentID would need to be fetched via round->bracket chain
			// For now using Nil as it's not critical for reconnect event
			tournamentID := uuid.Nil

			reconnectEvent := domain.NewPlayerConnectionRestored(
				matchID,
				tournamentID,
				userID,
				time.Now(),
				elapsedSeconds,
				map[string]string{
					"correlation_id": c.GetString("correlation_id"),
				},
			)
			if err := h.publisher.Publish(c.Request.Context(), reconnectEvent); err != nil {
				log.Printf("[WebSocket] Failed to publish PlayerConnectionRestored event: %v", err)
			}
		}
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocket] Failed to upgrade connection: %v", err)
		return
	}

	// Create client - constructor signature: NewClient(hub *Hub, conn *websocket.Conn, playerID, matchID uuid.UUID)
	client := wshub.NewClient(h.hub, conn, userID, matchID)

	// Register client with hub (CRITICAL: without this, broadcasts won't reach this client)
	h.hub.Register(client)

	// Publish appropriate event based on whether this is a reconnect or new join
	if !isReconnect {
		// Publish ParticipantJoinedLobby event (T060) for new connections
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
	}

	// Client will automatically receive the lobby state on connection
	// The hub handles initial state transmission via registerClient

	log.Printf("[WebSocket] Client connected: match=%s, user=%s", matchID, userID)

	// Start client goroutines for reading and writing
	go client.ReadPump()
	go client.WritePump()
}
