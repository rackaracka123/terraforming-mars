package websocket

import (
	"context"
	"net/http"

	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/repository"
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
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	sessionManager session.SessionManager,
) *WebSocketService {
	// Create hub with shared SessionManager
	hub := core.NewHub(eventBus, sessionManager)

	// Register specific message type handlers with middleware support
	RegisterHandlers(hub, gameService, playerService, standardProjectService, cardService, gameRepo, playerRepo)

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
