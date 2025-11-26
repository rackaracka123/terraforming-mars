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

// SetProductionAction handles the admin action to set player production
type SetProductionAction struct {
	action.BaseAction
	gameRepo game.Repository
}

// NewSetProductionAction creates a new set production admin action
func NewSetProductionAction(
	gameRepo game.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *SetProductionAction {
	return &SetProductionAction{
		BaseAction: action.NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the set production admin action
func (a *SetProductionAction) Execute(ctx context.Context, sess *session.Session, playerID string, production types.Production) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("🏭 Admin: Setting player production",
		zap.Int("credits", production.Credits),
		zap.Int("steel", production.Steel),
		zap.Int("titanium", production.Titanium),
		zap.Int("plants", production.Plants),
		zap.Int("energy", production.Energy),
		zap.Int("heat", production.Heat))

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

	// 3. Update player production
	p.Resources().SetProduction(production)

	log.Info("✅ Player production updated")

	// 4. Broadcast updated game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Admin set production completed")
	return nil
}
