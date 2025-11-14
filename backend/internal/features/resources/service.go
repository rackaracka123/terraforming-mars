package resources

import (
	"context"
"terraforming-mars-backend/internal/model"
	"fmt"

	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Service handles resource and production operations.
//
// Scope: Isolated resource management for a single player
//   - Resource validation (can afford X?)
//   - Resource updates (add/subtract resources)
//   - model.Production validation and updates
//   - Owns its own repository (independent storage)
//
// This feature is ISOLATED and should NOT:
//   - Call other feature services
//   - Handle game flow or orchestration
//   - Manage turn state or phases
//
// Note: Each Player instance has its own ResourcesService
type Service interface {
	// Resource operations
	AddResources(ctx context.Context, resources ResourceSet) error
	CanAffordCost(ctx context.Context, cost ResourceSet) (bool, error)
	PayCost(ctx context.Context, cost ResourceSet) error
	Get(ctx context.Context) (model.Resources, error)

	// Individual resource operations
	AddCredits(ctx context.Context, amount int) error
	DeductCredits(ctx context.Context, amount int) error

	AddSteel(ctx context.Context, amount int) error
	DeductSteel(ctx context.Context, amount int) error

	AddTitanium(ctx context.Context, amount int) error
	DeductTitanium(ctx context.Context, amount int) error

	AddPlants(ctx context.Context, amount int) error
	DeductPlants(ctx context.Context, amount int) error

	AddEnergy(ctx context.Context, amount int) error
	DeductEnergy(ctx context.Context, amount int) error

	AddHeat(ctx context.Context, amount int) error
	DeductHeat(ctx context.Context, amount int) error

	// model.Production operations
	AddProduction(ctx context.Context, production ResourceSet) error
	CanMeetProductionRequirement(ctx context.Context, requirement ResourceSet) (bool, error)
	GetProduction(ctx context.Context) (model.Production, error)

	// Special operations
	ConvertEnergyToHeat(ctx context.Context) error

	// Initialize
}

// ServiceImpl implements the model.Resources service
type ServiceImpl struct {
	repo Repository
}

// NewService creates a new model.Resources service
func NewService(repo Repository) Service {
	return &ServiceImpl{
		repo: repo,
	}
}

// AddResources adds a set of resources
func (s *ServiceImpl) AddResources(ctx context.Context, resources ResourceSet) error {
	if resources.Credits > 0 {
		if err := s.repo.AddCredits(ctx, resources.Credits); err != nil {
			return fmt.Errorf("failed to add credits: %w", err)
		}
	}
	if resources.Steel > 0 {
		if err := s.repo.AddSteel(ctx, resources.Steel); err != nil {
			return fmt.Errorf("failed to add steel: %w", err)
		}
	}
	if resources.Titanium > 0 {
		if err := s.repo.AddTitanium(ctx, resources.Titanium); err != nil {
			return fmt.Errorf("failed to add titanium: %w", err)
		}
	}
	if resources.Plants > 0 {
		if err := s.repo.AddPlants(ctx, resources.Plants); err != nil {
			return fmt.Errorf("failed to add plants: %w", err)
		}
	}
	if resources.Energy > 0 {
		if err := s.repo.AddEnergy(ctx, resources.Energy); err != nil {
			return fmt.Errorf("failed to add energy: %w", err)
		}
	}
	if resources.Heat > 0 {
		if err := s.repo.AddHeat(ctx, resources.Heat); err != nil {
			return fmt.Errorf("failed to add heat: %w", err)
		}
	}

	logger.Get().Debug("model.Resources added",
		zap.Int("credits", resources.Credits),
		zap.Int("steel", resources.Steel),
		zap.Int("titanium", resources.Titanium),
		zap.Int("plants", resources.Plants),
		zap.Int("energy", resources.Energy),
		zap.Int("heat", resources.Heat))

	// TODO Phase 5: Publish ResourcesChangedEvent

	return nil
}

// CanAffordCost checks if player can afford the resource cost
func (s *ServiceImpl) CanAffordCost(ctx context.Context, cost ResourceSet) (bool, error) {
	currentResources, err := s.repo.Get(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get resources: %w", err)
	}

	canAfford := currentResources.Credits >= cost.Credits &&
		currentResources.Steel >= cost.Steel &&
		currentResources.Titanium >= cost.Titanium &&
		currentResources.Plants >= cost.Plants &&
		currentResources.Energy >= cost.Energy &&
		currentResources.Heat >= cost.Heat

	return canAfford, nil
}

// PayCost deducts resource cost
func (s *ServiceImpl) PayCost(ctx context.Context, cost ResourceSet) error {
	// Validate can afford
	canAfford, err := s.CanAffordCost(ctx, cost)
	if err != nil {
		// return err
	}
	if !canAfford {
		return fmt.Errorf("insufficient resources to pay cost")
	}

	// Deduct resources using granular methods
	if cost.Credits > 0 {
		if err := s.repo.DeductCredits(ctx, cost.Credits); err != nil {
			return fmt.Errorf("failed to deduct credits: %w", err)
		}
	}
	if cost.Steel > 0 {
		if err := s.repo.DeductSteel(ctx, cost.Steel); err != nil {
			return fmt.Errorf("failed to deduct steel: %w", err)
		}
	}
	if cost.Titanium > 0 {
		if err := s.repo.DeductTitanium(ctx, cost.Titanium); err != nil {
			return fmt.Errorf("failed to deduct titanium: %w", err)
		}
	}
	if cost.Plants > 0 {
		if err := s.repo.DeductPlants(ctx, cost.Plants); err != nil {
			return fmt.Errorf("failed to deduct plants: %w", err)
		}
	}
	if cost.Energy > 0 {
		if err := s.repo.DeductEnergy(ctx, cost.Energy); err != nil {
			return fmt.Errorf("failed to deduct energy: %w", err)
		}
	}
	if cost.Heat > 0 {
		if err := s.repo.DeductHeat(ctx, cost.Heat); err != nil {
			return fmt.Errorf("failed to deduct heat: %w", err)
		}
	}

	logger.Get().Debug("Cost paid",
		zap.Int("credits", cost.Credits),
		zap.Int("steel", cost.Steel),
		zap.Int("titanium", cost.Titanium),
		zap.Int("plants", cost.Plants),
		zap.Int("energy", cost.Energy),
		zap.Int("heat", cost.Heat))

	// TODO Phase 5: Publish ResourcesChangedEvent

	return nil
}

// Get retrieves current resources
func (s *ServiceImpl) Get(ctx context.Context) (model.Resources, error) {
	return s.repo.Get(ctx)
}

// Individual resource operations
func (s *ServiceImpl) AddCredits(ctx context.Context, amount int) error {
	if err := s.repo.AddCredits(ctx, amount); err != nil {
		// return err
	}
	logger.Get().Debug("💰 Credits added", zap.Int("amount", amount))
	return nil
}

func (s *ServiceImpl) DeductCredits(ctx context.Context, amount int) error {
	if err := s.repo.DeductCredits(ctx, amount); err != nil {
		// return err
	}
	logger.Get().Debug("💸 Credits deducted", zap.Int("amount", amount))
	return nil
}

func (s *ServiceImpl) AddSteel(ctx context.Context, amount int) error {
	if err := s.repo.AddSteel(ctx, amount); err != nil {
		// return err
	}
	logger.Get().Debug("⚙️ Steel added", zap.Int("amount", amount))
	return nil
}

func (s *ServiceImpl) DeductSteel(ctx context.Context, amount int) error {
	return s.repo.DeductSteel(ctx, amount)
}

func (s *ServiceImpl) AddTitanium(ctx context.Context, amount int) error {
	if err := s.repo.AddTitanium(ctx, amount); err != nil {
		// return err
	}
	logger.Get().Debug("🔩 Titanium added", zap.Int("amount", amount))
	return nil
}

func (s *ServiceImpl) DeductTitanium(ctx context.Context, amount int) error {
	return s.repo.DeductTitanium(ctx, amount)
}

func (s *ServiceImpl) AddPlants(ctx context.Context, amount int) error {
	if err := s.repo.AddPlants(ctx, amount); err != nil {
		// return err
	}
	logger.Get().Debug("🌱 Plants added", zap.Int("amount", amount))
	return nil
}

func (s *ServiceImpl) DeductPlants(ctx context.Context, amount int) error {
	return s.repo.DeductPlants(ctx, amount)
}

func (s *ServiceImpl) AddEnergy(ctx context.Context, amount int) error {
	if err := s.repo.AddEnergy(ctx, amount); err != nil {
		// return err
	}
	logger.Get().Debug("⚡ Energy added", zap.Int("amount", amount))
	return nil
}

func (s *ServiceImpl) DeductEnergy(ctx context.Context, amount int) error {
	return s.repo.DeductEnergy(ctx, amount)
}

func (s *ServiceImpl) AddHeat(ctx context.Context, amount int) error {
	if err := s.repo.AddHeat(ctx, amount); err != nil {
		// return err
	}
	logger.Get().Debug("🔥 Heat added", zap.Int("amount", amount))
	return nil
}

func (s *ServiceImpl) DeductHeat(ctx context.Context, amount int) error {
	return s.repo.DeductHeat(ctx, amount)
}

// model.Production operations
func (s *ServiceImpl) AddProduction(ctx context.Context, production ResourceSet) error {
	if production.Credits != 0 {
		if production.Credits > 0 {
			s.repo.IncreaseCreditsProduction(ctx, production.Credits)
		} else {
			s.repo.DecreaseCreditsProduction(ctx, -production.Credits)
		}
	}
	if production.Steel != 0 {
		if production.Steel > 0 {
			s.repo.IncreaseSteelProduction(ctx, production.Steel)
		} else {
			s.repo.DecreaseSteelProduction(ctx, -production.Steel)
		}
	}
	if production.Titanium != 0 {
		if production.Titanium > 0 {
			s.repo.IncreaseTitaniumProduction(ctx, production.Titanium)
		} else {
			s.repo.DecreaseTitaniumProduction(ctx, -production.Titanium)
		}
	}
	if production.Plants != 0 {
		if production.Plants > 0 {
			s.repo.IncreasePlantsProduction(ctx, production.Plants)
		} else {
			s.repo.DecreasePlantsProduction(ctx, -production.Plants)
		}
	}
	if production.Energy != 0 {
		if production.Energy > 0 {
			s.repo.IncreaseEnergyProduction(ctx, production.Energy)
		} else {
			s.repo.DecreaseEnergyProduction(ctx, -production.Energy)
		}
	}
	if production.Heat != 0 {
		if production.Heat > 0 {
			s.repo.IncreaseHeatProduction(ctx, production.Heat)
		} else {
			s.repo.DecreaseHeatProduction(ctx, -production.Heat)
		}
	}

	logger.Get().Debug("model.Production updated",
		zap.Int("credits", production.Credits),
		zap.Int("steel", production.Steel),
		zap.Int("titanium", production.Titanium),
		zap.Int("plants", production.Plants),
		zap.Int("energy", production.Energy),
		zap.Int("heat", production.Heat))

	// TODO Phase 5: Publish ProductionChangedEvent

	return nil
}

// CanMeetProductionRequirement checks if player meets production requirements
func (s *ServiceImpl) CanMeetProductionRequirement(ctx context.Context, requirement ResourceSet) (bool, error) {
	currentProduction, err := s.repo.GetProduction(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get production: %w", err)
	}

	canMeet := currentProduction.Credits >= requirement.Credits &&
		currentProduction.Steel >= requirement.Steel &&
		currentProduction.Titanium >= requirement.Titanium &&
		currentProduction.Plants >= requirement.Plants &&
		currentProduction.Energy >= requirement.Energy &&
		currentProduction.Heat >= requirement.Heat

	return canMeet, nil
}

// GetProduction retrieves current production
func (s *ServiceImpl) GetProduction(ctx context.Context) (model.Production, error) {
	return s.repo.GetProduction(ctx)
}

// ConvertEnergyToHeat converts all energy to heat (production phase)
func (s *ServiceImpl) ConvertEnergyToHeat(ctx context.Context) error {
	if err := s.repo.ConvertEnergyToHeat(ctx); err != nil {
		return fmt.Errorf("failed to convert energy to heat: %w", err)
	}

	logger.Get().Debug("⚡→🔥 Energy converted to heat")

	return nil
}
