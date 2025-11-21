package tile

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/session/board"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
)

// Processor handles tile queue processing and tile placement logic
type Processor struct {
	gameRepo       game.Repository
	playerRepo     player.Repository
	boardRepo      board.Repository
	boardProcessor *board.BoardProcessor
	eventBus       *events.EventBusImpl
}

// NewProcessor creates a new TileProcessor instance
func NewProcessor(
	gameRepo game.Repository,
	playerRepo player.Repository,
	boardRepo board.Repository,
	boardProcessor *board.BoardProcessor,
	eventBus *events.EventBusImpl,
) *Processor {
	return &Processor{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		boardRepo:      boardRepo,
		boardProcessor: boardProcessor,
		eventBus:       eventBus,
	}
}

// SubscribeToEvents sets up event subscriptions for automatic tile queue processing
func (p *Processor) SubscribeToEvents() {
	log := logger.Get()

	// Subscribe to TileQueueCreatedEvent for automatic processing
	events.Subscribe(p.eventBus, func(event player.TileQueueCreatedEvent) {
		log.Info("🎯 TileQueueCreatedEvent received, processing queue automatically",
			zap.String("game_id", event.GameID),
			zap.String("player_id", event.PlayerID),
			zap.Int("queue_size", event.QueueSize),
			zap.String("source", event.Source))

		// Process tile queue immediately and synchronously
		ctx := context.Background()
		err := p.ProcessTileQueue(ctx, event.GameID, event.PlayerID)
		if err != nil {
			log.Warn("⚠️ Failed to process tile queue from event",
				zap.String("game_id", event.GameID),
				zap.String("player_id", event.PlayerID),
				zap.Error(err))
			// Non-fatal - don't crash if tile processing fails
		}
	})

	log.Info("✅ TileProcessor event subscriptions initialized")
}

// ProcessTileQueue processes the tile queue, validating and setting up the first valid tile selection
// This should be called after any operation that creates a tile queue (e.g., card play, standard project)
// Returns nil if queue is empty or doesn't exist
func (p *Processor) ProcessTileQueue(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🎯 Processing tile queue")

	// Process the queue through the private validation method
	return p.processNextTileInQueueWithValidation(ctx, gameID, playerID)
}

// processNextTileInQueueWithValidation processes the next tile in queue with business logic validation
func (p *Processor) processNextTileInQueueWithValidation(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get the queue to extract the source (card ID or project ID)
	queue, err := p.playerRepo.GetPendingTileSelectionQueue(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get tile queue: %w", err)
	}

	// If no queue, we're done
	if queue == nil {
		log.Debug("No tile queue exists")
		return nil
	}

	source := queue.Source // Store the source (card/project ID)

	for {
		// Pop the next tile type from repository (pure data operation)
		nextTileType, err := p.playerRepo.ProcessNextTileInQueue(ctx, gameID, playerID)
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
		canPlace, err := p.validateTilePlacement(ctx, gameID, nextTileType)
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

			availableHexes, err := p.calculateAvailableHexesForTileType(ctx, gameID, playerIDPtr, nextTileType)
			if err != nil {
				return fmt.Errorf("failed to calculate available hexes: %w", err)
			}

			// Create and set the pending tile selection with available hexes
			selection := &model.PendingTileSelection{
				TileType:       nextTileType,
				AvailableHexes: availableHexes,
				Source:         source,
			}

			if err := p.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, selection); err != nil {
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
func (p *Processor) validateTilePlacement(ctx context.Context, gameID, tileType string) (bool, error) {
	log := logger.WithGameContext(gameID, "")

	// Get board state to count ocean tiles
	b, err := p.boardRepo.GetByGameID(ctx, gameID)
	if err != nil {
		return false, fmt.Errorf("failed to get board state: %w", err)
	}

	switch tileType {
	case "ocean":
		// Check if we've reached the maximum of 9 ocean tiles
		// Count existing ocean tiles on the board
		oceanCount := 0
		for _, tile := range b.Tiles {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type == board.ResourceOceanTile {
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
func (p *Processor) calculateAvailableHexesForTileType(ctx context.Context, gameID string, playerID *string, tileType string) ([]string, error) {
	log := logger.WithGameContext(gameID, "")
	if playerID != nil {
		log = logger.WithGameContext(gameID, *playerID)
	}

	log.Info("🏙️ Starting available hexes calculation",
		zap.String("tile_type", tileType),
		zap.Bool("player_specific", playerID != nil))

	// Get the current board state
	b, err := p.boardRepo.GetByGameID(ctx, gameID)
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
