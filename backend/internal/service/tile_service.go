package service

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// TileService handles tile queue processing and tile placement logic
type TileService interface {
	// ProcessTileQueue processes the tile queue, validating and setting up the first valid tile selection
	// This should be called after any operation that creates a tile queue (e.g., card play)
	// Returns nil if queue is empty or doesn't exist
	ProcessTileQueue(ctx context.Context, gameID, playerID string) error
}

// TileServiceImpl implements TileService interface
type TileServiceImpl struct {
	gameRepo     repository.GameRepository
	playerRepo   repository.PlayerRepository
	boardService BoardService
}

// NewTileService creates a new TileService instance
func NewTileService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, boardService BoardService) TileService {
	return &TileServiceImpl{
		gameRepo:     gameRepo,
		playerRepo:   playerRepo,
		boardService: boardService,
	}
}

// ProcessTileQueue processes the tile queue, validating and setting up the first valid tile selection
// This is the public API method that should be called after card effects that create tile queues
func (s *TileServiceImpl) ProcessTileQueue(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🎯 Processing tile queue")

	// Process the queue through the private validation method
	return s.processNextTileInQueueWithValidation(ctx, gameID, playerID)
}

// processNextTileInQueueWithValidation processes the next tile in queue with business logic validation
func (s *TileServiceImpl) processNextTileInQueueWithValidation(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get the queue to extract the source (card ID)
	queue, err := s.playerRepo.GetPendingTileSelectionQueue(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get tile queue: %w", err)
	}

	// If no queue, we're done
	if queue == nil {
		log.Debug("No tile queue exists")
		return nil
	}

	source := queue.Source // Store the source card ID

	for {
		// Pop the next tile type from repository (pure data operation)
		nextTileType, err := s.playerRepo.ProcessNextTileInQueue(ctx, gameID, playerID)
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
		canPlace, err := s.validateTilePlacement(ctx, gameID, nextTileType)
		if err != nil {
			return fmt.Errorf("failed to validate tile placement: %w", err)
		}

		log.Info("🏙️ Tile placement validation result",
			zap.String("tile_type", nextTileType),
			zap.Bool("can_place", canPlace))

		if canPlace {
			// Tile is valid, now calculate available hexes for this tile type
			// For greenery, use player-specific calculation to enforce adjacency rule
			availableHexes, err := s.calculateAvailableHexesForTileTypeWithPlayer(ctx, gameID, playerID, nextTileType)
			if err != nil {
				return fmt.Errorf("failed to calculate available hexes: %w", err)
			}

			// Create and set the pending tile selection with available hexes
			selection := &model.PendingTileSelection{
				TileType:       nextTileType,
				AvailableHexes: availableHexes,
				Source:         source,
			}

			if err := s.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, selection); err != nil {
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
func (s *TileServiceImpl) validateTilePlacement(ctx context.Context, gameID, tileType string) (bool, error) {
	log := logger.WithGameContext(gameID, "")

	// Get game state to check global parameters
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return false, fmt.Errorf("failed to get game state: %w", err)
	}

	switch tileType {
	case "ocean":
		// Check if we've reached the maximum of 9 ocean tiles
		// Count existing ocean tiles on the board
		oceanCount := 0
		for _, tile := range game.Board.Tiles {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type == model.ResourceOceanTile {
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
		// More complex validation (adjacency rules, etc.) should be handled elsewhere
		return true, nil

	default:
		// Unknown tile types are considered valid for now
		log.Warn("Unknown tile type for validation", zap.String("tile_type", tileType))
		return true, nil
	}
}

// calculateAvailableHexesForTileTypeWithPlayer returns available hexes with player context
// Used for greenery placement which requires adjacency to player's tiles
func (s *TileServiceImpl) calculateAvailableHexesForTileTypeWithPlayer(ctx context.Context, gameID, playerID, tileType string) ([]string, error) {
	log := logger.WithGameContext(gameID, playerID)

	log.Info("🏙️ Starting available hexes calculation with player context",
		zap.String("tile_type", tileType))

	// Get the current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for hex calculation", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	log.Info("🏙️ Got game state, delegating to BoardService",
		zap.String("tile_type", tileType))

	// Delegate to BoardService for hex calculation with player context
	availableHexes, err := s.boardService.CalculateAvailableHexesForTileTypeWithPlayer(game, tileType, playerID)
	if err != nil {
		log.Error("🏙️ BoardService calculation failed", zap.Error(err))
		return nil, err
	}

	log.Info("🏙️ BoardService calculation completed",
		zap.String("tile_type", tileType),
		zap.Int("available_count", len(availableHexes)))

	return availableHexes, nil
}
