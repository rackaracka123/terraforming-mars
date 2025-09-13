package action

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_aquifer"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_city"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/build_power_plant"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/launch_asteroid"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/plant_greenery"
	"terraforming-mars-backend/internal/delivery/websocket/handler/action/sell_patents"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	websocketmiddleware "terraforming-mars-backend/internal/middleware/websocket"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/transaction"
)

// SetupActionRegistry registers all game action handlers with the registry (standard projects only)
func SetupActionRegistry(
	standardProjectService service.StandardProjectService,
	transactionManager *transaction.Manager,
) *core.ActionRegistry {
	registry := core.NewActionRegistry()
	parser := utils.NewMessageParser()

	// Create the middleware that validates turn + actions
	turnValidator := websocketmiddleware.CreateTurnValidatorMiddleware(transactionManager)
	actionValidator := websocketmiddleware.CreateActionValidatorMiddleware(transactionManager)

	// Chain both middleware together - turn validation + action consumption
	chainedMiddleware := websocketmiddleware.ChainMiddleware(turnValidator, actionValidator)

	// Register standard project actions with middleware
	// Each action gets wrapped with turn validation + action consumption middleware

	registry.Register(dto.ActionTypeSellPatents,
		websocketmiddleware.WrapWithMiddleware(
			sell_patents.NewHandler(standardProjectService, parser),
			chainedMiddleware,
		))

	registry.Register(dto.ActionTypeBuildPowerPlant,
		websocketmiddleware.WrapWithMiddleware(
			build_power_plant.NewHandler(standardProjectService),
			chainedMiddleware,
		))

	registry.Register(dto.ActionTypeLaunchAsteroid,
		websocketmiddleware.WrapWithMiddleware(
			launch_asteroid.NewHandler(standardProjectService),
			chainedMiddleware,
		))

	registry.Register(dto.ActionTypeBuildAquifer,
		websocketmiddleware.WrapWithMiddleware(
			build_aquifer.NewHandler(standardProjectService, parser),
			chainedMiddleware,
		))

	registry.Register(dto.ActionTypePlantGreenery,
		websocketmiddleware.WrapWithMiddleware(
			plant_greenery.NewHandler(standardProjectService, parser),
			chainedMiddleware,
		))

	registry.Register(dto.ActionTypeBuildCity,
		websocketmiddleware.WrapWithMiddleware(
			build_city.NewHandler(standardProjectService, parser),
			chainedMiddleware,
		))

	return registry
}
