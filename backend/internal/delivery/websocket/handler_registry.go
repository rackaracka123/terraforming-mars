package websocket

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handlers"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
)

// RegisterHandlers registers all message type handlers with the hub
func RegisterHandlers(hub *core.Hub, gameService service.GameService, playerService service.PlayerService, standardProjectService service.StandardProjectService, cardService service.CardService, gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) {
	parser := utils.NewMessageParser()
	broadcaster := hub.GetBroadcaster()
	manager := hub.GetManager()

	// Create consolidated handlers
	connectionHandler := handlers.NewConnectionHandler(gameService, playerService, broadcaster, manager)
	gameHandler := handlers.NewGameHandler(gameService)
	playerHandler := handlers.NewPlayerHandler(gameService, playerService, standardProjectService, cardService, broadcaster, parser)

	// Register connection handler
	hub.RegisterHandler(dto.MessageTypePlayerConnect, connectionHandler)

	// Register game management handler
	hub.RegisterHandler(dto.MessageTypeActionStartGame, gameHandler)

	// Register all player actions with the player handler
	hub.RegisterHandler(dto.MessageTypeActionLaunchAsteroid, playerHandler)
	hub.RegisterHandler(dto.MessageTypeActionSellPatents, playerHandler)
	hub.RegisterHandler(dto.MessageTypeActionBuildPowerPlant, playerHandler)
	hub.RegisterHandler(dto.MessageTypeActionBuildAquifer, playerHandler)
	hub.RegisterHandler(dto.MessageTypeActionPlantGreenery, playerHandler)
	hub.RegisterHandler(dto.MessageTypeActionBuildCity, playerHandler)
	hub.RegisterHandler(dto.MessageTypeActionPlayCard, playerHandler)
	hub.RegisterHandler(dto.MessageTypeActionSelectStartingCard, playerHandler)
	hub.RegisterHandler(dto.MessageTypeActionSelectCards, playerHandler)
	hub.RegisterHandler(dto.MessageTypeActionSkipAction, playerHandler)
}