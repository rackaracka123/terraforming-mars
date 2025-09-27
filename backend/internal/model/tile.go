package model

// TileBonus represents a resource bonus provided by a tile when occupied
type TileBonus struct {
	// Type specifies the resource type granted by this bonus
	Type ResourceType `json:"type" ts:"ResourceType"`
	// Amount specifies the quantity of the bonus granted
	Amount int `json:"amount" ts:"number"`
}

// TileOccupant represents what currently occupies a tile
type TileOccupant struct {
	// Type specifies the type of occupant (city-tile, ocean-tile, greenery-tile, etc.)
	Type ResourceType `json:"type" ts:"ResourceType"`
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
	Type ResourceType `json:"type" ts:"ResourceType"`
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
