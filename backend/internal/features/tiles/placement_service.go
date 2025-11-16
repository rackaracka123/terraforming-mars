package tiles

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// PlacementService handles tile placement calculations and validation
// CRITICAL: Takes IDs only, never Game/Player objects
type PlacementService interface {
	// CalculateAvailablePositions calculates available positions for a tile type
	CalculateAvailablePositions(ctx context.Context, tileType string) ([]HexPosition, error)

	// CalculateAvailablePositionsForPlayer calculates available positions considering player restrictions
	CalculateAvailablePositionsForPlayer(ctx context.Context, playerID string, tileType string) ([]HexPosition, error)

	// ValidateTilePlacement checks if a tile can be placed at position
	ValidateTilePlacement(ctx context.Context, playerID string, tileType string, position HexPosition) (bool, error)
}

// PlacementServiceImpl implements PlacementService
type PlacementServiceImpl struct {
	boardService      BoardService
	parametersService parameters.Service
}

// NewPlacementService creates a new PlacementService
func NewPlacementService(boardService BoardService, parametersService parameters.Service) PlacementService {
	return &PlacementServiceImpl{
		boardService:      boardService,
		parametersService: parametersService,
	}
}

// CalculateAvailablePositions calculates available tile positions
func (s *PlacementServiceImpl) CalculateAvailablePositions(ctx context.Context, tileType string) ([]HexPosition, error) {
	switch tileType {
	case "ocean":
		return s.calculateAvailableOceanHexes(ctx)
	case "city":
		return s.calculateAvailableCityHexes(ctx)
	case "greenery":
		return s.calculateAvailableGreeneryHexes(ctx)
	default:
		// Unknown tile types return empty list
		return []HexPosition{}, nil
	}
}

// CalculateAvailablePositionsForPlayer calculates positions for specific player
func (s *PlacementServiceImpl) CalculateAvailablePositionsForPlayer(ctx context.Context, playerID string, tileType string) ([]HexPosition, error) {
	switch tileType {
	case "greenery":
		return s.calculateAvailableGreeneryHexesForPlayer(ctx, playerID)
	default:
		// For non-greenery tiles, use the standard method
		return s.CalculateAvailablePositions(ctx, tileType)
	}
}

// ValidateTilePlacement validates a tile placement
func (s *PlacementServiceImpl) ValidateTilePlacement(ctx context.Context, playerID string, tileType string, position HexPosition) (bool, error) {
	// Get available positions for this player/tile type
	availablePositions, err := s.CalculateAvailablePositionsForPlayer(ctx, playerID, tileType)
	if err != nil {
		return false, fmt.Errorf("failed to get available positions: %w", err)
	}

	// Check if the position is in the available list
	for _, availablePos := range availablePositions {
		if availablePos.Q == position.Q && availablePos.R == position.R && availablePos.S == position.S {
			return true, nil
		}
	}

	return false, nil
}

// calculateAvailableOceanHexes returns available ocean hexes based on board state
func (s *PlacementServiceImpl) calculateAvailableOceanHexes(ctx context.Context) ([]HexPosition, error) {
	// Check if oceans are maxed
	oceansMaxed, err := s.parametersService.IsOceansMaxed(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check oceans maxed: %w", err)
	}
	if oceansMaxed {
		return []HexPosition{}, nil
	}

	// Get board state
	board, err := s.boardService.GetBoard(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	availableHexes := make([]HexPosition, 0)

	// Find unoccupied ocean spaces
	for _, tile := range board.Tiles {
		if tile.Type == domain.ResourceOceanTile {
			if tile.OccupiedBy == nil {
				availableHexes = append(availableHexes, tile.Coordinates)
			}
		}
	}

	return availableHexes, nil
}

// calculateAvailableCityHexes returns available hexes for city placement
func (s *PlacementServiceImpl) calculateAvailableCityHexes(ctx context.Context) ([]HexPosition, error) {
	// Get board state
	board, err := s.boardService.GetBoard(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	logger.Get().Info("🏙️ Calculating city placement hexes",
		zap.Int("total_tiles", len(board.Tiles)))

	availableHexes := make([]HexPosition, 0)

	// Build a map of city positions for adjacency checks
	cityPositions := make(map[string]bool)
	for _, tile := range board.Tiles {
		if tile.OccupiedBy != nil && tile.OccupiedBy.Type == domain.ResourceCityTile {
			cityPositions[tile.Coordinates.String()] = true
		}
	}

	// Check each tile for city placement eligibility
	oceanCount := 0
	occupiedCount := 0
	adjacentToCityCount := 0
	availableCount := 0

	for _, tile := range board.Tiles {
		// Skip ocean tiles
		if tile.Type == domain.ResourceOceanTile {
			oceanCount++
			continue
		}

		// Skip occupied tiles
		if tile.OccupiedBy != nil {
			occupiedCount++
			logger.Get().Debug("🏙️ Tile occupied",
				zap.String("coordinate", tile.Coordinates.String()),
				zap.String("type", string(tile.OccupiedBy.Type)))
			continue
		}

		// Check if any adjacent hex has a city (cities cannot be adjacent to each other)
		if s.isAdjacentToCity(tile.Coordinates, cityPositions) {
			adjacentToCityCount++
			logger.Get().Debug("🏙️ Tile adjacent to city, skipping",
				zap.String("coordinate", tile.Coordinates.String()))
			continue
		}

		// This tile is available for city placement
		availableCount++
		availableHexes = append(availableHexes, tile.Coordinates)
	}

	logger.Get().Info("🏙️ City placement summary",
		zap.Int("ocean_tiles", oceanCount),
		zap.Int("occupied_tiles", occupiedCount),
		zap.Int("adjacent_to_city", adjacentToCityCount),
		zap.Int("available_tiles", availableCount))

	return availableHexes, nil
}

// calculateAvailableGreeneryHexes returns available hexes for greenery placement
func (s *PlacementServiceImpl) calculateAvailableGreeneryHexes(ctx context.Context) ([]HexPosition, error) {
	// Get board state
	board, err := s.boardService.GetBoard(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	logger.Get().Info("🌿 Calculating greenery placement hexes",
		zap.Int("total_tiles", len(board.Tiles)))

	availableHexes := make([]HexPosition, 0)
	oceanCount := 0
	occupiedCount := 0
	availableCount := 0

	// Greenery can be placed on any empty land tile (not ocean tiles)
	for _, tile := range board.Tiles {
		// Skip ocean tiles
		if tile.Type == domain.ResourceOceanTile {
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
		availableHexes = append(availableHexes, tile.Coordinates)
	}

	logger.Get().Info("🌿 Greenery placement summary",
		zap.Int("ocean_tiles", oceanCount),
		zap.Int("occupied_tiles", occupiedCount),
		zap.Int("available_tiles", availableCount))

	return availableHexes, nil
}

// calculateAvailableGreeneryHexesForPlayer returns available hexes for greenery placement
// with preference for tiles adjacent to the player's existing tiles
func (s *PlacementServiceImpl) calculateAvailableGreeneryHexesForPlayer(ctx context.Context, playerID string) ([]HexPosition, error) {
	// Get board state
	board, err := s.boardService.GetBoard(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	logger.Get().Info("🌿 Calculating greenery placement hexes for player",
		zap.String("player_id", playerID))

	adjacentHexes := make([]HexPosition, 0)
	allAvailableHexes := make([]HexPosition, 0)

	// Build a map of player's tile positions
	playerTilePositions := make(map[string]bool)
	for _, tile := range board.Tiles {
		if tile.OwnerID != nil && *tile.OwnerID == playerID {
			playerTilePositions[tile.Coordinates.String()] = true
		}
	}

	logger.Get().Info("🌿 Player tile count",
		zap.Int("count", len(playerTilePositions)))

	// Check each tile for greenery placement eligibility
	for _, tile := range board.Tiles {
		// Skip ocean tiles
		if tile.Type == domain.ResourceOceanTile {
			continue
		}

		// Skip occupied tiles
		if tile.OccupiedBy != nil {
			continue
		}

		// Check if this tile is adjacent to any of the player's tiles
		if s.isAdjacentToPlayerTile(tile.Coordinates, playerTilePositions) {
			adjacentHexes = append(adjacentHexes, tile.Coordinates)
		}

		// Collect all available hexes as fallback
		allAvailableHexes = append(allAvailableHexes, tile.Coordinates)
	}

	// If player has tiles adjacent to empty spaces, return only those
	// This enforces the "must place adjacent if possible" rule
	if len(adjacentHexes) > 0 {
		logger.Get().Info("🌿 Found tiles adjacent to player's tiles",
			zap.Int("count", len(adjacentHexes)))
		return adjacentHexes, nil
	}

	// Otherwise, return all available hexes (player has no tiles or no adjacent spaces)
	logger.Get().Info("🌿 No adjacent tiles found, returning all available",
		zap.Int("count", len(allAvailableHexes)))
	return allAvailableHexes, nil
}

// isAdjacentToCity checks if a hex coordinate is adjacent to any city
func (s *PlacementServiceImpl) isAdjacentToCity(coord HexPosition, cityPositions map[string]bool) bool {
	// Check each adjacent position
	for _, adjacentCoord := range coord.GetNeighbors() {
		if cityPositions[adjacentCoord.String()] {
			return true
		}
	}

	return false
}

// isAdjacentToPlayerTile checks if a hex coordinate is adjacent to any of the player's tiles
func (s *PlacementServiceImpl) isAdjacentToPlayerTile(coord HexPosition, playerTilePositions map[string]bool) bool {
	// Check each adjacent position
	for _, adjacentCoord := range coord.GetNeighbors() {
		if playerTilePositions[adjacentCoord.String()] {
			return true
		}
	}

	return false
}
