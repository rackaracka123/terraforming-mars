package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// ListGamesAction handles the query for listing all games
type ListGamesAction struct {
	action.BaseAction
	oldGameRepo repository.GameRepository
}

// NewListGamesAction creates a new list games query action
func NewListGamesAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	oldGameRepo repository.GameRepository,
	sessionMgr session.SessionManager,
) *ListGamesAction {
	return &ListGamesAction{
		BaseAction:  action.NewBaseAction(gameRepo, playerRepo, sessionMgr),
		oldGameRepo: oldGameRepo,
	}
}

// Execute performs the list games query
func (a *ListGamesAction) Execute(ctx context.Context, status string) ([]model.Game, error) {
	log := a.GetLogger()
	log.Info("🔍 Querying all games",
		zap.String("status_filter", status))

	// Use old game repository to list games
	// status parameter is ignored for now as List() returns all games
	games, err := a.oldGameRepo.List(ctx, status)
	if err != nil {
		log.Error("Failed to list games", zap.Error(err))
		return nil, err
	}

	log.Info("✅ Games query completed",
		zap.Int("count", len(games)))

	return games, nil
}
