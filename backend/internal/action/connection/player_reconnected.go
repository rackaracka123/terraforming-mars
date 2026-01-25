package connection

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
)

// PlayerReconnectedAction handles the business logic for player reconnection
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type PlayerReconnectedAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewPlayerReconnectedAction creates a new player reconnected action
func NewPlayerReconnectedAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *PlayerReconnectedAction {
	return &PlayerReconnectedAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the player reconnected action
func (a *PlayerReconnectedAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "player_reconnected"),
	)
	log.Info("🔗 Player reconnecting")

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

	player.SetConnected(true)

	log.Info("✅ Player reconnected successfully")
	return nil
}
