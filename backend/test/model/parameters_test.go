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

func TestGlobalParameters_DeepCopy(t *testing.T) {
	original := &model.GlobalParameters{
		Temperature: 5,
		Oxygen:      10,
		Oceans:      3,
	}

	copy := original.DeepCopy()

	// Should be equal but different pointers
	assert.Equal(t, original.Temperature, copy.Temperature)
	assert.Equal(t, original.Oxygen, copy.Oxygen)
	assert.Equal(t, original.Oceans, copy.Oceans)
	assert.NotSame(t, original, copy)

	// Modifying copy should not affect original
	copy.Temperature = 8
	assert.Equal(t, 5, original.Temperature)
	assert.Equal(t, 8, copy.Temperature)
}

func TestGameSettings_DeepCopy(t *testing.T) {
	temp := -25
	oxygen := 5
	oceans := 2
	
	original := &model.GameSettings{
		MaxPlayers:  3,
		Temperature: &temp,
		Oxygen:      &oxygen,
		Oceans:      &oceans,
	}

	copy := original.DeepCopy()

	// Should be equal but different pointers
	assert.Equal(t, original.MaxPlayers, copy.MaxPlayers)
	assert.Equal(t, *original.Temperature, *copy.Temperature)
	assert.Equal(t, *original.Oxygen, *copy.Oxygen)
	assert.Equal(t, *original.Oceans, *copy.Oceans)
	assert.NotSame(t, original, copy)
	assert.NotSame(t, original.Temperature, copy.Temperature)
	assert.NotSame(t, original.Oxygen, copy.Oxygen)
	assert.NotSame(t, original.Oceans, copy.Oceans)

	// Modifying copy should not affect original
	*copy.Temperature = -20
	assert.Equal(t, -25, *original.Temperature)
	assert.Equal(t, -20, *copy.Temperature)
}

func TestHexPosition_DeepCopy(t *testing.T) {
	original := &model.HexPosition{
		Q: 1,
		R: -1,
		S: 0,
	}

	copy := original.DeepCopy()

	// Should be equal but different pointers
	assert.Equal(t, original.Q, copy.Q)
	assert.Equal(t, original.R, copy.R)
	assert.Equal(t, original.S, copy.S)
	assert.NotSame(t, original, copy)

	// Modifying copy should not affect original
	copy.Q = 2
	assert.Equal(t, 1, original.Q)
	assert.Equal(t, 2, copy.Q)
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