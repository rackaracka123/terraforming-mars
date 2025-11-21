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

// GetPlayerAction handles the query for getting a single player
type GetPlayerAction struct {
	action.BaseAction
	oldPlayerRepo repository.PlayerRepository
}

// NewGetPlayerAction creates a new get player query action
func NewGetPlayerAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	oldPlayerRepo repository.PlayerRepository,
	sessionMgr session.SessionManager,
) *GetPlayerAction {
	return &GetPlayerAction{
		BaseAction:    action.NewBaseAction(gameRepo, playerRepo, sessionMgr),
		oldPlayerRepo: oldPlayerRepo,
	}
}

// Execute performs the get player query
func (a *GetPlayerAction) Execute(ctx context.Context, gameID, playerID string) (model.Player, error) {
	log := a.InitLogger(gameID, playerID)
	log.Info("🔍 Querying player")

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.GetGameRepo(), gameID, log)
	if err != nil {
		return model.Player{}, err
	}

	// 2. Get player from old repository
	// TODO: This will need to be migrated when old repository is deleted
	player, err := a.oldPlayerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return model.Player{}, err
	}

	log.Info("✅ Player query completed")

	// Player is already returned as value, not pointer
	return player, nil
}
