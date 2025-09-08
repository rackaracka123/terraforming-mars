package websocket

import (
	"context"
	"net/http"

	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handlers"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/service"
)

// WebSocketService provides the main WebSocket functionality
type WebSocketService struct {
	hub     *core.Hub
	handler *core.Handler
}

// NewWebSocketService creates a new WebSocket service with clean architecture
func NewWebSocketService(
	gameService service.GameService,
	playerService service.PlayerService,
	standardProjectService service.StandardProjectService,
	cardService service.CardService,
	eventBus events.EventBus,
) *WebSocketService {
	// Create hub first (without handlers to break circular dependency)
	hub := core.NewHub(gameService, playerService, standardProjectService, cardService, eventBus, nil, nil, nil)

	// Now create message handlers with hub components
	manager := hub.GetManager()
	broadcaster := hub.GetBroadcaster()
	connectionHandler := handlers.NewConnectionHandler(gameService, playerService, broadcaster, manager)
	actionHandler := handlers.NewActionHandler(gameService, playerService, standardProjectService, cardService)
	eventHandler := handlers.NewEventHandler(broadcaster)

	// Set handlers in hub
	hub.SetHandlers(connectionHandler, actionHandler)
	hub.SetEventHandler(eventHandler)

	// Create HTTP handler
	httpHandler := core.NewHandler(hub)

	return &WebSocketService{
		hub:     hub,
		handler: httpHandler,
	}
}

// ServeWS handles WebSocket upgrade requests (for HTTP routing)
func (ws *WebSocketService) ServeWS(w http.ResponseWriter, r *http.Request) {
	ws.handler.ServeWS(w, r)
}

// Run starts the WebSocket service
func (ws *WebSocketService) Run(ctx context.Context) {
	ws.hub.Run(ctx)
}

// Stop gracefully stops the WebSocket service
func (ws *WebSocketService) Stop() {
	// The hub will clean up connections when the context is cancelled
}

// GetHub returns the hub for testing purposes
func (ws *WebSocketService) GetHub() *core.Hub {
	return ws.hub
}
