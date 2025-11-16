package tiles

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// BoardService handles board tile operations
//
// Scope: Isolated board management for a game
//   - Board generation
//   - Tile placement
//   - Tile occupancy checking
//   - Board state retrieval
type BoardService interface {
	// Board generation
	GenerateDefaultBoard() Board

	// Runtime board operations
	GetBoard(ctx context.Context) (Board, error)
	PlaceTile(ctx context.Context, coordinate HexPosition, occupant TileOccupant, ownerID *string) error
	GetTile(ctx context.Context, coordinate HexPosition) (*Tile, error)
	IsTileOccupied(ctx context.Context, coordinate HexPosition) (bool, error)
}

// BoardServiceImpl implements the board service
type BoardServiceImpl struct {
	repo BoardRepository
}

// NewBoardService creates a new board service
func NewBoardService(repo BoardRepository) BoardService {
	return &BoardServiceImpl{
		repo: repo,
	}
}

// GetBoard retrieves the complete board
func (s *BoardServiceImpl) GetBoard(ctx context.Context) (Board, error) {
	return s.repo.GetBoard(ctx)
}

// PlaceTile places a tile on the board
func (s *BoardServiceImpl) PlaceTile(ctx context.Context, coordinate HexPosition, occupant TileOccupant, ownerID *string) error {
	if err := s.repo.PlaceTile(ctx, coordinate, occupant, ownerID); err != nil {
		return fmt.Errorf("failed to place tile: %w", err)
	}

	logger.Get().Info("🏗️ Tile placed",
		zap.String("coordinate", coordinate.String()),
		zap.String("type", string(occupant.Type)),
		zap.Stringp("owner_id", ownerID))

	// TODO Phase 6: Publish TilePlacedEvent

	return nil
}

// GetTile retrieves a specific tile
func (s *BoardServiceImpl) GetTile(ctx context.Context, coordinate HexPosition) (*Tile, error) {
	tile, err := s.repo.GetTile(ctx, coordinate)
	if err != nil {
		return nil, fmt.Errorf("failed to get tile: %w", err)
	}
	return tile, nil
}

// IsTileOccupied checks if a tile is occupied
func (s *BoardServiceImpl) IsTileOccupied(ctx context.Context, coordinate HexPosition) (bool, error) {
	occupied, err := s.repo.IsTileOccupied(ctx, coordinate)
	if err != nil {
		return false, fmt.Errorf("failed to check tile occupancy: %w", err)
	}
	return occupied, nil
}

// GenerateDefaultBoard creates the default Mars board with 42 tiles
// Uses the same grid pattern as the frontend: 5-6-7-8-9-8-7-6-5
func (s *BoardServiceImpl) GenerateDefaultBoard() Board {
	allTiles := s.generateTiles()
	return Board{
		Tiles: allTiles,
	}
}

// generateTiles creates the default tile layout matching the frontend pattern
func (s *BoardServiceImpl) generateTiles() []Tile {
	var allTiles []Tile

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
			coordS := -q - r

			coordinate := HexPosition{
				Q: q,
				R: r,
				S: coordS,
			}

			// Determine if this is an ocean space
			isOceanSpace := s.isOceanPosition(rowIdx, colIdx)
			tileType := domain.ResourceType("empty") // Default type
			if isOceanSpace {
				tileType = domain.ResourceOceanTile
			}

			// Calculate resource bonuses for this tile
			bonuses := s.calculateBonuses(rowIdx, colIdx)

			// Create special tiles with tags
			tags := s.generateTileTags(coordinate)

			// Create display name for special tiles
			var displayName *string
			if len(tags) > 0 {
				name := s.getDisplayNameFromTags(tags)
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

			allTiles = append(allTiles, tile)
		}
	}

	return allTiles
}

// isOceanPosition determines if a tile should be an ocean space (matches frontend logic)
func (s *BoardServiceImpl) isOceanPosition(row, col int) bool {
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
func (s *BoardServiceImpl) calculateBonuses(row, col int) []TileBonus {
	bonuses := make([]TileBonus, 0)
	tileIndex := row*10 + col

	// Same bonus logic as frontend
	if tileIndex%8 == 0 {
		bonuses = append(bonuses, TileBonus{
			Type:   domain.ResourceType(domain.ResourceSteel),
			Amount: 2,
		})
	}
	if tileIndex%9 == 0 {
		bonuses = append(bonuses, TileBonus{
			Type:   domain.ResourceType(domain.ResourceTitanium),
			Amount: 1,
		})
	}
	if tileIndex%11 == 0 {
		bonuses = append(bonuses, TileBonus{
			Type:   domain.ResourceType(domain.ResourcePlants),
			Amount: 1,
		})
	}
	if tileIndex%13 == 0 {
		bonuses = append(bonuses, TileBonus{
			Type:   domain.ResourceType(domain.ResourceCardDraw),
			Amount: 1,
		})
	}

	return bonuses
}

// generateTileTags creates special tags for certain tiles
func (s *BoardServiceImpl) generateTileTags(coord HexPosition) []string {
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
func (s *BoardServiceImpl) getDisplayNameFromTags(tags []string) string {
	for _, tag := range tags {
		switch tag {
		case "noctis-city":
			return "Noctis City"
			// Add other special location names
		}
	}
	return ""
}
