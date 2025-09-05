package websocket

import (
	"context"
	"sync"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Connection represents a WebSocket connection
type Connection struct {
	ID       string
	PlayerID string
	GameID   string
	Conn     *websocket.Conn
	Send     chan dto.WebSocketMessage
	Hub      *Hub
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewConnection creates a new WebSocket connection
func NewConnection(id string, conn *websocket.Conn, hub *Hub) *Connection {
	return &Connection{
		ID:     id,
		Conn:   conn,
		Send:   make(chan dto.WebSocketMessage, 256),
		Hub:    hub,
		logger: logger.Get(),
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

// ReadPump pumps messages from the websocket connection to the hub
func (c *Connection) ReadPump(ctx context.Context) {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Connection read pump stopping due to context cancellation", zap.String("connection_id", c.ID))
			return
		default:
			var message dto.WebSocketMessage
			err := c.Conn.ReadJSON(&message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error("WebSocket read error", zap.Error(err), zap.String("connection_id", c.ID))
				} else {
					c.logger.Info("WebSocket connection closed", zap.String("connection_id", c.ID))
				}
				return
			}

			c.logger.Debug("Received WebSocket message",
				zap.String("connection_id", c.ID),
				zap.String("message_type", string(message.Type)))

			// Send message to hub for processing
			select {
			case c.Hub.Broadcast <- HubMessage{
				Connection: c,
				Message:    message,
			}:
			default:
				c.logger.Warn("Hub broadcast channel is full", zap.String("connection_id", c.ID))
				return
			}
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Connection) WritePump(ctx context.Context) {
	defer c.Conn.Close()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Connection write pump stopping due to context cancellation", zap.String("connection_id", c.ID))
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