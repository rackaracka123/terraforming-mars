package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
)

// SkipAction handles the player skip turn action
// This action advances the turn to the next player
type SkipAction struct {
	gameRepo       game.Repository
	sessionManager session.SessionManager
}

// NewSkipAction creates a new skip action orchestrator
func NewSkipAction(
	gameRepo game.Repository,
	sessionManager session.SessionManager,
) *SkipAction {
	return &SkipAction{
		gameRepo:       gameRepo,
		sessionManager: sessionManager,
	}
}

// Execute performs the skip turn action
// This advances the turn to the next player in turn order
func (a *SkipAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("⏭️ Executing skip turn action")

	// Get current game state
	gameState, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Verify it's the current player's turn
	if gameState.CurrentPlayerID != playerID {
		log.Warn("Player attempted to skip but it's not their turn",
			zap.String("current_player", gameState.CurrentPlayerID))
		return fmt.Errorf("it's not your turn")
	}

	// Find next player in turn order
	playerIDs := gameState.PlayerIDs
	if len(playerIDs) == 0 {
		return fmt.Errorf("no players in game")
	}

	// Find current player index
	currentIndex := -1
	for i, pid := range playerIDs {
		if pid == playerID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return fmt.Errorf("current player not found in player list")
	}

	// Calculate next player index (wrap around)
	nextIndex := (currentIndex + 1) % len(playerIDs)
	nextPlayerID := playerIDs[nextIndex]

	// Set next player as current
	if err := a.gameRepo.SetCurrentPlayer(ctx, gameID, nextPlayerID); err != nil {
		log.Error("Failed to set next player", zap.Error(err))
		return fmt.Errorf("failed to set next player: %w", err)
	}

	log.Info("⏭️ Turn advanced to next player",
		zap.String("previous_player", playerID),
		zap.String("next_player", nextPlayerID))

	// Broadcast updated game state to all players
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after skip turn", zap.Error(err))
		// Don't fail the skip operation, just log the error
	}

	log.Info("✅ Skip turn action completed successfully")
	return nil
}
