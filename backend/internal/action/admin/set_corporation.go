package admin

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

// SetCorporationAction handles the admin action to set a player's corporation
type SetCorporationAction struct {
	action.BaseAction
	gameRepo game.Repository
}

// NewSetCorporationAction creates a new set corporation admin action
func NewSetCorporationAction(
	gameRepo game.Repository,
	sessionFactory session.SessionFactory,
	sessionMgrFactory session.SessionManagerFactory,
) *SetCorporationAction {
	return &SetCorporationAction{
		BaseAction: action.NewBaseAction(sessionFactory, sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the set corporation admin action
func (a *SetCorporationAction) Execute(ctx context.Context, gameID, playerID, corporationID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🏢 Admin: Setting player corporation",
		zap.String("corporation_id", corporationID))

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Get session and player
	sess := a.GetSessionFactory().Get(gameID)
	if sess == nil {
		log.Error("Game session not found")
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. Update player corporation
	err = player.Corporation.Set(ctx, corporationID)
	if err != nil {
		log.Error("Failed to update corporation", zap.Error(err))
		return err
	}

	log.Info("✅ Player corporation updated")

	// 4. Broadcast updated game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Admin set corporation completed")
	return nil
}
