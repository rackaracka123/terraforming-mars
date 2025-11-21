package admin

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// SetProductionAction handles the admin action to set player production
type SetProductionAction struct {
	action.BaseAction
}

// NewSetProductionAction creates a new set production admin action
func NewSetProductionAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
) *SetProductionAction {
	return &SetProductionAction{
		BaseAction: action.NewBaseAction(gameRepo, playerRepo, sessionMgr),
	}
}

// Execute performs the set production admin action
func (a *SetProductionAction) Execute(ctx context.Context, gameID, playerID string, production types.Production) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🏭 Admin: Setting player production",
		zap.Int("credits", production.Credits),
		zap.Int("steel", production.Steel),
		zap.Int("titanium", production.Titanium),
		zap.Int("plants", production.Plants),
		zap.Int("energy", production.Energy),
		zap.Int("heat", production.Heat))

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

	// 3. Update player production
	err = a.GetPlayerRepo().UpdateProduction(ctx, gameID, playerID, production)
	if err != nil {
		log.Error("Failed to update production", zap.Error(err))
		return err
	}

	log.Info("✅ Player production updated")

	// 4. Broadcast updated game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Admin set production completed")
	return nil
}
