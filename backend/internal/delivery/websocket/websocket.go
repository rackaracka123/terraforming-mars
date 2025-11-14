package websocket

import (
	"context"
	"net/http"

	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/actions"
	"terraforming-mars-backend/internal/actions/card_selection"
	"terraforming-mars-backend/internal/actions/standard_projects"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
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
	lobbyService lobby.Service,
	playerService service.PlayerService,
	standardProjectService service.StandardProjectService,
	cardService service.CardService,
	adminService service.AdminService,
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardRepo game.CardRepository,
	hub *core.Hub,
	buildAquiferAction *standard_projects.BuildAquiferAction,
	launchAsteroidAction *standard_projects.LaunchAsteroidAction,
	buildPowerPlantAction *standard_projects.BuildPowerPlantAction,
	plantGreeneryAction *standard_projects.PlantGreeneryAction,
	buildCityAction *standard_projects.BuildCityAction,
	skipAction *actions.SkipAction,
	convertHeatAction *actions.ConvertHeatToTemperatureAction,
	convertPlantsAction *actions.ConvertPlantsToGreeneryAction,
	sellPatentsAction *standard_projects.SellPatentsAction,
	submitSellPatentsAction *card_selection.SubmitSellPatentsAction,
	selectStartingCardsAction *card_selection.SelectStartingCardsAction,
	selectProductionCardsAction *card_selection.SelectProductionCardsAction,
	confirmCardDrawAction *card_selection.ConfirmCardDrawAction,
	playCardAction *actions.PlayCardAction,
	selectTileAction *actions.SelectTileAction,
	playCardActionAction *actions.PlayCardActionAction,
) *WebSocketService {
	// Use the provided hub

	// Register specific message type handlers with middleware support
	RegisterHandlers(
		hub,
		gameService,
		lobbyService,
		playerService,
		standardProjectService,
		cardService,
		adminService,
		gameRepo,
		playerRepo,
		cardRepo,
		buildAquiferAction,
		launchAsteroidAction,
		buildPowerPlantAction,
		plantGreeneryAction,
		buildCityAction,
		skipAction,
		convertHeatAction,
		convertPlantsAction,
		sellPatentsAction,
		submitSellPatentsAction,
		selectStartingCardsAction,
		selectProductionCardsAction,
		confirmCardDrawAction,
		playCardAction,
		selectTileAction,
		playCardActionAction,
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
