package board

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game/player"
)

// GameInterface defines the methods needed from Game for tile processing
// This avoids circular dependency between board and game packages
type GameInterface interface {
	GetPendingTileSelectionQueue(playerID string) *player.PendingTileSelectionQueue
	ProcessNextTile(ctx context.Context, playerID string) (string, error)
	SetPendingTileSelection(ctx context.Context, playerID string, selection *player.PendingTileSelection) error
}

// Processor handles tile queue processing and tile placement logic
type Processor struct {
	boardRepo      Repository
	boardProcessor *BoardProcessor
}

// NewProcessor creates a new TileProcessor instance
func NewProcessor(
	boardRepo Repository,
	boardProcessor *BoardProcessor,
) *Processor {
	return &Processor{
		boardRepo:      boardRepo,
		boardProcessor: boardProcessor,
	}
}

// ProcessTileQueue processes the tile queue, validating and setting up the first valid tile selection
// This should be called after any operation that creates a tile queue (e.g., card play, standard project)
// Returns nil if queue is empty or doesn't exist
// Requires Game to manage phase state (tile queue, pending selections)
func (p *Processor) ProcessTileQueue(ctx context.Context, game GameInterface, plr *player.Player) error {
	log := logger.Get().With(zap.String("player_id", plr.ID()))
	log.Debug("🎯 Processing tile queue")

	// Process the queue through the private validation method
	return p.processNextTileInQueueWithValidation(ctx, game, plr)
}

// processNextTileInQueueWithValidation processes the next tile in queue with business logic validation
func (p *Processor) processNextTileInQueueWithValidation(ctx context.Context, game GameInterface, plr *player.Player) error {
	log := logger.Get().With(zap.String("player_id", plr.ID()))
	playerID := plr.ID()

	// Get the queue to extract the source (card ID or project ID)
	queue := game.GetPendingTileSelectionQueue(playerID)

	// If no queue, we're done
	if queue == nil {
		log.Debug("No tile queue exists")
		return nil
	}

	source := queue.Source // Store the source (card/project ID)

	for {
		// Pop the next tile type from game (pure data operation)
		nextTileType, err := game.ProcessNextTile(ctx, playerID)
		if err != nil {
			return fmt.Errorf("failed to pop next tile from queue: %w", err)
		}

		// If no tile type returned, we're done
		if nextTileType == "" {
			log.Debug("No more tiles in queue")
			return nil
		}

		log.Info("🎯 Validating next tile from queue",
			zap.String("tile_type", nextTileType),
			zap.String("source", source))

		// Validate this tile placement is still possible (especially important for oceans)
		canPlace, err := p.validateTilePlacement(ctx, nextTileType)
		if err != nil {
			return fmt.Errorf("failed to validate tile placement: %w", err)
		}

		log.Info("🏙️ Tile placement validation result",
			zap.String("tile_type", nextTileType),
			zap.Bool("can_place", canPlace))

		if canPlace {
			// Tile is valid, now calculate available hexes for this tile type
			// For greenery, pass playerID to enforce adjacency rule
			var playerIDPtr *string
			if nextTileType == "greenery" {
				playerIDPtr = &playerID
			}

			availableHexes, err := p.calculateAvailableHexesForTileType(ctx, playerIDPtr, nextTileType)
			if err != nil {
				return fmt.Errorf("failed to calculate available hexes: %w", err)
			}

			// Create and set the pending tile selection with available hexes
			selection := &player.PendingTileSelection{
				TileType:       nextTileType,
				AvailableHexes: availableHexes,
				Source:         source,
			}

			if err := game.SetPendingTileSelection(ctx, playerID, selection); err != nil {
				return fmt.Errorf("failed to set pending tile selection: %w", err)
			}

			log.Info("🎯 Tile validation successful and available hexes calculated",
				zap.String("tile_type", nextTileType),
				zap.Int("available_hexes", len(availableHexes)))
			return nil
		}

		// Tile is no longer valid, skip it and try next
		log.Info("⚠️ Tile placement no longer possible, skipping and checking next",
			zap.String("tile_type", nextTileType))

		// Continue loop to pop and process next tile
	}
}

// validateTilePlacement checks if a tile type can still be placed in the game
func (p *Processor) validateTilePlacement(ctx context.Context, tileType string) (bool, error) {
	log := logger.Get()

	// Get board state to count ocean tiles
	b, err := p.boardRepo.Get(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get board state: %w", err)
	}

	switch tileType {
	case "ocean":
		// Check if we've reached the maximum of 9 ocean tiles
		// Count existing ocean tiles on the board
		oceanCount := 0
		for _, tile := range b.Tiles {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type == ResourceOceanTile {
				oceanCount++
			}
		}

		canPlace := oceanCount < 9
		log.Debug("Ocean tile validation",
			zap.String("tile_type", tileType),
			zap.Int("current_oceans", oceanCount),
			zap.Bool("can_place", canPlace))
		return canPlace, nil

	case "city", "greenery":
		// Cities and greenery can generally always be placed if there are available spaces
		// More complex validation (adjacency rules, etc.) is handled by hex calculation
		return true, nil

	default:
		// Unknown tile types are considered valid for now
		log.Warn("Unknown tile type for validation", zap.String("tile_type", tileType))
		return true, nil
	}
}

// calculateAvailableHexesForTileType returns available hexes with optional player context
// Used for greenery placement which requires adjacency to player's tiles
// Note: gameID is not needed as processor is scoped to a specific game instance
func (p *Processor) calculateAvailableHexesForTileType(ctx context.Context, playerID *string, tileType string) ([]string, error) {
	log := logger.Get()
	if playerID != nil {
		log = log.With(zap.String("player_id", *playerID))
	}

	log.Info("🏙️ Starting available hexes calculation",
		zap.String("tile_type", tileType),
		zap.Bool("player_specific", playerID != nil))

	// Get the current board state
	b, err := p.boardRepo.Get(ctx)
	if err != nil {
		log.Error("Failed to get board for hex calculation", zap.Error(err))
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	log.Info("🏙️ Got board state, delegating to BoardProcessor",
		zap.String("tile_type", tileType))

	// Delegate to BoardProcessor for hex calculation with optional player context
	availableHexes := p.boardProcessor.CalculateAvailableHexesForTileType(b, tileType, playerID)

	log.Info("🏙️ BoardProcessor calculation completed",
		zap.String("tile_type", tileType),
		zap.Int("available_count", len(availableHexes)))

	return availableHexes, nil
}
