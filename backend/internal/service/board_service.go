package service

import (
	"fmt"
	"math"

	"terraforming-mars-backend/internal/model"
)

// BoardService interface defines board-related operations
type BoardService interface {
	GenerateDefaultBoard() model.Board
	CalculateAvailableHexesForTileType(game model.Game, tileType string) ([]string, error)
}

// BoardServiceImpl handles board-related operations
type BoardServiceImpl struct{}

// NewBoardService creates a new board service
func NewBoardService() BoardService {
	return &BoardServiceImpl{}
}

// GenerateDefaultBoard creates the default Mars board with 42 tiles
// Uses the same grid pattern as the frontend: 5-6-7-8-9-8-7-6-5
func (srv *BoardServiceImpl) GenerateDefaultBoard() model.Board {
	tiles := srv.generateTiles()
	return model.Board{
		Tiles: tiles,
	}
}

// generateTiles creates the default tile layout matching the frontend pattern
func (srv *BoardServiceImpl) generateTiles() []model.Tile {
	var tiles []model.Tile

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

			coordinate := model.HexPosition{
				Q: q,
				R: r,
				S: s,
			}

			// Determine if this is an ocean space
			isOceanSpace := srv.isOceanPosition(rowIdx, colIdx)
			tileType := model.ResourceType("empty") // Default type
			if isOceanSpace {
				tileType = model.ResourceOceanTile
			}

			// Calculate resource bonuses for this tile
			bonuses := srv.calculateBonuses(rowIdx, colIdx)

			// Create special tiles with tags
			tags := srv.generateTileTags(coordinate)

			// Create display name for special tiles
			var displayName *string
			if len(tags) > 0 {
				name := srv.getDisplayNameFromTags(tags)
				displayName = &name
			}

			tile := model.Tile{
				Coordinates: coordinate,
				Tags:        tags,
				Type:        tileType,
				Location:    model.TileLocationMars,
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
func (srv *BoardServiceImpl) isOceanPosition(row, col int) bool {
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
func (srv *BoardServiceImpl) calculateBonuses(row, col int) []model.TileBonus {
	bonuses := make([]model.TileBonus, 0)
	tileIndex := row*10 + col

	// Same bonus logic as frontend
	if tileIndex%8 == 0 {
		bonuses = append(bonuses, model.TileBonus{
			Type:   model.ResourceSteel,
			Amount: 2,
		})
	}
	if tileIndex%9 == 0 {
		bonuses = append(bonuses, model.TileBonus{
			Type:   model.ResourceTitanium,
			Amount: 1,
		})
	}
	if tileIndex%11 == 0 {
		bonuses = append(bonuses, model.TileBonus{
			Type:   model.ResourcePlants,
			Amount: 1,
		})
	}
	if tileIndex%13 == 0 {
		bonuses = append(bonuses, model.TileBonus{
			Type:   model.ResourceCardDraw,
			Amount: 1,
		})
	}

	return bonuses
}

// generateTileTags creates special tags for certain tiles
func (srv *BoardServiceImpl) generateTileTags(coord model.HexPosition) []string {
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
func (srv *BoardServiceImpl) getDisplayNameFromTags(tags []string) string {
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
func (srv *BoardServiceImpl) hexToPixel(coord model.HexPosition) (float64, float64) {
	size := 0.3 // Same as frontend HEX_SIZE

	// Pointy-top hex positioning (same as frontend)
	x := size * math.Sqrt(3) * (float64(coord.Q) + float64(coord.R)/2)
	y := (size * 3 / 2) * float64(coord.R)

	return x, y
}

// CalculateAvailableHexesForTileType returns available hexes for a specific tile type
func (srv *BoardServiceImpl) CalculateAvailableHexesForTileType(game model.Game, tileType string) ([]string, error) {
	switch tileType {
	case "ocean":
		return srv.calculateAvailableOceanHexes(game)
	case "city":
		// TODO: Implement city placement logic (need to avoid ocean tiles and check adjacency rules)
		return srv.calculateAvailableCityHexes(game)
	case "greenery":
		// TODO: Implement greenery placement logic (any empty land tile)
		return srv.calculateAvailableGreeneryHexes(game)
	default:
		// Unknown tile types return empty list
		return []string{}, nil
	}
}

// calculateAvailableOceanHexes returns available ocean hexes based on board state
func (srv *BoardServiceImpl) calculateAvailableOceanHexes(game model.Game) ([]string, error) {
	// Count actual oceans placed on board (board is source of truth)
	oceansPlaced := 0
	availableHexes := make([]string, 0)

	for _, tile := range game.Board.Tiles {
		if tile.Type == model.ResourceOceanTile {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type == model.ResourceOceanTile {
				// This ocean space is occupied
				oceansPlaced++
			} else {
				// This ocean space is available for placement
				hexKey := srv.formatHexCoordinate(tile.Coordinates)
				availableHexes = append(availableHexes, hexKey)
			}
		}
	}

	// Check if we've reached maximum oceans based on actual board state
	if oceansPlaced >= model.MaxOceans {
		return []string{}, nil
	}

	return availableHexes, nil
}

// calculateAvailableCityHexes returns available hexes for city placement
func (srv *BoardServiceImpl) calculateAvailableCityHexes(game model.Game) ([]string, error) {
	availableHexes := make([]string, 0)

	fmt.Printf("🏙️ Calculating city placement hexes - Total board tiles: %d\n", len(game.Board.Tiles))

	oceanCount := 0
	occupiedCount := 0
	availableCount := 0

	// For now, allow cities on any empty land tile (not ocean tiles)
	// TODO: Add adjacency rules later (cities cannot be adjacent to each other)
	for _, tile := range game.Board.Tiles {
		// Skip ocean tiles
		if tile.Type == model.ResourceOceanTile {
			oceanCount++
			continue
		}

		// Skip occupied tiles
		if tile.OccupiedBy != nil {
			occupiedCount++
			fmt.Printf("🏙️ Tile %d,%d,%d is occupied by: %v\n", tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S, tile.OccupiedBy.Type)
			continue
		}

		// This tile is available for city placement
		availableCount++
		hexKey := fmt.Sprintf("%d,%d,%d", tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S)
		availableHexes = append(availableHexes, hexKey)
	}

	fmt.Printf("🏙️ City placement summary - Ocean tiles: %d, Occupied tiles: %d, Available tiles: %d\n",
		oceanCount, occupiedCount, availableCount)

	return availableHexes, nil
}

// calculateAvailableGreeneryHexes returns available hexes for greenery placement
func (srv *BoardServiceImpl) calculateAvailableGreeneryHexes(game model.Game) ([]string, error) {
	// TODO: Implement greenery placement logic
	// Greenery can be placed on any empty land tile
	return []string{}, nil
}

// formatHexCoordinate converts hex coordinates to string format
func (srv *BoardServiceImpl) formatHexCoordinate(coord model.HexPosition) string {
	return fmt.Sprintf("%d,%d,%d", coord.Q, coord.R, coord.S)
}
