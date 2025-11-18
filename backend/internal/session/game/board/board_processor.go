package board

import (
	"math"
)

// BoardProcessor handles board generation and calculation logic
type BoardProcessor struct{}

// NewBoardProcessor creates a new board processor
func NewBoardProcessor() *BoardProcessor {
	return &BoardProcessor{}
}

// GenerateTiles creates the default 42-tile Mars board layout
// Uses the same grid pattern as the frontend: 5-6-7-8-9-8-7-6-5
func (bp *BoardProcessor) GenerateTiles() []Tile {
	var tiles []Tile

	// Row pattern: 5, 6, 7, 8, 9, 8, 7, 6, 5 (matches frontend HexGrid2D)
	rowPattern := []int{5, 6, 7, 8, 9, 8, 7, 6, 5}

	for rowIdx := 0; rowIdx < len(rowPattern); rowIdx++ {
		hexCount := rowPattern[rowIdx]
		r := rowIdx - len(rowPattern)/2 // Center the rows: -4 to +4

		for colIdx := 0; colIdx < hexCount; colIdx++ {
			// Calculate axial coordinates for honeycomb pattern (same as frontend)
			// Use integer division that matches Math.floor behavior
			q := colIdx - hexCount/2
			if r < 0 {
				// For negative r values, we need to subtract the floor division
				q = q - (r-1)/2
			} else {
				// For positive r values, regular integer division works
				q = q - r/2
			}
			s := -q - r

			coordinate := HexPosition{
				Q: q,
				R: r,
				S: s,
			}

			// Determine if this is an ocean space
			isOceanSpace := bp.isOceanPosition(rowIdx, colIdx)
			tileType := ResourceType("empty") // Default type
			if isOceanSpace {
				tileType = ResourceOceanTile
			}

			// Calculate resource bonuses for this tile
			bonuses := bp.calculateBonuses(rowIdx, colIdx)

			// Create special tiles with tags
			tags := bp.generateTileTags(coordinate)

			// Create display name for special tiles
			var displayName *string
			if len(tags) > 0 {
				name := bp.getDisplayNameFromTags(tags)
				displayName = &name
			}

			tile := Tile{
				Coordinates: coordinate,
				Tags:        tags,
				Type:        tileType,
				Location:    TileLocationMars,
				DisplayName: displayName,
				Bonuses:     bonuses,
				OccupiedBy:  nil, // All tiles start empty
				OwnerID:     nil, // All tiles start unowned
			}

			tiles = append(tiles, tile)
		}
	}

	return tiles
}

// isOceanPosition determines if a tile should be an ocean space (matches frontend logic)
func (bp *BoardProcessor) isOceanPosition(row, col int) bool {
	oceanPositions := []struct{ row, col int }{
		{1, 2}, {2, 1}, {2, 5}, {3, 3}, {4, 1},
		{4, 7}, {5, 4}, {6, 2}, {7, 3},
	}

	for _, pos := range oceanPositions {
		if pos.row == row && pos.col == col {
			return true
		}
	}
	return false
}

// calculateBonuses generates resource bonuses for tiles (matches frontend logic)
func (bp *BoardProcessor) calculateBonuses(row, col int) []TileBonus {
	bonuses := make([]TileBonus, 0)
	tileIndex := row*10 + col

	// Same bonus logic as frontend
	if tileIndex%8 == 0 {
		bonuses = append(bonuses, TileBonus{
			Type:   ResourceSteel,
			Amount: 2,
		})
	}
	if tileIndex%9 == 0 {
		bonuses = append(bonuses, TileBonus{
			Type:   ResourceTitanium,
			Amount: 1,
		})
	}
	if tileIndex%11 == 0 {
		bonuses = append(bonuses, TileBonus{
			Type:   ResourcePlants,
			Amount: 1,
		})
	}
	if tileIndex%13 == 0 {
		bonuses = append(bonuses, TileBonus{
			Type:   ResourceCardDraw,
			Amount: 1,
		})
	}

	return bonuses
}

// generateTileTags creates special tags for certain tiles
func (bp *BoardProcessor) generateTileTags(coord HexPosition) []string {
	tags := make([]string, 0)

	// Add Noctis City location (example special placement)
	// This is roughly in the center-left area of the board
	if coord.Q == -2 && coord.R == 0 && coord.S == 2 {
		tags = append(tags, "noctis-city")
	}

	// Add other special locations as needed
	// Example: Tharsis locations, polar areas, etc.

	return tags
}

// getDisplayNameFromTags returns a human-readable name for special tiles
func (bp *BoardProcessor) getDisplayNameFromTags(tags []string) string {
	for _, tag := range tags {
		switch tag {
		case "noctis-city":
			return "Noctis City"
			// Add other special location names
		}
	}
	return ""
}

// hexToPixel converts hex coordinates to 2D position (for reference, not used in backend)
// Kept for documentation of the coordinate system
func (bp *BoardProcessor) hexToPixel(coord HexPosition) (float64, float64) {
	size := 0.3 // Same as frontend HEX_SIZE

	// Pointy-top hex positioning (same as frontend)
	x := size * math.Sqrt(3) * (float64(coord.Q) + float64(coord.R)/2)
	y := (size * 3 / 2) * float64(coord.R)

	return x, y
}

// CalculateAvailableHexesForTileType returns available hexes for a specific tile type
func (bp *BoardProcessor) CalculateAvailableHexesForTileType(board *Board, tileType string, playerID *string) []string {
	switch tileType {
	case TileTypeOcean:
		return bp.calculateAvailableOceanHexes(board)
	case TileTypeCity:
		return bp.calculateAvailableCityHexes(board)
	case TileTypeGreenery:
		if playerID != nil {
			return bp.calculateAvailableGreeneryHexesForPlayer(board, *playerID)
		}
		return bp.calculateAvailableGreeneryHexes(board)
	default:
		// Unknown tile types return empty list
		return []string{}
	}
}

// calculateAvailableOceanHexes returns available ocean hexes based on board state
func (bp *BoardProcessor) calculateAvailableOceanHexes(board *Board) []string {
	const MaxOceans = 9

	// Count actual oceans placed on board (board is source of truth)
	oceansPlaced := 0
	availableHexes := make([]string, 0)

	for _, tile := range board.Tiles {
		if tile.Type == ResourceOceanTile {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type == ResourceOceanTile {
				// This ocean space is occupied
				oceansPlaced++
			} else {
				// This ocean space is available for placement
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}
		}
	}

	// Check if we've reached maximum oceans based on actual board state
	if oceansPlaced >= MaxOceans {
		return []string{}
	}

	return availableHexes
}

// calculateAvailableCityHexes returns available hexes for city placement
func (bp *BoardProcessor) calculateAvailableCityHexes(board *Board) []string {
	availableHexes := make([]string, 0)

	// Build a map of city positions for adjacency checks
	cityPositions := make(map[string]bool)
	for _, tile := range board.Tiles {
		if tile.OccupiedBy != nil && tile.OccupiedBy.Type == ResourceCityTile {
			cityPositions[tile.Coordinates.String()] = true
		}
	}

	// Check each tile for city placement eligibility
	for _, tile := range board.Tiles {
		// Skip ocean tiles
		if tile.Type == ResourceOceanTile {
			continue
		}

		// Skip occupied tiles
		if tile.OccupiedBy != nil {
			continue
		}

		// Check if any adjacent hex has a city (cities cannot be adjacent to each other)
		if bp.isAdjacentToCity(tile.Coordinates, cityPositions) {
			continue
		}

		// This tile is available for city placement
		availableHexes = append(availableHexes, tile.Coordinates.String())
	}

	return availableHexes
}

// calculateAvailableGreeneryHexes returns available hexes for greenery placement
func (bp *BoardProcessor) calculateAvailableGreeneryHexes(board *Board) []string {
	availableHexes := make([]string, 0)

	// Greenery can be placed on any empty land tile (not ocean tiles)
	for _, tile := range board.Tiles {
		// Skip ocean tiles
		if tile.Type == ResourceOceanTile {
			continue
		}

		// Skip occupied tiles
		if tile.OccupiedBy != nil {
			continue
		}

		// This tile is available for greenery placement
		availableHexes = append(availableHexes, tile.Coordinates.String())
	}

	return availableHexes
}

// calculateAvailableGreeneryHexesForPlayer returns available hexes for greenery placement
// with preference for tiles adjacent to the player's existing tiles
func (bp *BoardProcessor) calculateAvailableGreeneryHexesForPlayer(board *Board, playerID string) []string {
	adjacentHexes := make([]string, 0)
	allAvailableHexes := make([]string, 0)

	// Build a map of player's tile positions
	playerTilePositions := make(map[string]bool)
	for _, tile := range board.Tiles {
		if tile.OwnerID != nil && *tile.OwnerID == playerID {
			playerTilePositions[tile.Coordinates.String()] = true
		}
	}

	// Check each tile for greenery placement eligibility
	for _, tile := range board.Tiles {
		// Skip ocean tiles
		if tile.Type == ResourceOceanTile {
			continue
		}

		// Skip occupied tiles
		if tile.OccupiedBy != nil {
			continue
		}

		// Check if this tile is adjacent to any of the player's tiles
		if bp.isAdjacentToPlayerTile(tile.Coordinates, playerTilePositions) {
			adjacentHexes = append(adjacentHexes, tile.Coordinates.String())
		}

		// Collect all available hexes as fallback
		allAvailableHexes = append(allAvailableHexes, tile.Coordinates.String())
	}

	// If player has tiles adjacent to empty spaces, return only those
	// This enforces the "must place adjacent if possible" rule
	if len(adjacentHexes) > 0 {
		return adjacentHexes
	}

	// Otherwise, return all available hexes (player has no tiles or no adjacent spaces)
	return allAvailableHexes
}

// isAdjacentToCity checks if a hex coordinate is adjacent to any city
func (bp *BoardProcessor) isAdjacentToCity(coord HexPosition, cityPositions map[string]bool) bool {
	// Check each adjacent position
	for _, adjacentCoord := range coord.GetNeighbors() {
		if cityPositions[adjacentCoord.String()] {
			return true
		}
	}

	return false
}

// isAdjacentToPlayerTile checks if a hex coordinate is adjacent to any of the player's tiles
func (bp *BoardProcessor) isAdjacentToPlayerTile(coord HexPosition, playerTilePositions map[string]bool) bool {
	// Check each adjacent position
	for _, adjacentCoord := range coord.GetNeighbors() {
		if playerTilePositions[adjacentCoord.String()] {
			return true
		}
	}

	return false
}
