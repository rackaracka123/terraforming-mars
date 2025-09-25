package websocket

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_aquifer"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_city"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_power_plant"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/launch_asteroid"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/plant_greenery"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/play_card"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/sell_patents"
	"terraforming-mars-backend/internal/delivery/websocket/handler/card_selection/select_cards"
	"terraforming-mars-backend/internal/delivery/websocket/handler/card_selection/select_starting_card"
	"terraforming-mars-backend/internal/delivery/websocket/handler/connect"
	"terraforming-mars-backend/internal/delivery/websocket/handler/game/skip_action"
	"terraforming-mars-backend/internal/delivery/websocket/handler/game/start_game"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
)

// RegisterHandlers registers all message type handlers with the hub
func RegisterHandlers(hub *core.Hub, gameService service.GameService, playerService service.PlayerService, standardProjectService service.StandardProjectService, cardService service.CardService, gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) {
	parser := utils.NewMessageParser()

	// Register connection handler
	connectionHandler := connect.NewConnectionHandler(gameService, playerService)
	hub.RegisterHandler(dto.MessageTypePlayerConnect, connectionHandler)

	// Register standard project handlers (validation moved to individual handlers)
	hub.RegisterHandler(dto.MessageTypeActionLaunchAsteroid, launch_asteroid.NewHandler(standardProjectService))
	hub.RegisterHandler(dto.MessageTypeActionSellPatents, sell_patents.NewHandler(standardProjectService, parser))
	hub.RegisterHandler(dto.MessageTypeActionBuildPowerPlant, build_power_plant.NewHandler(standardProjectService))
	hub.RegisterHandler(dto.MessageTypeActionBuildAquifer, build_aquifer.NewHandler(standardProjectService, parser))
	hub.RegisterHandler(dto.MessageTypeActionPlantGreenery, plant_greenery.NewHandler(standardProjectService, parser))
	hub.RegisterHandler(dto.MessageTypeActionBuildCity, build_city.NewHandler(standardProjectService, parser))

	// Register skip action handler
	hub.RegisterHandler(dto.MessageTypeActionSkipAction, skip_action.NewHandler(gameService, playerService))

	// Register game management handlers
	hub.RegisterHandler(dto.MessageTypeActionStartGame, start_game.NewHandler(gameService))

	// Register card selection handlers
	hub.RegisterHandler(dto.MessageTypeActionSelectStartingCard, select_starting_card.NewHandler(cardService, gameService, parser))
	hub.RegisterHandler(dto.MessageTypeActionSelectCards, select_cards.NewHandler(cardService, gameService, parser))

	// Register play card handler
	hub.RegisterHandler(dto.MessageTypeActionPlayCard, play_card.NewHandler(cardService, parser))
}

