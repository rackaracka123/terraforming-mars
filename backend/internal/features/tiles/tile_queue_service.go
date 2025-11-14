package tiles

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// TileQueueService handles player tile selection queue
//
// Scope: Isolated tile queue management for a player
//   - Add tiles to queue
//   - Pop tiles from queue
//   - Manage pending selection state
type TileQueueService interface {
	GetQueue(ctx context.Context) (*PendingTileSelectionQueue, error)
	GetPendingSelection(ctx context.Context) (*PendingTileSelection, error)

	AddToQueue(ctx context.Context, tileType string) error
	PopFromQueue(ctx context.Context) (string, error)
	ClearQueue(ctx context.Context) error
	GetQueueLength(ctx context.Context) (int, error)

	SetPendingSelection(ctx context.Context, selection *PendingTileSelection) error
	ClearPendingSelection(ctx context.Context) error
}

// TileQueueServiceImpl implements the tile queue service
type TileQueueServiceImpl struct {
	repo TileQueueRepository
}

// NewTileQueueService creates a new tile queue service
func NewTileQueueService(repo TileQueueRepository) TileQueueService {
	return &TileQueueServiceImpl{
		repo: repo,
	}
}

// GetQueue retrieves the current queue
func (s *TileQueueServiceImpl) GetQueue(ctx context.Context) (*PendingTileSelectionQueue, error) {
	return s.repo.GetQueue(ctx)
}

// GetPendingSelection retrieves the current pending selection
func (s *TileQueueServiceImpl) GetPendingSelection(ctx context.Context) (*PendingTileSelection, error) {
	return s.repo.GetPendingSelection(ctx)
}

// AddToQueue adds a tile to the queue
func (s *TileQueueServiceImpl) AddToQueue(ctx context.Context, tileType string) error {
	if err := s.repo.AddToQueue(ctx, tileType); err != nil {
		return fmt.Errorf("failed to add to queue: %w", err)
	}

	logger.Get().Debug("Added tile to queue", zap.String("tile_type", tileType))

	return nil
}

// PopFromQueue removes and returns the first tile from queue
func (s *TileQueueServiceImpl) PopFromQueue(ctx context.Context) (string, error) {
	tileType, err := s.repo.PopFromQueue(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to pop from queue: %w", err)
	}

	logger.Get().Debug("Popped tile from queue", zap.String("tile_type", tileType))

	return tileType, nil
}

// ClearQueue removes all items from queue
func (s *TileQueueServiceImpl) ClearQueue(ctx context.Context) error {
	if err := s.repo.ClearQueue(ctx); err != nil {
		return fmt.Errorf("failed to clear queue: %w", err)
	}
	return nil
}

// GetQueueLength returns the number of items in queue
func (s *TileQueueServiceImpl) GetQueueLength(ctx context.Context) (int, error) {
	length, err := s.repo.GetQueueLength(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}
	return length, nil
}

// SetPendingSelection sets the current pending tile selection
func (s *TileQueueServiceImpl) SetPendingSelection(ctx context.Context, selection *PendingTileSelection) error {
	if err := s.repo.SetPendingSelection(ctx, selection); err != nil {
		return fmt.Errorf("failed to set pending selection: %w", err)
	}

	if selection != nil {
		logger.Get().Debug("Set pending tile selection",
			zap.String("tile_type", selection.TileType),
			zap.Int("available_hexes", len(selection.AvailableHexes)))
	}

	return nil
}

// ClearPendingSelection clears the current pending selection
func (s *TileQueueServiceImpl) ClearPendingSelection(ctx context.Context) error {
	if err := s.repo.ClearPendingSelection(ctx); err != nil {
		return fmt.Errorf("failed to clear pending selection: %w", err)
	}
	return nil
}
