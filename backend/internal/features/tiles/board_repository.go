package tiles

import (
	"context"
"terraforming-mars-backend/internal/model"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	// Removed - events now in events package
)

// BoardRepository manages the Mars board tiles
type BoardRepository interface {
	// Get board state
	GetBoard(ctx context.Context) (model.Board, error)

	// Granular tile operations
	PlaceTile(ctx context.Context, coordinate model.HexPosition, occupant model.TileOccupant, ownerID *string) error
	GetTile(ctx context.Context, coordinate model.HexPosition) (*model.Tile, error)
	IsTileOccupied(ctx context.Context, coordinate model.HexPosition) (bool, error)
}

// BoardRepositoryImpl implements independent in-memory storage for the board
type BoardRepositoryImpl struct {
	mu       sync.RWMutex
	gameID   string
	tiles    map[string]*model.Tile // map[coordinateString]*model.Tile for O(1) lookup
	eventBus *events.EventBusImpl
}

// NewBoardRepository creates a new independent board repository with initial board state
func NewBoardRepository(gameID string, initialBoard model.Board, eventBus *events.EventBusImpl) BoardRepository {
	tiles := make(map[string]*model.Tile)
	for i := range initialBoard.Tiles {
		tile := initialBoard.Tiles[i]
		coordKey := tile.Coordinates.String()
		tileCopy := tile
		tiles[coordKey] = &tileCopy
	}

	return &BoardRepositoryImpl{
		gameID:   gameID,
		tiles:    tiles,
		eventBus: eventBus,
	}
}

// GetBoard retrieves the complete board state
func (r *BoardRepositoryImpl) GetBoard(ctx context.Context) (model.Board, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tiles := make([]model.Tile, 0, len(r.tiles))
	for _, tile := range r.tiles {
		tiles = append(tiles, *tile)
	}

	return model.Board{Tiles: tiles}, nil
}

// PlaceTile places a tile occupant at the given coordinates
func (r *BoardRepositoryImpl) PlaceTile(ctx context.Context, coordinate model.HexPosition, occupant model.TileOccupant, ownerID *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	coordKey := coordinate.String()
	tile, exists := r.tiles[coordKey]
	if !exists {
		return fmt.Errorf("tile does not exist at %s", coordKey)
	}

	if tile.OccupiedBy != nil {
		return fmt.Errorf("tile at %s is already occupied", coordKey)
	}

	tile.OccupiedBy = &occupant
	tile.OwnerID = ownerID

	// Publish event if eventBus is available
	if r.eventBus != nil && ownerID != nil {
		events.Publish(r.eventBus, events.TilePlacedEvent{
			GameID:    r.gameID,
			PlayerID:  *ownerID,
			TileType:  string(occupant.Type),
			Q:         coordinate.Q,
			R:         coordinate.R,
			S:         coordinate.S,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// GetTile retrieves a specific tile by coordinates
func (r *BoardRepositoryImpl) GetTile(ctx context.Context, coordinate model.HexPosition) (*model.Tile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	coordKey := coordinate.String()
	tile, exists := r.tiles[coordKey]
	if !exists {
		return nil, fmt.Errorf("tile does not exist at %s", coordKey)
	}

	// Return a copy
	tileCopy := *tile
	return &tileCopy, nil
}

// IsTileOccupied checks if a tile is occupied
func (r *BoardRepositoryImpl) IsTileOccupied(ctx context.Context, coordinate model.HexPosition) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	coordKey := coordinate.String()
	tile, exists := r.tiles[coordKey]
	if !exists {
		return false, fmt.Errorf("tile does not exist at %s", coordKey)
	}

	return tile.OccupiedBy != nil, nil
}
