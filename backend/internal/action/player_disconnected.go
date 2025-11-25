package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"

	"go.uber.org/zap"
)

// PlayerDisconnectedAction handles the business logic for player disconnection
type PlayerDisconnectedAction struct {
	BaseAction
}

// NewPlayerDisconnectedAction creates a new player disconnected action
func NewPlayerDisconnectedAction(
	sessionFactory session.SessionFactory,
	sessionMgrFactory session.SessionManagerFactory,
) *PlayerDisconnectedAction {
	return &PlayerDisconnectedAction{
		BaseAction: NewBaseAction(sessionFactory, sessionMgrFactory),
	}
}

// Execute performs the player disconnected action
func (a *PlayerDisconnectedAction) Execute(ctx context.Context, gameID, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🔌 Player disconnecting")

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

	// 3. Update player connection status to disconnected
	err := player.Turn.UpdateConnectionStatus(ctx, false)
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
