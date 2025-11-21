package admin

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// SetPhaseAction handles the admin action to set the game phase
type SetPhaseAction struct {
	action.BaseAction
}

// NewSetPhaseAction creates a new set phase admin action
func NewSetPhaseAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
) *SetPhaseAction {
	return &SetPhaseAction{
		BaseAction: action.NewBaseAction(gameRepo, playerRepo, sessionMgr),
	}
}

// Execute performs the set phase admin action
func (a *SetPhaseAction) Execute(ctx context.Context, gameID string, phase game.GamePhase) error {
	log := a.InitLogger(gameID, "")
	log.Info("🎬 Admin: Setting game phase",
		zap.String("phase", string(phase)))

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.GetGameRepo(), gameID, log)
	if err != nil {
		return err
	}

	// 2. Update game phase
	err = a.GetGameRepo().UpdatePhase(ctx, gameID, phase)
	if err != nil {
		log.Error("Failed to update phase", zap.Error(err))
		return err
	}

	log.Info("✅ Game phase updated")

	// 3. Broadcast updated game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Admin set phase completed")
	return nil
}
