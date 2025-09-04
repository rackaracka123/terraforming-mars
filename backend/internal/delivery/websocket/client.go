package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost during development
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:3000" || origin == ""
	},
}

// Client represents a WebSocket client connection
type Client struct {
	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// The hub that manages this client
	hub *Hub

	// Client metadata
	ID       string
	PlayerID string
	GameID   string
	mutex    sync.RWMutex
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte, 256),
		hub:  hub,
		ID:   generateClientID(),
	}
}

// SetPlayerInfo sets the player and game information for this client
func (c *Client) SetPlayerInfo(playerID, gameID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.PlayerID = playerID
	c.GameID = gameID
}

// GetPlayerInfo returns the player and game information for this client
func (c *Client) GetPlayerInfo() (playerID, gameID string) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.PlayerID, c.GameID
}

// readPump handles reading messages from the WebSocket connection
func (c *Client) readPump() {
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
				playerID, gameID := c.GetPlayerInfo()
				logger.WithClientContext(c.ID, playerID, gameID).Error("Unexpected WebSocket close", zap.Error(err))
			}
			break
		}

		// Parse the message
		var msg dto.WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			playerID, gameID := c.GetPlayerInfo()
			logger.WithClientContext(c.ID, playerID, gameID).Error("Error parsing WebSocket message", zap.Error(err))
			c.sendError("Invalid message format")
			continue
		}

		// Handle the message
		c.handleMessage(&msg)
	}
}

// writePump handles writing messages to the WebSocket connection
func (c *Client) writePump() {
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
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write each message as a separate WebSocket frame
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

			// Send any additional queued messages as separate frames
			n := len(c.send)
			for i := 0; i < n; i++ {
				additionalMessage := <-c.send
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.conn.WriteMessage(websocket.TextMessage, additionalMessage); err != nil {
					return
				}
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(msg *dto.WebSocketMessage) {
	playerID, gameID := c.GetPlayerInfo()
	log := logger.WithClientContext(c.ID, playerID, gameID)
	
	log.Debug("Received WebSocket message", zap.String("message_type", string(msg.Type)))
	
	switch msg.Type {
	case dto.MessageTypePlayerConnect:
		c.hub.handlePlayerConnect(c, msg)
	case dto.MessageTypePlayAction:
		c.hub.handlePlayAction(c, msg)
	default:
		log.Warn("Unknown message type", zap.String("message_type", string(msg.Type)))
		c.sendError("Unknown message type")
	}
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(msg *dto.WebSocketMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		playerID, gameID := c.GetPlayerInfo()
		logger.WithClientContext(c.ID, playerID, gameID).Error("Error marshaling message", zap.Error(err))
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
		delete(c.hub.clients, c)
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(message string) {
	errorMsg := &dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: dto.ErrorPayload{
			Message: message,
		},
	}
	c.SendMessage(errorMsg)
}

// ServeWS handles websocket requests from clients
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Get().Error("WebSocket upgrade error", 
			zap.Error(err),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		)
		return
	}

	client := NewClient(conn, hub)
	hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// generateClientID generates a unique client ID
func generateClientID() string {
	// Simple timestamp-based ID for now
	return time.Now().Format("20060102150405") + "-" + time.Now().Format("000")
}
