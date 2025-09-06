package model_test

import (
	"terraforming-mars-backend/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResources_Zero(t *testing.T) {
	resources := model.Resources{}

	assert.Equal(t, 0, resources.Credits)
	assert.Equal(t, 0, resources.Steel)
	assert.Equal(t, 0, resources.Titanium)
	assert.Equal(t, 0, resources.Plants)
	assert.Equal(t, 0, resources.Energy)
	assert.Equal(t, 0, resources.Heat)
}

func TestResources_Add(t *testing.T) {
	r1 := model.Resources{
		Credits:  10,
		Steel:    5,
		Titanium: 3,
		Plants:   8,
		Energy:   2,
		Heat:     4,
	}

	r2 := model.Resources{
		Credits:  5,
		Steel:    2,
		Titanium: 1,
		Plants:   3,
		Energy:   4,
		Heat:     2,
	}

	expected := model.Resources{
		Credits:  15,
		Steel:    7,
		Titanium: 4,
		Plants:   11,
		Energy:   6,
		Heat:     6,
	}

	result := r1.Add(r2)
	assert.Equal(t, expected, result)

	// Verify original resources are unchanged
	assert.Equal(t, 10, r1.Credits)
	assert.Equal(t, 5, r2.Credits)
}

func TestResources_Subtract(t *testing.T) {
	r1 := model.Resources{
		Credits:  20,
		Steel:    10,
		Titanium: 5,
		Plants:   15,
		Energy:   8,
		Heat:     12,
	}

	r2 := model.Resources{
		Credits:  5,
		Steel:    3,
		Titanium: 2,
		Plants:   4,
		Energy:   3,
		Heat:     6,
	}

	expected := model.Resources{
		Credits:  15,
		Steel:    7,
		Titanium: 3,
		Plants:   11,
		Energy:   5,
		Heat:     6,
	}

	result := r1.Subtract(r2)
	assert.Equal(t, expected, result)
}

func TestResources_Subtract_NegativeResults(t *testing.T) {
	r1 := model.Resources{
		Credits:  5,
		Steel:    2,
		Titanium: 1,
		Plants:   3,
		Energy:   1,
		Heat:     2,
	}

	r2 := model.Resources{
		Credits:  10,
		Steel:    5,
		Titanium: 3,
		Plants:   6,
		Energy:   4,
		Heat:     8,
	}

	expected := model.Resources{
		Credits:  -5,
		Steel:    -3,
		Titanium: -2,
		Plants:   -3,
		Energy:   -3,
		Heat:     -6,
	}

	result := r1.Subtract(r2)
	assert.Equal(t, expected, result)
}

func TestResources_HasNegative(t *testing.T) {
	tests := []struct {
		name      string
		resources model.Resources
		expected  bool
	}{
		{
			"All positive",
			model.Resources{Credits: 10, Steel: 5, Titanium: 3, Plants: 8, Energy: 2, Heat: 4},
			false,
		},
		{
			"All zero",
			model.Resources{Credits: 0, Steel: 0, Titanium: 0, Plants: 0, Energy: 0, Heat: 0},
			false,
		},
		{
			"Negative credits",
			model.Resources{Credits: -1, Steel: 5, Titanium: 3, Plants: 8, Energy: 2, Heat: 4},
			true,
		},
		{
			"Negative steel",
			model.Resources{Credits: 10, Steel: -2, Titanium: 3, Plants: 8, Energy: 2, Heat: 4},
			true,
		},
		{
			"Multiple negative",
			model.Resources{Credits: -5, Steel: 5, Titanium: -1, Plants: 8, Energy: 2, Heat: 4},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.resources.HasNegative()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResources_CanAfford(t *testing.T) {
	playerResources := model.Resources{
		Credits:  20,
		Steel:    10,
		Titanium: 5,
		Plants:   15,
		Energy:   8,
		Heat:     12,
	}

	tests := []struct {
		name     string
		cost     model.Resources
		expected bool
	}{
		{
			"Can afford all resources",
			model.Resources{Credits: 10, Steel: 5, Titanium: 2, Plants: 8, Energy: 4, Heat: 6},
			true,
		},
		{
			"Can afford exact resources",
			model.Resources{Credits: 20, Steel: 10, Titanium: 5, Plants: 15, Energy: 8, Heat: 12},
			true,
		},
		{
			"Cannot afford - credits too high",
			model.Resources{Credits: 25, Steel: 5, Titanium: 2, Plants: 8, Energy: 4, Heat: 6},
			false,
		},
		{
			"Cannot afford - steel too high",
			model.Resources{Credits: 10, Steel: 15, Titanium: 2, Plants: 8, Energy: 4, Heat: 6},
			false,
		},
		{
			"Can afford zero cost",
			model.Resources{Credits: 0, Steel: 0, Titanium: 0, Plants: 0, Energy: 0, Heat: 0},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := playerResources.CanAfford(tt.cost)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProduction_Zero(t *testing.T) {
	production := model.Production{}

	assert.Equal(t, 0, production.Credits)
	assert.Equal(t, 0, production.Steel)
	assert.Equal(t, 0, production.Titanium)
	assert.Equal(t, 0, production.Plants)
	assert.Equal(t, 0, production.Energy)
	assert.Equal(t, 0, production.Heat)
}

func TestProduction_Add(t *testing.T) {
	p1 := model.Production{
		Credits:  2,
		Steel:    1,
		Titanium: 1,
		Plants:   3,
		Energy:   2,
		Heat:     1,
	}

	p2 := model.Production{
		Credits:  1,
		Steel:    2,
		Titanium: 0,
		Plants:   1,
		Energy:   1,
		Heat:     2,
	}

	expected := model.Production{
		Credits:  3,
		Steel:    3,
		Titanium: 1,
		Plants:   4,
		Energy:   3,
		Heat:     3,
	}

	result := p1.Add(p2)
	assert.Equal(t, expected, result)
}

func TestProduction_Subtract(t *testing.T) {
	p1 := model.Production{
		Credits:  5,
		Steel:    4,
		Titanium: 3,
		Plants:   6,
		Energy:   4,
		Heat:     3,
	}

	p2 := model.Production{
		Credits:  2,
		Steel:    1,
		Titanium: 1,
		Plants:   2,
		Energy:   2,
		Heat:     1,
	}

	expected := model.Production{
		Credits:  3,
		Steel:    3,
		Titanium: 2,
		Plants:   4,
		Energy:   2,
		Heat:     2,
	}

	result := p1.Subtract(p2)
	assert.Equal(t, expected, result)
}
