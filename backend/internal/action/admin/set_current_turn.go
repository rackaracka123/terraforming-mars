package admin

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// SetCurrentTurnAction handles the admin action to set the current turn
type SetCurrentTurnAction struct {
	action.BaseAction
}

// NewSetCurrentTurnAction creates a new set current turn admin action
func NewSetCurrentTurnAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *SetCurrentTurnAction {
	return &SetCurrentTurnAction{
		BaseAction: action.NewBaseAction(gameRepo, playerRepo, sessionMgrFactory),
	}
}

// Execute performs the set current turn admin action
func (a *SetCurrentTurnAction) Execute(ctx context.Context, gameID, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🎲 Admin: Setting current turn")

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

	// 3. Update current turn
	err = a.GetGameRepo().SetCurrentTurn(ctx, gameID, &playerID)
	if err != nil {
		log.Error("Failed to update current turn", zap.Error(err))
		return err
	}

	log.Info("✅ Current turn updated")

	// 4. Broadcast updated game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Admin set current turn completed")
	return nil
}
