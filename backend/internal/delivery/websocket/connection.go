package websocket

import (
	"sync"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Connection represents a WebSocket connection
type Connection struct {
	ID         string
	PlayerID   string
	GameID     string
	Conn       *websocket.Conn
	Send       chan dto.WebSocketMessage
	Hub        *Hub
	mu         sync.RWMutex
	logger     *zap.Logger
	Done       chan struct{} // Signal channel for connection cleanup
	closeOnce  sync.Once     // Ensure cleanup only happens once
	sendClosed bool          // Track if Send channel is closed
}

// NewConnection creates a new WebSocket connection
func NewConnection(id string, conn *websocket.Conn, hub *Hub) *Connection {
	return &Connection{
		ID:     id,
		Conn:   conn,
		Send:   make(chan dto.WebSocketMessage, 256),
		Hub:    hub,
		logger: logger.Get(),
		Done:   make(chan struct{}),
	}
}

// SetPlayer associates this connection with a player
func (c *Connection) SetPlayer(playerID, gameID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.PlayerID = playerID
	c.GameID = gameID
}

// GetPlayer returns the player and game IDs for this connection
func (c *Connection) GetPlayer() (playerID, gameID string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.PlayerID, c.GameID
}

// CloseSend safely closes the Send channel
func (c *Connection) CloseSend() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.sendClosed {
		c.sendClosed = true
		close(c.Send)
	}
}

// Close closes the connection and signals cleanup
func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		close(c.Done)
		c.Conn.Close()
	})
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Connection) ReadPump() {
	c.logger.Info("🔄 Starting ReadPump for connection", zap.String("connection_id", c.ID))
	defer func() {
		c.logger.Info("🛑 ReadPump stopping for connection", zap.String("connection_id", c.ID))
		c.Close()
		c.Hub.Unregister <- c
	}()

	for {
		c.logger.Debug("📡 Waiting to read WebSocket message", zap.String("connection_id", c.ID))
		var message dto.WebSocketMessage
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("❌ WebSocket read error", zap.Error(err), zap.String("connection_id", c.ID))
			} else {
				c.logger.Info("🔌 WebSocket connection closed normally", zap.String("connection_id", c.ID), zap.Error(err))
			}
			return
		}

		c.logger.Info("📨 Received WebSocket message from client→server",
			zap.String("connection_id", c.ID),
			zap.String("message_type", string(message.Type)),
			zap.Any("payload", message.Payload))

		// Send message to hub for processing
		c.logger.Debug("📤 Sending message to hub broadcast channel", zap.String("connection_id", c.ID))
		select {
		case c.Hub.Broadcast <- HubMessage{
			Connection: c,
			Message:    message,
		}:
			c.logger.Debug("✅ Message sent to hub successfully", zap.String("connection_id", c.ID))
		default:
			c.logger.Warn("❌ Hub broadcast channel is full", zap.String("connection_id", c.ID))
			return
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Connection) WritePump() {
	defer c.Close()

	for {
		select {
		case <-c.Done:
			c.logger.Debug("Write pump stopping - connection closed", zap.String("connection_id", c.ID))
			return
		case message, ok := <-c.Send:
			if !ok {
				c.logger.Info("Send channel closed", zap.String("connection_id", c.ID))
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.logger.Debug("Sending WebSocket message",
				zap.String("connection_id", c.ID),
				zap.String("message_type", string(message.Type)))

			if err := c.Conn.WriteJSON(message); err != nil {
				c.logger.Error("WebSocket write error", zap.Error(err), zap.String("connection_id", c.ID))
				return
			}
		}
	}
}

// SendMessage sends a message to this connection
func (c *Connection) SendMessage(message dto.WebSocketMessage) {
	select {
	case c.Send <- message:
	default:
		c.logger.Warn("Connection send channel is full, closing connection", zap.String("connection_id", c.ID))
		close(c.Send)
	}
}
