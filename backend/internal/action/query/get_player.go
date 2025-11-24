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

// GetPlayerAction handles the query for getting a single player
type GetPlayerAction struct {
	action.BaseAction
}

// NewGetPlayerAction creates a new get player query action
func NewGetPlayerAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *GetPlayerAction {
	return &GetPlayerAction{
		BaseAction: action.NewBaseAction(gameRepo, playerRepo, sessionMgrFactory),
	}
}

// Execute performs the get player query
func (a *GetPlayerAction) Execute(ctx context.Context, gameID, playerID string) (types.Player, error) {
	log := a.InitLogger(gameID, playerID)
	log.Info("🔍 Querying player")

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.GetGameRepo(), gameID, log)
	if err != nil {
		return types.Player{}, err
	}

	// 2. Get player from repository
	player, err := a.GetPlayerRepo().GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return types.Player{}, err
	}

	log.Info("✅ Player query completed")

	// Convert pointer to value for return
	return *player, nil
}
