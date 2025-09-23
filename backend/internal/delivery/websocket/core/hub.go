package core

import (
	"context"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// MessageHandler defines the interface for handling different message types
type MessageHandler interface {
	HandleMessage(ctx context.Context, connection *Connection, message dto.WebSocketMessage)
}

// HubMessage represents a message to be processed by the hub
type HubMessage struct {
	Connection *Connection
	Message    dto.WebSocketMessage
}

// EventHandler interface for handling domain events
type EventHandler interface {
}

// Hub manages WebSocket connections and message routing
type Hub struct {
	// Core channels
	Register   chan *Connection
	Unregister chan *Connection
	Messages   chan HubMessage

	// Components
	manager        *Manager
	sessionManager session.SessionManager
	logger         *zap.Logger

	// Handler registry for specific message types
	handlers map[dto.MessageType]MessageHandler

	// Event bus for subscribing to domain events
	eventBus events.EventBus
}

// NewHub creates a new WebSocket hub with clean architecture
func NewHub(eventBus events.EventBus, sessionManager session.SessionManager) *Hub {
	manager := NewManager()

	return &Hub{
		Register:       make(chan *Connection),
		Unregister:     make(chan *Connection),
		Messages:       make(chan HubMessage),
		manager:        manager,
		sessionManager: sessionManager,
		logger:         logger.Get(),
		handlers:       make(map[dto.MessageType]MessageHandler),
		eventBus:       eventBus,
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run(ctx context.Context) {
	h.logger.Info("🚀 Starting WebSocket hub")
	h.logger.Info("✅ WebSocket hub ready to process messages")

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("🛑 WebSocket hub shutting down")
			h.manager.CloseAllConnections()
			return

		case connection := <-h.Register:
			h.manager.RegisterConnection(connection)
			// Session registration will happen when first message is received

		case connection := <-h.Unregister:
			playerID, gameID, _ := h.manager.UnregisterConnection(connection)
			// Also unregister from session manager
			if playerID != "" && gameID != "" {
				h.sessionManager.UnregisterSession(playerID, gameID)
			}

		case hubMessage := <-h.Messages:
			// Route message to appropriate handler
			h.routeMessage(ctx, hubMessage)
		}
	}
}

// RegisterHandler registers a message handler for a specific message type
func (h *Hub) RegisterHandler(messageType dto.MessageType, handler MessageHandler) {
	h.handlers[messageType] = handler
}

// routeMessage routes incoming messages to appropriate handlers
func (h *Hub) routeMessage(ctx context.Context, hubMessage HubMessage) {
	connection := hubMessage.Connection
	message := hubMessage.Message

	h.logger.Info("🔄 Routing WebSocket message",
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)))

	// Register connection with session manager if it has player info
	playerID, gameID := connection.GetPlayer()
	if playerID != "" && gameID != "" {
		h.sessionManager.RegisterSession(playerID, gameID, connection.SendMessage)
	}

	// Check if we have a registered handler for this message type
	if handler, exists := h.handlers[message.Type]; exists {
		h.logger.Debug("🎯 Routing to registered message handler",
			zap.String("message_type", string(message.Type)))
		handler.HandleMessage(ctx, connection, message)
	} else {
		h.logger.Warn("❓ Unknown message type",
			zap.String("message_type", string(message.Type)))
		h.sendError(connection, ErrUnknownMessageType)
	}
}

// sendError sends an error message to a connection
func (h *Hub) sendError(connection *Connection, errorMessage string) {
	_, gameID := connection.GetPlayer()

	message := dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: dto.ErrorPayload{
			Message: errorMessage,
		},
		GameID: gameID,
	}

	connection.SendMessage(message)
}

// Standard error messages for hub operations
const (
	ErrHandlerNotAvailable = "Handler not available"
	ErrUnknownMessageType  = "Unknown message type"
)
