package websocket

import (
	"context"
	"net/http"

	"terraforming-mars-backend/internal/actions"
	"terraforming-mars-backend/internal/actions/card_selection"
	"terraforming-mars-backend/internal/actions/standard_projects"
	"terraforming-mars-backend/internal/admin"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/lobby"
)

// WebSocketService provides the main WebSocket functionality
type WebSocketService struct {
	hub     *core.Hub
	handler *core.Handler
}

// NewWebSocketService creates a new WebSocket service with clean architecture
func NewWebSocketService(
	lobbyService lobby.Service,
	adminService admin.AdminService,
	gameRepo game.Repository,
	cardRepo card.CardRepository,
	hub *core.Hub,
	// Connection management actions
	connectPlayerAction *actions.ConnectPlayerAction,
	disconnectPlayerAction *actions.DisconnectPlayerAction,
	// Standard project actions
	buildAquiferAction *standard_projects.BuildAquiferAction,
	launchAsteroidAction *standard_projects.LaunchAsteroidAction,
	buildPowerPlantAction *standard_projects.BuildPowerPlantAction,
	plantGreeneryAction *standard_projects.PlantGreeneryAction,
	buildCityAction *standard_projects.BuildCityAction,
	sellPatentsAction *standard_projects.SellPatentsAction,
	// Game actions
	skipAction *actions.SkipAction,
	convertHeatAction *actions.ConvertHeatToTemperatureAction,
	convertPlantsAction *actions.ConvertPlantsToGreeneryAction,
	playCardAction *actions.PlayCardAction,
	selectTileAction *actions.SelectTileAction,
	playCardActionAction *actions.PlayCardActionAction,
	// Card selection actions
	submitSellPatentsAction *card_selection.SubmitSellPatentsAction,
	selectStartingCardsAction *card_selection.SelectStartingCardsAction,
	selectProductionCardsAction *card_selection.SelectProductionCardsAction,
	confirmCardDrawAction *card_selection.ConfirmCardDrawAction,
	selectCardsAction *card_selection.SelectCardsAction,
) *WebSocketService {
	// Use the provided hub

	// Register specific message type handlers with middleware support
	RegisterHandlers(
		hub,
		lobbyService,
		adminService,
		gameRepo,
		cardRepo,
		connectPlayerAction,
		disconnectPlayerAction,
		buildAquiferAction,
		launchAsteroidAction,
		buildPowerPlantAction,
		plantGreeneryAction,
		buildCityAction,
		sellPatentsAction,
		skipAction,
		convertHeatAction,
		convertPlantsAction,
		playCardAction,
		selectTileAction,
		playCardActionAction,
		submitSellPatentsAction,
		selectStartingCardsAction,
		selectProductionCardsAction,
		confirmCardDrawAction,
		selectCardsAction,
	)

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
