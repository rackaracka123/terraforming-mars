package game_rules_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"terraforming-mars-backend/internal/domain"
)

// Test standard project costs from TERRAFORMING_MARS_RULES.md

func TestStandardProject_SellPatents(t *testing.T) {
	// Sell Patents: Discard cards for 1 MC each
	cost := domain.StandardProjectCosts.SellPatents

	assert.True(t, cost.IsEmpty(), "Sell Patents should have no upfront cost")
	assert.Equal(t, 0, cost.Credits, "Sell Patents costs 0 MC upfront")

	// Note: The actual revenue (1 MC per card) is calculated per card sold
}

func TestStandardProject_PowerPlant(t *testing.T) {
	// Power Plant: Spend 11 MC for +1 energy production
	cost := domain.StandardProjectCosts.PowerPlant

	assert.Equal(t, 11, cost.Credits, "Power Plant should cost 11 MC")
	assert.Equal(t, 0, cost.Steel, "Power Plant should not cost steel")
	assert.Equal(t, 0, cost.Titanium, "Power Plant should not cost titanium")
	assert.Equal(t, 0, cost.Plants, "Power Plant should not cost plants")
	assert.Equal(t, 0, cost.Energy, "Power Plant should not cost energy")
	assert.Equal(t, 0, cost.Heat, "Power Plant should not cost heat")
}

func TestStandardProject_Asteroid(t *testing.T) {
	// Asteroid: Spend 14 MC for +1 temperature
	cost := domain.StandardProjectCosts.Asteroid

	assert.Equal(t, 14, cost.Credits, "Launch Asteroid should cost 14 MC")
	assert.Equal(t, 0, cost.Steel, "Launch Asteroid should not cost other resources")
}

func TestStandardProject_Aquifer(t *testing.T) {
	// Aquifer: Spend 18 MC for 1 ocean tile
	cost := domain.StandardProjectCosts.Aquifer

	assert.Equal(t, 18, cost.Credits, "Build Aquifer should cost 18 MC")
	assert.Equal(t, 0, cost.Steel, "Build Aquifer should not cost other resources")
}

func TestStandardProject_Greenery(t *testing.T) {
	// Greenery: Spend 23 MC for 1 greenery tile
	cost := domain.StandardProjectCosts.Greenery

	assert.Equal(t, 23, cost.Credits, "Plant Greenery should cost 23 MC")
	assert.Equal(t, 0, cost.Steel, "Plant Greenery should not cost other resources")
}

func TestStandardProject_City(t *testing.T) {
	// City: Spend 25 MC for 1 city tile
	cost := domain.StandardProjectCosts.City

	assert.Equal(t, 25, cost.Credits, "Build City should cost 25 MC")
	assert.Equal(t, 0, cost.Steel, "Build City should not cost other resources")
}

func TestStandardProjects_CostProgression(t *testing.T) {
	// Standard projects should be in ascending order of cost
	costs := []struct {
		name   string
		amount int
	}{
		{"Sell Patents", 0},
		{"Power Plant", 11},
		{"Asteroid", 14},
		{"Aquifer", 18},
		{"Greenery", 23},
		{"City", 25},
	}

	for i := 1; i < len(costs); i++ {
		assert.Greater(t, costs[i].amount, costs[i-1].amount,
			"%s (%d MC) should cost more than %s (%d MC)",
			costs[i].name, costs[i].amount,
			costs[i-1].name, costs[i-1].amount)
	}
}

func TestStandardProjects_Affordability(t *testing.T) {
	tests := []struct {
		name      string
		credits   int
		project   string
		cost      domain.ResourceSet
		canAfford bool
	}{
		{
			name:      "Can afford Power Plant with exact amount",
			credits:   11,
			project:   "Power Plant",
			cost:      domain.StandardProjectCosts.PowerPlant,
			canAfford: true,
		},
		{
			name:      "Can afford Power Plant with surplus",
			credits:   20,
			project:   "Power Plant",
			cost:      domain.StandardProjectCosts.PowerPlant,
			canAfford: true,
		},
		{
			name:      "Cannot afford Power Plant",
			credits:   10,
			project:   "Power Plant",
			cost:      domain.StandardProjectCosts.PowerPlant,
			canAfford: false,
		},
		{
			name:      "Can afford Asteroid",
			credits:   14,
			project:   "Asteroid",
			cost:      domain.StandardProjectCosts.Asteroid,
			canAfford: true,
		},
		{
			name:      "Cannot afford Asteroid",
			credits:   13,
			project:   "Asteroid",
			cost:      domain.StandardProjectCosts.Asteroid,
			canAfford: false,
		},
		{
			name:      "Can afford Aquifer",
			credits:   18,
			project:   "Aquifer",
			cost:      domain.StandardProjectCosts.Aquifer,
			canAfford: true,
		},
		{
			name:      "Can afford Greenery",
			credits:   23,
			project:   "Greenery",
			cost:      domain.StandardProjectCosts.Greenery,
			canAfford: true,
		},
		{
			name:      "Cannot afford Greenery",
			credits:   22,
			project:   "Greenery",
			cost:      domain.StandardProjectCosts.Greenery,
			canAfford: false,
		},
		{
			name:      "Can afford City",
			credits:   25,
			project:   "City",
			cost:      domain.StandardProjectCosts.City,
			canAfford: true,
		},
		{
			name:      "Cannot afford City",
			credits:   24,
			project:   "City",
			cost:      domain.StandardProjectCosts.City,
			canAfford: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := domain.ResourceSet{Credits: tt.credits}
			canAfford := resources.CanAfford(tt.cost)
			assert.Equal(t, tt.canAfford, canAfford)
		})
	}
}

func TestStandardProjects_ResourceDeduction(t *testing.T) {
	t.Run("Power Plant deduction", func(t *testing.T) {
		resources := domain.ResourceSet{Credits: 30}
		cost := domain.StandardProjectCosts.PowerPlant

		resources.Subtract(cost)
		assert.Equal(t, 19, resources.Credits, "Should have 19 MC remaining after Power Plant")
	})

	t.Run("Asteroid deduction", func(t *testing.T) {
		resources := domain.ResourceSet{Credits: 30}
		cost := domain.StandardProjectCosts.Asteroid

		resources.Subtract(cost)
		assert.Equal(t, 16, resources.Credits, "Should have 16 MC remaining after Asteroid")
	})

	t.Run("Aquifer deduction", func(t *testing.T) {
		resources := domain.ResourceSet{Credits: 30}
		cost := domain.StandardProjectCosts.Aquifer

		resources.Subtract(cost)
		assert.Equal(t, 12, resources.Credits, "Should have 12 MC remaining after Aquifer")
	})

	t.Run("Greenery deduction", func(t *testing.T) {
		resources := domain.ResourceSet{Credits: 30}
		cost := domain.StandardProjectCosts.Greenery

		resources.Subtract(cost)
		assert.Equal(t, 7, resources.Credits, "Should have 7 MC remaining after Greenery")
	})

	t.Run("City deduction", func(t *testing.T) {
		resources := domain.ResourceSet{Credits: 30}
		cost := domain.StandardProjectCosts.City

		resources.Subtract(cost)
		assert.Equal(t, 5, resources.Credits, "Should have 5 MC remaining after City")
	})
}

func TestStandardProjects_MultipleActions(t *testing.T) {
	// Test player performing multiple standard projects in sequence
	resources := domain.ResourceSet{Credits: 100}

	// Buy Power Plant (11 MC)
	resources.Subtract(domain.StandardProjectCosts.PowerPlant)
	assert.Equal(t, 89, resources.Credits)

	// Launch Asteroid (14 MC)
	resources.Subtract(domain.StandardProjectCosts.Asteroid)
	assert.Equal(t, 75, resources.Credits)

	// Build Aquifer (18 MC)
	resources.Subtract(domain.StandardProjectCosts.Aquifer)
	assert.Equal(t, 57, resources.Credits)

	// Plant Greenery (23 MC)
	resources.Subtract(domain.StandardProjectCosts.Greenery)
	assert.Equal(t, 34, resources.Credits)

	// Build City (25 MC)
	resources.Subtract(domain.StandardProjectCosts.City)
	assert.Equal(t, 9, resources.Credits)

	// Total spent: 11 + 14 + 18 + 23 + 25 = 91 MC
	// Remaining: 100 - 91 = 9 MC
}
