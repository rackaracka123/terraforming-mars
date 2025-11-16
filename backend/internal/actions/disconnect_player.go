package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"
)

// DisconnectPlayerAction handles player disconnection logic
// This action encapsulates all business logic for disconnecting a player from a game
type DisconnectPlayerAction struct {
	playerRepo     player.Repository
	sessionManager session.SessionManager
}

// NewDisconnectPlayerAction creates a new disconnect player action
func NewDisconnectPlayerAction(
	playerRepo player.Repository,
	sessionManager session.SessionManager,
) *DisconnectPlayerAction {
	return &DisconnectPlayerAction{
		playerRepo:     playerRepo,
		sessionManager: sessionManager,
	}
}

// Execute handles player disconnection
// Updates the player's connection status and broadcasts the change to all players
func (a *DisconnectPlayerAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🔌 Executing disconnect player action")

	// Update player connection status to false
	err := a.playerRepo.UpdateConnectionStatus(ctx, gameID, playerID, false)
	if err != nil {
		log.Error("Failed to update connection status during disconnection", zap.Error(err))
		return fmt.Errorf("failed to update connection status: %w", err)
	}

	// Broadcast updated game state to other players
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after player disconnection", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	log.Info("✅ Player disconnection processed successfully")
	return nil
}
