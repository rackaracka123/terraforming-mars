package core

import (
	"context"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

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
	HandlePlayerStartingCardOptions(ctx context.Context, event events.Event) error
}

// Hub manages WebSocket connections and message routing
type Hub struct {
	// Core channels
	Register   chan *Connection
	Unregister chan *Connection
	Messages   chan HubMessage

	// Components
	manager           *Manager
	broadcaster       *Broadcaster
	connectionHandler MessageHandler
	actionHandler     MessageHandler
	eventHandler      EventHandler // Delegate for domain events
	logger            *zap.Logger

	// Services (for routing to handlers)
	gameService            service.GameService
	playerService          service.PlayerService
	standardProjectService service.StandardProjectService
	cardService            service.CardService
	eventBus               events.EventBus
}

// NewHub creates a new WebSocket hub with clean architecture
func NewHub(
	gameService service.GameService,
	playerService service.PlayerService,
	standardProjectService service.StandardProjectService,
	cardService service.CardService,
	eventBus events.EventBus,
	connectionHandler MessageHandler,
	actionHandler MessageHandler,
	eventHandler EventHandler,
) *Hub {
	manager := NewManager()
	broadcaster := NewBroadcaster(manager, gameService, playerService)

	return &Hub{
		Register:               make(chan *Connection),
		Unregister:             make(chan *Connection),
		Messages:               make(chan HubMessage),
		manager:                manager,
		broadcaster:            broadcaster,
		connectionHandler:      connectionHandler,
		actionHandler:          actionHandler,
		eventHandler:           eventHandler,
		logger:                 logger.Get(),
		gameService:            gameService,
		playerService:          playerService,
		standardProjectService: standardProjectService,
		cardService:            cardService,
		eventBus:               eventBus,
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run(ctx context.Context) {
	h.logger.Info("🚀 Starting WebSocket hub")
	h.subscribeToEvents()
	h.logger.Info("✅ WebSocket hub ready to process messages")

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("🛑 WebSocket hub shutting down")
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

// GetServices returns services for handlers (clean dependency injection)
func (h *Hub) GetServices() (service.GameService, service.PlayerService, service.StandardProjectService, service.CardService) {
	return h.gameService, h.playerService, h.standardProjectService, h.cardService
}

// GetBroadcaster returns the broadcaster for handlers
func (h *Hub) GetBroadcaster() *Broadcaster {
	return h.broadcaster
}

// GetManager returns the manager for handlers
func (h *Hub) GetManager() *Manager {
	return h.manager
}

// SetHandlers sets the message handlers (used to break circular dependency)
func (h *Hub) SetHandlers(connectionHandler, actionHandler MessageHandler) {
	h.connectionHandler = connectionHandler
	h.actionHandler = actionHandler
}

// SetEventHandler sets the event handler (used to break circular dependency)
func (h *Hub) SetEventHandler(eventHandler EventHandler) {
	h.eventHandler = eventHandler
}

// routeMessage routes incoming messages to appropriate handlers
func (h *Hub) routeMessage(ctx context.Context, hubMessage HubMessage) {
	connection := hubMessage.Connection
	message := hubMessage.Message

	h.logger.Info("🔄 Routing WebSocket message",
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)))

	// Route to appropriate handler based on message type
	switch message.Type {
	case dto.MessageTypePlayerConnect, dto.MessageTypePlayerReconnect:
		if h.connectionHandler != nil {
			h.logger.Debug("🚪 Routing to connection handler")
			h.connectionHandler.HandleMessage(ctx, connection, message)
		} else {
			h.logger.Error("Connection handler not set")
			h.sendError(connection, ErrHandlerNotAvailable)
		}
	case dto.MessageTypePlayAction:
		if h.actionHandler != nil {
			h.logger.Debug("🎮 Routing to action handler")
			h.actionHandler.HandleMessage(ctx, connection, message)
		} else {
			h.logger.Error("Action handler not set")
			h.sendError(connection, ErrHandlerNotAvailable)
		}
	default:
		h.logger.Warn("❓ Unknown message type",
			zap.String("message_type", string(message.Type)))
		h.sendError(connection, ErrUnknownMessageType)
	}
}

// subscribeToEvents sets up event listeners
func (h *Hub) subscribeToEvents() {
	// Subscribe to game updates for broadcasting updates
	h.eventBus.Subscribe(events.EventTypeGameUpdated, h.handleGameUpdated)

	// Subscribe to card events (using new consolidated event names)
	h.eventBus.Subscribe(events.EventTypeCardDealt, h.handlePlayerStartingCardOptions) // Renamed from PlayerStartingCardOptions

	// Subscribe to global parameter changes to trigger game updates (consolidated event only)
	h.eventBus.Subscribe(events.EventTypeGlobalParametersChanged, h.handleGlobalParameterChange)

	h.logger.Info("📡 WebSocket hub subscribed to events")
}

// handleGameUpdated processes game updated events
func (h *Hub) handleGameUpdated(ctx context.Context, event events.Event) error {
	payload := event.GetPayload().(events.GameUpdatedEventData)
	gameID := payload.GameID

	h.logger.Info("🎮 Processing game updated broadcast",
		zap.String("game_id", gameID))

	// Delegate to broadcaster
	h.broadcaster.SendPersonalizedGameUpdates(ctx, gameID)

	h.logger.Info("✅ Game updated broadcast completed", zap.String("game_id", gameID))
	return nil
}

// handlePlayerStartingCardOptions handles card option events
func (h *Hub) handlePlayerStartingCardOptions(ctx context.Context, event events.Event) error {
	h.logger.Debug("🃏 Card options event received - delegating to event handler")

	// Delegate to the proper event handler
	if h.eventHandler != nil {
		return h.eventHandler.HandlePlayerStartingCardOptions(ctx, event)
	}

	h.logger.Warn("⚠️ No event handler configured")
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
		h.logger.Debug("🌍 Processing global parameters change event",
			zap.String("game_id", gameID),
			zap.Strings("change_types", payload.ChangeTypes))
	default:
		h.logger.Warn("⚠️ Unknown global parameter event type", zap.String("event_type", event.GetType()))
		return nil
	}

	// Trigger game update broadcast to notify clients of parameter changes
	h.broadcaster.SendPersonalizedGameUpdates(ctx, gameID)

	h.logger.Debug("✅ Global parameter change broadcast completed", zap.String("game_id", gameID))
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
