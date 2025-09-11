package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/core/broadcast"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"
)

// SkipActions handles skip turn and production-related actions
type SkipActions struct {
	gameService   service.GameService
	playerService service.PlayerService
	broadcaster   *broadcast.Broadcaster
}

// NewSkipActions creates a new skip actions handler
func NewSkipActions(gameService service.GameService, playerService service.PlayerService, broadcaster *broadcast.Broadcaster) *SkipActions {
	return &SkipActions{
		gameService:   gameService,
		playerService: playerService,
		broadcaster:   broadcaster,
	}
}

// SkipAction handles the skip action with production phase triggering
func (sa *SkipActions) SkipAction(ctx context.Context, gameID, playerID string) error {
	game, err := sa.gameService.GetGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state for phase validation: %w", err)
	}

	if game.CurrentPhase != model.GamePhaseAction {
		return fmt.Errorf("skip action not allowed in phase %s, must be in action phase", game.CurrentPhase)
	}

	result, err := sa.gameService.SkipPlayerTurn(ctx, gameID, playerID)
	if err != nil {
		return err
	}

	// Check if all players have passed - trigger production phase
	if result.AllPlayersPassed {
		if err := sa.triggerProductionPhase(ctx, gameID); err != nil {
			return fmt.Errorf("failed to trigger production phase: %w", err)
		}
	}

	return nil
}

// triggerProductionPhase handles the transition to production phase and broadcasts production data
func (sa *SkipActions) triggerProductionPhase(ctx context.Context, gameID string) error {
	_, err := sa.gameService.ExecuteProductionPhase(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to execute production phase: %w", err)
	}

	// Note: The production data broadcasting should be handled by the service layer
	// via events rather than directly in the delivery layer. This is a temporary approach.
	// TODO: Remove this when proper event-driven production broadcasting is implemented.

	return nil
}
