package websocket

import (
	"terraforming-mars-backend/internal/action"
	adminaction "terraforming-mars-backend/internal/action/admin"
	executecardaction "terraforming-mars-backend/internal/action/execute_card_action"
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
	"terraforming-mars-backend/internal/session"
	sessionGame "terraforming-mars-backend/internal/session/game"
	sessionPlayer "terraforming-mars-backend/internal/session/player"
)

// RegisterHandlers registers all message type handlers with the hub
func RegisterHandlers(
	hub *core.Hub,
	sessionManagerFactory session.SessionManagerFactory,
	newGameRepo sessionGame.Repository,
	newPlayerRepo sessionPlayer.Repository,
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
) {
	parser := utils.NewMessageParser()

	// Register connection handler
	// NEW ARCHITECTURE: Using action pattern for join_game + reconnection + explicit broadcast timing
	// SessionManagerFactory injected so handler can get game-specific broadcaster
	connectionHandler := connect.NewConnectionHandler(hub, sessionManagerFactory, joinGameAction, playerReconnectedAction)
	hub.RegisterHandler(dto.MessageTypePlayerConnect, connectionHandler)

	// Register disconnect handler
	// NEW ARCHITECTURE: Using action pattern for disconnect
	disconnectHandler := disconnect.NewDisconnectHandler(playerDisconnectedAction)
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
	hub.RegisterHandler(dto.MessageTypeActionSkipAction, skip_action.NewHandler(skipActionAction))

	// Register game management handlers
	hub.RegisterHandler(dto.MessageTypeActionStartGame, start_game.NewHandler(startGameAction))

	// Register card selection handlers
	// NEW ARCHITECTURE: Using action pattern for select_starting_card and select_cards
	hub.RegisterHandler(dto.MessageTypeActionSelectStartingCard, select_starting_card.NewHandler(selectStartingCardsAction, parser))
	hub.RegisterHandler(dto.MessageTypeActionSelectCards, select_cards.NewHandler(confirmSellPatentsAction, confirmProductionCardsAction, newPlayerRepo, parser))
	hub.RegisterHandler(dto.MessageTypeActionConfirmProductionCards, confirm_cards.NewHandler(confirmProductionCardsAction, parser))
	hub.RegisterHandler(dto.MessageTypeActionCardDrawConfirmed, card_draw_confirmed.NewHandler(confirmCardDrawAction, parser))

	// Register play card handler
	// NEW ARCHITECTURE: Using action pattern for play_card
	hub.RegisterHandler(dto.MessageTypeActionPlayCard, play_card.NewHandler(playCardAction, parser))

	// Register play card action handler
	// NEW ARCHITECTURE: Using action pattern for card action execution
	hub.RegisterHandler(dto.MessageTypeActionCardAction, play_card_action.NewHandler(executeCardActionAction, parser))

	// Register tile selection handlers
	// NEW ARCHITECTURE: Using action pattern for tile_selected
	hub.RegisterHandler(dto.MessageTypeActionTileSelected, tile_selected.NewHandler(selectTileAction, parser))

	// Register admin command handler WITHOUT middleware (development mode validation is handled internally)
	// NEW ARCHITECTURE: Using admin action pattern for all admin commands
	hub.RegisterHandler(dto.MessageTypeAdminCommand, admin_command.NewHandler(
		newGameRepo,
		giveCardAdminAction,
		setPhaseAdminAction,
		setResourcesAdminAction,
		setProductionAdminAction,
		setGlobalParametersAdminAction,
		startTileSelectionAdminAction,
		setCurrentTurnAdminAction,
		setCorporationAdminAction,
	))
}
