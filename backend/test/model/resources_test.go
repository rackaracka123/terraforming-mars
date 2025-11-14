package model_test

import (
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResources_InitialState(t *testing.T) {
	resources := resources.Resources{}

	// Test zero values
	assert.Equal(t, 0, resources.Credits)
	assert.Equal(t, 0, resources.Steel)
	assert.Equal(t, 0, resources.Titanium)
	assert.Equal(t, 0, resources.Plants)
	assert.Equal(t, 0, resources.Energy)
	assert.Equal(t, 0, resources.Heat)
}

func TestProduction_InitialState(t *testing.T) {
	production := resources.Production{}

	// Test zero values
	assert.Equal(t, 0, production.Credits)
	assert.Equal(t, 0, production.Steel)
	assert.Equal(t, 0, production.Titanium)
	assert.Equal(t, 0, production.Plants)
	assert.Equal(t, 0, production.Energy)
	assert.Equal(t, 0, production.Heat)
}

func TestResources_AllFields(t *testing.T) {
	resources := resources.Resources{
		Credits:  100,
		Steel:    25,
		Titanium: 15,
		Plants:   30,
		Energy:   20,
		Heat:     45,
	}

	// Test all fields are set correctly
	assert.Equal(t, 100, resources.Credits)
	assert.Equal(t, 25, resources.Steel)
	assert.Equal(t, 15, resources.Titanium)
	assert.Equal(t, 30, resources.Plants)
	assert.Equal(t, 20, resources.Energy)
	assert.Equal(t, 45, resources.Heat)
}

func TestProduction_AllFields(t *testing.T) {
	production := resources.Production{
		Credits:  3,
		Steel:    1,
		Titanium: 2,
		Plants:   1,
		Energy:   4,
		Heat:     2,
	}

	// Test all fields are set correctly
	assert.Equal(t, 3, production.Credits)
	assert.Equal(t, 1, production.Steel)
	assert.Equal(t, 2, production.Titanium)
	assert.Equal(t, 1, production.Plants)
	assert.Equal(t, 4, production.Energy)
	assert.Equal(t, 2, production.Heat)
}
