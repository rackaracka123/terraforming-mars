package action

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_aquifer"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_city"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_power_plant"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/launch_asteroid"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/plant_greenery"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/select_cards"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/select_starting_card"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/sell_patents"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/skip_action"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/start_game"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/service"
)

// SetupActionRegistry registers all action handlers with the registry
func SetupActionRegistry(
	gameService service.GameService,
	playerService service.PlayerService,
	standardProjectService service.StandardProjectService,
	cardService service.CardService,
	broadcaster *core.Broadcaster,
) *core.ActionRegistry {
	registry := core.NewActionRegistry()
	parser := utils.NewMessageParser()

	// Register game actions
	registry.Register(dto.ActionTypeStartGame, start_game.NewHandler(gameService))
	registry.Register(dto.ActionTypeSkipAction, skip_action.NewHandler(gameService, playerService, broadcaster))

	// Register card actions
	registry.Register(dto.ActionTypeSelectStartingCard, select_starting_card.NewHandler(cardService, gameService, parser))
	registry.Register(dto.ActionTypeSelectCards, select_cards.NewHandler(cardService, gameService, parser))

	// Register standard project actions
	registry.Register(dto.ActionTypeSellPatents, sell_patents.NewHandler(standardProjectService, parser))
	registry.Register(dto.ActionTypeBuildPowerPlant, build_power_plant.NewHandler(standardProjectService))
	registry.Register(dto.ActionTypeLaunchAsteroid, launch_asteroid.NewHandler(standardProjectService))
	registry.Register(dto.ActionTypeBuildAquifer, build_aquifer.NewHandler(standardProjectService, parser))
	registry.Register(dto.ActionTypePlantGreenery, plant_greenery.NewHandler(standardProjectService, parser))
	registry.Register(dto.ActionTypeBuildCity, build_city.NewHandler(standardProjectService, parser))

	return registry
}
