package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DisconnectCallback is called when a player disconnects
type DisconnectCallback func(matchID, playerID uuid.UUID, disconnectedAt time.Time)

// Hub maintains active WebSocket connections and broadcasts messages
type Hub struct {
	// Registered clients organized by match ID
	matches map[uuid.UUID]map[uuid.UUID]*Client

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast message to specific match
	broadcast chan *BroadcastMessage

	// Callback for disconnect handling
	onDisconnect DisconnectCallback

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// BroadcastMessage represents a message to broadcast to a match
type BroadcastMessage struct {
	MatchID uuid.UUID
	Message []byte
}

// NewHub creates a new WebSocket hub
func NewHub(onDisconnect DisconnectCallback) *Hub {
	return &Hub{
		matches:      make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:     make(chan *Client, 256),
		unregister:   make(chan *Client, 256),
		broadcast:    make(chan *BroadcastMessage, 256),
		onDisconnect: onDisconnect,
	}
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
			h.broadcastToMatch(message.MatchID, message.Message)
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

	// Register client
	h.matches[client.MatchID][client.PlayerID] = client

	log.Printf("[Hub] Player %s connected to match %s (total clients: %d)",
		client.PlayerID, client.MatchID, len(h.matches[client.MatchID]))

	// Send current lobby state to the newly connected client
	h.sendLobbyStateToClient(client)
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

		log.Printf("[Hub] Player %s disconnected from match %s (remaining clients: %d)",
			client.PlayerID, client.MatchID, len(matchClients))

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
func (h *Hub) BroadcastToMatch(matchID uuid.UUID, message *Message) {
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
	// This would fetch current match state from database
	// For now, just acknowledge connection
	payload := map[string]interface{}{
		"match_id":  client.MatchID,
		"player_id": client.PlayerID,
		"connected": true,
	}

	msg, err := NewMessage(MessageTypeLobbyState, payload)
	if err != nil {
		log.Printf("[Hub] ERROR: Failed to create lobby state message: %v", err)
		return
	}

	messageBytes, _ := json.Marshal(msg)
	client.Send(messageBytes)
}

