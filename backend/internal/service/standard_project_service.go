package service

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// StandardProjectService handles standard project operations
type StandardProjectService interface {
	// SellPatents exchanges hand cards for megacredits (1 M€ per card)
	SellPatents(ctx context.Context, gameID, playerID string, cardCount int) error

	// BuildPowerPlant increases energy production for 11 M€
	BuildPowerPlant(ctx context.Context, gameID, playerID string) error

	// LaunchAsteroid raises temperature for 14 M€ and grants TR
	LaunchAsteroid(ctx context.Context, gameID, playerID string) error

	// BuildAquifer places ocean tile for 18 M€ and grants TR (creates tile queue)
	BuildAquifer(ctx context.Context, gameID, playerID string) error

	// PlantGreenery places greenery tile for 23 M€ and grants TR (creates tile queue)
	PlantGreenery(ctx context.Context, gameID, playerID string) error

	// BuildCity places city tile for 25 M€ (creates tile queue)
	BuildCity(ctx context.Context, gameID, playerID string) error
}

// StandardProjectServiceImpl implements StandardProjectService interface
type StandardProjectServiceImpl struct {
	gameRepo       repository.GameRepository
	playerRepo     repository.PlayerRepository
	sessionManager session.SessionManager
	tileService    TileService
}

// NewStandardProjectService creates a new StandardProjectService instance
func NewStandardProjectService(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	sessionManager session.SessionManager,
	tileService TileService,
) StandardProjectService {
	return &StandardProjectServiceImpl{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		sessionManager: sessionManager,
		tileService:    tileService,
	}
}

// SellPatents exchanges hand cards for megacredits (1 M€ per card)
func (s *StandardProjectServiceImpl) SellPatents(ctx context.Context, gameID, playerID string, cardCount int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get player
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for sell patents", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Validate player can sell cards
	if len(player.Cards) < cardCount || cardCount <= 0 {
		log.Warn("Player attempted to sell more cards than available",
			zap.Int("requested", cardCount),
			zap.Int("available", len(player.Cards)))
		return fmt.Errorf("player only has %d cards, cannot sell %d", len(player.Cards), cardCount)
	}

	// Calculate credits gained (1 M€ per card)
	creditsGained := cardCount

	// Update player resources
	updatedPlayer := player
	updatedPlayer.Resources.Credits += creditsGained

	// Update player resources first
	if err := s.playerRepo.UpdateResources(ctx, gameID, updatedPlayer.ID, updatedPlayer.Resources); err != nil {
		log.Error("Failed to update player resources after selling patents", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	// Remove cards from hand (remove first N cards)
	for i := 0; i < cardCount && i < len(player.Cards); i++ {
		cardToRemove := player.Cards[i]
		if err := s.playerRepo.RemoveCard(ctx, gameID, playerID, cardToRemove); err != nil {
			log.Error("Failed to remove card after selling patents",
				zap.String("card_id", cardToRemove),
				zap.Error(err))
			return fmt.Errorf("failed to remove card %s: %w", cardToRemove, err)
		}
	}

	// Clean architecture: no manual game state sync needed

	log.Info("Player sold patents",
		zap.Int("cards_sold", cardCount),
		zap.Int("credits_gained", creditsGained))

	return nil
}

// BuildPowerPlant increases energy production for 11 M€
func (s *StandardProjectServiceImpl) BuildPowerPlant(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	return s.executeStandardProject(ctx, gameID, playerID, model.StandardProjectPowerPlant, func(player *model.Player) error {
		// Increase energy production by 1
		player.Production.Energy++

		log.Info("Player built power plant",
			zap.Int("new_energy_production", player.Production.Energy))

		return nil
	})
}

// LaunchAsteroid raises temperature for 14 M€ and grants TR
func (s *StandardProjectServiceImpl) LaunchAsteroid(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	return s.executeStandardProject(ctx, gameID, playerID, model.StandardProjectAsteroid, func(player *model.Player) error {
		// Increase terraform rating
		player.TerraformRating++

		// Increase temperature by 1 step (2°C) - asteroid effect
		game, err := s.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			log.Error("Failed to get game", zap.Error(err))
			return fmt.Errorf("failed to get game: %w", err)
		}

		newParams := game.GlobalParameters
		newParams.Temperature += 2 // Each step is 2°C
		if newParams.Temperature > model.MaxTemperature {
			newParams.Temperature = model.MaxTemperature
		}

		if err := s.gameRepo.UpdateGlobalParameters(ctx, gameID, newParams); err != nil {
			log.Error("Failed to update temperature", zap.Error(err))
			return fmt.Errorf("failed to update temperature: %w", err)
		}

		log.Info("Player launched asteroid",
			zap.Int("new_terraform_rating", player.TerraformRating))

		return nil
	})
}

// BuildAquifer places ocean tile for 18 M€ and grants TR
func (s *StandardProjectServiceImpl) BuildAquifer(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	return s.executeTileQueuedProject(ctx, gameID, playerID, model.StandardProjectAquifer, "ocean", func(player *model.Player) error {
		// Increase terraform rating
		player.TerraformRating++

		// Increase ocean count by 1
		game, err := s.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			log.Error("Failed to get game", zap.Error(err))
			return fmt.Errorf("failed to get game: %w", err)
		}

		newParams := game.GlobalParameters
		newParams.Oceans++
		if newParams.Oceans > model.MaxOceans {
			newParams.Oceans = model.MaxOceans
		}

		if err := s.gameRepo.UpdateGlobalParameters(ctx, gameID, newParams); err != nil {
			log.Error("Failed to place ocean", zap.Error(err))
			return fmt.Errorf("failed to place ocean: %w", err)
		}

		log.Info("Player built aquifer (queuing tile placement)", zap.Int("new_terraform_rating", player.TerraformRating))
		return nil
	})
}

// PlantGreenery places greenery tile for 23 M€ and grants TR
func (s *StandardProjectServiceImpl) PlantGreenery(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	return s.executeTileQueuedProject(ctx, gameID, playerID, model.StandardProjectGreenery, "greenery", func(player *model.Player) error {
		// Increase terraform rating
		player.TerraformRating++

		// Increase oxygen by 1% - greenery effect
		game, err := s.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			log.Error("Failed to get game", zap.Error(err))
			return fmt.Errorf("failed to get game: %w", err)
		}

		newParams := game.GlobalParameters
		newParams.Oxygen++
		if newParams.Oxygen > model.MaxOxygen {
			newParams.Oxygen = model.MaxOxygen
		}

		if err := s.gameRepo.UpdateGlobalParameters(ctx, gameID, newParams); err != nil {
			log.Error("Failed to increase oxygen", zap.Error(err))
			return fmt.Errorf("failed to increase oxygen: %w", err)
		}

		log.Info("Player planted greenery (queuing tile placement)", zap.Int("new_terraform_rating", player.TerraformRating))
		return nil
	})
}

// BuildCity places city tile for 25 M€
func (s *StandardProjectServiceImpl) BuildCity(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	return s.executeTileQueuedProject(ctx, gameID, playerID, model.StandardProjectCity, "city", func(player *model.Player) error {
		// Increase megacredit production by 1 (cities provide income)
		player.Production.Credits++
		log.Info("Player built city (queuing tile placement)", zap.Int("new_credit_production", player.Production.Credits))
		return nil
	})
}

// executeTileQueuedProject executes a standard project that requires tile placement
func (s *StandardProjectServiceImpl) executeTileQueuedProject(ctx context.Context, gameID, playerID string, project model.StandardProject, tileType string, projectAction func(*model.Player) error) error {
	log := logger.WithGameContext(gameID, playerID)

	// Execute standard project (cost deduction, effects)
	if err := s.executeStandardProject(ctx, gameID, playerID, project, projectAction); err != nil {
		return err
	}

	// Create tile queue
	queueSource := fmt.Sprintf("standard-project-%s", tileType)
	if err := s.playerRepo.CreateTileQueue(ctx, gameID, playerID, queueSource, []string{tileType}); err != nil {
		log.Error("Failed to create tile queue", zap.Error(err))
		return fmt.Errorf("failed to create tile queue: %w", err)
	}

	// Process tile queue to set pendingTileSelection with available hexes
	if err := s.tileService.ProcessTileQueue(ctx, gameID, playerID); err != nil {
		log.Error("Failed to process tile queue", zap.Error(err))
		return fmt.Errorf("failed to process tile queue: %w", err)
	}

	// Broadcast game state (includes pendingTileSelection)
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
	}

	log.Info("✅ Standard project executed, tile queued for placement", zap.String("tile_type", tileType))
	return nil
}

// executeStandardProject executes a standard project with common validation and resource deduction
func (s *StandardProjectServiceImpl) executeStandardProject(ctx context.Context, gameID, playerID string, project model.StandardProject, projectAction func(*model.Player) error) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get player
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for standard project", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Validate player can afford the project
	cost, exists := model.StandardProjectCost[project]
	if !exists {
		return fmt.Errorf("unknown standard project: %s", project)
	}
	if player.Resources.Credits < cost {
		log.Warn("Player cannot afford standard project",
			zap.String("project", string(project)),
			zap.Int("cost", cost),
			zap.Int("player_credits", player.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", cost, player.Resources.Credits)
	}

	// Create updated player copy
	updatedPlayer := player

	// Deduct cost
	updatedPlayer.Resources.Credits -= cost

	// Execute project-specific action
	if err := projectAction(&updatedPlayer); err != nil {
		return err
	}

	// Update player resources
	if err := s.playerRepo.UpdateResources(ctx, gameID, updatedPlayer.ID, updatedPlayer.Resources); err != nil {
		log.Error("Failed to update player resources after standard project", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	// Update player production if it changed
	if err := s.playerRepo.UpdateProduction(ctx, gameID, updatedPlayer.ID, updatedPlayer.Production); err != nil {
		log.Error("Failed to update player production after standard project", zap.Error(err))
		return fmt.Errorf("failed to update player production: %w", err)
	}

	// Update terraform rating if it changed
	if updatedPlayer.TerraformRating != player.TerraformRating {
		if err := s.playerRepo.UpdateTerraformRating(ctx, gameID, updatedPlayer.ID, updatedPlayer.TerraformRating); err != nil {
			log.Error("Failed to update player terraform rating after standard project", zap.Error(err))
			return fmt.Errorf("failed to update player terraform rating: %w", err)
		}
	}

	// Broadcast updated game state to all players
	s.sessionManager.Broadcast(gameID)

	log.Info("Standard project executed",
		zap.String("project", string(project)),
		zap.Int("cost", cost))

	return nil
}
