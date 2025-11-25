package action

import (
	"context"
	"fmt"

	sessionGame "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

// CheckAllPlayersComplete is deprecated - use session.GetAllPlayers() directly
// This helper function is no longer needed with the new session-based architecture

// TransitionGamePhase updates the game to a new phase with proper logging
// This is a common pattern used when advancing through game phases
func TransitionGamePhase(
	ctx context.Context,
	gameRepo sessionGame.Repository,
	gameID string,
	newPhase sessionGame.GamePhase,
	log *zap.Logger,
) error {
	err := gameRepo.UpdatePhase(ctx, gameID, newPhase)
	if err != nil {
		log.Error("Failed to update game phase",
			zap.Error(err),
			zap.String("new_phase", string(newPhase)))
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	log.Info("✅ Game phase updated", zap.String("new_phase", string(newPhase)))
	return nil
}

// TransitionGameStatus updates the game to a new status with proper logging
// This is used when changing game lifecycle status (lobby -> active, active -> completed)
func TransitionGameStatus(
	ctx context.Context,
	gameRepo sessionGame.Repository,
	gameID string,
	newStatus sessionGame.GameStatus,
	log *zap.Logger,
) error {
	err := gameRepo.UpdateStatus(ctx, gameID, newStatus)
	if err != nil {
		log.Error("Failed to update game status",
			zap.Error(err),
			zap.String("new_status", string(newStatus)))
		return fmt.Errorf("failed to update game status: %w", err)
	}

	log.Info("✅ Game status updated", zap.String("new_status", string(newStatus)))
	return nil
}

// SetCurrentTurn sets the current turn to a specific player
// Handles nil player ID for clearing the current turn
func SetCurrentTurn(
	ctx context.Context,
	gameRepo sessionGame.Repository,
	gameID string,
	playerID *string,
	log *zap.Logger,
) error {
	err := gameRepo.SetCurrentTurn(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to set current turn", zap.Error(err))
		return fmt.Errorf("failed to set current turn: %w", err)
	}

	if playerID != nil {
		log.Debug("Current turn set", zap.String("player_id", *playerID))
	} else {
		log.Debug("Current turn cleared")
	}

	return nil
}

// GetAllPlayers is deprecated - use session.GetAllPlayers() directly
// This helper function is no longer needed with the new session-based architecture
