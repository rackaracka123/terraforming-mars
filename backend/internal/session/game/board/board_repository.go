package board

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Repository defines the interface for board data access
// This repository is scoped to a specific game instance
type Repository interface {
	// Get returns a deep copy of the board (immutable getter)
	Get(ctx context.Context) (*Board, error)

	// GenerateBoard creates and initializes the default 42-tile board
	GenerateBoard(ctx context.Context) error

	// UpdateTileOccupancy updates what occupies a tile and who owns it
	UpdateTileOccupancy(ctx context.Context, coord HexPosition, occupant *TileOccupant, ownerID *string) error

	// GetTile returns a specific tile by coordinates (immutable getter)
	GetTile(ctx context.Context, coord HexPosition) (*Tile, error)
}

// RepositoryImpl implements the Repository interface with in-memory storage
// Scoped to a specific game instance
type RepositoryImpl struct {
	gameID    string // Bound to specific game
	mu        sync.RWMutex
	boards    map[string]*Board // gameID -> Board (shared storage across instances)
	processor *BoardProcessor
	eventBus  *events.EventBusImpl
	logger    *zap.Logger
}

// NewRepository creates a new board repository bound to a specific game
func NewRepository(gameID string, boards map[string]*Board, eventBus *events.EventBusImpl) Repository {
	return &RepositoryImpl{
		gameID:    gameID,
		boards:    boards,
		processor: NewBoardProcessor(),
		eventBus:  eventBus,
		logger:    logger.Get(),
	}
}

// Get returns a deep copy of the board (immutable getter)
func (r *RepositoryImpl) Get(ctx context.Context) (*Board, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	board, exists := r.boards[r.gameID]
	if !exists {
		return nil, fmt.Errorf("board not found for game: %s", r.gameID)
	}

	// Return deep copy to maintain immutability
	return board.DeepCopy(), nil
}

// GenerateBoard creates and initializes the default 42-tile board
func (r *RepositoryImpl) GenerateBoard(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	log := r.logger.With(zap.String("game_id", r.gameID))

	// Check if board already exists
	if _, exists := r.boards[r.gameID]; exists {
		log.Warn("⚠️  Board already exists for game, skipping generation")
		return fmt.Errorf("board already exists for game: %s", r.gameID)
	}

	// Generate tiles using processor
	tiles := r.processor.GenerateTiles()

	// Create board
	board := &Board{
		Tiles: tiles,
	}

	// Store board
	r.boards[r.gameID] = board

	log.Info("🗺️  Generated board with tiles", zap.Int("tile_count", len(tiles)))

	return nil
}

// UpdateTileOccupancy updates what occupies a tile and who owns it
func (r *RepositoryImpl) UpdateTileOccupancy(ctx context.Context, coord HexPosition, occupant *TileOccupant, ownerID *string) error {
	r.mu.Lock()

	log := r.logger.With(
		zap.String("game_id", r.gameID),
		zap.String("coordinate", coord.String()),
	)

	board, exists := r.boards[r.gameID]
	if !exists {
		r.mu.Unlock()
		return fmt.Errorf("board not found for game: %s", r.gameID)
	}

	// Find the tile
	tileIndex := -1
	for i, tile := range board.Tiles {
		if tile.Coordinates.Q == coord.Q && tile.Coordinates.R == coord.R && tile.Coordinates.S == coord.S {
			tileIndex = i
			break
		}
	}

	if tileIndex == -1 {
		r.mu.Unlock()
		return fmt.Errorf("tile not found at coordinates: %s", coord.String())
	}

	// Update tile occupancy
	board.Tiles[tileIndex].OccupiedBy = occupant
	board.Tiles[tileIndex].OwnerID = ownerID

	// Prepare event data before releasing lock
	var eventToPublish *events.TilePlacedEvent
	if occupant != nil && ownerID != nil {
		eventToPublish = &events.TilePlacedEvent{
			GameID:    r.gameID,
			PlayerID:  *ownerID,
			TileType:  string(occupant.Type),
			Q:         coord.Q,
			R:         coord.R,
			S:         coord.S,
			Timestamp: time.Now(),
		}
	}

	// Release lock BEFORE publishing event
	r.mu.Unlock()

	// Log and publish after lock is released
	if occupant != nil {
		log.Info("🎯 Updated tile occupancy",
			zap.String("occupant_type", string(occupant.Type)),
			zap.Strings("occupant_tags", occupant.Tags),
		)

		if eventToPublish != nil {
			events.Publish(r.eventBus, *eventToPublish)
			log.Debug("📢 Published TilePlacedEvent",
				zap.String("player_id", eventToPublish.PlayerID),
				zap.String("tile_type", eventToPublish.TileType))
		}
	} else {
		log.Info("🧹 Cleared tile occupancy")
	}

	return nil
}

// GetTile returns a specific tile by coordinates (immutable getter)
func (r *RepositoryImpl) GetTile(ctx context.Context, coord HexPosition) (*Tile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	board, exists := r.boards[r.gameID]
	if !exists {
		return nil, fmt.Errorf("board not found for game: %s", r.gameID)
	}

	// Find the tile
	for _, tile := range board.Tiles {
		if tile.Coordinates.Q == coord.Q && tile.Coordinates.R == coord.R && tile.Coordinates.S == coord.S {
			// Return a copy to maintain immutability
			tileCopy := tile

			// Deep copy slices and pointers
			tileCopy.Tags = make([]string, len(tile.Tags))
			copy(tileCopy.Tags, tile.Tags)

			tileCopy.Bonuses = make([]TileBonus, len(tile.Bonuses))
			copy(tileCopy.Bonuses, tile.Bonuses)

			if tile.DisplayName != nil {
				displayNameCopy := *tile.DisplayName
				tileCopy.DisplayName = &displayNameCopy
			}

			if tile.OccupiedBy != nil {
				occupantCopy := *tile.OccupiedBy
				occupantCopy.Tags = make([]string, len(tile.OccupiedBy.Tags))
				copy(occupantCopy.Tags, tile.OccupiedBy.Tags)
				tileCopy.OccupiedBy = &occupantCopy
			}

			if tile.OwnerID != nil {
				ownerIDCopy := *tile.OwnerID
				tileCopy.OwnerID = &ownerIDCopy
			}

			return &tileCopy, nil
		}
	}

	return nil, fmt.Errorf("tile not found at coordinates: %s", coord.String())
}
