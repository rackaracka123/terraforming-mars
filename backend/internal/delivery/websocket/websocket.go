package websocket

import (
	"context"
	"net/http"

	"terraforming-mars-backend/internal/action"
	adminaction "terraforming-mars-backend/internal/action/admin"
	executecardaction "terraforming-mars-backend/internal/action/execute_card_action"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/session"
	sessionGame "terraforming-mars-backend/internal/session/game/core"
)

// WebSocketService provides the main WebSocket functionality
type WebSocketService struct {
	hub     *core.Hub
	handler *core.Handler
}

// NewWebSocketService creates a new WebSocket service with clean architecture
func NewWebSocketService(
	newGameRepo sessionGame.Repository,
	sessionFactory session.SessionFactory,
	hub *core.Hub,
	sessionManagerFactory session.SessionManagerFactory,
	startGameAction *action.StartGameAction,
	joinGameAction *action.JoinGameAction,
	playerReconnectedAction *action.PlayerReconnectedAction,
	playerDisconnectedAction *action.PlayerDisconnectedAction,
	selectStartingCardsAction *action.SelectStartingCardsAction,
	skipActionAction *action.SkipActionAction,
	confirmProductionCardsAction *action.ConfirmProductionCardsAction,
	buildCityAction *action.BuildCityAction,
	selectTileAction *action.SelectTileAction,
	playCardAction *action.PlayCardAction,
	executeCardActionAction *executecardaction.ExecuteCardActionAction,
	launchAsteroidAction *action.LaunchAsteroidAction,
	buildPowerPlantAction *action.BuildPowerPlantAction,
	buildAquiferAction *action.BuildAquiferAction,
	plantGreeneryAction *action.PlantGreeneryAction,
	sellPatentsAction *action.SellPatentsAction,
	confirmSellPatentsAction *action.ConfirmSellPatentsAction,
	convertHeatAction *action.ConvertHeatToTemperatureAction,
	convertPlantsAction *action.ConvertPlantsToGreeneryAction,
	confirmCardDrawAction *action.ConfirmCardDrawAction,
	giveCardAdminAction *adminaction.GiveCardAction,
	setPhaseAdminAction *adminaction.SetPhaseAction,
	setResourcesAdminAction *adminaction.SetResourcesAction,
	setProductionAdminAction *adminaction.SetProductionAction,
	setGlobalParametersAdminAction *adminaction.SetGlobalParametersAction,
	startTileSelectionAdminAction *adminaction.StartTileSelectionAction,
	setCurrentTurnAdminAction *adminaction.SetCurrentTurnAction,
	setCorporationAdminAction *adminaction.SetCorporationAction,
) *WebSocketService {
	// Use the provided hub

	// Register specific message type handlers with middleware support
	RegisterHandlers(hub, sessionManagerFactory, newGameRepo, sessionFactory, startGameAction, joinGameAction, playerReconnectedAction, playerDisconnectedAction, selectStartingCardsAction, skipActionAction, confirmProductionCardsAction, buildCityAction, selectTileAction, playCardAction, executeCardActionAction, launchAsteroidAction, buildPowerPlantAction, buildAquiferAction, plantGreeneryAction, sellPatentsAction, confirmSellPatentsAction, convertHeatAction, convertPlantsAction, confirmCardDrawAction, giveCardAdminAction, setPhaseAdminAction, setResourcesAdminAction, setProductionAdminAction, setGlobalParametersAdminAction, startTileSelectionAdminAction, setCurrentTurnAdminAction, setCorporationAdminAction)

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
