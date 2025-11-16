package game_rules_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/parameters"
)

// Test global parameter ranges and rules from TERRAFORMING_MARS_RULES.md

func TestTemperature_Range(t *testing.T) {
	// Temperature range: -30°C to +8°C
	assert.Equal(t, -30, parameters.MinTemperature, "Minimum temperature should be -30°C")
	assert.Equal(t, 8, parameters.MaxTemperature, "Maximum temperature should be +8°C")
}

func TestTemperature_StepIncrement(t *testing.T) {
	// Each temperature step = 2°C
	eventBus := events.NewEventBus()
	repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
		Temperature: -30,
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	// Increase by 1 step
	actualSteps, err := repo.IncreaseTemperature(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1, actualSteps)

	temp, err := repo.GetTemperature(context.Background())
	require.NoError(t, err)
	assert.Equal(t, -28, temp, "1 step should increase temperature by 2°C")

	// Increase by 5 more steps
	actualSteps, err = repo.IncreaseTemperature(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, 5, actualSteps)

	temp, err = repo.GetTemperature(context.Background())
	require.NoError(t, err)
	assert.Equal(t, -18, temp, "5 more steps should increase temperature by 10°C total")
}

func TestTemperature_TotalSteps(t *testing.T) {
	// From -30°C to +8°C = 38°C total
	// 38°C / 2°C per step = 19 steps
	totalDegrees := parameters.MaxTemperature - parameters.MinTemperature
	assert.Equal(t, 38, totalDegrees)

	expectedSteps := totalDegrees / 2
	assert.Equal(t, 19, expectedSteps, "Should be 19 steps from min to max temperature")
}

func TestOxygen_Range(t *testing.T) {
	// Oxygen range: 0% to 14%
	assert.Equal(t, 0, parameters.MinOxygen, "Minimum oxygen should be 0%")
	assert.Equal(t, 14, parameters.MaxOxygen, "Maximum oxygen should be 14%")
}

func TestOxygen_StepIncrement(t *testing.T) {
	// Each oxygen step = 1%
	eventBus := events.NewEventBus()
	repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
		Temperature: 0,
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	// Increase by 1 step
	actualSteps, err := repo.IncreaseOxygen(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1, actualSteps)

	oxygen, err := repo.GetOxygen(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, oxygen, "1 step should increase oxygen by 1%")

	// Increase by 7 more steps
	actualSteps, err = repo.IncreaseOxygen(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, 7, actualSteps)

	oxygen, err = repo.GetOxygen(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 8, oxygen, "Should now be at 8%")
}

func TestOxygen_TotalSteps(t *testing.T) {
	// From 0% to 14% = 14 steps
	totalSteps := parameters.MaxOxygen - parameters.MinOxygen
	assert.Equal(t, 14, totalSteps, "Should be 14 steps from min to max oxygen")
}

func TestOceans_Range(t *testing.T) {
	// Oceans: 0 to 9
	assert.Equal(t, 0, parameters.MinOceans, "Minimum oceans should be 0")
	assert.Equal(t, 9, parameters.MaxOceans, "Maximum oceans should be 9")
}

func TestOceans_PlacementCount(t *testing.T) {
	eventBus := events.NewEventBus()
	repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
		Temperature: 0,
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	// Place 3 oceans
	actualCount, err := repo.IncreaseOceans(context.Background(), 3)
	require.NoError(t, err)
	assert.Equal(t, 3, actualCount)

	oceans, err := repo.GetOceans(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 3, oceans)

	// Place 6 more oceans
	actualCount, err = repo.IncreaseOceans(context.Background(), 6)
	require.NoError(t, err)
	assert.Equal(t, 6, actualCount)

	oceans, err = repo.GetOceans(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 9, oceans, "Should now have all 9 oceans")
}

func TestGlobalParameters_MaximumCapping(t *testing.T) {
	eventBus := events.NewEventBus()

	t.Run("Temperature caps at +8°C", func(t *testing.T) {
		repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
			Temperature: 6, // Close to max
			Oxygen:      0,
			Oceans:      0,
		}, eventBus)
		require.NoError(t, err)

		// Try to increase by 5 steps (should only increase by 1)
		actualSteps, err := repo.IncreaseTemperature(context.Background(), 5)
		require.NoError(t, err)
		assert.Equal(t, 1, actualSteps, "Only 1 step possible to reach max")

		temp, err := repo.GetTemperature(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 8, temp, "Temperature should be capped at +8°C")
	})

	t.Run("Oxygen caps at 14%", func(t *testing.T) {
		repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
			Temperature: 0,
			Oxygen:      12, // Close to max
			Oceans:      0,
		}, eventBus)
		require.NoError(t, err)

		// Try to increase by 5 steps (should only increase by 2)
		actualSteps, err := repo.IncreaseOxygen(context.Background(), 5)
		require.NoError(t, err)
		assert.Equal(t, 2, actualSteps, "Only 2 steps possible to reach max")

		oxygen, err := repo.GetOxygen(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 14, oxygen, "Oxygen should be capped at 14%")
	})

	t.Run("Oceans cap at 9", func(t *testing.T) {
		repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
			Temperature: 0,
			Oxygen:      0,
			Oceans:      7, // Close to max
		}, eventBus)
		require.NoError(t, err)

		// Try to place 5 oceans (should only place 2)
		actualCount, err := repo.IncreaseOceans(context.Background(), 5)
		require.NoError(t, err)
		assert.Equal(t, 2, actualCount, "Only 2 oceans possible to reach max")

		oceans, err := repo.GetOceans(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 9, oceans, "Oceans should be capped at 9")
	})
}

func TestGlobalParameters_AlreadyMaxed(t *testing.T) {
	eventBus := events.NewEventBus()

	t.Run("Cannot increase temperature when already maxed", func(t *testing.T) {
		repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
			Temperature: 8, // Already maxed
			Oxygen:      0,
			Oceans:      0,
		}, eventBus)
		require.NoError(t, err)

		actualSteps, err := repo.IncreaseTemperature(context.Background(), 3)
		require.NoError(t, err)
		assert.Equal(t, 0, actualSteps, "No steps should be possible")

		temp, err := repo.GetTemperature(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 8, temp)
	})

	t.Run("Cannot increase oxygen when already maxed", func(t *testing.T) {
		repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
			Temperature: 0,
			Oxygen:      14, // Already maxed
			Oceans:      0,
		}, eventBus)
		require.NoError(t, err)

		actualSteps, err := repo.IncreaseOxygen(context.Background(), 5)
		require.NoError(t, err)
		assert.Equal(t, 0, actualSteps, "No steps should be possible")

		oxygen, err := repo.GetOxygen(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 14, oxygen)
	})

	t.Run("Cannot place oceans when already maxed", func(t *testing.T) {
		repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
			Temperature: 0,
			Oxygen:      0,
			Oceans:      9, // Already maxed
		}, eventBus)
		require.NoError(t, err)

		actualCount, err := repo.IncreaseOceans(context.Background(), 3)
		require.NoError(t, err)
		assert.Equal(t, 0, actualCount, "No oceans should be placeable")

		oceans, err := repo.GetOceans(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 9, oceans)
	})
}

func TestGameEnd_AllParametersMaxed(t *testing.T) {
	// Game ends when all three global parameters reach maximum
	eventBus := events.NewEventBus()
	repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
		Temperature: 8,  // Maxed
		Oxygen:      14, // Maxed
		Oceans:      9,  // Maxed
	}, eventBus)
	require.NoError(t, err)

	params, err := repo.Get(context.Background())
	require.NoError(t, err)

	// Verify all parameters are at maximum
	assert.Equal(t, parameters.MaxTemperature, params.Temperature)
	assert.Equal(t, parameters.MaxOxygen, params.Oxygen)
	assert.Equal(t, parameters.MaxOceans, params.Oceans)

	// This would trigger game end condition
	isGameEndCondition := params.Temperature == parameters.MaxTemperature &&
		params.Oxygen == parameters.MaxOxygen &&
		params.Oceans == parameters.MaxOceans

	assert.True(t, isGameEndCondition, "All parameters maxed should trigger game end")
}
