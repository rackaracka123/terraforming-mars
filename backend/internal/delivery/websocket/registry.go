package websocket

import (
	"terraforming-mars-backend/internal/action"
	adminaction "terraforming-mars-backend/internal/action/admin"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_aquifer"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_city"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_power_plant"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/confirm_sell_patents"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/convert_heat_to_temperature"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/convert_plants_to_greenery"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/launch_asteroid"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/plant_greenery"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/play_card"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/play_card_action"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/sell_patents"
	"terraforming-mars-backend/internal/delivery/websocket/handler/admin/admin_command"
	"terraforming-mars-backend/internal/delivery/websocket/handler/card_selection/card_draw_confirmed"
	"terraforming-mars-backend/internal/delivery/websocket/handler/card_selection/select_cards"
	"terraforming-mars-backend/internal/delivery/websocket/handler/card_selection/select_starting_card"
	"terraforming-mars-backend/internal/delivery/websocket/handler/connect"
	"terraforming-mars-backend/internal/delivery/websocket/handler/disconnect"
	"terraforming-mars-backend/internal/delivery/websocket/handler/game/skip_action"
	"terraforming-mars-backend/internal/delivery/websocket/handler/game/start_game"
	"terraforming-mars-backend/internal/delivery/websocket/handler/production/confirm_cards"
	"terraforming-mars-backend/internal/delivery/websocket/handler/tile_selection/tile_selected"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/session"
	sessionGame "terraforming-mars-backend/internal/session/game"
)

// RegisterHandlers registers all message type handlers with the hub
func RegisterHandlers(
	hub *core.Hub,
	sessionManager session.SessionManager,
	gameService service.GameService,
	playerService service.PlayerService,
	standardProjectService service.StandardProjectService,
	cardService service.CardService,
	adminService service.AdminService,
	resourceConversionService service.ResourceConversionService,
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardRepo repository.CardRepository,
	newGameRepo sessionGame.Repository,
	eventBus *events.EventBusImpl,
	startGameAction *action.StartGameAction,
	joinGameAction *action.JoinGameAction,
	selectStartingCardsAction *action.SelectStartingCardsAction,
	skipActionAction *action.SkipActionAction,
	confirmProductionCardsAction *action.ConfirmProductionCardsAction,
	buildCityAction *action.BuildCityAction,
	selectTileAction *action.SelectTileAction,
	playCardAction *action.PlayCardAction,
	launchAsteroidAction *action.LaunchAsteroidAction,
	buildPowerPlantAction *action.BuildPowerPlantAction,
	buildAquiferAction *action.BuildAquiferAction,
	plantGreeneryAction *action.PlantGreeneryAction,
	sellPatentsAction *action.SellPatentsAction,
	confirmSellPatentsAction *action.ConfirmSellPatentsAction,
	convertHeatAction *action.ConvertHeatToTemperatureAction,
	convertPlantsAction *action.ConvertPlantsToGreeneryAction,
	confirmCardDrawAction *action.ConfirmCardDrawAction,
	startTileSelectionAdminAction *adminaction.StartTileSelectionAction,
) {
	parser := utils.NewMessageParser()

	// Register connection handler
	// NEW ARCHITECTURE: Using action pattern for join_game + explicit broadcast timing
	// SessionManager injected so handler can control broadcast timing after join completes
	connectionHandler := connect.NewConnectionHandler(hub, sessionManager, gameService, playerService, joinGameAction)
	hub.RegisterHandler(dto.MessageTypePlayerConnect, connectionHandler)

	// Register disconnect handler
	disconnectHandler := disconnect.NewDisconnectHandler(playerService)
	hub.RegisterHandler(dto.MessageTypePlayerDisconnected, disconnectHandler)

	// Register standard project handlers
	// NEW ARCHITECTURE: Using action pattern for standard projects
	hub.RegisterHandler(dto.MessageTypeActionLaunchAsteroid, launch_asteroid.NewHandler(launchAsteroidAction))
	hub.RegisterHandler(dto.MessageTypeActionSellPatents, sell_patents.NewHandler(sellPatentsAction))
	hub.RegisterHandler(dto.MessageTypeActionConfirmSellPatents, confirm_sell_patents.NewHandler(confirmSellPatentsAction))
	hub.RegisterHandler(dto.MessageTypeActionBuildPowerPlant, build_power_plant.NewHandler(buildPowerPlantAction))
	hub.RegisterHandler(dto.MessageTypeActionBuildAquifer, build_aquifer.NewHandler(buildAquiferAction))
	hub.RegisterHandler(dto.MessageTypeActionPlantGreenery, plant_greenery.NewHandler(plantGreeneryAction))
	hub.RegisterHandler(dto.MessageTypeActionBuildCity, build_city.NewHandler(buildCityAction))

	// Register resource conversion handlers
	// NEW ARCHITECTURE: Using action pattern for resource conversions
	hub.RegisterHandler(dto.MessageTypeActionConvertPlantsToGreenery, convert_plants_to_greenery.NewHandler(convertPlantsAction))
	hub.RegisterHandler(dto.MessageTypeActionConvertHeatToTemperature, convert_heat_to_temperature.NewHandler(convertHeatAction))

	// Register skip action handler
	// NEW ARCHITECTURE: Using action pattern for skip_action
	hub.RegisterHandler(dto.MessageTypeActionSkipAction, skip_action.NewHandlerWithAction(skipActionAction))

	// Register game management handlers
	// NEW ARCHITECTURE: Using action pattern for start_game
	hub.RegisterHandler(dto.MessageTypeActionStartGame, start_game.NewHandlerWithAction(startGameAction))

	// Register card selection handlers
	// NEW ARCHITECTURE: Using action pattern for select_starting_card
	hub.RegisterHandler(dto.MessageTypeActionSelectStartingCard, select_starting_card.NewHandler(selectStartingCardsAction, parser))
	hub.RegisterHandler(dto.MessageTypeActionSelectCards, select_cards.NewHandler(cardService, gameService, standardProjectService, playerRepo, parser))
	hub.RegisterHandler(dto.MessageTypeActionConfirmProductionCards, confirm_cards.NewHandler(confirmProductionCardsAction, parser))
	hub.RegisterHandler(dto.MessageTypeActionCardDrawConfirmed, card_draw_confirmed.NewHandler(confirmCardDrawAction, parser))

	// Register play card handler
	// NEW ARCHITECTURE: Using action pattern for play_card
	hub.RegisterHandler(dto.MessageTypeActionPlayCard, play_card.NewHandler(playCardAction, parser))

	// Register play card action handler
	hub.RegisterHandler(dto.MessageTypeActionCardAction, play_card_action.NewHandler(cardService, parser))

	// Register tile selection handlers
	// NEW ARCHITECTURE: Using action pattern for tile_selected
	hub.RegisterHandler(dto.MessageTypeActionTileSelected, tile_selected.NewHandler(selectTileAction, parser))

	// Register admin command handler WITHOUT middleware (development mode validation is handled internally)
	hub.RegisterHandler(dto.MessageTypeAdminCommand, admin_command.NewHandler(newGameRepo, gameService, playerService, cardService, adminService, startTileSelectionAdminAction))
}
