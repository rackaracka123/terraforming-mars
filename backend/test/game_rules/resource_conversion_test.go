package game_rules_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"terraforming-mars-backend/internal/domain"
)

// Test resource conversion rules from TERRAFORMING_MARS_RULES.md

func TestHeatToTemperature_Cost(t *testing.T) {
	// Rule: 8 heat → +1 temperature step
	cost := domain.StandardProjectCosts.ConvertHeatToTemperature

	assert.Equal(t, 8, cost.Heat, "Converting heat to temperature should cost 8 heat")
	assert.Equal(t, 0, cost.Credits, "Heat conversion should not cost credits")
	assert.Equal(t, 0, cost.Steel, "Heat conversion should not cost steel")
	assert.Equal(t, 0, cost.Titanium, "Heat conversion should not cost titanium")
	assert.Equal(t, 0, cost.Plants, "Heat conversion should not cost plants")
	assert.Equal(t, 0, cost.Energy, "Heat conversion should not cost energy")
}

func TestPlantsToGreenery_Cost(t *testing.T) {
	// Rule: 8 plants → 1 greenery tile
	cost := domain.StandardProjectCosts.ConvertPlantsToGreenery

	assert.Equal(t, 8, cost.Plants, "Converting plants to greenery should cost 8 plants")
	assert.Equal(t, 0, cost.Credits, "Plants conversion should not cost credits")
	assert.Equal(t, 0, cost.Steel, "Plants conversion should not cost steel")
	assert.Equal(t, 0, cost.Titanium, "Plants conversion should not cost titanium")
	assert.Equal(t, 0, cost.Energy, "Plants conversion should not cost energy")
	assert.Equal(t, 0, cost.Heat, "Plants conversion should not cost heat")
}

func TestResourceSet_CanAfford_HeatConversion(t *testing.T) {
	tests := []struct {
		name      string
		heat      int
		canAfford bool
	}{
		{
			name:      "Exactly 8 heat",
			heat:      8,
			canAfford: true,
		},
		{
			name:      "More than 8 heat",
			heat:      15,
			canAfford: true,
		},
		{
			name:      "Less than 8 heat",
			heat:      7,
			canAfford: false,
		},
		{
			name:      "No heat",
			heat:      0,
			canAfford: false,
		},
		{
			name:      "16 heat (can convert twice)",
			heat:      16,
			canAfford: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := domain.ResourceSet{Heat: tt.heat}
			cost := domain.StandardProjectCosts.ConvertHeatToTemperature

			canAfford := resources.CanAfford(cost)
			assert.Equal(t, tt.canAfford, canAfford)
		})
	}
}

func TestResourceSet_CanAfford_PlantsConversion(t *testing.T) {
	tests := []struct {
		name      string
		plants    int
		canAfford bool
	}{
		{
			name:      "Exactly 8 plants",
			plants:    8,
			canAfford: true,
		},
		{
			name:      "More than 8 plants",
			plants:    12,
			canAfford: true,
		},
		{
			name:      "Less than 8 plants",
			plants:    7,
			canAfford: false,
		},
		{
			name:      "No plants",
			plants:    0,
			canAfford: false,
		},
		{
			name:      "16 plants (can convert twice)",
			plants:    16,
			canAfford: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := domain.ResourceSet{Plants: tt.plants}
			cost := domain.StandardProjectCosts.ConvertPlantsToGreenery

			canAfford := resources.CanAfford(cost)
			assert.Equal(t, tt.canAfford, canAfford)
		})
	}
}

func TestHeatConversion_DeductResources(t *testing.T) {
	// Simulate deducting heat for temperature conversion
	resources := domain.ResourceSet{
		Credits: 20,
		Heat:    15,
	}

	cost := domain.StandardProjectCosts.ConvertHeatToTemperature

	// Deduct the cost
	resources.Subtract(cost)

	assert.Equal(t, 20, resources.Credits, "Credits should remain unchanged")
	assert.Equal(t, 7, resources.Heat, "Should have 7 heat remaining after conversion")
}

func TestPlantsConversion_DeductResources(t *testing.T) {
	// Simulate deducting plants for greenery placement
	resources := domain.ResourceSet{
		Credits: 30,
		Plants:  12,
	}

	cost := domain.StandardProjectCosts.ConvertPlantsToGreenery

	// Deduct the cost
	resources.Subtract(cost)

	assert.Equal(t, 30, resources.Credits, "Credits should remain unchanged")
	assert.Equal(t, 4, resources.Plants, "Should have 4 plants remaining after conversion")
}

func TestMultipleConversions(t *testing.T) {
	t.Run("Multiple heat conversions", func(t *testing.T) {
		resources := domain.ResourceSet{Heat: 24} // Enough for 3 conversions

		cost := domain.StandardProjectCosts.ConvertHeatToTemperature

		// First conversion
		assert.True(t, resources.CanAfford(cost))
		resources.Subtract(cost)
		assert.Equal(t, 16, resources.Heat)

		// Second conversion
		assert.True(t, resources.CanAfford(cost))
		resources.Subtract(cost)
		assert.Equal(t, 8, resources.Heat)

		// Third conversion
		assert.True(t, resources.CanAfford(cost))
		resources.Subtract(cost)
		assert.Equal(t, 0, resources.Heat)

		// Fourth conversion - should not be affordable
		assert.False(t, resources.CanAfford(cost))
	})

	t.Run("Multiple plants conversions", func(t *testing.T) {
		resources := domain.ResourceSet{Plants: 25} // Enough for 3 conversions

		cost := domain.StandardProjectCosts.ConvertPlantsToGreenery

		// First conversion
		assert.True(t, resources.CanAfford(cost))
		resources.Subtract(cost)
		assert.Equal(t, 17, resources.Plants)

		// Second conversion
		assert.True(t, resources.CanAfford(cost))
		resources.Subtract(cost)
		assert.Equal(t, 9, resources.Plants)

		// Third conversion
		assert.True(t, resources.CanAfford(cost))
		resources.Subtract(cost)
		assert.Equal(t, 1, resources.Plants)

		// Fourth conversion - should not be affordable
		assert.False(t, resources.CanAfford(cost))
	})
}
