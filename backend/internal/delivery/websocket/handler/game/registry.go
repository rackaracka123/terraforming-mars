package game

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/game/skip_action"
	"terraforming-mars-backend/internal/delivery/websocket/handler/game/start_game"
	websocketmiddleware "terraforming-mars-backend/internal/middleware/websocket"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/transaction"
)

// SetupGameRegistry registers all game management handlers with the registry
func SetupGameRegistry(
	gameService service.GameService,
	playerService service.PlayerService,
	broadcaster *core.Broadcaster,
	transactionManager *transaction.Manager,
) *core.ActionRegistry {
	registry := core.NewActionRegistry()

	// Create turn validation middleware for skip action
	turnValidator := websocketmiddleware.CreateTurnValidatorMiddleware(transactionManager)

	// Register game management actions
	// start_game doesn't need turn validation (happens during lobby phase)
	registry.Register(dto.ActionTypeStartGame, start_game.NewHandler(gameService))

	// skip_action needs turn validation - only current player can skip their turn
	registry.Register(dto.ActionTypeSkipAction,
		websocketmiddleware.WrapWithMiddleware(
			skip_action.NewHandler(gameService, playerService, broadcaster),
			turnValidator,
		))

	return registry
}
