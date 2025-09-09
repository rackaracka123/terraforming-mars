package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"
)

// GameActions handles game lifecycle actions
type GameActions struct {
	gameService service.GameService
}

// NewGameActions creates a new game actions handler
func NewGameActions(gameService service.GameService) *GameActions {
	return &GameActions{
		gameService: gameService,
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
		h.logger.Info("🏭 All players passed - triggering production phase",
			zap.String("game_id", gameID),
			zap.Int("generation", result.Game.Generation))

		// Trigger production phase broadcast in a separate goroutine
		go func() {
			if err := h.BroadcastProductionPhaseStarted(ctx, gameID); err != nil {
				h.logger.Error("Failed to broadcast production phase started",
					zap.String("game_id", gameID),
					zap.Error(err))
			}
		}()
	}

	return nil

	return ga.gameService.SkipPlayerTurn(ctx, gameID, playerID)
}
