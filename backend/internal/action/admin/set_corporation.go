package admin

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// SetCorporationAction handles the admin action to set a player's corporation
type SetCorporationAction struct {
	action.BaseAction
}

// NewSetCorporationAction creates a new set corporation admin action
func NewSetCorporationAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
) *SetCorporationAction {
	return &SetCorporationAction{
		BaseAction: action.NewBaseAction(gameRepo, playerRepo, sessionMgr),
	}
}

// Execute performs the set corporation admin action
func (a *SetCorporationAction) Execute(ctx context.Context, gameID, playerID, corporationID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🏢 Admin: Setting player corporation",
		zap.String("corporation_id", corporationID))

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

	// 3. Update player corporation
	err = a.GetPlayerRepo().SetCorporation(ctx, gameID, playerID, corporationID)
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
