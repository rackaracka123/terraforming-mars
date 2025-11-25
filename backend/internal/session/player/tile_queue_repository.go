package player

import (
	"context"
	"time"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/types"
)

var logTileQueue = logger.Get()

// TileQueueRepository handles tile placement queue operations for a specific player
// Auto-saves changes after every operation
type TileQueueRepository struct {
	player   *Player // Reference to parent player
	eventBus *events.EventBusImpl
}

// NewTileQueueRepository creates a new tile queue repository for a specific player
func NewTileQueueRepository(player *Player, eventBus *events.EventBusImpl) *TileQueueRepository {
	return &TileQueueRepository{
		player:   player,
		eventBus: eventBus,
	}
}

// CreateQueue creates a tile placement queue for the player
// Auto-saves changes to the player
func (r *TileQueueRepository) CreateQueue(ctx context.Context, cardID string, tileTypes []string) error {
	// Create tile queue
	if len(tileTypes) > 0 {
		r.player.PendingTileSelectionQueue = &types.PendingTileSelectionQueue{
			Items:  tileTypes,
			Source: cardID,
		}

		// Publish TileQueueCreatedEvent
		logTileQueue.Debug("📡 Publishing TileQueueCreatedEvent",
			zap.String("game_id", r.player.GameID),
			zap.String("player_id", r.player.ID),
			zap.Int("queue_size", len(tileTypes)),
			zap.String("source", cardID))

		events.Publish(r.eventBus, TileQueueCreatedEvent{
			GameID:    r.player.GameID,
			PlayerID:  r.player.ID,
			QueueSize: len(tileTypes),
			Source:    cardID,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// GetQueue retrieves the pending tile selection queue
func (r *TileQueueRepository) GetQueue(ctx context.Context) (*types.PendingTileSelectionQueue, error) {
	if r.player.PendingTileSelectionQueue == nil {
		return nil, nil
	}

	// Return a copy to prevent external mutation
	itemsCopy := make([]string, len(r.player.PendingTileSelectionQueue.Items))
	copy(itemsCopy, r.player.PendingTileSelectionQueue.Items)

	return &types.PendingTileSelectionQueue{
		Items:  itemsCopy,
		Source: r.player.PendingTileSelectionQueue.Source,
	}, nil
}

// ProcessNext pops the next tile type from the queue and returns it
// Auto-saves changes to the player
func (r *TileQueueRepository) ProcessNext(ctx context.Context) (string, error) {
	// If no queue exists or queue is empty, nothing to process
	if r.player.PendingTileSelectionQueue == nil || len(r.player.PendingTileSelectionQueue.Items) == 0 {
		logTileQueue.Debug("No tile placements in queue",
			zap.String("game_id", r.player.GameID),
			zap.String("player_id", r.player.ID))
		return "", nil
	}

	// Pop the first item from the queue
	nextTileType := r.player.PendingTileSelectionQueue.Items[0]
	remainingItems := r.player.PendingTileSelectionQueue.Items[1:]

	logTileQueue.Info("🎯 Popping next tile from queue",
		zap.String("game_id", r.player.GameID),
		zap.String("player_id", r.player.ID),
		zap.String("tile_type", nextTileType),
		zap.String("source", r.player.PendingTileSelectionQueue.Source),
		zap.Int("remaining_in_queue", len(remainingItems)))

	source := r.player.PendingTileSelectionQueue.Source

	// Update queue with remaining items or clear if empty
	if len(remainingItems) > 0 {
		r.player.PendingTileSelectionQueue = &types.PendingTileSelectionQueue{
			Items:  remainingItems,
			Source: r.player.PendingTileSelectionQueue.Source,
		}
	} else {
		// This is the last item, clear the queue
		r.player.PendingTileSelectionQueue = nil
	}

	logTileQueue.Debug("✅ Tile popped from queue",
		zap.String("game_id", r.player.GameID),
		zap.String("player_id", r.player.ID),
		zap.String("tile_type", nextTileType))

	// If there are more tiles in queue after popping, publish event to trigger processing
	// This ensures the next tile is automatically processed after the current one is placed
	if len(remainingItems) > 0 {
		logTileQueue.Debug("📡 Publishing TileQueueCreatedEvent for remaining tiles",
			zap.String("game_id", r.player.GameID),
			zap.String("player_id", r.player.ID),
			zap.Int("remaining_count", len(remainingItems)),
			zap.String("source", source))

		events.Publish(r.eventBus, TileQueueCreatedEvent{
			GameID:    r.player.GameID,
			PlayerID:  r.player.ID,
			QueueSize: len(remainingItems),
			Source:    source,
			Timestamp: time.Now(),
		})
	}

	return nextTileType, nil
}

// UpdatePendingTileSelection updates the pending tile selection
// Auto-saves changes to the player
func (r *TileQueueRepository) UpdatePendingTileSelection(ctx context.Context, selection *types.PendingTileSelection) error {
	r.player.PendingTileSelection = selection
	return nil
}

// ClearPendingTileSelection clears the pending tile selection
// Auto-saves changes to the player
func (r *TileQueueRepository) ClearPendingTileSelection(ctx context.Context) error {
	r.player.PendingTileSelection = nil
	return nil
}

// UpdateQueue updates the pending tile selection queue
// Auto-saves changes to the player
func (r *TileQueueRepository) UpdateQueue(ctx context.Context, queue *types.PendingTileSelectionQueue) error {
	r.player.PendingTileSelectionQueue = queue
	return nil
}

// ClearQueue clears the pending tile selection queue
// Auto-saves changes to the player
func (r *TileQueueRepository) ClearQueue(ctx context.Context) error {
	r.player.PendingTileSelectionQueue = nil
	return nil
}
