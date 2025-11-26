package admin

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// SetResourcesAction handles the admin action to set player resources
type SetResourcesAction struct {
	action.BaseAction
	gameRepo game.Repository
}

// NewSetResourcesAction creates a new set resources admin action
func NewSetResourcesAction(
	gameRepo game.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *SetResourcesAction {
	return &SetResourcesAction{
		BaseAction: action.NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the set resources admin action
func (a *SetResourcesAction) Execute(ctx context.Context, sess *session.Session, playerID string, resources types.Resources) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("💰 Admin: Setting player resources",
		zap.Int("credits", resources.Credits),
		zap.Int("steel", resources.Steel),
		zap.Int("titanium", resources.Titanium),
		zap.Int("plants", resources.Plants),
		zap.Int("energy", resources.Energy),
		zap.Int("heat", resources.Heat))

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Get session and player
	p, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. Update player resources
	p.Resources().Set(resources)

	log.Info("✅ Player resources updated")

	// 4. Broadcast updated game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Admin set resources completed")
	return nil
}
