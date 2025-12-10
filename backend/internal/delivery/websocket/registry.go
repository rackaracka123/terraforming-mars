package websocket

import (
	adminAction "terraforming-mars-backend/internal/action/admin"
	cardAction "terraforming-mars-backend/internal/action/card"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	connAction "terraforming-mars-backend/internal/action/connection"
	gameAction "terraforming-mars-backend/internal/action/game"
	resconvAction "terraforming-mars-backend/internal/action/resource_conversion"
	stdprojAction "terraforming-mars-backend/internal/action/standard_project"
	tileAction "terraforming-mars-backend/internal/action/tile"
	turnAction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/admin"
	"terraforming-mars-backend/internal/delivery/websocket/handler/card"
	"terraforming-mars-backend/internal/delivery/websocket/handler/confirmation"
	"terraforming-mars-backend/internal/delivery/websocket/handler/connection"
	"terraforming-mars-backend/internal/delivery/websocket/handler/game"
	"terraforming-mars-backend/internal/delivery/websocket/handler/resource_conversion"
	"terraforming-mars-backend/internal/delivery/websocket/handler/standard_project"
	"terraforming-mars-backend/internal/delivery/websocket/handler/tile"
	"terraforming-mars-backend/internal/delivery/websocket/handler/turn_management"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// RegisterHandlers registers migrated action handlers with the hub
// Handlers call broadcaster explicitly after actions complete
func RegisterHandlers(
	hub *core.Hub,
	broadcaster *Broadcaster,
	// Game lifecycle
	createGameAction *gameAction.CreateGameAction,
	joinGameAction *gameAction.JoinGameAction,
	// Card actions
	playCardAction *cardAction.PlayCardAction,
	useCardActionAction *cardAction.UseCardActionAction,
	// Standard projects
	launchAsteroidAction *stdprojAction.LaunchAsteroidAction,
	buildPowerPlantAction *stdprojAction.BuildPowerPlantAction,
	buildAquiferAction *stdprojAction.BuildAquiferAction,
	buildCityAction *stdprojAction.BuildCityAction,
	plantGreeneryAction *stdprojAction.PlantGreeneryAction,
	sellPatentsAction *stdprojAction.SellPatentsAction,
	// Resource conversions
	convertHeatAction *resconvAction.ConvertHeatToTemperatureAction,
	convertPlantsAction *resconvAction.ConvertPlantsToGreeneryAction,
	// Tile selection
	selectTileAction *tileAction.SelectTileAction,
	// Turn management
	startGameAction *turnAction.StartGameAction,
	skipActionAction *turnAction.SkipActionAction,
	selectStartingCardsAction *turnAction.SelectStartingCardsAction,
	// Confirmations
	confirmSellPatentsAction *confirmAction.ConfirmSellPatentsAction,
	confirmProductionCardsAction *confirmAction.ConfirmProductionCardsAction,
	confirmCardDrawAction *confirmAction.ConfirmCardDrawAction,
	// Connection
	playerReconnectedAction *connAction.PlayerReconnectedAction,
	playerDisconnectedAction *connAction.PlayerDisconnectedAction,
	// Admin actions
	adminSetPhaseAction *adminAction.SetPhaseAction,
	adminSetCurrentTurnAction *adminAction.SetCurrentTurnAction,
	adminSetResourcesAction *adminAction.SetResourcesAction,
	adminSetProductionAction *adminAction.SetProductionAction,
	adminSetGlobalParametersAction *adminAction.SetGlobalParametersAction,
	adminGiveCardAction *adminAction.GiveCardAction,
	adminSetCorporationAction *adminAction.SetCorporationAction,
	adminStartTileSelectionAction *adminAction.StartTileSelectionAction,
) {
	log := logger.Get()
	log.Info("🔄 Registering migration handlers with explicit broadcasting")

	// ========== Game Lifecycle ==========
	createGameHandler := game.NewCreateGameHandler(createGameAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeCreateGame, createGameHandler)

	joinGameHandler := game.NewJoinGameHandler(joinGameAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypePlayerConnect, joinGameHandler) // Primary handler
	hub.RegisterHandler(dto.MessageTypeJoinGame, joinGameHandler)      // Alternative for backwards compatibility

	// ========== Card Actions ==========
	playCardHandler := card.NewPlayCardHandler(playCardAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionPlayCard, playCardHandler)

	useCardActionHandler := card.NewUseCardActionHandler(useCardActionAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionCardAction, useCardActionHandler)

	// ========== Standard Projects ==========
	launchAsteroidHandler := standard_project.NewLaunchAsteroidHandler(launchAsteroidAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionLaunchAsteroid, launchAsteroidHandler)

	buildPowerPlantHandler := standard_project.NewBuildPowerPlantHandler(buildPowerPlantAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionBuildPowerPlant, buildPowerPlantHandler)

	buildAquiferHandler := standard_project.NewBuildAquiferHandler(buildAquiferAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionBuildAquifer, buildAquiferHandler)

	buildCityHandler := standard_project.NewBuildCityHandler(buildCityAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionBuildCity, buildCityHandler)

	plantGreeneryHandler := standard_project.NewPlantGreeneryHandler(plantGreeneryAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionPlantGreenery, plantGreeneryHandler)

	sellPatentsHandler := standard_project.NewSellPatentsHandler(sellPatentsAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionSellPatents, sellPatentsHandler)

	// ========== Resource Conversions ==========
	convertHeatHandler := resource_conversion.NewConvertHeatHandler(convertHeatAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConvertHeatToTemperature, convertHeatHandler)

	convertPlantsHandler := resource_conversion.NewConvertPlantsHandler(convertPlantsAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConvertPlantsToGreenery, convertPlantsHandler)

	// ========== Tile Selection ==========
	selectTileHandler := tile.NewSelectTileHandler(selectTileAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionTileSelected, selectTileHandler)

	// ========== Turn Management ==========
	startGameHandler := turn_management.NewStartGameHandler(startGameAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionStartGame, startGameHandler)

	skipActionHandler := turn_management.NewSkipActionHandler(skipActionAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionSkipAction, skipActionHandler)

	selectStartingCardsHandler := turn_management.NewSelectStartingCardsHandler(selectStartingCardsAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionSelectStartingCard, selectStartingCardsHandler)

	// ========== Confirmations ==========
	confirmSellPatentsHandler := confirmation.NewConfirmSellPatentsHandler(confirmSellPatentsAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConfirmSellPatents, confirmSellPatentsHandler)

	confirmProductionCardsHandler := confirmation.NewConfirmProductionCardsHandler(confirmProductionCardsAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConfirmProductionCards, confirmProductionCardsHandler)

	confirmCardDrawHandler := confirmation.NewConfirmCardDrawHandler(confirmCardDrawAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionCardDrawConfirmed, confirmCardDrawHandler)

	// ========== Connection Management ==========
	// NOTE: PlayerReconnectedHandler is NOT registered separately because:
	// - JoinGameHandler (on 'player-connect') handles BOTH new joins AND reconnections
	// - It checks for playerID in payload to determine if it's a reconnect
	// - MessageTypePlayerReconnected is a SERVER->CLIENT response type, not CLIENT->SERVER request
	// If reconnection logic needs to be different, integrate it into JoinGameHandler
	_ = playerReconnectedAction // Keep action available for future use

	playerDisconnectedHandler := connection.NewPlayerDisconnectedHandler(playerDisconnectedAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypePlayerDisconnected, playerDisconnectedHandler)

	// ========== Admin Commands (Development Mode) ==========
	adminCommandHandler := admin.NewAdminCommandHandler(
		adminSetPhaseAction,
		adminSetCurrentTurnAction,
		adminSetResourcesAction,
		adminSetProductionAction,
		adminSetGlobalParametersAction,
		adminGiveCardAction,
		adminSetCorporationAction,
		adminStartTileSelectionAction,
		broadcaster,
	)
	hub.RegisterHandler(dto.MessageTypeAdminCommand, adminCommandHandler)

	log.Info("🎯 Migration handlers registered successfully")
	log.Info("   ✅ Game Lifecycle (2): create-game, player-connect/join-game (both supported)")
	log.Info("   ✅ Card Actions (2): PlayCard, UseCardAction")
	log.Info("   ✅ Standard Projects (6): LaunchAsteroid, BuildPowerPlant, BuildAquifer, BuildCity, PlantGreenery, SellPatents")
	log.Info("   ✅ Resource Conversions (2): ConvertHeat, ConvertPlants")
	log.Info("   ✅ Tile Selection (1): SelectTile")
	log.Info("   ✅ Turn Management (3): StartGame, SkipAction, SelectStartingCards")
	log.Info("   ✅ Confirmations (3): ConfirmSellPatents, ConfirmProductionCards, ConfirmCardDraw")
	log.Info("   ✅ Connection (1): PlayerDisconnected")
	log.Info("   ✅ Admin (1): AdminCommand (routes to 8 sub-commands)")
	log.Info("   📌 Total: 21 handlers registered (OLD handlers overwritten)")
}

// MigrateSingleHandler migrates a specific message type from old to new handler
// This allows for gradual, controlled migration of individual handlers
func MigrateSingleHandler(
	hub *core.Hub,
	messageType dto.MessageType,
	newHandler core.MessageHandler,
) {
	log := logger.Get().With(zap.String("message_type", string(messageType)))

	log.Info("🔄 Migrating handler to new architecture")
	hub.RegisterHandler(messageType, newHandler)
	log.Info("✅ Handler migration complete")
}

// NOTE: All handlers have been migrated. The TODO list below is preserved for reference.
// Completed migrations:
// - ConvertHeatHandler ✓
// - ConvertPlantsHandler ✓
// - BuildPowerPlantHandler ✓
// - BuildCityHandler ✓
// - BuildAquiferHandler ✓
// - PlantGreeneryHandler ✓
// - LaunchAsteroidHandler ✓
// - SellPatentsHandler ✓
// - ConfirmSellPatentsHandler ✓
// - SkipActionHandler ✓
// - StartGameHandler ✓
// - SelectStartingCardsHandler ✓
// - ConfirmProductionCardsHandler ✓
// - ConfirmCardDrawHandler ✓
// - PlayerReconnectedHandler ✓ (handled by JoinGameHandler)
// - PlayerDisconnectedHandler ✓
// - Admin handlers (SetPhase, SetResources, SetProduction, SetGlobalParameters, GiveCard, SetCorporation, SetCurrentTurn) ✓
