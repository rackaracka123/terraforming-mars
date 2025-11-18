package tile

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game/player"
)

var log = logger.Get()

// Processor handles tile queue processing for the NEW session-based architecture
type Processor struct {
	playerRepo player.Repository
}

// NewProcessor creates a new tile processor
func NewProcessor(playerRepo player.Repository) *Processor {
	return &Processor{
		playerRepo: playerRepo,
	}
}

// ProcessTileQueue processes the tile queue for a player
// This is a simplified version that just processes the queue without validation
// Full validation and hex calculation will be added when needed
func (p *Processor) ProcessTileQueue(ctx context.Context, gameID string, playerID string) error {
	log.Debug("🎯 Processing tile queue (NEW session-based processor)",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID))

	// Get the tile queue
	queue, err := p.playerRepo.GetPendingTileSelectionQueue(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get tile queue: %w", err)
	}

	// If no queue exists, nothing to do
	if queue == nil || len(queue.Items) == 0 {
		log.Debug("No tile queue to process",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		return nil
	}

	log.Info("🎯 Tile queue exists but processing is not yet implemented",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Int("queue_length", len(queue.Items)),
		zap.String("source_card", queue.Source))

	// TODO: Implement full tile queue processing
	// For now, just log that we have a queue
	// The queue will be processed when the player selects where to place tiles

	return nil
}
