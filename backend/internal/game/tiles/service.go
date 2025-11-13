package tiles

import (
	"context"
	"fmt"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// BoardService interface for hex calculations (avoids import cycle with service package)
type BoardService interface {
	GenerateDefaultBoard() Board
	CalculateAvailableHexesForTileType(game Game, tileType string) ([]string, error)
	CalculateAvailableHexesForTileTypeWithPlayer(game Game, tileType, playerID string) ([]string, error)
}

// Service handles tile placement operations on the Mars board.
//
// Scope: Isolated tile mechanic
//   - Tile placement at coordinates
//   - Validation of tile placement legality
//   - Tile bonus calculation and awarding (steel, titanium, plants, card draw)
//   - Ocean adjacency bonus calculation (+2 MC per ocean)
//   - Tile placement queue processing
//
// This mechanic is ISOLATED and should NOT:
//   - Call other mechanic services
//   - Manage turn state or phases
//   - Handle global parameter updates (that's parameters mechanic)
//
// Dependencies:
//   - GameRepository (for reading board state and updating tile occupancy)
//   - PlayerRepository (for updating player resources from bonuses)
//   - BoardService (for hex calculations - stateless utility)
//   - EventBus (for publishing tile placement events)
type Service interface {
	// Tile placement operations
	PlaceTile(ctx context.Context, gameID, playerID, tileType string, coordinate HexPosition) error
	ValidatePlacement(ctx context.Context, gameID, tileType string) error
	CalculateAvailableHexes(ctx context.Context, gameID, playerID, tileType string) ([]string, error)

	// Bonus operations
	AwardPlacementBonuses(ctx context.Context, gameID, playerID string, coordinate HexPosition) error

	// Queue operations
	ProcessTileQueue(ctx context.Context, gameID, playerID string) error
}

// ServiceImpl implements the Tiles mechanic service
type ServiceImpl struct {
	repo         Repository
	boardService BoardService
	eventBus     *events.EventBusImpl
}

// NewService creates a new Tiles mechanic service
func NewService(repo Repository, boardService BoardService, eventBus *events.EventBusImpl) Service {
	return &ServiceImpl{
		repo:         repo,
		boardService: boardService,
		eventBus:     eventBus,
	}
}

// PlaceTile places a tile of the specified type at the given coordinate.
// Awards placement bonuses and updates board occupancy.
// Does NOT raise global parameters - that's the parameters mechanic's job.
func (s *ServiceImpl) PlaceTile(ctx context.Context, gameID, playerID, tileType string, coordinate HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🔧 Placing tile",
		zap.String("tile_type", tileType),
		zap.Int("q", coordinate.Q),
		zap.Int("r", coordinate.R),
		zap.Int("s", coordinate.S))

	// Convert tile type string to ResourceType
	resourceType, err := TileTypeToResourceType(tileType)
	if err != nil {
		log.Error("Unknown tile type", zap.String("tile_type", tileType))
		return err
	}

	// Create the tile occupant
	occupant := &TileOccupant{
		Type: resourceType,
		Tags: []string{},
	}

	// Award placement bonuses to the player
	if err := s.AwardPlacementBonuses(ctx, gameID, playerID, coordinate); err != nil {
		log.Error("Failed to award tile placement bonuses", zap.Error(err))
		return fmt.Errorf("failed to award tile placement bonuses: %w", err)
	}

	// Update the tile occupancy in the game board
	if err := s.repo.UpdateTileOccupancy(ctx, gameID, coordinate, occupant, &playerID); err != nil {
		log.Error("Failed to update tile occupancy", zap.Error(err))
		return fmt.Errorf("failed to update tile occupancy: %w", err)
	}

	// Passive effects are triggered automatically via TilePlacedEvent from the repository
	// No manual triggering needed here - event system handles it

	log.Info("✅ Tile placed successfully",
		zap.String("tile_type", tileType),
		zap.String("coordinate", coordinate.String()))

	return nil
}

// ValidatePlacement validates if a tile of the specified type can be placed.
// Checks ocean count limits for ocean tiles.
func (s *ServiceImpl) ValidatePlacement(ctx context.Context, gameID, tileType string) error {
	log := logger.WithGameContext(gameID, "")

	// Only ocean tiles have placement validation (count limit)
	if tileType == "ocean" {
		game, err := s.repo.GetGame(ctx, gameID)
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		// Count oceans on board (source of truth)
		oceansPlaced := 0
		for _, tile := range game.Board.Tiles {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type == ResourceOceanTile {
				oceansPlaced++
			}
		}

		if oceansPlaced >= MaxOceans {
			log.Warn("Cannot place ocean tile - maximum oceans reached",
				zap.Int("current_oceans", oceansPlaced),
				zap.Int("max_oceans", MaxOceans))
			return fmt.Errorf("maximum oceans already placed")
		}
	}

	return nil
}

// CalculateAvailableHexes calculates available hexes for tile placement.
// Uses BoardService for hex calculation logic.
func (s *ServiceImpl) CalculateAvailableHexes(ctx context.Context, gameID, playerID, tileType string) ([]string, error) {
	game, err := s.repo.GetGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Use BoardService for hex calculations
	if tileType == "greenery" && playerID != "" {
		// Greenery requires player context for adjacency rules
		return s.boardService.CalculateAvailableHexesForTileTypeWithPlayer(game, tileType, playerID)
	}

	return s.boardService.CalculateAvailableHexesForTileType(game, tileType)
}

// AwardPlacementBonuses awards bonuses to the player for placing a tile.
// Bonuses include: steel, titanium, plants, card draw (from board), ocean adjacency (from placement)
func (s *ServiceImpl) AwardPlacementBonuses(ctx context.Context, gameID, playerID string, coordinate HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get the game state to access the board
	game, err := s.repo.GetGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Find the placed tile in the board
	var placedTile *Tile
	for i, tile := range game.Board.Tiles {
		if tile.Coordinates.Equals(coordinate) {
			placedTile = &game.Board.Tiles[i]
			break
		}
	}

	if placedTile == nil {
		log.Warn("Placed tile not found in board", zap.Any("coordinate", coordinate))
		return nil // Not critical, just skip bonuses
	}

	// Get current player resources
	player, err := s.repo.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	newResources := player.Resources
	var totalCreditsBonus int
	var bonusesAwarded []string

	// Collect all placement bonuses for event publishing
	placementBonuses := make(map[string]int)

	// Award tile bonuses (steel, titanium, plants, card draw)
	for _, bonus := range placedTile.Bonuses {
		description, applied, resourceType, amount := s.applyTileBonus(ctx, gameID, playerID, coordinate, &newResources, bonus, log)
		if applied {
			bonusesAwarded = append(bonusesAwarded, description)
			// Collect bonus for event
			if resourceType != "" {
				placementBonuses[resourceType] += amount
			}
		}
	}

	// Award ocean adjacency bonus (+2 MC per adjacent ocean, or +3 with Lakefront Resorts)
	oceanAdjacencyBonus := s.calculateOceanAdjacencyBonus(game, player, coordinate)
	if oceanAdjacencyBonus > 0 {
		totalCreditsBonus += oceanAdjacencyBonus
		bonusesAwarded = append(bonusesAwarded, fmt.Sprintf("+%d MC (ocean adjacency)", oceanAdjacencyBonus))
		log.Info("🌊 Ocean adjacency bonus awarded",
			zap.Int("bonus", oceanAdjacencyBonus))
	}

	// Apply credits bonus if any
	if totalCreditsBonus > 0 {
		newResources.Credits += totalCreditsBonus
	}

	// Update player resources if any bonuses were awarded
	if len(bonusesAwarded) > 0 {
		if err := s.repo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
			return fmt.Errorf("failed to update player resources: %w", err)
		}

		log.Info("✅ All tile placement bonuses awarded",
			zap.Strings("bonuses", bonusesAwarded))
	}

	// Publish PlacementBonusGainedEvent with all resources if any bonuses were gained
	if len(placementBonuses) > 0 && s.eventBus != nil {
		events.Publish(s.eventBus, repository.PlacementBonusGainedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			Resources: placementBonuses,
			Q:         coordinate.Q,
			R:         coordinate.R,
			S:         coordinate.S,
			Timestamp: time.Now(),
		})

		log.Debug("📢 Published PlacementBonusGainedEvent",
			zap.Any("resources", placementBonuses))
	}

	return nil
}

// calculateOceanAdjacencyBonus calculates the bonus from adjacent ocean tiles.
// Base bonus is 2 MC per ocean, but can be modified by effects (e.g., Lakefront Resorts adds +1).
func (s *ServiceImpl) calculateOceanAdjacencyBonus(game Game, player Player, coordinate HexPosition) int {
	adjacentOceanCount := 0

	// Check each adjacent position for ocean tiles
	for _, adjacentCoord := range coordinate.GetNeighbors() {
		// Find the adjacent tile in the board
		for _, tile := range game.Board.Tiles {
			if tile.Coordinates.Equals(adjacentCoord) {
				// Check if this tile has an ocean
				if tile.OccupiedBy != nil && tile.OccupiedBy.Type == ResourceOceanTile {
					adjacentOceanCount++
				}
				break
			}
		}
	}

	// Base ocean adjacency bonus is 2 MC per ocean
	baseBonus := 2

	// TODO: Support ocean adjacency bonus modifiers from cards (e.g., Lakefront Resorts)
	// This will require checking card effects via CardEffectSubscriber or similar mechanism

	// Calculate total bonus: adjacent oceans * base bonus
	return adjacentOceanCount * baseBonus
}

// applyTileBonus applies a single tile bonus to player resources.
// Returns: description, applied (bool), resourceType, amount
func (s *ServiceImpl) applyTileBonus(ctx context.Context, gameID, playerID string, coordinate HexPosition, resources *Resources, bonus TileBonus, log *zap.Logger) (string, bool, string, int) {
	var resourceName string

	switch bonus.Type {
	case ResourceSteel:
		resources.Steel += bonus.Amount
		resourceName = "steel"

	case ResourceTitanium:
		resources.Titanium += bonus.Amount
		resourceName = "titanium"

	case ResourcePlants:
		resources.Plants += bonus.Amount
		resourceName = "plants"

	case ResourceCardDraw:
		// TODO: Implement card drawing bonus
		log.Info("🎁 Tile bonus awarded (card draw not yet implemented)",
			zap.String("type", "card-draw"),
			zap.Int("amount", bonus.Amount))
		return fmt.Sprintf("+%d cards", bonus.Amount), true, string(ResourceCardDraw), bonus.Amount

	default:
		// Unknown bonus type, skip it
		return "", false, "", 0
	}

	log.Info("🎁 Tile bonus awarded",
		zap.String("type", resourceName),
		zap.Int("amount", bonus.Amount))

	return fmt.Sprintf("+%d %s", bonus.Amount, resourceName), true, string(bonus.Type), bonus.Amount
}

// ProcessTileQueue processes the next tile in the player's tile placement queue.
// Validates tile placement and sets up the pending tile selection.
func (s *ServiceImpl) ProcessTileQueue(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get the next tile type from the queue (pops from queue)
	nextTileType, err := s.repo.ProcessNextTileInQueue(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to process next tile in queue: %w", err)
	}

	// If no tiles in queue, nothing to do
	if nextTileType == "" {
		log.Debug("No tiles in queue to process")
		return nil
	}

	// Get the queue to determine the source
	queue, err := s.repo.GetPendingTileSelectionQueue(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get tile queue: %w", err)
	}

	source := ""
	if queue != nil {
		source = queue.Source
	}

	log.Info("🔄 Processing next tile in queue",
		zap.String("tile_type", nextTileType),
		zap.String("source", source))

	// Validate that this tile type can be placed (e.g., ocean count check)
	if err := s.ValidatePlacement(ctx, gameID, nextTileType); err != nil {
		log.Warn("Tile placement validation failed, skipping tile",
			zap.String("tile_type", nextTileType),
			zap.Error(err))

		// Recursively process next tile
		return s.ProcessTileQueue(ctx, gameID, playerID)
	}

	// Calculate available hexes for this tile type
	availableHexes, err := s.CalculateAvailableHexes(ctx, gameID, playerID, nextTileType)
	if err != nil {
		return fmt.Errorf("failed to calculate available hexes: %w", err)
	}

	if len(availableHexes) == 0 {
		log.Warn("No available hexes for tile placement, skipping",
			zap.String("tile_type", nextTileType))

		// Recursively process next tile
		return s.ProcessTileQueue(ctx, gameID, playerID)
	}

	// Set pending tile selection with available hexes
	pendingSelection := PendingTileSelection{
		TileType:       nextTileType,
		Source:         source,
		AvailableHexes: availableHexes,
	}

	if err := s.repo.UpdatePendingTileSelection(ctx, gameID, playerID, &pendingSelection); err != nil {
		return fmt.Errorf("failed to set pending tile selection: %w", err)
	}

	log.Info("✅ Tile queue processed, pending selection set",
		zap.String("tile_type", nextTileType),
		zap.Int("available_hexes", len(availableHexes)))

	return nil
}
