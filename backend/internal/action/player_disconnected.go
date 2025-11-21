package action

import (
	"context"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// PlayerDisconnectedAction handles the business logic for player disconnection
type PlayerDisconnectedAction struct {
	BaseAction
}

// NewPlayerDisconnectedAction creates a new player disconnected action
func NewPlayerDisconnectedAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
) *PlayerDisconnectedAction {
	return &PlayerDisconnectedAction{
		BaseAction: NewBaseAction(gameRepo, playerRepo, sessionMgr),
	}
}

// Execute performs the player disconnected action
func (a *PlayerDisconnectedAction) Execute(ctx context.Context, gameID, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🔌 Player disconnecting")

	// 1. Validate game exists
	_, err := ValidateGameExists(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate player exists
	_, err = ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
	if err != nil {
		return err
	}

	// 3. Update player connection status to disconnected
	err = a.playerRepo.UpdateConnectionStatus(ctx, gameID, playerID, false)
	if err != nil {
		log.Error("Failed to update connection status", zap.Error(err))
		return err
	}

	log.Info("✅ Player connection status updated to disconnected")

	// 4. Broadcast updated game state to all players
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Player disconnected successfully")
	return nil
}
