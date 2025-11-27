package board

import (
	"context"
	"fmt"
	"sync"
	"terraforming-mars-backend/internal/game/shared"

	"terraforming-mars-backend/internal/events"
)

// Tile type string constants for placement operations
const (
	TileTypeCity     = "city"
	TileTypeGreenery = "greenery"
	TileTypeOcean    = "ocean"
)

// TileLocation represents the celestial body where tiles are located
type TileLocation string

const (
	// TileLocationMars represents tiles on the Mars surface
	TileLocationMars TileLocation = "mars"
)

// TileBonus represents a resource bonus provided by a tile when occupied
type TileBonus struct {
	Type   shared.ResourceType `json:"type"`
	Amount int                 `json:"amount"`
}

// TileOccupant represents what currently occupies a tile
type TileOccupant struct {
	Type shared.ResourceType `json:"type"`
	Tags []string            `json:"tags"`
}

// Tile represents a single hexagonal tile on the game board
type Tile struct {
	Coordinates shared.HexPosition  `json:"coordinates"`
	Tags        []string            `json:"tags"`
	Type        shared.ResourceType `json:"type"`
	Location    TileLocation        `json:"location"`
	DisplayName *string             `json:"displayName,omitempty"`
	Bonuses     []TileBonus         `json:"bonuses"`
	OccupiedBy  *TileOccupant       `json:"occupiedBy,omitempty"`
	OwnerID     *string             `json:"ownerId,omitempty"`
}

// Board represents the complete game board state with encapsulated tiles
type Board struct {
	mu       sync.RWMutex
	gameID   string
	tiles    []Tile
	eventBus *events.EventBusImpl
}

// NewBoard creates a new empty board
func NewBoard(gameID string, eventBus *events.EventBusImpl) *Board {
	return &Board{
		gameID:   gameID,
		tiles:    []Tile{},
		eventBus: eventBus,
	}
}

// NewBoardWithTiles creates a new board with the provided tiles
func NewBoardWithTiles(gameID string, tiles []Tile, eventBus *events.EventBusImpl) *Board {
	tilesCopy := make([]Tile, len(tiles))
	copy(tilesCopy, tiles)
	return &Board{
		gameID:   gameID,
		tiles:    tilesCopy,
		eventBus: eventBus,
	}
}

// ================== Getters ==================

// Tiles returns a deep copy of all tiles to prevent external mutation
func (b *Board) Tiles() []Tile {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.deepCopyTiles()
}

// GetTile returns a copy of a specific tile by coordinates
func (b *Board) GetTile(coords shared.HexPosition) (*Tile, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for i := range b.tiles {
		if b.tiles[i].Coordinates == coords {
			tileCopy := b.deepCopyTile(&b.tiles[i])
			return tileCopy, nil
		}
	}

	return nil, fmt.Errorf("tile not found at coordinates %v", coords)
}

// ================== Setters with Event Publishing ==================

// SetTiles replaces all tiles (used for board generation)
func (b *Board) SetTiles(ctx context.Context, tiles []Tile) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.Lock()
	b.tiles = make([]Tile, len(tiles))
	copy(b.tiles, tiles)
	b.mu.Unlock()

	return nil
}

// UpdateTileOccupancy updates a tile's occupancy state and publishes TilePlacedEvent
func (b *Board) UpdateTileOccupancy(ctx context.Context, coords shared.HexPosition, occupant TileOccupant, ownerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var found bool

	b.mu.Lock()
	for i := range b.tiles {
		if b.tiles[i].Coordinates == coords {
			b.tiles[i].OccupiedBy = &occupant
			b.tiles[i].OwnerID = &ownerID
			found = true
			break
		}
	}
	b.mu.Unlock()

	if !found {
		return fmt.Errorf("tile not found at coordinates %v", coords)
	}

	// Publish event AFTER releasing lock
	if b.eventBus != nil {
		events.Publish(b.eventBus, events.TilePlacedEvent{
			GameID:   b.gameID,
			PlayerID: ownerID,
			TileType: string(occupant.Type),
			Q:        coords.Q,
			R:        coords.R,
			S:        coords.S,
		})
		// Trigger client broadcast
		events.Publish(b.eventBus, events.BroadcastEvent{
			GameID:    b.gameID,
			PlayerIDs: nil, // Broadcast to all players
		})
	}

	return nil
}

// ================== Helper Methods ==================

// deepCopyTiles creates a deep copy of all tiles
func (b *Board) deepCopyTiles() []Tile {
	tiles := make([]Tile, len(b.tiles))
	for i := range b.tiles {
		tiles[i] = *b.deepCopyTile(&b.tiles[i])
	}
	return tiles
}

// deepCopyTile creates a deep copy of a single tile
func (b *Board) deepCopyTile(tile *Tile) *Tile {
	tileCopy := *tile

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

	return &tileCopy
}
