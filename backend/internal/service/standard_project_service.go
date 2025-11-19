package service

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/session"

	"go.uber.org/zap"
)

// StandardProjectService handles standard project operations
type StandardProjectService interface {
	// InitiateSellPatents initiates the sell patents flow by creating a pending card selection
	InitiateSellPatents(ctx context.Context, gameID, playerID string) error

	// ProcessCardSelection processes a card selection (sell patents, card effects, etc.)
	ProcessCardSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error

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

// InitiateSellPatents initiates the sell patents flow by creating a pending card selection
func (s *StandardProjectServiceImpl) InitiateSellPatents(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get player
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for sell patents", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Validate player has cards to sell
	if len(player.Cards) == 0 {
		log.Warn("Player attempted to sell patents with no cards in hand")
		return fmt.Errorf("player has no cards to sell")
	}

	// Create pending card selection with all player's cards
	// Each card costs 0 MC to "select" (sell) and rewards 1 MC
	cardCosts := make(map[string]int)
	cardRewards := make(map[string]int)
	for _, cardID := range player.Cards {
		cardCosts[cardID] = 0   // Free to select (selling)
		cardRewards[cardID] = 1 // Gain 1 MC per card sold
	}

	selection := &model.PendingCardSelection{
		AvailableCards: player.Cards, // All cards in hand available
		CardCosts:      cardCosts,
		CardRewards:    cardRewards,
		Source:         "sell-patents",
		MinCards:       0,                 // Can sell 0 cards (cancel action)
		MaxCards:       len(player.Cards), // Can sell all cards
	}

	// Store the pending card selection
	if err := s.playerRepo.UpdatePendingCardSelection(ctx, gameID, playerID, selection); err != nil {
		log.Error("Failed to create pending card selection", zap.Error(err))
		return fmt.Errorf("failed to create pending card selection: %w", err)
	}

	// Broadcast updated game state (includes pendingCardSelection)
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
	}

	log.Info("🃏 Sell patents initiated, awaiting card selection",
		zap.Int("available_cards", len(player.Cards)))

	return nil
}

// ProcessCardSelection processes a card selection (sell patents, card effects, etc.)
func (s *StandardProjectServiceImpl) ProcessCardSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get pending card selection
	selection, err := s.playerRepo.GetPendingCardSelection(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get pending card selection", zap.Error(err))
		return fmt.Errorf("failed to get pending card selection: %w", err)
	}

	if selection == nil {
		log.Warn("No pending card selection found")
		return fmt.Errorf("no pending card selection")
	}

	// Validate card count is within bounds
	if len(cardIDs) < selection.MinCards || len(cardIDs) > selection.MaxCards {
		log.Warn("Invalid card selection count",
			zap.Int("selected", len(cardIDs)),
			zap.Int("min", selection.MinCards),
			zap.Int("max", selection.MaxCards))
		return fmt.Errorf("must select between %d and %d cards, got %d", selection.MinCards, selection.MaxCards, len(cardIDs))
	}

	// Validate all selected cards are in the available list
	availableSet := make(map[string]bool)
	for _, cardID := range selection.AvailableCards {
		availableSet[cardID] = true
	}
	for _, cardID := range cardIDs {
		if !availableSet[cardID] {
			log.Warn("Player attempted to select unavailable card", zap.String("card_id", cardID))
			return fmt.Errorf("card %s is not available for selection", cardID)
		}
	}

	// Calculate total cost and reward
	totalCost := 0
	totalReward := 0
	for _, cardID := range cardIDs {
		totalCost += selection.CardCosts[cardID]
		totalReward += selection.CardRewards[cardID]
	}

	// Get player
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Validate player can afford the cost
	if player.Resources.Credits < totalCost {
		log.Warn("Player cannot afford card selection",
			zap.Int("cost", totalCost),
			zap.Int("player_credits", player.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", totalCost, player.Resources.Credits)
	}

	// Apply cost and reward
	player.Resources.Credits -= totalCost
	player.Resources.Credits += totalReward

	// Update player resources
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, player.Resources); err != nil {
		log.Error("Failed to update player resources", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	// Remove selected cards from hand
	for _, cardID := range cardIDs {
		if err := s.playerRepo.RemoveCard(ctx, gameID, playerID, cardID); err != nil {
			log.Error("Failed to remove card",
				zap.String("card_id", cardID),
				zap.Error(err))
			return fmt.Errorf("failed to remove card %s: %w", cardID, err)
		}
	}

	// Clear pending card selection
	if err := s.playerRepo.ClearPendingCardSelection(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear pending card selection", zap.Error(err))
		return fmt.Errorf("failed to clear pending card selection: %w", err)
	}

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
	}

	log.Info("✅ Card selection processed",
		zap.String("source", selection.Source),
		zap.Int("cards_selected", len(cardIDs)),
		zap.Int("total_cost", totalCost),
		zap.Int("total_reward", totalReward))

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

	// Tile queue processing (now automatic via TileQueueCreatedEvent)
	// No manual call needed - TileProcessor subscribes to events and processes automatically

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
