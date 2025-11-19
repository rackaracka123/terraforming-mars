package action

import (
	"context"
	"fmt"

	sessionGame "terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
)

// CheckAllPlayersComplete checks if all players in a game satisfy a given condition
// The checkFunc receives each player and should return true if the player has completed the condition
// Returns true if ALL players satisfy the condition, false otherwise
func CheckAllPlayersComplete(
	ctx context.Context,
	playerRepo player.Repository,
	gameID string,
	checkFunc func(*player.Player) bool,
) (bool, error) {
	players, err := playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return false, fmt.Errorf("failed to list players: %w", err)
	}

	for _, p := range players {
		if !checkFunc(p) {
			return false, nil
		}
	}

	return true, nil
}

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

// GetAllPlayers retrieves all players in a game
// Common helper to avoid repetitive error handling
func GetAllPlayers(
	ctx context.Context,
	playerRepo player.Repository,
	gameID string,
	log *zap.Logger,
) ([]*player.Player, error) {
	players, err := playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players", zap.Error(err))
		return nil, fmt.Errorf("failed to list players: %w", err)
	}
	return players, nil
}
