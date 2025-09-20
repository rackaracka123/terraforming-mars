package core

import (
	"context"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/store"

	"go.uber.org/zap"
)

// HubMessage represents a message to be processed by the hub
type HubMessage struct {
	Connection *Connection
	Message    dto.WebSocketMessage
}

// Hub manages WebSocket connections and message routing
type Hub struct {
	// Core channels
	Register   chan *Connection
	Unregister chan *Connection
	Messages   chan HubMessage

	// Components
	manager     *Manager
	broadcaster *Broadcaster

	eventBus events.EventBus

	// Store-based architecture fields
	appStore            *store.Store
	storeMessageHandler func(context.Context, *Connection, interface{})
}

// NewHubWithStore creates a new WebSocket hub using store-based architecture
func NewHubWithStore(appStore *store.Store, eventBus events.EventBus) *Hub {
	manager := NewManager()
	// Create a simplified broadcaster that works with the store
	broadcaster := NewBroadcasterWithStore(manager, appStore)

	return &Hub{
		Register:            make(chan *Connection),
		Unregister:          make(chan *Connection),
		Messages:            make(chan HubMessage),
		manager:             manager,
		broadcaster:         broadcaster,
		eventBus:            eventBus,
		appStore:            appStore,
		storeMessageHandler: nil,
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run(ctx context.Context) {
	logger.Info("🚀 Starting WebSocket hub")
	h.subscribeToEvents()
	logger.Info("✅ WebSocket hub ready to process messages")

	for {
		select {
		case <-ctx.Done():
			logger.Info("🛑 WebSocket hub shutting down")
			h.manager.CloseAllConnections()
			return

		case connection := <-h.Register:
			h.manager.RegisterConnection(connection)

		case connection := <-h.Unregister:
			playerID, gameID, shouldBroadcast := h.manager.UnregisterConnection(connection)
			if shouldBroadcast {
				h.handlePlayerDisconnection(ctx, playerID, gameID, connection)
			}

		case hubMessage := <-h.Messages:
			// Route message to appropriate handler
			h.routeMessage(ctx, hubMessage)
		}
	}
}

// GetBroadcaster returns the broadcaster for handlers
func (h *Hub) GetBroadcaster() *Broadcaster {
	return h.broadcaster
}

// GetManager returns the manager for handlers
func (h *Hub) GetManager() *Manager {
	return h.manager
}

// routeMessage routes incoming messages to the store-based handler
func (h *Hub) routeMessage(ctx context.Context, hubMessage HubMessage) {
	connection := hubMessage.Connection
	message := hubMessage.Message

	logger.Info("🔄 Routing WebSocket message",
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)))

	// Route to store message handler
	if h.storeMessageHandler != nil {
		logger.Debug("🎯 Routing to store message handler",
			zap.String("message_type", string(message.Type)))
		h.storeMessageHandler(ctx, connection, message)
	} else {
		logger.Warn("❓ No store message handler configured")
		h.sendError(connection, ErrHandlerNotAvailable)
	}
}

// subscribeToEvents sets up event listeners
func (h *Hub) subscribeToEvents() {
	// Subscribe to game updates for broadcasting updates
	h.eventBus.Subscribe(events.EventTypeGameUpdated, h.handleGameUpdated)

	// Subscribe to global parameter changes to trigger game updates (consolidated event only)
	h.eventBus.Subscribe(events.EventTypeGlobalParametersChanged, h.handleGlobalParameterChange)

	logger.Info("📡 WebSocket hub subscribed to events")
}

// handleGameUpdated processes game updated events
func (h *Hub) handleGameUpdated(ctx context.Context, event events.Event) error {
	payload := event.GetPayload().(events.GameUpdatedEventData)
	gameID := payload.GameID

	logger.Info("🎮 Processing game updated broadcast",
		zap.String("game_id", gameID))

	// Delegate to broadcaster
	h.broadcaster.SendPersonalizedGameUpdates(ctx, gameID)

	logger.Info("✅ Game updated broadcast completed", zap.String("game_id", gameID))
	return nil
}

// handleGlobalParameterChange handles global parameter changes (temperature, oceans, etc.)
func (h *Hub) handleGlobalParameterChange(ctx context.Context, event events.Event) error {
	// Extract game ID from the event payload
	var gameID string

	// Handle consolidated global parameter event
	switch event.GetType() {
	case events.EventTypeGlobalParametersChanged:
		payload := event.GetPayload().(events.GlobalParametersChangedEventData)
		gameID = payload.GameID
		logger.Debug("🌍 Processing global parameters change event",
			zap.String("game_id", gameID),
			zap.Strings("change_types", payload.ChangeTypes))
	default:
		logger.Warn("⚠️ Unknown global parameter event type", zap.String("event_type", event.GetType()))
		return nil
	}

	// Trigger game update broadcast to notify clients of parameter changes
	h.broadcaster.SendPersonalizedGameUpdates(ctx, gameID)

	logger.Debug("✅ Global parameter change broadcast completed", zap.String("game_id", gameID))
	return nil
}

// handlePlayerDisconnection handles player disconnection broadcasting
func (h *Hub) handlePlayerDisconnection(ctx context.Context, playerID, gameID string, connection *Connection) {
	h.broadcaster.BroadcastPlayerDisconnection(ctx, playerID, gameID, connection)
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

	h.broadcaster.SendToConnection(connection, message)
}

// Standard error messages for hub operations
const (
	ErrHandlerNotAvailable = "Handler not available"
	ErrUnknownMessageType  = "Unknown message type"
)

// SetStoreMessageHandler sets a handler function that uses the store
func (h *Hub) SetStoreMessageHandler(handler func(context.Context, *Connection, interface{})) {
	h.storeMessageHandler = handler
}

// GetStore returns the application store
func (h *Hub) GetStore() *store.Store {
	return h.appStore
}
