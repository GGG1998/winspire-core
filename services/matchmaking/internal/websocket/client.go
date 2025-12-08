package websocket

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// T136: Maximum message size allowed from peer (prevent DoS attacks)
	// Reasonable limit for game lobby messages (heartbeats, ready status, etc.)
	maxMessageSize = 8192 // 8KB
)

// Client represents a WebSocket client connection
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	PlayerID uuid.UUID
	MatchID  uuid.UUID

	// Heartbeat tracking
	lastHeartbeat time.Time
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, playerID, matchID uuid.UUID) *Client {
	return &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		PlayerID:      playerID,
		MatchID:       matchID,
		lastHeartbeat: time.Now(),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for player %s: %v", c.PlayerID, err)
			}
			break
		}

		// Update heartbeat timestamp
		c.lastHeartbeat = time.Now()

		// Forward message to hub for processing
		c.hub.handleClientMessage(c, message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send the message
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send sends a message to the client
func (c *Client) Send(message []byte) {
	select {
	case c.send <- message:
	default:
		// Client send buffer is full, close connection
		log.Printf("Client send buffer full for player %s, closing connection", c.PlayerID)
		close(c.send)
	}
}

// Close closes the client connection
func (c *Client) Close() {
	close(c.send)
	c.conn.Close()
}

// IsConnected returns true if client has sent heartbeat recently
func (c *Client) IsConnected() bool {
	return time.Since(c.lastHeartbeat) < pongWait
}

// TimeSinceLastHeartbeat returns duration since last heartbeat
func (c *Client) TimeSinceLastHeartbeat() time.Duration {
	return time.Since(c.lastHeartbeat)
}
