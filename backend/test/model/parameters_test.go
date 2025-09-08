package model_test

import (
	"terraforming-mars-backend/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalParameters_InitialState(t *testing.T) {
	params := model.GlobalParameters{}

	// Test zero values
	assert.Equal(t, 0, params.Temperature)
	assert.Equal(t, 0, params.Oxygen)
	assert.Equal(t, 0, params.Oceans)
}

func TestGlobalParameters_MarsStartingConditions(t *testing.T) {
	params := model.GlobalParameters{
		Temperature: -30,
		Oxygen:      0,
		Oceans:      0,
	}

	assert.Equal(t, -30, params.Temperature)
	assert.Equal(t, 0, params.Oxygen)
	assert.Equal(t, 0, params.Oceans)
}

func TestConstants(t *testing.T) {
	// Test model constants are defined
	assert.Equal(t, -30, model.MinTemperature)
	assert.Equal(t, 8, model.MaxTemperature)
	assert.Equal(t, 0, model.MinOxygen)
	assert.Equal(t, 14, model.MaxOxygen)
	assert.Equal(t, 0, model.MinOceans)
	assert.Equal(t, 9, model.MaxOceans)
}
