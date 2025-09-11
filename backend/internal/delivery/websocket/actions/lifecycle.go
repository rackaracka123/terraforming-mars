package actions

import (
	"context"

	"terraforming-mars-backend/internal/service"
)

// LifecycleActions handles basic game lifecycle actions (start game)
type LifecycleActions struct {
	gameService service.GameService
}

// NewLifecycleActions creates a new lifecycle actions handler
func NewLifecycleActions(gameService service.GameService) *LifecycleActions {
	return &LifecycleActions{
		gameService: gameService,
	}
}

// StartGame handles the start game action
func (la *LifecycleActions) StartGame(ctx context.Context, gameID, playerID string) error {
	return la.gameService.StartGame(ctx, gameID, playerID)
}
