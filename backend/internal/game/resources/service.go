package resources

import (
	"context"
	"fmt"
)

// Service handles resource and production operations for players.
//
// Scope: Isolated resource management mechanic - VERTICAL SLICE
//   - Resource validation (can player afford X?)
//   - Resource updates (add/subtract resources)
//   - Production validation and updates
//   - Owns its own repository abstraction
//
// This mechanic is ISOLATED and should NOT:
//   - Call other mechanic services
//   - Handle game flow or orchestration
//   - Manage turn state or phases
//
// Dependencies:
//   - Repository (local abstraction for resource data access)
type Service interface {
	// Resource operations
	AddResources(ctx context.Context, gameID, playerID string, resources ResourceSet) error
	ValidateResourceCost(ctx context.Context, gameID, playerID string, cost ResourceSet) error
	PayResourceCost(ctx context.Context, gameID, playerID string, cost ResourceSet) error

	// Production operations
	AddProduction(ctx context.Context, gameID, playerID string, production ResourceSet) error
	ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement ResourceSet) error

	// Helper operations
	CanAffordCost(ctx context.Context, gameID, playerID string, cost int) (bool, error)
	GetPlayerResources(ctx context.Context, gameID, playerID string) (Resources, error)
	GetPlayerProduction(ctx context.Context, gameID, playerID string) (Production, error)
}

// ServiceImpl implements the Resources mechanic service
type ServiceImpl struct {
	repo Repository // Use local repository abstraction
}

// NewService creates a new Resources mechanic service
func NewService(repo Repository) Service {
	return &ServiceImpl{
		repo: repo,
	}
}

// AddResources adds resources to a player
func (s *ServiceImpl) AddResources(ctx context.Context, gameID, playerID string, resources ResourceSet) error {
	currentResources, err := s.repo.GetPlayerResources(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player resources: %w", err)
	}

	// Add resources
	newResources := Resources{
		Credits:  currentResources.Credits + resources.Credits,
		Steel:    currentResources.Steel + resources.Steel,
		Titanium: currentResources.Titanium + resources.Titanium,
		Plants:   currentResources.Plants + resources.Plants,
		Energy:   currentResources.Energy + resources.Energy,
		Heat:     currentResources.Heat + resources.Heat,
	}

	// Update resources via repository
	if err := s.repo.UpdatePlayerResources(ctx, gameID, playerID, newResources); err != nil {
		return fmt.Errorf("failed to update player resources: %w", err)
	}
	return nil
}

// ValidateResourceCost validates if player can afford the resource cost
func (s *ServiceImpl) ValidateResourceCost(ctx context.Context, gameID, playerID string, cost ResourceSet) error {
	resources, err := s.repo.GetPlayerResources(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player resources: %w", err)
	}

	// Check if player has sufficient resources
	if resources.Credits < cost.Credits ||
		resources.Steel < cost.Steel ||
		resources.Titanium < cost.Titanium ||
		resources.Plants < cost.Plants ||
		resources.Energy < cost.Energy ||
		resources.Heat < cost.Heat {
		return fmt.Errorf("insufficient resources to pay cost")
	}

	return nil
}

// PayResourceCost deducts resource cost from player
func (s *ServiceImpl) PayResourceCost(ctx context.Context, gameID, playerID string, cost ResourceSet) error {
	resources, err := s.repo.GetPlayerResources(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Validate player can afford the cost
	if err := s.ValidateResourceCost(ctx, gameID, playerID, cost); err != nil {
		return err
	}

	// Deduct resources
	newResources := Resources{
		Credits:  resources.Credits - cost.Credits,
		Steel:    resources.Steel - cost.Steel,
		Titanium: resources.Titanium - cost.Titanium,
		Plants:   resources.Plants - cost.Plants,
		Energy:   resources.Energy - cost.Energy,
		Heat:     resources.Heat - cost.Heat,
	}

	// Update resources via repository
	if err := s.repo.UpdatePlayerResources(ctx, gameID, playerID, newResources); err != nil {
		return fmt.Errorf("failed to update player resources: %w", err)
	}
	return nil
}

// AddProduction adds production to a player
func (s *ServiceImpl) AddProduction(ctx context.Context, gameID, playerID string, production ResourceSet) error {
	currentProduction, err := s.repo.GetPlayerProduction(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player production: %w", err)
	}

	// Add production
	newProduction := Production{
		Credits:  currentProduction.Credits + production.Credits,
		Steel:    currentProduction.Steel + production.Steel,
		Titanium: currentProduction.Titanium + production.Titanium,
		Plants:   currentProduction.Plants + production.Plants,
		Energy:   currentProduction.Energy + production.Energy,
		Heat:     currentProduction.Heat + production.Heat,
	}

	// Update production via repository
	if err := s.repo.UpdatePlayerProduction(ctx, gameID, playerID, newProduction); err != nil {
		return fmt.Errorf("failed to update player production: %w", err)
	}
	return nil
}

// ValidateProductionRequirement validates if player meets production requirements
func (s *ServiceImpl) ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement ResourceSet) error {
	production, err := s.repo.GetPlayerProduction(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player production: %w", err)
	}

	// Check if player has sufficient production
	if production.Credits < requirement.Credits ||
		production.Steel < requirement.Steel ||
		production.Titanium < requirement.Titanium ||
		production.Plants < requirement.Plants ||
		production.Energy < requirement.Energy ||
		production.Heat < requirement.Heat {
		return fmt.Errorf("insufficient production to meet requirement")
	}

	return nil
}

// CanAffordCost checks if player can afford a specific credit cost
func (s *ServiceImpl) CanAffordCost(ctx context.Context, gameID, playerID string, cost int) (bool, error) {
	resources, err := s.repo.GetPlayerResources(ctx, gameID, playerID)
	if err != nil {
		return false, fmt.Errorf("failed to get player resources: %w", err)
	}

	return resources.Credits >= cost, nil
}

// GetPlayerResources retrieves a player's current resources
func (s *ServiceImpl) GetPlayerResources(ctx context.Context, gameID, playerID string) (Resources, error) {
	resources, err := s.repo.GetPlayerResources(ctx, gameID, playerID)
	if err != nil {
		return Resources{}, fmt.Errorf("failed to get player resources: %w", err)
	}

	return resources, nil
}

// GetPlayerProduction retrieves a player's current production
func (s *ServiceImpl) GetPlayerProduction(ctx context.Context, gameID, playerID string) (Production, error) {
	production, err := s.repo.GetPlayerProduction(ctx, gameID, playerID)
	if err != nil {
		return Production{}, fmt.Errorf("failed to get player production: %w", err)
	}

	return production, nil
}
