package actions

import (
	"context"
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