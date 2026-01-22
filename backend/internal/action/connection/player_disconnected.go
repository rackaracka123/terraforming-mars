package connection

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
)

// PlayerDisconnectedAction handles the business logic for player disconnection
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type PlayerDisconnectedAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewPlayerDisconnectedAction creates a new player disconnected action
func NewPlayerDisconnectedAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *PlayerDisconnectedAction {
	return &PlayerDisconnectedAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the player disconnected action
func (a *PlayerDisconnectedAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "player_disconnected"),
	)
	log.Info("🔌 Player disconnecting")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	player.SetConnected(false)

	log.Info("✅ Player disconnected successfully")
	return nil
}
