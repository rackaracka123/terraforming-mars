package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"

	"go.uber.org/zap"
)

// PlayerReconnectedAction handles the business logic for player reconnection
type PlayerReconnectedAction struct {
	BaseAction
}

// NewPlayerReconnectedAction creates a new player reconnected action
func NewPlayerReconnectedAction(
	sessionFactory session.SessionFactory,
	sessionMgrFactory session.SessionManagerFactory,
) *PlayerReconnectedAction {
	return &PlayerReconnectedAction{
		BaseAction: NewBaseAction(sessionFactory, sessionMgrFactory),
	}
}

// Execute performs the player reconnected action
func (a *PlayerReconnectedAction) Execute(ctx context.Context, gameID, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🔗 Player reconnecting")

	// 1. Get session
	sess := a.sessionFactory.Get(gameID)
	if sess == nil {
		log.Error("Game session not found")
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. Get player from session
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. Update player connection status to connected
	err := player.Turn.UpdateConnectionStatus(ctx, true)
	if err != nil {
		log.Error("Failed to update connection status", zap.Error(err))
		return err
	}

	log.Info("✅ Player connection status updated to connected")

	// 4. Send complete game state to reconnected player
	err = a.sessionMgrFactory.GetOrCreate(gameID).Send(playerID)
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
