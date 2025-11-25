package game

import (
	"testing"

	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/types"
)

func TestResourceManager_ApplyResourceChange(t *testing.T) {
	rm := game.NewResourceManager()

	tests := []struct {
		name         string
		initial      types.Resources
		resourceType types.ResourceType
		amount       int
		expected     types.Resources
		wantErr      bool
	}{
		{
			name:         "add credits",
			initial:      types.Resources{Credits: 10},
			resourceType: types.ResourceCredits,
			amount:       5,
			expected:     types.Resources{Credits: 15},
			wantErr:      false,
		},
		{
			name:         "subtract credits",
			initial:      types.Resources{Credits: 10},
			resourceType: types.ResourceCredits,
			amount:       -3,
			expected:     types.Resources{Credits: 7},
			wantErr:      false,
		},
		{
			name:         "add steel",
			initial:      types.Resources{Steel: 2},
			resourceType: types.ResourceSteel,
			amount:       4,
			expected:     types.Resources{Steel: 6},
			wantErr:      false,
		},
		{
			name:         "add titanium",
			initial:      types.Resources{Titanium: 1},
			resourceType: types.ResourceTitanium,
			amount:       2,
			expected:     types.Resources{Titanium: 3},
			wantErr:      false,
		},
		{
			name:         "add plants",
			initial:      types.Resources{Plants: 5},
			resourceType: types.ResourcePlants,
			amount:       3,
			expected:     types.Resources{Plants: 8},
			wantErr:      false,
		},
		{
			name:         "add energy",
			initial:      types.Resources{Energy: 2},
			resourceType: types.ResourceEnergy,
			amount:       4,
			expected:     types.Resources{Energy: 6},
			wantErr:      false,
		},
		{
			name:         "add heat",
			initial:      types.Resources{Heat: 3},
			resourceType: types.ResourceHeat,
			amount:       7,
			expected:     types.Resources{Heat: 10},
			wantErr:      false,
		},
		{
			name:         "unknown resource type",
			initial:      types.Resources{Credits: 10},
			resourceType: types.ResourceType("invalid"),
			amount:       5,
			expected:     types.Resources{Credits: 10},
			wantErr:      true,
		},
		{
			name:         "multiple resources unchanged",
			initial:      types.Resources{Credits: 10, Steel: 2, Titanium: 1},
			resourceType: types.ResourceCredits,
			amount:       5,
			expected:     types.Resources{Credits: 15, Steel: 2, Titanium: 1},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rm.ApplyResourceChange(tt.initial, tt.resourceType, tt.amount)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyResourceChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.expected {
				t.Errorf("ApplyResourceChange() = %+v, expected %+v", result, tt.expected)
			}
		})
	}
}

func TestResourceManager_ApplyProductionChange(t *testing.T) {
	rm := game.NewResourceManager()

	tests := []struct {
		name         string
		initial      types.Production
		resourceType types.ResourceType
		amount       int
		expected     types.Production
		wantErr      bool
	}{
		{
			name:         "increase credits production",
			initial:      types.Production{Credits: 2},
			resourceType: types.ResourceCreditsProduction,
			amount:       3,
			expected:     types.Production{Credits: 5},
			wantErr:      false,
		},
		{
			name:         "decrease credits production but not below zero",
			initial:      types.Production{Credits: 5},
			resourceType: types.ResourceCreditsProduction,
			amount:       -7,
			expected:     types.Production{Credits: 0},
			wantErr:      false,
		},
		{
			name:         "increase steel production",
			initial:      types.Production{Steel: 1},
			resourceType: types.ResourceSteelProduction,
			amount:       2,
			expected:     types.Production{Steel: 3},
			wantErr:      false,
		},
		{
			name:         "decrease steel production to zero",
			initial:      types.Production{Steel: 2},
			resourceType: types.ResourceSteelProduction,
			amount:       -2,
			expected:     types.Production{Steel: 0},
			wantErr:      false,
		},
		{
			name:         "increase titanium production",
			initial:      types.Production{Titanium: 0},
			resourceType: types.ResourceTitaniumProduction,
			amount:       1,
			expected:     types.Production{Titanium: 1},
			wantErr:      false,
		},
		{
			name:         "increase plants production",
			initial:      types.Production{Plants: 1},
			resourceType: types.ResourcePlantsProduction,
			amount:       2,
			expected:     types.Production{Plants: 3},
			wantErr:      false,
		},
		{
			name:         "increase energy production",
			initial:      types.Production{Energy: 2},
			resourceType: types.ResourceEnergyProduction,
			amount:       3,
			expected:     types.Production{Energy: 5},
			wantErr:      false,
		},
		{
			name:         "decrease energy production but clamped to zero",
			initial:      types.Production{Energy: 1},
			resourceType: types.ResourceEnergyProduction,
			amount:       -5,
			expected:     types.Production{Energy: 0},
			wantErr:      false,
		},
		{
			name:         "increase heat production",
			initial:      types.Production{Heat: 1},
			resourceType: types.ResourceHeatProduction,
			amount:       1,
			expected:     types.Production{Heat: 2},
			wantErr:      false,
		},
		{
			name:         "unknown production type",
			initial:      types.Production{Credits: 2},
			resourceType: types.ResourceType("invalid"),
			amount:       3,
			expected:     types.Production{Credits: 2},
			wantErr:      true,
		},
		{
			name:         "multiple production values unchanged",
			initial:      types.Production{Credits: 5, Steel: 2, Energy: 3},
			resourceType: types.ResourceCreditsProduction,
			amount:       2,
			expected:     types.Production{Credits: 7, Steel: 2, Energy: 3},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rm.ApplyProductionChange(tt.initial, tt.resourceType, tt.amount)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyProductionChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.expected {
				t.Errorf("ApplyProductionChange() = %+v, expected %+v", result, tt.expected)
			}
		})
	}
}

func TestResourceManager_GetResourceAmount(t *testing.T) {
	rm := game.NewResourceManager()

	resources := types.Resources{
		Credits:  10,
		Steel:    5,
		Titanium: 3,
		Plants:   8,
		Energy:   4,
		Heat:     12,
	}

	tests := []struct {
		name         string
		resourceType types.ResourceType
		expected     int
	}{
		{"get credits", types.ResourceCredits, 10},
		{"get steel", types.ResourceSteel, 5},
		{"get titanium", types.ResourceTitanium, 3},
		{"get plants", types.ResourcePlants, 8},
		{"get energy", types.ResourceEnergy, 4},
		{"get heat", types.ResourceHeat, 12},
		{"unknown type returns zero", types.ResourceType("invalid"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rm.GetResourceAmount(resources, tt.resourceType)
			if result != tt.expected {
				t.Errorf("GetResourceAmount() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestResourceManager_GetProductionAmount(t *testing.T) {
	rm := game.NewResourceManager()

	production := types.Production{
		Credits:  5,
		Steel:    2,
		Titanium: 1,
		Plants:   3,
		Energy:   4,
		Heat:     2,
	}

	tests := []struct {
		name         string
		resourceType types.ResourceType
		expected     int
	}{
		{"get credits production", types.ResourceCreditsProduction, 5},
		{"get steel production", types.ResourceSteelProduction, 2},
		{"get titanium production", types.ResourceTitaniumProduction, 1},
		{"get plants production", types.ResourcePlantsProduction, 3},
		{"get energy production", types.ResourceEnergyProduction, 4},
		{"get heat production", types.ResourceHeatProduction, 2},
		{"unknown type returns zero", types.ResourceType("invalid"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rm.GetProductionAmount(production, tt.resourceType)
			if result != tt.expected {
				t.Errorf("GetProductionAmount() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestResourceManager_ValidateHasResource(t *testing.T) {
	rm := game.NewResourceManager()

	resources := types.Resources{
		Credits:  10,
		Steel:    5,
		Titanium: 3,
		Plants:   8,
		Energy:   4,
		Heat:     12,
	}

	tests := []struct {
		name         string
		resourceType types.ResourceType
		required     int
		wantErr      bool
	}{
		{"sufficient credits", types.ResourceCredits, 10, false},
		{"insufficient credits", types.ResourceCredits, 11, true},
		{"sufficient steel", types.ResourceSteel, 5, false},
		{"insufficient steel", types.ResourceSteel, 6, true},
		{"sufficient titanium", types.ResourceTitanium, 2, false},
		{"insufficient titanium", types.ResourceTitanium, 4, true},
		{"sufficient plants", types.ResourcePlants, 8, false},
		{"insufficient plants", types.ResourcePlants, 9, true},
		{"sufficient energy", types.ResourceEnergy, 3, false},
		{"insufficient energy", types.ResourceEnergy, 5, true},
		{"sufficient heat", types.ResourceHeat, 12, false},
		{"insufficient heat", types.ResourceHeat, 13, true},
		{"zero required always passes", types.ResourceCredits, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rm.ValidateHasResource(resources, tt.resourceType, tt.required)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHasResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
