package board

import (

	"terraforming-mars-backend/internal/session/types"
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

// ResourceType represents different types of resources (subset needed for tiles)
type ResourceType string

const (
	ResourceOceanTile    ResourceType = "ocean-tile"
	ResourceCityTile     ResourceType = "city-tile"
	ResourceGreeneryTile ResourceType = "greenery-tile"
	ResourceSteel        ResourceType = "steel"
	ResourceTitanium     ResourceType = "titanium"
	ResourcePlants       ResourceType = "plants"
	ResourceCardDraw     ResourceType = "card-draw"
)

// HexPosition is re-exported from types to avoid import cycles
type HexPosition = types.HexPosition

// TileBonus represents a resource bonus provided by a tile when occupied
type TileBonus struct {
	Type   ResourceType `json:"type"`
	Amount int          `json:"amount"`
}

// TileOccupant represents what currently occupies a tile
type TileOccupant struct {
	Type ResourceType `json:"type"`
	Tags []string     `json:"tags"`
}

// Tile represents a single hexagonal tile on the game board
type Tile struct {
	Coordinates HexPosition   `json:"coordinates"`
	Tags        []string      `json:"tags"`
	Type        ResourceType  `json:"type"`
	Location    TileLocation  `json:"location"`
	DisplayName *string       `json:"displayName,omitempty"`
	Bonuses     []TileBonus   `json:"bonuses"`
	OccupiedBy  *TileOccupant `json:"occupiedBy,omitempty"`
	OwnerID     *string       `json:"ownerId,omitempty"`
}

// Board represents the complete game board state
type Board struct {
	Tiles []Tile `json:"tiles"`
}

// DeepCopy creates a deep copy of the board
func (b *Board) DeepCopy() *Board {
	if b == nil {
		return nil
	}

	tiles := make([]Tile, len(b.Tiles))
	for i, tile := range b.Tiles {
		// Copy tile
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

		tiles[i] = tileCopy
	}

	return &Board{Tiles: tiles}
}
