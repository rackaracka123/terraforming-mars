package service

import (
	"fmt"
	"math"

	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/shared/types"
)

// BoardService interface defines board-related operations
type BoardService interface {
	GenerateDefaultBoard() tiles.Board
	CalculateAvailableHexesForTileType(g game.Game, tileType string) ([]string, error)
	CalculateAvailableHexesForTileTypeWithPlayer(g game.Game, tileType, playerID string) ([]string, error)
}

// BoardServiceImpl handles board-related operations
type BoardServiceImpl struct{}

// NewBoardService creates a new board service
func NewBoardService() BoardService {
	return &BoardServiceImpl{}
}

// GenerateDefaultBoard creates the default Mars board with 42 tiles
// Uses the same grid pattern as the frontend: 5-6-7-8-9-8-7-6-5
func (srv *BoardServiceImpl) GenerateDefaultBoard() tiles.Board {
	allTiles := srv.generateTiles()
	return tiles.Board{
		Tiles: allTiles,
	}
}

// generateTiles creates the default tile layout matching the frontend pattern
func (srv *BoardServiceImpl) generateTiles() []tiles.Tile {
	var allTiles []tiles.Tile

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

			coordinate := types.HexPosition{
				Q: q,
				R: r,
				S: s,
			}

			// Determine if this is an ocean space
			isOceanSpace := srv.isOceanPosition(rowIdx, colIdx)
			tileType := tiles.ResourceType("empty") // Default type
			if isOceanSpace {
				tileType = tiles.ResourceOceanTile
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

			tile := tiles.Tile{
				Coordinates: coordinate,
				Tags:        tags,
				Type:        tileType,
				Location:    tiles.TileLocationMars,
				DisplayName: displayName,
				Bonuses:     bonuses,
				OccupiedBy:  nil, // All tiles start empty
				OwnerID:     nil, // All tiles start unowned
			}

			allTiles = append(allTiles, tile)
		}
	}

	return allTiles
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
func (srv *BoardServiceImpl) calculateBonuses(row, col int) []tiles.TileBonus {
	bonuses := make([]tiles.TileBonus, 0)
	tileIndex := row*10 + col

	// Same bonus logic as frontend
	if tileIndex%8 == 0 {
		bonuses = append(bonuses, tiles.TileBonus{
			Type:   tiles.ResourceType(resources.ResourceSteel),
			Amount: 2,
		})
	}
	if tileIndex%9 == 0 {
		bonuses = append(bonuses, tiles.TileBonus{
			Type:   tiles.ResourceType(resources.ResourceTitanium),
			Amount: 1,
		})
	}
	if tileIndex%11 == 0 {
		bonuses = append(bonuses, tiles.TileBonus{
			Type:   tiles.ResourceType(resources.ResourcePlants),
			Amount: 1,
		})
	}
	if tileIndex%13 == 0 {
		bonuses = append(bonuses, tiles.TileBonus{
			Type:   tiles.ResourceType(resources.ResourceCardDraw),
			Amount: 1,
		})
	}

	return bonuses
}

// generateTileTags creates special tags for certain tiles
func (srv *BoardServiceImpl) generateTileTags(coord types.HexPosition) []string {
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
func (srv *BoardServiceImpl) hexToPixel(coord types.HexPosition) (float64, float64) {
	size := 0.3 // Same as frontend HEX_SIZE

	// Pointy-top hex positioning (same as frontend)
	x := size * math.Sqrt(3) * (float64(coord.Q) + float64(coord.R)/2)
	y := (size * 3 / 2) * float64(coord.R)

	return x, y
}

// CalculateAvailableHexesForTileType returns available hexes for a specific tile type
func (srv *BoardServiceImpl) CalculateAvailableHexesForTileType(g game.Game, tileType string) ([]string, error) {
	switch tileType {
	case "ocean":
		return srv.calculateAvailableOceanHexes(g)
	case "city":
		return srv.calculateAvailableCityHexes(g)
	case "greenery":
		return srv.calculateAvailableGreeneryHexes(g)
	default:
		// Unknown tile types return empty list
		return []string{}, nil
	}
}

// CalculateAvailableHexesForTileTypeWithPlayer returns available hexes with player context
// This is used for greenery placement which requires adjacency to player's tiles
func (srv *BoardServiceImpl) CalculateAvailableHexesForTileTypeWithPlayer(g game.Game, tileType, playerID string) ([]string, error) {
	switch tileType {
	case "greenery":
		return srv.calculateAvailableGreeneryHexesForPlayer(g, playerID)
	default:
		// For non-greenery tiles, use the standard method
		return srv.CalculateAvailableHexesForTileType(g, tileType)
	}
}

// calculateAvailableOceanHexes returns available ocean hexes based on board state
func (srv *BoardServiceImpl) calculateAvailableOceanHexes(g game.Game) ([]string, error) {
	// Count actual oceans placed on board (board is source of truth)
	oceansPlaced := 0
	availableHexes := make([]string, 0)

	board, err := g.GetBoard()
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	for _, tile := range board.Tiles {
		if tile.Type == tiles.ResourceOceanTile {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type == tiles.ResourceOceanTile {
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
	if oceansPlaced >= parameters.MaxOceans {
		return []string{}, nil
	}

	return availableHexes, nil
}

// calculateAvailableCityHexes returns available hexes for city placement
func (srv *BoardServiceImpl) calculateAvailableCityHexes(g game.Game) ([]string, error) {
	availableHexes := make([]string, 0)

	board, err := g.GetBoard()
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	fmt.Printf("🏙️ Calculating city placement hexes - Total board tiles: %d\n", len(board.Tiles))

	oceanCount := 0
	occupiedCount := 0
	adjacentToCityCount := 0
	availableCount := 0

	// Build a map of city positions for adjacency checks
	cityPositions := make(map[string]bool)
	for _, tile := range board.Tiles {
		if tile.OccupiedBy != nil && tile.OccupiedBy.Type == tiles.ResourceCityTile {
			cityPositions[tile.Coordinates.String()] = true
		}
	}

	// Check each tile for city placement eligibility
	for _, tile := range board.Tiles {
		// Skip ocean tiles
		if tile.Type == tiles.ResourceOceanTile {
			oceanCount++
			continue
		}

		// Skip occupied tiles
		if tile.OccupiedBy != nil {
			occupiedCount++
			fmt.Printf("🏙️ Tile %d,%d,%d is occupied by: %v\n", tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S, tile.OccupiedBy.Type)
			continue
		}

		// Check if any adjacent hex has a city (cities cannot be adjacent to each other)
		if srv.isAdjacentToCity(tile.Coordinates, cityPositions) {
			adjacentToCityCount++
			fmt.Printf("🏙️ Tile %d,%d,%d is adjacent to a city, skipping\n", tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S)
			continue
		}

		// This tile is available for city placement
		availableCount++
		availableHexes = append(availableHexes, tile.Coordinates.String())
	}

	fmt.Printf("🏙️ City placement summary - Ocean tiles: %d, Occupied tiles: %d, Adjacent to city: %d, Available tiles: %d\n",
		oceanCount, occupiedCount, adjacentToCityCount, availableCount)

	return availableHexes, nil
}

// calculateAvailableGreeneryHexes returns available hexes for greenery placement
func (srv *BoardServiceImpl) calculateAvailableGreeneryHexes(g game.Game) ([]string, error) {
	availableHexes := make([]string, 0)

	board, err := g.GetBoard()
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	fmt.Printf("🌿 Calculating greenery placement hexes - Total board tiles: %d\n", len(board.Tiles))

	oceanCount := 0
	occupiedCount := 0
	availableCount := 0

	// Greenery can be placed on any empty land tile (not ocean tiles)
	for _, tile := range board.Tiles {
		// Skip ocean tiles
		if tile.Type == tiles.ResourceOceanTile {
			oceanCount++
			continue
		}

		// Skip occupied tiles
		if tile.OccupiedBy != nil {
			occupiedCount++
			continue
		}

		// This tile is available for greenery placement
		availableCount++
		availableHexes = append(availableHexes, tile.Coordinates.String())
	}

	fmt.Printf("🌿 Greenery placement summary - Ocean tiles: %d, Occupied tiles: %d, Available tiles: %d\n",
		oceanCount, occupiedCount, availableCount)

	return availableHexes, nil
}

// formatHexCoordinate converts hex coordinates to string format
func (srv *BoardServiceImpl) formatHexCoordinate(coord types.HexPosition) string {
	return coord.String()
}

// isAdjacentToCity checks if a hex coordinate is adjacent to any city
func (srv *BoardServiceImpl) isAdjacentToCity(coord types.HexPosition, cityPositions map[string]bool) bool {
	// Check each adjacent position
	for _, adjacentCoord := range coord.GetNeighbors() {
		if cityPositions[adjacentCoord.String()] {
			return true
		}
	}

	return false
}

// calculateAvailableGreeneryHexesForPlayer returns available hexes for greenery placement
// with preference for tiles adjacent to the player's existing tiles
func (srv *BoardServiceImpl) calculateAvailableGreeneryHexesForPlayer(g game.Game, playerID string) ([]string, error) {
	adjacentHexes := make([]string, 0)
	allAvailableHexes := make([]string, 0)

	fmt.Printf("🌿 Calculating greenery placement hexes for player %s\n", playerID)

	board, err := g.GetBoard()
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	// Build a map of player's tile positions
	playerTilePositions := make(map[string]bool)
	for _, tile := range board.Tiles {
		if tile.OwnerID != nil && *tile.OwnerID == playerID {
			playerTilePositions[tile.Coordinates.String()] = true
		}
	}

	fmt.Printf("🌿 Player has %d tiles on the board\n", len(playerTilePositions))

	// Check each tile for greenery placement eligibility
	for _, tile := range board.Tiles {
		// Skip ocean tiles
		if tile.Type == tiles.ResourceOceanTile {
			continue
		}

		// Skip occupied tiles
		if tile.OccupiedBy != nil {
			continue
		}

		// Check if this tile is adjacent to any of the player's tiles
		if srv.isAdjacentToPlayerTile(tile.Coordinates, playerTilePositions) {
			adjacentHexes = append(adjacentHexes, tile.Coordinates.String())
		}

		// Collect all available hexes as fallback
		allAvailableHexes = append(allAvailableHexes, tile.Coordinates.String())
	}

	// If player has tiles adjacent to empty spaces, return only those
	// This enforces the "must place adjacent if possible" rule
	if len(adjacentHexes) > 0 {
		fmt.Printf("🌿 Found %d tiles adjacent to player's tiles (returning only these)\n", len(adjacentHexes))
		return adjacentHexes, nil
	}

	// Otherwise, return all available hexes (player has no tiles or no adjacent spaces)
	fmt.Printf("🌿 No adjacent tiles found, returning all %d available tiles\n", len(allAvailableHexes))
	return allAvailableHexes, nil
}

// isAdjacentToPlayerTile checks if a hex coordinate is adjacent to any of the player's tiles
func (srv *BoardServiceImpl) isAdjacentToPlayerTile(coord types.HexPosition, playerTilePositions map[string]bool) bool {
	// Check each adjacent position
	for _, adjacentCoord := range coord.GetNeighbors() {
		if playerTilePositions[adjacentCoord.String()] {
			return true
		}
	}

	return false
}
