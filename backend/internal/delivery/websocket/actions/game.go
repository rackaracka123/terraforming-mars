package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"
)

// GameActions handles game lifecycle actions
type GameActions struct {
	gameService   service.GameService
	playerService service.PlayerService
	broadcaster   *core.Broadcaster
}

// NewGameActions creates a new game actions handler
func NewGameActions(gameService service.GameService, playerService service.PlayerService, broadcaster *core.Broadcaster) *GameActions {
	return &GameActions{
		gameService:   gameService,
		playerService: playerService,
		broadcaster:   broadcaster,
	}
}

// StartGame handles the start game action
func (ga *GameActions) StartGame(ctx context.Context, gameID, playerID string) error {
	return ga.gameService.StartGame(ctx, gameID, playerID)
}

// SkipAction handles the skip action
func (ga *GameActions) SkipAction(ctx context.Context, gameID, playerID string) error {
	game, err := ga.gameService.GetGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state for phase validation: %w", err)
	}

	if game.CurrentPhase != model.GamePhaseAction {
		return fmt.Errorf("skip action not allowed in phase %s, must be in action phase", game.CurrentPhase)
	}

	result, err := ga.gameService.SkipPlayerTurn(ctx, gameID, playerID)
	if err != nil {
		return err
	}

	// Check if all players have passed - trigger production phase
	if result.AllPlayersPassed {
		_, err := ga.gameService.ExecuteProductionPhase(ctx, gameID)
		if err != nil {
			return fmt.Errorf("failed to execute production phase: %w", err)
		}

		// Note: Production data broadcasting should be handled by the service layer
		// via events rather than directly in the delivery layer. The service should
		// emit production phase events that the WebSocket event handlers pick up.
		// TODO: Remove direct broadcasting when proper event-driven architecture is implemented.
	}

	return nil
}

