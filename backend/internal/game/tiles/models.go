package tiles

import "fmt"

// Constants
const (
	MaxOceans = 9 // Maximum number of ocean tiles
)

// ResourceType represents different types of resources and tiles in the game
type ResourceType string

const (
	ResourceSteel        ResourceType = "steel"
	ResourceTitanium     ResourceType = "titanium"
	ResourcePlants       ResourceType = "plants"
	ResourceOceanTile    ResourceType = "ocean-tile"
	ResourceCityTile     ResourceType = "city-tile"
	ResourceGreeneryTile ResourceType = "greenery-tile"
	ResourceCardDraw     ResourceType = "card-draw"
)

// HexPosition represents a position in cube coordinates on the hexagonal board
type HexPosition struct {
	Q int `json:"q"`
	R int `json:"r"`
	S int `json:"s"`
}

// String returns a string representation of the hex position
func (h HexPosition) String() string {
	return fmt.Sprintf("(%d,%d,%d)", h.Q, h.R, h.S)
}

// Equals checks if two hex positions are equal
func (h HexPosition) Equals(other HexPosition) bool {
	return h.Q == other.Q && h.R == other.R && h.S == other.S
}

// GetNeighbors returns all adjacent hex positions
func (h HexPosition) GetNeighbors() []HexPosition {
	return []HexPosition{
		{Q: h.Q + 1, R: h.R - 1, S: h.S}, // Northeast
		{Q: h.Q + 1, R: h.R, S: h.S - 1}, // East
		{Q: h.Q, R: h.R + 1, S: h.S - 1}, // Southeast
		{Q: h.Q - 1, R: h.R + 1, S: h.S}, // Southwest
		{Q: h.Q - 1, R: h.R, S: h.S + 1}, // West
		{Q: h.Q, R: h.R - 1, S: h.S + 1}, // Northwest
	}
}

// TileOccupant represents what occupies a tile (e.g., city, greenery, ocean)
type TileOccupant struct {
	Type ResourceType `json:"type"`
	Tags []string     `json:"tags"`
}

// TileBonus represents a bonus that a tile provides when placed
type TileBonus struct {
	Type   ResourceType `json:"type"`
	Amount int          `json:"amount"`
}

// Tile represents a single hexagonal tile on the Mars board
type Tile struct {
	Coordinates HexPosition   `json:"coordinates"`
	OccupiedBy  *TileOccupant `json:"occupiedBy"`
	OwnerID     *string       `json:"ownerId"`
	Bonuses     []TileBonus   `json:"bonuses"`
}

// Board represents the Mars game board
type Board struct {
	Tiles []Tile `json:"tiles"`
}

// Game represents the minimal game state needed for tile management
type Game struct {
	ID    string
	Board Board
}

// Player represents player state needed for tile management
type Player struct {
	ID          string
	Resources   Resources
	PlayedCards []string
}

// Resources represents a player's resource state
type Resources struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// PendingTileSelection represents a tile selection awaiting player input
type PendingTileSelection struct {
	TileType       string   `json:"tileType"`
	Source         string   `json:"source"`
	AvailableHexes []string `json:"availableHexes"`
}

// PendingTileSelectionQueue represents a queue of tiles to be placed
type PendingTileSelectionQueue struct {
	Items  []string `json:"items"`
	Source string   `json:"source"`
}

// TileTypeToResourceType converts a tile type string to ResourceType
func TileTypeToResourceType(tileType string) (ResourceType, error) {
	switch tileType {
	case "ocean":
		return ResourceOceanTile, nil
	case "greenery":
		return ResourceGreeneryTile, nil
	case "city":
		return ResourceCityTile, nil
	default:
		return "", fmt.Errorf("unknown tile type: %s", tileType)
	}
}
