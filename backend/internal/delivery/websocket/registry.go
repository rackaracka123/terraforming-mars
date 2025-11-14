package websocket

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_aquifer"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_city"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_power_plant"
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
	"terraforming-mars-backend/internal/delivery/websocket/handler/tile_selection/tile_selected"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/actions"
	"terraforming-mars-backend/internal/actions/card_selection"
	"terraforming-mars-backend/internal/actions/standard_projects"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/internal/service"
)

// RegisterHandlers registers all message type handlers with the hub
func RegisterHandlers(
	hub *core.Hub,
	gameService service.GameService,
	lobbyService lobby.Service,
	playerService service.PlayerService,
	standardProjectService service.StandardProjectService,
	cardService service.CardService,
	adminService service.AdminService,
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardRepo game.CardRepository,
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
) {
	parser := utils.NewMessageParser()

	// Register connection handler (uses lobby for joining, game for reconnection)
	connectionHandler := connect.NewConnectionHandler(lobbyService, gameService, playerService)
	hub.RegisterHandler(dto.MessageTypePlayerConnect, connectionHandler)

	// Register disconnect handler
	disconnectHandler := disconnect.NewDisconnectHandler(playerService)
	hub.RegisterHandler(dto.MessageTypePlayerDisconnected, disconnectHandler)

	// Register standard project handlers
	hub.RegisterHandler(dto.MessageTypeActionLaunchAsteroid, launch_asteroid.NewHandler(launchAsteroidAction, parser))
	hub.RegisterHandler(dto.MessageTypeActionSellPatents, sell_patents.NewHandler(sellPatentsAction, parser))
	hub.RegisterHandler(dto.MessageTypeActionBuildPowerPlant, build_power_plant.NewHandler(buildPowerPlantAction))
	hub.RegisterHandler(dto.MessageTypeActionBuildAquifer, build_aquifer.NewHandler(buildAquiferAction, parser))
	hub.RegisterHandler(dto.MessageTypeActionPlantGreenery, plant_greenery.NewHandler(plantGreeneryAction, parser))
	hub.RegisterHandler(dto.MessageTypeActionBuildCity, build_city.NewHandler(buildCityAction, parser))

	// Register resource conversion handlers
	hub.RegisterHandler(dto.MessageTypeActionConvertPlantsToGreenery, convert_plants_to_greenery.NewHandler(convertPlantsAction))
	hub.RegisterHandler(dto.MessageTypeActionConvertHeatToTemperature, convert_heat_to_temperature.NewHandler(convertHeatAction))

	// Register skip action handler
	hub.RegisterHandler(dto.MessageTypeActionSkipAction, skip_action.NewHandler(skipAction))

	// Register game management handlers
	hub.RegisterHandler(dto.MessageTypeActionStartGame, start_game.NewHandler(lobbyService))

	// Register card selection handlers
	hub.RegisterHandler(dto.MessageTypeActionSelectStartingCard, select_starting_card.NewHandler(selectStartingCardsAction, parser))
	hub.RegisterHandler(dto.MessageTypeActionSelectCards, select_cards.NewHandler(submitSellPatentsAction, selectProductionCardsAction, playerRepo, parser))
	hub.RegisterHandler(dto.MessageTypeActionCardDrawConfirmed, card_draw_confirmed.NewHandler(confirmCardDrawAction, parser))

	// Register play card handler
	hub.RegisterHandler(dto.MessageTypeActionPlayCard, play_card.NewHandler(playCardAction, parser))

	// Register play card action handler
	hub.RegisterHandler(dto.MessageTypeActionCardAction, play_card_action.NewHandler(playCardActionAction, parser))

	// Register tile selection handlers
	hub.RegisterHandler(dto.MessageTypeActionTileSelected, tile_selected.NewHandler(selectTileAction, parser))

	// Register admin command handler WITHOUT middleware (development mode validation is handled internally)
	hub.RegisterHandler(dto.MessageTypeAdminCommand, admin_command.NewHandler(gameService, playerService, cardService, adminService))
}
