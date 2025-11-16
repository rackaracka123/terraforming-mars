package tiles

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// PlacementResult contains the outcome of a tile placement operation
type PlacementResult struct {
	Coordinate HexPosition
	TileType   string
	Bonuses    []TilePlacementBonus
}

// TilePlacementBonus represents a bonus awarded from tile placement
type TilePlacementBonus struct {
	Type   domain.ResourceType // steel, titanium, plants, cards, etc.
	Amount int
}

// SelectionService handles tile selection and placement operations
// IMPORTANT: This is a PURE feature service - NO Player/Game knowledge
// The Action layer orchestrates getting pending selections and awarding bonuses
type SelectionService interface {
	// ProcessTileSelection validates coordinate and places tile on board
	// Returns placement result with bonuses that the Action can award
	ProcessTileSelection(
		ctx context.Context,
		coordinate HexPosition,
		tileType string,
		ownerID *string,
	) (*PlacementResult, error)
}

// SelectionServiceImpl implements the tile selection service
type SelectionServiceImpl struct {
	boardService     BoardService
	placementService PlacementService
}

// NewSelectionService creates a new selection service
func NewSelectionService(boardService BoardService, placementService PlacementService) SelectionService {
	return &SelectionServiceImpl{
		boardService:     boardService,
		placementService: placementService,
	}
}

// ProcessTileSelection validates and places a tile on the board
// This is a PURE domain operation - validates placement and updates board state
// The Action layer handles:
// - Getting pending selection from player
// - Validating coordinate is in available hexes
// - Awarding bonuses from result
// - Special handling (plant conversion oxygen/TR)
// - Clearing pending selection
func (s *SelectionServiceImpl) ProcessTileSelection(
	ctx context.Context,
	coordinate HexPosition,
	tileType string,
	ownerID *string,
) (*PlacementResult, error) {
	log := logger.Get()
	log.Debug("🎯 Processing tile placement",
		zap.String("coordinate", coordinate.String()),
		zap.String("tile_type", tileType),
		zap.Stringp("owner_id", ownerID))

	// Validate the tile can be placed at this coordinate
	// Note: The Action already validated coordinate is in available hexes
	// This is a double-check for tile occupancy
	occupied, err := s.boardService.IsTileOccupied(ctx, coordinate)
	if err != nil {
		return nil, fmt.Errorf("failed to check tile occupancy: %w", err)
	}
	if occupied {
		return nil, fmt.Errorf("tile at %s is already occupied", coordinate.String())
	}

	// Get the tile to determine bonuses
	tile, err := s.boardService.GetTile(ctx, coordinate)
	if err != nil {
		return nil, fmt.Errorf("failed to get tile: %w", err)
	}

	// Create tile occupant based on tile type
	var occupant TileOccupant
	switch tileType {
	case "ocean":
		occupant = TileOccupant{Type: domain.ResourceOceanTile}
	case "city":
		occupant = TileOccupant{Type: domain.ResourceCityTile}
	case "greenery":
		occupant = TileOccupant{Type: domain.ResourceGreeneryTile}
	default:
		return nil, fmt.Errorf("unknown tile type: %s", tileType)
	}

	// Place the tile on the board (updates board state and publishes TilePlacedEvent)
	if err := s.boardService.PlaceTile(ctx, coordinate, occupant, ownerID); err != nil {
		return nil, fmt.Errorf("failed to place tile: %w", err)
	}

	log.Info("🎯 Tile placed successfully",
		zap.String("coordinate", coordinate.String()),
		zap.String("tile_type", tileType))

	// Calculate bonuses from this tile placement
	bonuses := s.calculateTilePlacementBonuses(tile)

	result := &PlacementResult{
		Coordinate: coordinate,
		TileType:   tileType,
		Bonuses:    bonuses,
	}

	log.Debug("✅ Tile placement result",
		zap.Int("bonus_count", len(bonuses)))

	return result, nil
}

// calculateTilePlacementBonuses determines what bonuses the player gets from placing on this tile
// These are board-based bonuses, not card passive effects (those are handled via events)
func (s *SelectionServiceImpl) calculateTilePlacementBonuses(tile *Tile) []TilePlacementBonus {
	if tile == nil {
		return []TilePlacementBonus{}
	}

	tileBonuses := tile.Bonuses
	if len(tileBonuses) == 0 {
		return []TilePlacementBonus{}
	}

	bonuses := make([]TilePlacementBonus, 0, len(tileBonuses))

	// Convert tile bonuses to placement bonuses
	for _, tileBonus := range tileBonuses {
		bonuses = append(bonuses, TilePlacementBonus{
			Type:   tileBonus.Type,
			Amount: tileBonus.Amount,
		})
	}

	return bonuses
}
