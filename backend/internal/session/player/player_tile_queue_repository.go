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

// PlayerTileQueueRepository handles player tile placement queue operations
type PlayerTileQueueRepository struct {
	storage  *PlayerStorage
	eventBus *events.EventBusImpl
}

// NewPlayerTileQueueRepository creates a new player tile queue repository
func NewPlayerTileQueueRepository(storage *PlayerStorage, eventBus *events.EventBusImpl) *PlayerTileQueueRepository {
	return &PlayerTileQueueRepository{
		storage:  storage,
		eventBus: eventBus,
	}
}

// CreateTileQueue creates a tile placement queue for the player
func (r *PlayerTileQueueRepository) CreateTileQueue(ctx context.Context, gameID string, playerID string, cardID string, tileTypes []string) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	// Create tile queue
	if len(tileTypes) > 0 {
		p.PendingTileSelectionQueue = &types.PendingTileSelectionQueue{
			Items:  tileTypes,
			Source: cardID,
		}

		err = r.storage.Set(gameID, playerID, p)
		if err != nil {
			return err
		}

		// Publish TileQueueCreatedEvent
		logTileQueue.Debug("📡 Publishing TileQueueCreatedEvent",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Int("queue_size", len(tileTypes)),
			zap.String("source", cardID))

		events.Publish(r.eventBus, TileQueueCreatedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			QueueSize: len(tileTypes),
			Source:    cardID,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// GetPendingTileSelectionQueue retrieves the pending tile selection queue for a player
func (r *PlayerTileQueueRepository) GetPendingTileSelectionQueue(ctx context.Context, gameID string, playerID string) (*types.PendingTileSelectionQueue, error) {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return nil, err
	}

	if p.PendingTileSelectionQueue == nil {
		return nil, nil
	}

	// Return a copy to prevent external mutation
	itemsCopy := make([]string, len(p.PendingTileSelectionQueue.Items))
	copy(itemsCopy, p.PendingTileSelectionQueue.Items)

	return &types.PendingTileSelectionQueue{
		Items:  itemsCopy,
		Source: p.PendingTileSelectionQueue.Source,
	}, nil
}

// ProcessNextTileInQueue pops the next tile type from the queue and returns it
func (r *PlayerTileQueueRepository) ProcessNextTileInQueue(ctx context.Context, gameID string, playerID string) (string, error) {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return "", err
	}

	// If no queue exists or queue is empty, nothing to process
	if p.PendingTileSelectionQueue == nil || len(p.PendingTileSelectionQueue.Items) == 0 {
		logTileQueue.Debug("No tile placements in queue",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		return "", nil
	}

	// Pop the first item from the queue
	nextTileType := p.PendingTileSelectionQueue.Items[0]
	remainingItems := p.PendingTileSelectionQueue.Items[1:]

	logTileQueue.Info("🎯 Popping next tile from queue",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("tile_type", nextTileType),
		zap.String("source", p.PendingTileSelectionQueue.Source),
		zap.Int("remaining_in_queue", len(remainingItems)))

	source := p.PendingTileSelectionQueue.Source

	// Update queue with remaining items or clear if empty
	if len(remainingItems) > 0 {
		p.PendingTileSelectionQueue = &types.PendingTileSelectionQueue{
			Items:  remainingItems,
			Source: p.PendingTileSelectionQueue.Source,
		}
	} else {
		// This is the last item, clear the queue
		p.PendingTileSelectionQueue = nil
	}

	err = r.storage.Set(gameID, playerID, p)
	if err != nil {
		return "", err
	}

	logTileQueue.Debug("✅ Tile popped from queue",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("tile_type", nextTileType))

	// If there are more tiles in queue after popping, publish event to trigger processing
	// This ensures the next tile is automatically processed after the current one is placed
	if len(remainingItems) > 0 {
		logTileQueue.Debug("📡 Publishing TileQueueCreatedEvent for remaining tiles",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Int("remaining_count", len(remainingItems)),
			zap.String("source", source))

		events.Publish(r.eventBus, TileQueueCreatedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			QueueSize: len(remainingItems),
			Source:    source,
			Timestamp: time.Now(),
		})
	}

	return nextTileType, nil
}

// UpdatePendingTileSelection updates the pending tile selection for a player
func (r *PlayerTileQueueRepository) UpdatePendingTileSelection(ctx context.Context, gameID string, playerID string, selection *types.PendingTileSelection) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.PendingTileSelection = selection

	return r.storage.Set(gameID, playerID, p)
}

// ClearPendingTileSelection clears the pending tile selection for a player
func (r *PlayerTileQueueRepository) ClearPendingTileSelection(ctx context.Context, gameID string, playerID string) error {
	return r.UpdatePendingTileSelection(ctx, gameID, playerID, nil)
}

// UpdatePendingTileSelectionQueue updates the pending tile selection queue
func (r *PlayerTileQueueRepository) UpdatePendingTileSelectionQueue(ctx context.Context, gameID string, playerID string, queue *types.PendingTileSelectionQueue) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.PendingTileSelectionQueue = queue

	return r.storage.Set(gameID, playerID, p)
}

// ClearPendingTileSelectionQueue clears the pending tile selection queue
func (r *PlayerTileQueueRepository) ClearPendingTileSelectionQueue(ctx context.Context, gameID string, playerID string) error {
	return r.UpdatePendingTileSelectionQueue(ctx, gameID, playerID, nil)
}
