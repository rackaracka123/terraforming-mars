package service

import (
	"math"

	"terraforming-mars-backend/internal/model"
)

// BoardService interface defines board-related operations
type BoardService interface {
	GenerateDefaultBoard() model.Board
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
			q := colIdx - hexCount/2 - r/2
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
