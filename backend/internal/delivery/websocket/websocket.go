package websocket

import (
	"context"
	"net/http"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/session"
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
	adminService service.AdminService,
	resourceConversionService service.ResourceConversionService,
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardRepo repository.CardRepository,
	hub *core.Hub,
	sessionManager session.SessionManager,
	startGameAction *action.StartGameAction,
	joinGameAction *action.JoinGameAction,
) *WebSocketService {
	// Use the provided hub

	// Register specific message type handlers with middleware support
	RegisterHandlers(hub, sessionManager, gameService, playerService, standardProjectService, cardService, adminService, resourceConversionService, gameRepo, playerRepo, cardRepo, startGameAction, joinGameAction)

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
