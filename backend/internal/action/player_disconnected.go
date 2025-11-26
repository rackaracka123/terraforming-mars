package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
)

// PlayerDisconnectedAction handles the business logic for player disconnection
type PlayerDisconnectedAction struct {
	BaseAction
}

// NewPlayerDisconnectedAction creates a new player disconnected action
func NewPlayerDisconnectedAction(
	sessionMgrFactory session.SessionManagerFactory,
) *PlayerDisconnectedAction {
	return &PlayerDisconnectedAction{
		BaseAction: NewBaseAction(sessionMgrFactory),
	}
}

// Execute performs the player disconnected action
func (a *PlayerDisconnectedAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("🔌 Player disconnecting")

	// 1. Get session	// 2. Get player from session
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. Update player connection status to disconnected
	player.Turn().SetConnectionStatus(false)

	log.Info("✅ Player connection status updated to disconnected")

	// 4. Broadcast updated game state to all players
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Player disconnected successfully")
	return nil
}
