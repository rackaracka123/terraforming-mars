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

	// BuildAquifer places ocean tile for 18 M€ and grants TR
	BuildAquifer(ctx context.Context, gameID, playerID string, hexPosition model.HexPosition) error

	// PlantGreenery places greenery tile for 23 M€ and grants TR
	PlantGreenery(ctx context.Context, gameID, playerID string, hexPosition model.HexPosition) error

	// BuildCity places city tile for 25 M€
	BuildCity(ctx context.Context, gameID, playerID string, hexPosition model.HexPosition) error

	// IsValidHexPosition validates hex coordinate positioning
	IsValidHexPosition(h *model.HexPosition) bool
}

// StandardProjectServiceImpl implements StandardProjectService interface
type StandardProjectServiceImpl struct {
	gameRepo       repository.GameRepository
	playerRepo     repository.PlayerRepository
	gameService    GameService
	sessionManager session.SessionManager
}

// NewStandardProjectService creates a new StandardProjectService instance
func NewStandardProjectService(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	gameService GameService,
	sessionManager session.SessionManager,
) StandardProjectService {
	return &StandardProjectServiceImpl{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		gameService:    gameService,
		sessionManager: sessionManager,
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

		// Increase temperature by 1 step (2°C)
		if err := s.gameService.IncreaseTemperature(ctx, gameID, 1); err != nil {
			log.Error("Failed to increase temperature", zap.Error(err))
			return fmt.Errorf("failed to increase temperature: %w", err)
		}

		log.Info("Player launched asteroid",
			zap.Int("new_terraform_rating", player.TerraformRating))

		return nil
	})
}

// BuildAquifer places ocean tile for 18 M€ and grants TR
func (s *StandardProjectServiceImpl) BuildAquifer(ctx context.Context, gameID, playerID string, hexPosition model.HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)

	// Validate hex position
	if !s.IsValidHexPosition(&hexPosition) {
		log.Warn("Invalid hex position for aquifer",
			zap.Int("q", hexPosition.Q),
			zap.Int("r", hexPosition.R),
			zap.Int("s", hexPosition.S))
		return fmt.Errorf("invalid hex position: coordinates must sum to 0")
	}

	return s.executeStandardProject(ctx, gameID, playerID, model.StandardProjectAquifer, func(player *model.Player) error {
		// Increase terraform rating
		player.TerraformRating++

		// Place ocean tile (increase ocean count)
		if err := s.gameService.PlaceOcean(ctx, gameID, 1); err != nil {
			log.Error("Failed to place ocean", zap.Error(err))
			return fmt.Errorf("failed to place ocean: %w", err)
		}

		log.Info("Player built aquifer",
			zap.Int("new_terraform_rating", player.TerraformRating),
			zap.Any("hex_position", hexPosition))

		return nil
	})
}

// PlantGreenery places greenery tile for 23 M€ and grants TR
func (s *StandardProjectServiceImpl) PlantGreenery(ctx context.Context, gameID, playerID string, hexPosition model.HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)

	// Validate hex position
	if !s.IsValidHexPosition(&hexPosition) {
		log.Warn("Invalid hex position for greenery",
			zap.Int("q", hexPosition.Q),
			zap.Int("r", hexPosition.R),
			zap.Int("s", hexPosition.S))
		return fmt.Errorf("invalid hex position: coordinates must sum to 0")
	}

	return s.executeStandardProject(ctx, gameID, playerID, model.StandardProjectGreenery, func(player *model.Player) error {
		// Increase terraform rating
		player.TerraformRating++

		// Increase oxygen by 1 step
		if err := s.gameService.IncreaseOxygen(ctx, gameID, 1); err != nil {
			log.Error("Failed to increase oxygen", zap.Error(err))
			return fmt.Errorf("failed to increase oxygen: %w", err)
		}

		log.Info("Player planted greenery",
			zap.Int("new_terraform_rating", player.TerraformRating),
			zap.Any("hex_position", hexPosition))

		return nil
	})
}

// BuildCity places city tile for 25 M€
func (s *StandardProjectServiceImpl) BuildCity(ctx context.Context, gameID, playerID string, hexPosition model.HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)

	// Validate hex position
	if !s.IsValidHexPosition(&hexPosition) {
		log.Warn("Invalid hex position for city",
			zap.Int("q", hexPosition.Q),
			zap.Int("r", hexPosition.R),
			zap.Int("s", hexPosition.S))
		return fmt.Errorf("invalid hex position: coordinates must sum to 0")
	}

	return s.executeStandardProject(ctx, gameID, playerID, model.StandardProjectCity, func(player *model.Player) error {
		// Increase megacredit production by 1 (cities provide income)
		player.Production.Credits++

		log.Info("Player built city",
			zap.Int("new_credit_production", player.Production.Credits),
			zap.Any("hex_position", hexPosition))

		return nil
	})
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

// StandardProjectRequiresHexPosition returns true if the standard project requires a hex position (business logic from StandardProject functions)
func (s *StandardProjectServiceImpl) StandardProjectRequiresHexPosition(project model.StandardProject) bool {
	switch project {
	case model.StandardProjectAquifer, model.StandardProjectGreenery, model.StandardProjectCity:
		return true
	default:
		return false
	}
}

// StandardProjectProvidesTR returns true if the standard project increases terraform rating (business logic from StandardProject functions)
func (s *StandardProjectServiceImpl) StandardProjectProvidesTR(project model.StandardProject) bool {
	switch project {
	case model.StandardProjectAsteroid, model.StandardProjectAquifer, model.StandardProjectGreenery:
		return true
	default:
		return false
	}
}

// IsValidHexPosition validates that the hex position follows cube coordinate rules (business logic from HexPosition model)
func (s *StandardProjectServiceImpl) IsValidHexPosition(h *model.HexPosition) bool {
	return h.Q+h.R+h.S == 0
}

// DistanceHexPosition calculates the distance between two hex positions (business logic from HexPosition model)
func (s *StandardProjectServiceImpl) DistanceHexPosition(h1, h2 *model.HexPosition) int {
	return (abs(h1.Q-h2.Q) + abs(h1.R-h2.R) + abs(h1.S-h2.S)) / 2
}

// GetHexNeighbors returns all adjacent hex positions (business logic from HexPosition model)
func (s *StandardProjectServiceImpl) GetHexNeighbors(h *model.HexPosition) []model.HexPosition {
	directions := []model.HexPosition{
		{Q: 1, R: -1, S: 0}, // East
		{Q: 1, R: 0, S: -1}, // Southeast
		{Q: 0, R: 1, S: -1}, // Southwest
		{Q: -1, R: 1, S: 0}, // West
		{Q: -1, R: 0, S: 1}, // Northwest
		{Q: 0, R: -1, S: 1}, // Northeast
	}

	neighbors := make([]model.HexPosition, 6)
	for i, dir := range directions {
		neighbors[i] = model.HexPosition{
			Q: h.Q + dir.Q,
			R: h.R + dir.R,
			S: h.S + dir.S,
		}
	}

	return neighbors
}

// abs returns the absolute value of an integer (helper function)
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
