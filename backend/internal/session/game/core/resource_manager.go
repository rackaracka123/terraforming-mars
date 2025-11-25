package core

import (
	"fmt"

	"terraforming-mars-backend/internal/session/types"
)

// ResourceManager provides utility methods for handling resource and production changes
// Consolidates repetitive switch logic found across multiple files
type ResourceManager struct{}

// NewResourceManager creates a new ResourceManager instance
func NewResourceManager() *ResourceManager {
	return &ResourceManager{}
}

// ApplyResourceChange applies a resource change to the given Resources struct
// Returns the updated Resources and an error if the resource type is unknown
func (rm *ResourceManager) ApplyResourceChange(resources types.Resources, resourceType types.ResourceType, amount int) (types.Resources, error) {
	switch resourceType {
	case types.ResourceCredits:
		resources.Credits += amount
	case types.ResourceSteel:
		resources.Steel += amount
	case types.ResourceTitanium:
		resources.Titanium += amount
	case types.ResourcePlants:
		resources.Plants += amount
	case types.ResourceEnergy:
		resources.Energy += amount
	case types.ResourceHeat:
		resources.Heat += amount
	default:
		return resources, fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return resources, nil
}

// ApplyProductionChange applies a production change to the given Production struct
// Production values are clamped to a minimum of 0
// Returns the updated Production and an error if the resource type is unknown
func (rm *ResourceManager) ApplyProductionChange(production types.Production, resourceType types.ResourceType, amount int) (types.Production, error) {
	switch resourceType {
	case types.ResourceCreditsProduction:
		production.Credits += amount
		if production.Credits < 0 {
			production.Credits = 0
		}
	case types.ResourceSteelProduction:
		production.Steel += amount
		if production.Steel < 0 {
			production.Steel = 0
		}
	case types.ResourceTitaniumProduction:
		production.Titanium += amount
		if production.Titanium < 0 {
			production.Titanium = 0
		}
	case types.ResourcePlantsProduction:
		production.Plants += amount
		if production.Plants < 0 {
			production.Plants = 0
		}
	case types.ResourceEnergyProduction:
		production.Energy += amount
		if production.Energy < 0 {
			production.Energy = 0
		}
	case types.ResourceHeatProduction:
		production.Heat += amount
		if production.Heat < 0 {
			production.Heat = 0
		}
	default:
		return production, fmt.Errorf("unknown production type: %s", resourceType)
	}
	return production, nil
}

// GetResourceAmount returns the amount of a specific resource type from Resources struct
// Returns 0 for unknown resource types
func (rm *ResourceManager) GetResourceAmount(resources types.Resources, resourceType types.ResourceType) int {
	switch resourceType {
	case types.ResourceCredits:
		return resources.Credits
	case types.ResourceSteel:
		return resources.Steel
	case types.ResourceTitanium:
		return resources.Titanium
	case types.ResourcePlants:
		return resources.Plants
	case types.ResourceEnergy:
		return resources.Energy
	case types.ResourceHeat:
		return resources.Heat
	default:
		return 0
	}
}

// GetProductionAmount returns the amount of a specific production type from Production struct
// Returns 0 for unknown production types
func (rm *ResourceManager) GetProductionAmount(production types.Production, resourceType types.ResourceType) int {
	switch resourceType {
	case types.ResourceCreditsProduction:
		return production.Credits
	case types.ResourceSteelProduction:
		return production.Steel
	case types.ResourceTitaniumProduction:
		return production.Titanium
	case types.ResourcePlantsProduction:
		return production.Plants
	case types.ResourceEnergyProduction:
		return production.Energy
	case types.ResourceHeatProduction:
		return production.Heat
	default:
		return 0
	}
}

// ValidateHasResource validates that the given Resources has at least the specified amount
// Returns an error if insufficient resources are available
func (rm *ResourceManager) ValidateHasResource(resources types.Resources, resourceType types.ResourceType, requiredAmount int) error {
	currentAmount := rm.GetResourceAmount(resources, resourceType)
	if currentAmount < requiredAmount {
		return fmt.Errorf("insufficient %s: need %d, have %d", resourceType, requiredAmount, currentAmount)
	}
	return nil
}
