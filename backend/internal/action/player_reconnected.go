package action

import (
	"context"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// PlayerReconnectedAction handles the business logic for player reconnection
type PlayerReconnectedAction struct {
	BaseAction
}

// NewPlayerReconnectedAction creates a new player reconnected action
func NewPlayerReconnectedAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
) *PlayerReconnectedAction {
	return &PlayerReconnectedAction{
		BaseAction: NewBaseAction(gameRepo, playerRepo, sessionMgr),
	}
}

// Execute performs the player reconnected action
func (a *PlayerReconnectedAction) Execute(ctx context.Context, gameID, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🔗 Player reconnecting")

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

	// 3. Update player connection status to connected
	err = a.playerRepo.UpdateConnectionStatus(ctx, gameID, playerID, true)
	if err != nil {
		log.Error("Failed to update connection status", zap.Error(err))
		return err
	}

	log.Info("✅ Player connection status updated to connected")

	// 4. Send complete game state to reconnected player
	err = a.sessionMgr.Send(gameID, playerID)
	if err != nil {
		log.Error("Failed to send game state to reconnected player", zap.Error(err))
		return err
	}

	log.Info("📤 Game state sent to reconnected player")

	// 5. Broadcast to all players that this player has reconnected
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Player reconnected successfully")
	return nil
}
