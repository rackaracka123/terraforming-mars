package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	gamePackage "terraforming-mars-backend/internal/session/game"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// ListGamesAction handles the query for listing all games
type ListGamesAction struct {
	action.BaseAction
	gameRepo game.Repository
}

// NewListGamesAction creates a new list games query action
func NewListGamesAction(
	gameRepo game.Repository,
) *ListGamesAction {
	return &ListGamesAction{
		BaseAction: action.NewBaseAction(nil),
		gameRepo:   gameRepo,
	}
}

// Execute performs the list games query
func (a *ListGamesAction) Execute(ctx context.Context, statusStr string) ([]*gamePackage.Game, error) {
	log := a.GetLogger()
	log.Info("🔍 Querying all games",
		zap.String("status_filter", statusStr))

	// Convert string to GameStatus type
	status := types.GameStatus(statusStr)

	// List games from repository
	games, err := a.gameRepo.List(ctx, status)
	if err != nil {
		log.Error("Failed to list games", zap.Error(err))
		return nil, err
	}

	log.Info("✅ Games query completed",
		zap.Int("count", len(games)))

	return games, nil
}
