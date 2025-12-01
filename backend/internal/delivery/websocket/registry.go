package websocket

import (
	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
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
// This is a parallel registry for the new architecture, allowing gradual migration
// Eventually this will replace RegisterHandlers entirely
func RegisterHandlers(
	hub *core.Hub,
	// Game lifecycle
	createGameAction *action.CreateGameAction,
	joinGameAction *action.JoinGameAction,
	// Card actions
	playCardAction *action.PlayCardAction,
	useCardActionAction *action.UseCardActionAction,
	// Standard projects
	launchAsteroidAction *action.LaunchAsteroidAction,
	buildPowerPlantAction *action.BuildPowerPlantAction,
	buildAquiferAction *action.BuildAquiferAction,
	buildCityAction *action.BuildCityAction,
	plantGreeneryAction *action.PlantGreeneryAction,
	sellPatentsAction *action.SellPatentsAction,
	// Resource conversions
	convertHeatAction *action.ConvertHeatToTemperatureAction,
	convertPlantsAction *action.ConvertPlantsToGreeneryAction,
	// Tile selection
	selectTileAction *action.SelectTileAction,
	// Turn management
	startGameAction *action.StartGameAction,
	skipActionAction *action.SkipActionAction,
	selectStartingCardsAction *action.SelectStartingCardsAction,
	// Confirmations
	confirmSellPatentsAction *action.ConfirmSellPatentsAction,
	confirmProductionCardsAction *action.ConfirmProductionCardsAction,
	confirmCardDrawAction *action.ConfirmCardDrawAction,
	// Connection
	playerReconnectedAction *action.PlayerReconnectedAction,
	playerDisconnectedAction *action.PlayerDisconnectedAction,
) {
	log := logger.Get()
	log.Info("🔄 Registering migration handlers for new architecture")

	// ========== Game Lifecycle ==========
	createGameHandler := game.NewCreateGameHandler(createGameAction)
	hub.RegisterHandler(dto.MessageTypeCreateGame, createGameHandler)

	joinGameHandler := game.NewJoinGameHandler(joinGameAction)
	hub.RegisterHandler(dto.MessageTypePlayerConnect, joinGameHandler) // Primary handler
	hub.RegisterHandler(dto.MessageTypeJoinGame, joinGameHandler)      // Alternative for backwards compatibility

	// ========== Card Actions ==========
	playCardHandler := card.NewPlayCardHandler(playCardAction)
	hub.RegisterHandler(dto.MessageTypeActionPlayCard, playCardHandler)

	useCardActionHandler := card.NewUseCardActionHandler(useCardActionAction)
	hub.RegisterHandler(dto.MessageTypeActionCardAction, useCardActionHandler)

	// ========== Standard Projects ==========
	launchAsteroidHandler := standard_project.NewLaunchAsteroidHandler(launchAsteroidAction)
	hub.RegisterHandler(dto.MessageTypeActionLaunchAsteroid, launchAsteroidHandler)

	buildPowerPlantHandler := standard_project.NewBuildPowerPlantHandler(buildPowerPlantAction)
	hub.RegisterHandler(dto.MessageTypeActionBuildPowerPlant, buildPowerPlantHandler)

	buildAquiferHandler := standard_project.NewBuildAquiferHandler(buildAquiferAction)
	hub.RegisterHandler(dto.MessageTypeActionBuildAquifer, buildAquiferHandler)

	buildCityHandler := standard_project.NewBuildCityHandler(buildCityAction)
	hub.RegisterHandler(dto.MessageTypeActionBuildCity, buildCityHandler)

	plantGreeneryHandler := standard_project.NewPlantGreeneryHandler(plantGreeneryAction)
	hub.RegisterHandler(dto.MessageTypeActionPlantGreenery, plantGreeneryHandler)

	sellPatentsHandler := standard_project.NewSellPatentsHandler(sellPatentsAction)
	hub.RegisterHandler(dto.MessageTypeActionSellPatents, sellPatentsHandler)

	// ========== Resource Conversions ==========
	convertHeatHandler := resource_conversion.NewConvertHeatHandler(convertHeatAction)
	hub.RegisterHandler(dto.MessageTypeActionConvertHeatToTemperature, convertHeatHandler)

	convertPlantsHandler := resource_conversion.NewConvertPlantsHandler(convertPlantsAction)
	hub.RegisterHandler(dto.MessageTypeActionConvertPlantsToGreenery, convertPlantsHandler)

	// ========== Tile Selection ==========
	selectTileHandler := tile.NewSelectTileHandler(selectTileAction)
	hub.RegisterHandler(dto.MessageTypeActionTileSelected, selectTileHandler)

	// ========== Turn Management ==========
	startGameHandler := turn_management.NewStartGameHandler(startGameAction)
	hub.RegisterHandler(dto.MessageTypeActionStartGame, startGameHandler)

	skipActionHandler := turn_management.NewSkipActionHandler(skipActionAction)
	hub.RegisterHandler(dto.MessageTypeActionSkipAction, skipActionHandler)

	selectStartingCardsHandler := turn_management.NewSelectStartingCardsHandler(selectStartingCardsAction)
	hub.RegisterHandler(dto.MessageTypeActionSelectStartingCard, selectStartingCardsHandler)

	// ========== Confirmations ==========
	confirmSellPatentsHandler := confirmation.NewConfirmSellPatentsHandler(confirmSellPatentsAction)
	hub.RegisterHandler(dto.MessageTypeActionConfirmSellPatents, confirmSellPatentsHandler)

	confirmProductionCardsHandler := confirmation.NewConfirmProductionCardsHandler(confirmProductionCardsAction)
	hub.RegisterHandler(dto.MessageTypeActionConfirmProductionCards, confirmProductionCardsHandler)

	confirmCardDrawHandler := confirmation.NewConfirmCardDrawHandler(confirmCardDrawAction)
	hub.RegisterHandler(dto.MessageTypeActionCardDrawConfirmed, confirmCardDrawHandler)

	// ========== Connection Management ==========
	// NOTE: PlayerReconnectedHandler is NOT registered separately because:
	// - JoinGameHandler (on 'player-connect') handles BOTH new joins AND reconnections
	// - It checks for playerID in payload to determine if it's a reconnect
	// - MessageTypePlayerReconnected is a SERVER->CLIENT response type, not CLIENT->SERVER request
	// If reconnection logic needs to be different, integrate it into JoinGameHandler
	_ = playerReconnectedAction // Keep action available for future use

	playerDisconnectedHandler := connection.NewPlayerDisconnectedHandler(playerDisconnectedAction)
	hub.RegisterHandler(dto.MessageTypePlayerDisconnected, playerDisconnectedHandler)

	log.Info("🎯 Migration handlers registered successfully")
	log.Info("   ✅ Game Lifecycle (2): create-game, player-connect/join-game (both supported)")
	log.Info("   ✅ Card Actions (2): PlayCard, UseCardAction")
	log.Info("   ✅ Standard Projects (6): LaunchAsteroid, BuildPowerPlant, BuildAquifer, BuildCity, PlantGreenery, SellPatents")
	log.Info("   ✅ Resource Conversions (2): ConvertHeat, ConvertPlants")
	log.Info("   ✅ Tile Selection (1): SelectTile")
	log.Info("   ✅ Turn Management (3): StartGame, SkipAction, SelectStartingCards")
	log.Info("   ✅ Confirmations (3): ConfirmSellPatents, ConfirmProductionCards, ConfirmCardDraw")
	log.Info("   ✅ Connection (1): PlayerDisconnected")
	log.Info("   📌 Total: 20 handlers registered (OLD handlers overwritten)")
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

// TODO: Add more migration handlers as they are implemented:
// - ConvertHeatHandler
// - ConvertPlantsHandler
// - BuildPowerPlantHandler
// - BuildCityHandler
// - BuildAquiferHandler
// - PlantGreeneryHandler
// - LaunchAsteroidHandler
// - SellPatentsHandler
// - ConfirmSellPatentsHandler
// - SkipActionHandler
// - StartGameHandler
// - SelectStartingCardsHandler
// - ConfirmProductionCardsHandler
// - ConfirmCardDrawHandler
// - PlayerReconnectedHandler
// - PlayerDisconnectedHandler
// - Admin handlers (SetPhase, SetResources, SetProduction, SetGlobalParameters)
