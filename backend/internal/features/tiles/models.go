package tiles

import "terraforming-mars-backend/internal/model"

// Constants
const (
	MaxOceans = 9 // Maximum number of ocean tiles
)

// NewStandardBoard creates the standard Mars board with 42 hexagonal tiles
// Uses the same 5-6-7-8-9-8-7-6-5 row pattern as the frontend
func NewStandardBoard() model.Board {
	var tiles []model.Tile

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

			tiles = append(tiles, model.Tile{
				Coordinates: model.HexPosition{Q: q, R: r, S: s},
				OccupiedBy:  nil,
				OwnerID:     nil,
				Bonuses:     []model.TileBonus{},
			})
		}
	}

	return model.Board{Tiles: tiles}
}
