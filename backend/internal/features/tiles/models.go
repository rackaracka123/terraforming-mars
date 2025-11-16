package tiles

import (
	"fmt"

	"terraforming-mars-backend/internal/domain"
)

// HexPosition represents a position on the Mars board using cube coordinates
type HexPosition struct {
	Q int `json:"q" ts:"number"` // Column coordinate
	R int `json:"r" ts:"number"` // Row coordinate
	S int `json:"s" ts:"number"` // Third coordinate (Q + R + S = 0)
}

// GetNeighbors returns all 6 adjacent hex positions using cube coordinate system
func (h HexPosition) GetNeighbors() []HexPosition {
	// Six adjacent directions in cube coordinates (hexagonal grid)
	directions := []HexPosition{
		{Q: 1, R: -1, S: 0}, // East
		{Q: 1, R: 0, S: -1}, // Southeast
		{Q: 0, R: 1, S: -1}, // Southwest
		{Q: -1, R: 1, S: 0}, // West
		{Q: -1, R: 0, S: 1}, // Northwest
		{Q: 0, R: -1, S: 1}, // Northeast
	}

	neighbors := make([]HexPosition, 6)
	for i, dir := range directions {
		neighbors[i] = HexPosition{
			Q: h.Q + dir.Q,
			R: h.R + dir.R,
			S: h.S + dir.S,
		}
	}
	return neighbors
}

// Equals checks if two hex positions are equal
func (h HexPosition) Equals(other HexPosition) bool {
	return h.Q == other.Q && h.R == other.R && h.S == other.S
}

// String returns the hex position as a formatted string "q,r,s"
func (h HexPosition) String() string {
	return fmt.Sprintf("%d,%d,%d", h.Q, h.R, h.S)
}

// Tile type string constants for placement operations
const (
	TileTypeCity     = "city"
	TileTypeGreenery = "greenery"
	TileTypeOcean    = "ocean"
)

// TileTypeToResourceType converts a tile type string to its corresponding ResourceType
func TileTypeToResourceType(tileType string) (domain.ResourceType, error) {
	switch tileType {
	case TileTypeCity:
		return domain.ResourceCityTile, nil
	case TileTypeGreenery:
		return domain.ResourceGreeneryTile, nil
	case TileTypeOcean:
		return domain.ResourceOceanTile, nil
	default:
		return "", fmt.Errorf("unknown tile type: %s", tileType)
	}
}

// TileBonus represents a resource bonus provided by a tile when occupied
type TileBonus struct {
	// Type specifies the resource type granted by this bonus
	Type domain.ResourceType `json:"type" ts:"ResourceType"`
	// Amount specifies the quantity of the bonus granted
	Amount int `json:"amount" ts:"number"`
}

// TileOccupant represents what currently occupies a tile
type TileOccupant struct {
	// Type specifies the type of occupant (city-tile, ocean-tile, greenery-tile, etc.)
	Type domain.ResourceType `json:"type" ts:"ResourceType"`
	// Tags specifies special properties of the occupant (e.g., "capital" for cities)
	Tags []string `json:"tags" ts:"string[]"`
}

// Tile represents a single hexagonal tile on the game board
type Tile struct {
	// Coordinates specifies the hex position of this tile
	Coordinates HexPosition `json:"coordinates" ts:"HexPosition"`
	// Tags specifies special properties for placement restrictions (e.g., "noctis-city")
	Tags []string `json:"tags" ts:"string[]"`
	// Type specifies the base type of tile (ocean-tile for ocean spaces, etc.)
	Type domain.ResourceType `json:"type" ts:"ResourceType"`
	// Location specifies which celestial body this tile is on
	Location TileLocation `json:"location" ts:"TileLocation"`
	// DisplayName specifies the optional display name shown on the tile
	DisplayName *string `json:"displayName,omitempty" ts:"string|null"`
	// Bonuses specifies the resource bonuses provided by this tile
	Bonuses []TileBonus `json:"bonuses" ts:"TileBonus[]"`
	// OccupiedBy specifies what currently occupies this tile, if anything
	OccupiedBy *TileOccupant `json:"occupiedBy,omitempty" ts:"TileOccupant|null"`
	// OwnerID specifies the player who owns this tile, if any
	OwnerID *string `json:"ownerId,omitempty" ts:"string|null"`
}

// TileLocation represents the celestial body where tiles are located
type TileLocation string

const (
	// TileLocationMars represents tiles on the Mars surface
	TileLocationMars TileLocation = "mars"
)

// Board represents the complete game board state
type Board struct {
	// Tiles contains all tiles on the game board
	Tiles []Tile `json:"tiles" ts:"Tile[]"`
}

// NewStandardBoard creates the standard Mars board with 42 hexagonal tiles
// Uses the same 5-6-7-8-9-8-7-6-5 row pattern as the frontend
func NewStandardBoard() Board {
	var boardTiles []Tile

	// Row pattern matches frontend HexGrid2D
	rowPattern := []int{5, 6, 7, 8, 9, 8, 7, 6, 5}

	for rowIdx := 0; rowIdx < len(rowPattern); rowIdx++ {
		hexCount := rowPattern[rowIdx]
		r := rowIdx - len(rowPattern)/2 // Center rows: -4 to +4

		for colIdx := 0; colIdx < hexCount; colIdx++ {
			// Calculate axial coordinates
			q := colIdx - hexCount/2
			if r < 0 {
				q = q - (r-1)/2
			} else {
				q = q - r/2
			}
			s := -(q + r)

			boardTiles = append(boardTiles, Tile{
				Coordinates: HexPosition{Q: q, R: r, S: s},
				OccupiedBy:  nil,
				OwnerID:     nil,
				Bonuses:     []TileBonus{},
			})
		}
	}

	return Board{Tiles: boardTiles}
}

// Constants
const (
	MaxOceans = 9 // Maximum number of ocean tiles
)
