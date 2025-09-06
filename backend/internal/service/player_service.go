package service

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// PlayerService handles player-specific operations
type PlayerService interface {
	// Update player resources
	UpdatePlayerResources(ctx context.Context, gameID, playerID string, newResources model.Resources) error

	// Update player production
	UpdatePlayerProduction(ctx context.Context, gameID, playerID string, newProduction model.Production) error

	// Get player information
	GetPlayer(ctx context.Context, gameID, playerID string) (*model.Player, error)

	// Validation methods for card system
	ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement model.ResourceSet) error
	ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error

	// Card effect methods
	AddProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error
	PayResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error
	AddResources(ctx context.Context, gameID, playerID string, resources model.ResourceSet) error
}

// PlayerServiceImpl implements PlayerService interface
type PlayerServiceImpl struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

// NewPlayerService creates a new PlayerService instance
func NewPlayerService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) PlayerService {
	return &PlayerServiceImpl{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
	}
}

// UpdatePlayerResources updates a player's resources
func (s *PlayerServiceImpl) UpdatePlayerResources(ctx context.Context, gameID, playerID string, newResources model.Resources) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player
	player, err := s.playerRepo.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Create a copy of the player to avoid modifying the stored one
	updatedPlayer := *player
	updatedPlayer.Resources = newResources

	// Update through PlayerRepository (this will publish events)
	if err := s.playerRepo.UpdatePlayer(ctx, gameID, &updatedPlayer); err != nil {
		log.Error("Failed to update player resources", zap.Error(err))
		return fmt.Errorf("failed to update player: %w", err)
	}

	// Also need to update the game state to keep the main Game entity in sync
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for player update", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Find and update player in game
	for i, p := range game.Players {
		if p.ID == playerID {
			game.Players[i] = updatedPlayer
			break
		}
	}

	// Update game state
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game after player resource change", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Player resources updated successfully")
	return nil
}

// UpdatePlayerProduction updates a player's production
func (s *PlayerServiceImpl) UpdatePlayerProduction(ctx context.Context, gameID, playerID string, newProduction model.Production) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player
	player, err := s.playerRepo.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Create a copy of the player to avoid modifying the stored one
	updatedPlayer := *player
	updatedPlayer.Production = newProduction

	// Update through PlayerRepository (this will publish events)
	if err := s.playerRepo.UpdatePlayer(ctx, gameID, &updatedPlayer); err != nil {
		log.Error("Failed to update player production", zap.Error(err))
		return fmt.Errorf("failed to update player: %w", err)
	}

	// Also need to update the game state to keep the main Game entity in sync
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for player update", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Find and update player in game
	for i, p := range game.Players {
		if p.ID == playerID {
			game.Players[i] = updatedPlayer
			break
		}
	}

	// Update game state
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game after player production change", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Player production updated successfully")
	return nil
}

// GetPlayer retrieves player information
func (s *PlayerServiceImpl) GetPlayer(ctx context.Context, gameID, playerID string) (*model.Player, error) {
	return s.playerRepo.GetPlayer(ctx, gameID, playerID)
}

// ValidateProductionRequirement validates if player meets production requirements
func (s *PlayerServiceImpl) ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement model.ResourceSet) error {
	player, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player has sufficient production
	if player.Production.Credits < requirement.Credits ||
		player.Production.Steel < requirement.Steel ||
		player.Production.Titanium < requirement.Titanium ||
		player.Production.Plants < requirement.Plants ||
		player.Production.Energy < requirement.Energy ||
		player.Production.Heat < requirement.Heat {
		return fmt.Errorf("insufficient production to meet requirement")
	}

	return nil
}

// ValidateResourceCost validates if player can afford the resource cost
func (s *PlayerServiceImpl) ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	player, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player has sufficient resources
	if player.Resources.Credits < cost.Credits ||
		player.Resources.Steel < cost.Steel ||
		player.Resources.Titanium < cost.Titanium ||
		player.Resources.Plants < cost.Plants ||
		player.Resources.Energy < cost.Energy ||
		player.Resources.Heat < cost.Heat {
		return fmt.Errorf("insufficient resources to pay cost")
	}

	return nil
}

// AddProduction adds production to a player
func (s *PlayerServiceImpl) AddProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error {
	player, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Add production
	newProduction := model.Production{
		Credits:  player.Production.Credits + production.Credits,
		Steel:    player.Production.Steel + production.Steel,
		Titanium: player.Production.Titanium + production.Titanium,
		Plants:   player.Production.Plants + production.Plants,
		Energy:   player.Production.Energy + production.Energy,
		Heat:     player.Production.Heat + production.Heat,
	}

	return s.UpdatePlayerProduction(ctx, gameID, playerID, newProduction)
}

// PayResourceCost deducts resource cost from player
func (s *PlayerServiceImpl) PayResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	player, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Validate player can afford the cost
	if err := s.ValidateResourceCost(ctx, gameID, playerID, cost); err != nil {
		return err
	}

	// Deduct resources
	newResources := model.Resources{
		Credits:  player.Resources.Credits - cost.Credits,
		Steel:    player.Resources.Steel - cost.Steel,
		Titanium: player.Resources.Titanium - cost.Titanium,
		Plants:   player.Resources.Plants - cost.Plants,
		Energy:   player.Resources.Energy - cost.Energy,
		Heat:     player.Resources.Heat - cost.Heat,
	}

	return s.UpdatePlayerResources(ctx, gameID, playerID, newResources)
}

// AddResources adds resources to a player
func (s *PlayerServiceImpl) AddResources(ctx context.Context, gameID, playerID string, resources model.ResourceSet) error {
	player, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Add resources
	newResources := model.Resources{
		Credits:  player.Resources.Credits + resources.Credits,
		Steel:    player.Resources.Steel + resources.Steel,
		Titanium: player.Resources.Titanium + resources.Titanium,
		Plants:   player.Resources.Plants + resources.Plants,
		Energy:   player.Resources.Energy + resources.Energy,
		Heat:     player.Resources.Heat + resources.Heat,
	}

	return s.UpdatePlayerResources(ctx, gameID, playerID, newResources)
}