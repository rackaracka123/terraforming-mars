package resources

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// Service handles resource and production operations for players.
//
// Scope: Isolated resource management mechanic
//   - Resource validation (can player afford X?)
//   - Resource updates (add/subtract resources)
//   - Production validation and updates
//
// This mechanic is ISOLATED and should NOT:
//   - Call other mechanic services
//   - Handle game flow or orchestration
//   - Manage turn state or phases
//
// Dependencies:
//   - PlayerRepository (for reading/updating player resources)
type Service interface {
	// Resource operations
	AddResources(ctx context.Context, gameID, playerID string, resources model.ResourceSet) error
	ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error
	PayResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error

	// Production operations
	AddProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error
	ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement model.ResourceSet) error

	// Helper operations
	CanAffordCost(ctx context.Context, gameID, playerID string, cost int) (bool, error)
	GetPlayerResources(ctx context.Context, gameID, playerID string) (model.Resources, error)
	GetPlayerProduction(ctx context.Context, gameID, playerID string) (model.Production, error)
}

// ServiceImpl implements the Resources mechanic service
type ServiceImpl struct {
	playerRepo repository.PlayerRepository
}

// NewService creates a new Resources mechanic service
func NewService(playerRepo repository.PlayerRepository) Service {
	return &ServiceImpl{
		playerRepo: playerRepo,
	}
}

// AddResources adds resources to a player
func (s *ServiceImpl) AddResources(ctx context.Context, gameID, playerID string, resources model.ResourceSet) error {
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
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

	// Update resources via repository
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
		return fmt.Errorf("failed to update player resources: %w", err)
	}
	return nil
}

// ValidateResourceCost validates if player can afford the resource cost
func (s *ServiceImpl) ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
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

// PayResourceCost deducts resource cost from player
func (s *ServiceImpl) PayResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
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

	// Update resources via repository
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
		return fmt.Errorf("failed to update player resources: %w", err)
	}
	return nil
}

// AddProduction adds production to a player
func (s *ServiceImpl) AddProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error {
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
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

	// Update production via repository
	if err := s.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
		return fmt.Errorf("failed to update player production: %w", err)
	}
	return nil
}

// ValidateProductionRequirement validates if player meets production requirements
func (s *ServiceImpl) ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement model.ResourceSet) error {
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
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

// CanAffordCost checks if player can afford a specific credit cost
func (s *ServiceImpl) CanAffordCost(ctx context.Context, gameID, playerID string, cost int) (bool, error) {
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return false, fmt.Errorf("failed to get player: %w", err)
	}

	return player.Resources.Credits >= cost, nil
}

// GetPlayerResources retrieves a player's current resources
func (s *ServiceImpl) GetPlayerResources(ctx context.Context, gameID, playerID string) (model.Resources, error) {
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return model.Resources{}, fmt.Errorf("failed to get player: %w", err)
	}

	return player.Resources, nil
}

// GetPlayerProduction retrieves a player's current production
func (s *ServiceImpl) GetPlayerProduction(ctx context.Context, gameID, playerID string) (model.Production, error) {
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return model.Production{}, fmt.Errorf("failed to get player: %w", err)
	}

	return player.Production, nil
}
