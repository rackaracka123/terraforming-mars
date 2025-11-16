package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"terraforming-mars-backend/internal/domain"
)

func TestNewResourceSet(t *testing.T) {
	rs := domain.NewResourceSet()

	assert.Equal(t, 0, rs.Credits)
	assert.Equal(t, 0, rs.Steel)
	assert.Equal(t, 0, rs.Titanium)
	assert.Equal(t, 0, rs.Plants)
	assert.Equal(t, 0, rs.Energy)
	assert.Equal(t, 0, rs.Heat)
}

func TestResourceSet_Add(t *testing.T) {
	tests := []struct {
		name     string
		initial  domain.ResourceSet
		toAdd    domain.ResourceSet
		expected domain.ResourceSet
	}{
		{
			name:     "Add to empty set",
			initial:  domain.ResourceSet{},
			toAdd:    domain.ResourceSet{Credits: 10, Steel: 5, Titanium: 3},
			expected: domain.ResourceSet{Credits: 10, Steel: 5, Titanium: 3},
		},
		{
			name:     "Add to existing resources",
			initial:  domain.ResourceSet{Credits: 15, Plants: 8, Heat: 4},
			toAdd:    domain.ResourceSet{Credits: 5, Plants: 2, Energy: 3},
			expected: domain.ResourceSet{Credits: 20, Plants: 10, Heat: 4, Energy: 3},
		},
		{
			name:     "Add all resource types",
			initial:  domain.ResourceSet{Credits: 1, Steel: 1, Titanium: 1, Plants: 1, Energy: 1, Heat: 1},
			toAdd:    domain.ResourceSet{Credits: 2, Steel: 2, Titanium: 2, Plants: 2, Energy: 2, Heat: 2},
			expected: domain.ResourceSet{Credits: 3, Steel: 3, Titanium: 3, Plants: 3, Energy: 3, Heat: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := tt.initial
			rs.Add(tt.toAdd)
			assert.Equal(t, tt.expected, rs)
		})
	}
}

func TestResourceSet_Subtract(t *testing.T) {
	tests := []struct {
		name        string
		initial     domain.ResourceSet
		toSubtract  domain.ResourceSet
		expected    domain.ResourceSet
		description string
	}{
		{
			name:        "Subtract from resources",
			initial:     domain.ResourceSet{Credits: 20, Steel: 10, Titanium: 5},
			toSubtract:  domain.ResourceSet{Credits: 5, Steel: 3, Titanium: 2},
			expected:    domain.ResourceSet{Credits: 15, Steel: 7, Titanium: 3},
			description: "Basic subtraction",
		},
		{
			name:        "Subtract to zero",
			initial:     domain.ResourceSet{Credits: 10, Plants: 8},
			toSubtract:  domain.ResourceSet{Credits: 10, Plants: 8},
			expected:    domain.ResourceSet{},
			description: "Subtract exact amount",
		},
		{
			name:        "Subtract can go negative",
			initial:     domain.ResourceSet{Credits: 5},
			toSubtract:  domain.ResourceSet{Credits: 10},
			expected:    domain.ResourceSet{Credits: -5},
			description: "Allow negative values (for production)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := tt.initial
			rs.Subtract(tt.toSubtract)
			assert.Equal(t, tt.expected, rs)
		})
	}
}

func TestResourceSet_CanAfford(t *testing.T) {
	tests := []struct {
		name      string
		resources domain.ResourceSet
		cost      domain.ResourceSet
		canAfford bool
	}{
		{
			name:      "Can afford exact amount",
			resources: domain.ResourceSet{Credits: 10, Steel: 5},
			cost:      domain.ResourceSet{Credits: 10, Steel: 5},
			canAfford: true,
		},
		{
			name:      "Can afford with surplus",
			resources: domain.ResourceSet{Credits: 20, Steel: 10, Titanium: 5},
			cost:      domain.ResourceSet{Credits: 10, Steel: 3},
			canAfford: true,
		},
		{
			name:      "Cannot afford - insufficient credits",
			resources: domain.ResourceSet{Credits: 5, Steel: 10},
			cost:      domain.ResourceSet{Credits: 10, Steel: 5},
			canAfford: false,
		},
		{
			name:      "Cannot afford - insufficient steel",
			resources: domain.ResourceSet{Credits: 20, Steel: 2},
			cost:      domain.ResourceSet{Credits: 10, Steel: 5},
			canAfford: false,
		},
		{
			name:      "Can afford zero cost",
			resources: domain.ResourceSet{Credits: 5},
			cost:      domain.ResourceSet{},
			canAfford: true,
		},
		{
			name:      "Cannot afford with zero resources",
			resources: domain.ResourceSet{},
			cost:      domain.ResourceSet{Credits: 1},
			canAfford: false,
		},
		{
			name:      "Can afford all resource types",
			resources: domain.ResourceSet{Credits: 30, Steel: 10, Titanium: 8, Plants: 12, Energy: 6, Heat: 15},
			cost:      domain.ResourceSet{Credits: 25, Steel: 8, Titanium: 5, Plants: 10, Energy: 5, Heat: 12},
			canAfford: true,
		},
		{
			name:      "Cannot afford - one resource insufficient",
			resources: domain.ResourceSet{Credits: 30, Steel: 10, Titanium: 3, Plants: 12, Energy: 6, Heat: 15},
			cost:      domain.ResourceSet{Credits: 25, Steel: 8, Titanium: 5, Plants: 10, Energy: 5, Heat: 12},
			canAfford: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.resources.CanAfford(tt.cost)
			assert.Equal(t, tt.canAfford, result)
		})
	}
}

func TestResourceSet_IsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		resources domain.ResourceSet
		isEmpty   bool
	}{
		{
			name:      "Empty resource set",
			resources: domain.ResourceSet{},
			isEmpty:   true,
		},
		{
			name:      "Has credits only",
			resources: domain.ResourceSet{Credits: 1},
			isEmpty:   false,
		},
		{
			name:      "Has steel only",
			resources: domain.ResourceSet{Steel: 1},
			isEmpty:   false,
		},
		{
			name:      "Has multiple resources",
			resources: domain.ResourceSet{Credits: 5, Plants: 3, Heat: 2},
			isEmpty:   false,
		},
		{
			name:      "All zeros (explicit)",
			resources: domain.ResourceSet{Credits: 0, Steel: 0, Titanium: 0, Plants: 0, Energy: 0, Heat: 0},
			isEmpty:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.resources.IsEmpty()
			assert.Equal(t, tt.isEmpty, result)
		})
	}
}

func TestStandardProjectCosts_Validation(t *testing.T) {
	t.Run("Sell Patents cost is zero", func(t *testing.T) {
		assert.True(t, domain.StandardProjectCosts.SellPatents.IsEmpty())
	})

	t.Run("Power Plant costs 11 MC", func(t *testing.T) {
		assert.Equal(t, 11, domain.StandardProjectCosts.PowerPlant.Credits)
		assert.Equal(t, 0, domain.StandardProjectCosts.PowerPlant.Steel)
	})

	t.Run("Asteroid costs 14 MC", func(t *testing.T) {
		assert.Equal(t, 14, domain.StandardProjectCosts.Asteroid.Credits)
	})

	t.Run("Aquifer costs 18 MC", func(t *testing.T) {
		assert.Equal(t, 18, domain.StandardProjectCosts.Aquifer.Credits)
	})

	t.Run("Greenery costs 23 MC", func(t *testing.T) {
		assert.Equal(t, 23, domain.StandardProjectCosts.Greenery.Credits)
	})

	t.Run("City costs 25 MC", func(t *testing.T) {
		assert.Equal(t, 25, domain.StandardProjectCosts.City.Credits)
	})

	t.Run("Convert Heat to Temperature costs 8 heat", func(t *testing.T) {
		assert.Equal(t, 8, domain.StandardProjectCosts.ConvertHeatToTemperature.Heat)
		assert.Equal(t, 0, domain.StandardProjectCosts.ConvertHeatToTemperature.Credits)
	})

	t.Run("Convert Plants to Greenery costs 8 plants", func(t *testing.T) {
		assert.Equal(t, 8, domain.StandardProjectCosts.ConvertPlantsToGreenery.Plants)
		assert.Equal(t, 0, domain.StandardProjectCosts.ConvertPlantsToGreenery.Credits)
	})
}

func TestResourceType_Constants(t *testing.T) {
	// Verify resource type constants are defined correctly
	assert.Equal(t, domain.ResourceType("credits"), domain.ResourceTypeCredits)
	assert.Equal(t, domain.ResourceType("steel"), domain.ResourceTypeSteel)
	assert.Equal(t, domain.ResourceType("titanium"), domain.ResourceTypeTitanium)
	assert.Equal(t, domain.ResourceType("plants"), domain.ResourceTypePlants)
	assert.Equal(t, domain.ResourceType("energy"), domain.ResourceTypeEnergy)
	assert.Equal(t, domain.ResourceType("heat"), domain.ResourceTypeHeat)

	// Verify production constants
	assert.Equal(t, domain.ResourceType("credits-production"), domain.ResourceCreditsProduction)
	assert.Equal(t, domain.ResourceType("energy-production"), domain.ResourceEnergyProduction)

	// Verify tile constants
	assert.Equal(t, domain.ResourceType("city-tile"), domain.ResourceCityTile)
	assert.Equal(t, domain.ResourceType("greenery-tile"), domain.ResourceGreeneryTile)
	assert.Equal(t, domain.ResourceType("ocean-tile"), domain.ResourceOceanTile)

	// Verify global parameter constants
	assert.Equal(t, domain.ResourceType("temperature"), domain.ResourceTemperature)
	assert.Equal(t, domain.ResourceType("oxygen"), domain.ResourceOxygen)
	assert.Equal(t, domain.ResourceType("oceans"), domain.ResourceOceans)
}
