package admin

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// SetResourcesAction handles the admin action to set player resources
type SetResourcesAction struct {
	action.BaseAction
}

// NewSetResourcesAction creates a new set resources admin action
func NewSetResourcesAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
) *SetResourcesAction {
	return &SetResourcesAction{
		BaseAction: action.NewBaseAction(gameRepo, playerRepo, sessionMgr),
	}
}

// Execute performs the set resources admin action
func (a *SetResourcesAction) Execute(ctx context.Context, gameID, playerID string, resources model.Resources) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("💰 Admin: Setting player resources",
		zap.Int("credits", resources.Credits),
		zap.Int("steel", resources.Steel),
		zap.Int("titanium", resources.Titanium),
		zap.Int("plants", resources.Plants),
		zap.Int("energy", resources.Energy),
		zap.Int("heat", resources.Heat))

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.GetGameRepo(), gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate player exists
	_, err = action.ValidatePlayer(ctx, a.GetPlayerRepo(), gameID, playerID, log)
	if err != nil {
		return err
	}

	// 3. Update player resources
	err = a.GetPlayerRepo().UpdateResources(ctx, gameID, playerID, resources)
	if err != nil {
		log.Error("Failed to update resources", zap.Error(err))
		return err
	}

	log.Info("✅ Player resources updated")

	// 4. Broadcast updated game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Admin set resources completed")
	return nil
}
