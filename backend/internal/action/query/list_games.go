package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// ListGamesAction handles the query for listing all games
type ListGamesAction struct {
	action.BaseAction
}

// NewListGamesAction creates a new list games query action
func NewListGamesAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *ListGamesAction {
	return &ListGamesAction{
		BaseAction: action.NewBaseAction(gameRepo, playerRepo, sessionMgrFactory),
	}
}

// Execute performs the list games query
func (a *ListGamesAction) Execute(ctx context.Context, status string) ([]types.Game, error) {
	log := a.GetLogger()
	log.Info("🔍 Querying all games",
		zap.String("status_filter", status))

	// List games from repository
	gamePointers, err := a.GetGameRepo().List(ctx, status)
	if err != nil {
		log.Error("Failed to list games", zap.Error(err))
		return nil, err
	}

	// Convert game pointers to values
	games := make([]types.Game, len(gamePointers))
	for i, gamePtr := range gamePointers {
		games[i] = types.Game(*gamePtr)
	}

	log.Info("✅ Games query completed",
		zap.Int("count", len(games)))

	return games, nil
}
