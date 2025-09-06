package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalParameters_InitialState(t *testing.T) {
	params := GlobalParameters{}

	// Test zero values
	assert.Equal(t, 0, params.Temperature)
	assert.Equal(t, 0, params.Oxygen)
	assert.Equal(t, 0, params.Oceans)
}

func TestGlobalParameters_MarsStartingConditions(t *testing.T) {
	params := GlobalParameters{
		Temperature: -30,
		Oxygen:      0,
		Oceans:      0,
	}

	assert.Equal(t, -30, params.Temperature)
	assert.Equal(t, 0, params.Oxygen)
	assert.Equal(t, 0, params.Oceans)
}

func TestGlobalParameters_CanIncreaseTemperature(t *testing.T) {
	tests := []struct {
		name      string
		current   int
		increase  int
		expected  bool
		finalTemp int
	}{
		{"Can increase from minimum", -30, 5, true, -20},
		{"Can increase to maximum", 6, 1, true, 8},
		{"Cannot exceed maximum", 8, 1, false, 8},
		{"Already at maximum", 8, 0, true, 8},
		{"Large increase capped at maximum", -30, 50, true, 8},
		{"Zero increase", -20, 0, true, -20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &GlobalParameters{Temperature: tt.current}
			result := params.CanIncreaseTemperature(tt.increase)

			if tt.expected {
				assert.True(t, result)
				params.IncreaseTemperature(tt.increase)
				assert.Equal(t, tt.finalTemp, params.Temperature)
			} else {
				assert.False(t, result)
			}
		})
	}
}

func TestGlobalParameters_CanIncreaseOxygen(t *testing.T) {
	tests := []struct {
		name        string
		current     int
		increase    int
		expected    bool
		finalOxygen int
	}{
		{"Can increase from zero", 0, 5, true, 5},
		{"Can increase to maximum", 13, 1, true, 14},
		{"Cannot exceed maximum", 14, 1, false, 14},
		{"Already at maximum", 14, 0, true, 14},
		{"Large increase capped at maximum", 0, 20, true, 14},
		{"Zero increase", 10, 0, true, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &GlobalParameters{Oxygen: tt.current}
			result := params.CanIncreaseOxygen(tt.increase)

			if tt.expected {
				assert.True(t, result)
				params.IncreaseOxygen(tt.increase)
				assert.Equal(t, tt.finalOxygen, params.Oxygen)
			} else {
				assert.False(t, result)
			}
		})
	}
}

func TestGlobalParameters_CanPlaceOcean(t *testing.T) {
	tests := []struct {
		name        string
		current     int
		place       int
		expected    bool
		finalOceans int
	}{
		{"Can place from zero", 0, 3, true, 3},
		{"Can place to maximum", 8, 1, true, 9},
		{"Cannot exceed maximum", 9, 1, false, 9},
		{"Already at maximum", 9, 0, true, 9},
		{"Large placement capped at maximum", 0, 15, true, 9},
		{"Zero placement", 5, 0, true, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &GlobalParameters{Oceans: tt.current}
			result := params.CanPlaceOcean(tt.place)

			if tt.expected {
				assert.True(t, result)
				params.PlaceOcean(tt.place)
				assert.Equal(t, tt.finalOceans, params.Oceans)
			} else {
				assert.False(t, result)
			}
		})
	}
}

func TestGlobalParameters_IsFullyTerraformed(t *testing.T) {
	tests := []struct {
		name        string
		temperature int
		oxygen      int
		oceans      int
		expected    bool
	}{
		{"Not terraformed - all minimum", -30, 0, 0, false},
		{"Not terraformed - temperature not max", 6, 14, 9, false},
		{"Not terraformed - oxygen not max", 8, 13, 9, false},
		{"Not terraformed - oceans not max", 8, 14, 8, false},
		{"Fully terraformed", 8, 14, 9, true},
		{"Partially terraformed", -10, 8, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := GlobalParameters{
				Temperature: tt.temperature,
				Oxygen:      tt.oxygen,
				Oceans:      tt.oceans,
			}

			result := params.IsFullyTerraformed()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGlobalParameters_GetTerraformingProgress(t *testing.T) {
	tests := []struct {
		name        string
		temperature int
		oxygen      int
		oceans      int
		expected    float64
	}{
		{"No progress", -30, 0, 0, 0.0},
		{"Full progress", 8, 14, 9, 100.0},
		{"Near half progress", -11, 7, 4, 48.15},
		{"Temperature only", 8, 0, 0, 100.0 / 3},
		{"Oxygen only", -30, 14, 0, 100.0 / 3},
		{"Oceans only", -30, 0, 9, 100.0 / 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := GlobalParameters{
				Temperature: tt.temperature,
				Oxygen:      tt.oxygen,
				Oceans:      tt.oceans,
			}

			result := params.GetTerraformingProgress()
			assert.InDelta(t, tt.expected, result, 0.1) // Allow small floating point differences
		})
	}
}

func TestHexPosition_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		position HexPosition
		expected bool
	}{
		{"Valid center", HexPosition{Q: 0, R: 0, S: 0}, true},
		{"Valid positive Q", HexPosition{Q: 1, R: -1, S: 0}, true},
		{"Valid negative Q", HexPosition{Q: -1, R: 1, S: 0}, true},
		{"Valid positive R", HexPosition{Q: 0, R: 1, S: -1}, true},
		{"Valid negative R", HexPosition{Q: 0, R: -1, S: 1}, true},
		{"Valid positive S", HexPosition{Q: -1, R: 0, S: 1}, true},
		{"Valid negative S", HexPosition{Q: 1, R: 0, S: -1}, true},
		{"Valid complex", HexPosition{Q: 2, R: -1, S: -1}, true},
		{"Invalid sum positive", HexPosition{Q: 1, R: 1, S: 1}, false},
		{"Invalid sum negative", HexPosition{Q: -1, R: -1, S: -1}, false},
		{"Invalid sum two", HexPosition{Q: 2, R: 1, S: 1}, false},
		{"Invalid partial", HexPosition{Q: 1, R: 0, S: 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.position.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHexPosition_Distance(t *testing.T) {
	tests := []struct {
		name     string
		pos1     HexPosition
		pos2     HexPosition
		expected int
	}{
		{"Same position", HexPosition{0, 0, 0}, HexPosition{0, 0, 0}, 0},
		{"Adjacent horizontal", HexPosition{0, 0, 0}, HexPosition{1, -1, 0}, 1},
		{"Adjacent vertical", HexPosition{0, 0, 0}, HexPosition{0, 1, -1}, 1},
		{"Distance 2", HexPosition{0, 0, 0}, HexPosition{2, -1, -1}, 2},
		{"Distance 3", HexPosition{0, 0, 0}, HexPosition{-2, 1, 1}, 2},
		{"Complex distance", HexPosition{1, -1, 0}, HexPosition{-1, 2, -1}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pos1.Distance(tt.pos2)
			assert.Equal(t, tt.expected, result)

			// Distance should be symmetric
			reverseResult := tt.pos2.Distance(tt.pos1)
			assert.Equal(t, result, reverseResult)
		})
	}
}

func TestHexPosition_GetNeighbors(t *testing.T) {
	center := HexPosition{0, 0, 0}
	neighbors := center.GetNeighbors()

	expectedNeighbors := []HexPosition{
		{1, -1, 0}, // East
		{1, 0, -1}, // Southeast
		{0, 1, -1}, // Southwest
		{-1, 1, 0}, // West
		{-1, 0, 1}, // Northwest
		{0, -1, 1}, // Northeast
	}

	assert.Len(t, neighbors, 6)

	// Check that all expected neighbors are present
	for _, expected := range expectedNeighbors {
		found := false
		for _, neighbor := range neighbors {
			if neighbor == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected neighbor %+v not found", expected)
	}

	// Check that all neighbors are valid hex positions
	for _, neighbor := range neighbors {
		assert.True(t, neighbor.IsValid())
	}

	// Check that all neighbors are distance 1 from center
	for _, neighbor := range neighbors {
		assert.Equal(t, 1, center.Distance(neighbor))
	}
}
