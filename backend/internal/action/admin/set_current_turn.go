package admin

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"

	"go.uber.org/zap"
)

// SetCurrentTurnAction handles the admin action to set the current turn
type SetCurrentTurnAction struct {
	action.BaseAction
	gameRepo game.Repository
}

// NewSetCurrentTurnAction creates a new set current turn admin action
func NewSetCurrentTurnAction(
	gameRepo game.Repository,
	sessionFactory session.SessionFactory,
	sessionMgrFactory session.SessionManagerFactory,
) *SetCurrentTurnAction {
	return &SetCurrentTurnAction{
		BaseAction: action.NewBaseAction(sessionFactory, sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the set current turn admin action
func (a *SetCurrentTurnAction) Execute(ctx context.Context, gameID, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🎲 Admin: Setting current turn")

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate player exists in session
	sess := a.GetSessionFactory().Get(gameID)
	if sess == nil {
		log.Error("Game session not found")
		return fmt.Errorf("game not found: %s", gameID)
	}

	_, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. Update current turn
	err = a.gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
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
