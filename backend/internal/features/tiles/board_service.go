package tiles

import (
	"context"
"terraforming-mars-backend/internal/model"
	"fmt"

	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// BoardService handles board tile operations
//
// Scope: Isolated board management for a game
//   - model.Tile placement
//   - model.Tile occupancy checking
//   - model.Board state retrieval
type BoardService interface {
	GetBoard(ctx context.Context) (model.Board, error)
	PlaceTile(ctx context.Context, coordinate model.HexPosition, occupant model.TileOccupant, ownerID *string) error
	GetTile(ctx context.Context, coordinate model.HexPosition) (*model.Tile, error)
	IsTileOccupied(ctx context.Context, coordinate model.HexPosition) (bool, error)
}

// BoardServiceImpl implements the board service
type BoardServiceImpl struct {
	repo BoardRepository
}

// NewBoardService creates a new board service
func NewBoardService(repo BoardRepository) BoardService {
	return &BoardServiceImpl{
		repo: repo,
	}
}

// GetBoard retrieves the complete board
func (s *BoardServiceImpl) GetBoard(ctx context.Context) (model.Board, error) {
	return s.repo.GetBoard(ctx)
}

// PlaceTile places a tile on the board
func (s *BoardServiceImpl) PlaceTile(ctx context.Context, coordinate model.HexPosition, occupant model.TileOccupant, ownerID *string) error {
	if err := s.repo.PlaceTile(ctx, coordinate, occupant, ownerID); err != nil {
		return fmt.Errorf("failed to place tile: %w", err)
	}

	logger.Get().Info("🏗️ model.Tile placed",
		zap.String("coordinate", coordinate.String()),
		zap.String("type", string(occupant.Type)),
		zap.Stringp("owner_id", ownerID))

	// TODO Phase 6: Publish TilePlacedEvent

	return nil
}

// GetTile retrieves a specific tile
func (s *BoardServiceImpl) GetTile(ctx context.Context, coordinate model.HexPosition) (*model.Tile, error) {
	tile, err := s.repo.GetTile(ctx, coordinate)
	if err != nil {
		return nil, fmt.Errorf("failed to get tile: %w", err)
	}
	return tile, nil
}

// IsTileOccupied checks if a tile is occupied
func (s *BoardServiceImpl) IsTileOccupied(ctx context.Context, coordinate model.HexPosition) (bool, error) {
	occupied, err := s.repo.IsTileOccupied(ctx, coordinate)
	if err != nil {
		return false, fmt.Errorf("failed to check tile occupancy: %w", err)
	}
	return occupied, nil
}
