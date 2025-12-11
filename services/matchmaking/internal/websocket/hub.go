package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/winspire-core/services/matchmaking/internal/repository"
)

// RoomType identifies the type of room (match or tournament pre-lobby)
type RoomType string

const (
	RoomTypeMatch      RoomType = "match"
	RoomTypeTournament RoomType = "tournament"
)

// DisconnectCallback is called when a player disconnects (T112)
type DisconnectCallback func(matchID, playerID uuid.UUID, disconnectedAt time.Time)

// TournamentDisconnectCallback is called when a participant disconnects from pre-lobby
type TournamentDisconnectCallback func(tournamentID, userID uuid.UUID, disconnectedAt time.Time)

// Hub maintains active WebSocket connections and broadcasts messages
type Hub struct {
	// Registered clients organized by match ID
	matches map[uuid.UUID]map[uuid.UUID]*Client

	// T081: Track lobby join timestamps for no-show detection
	lobbyJoinTimes map[uuid.UUID]map[uuid.UUID]time.Time

	// Pre-lobby connections organized by tournament ID
	tournaments map[uuid.UUID]map[uuid.UUID]*Client

	// Track tournament join timestamps
	tournamentJoinTimes map[uuid.UUID]map[uuid.UUID]time.Time

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast message to specific match
	broadcast chan *BroadcastMessage

	// Callback for disconnect handling
	onDisconnect DisconnectCallback

	// Callback for tournament pre-lobby disconnect handling
	onTournamentDisconnect TournamentDisconnectCallback

	// Redis-based session persistence
	sessionStore *SessionStore

	// Match repository for fetching full match state
	matchRepo repository.MatchRepository

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// BroadcastMessage represents a message to broadcast to a match or tournament
type BroadcastMessage struct {
	RoomType RoomType
	RoomID   uuid.UUID // MatchID or TournamentID
	MatchID  uuid.UUID // Deprecated: use RoomID with RoomType
	Message  []byte
}

// NewHub creates a new WebSocket hub
func NewHub(sessionStore *SessionStore, matchRepo repository.MatchRepository, onDisconnect DisconnectCallback) *Hub {
	return &Hub{
		matches:             make(map[uuid.UUID]map[uuid.UUID]*Client),
		lobbyJoinTimes:      make(map[uuid.UUID]map[uuid.UUID]time.Time),
		tournaments:         make(map[uuid.UUID]map[uuid.UUID]*Client),
		tournamentJoinTimes: make(map[uuid.UUID]map[uuid.UUID]time.Time),
		register:            make(chan *Client, 256),
		unregister:          make(chan *Client, 256),
		broadcast:           make(chan *BroadcastMessage, 256),
		onDisconnect:        onDisconnect,
		sessionStore:        sessionStore,
		matchRepo:           matchRepo,
	}
}

// SetDisconnectCallback sets the callback for match disconnections
func (h *Hub) SetDisconnectCallback(cb DisconnectCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onDisconnect = cb
}

// Register registers a client with the hub for receiving broadcasts
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// SetTournamentDisconnectCallback sets the callback for tournament disconnections
func (h *Hub) SetTournamentDisconnectCallback(cb TournamentDisconnectCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onTournamentDisconnect = cb
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	// Start heartbeat monitor
	go h.monitorHeartbeats()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			// Route to appropriate room type
			switch message.RoomType {
			case RoomTypeTournament:
				h.broadcastToTournament(message.RoomID, message.Message)
			default:
				// Default to match broadcast for backward compatibility
				roomID := message.RoomID
				if roomID == uuid.Nil {
					roomID = message.MatchID // Backward compatibility
				}
				h.broadcastToMatch(roomID, message.Message)
			}
		}
	}
}

// registerClient registers a new client connection
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Initialize match map if it doesn't exist
	if h.matches[client.MatchID] == nil {
		h.matches[client.MatchID] = make(map[uuid.UUID]*Client)
	}

	// T081: Initialize lobby join times map if needed
	if h.lobbyJoinTimes[client.MatchID] == nil {
		h.lobbyJoinTimes[client.MatchID] = make(map[uuid.UUID]time.Time)
	}

	// Register client and track join time
	h.matches[client.MatchID][client.PlayerID] = client
	h.lobbyJoinTimes[client.MatchID][client.PlayerID] = time.Now()

	// Persist session to Redis
	if h.sessionStore != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.sessionStore.AddSession(ctx, client.MatchID, client.PlayerID); err != nil {
			log.Printf("[Hub] ERROR: Failed to add session to Redis: %v", err)
		}
	}

	log.Printf("[Hub] Player %s connected to match %s (total clients: %d)",
		client.PlayerID, client.MatchID, len(h.matches[client.MatchID]))

	log.Printf("[Hub-H12] About to call sendLobbyStateToClient for newly connected player=%s", client.PlayerID)
	// Send current lobby state to the newly connected client
	h.sendLobbyStateToClient(client)
	log.Printf("[Hub-H12] Finished initial sendLobbyStateToClient for player=%s", client.PlayerID)
}

// unregisterClient unregisters a client connection
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	matchClients, exists := h.matches[client.MatchID]
	if !exists {
		return
	}

	if _, exists := matchClients[client.PlayerID]; exists {
		delete(matchClients, client.PlayerID)
		close(client.send)

		// Remove session from Redis
		if h.sessionStore != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := h.sessionStore.RemoveSession(ctx, client.MatchID, client.PlayerID); err != nil {
				log.Printf("[Hub] ERROR: Failed to remove session from Redis: %v", err)
			}
		}

		// Clean up lobby join time tracking
		if joinTimes, exists := h.lobbyJoinTimes[client.MatchID]; exists {
			delete(joinTimes, client.PlayerID)
			if len(joinTimes) == 0 {
				delete(h.lobbyJoinTimes, client.MatchID)
			}
		}

		log.Printf("[Hub] Player %s disconnected from match %s (remaining clients: %d)",
			client.PlayerID, client.MatchID, len(matchClients))

		// Broadcast player_left to remaining clients in the lobby
		if len(matchClients) > 0 {
			msg, err := NewMessage(MessageType("player_left"), map[string]interface{}{
				"playerId": client.PlayerID,
			})
			if err == nil {
				msgBytes, _ := json.Marshal(msg)
				h.mu.Unlock()
				h.broadcastToMatch(client.MatchID, msgBytes)
				h.mu.Lock()
			} else {
				log.Printf("[Hub] ERROR: Failed to create player_left message: %v", err)
			}
		}

		// Clean up empty match
		if len(matchClients) == 0 {
			delete(h.matches, client.MatchID)
			log.Printf("[Hub] Match %s has no more clients, cleaned up", client.MatchID)
		}

		// Trigger disconnect callback
		if h.onDisconnect != nil {
			go h.onDisconnect(client.MatchID, client.PlayerID, time.Now())
		}
	}
}

// BroadcastToMatch sends a message to all clients in a match
// Accepts either *Message or any other type that can be marshaled to JSON
func (h *Hub) BroadcastToMatch(matchID uuid.UUID, message interface{}) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("[Hub] ERROR: Failed to marshal message for match %s: %v", matchID, err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		MatchID: matchID,
		Message: messageBytes,
	}
}

// broadcastToMatch (internal) sends bytes to all match clients
func (h *Hub) broadcastToMatch(matchID uuid.UUID, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	matchClients, exists := h.matches[matchID]
	if !exists {
		return
	}

	for _, client := range matchClients {
		select {
		case client.send <- message:
		default:
			// Client buffer full, will be cleaned up by unregister
			log.Printf("[Hub] WARN: Client buffer full for player %s", client.PlayerID)
		}
	}
}

// SendToPlayer sends a message to a specific player
func (h *Hub) SendToPlayer(matchID, playerID uuid.UUID, message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	matchClients, exists := h.matches[matchID]
	if !exists {
		return
	}

	client, exists := matchClients[playerID]
	if !exists {
		return
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("[Hub] ERROR: Failed to marshal message for player %s: %v", playerID, err)
		return
	}

	client.Send(messageBytes)
}

// GetConnectedPlayers returns list of connected player IDs for a match
func (h *Hub) GetConnectedPlayers(matchID uuid.UUID) []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	matchClients, exists := h.matches[matchID]
	if !exists {
		return []uuid.UUID{}
	}

	playerIDs := make([]uuid.UUID, 0, len(matchClients))
	for playerID := range matchClients {
		playerIDs = append(playerIDs, playerID)
	}

	return playerIDs
}

// IsPlayerConnected checks if a player is connected to a match
func (h *Hub) IsPlayerConnected(matchID, playerID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	matchClients, exists := h.matches[matchID]
	if !exists {
		return false
	}

	client, exists := matchClients[playerID]
	if !exists {
		return false
	}

	return client.IsConnected()
}

// monitorHeartbeats checks for stale connections and triggers disconnects
func (h *Hub) monitorHeartbeats() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		for matchID, matchClients := range h.matches {
			for playerID, client := range matchClients {
				timeSinceHeartbeat := client.TimeSinceLastHeartbeat()

				// Update heartbeat in Redis
				if h.sessionStore != nil && timeSinceHeartbeat < pongWait {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					if err := h.sessionStore.UpdateHeartbeat(ctx, matchID, playerID); err != nil {
						log.Printf("[Hub] ERROR: Failed to update heartbeat in Redis: %v", err)
					}
					cancel()
				}

				if timeSinceHeartbeat > pongWait {
					log.Printf("[Hub] Player %s heartbeat timeout (%v), disconnecting", playerID, timeSinceHeartbeat)
					h.mu.Unlock()
					h.unregister <- client
					h.mu.Lock()

					// Trigger disconnect callback
					if h.onDisconnect != nil {
						disconnectedAt := time.Now().Add(-timeSinceHeartbeat)
						go h.onDisconnect(matchID, playerID, disconnectedAt)
					}
				}
			}
		}
		h.mu.Unlock()
	}
}

// handleClientMessage processes incoming client messages
func (h *Hub) handleClientMessage(client *Client, messageBytes []byte) {
	var msg Message
	if err := json.Unmarshal(messageBytes, &msg); err != nil {
		log.Printf("[Hub] ERROR: Failed to unmarshal message from player %s: %v", client.PlayerID, err)
		return
	}

	switch msg.Type {
	case MessageTypeHeartbeat:
		// Heartbeat handled by ReadPump (updates lastHeartbeat)
		// No additional action needed

	default:
		log.Printf("[Hub] Received message type %s from player %s in match %s",
			msg.Type, client.PlayerID, client.MatchID)
		// Additional message types handled by application layer
	}
}

// sendLobbyStateToClient sends current lobby state to a client
func (h *Hub) sendLobbyStateToClient(client *Client) {
	fmt.Printf("[Hub-H12] >>>> sendLobbyStateToClient CALLED for player=%s match=%s\n", client.PlayerID, client.MatchID)

	// Fetch full match state from database
	if h.matchRepo == nil {
		log.Printf("[Hub] WARN: matchRepo is nil, cannot fetch match state")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	match, err := h.matchRepo.GetByID(ctx, client.MatchID)
	if err != nil {
		log.Printf("[Hub] ERROR: Failed to fetch match %s: %v", client.MatchID, err)
		return
	}

	// Build lobby state payload with match data
	// Note: Frontend already has participant details from initial REST API call
	// We only need to send updated match state here
	payload := map[string]interface{}{
		"match": map[string]interface{}{
			"id":                       match.ID,
			"status":                   match.Status,
			"participant1_id":          match.Participant1ID,
			"participant2_id":          match.Participant2ID,
			"participant1_ready":       match.Participant1Ready,
			"participant2_ready":       match.Participant2Ready,
			"participant1_game_loaded": match.Participant1GameLoaded,
			"participant2_game_loaded": match.Participant2GameLoaded,
			"winner_id":                match.WinnerID,
			"score_player1":            match.ScorePlayer1,
			"score_player2":            match.ScorePlayer2,
			"disconnected_player_id":   match.DisconnectedPlayerID,
			"disconnected_at":          match.DisconnectedAt,
			"started_at":               match.StartedAt,
			"completed_at":             match.CompletedAt,
		},
	}

	fmt.Printf("[Hub-H12] Payload constructed with: p1_game_loaded=%v p2_game_loaded=%v\n", payload["match"].(map[string]interface{})["participant1_game_loaded"], payload["match"].(map[string]interface{})["participant2_game_loaded"])

	// #region agent log
	func() {
		f, _ := os.OpenFile("/Users/gabrieldomanowski/programming/winspire-core/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			defer f.Close()
			json.NewEncoder(f).Encode(map[string]interface{}{"location": "hub.go:428", "message": "lobby_state payload constructed", "data": map[string]interface{}{"payloadMatch": payload["match"]}, "timestamp": time.Now().UnixMilli(), "sessionId": "debug-session", "hypothesisId": "H10"})
		}
	}()
	// #endregion

	msg, err := NewMessage(MessageTypeLobbyState, payload)
	if err != nil {
		log.Printf("[Hub] ERROR: Failed to create lobby state message: %v", err)
		return
	}

	messageBytes, _ := json.Marshal(msg)
	client.Send(messageBytes)
}

// ============================================================================
// Tournament Pre-Lobby Methods
// ============================================================================

// TournamentClient represents a client connected to a tournament pre-lobby
type TournamentClient struct {
	*Client
	TournamentID uuid.UUID
	UserID       uuid.UUID
	DisplayName  string
	AvatarURL    *string
}

// RegisterTournamentClient registers a client to a tournament pre-lobby
func (h *Hub) RegisterTournamentClient(tournamentID, userID uuid.UUID, displayName string, avatarURL *string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Initialize tournament map if it doesn't exist
	if h.tournaments[tournamentID] == nil {
		h.tournaments[tournamentID] = make(map[uuid.UUID]*Client)
	}

	// Initialize tournament join times map if needed
	if h.tournamentJoinTimes[tournamentID] == nil {
		h.tournamentJoinTimes[tournamentID] = make(map[uuid.UUID]time.Time)
	}

	// Register client and track join time
	h.tournaments[tournamentID][userID] = client
	h.tournamentJoinTimes[tournamentID][userID] = time.Now()

	log.Printf("[Hub] User %s (%s) connected to tournament %s (total participants: %d)",
		userID, displayName, tournamentID, len(h.tournaments[tournamentID]))
}

// UnregisterTournamentClient unregisters a client from a tournament pre-lobby
func (h *Hub) UnregisterTournamentClient(tournamentID, userID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	tournamentClients, exists := h.tournaments[tournamentID]
	if !exists {
		return
	}

	client, exists := tournamentClients[userID]
	if !exists {
		return
	}

	delete(tournamentClients, userID)
	close(client.send)

	// Clean up tournament join time tracking
	if joinTimes, exists := h.tournamentJoinTimes[tournamentID]; exists {
		delete(joinTimes, userID)
		if len(joinTimes) == 0 {
			delete(h.tournamentJoinTimes, tournamentID)
		}
	}

	log.Printf("[Hub] User %s disconnected from tournament %s (remaining participants: %d)",
		userID, tournamentID, len(tournamentClients))

	// Clean up empty tournament
	if len(tournamentClients) == 0 {
		delete(h.tournaments, tournamentID)
		log.Printf("[Hub] Tournament %s has no more participants, cleaned up", tournamentID)
	}

	// Trigger disconnect callback
	if h.onTournamentDisconnect != nil {
		go h.onTournamentDisconnect(tournamentID, userID, time.Now())
	}
}

// GetTournamentParticipants returns list of connected user IDs for a tournament
func (h *Hub) GetTournamentParticipants(tournamentID uuid.UUID) []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tournamentClients, exists := h.tournaments[tournamentID]
	if !exists {
		return []uuid.UUID{}
	}

	userIDs := make([]uuid.UUID, 0, len(tournamentClients))
	for userID := range tournamentClients {
		userIDs = append(userIDs, userID)
	}

	return userIDs
}

// GetTournamentParticipantCount returns the number of connected participants
func (h *Hub) GetTournamentParticipantCount(tournamentID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tournamentClients, exists := h.tournaments[tournamentID]
	if !exists {
		return 0
	}

	return len(tournamentClients)
}

// IsTournamentParticipantConnected checks if a user is connected to a tournament
func (h *Hub) IsTournamentParticipantConnected(tournamentID, userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tournamentClients, exists := h.tournaments[tournamentID]
	if !exists {
		return false
	}

	client, exists := tournamentClients[userID]
	if !exists {
		return false
	}

	return client.IsConnected()
}

// BroadcastToTournament sends a message to all clients in a tournament pre-lobby
func (h *Hub) BroadcastToTournament(tournamentID uuid.UUID, message *Message) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("[Hub] ERROR: Failed to marshal message for tournament %s: %v", tournamentID, err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		RoomType: RoomTypeTournament,
		RoomID:   tournamentID,
		Message:  messageBytes,
	}
}

// broadcastToTournament (internal) sends bytes to all tournament clients
func (h *Hub) broadcastToTournament(tournamentID uuid.UUID, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tournamentClients, exists := h.tournaments[tournamentID]
	if !exists {
		return
	}

	for _, client := range tournamentClients {
		select {
		case client.send <- message:
		default:
			// Client buffer full, will be cleaned up by unregister
			log.Printf("[Hub] WARN: Client buffer full for tournament participant")
		}
	}
}

// SendToTournamentParticipant sends a message to a specific participant in a tournament
func (h *Hub) SendToTournamentParticipant(tournamentID, userID uuid.UUID, message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tournamentClients, exists := h.tournaments[tournamentID]
	if !exists {
		return
	}

	client, exists := tournamentClients[userID]
	if !exists {
		return
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("[Hub] ERROR: Failed to marshal message for tournament participant %s: %v", userID, err)
		return
	}

	client.Send(messageBytes)
}

// GetTournamentJoinTime returns when a user joined a tournament pre-lobby
func (h *Hub) GetTournamentJoinTime(tournamentID, userID uuid.UUID) (time.Time, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	joinTimes, exists := h.tournamentJoinTimes[tournamentID]
	if !exists {
		return time.Time{}, false
	}

	joinTime, exists := joinTimes[userID]
	return joinTime, exists
}

// BroadcastServerRestarting sends a server_restarting message to all connected clients
// This is called during graceful shutdown to notify clients to prepare for reconnection
func (h *Hub) BroadcastServerRestarting() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg, err := NewMessage(MessageType("server_restarting"), map[string]interface{}{
		"message": "Server is restarting, please wait for reconnection",
	})
	if err != nil {
		log.Printf("[Hub] ERROR: Failed to create server_restarting message: %v", err)
		return
	}

	messageBytes, _ := json.Marshal(msg)

	// Send to all match lobby clients
	for matchID := range h.matches {
		h.mu.RUnlock()
		h.broadcastToMatch(matchID, messageBytes)
		h.mu.RLock()
	}

	// Send to all tournament pre-lobby clients
	for tournamentID := range h.tournaments {
		h.mu.RUnlock()
		h.broadcastToTournament(tournamentID, messageBytes)
		h.mu.RLock()
	}

	log.Printf("[Hub] Sent server_restarting message to all clients")
}
