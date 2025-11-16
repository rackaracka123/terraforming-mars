package actions_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"terraforming-mars-backend/internal/actions"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/test/fixtures"
)

// Mock session manager for testing
type mockSessionManager struct{}

func (m *mockSessionManager) Broadcast(gameID string) error {
	return nil
}

func (m *mockSessionManager) Send(gameID string, playerID string) error {
	return nil
}

func TestConvertHeatToTemperature_Success(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()
	gameID := "test-game"
	playerID := "player-1"

	// Setup repositories
	playerRepo := player.NewRepository(eventBus)
	paramsRepo, err := parameters.NewRepository(gameID, parameters.GlobalParameters{
		Temperature: -30, // Starting temperature
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	// Create service
	paramsService := parameters.NewService(paramsRepo)

	// Create player with heat
	testPlayer := fixtures.NewTestPlayer(
		fixtures.WithID(playerID),
		fixtures.WithTR(20),
		fixtures.WithHeat(8), // Exactly enough for conversion
	)
	err = playerRepo.Create(ctx, gameID, *testPlayer)
	require.NoError(t, err)

	// Create action
	action := actions.NewConvertHeatToTemperatureAction(
		playerRepo,
		paramsService,
		&mockSessionManager{},
	)

	// Execute conversion
	err = action.Execute(ctx, gameID, playerID)
	assert.NoError(t, err)

	// Verify heat was deducted
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Equal(t, 0, updatedPlayer.Resources.Heat, "Heat should be depleted")

	// Verify temperature was raised
	temp, err := paramsService.GetTemperature(ctx)
	require.NoError(t, err)
	assert.Equal(t, -28, temp, "Temperature should increase by 2°C (1 step)")

	// Verify TR was awarded
	assert.Equal(t, 21, updatedPlayer.TerraformRating, "TR should increase by 1")
}

func TestConvertHeatToTemperature_InsufficientHeat(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()
	gameID := "test-game"
	playerID := "player-1"

	// Setup repositories
	playerRepo := player.NewRepository(eventBus)
	paramsRepo, err := parameters.NewRepository(gameID, parameters.GlobalParameters{
		Temperature: -30,
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	paramsService := parameters.NewService(paramsRepo)

	// Create player with insufficient heat
	testPlayer := fixtures.NewTestPlayer(
		fixtures.WithID(playerID),
		fixtures.WithTR(20),
		fixtures.WithHeat(7), // Not enough (need 8)
	)
	err = playerRepo.Create(ctx, gameID, *testPlayer)
	require.NoError(t, err)

	// Create action
	action := actions.NewConvertHeatToTemperatureAction(
		playerRepo,
		paramsService,
		&mockSessionManager{},
	)

	// Execute conversion - should fail
	err = action.Execute(ctx, gameID, playerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient heat")

	// Verify heat was NOT deducted
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Equal(t, 7, updatedPlayer.Resources.Heat, "Heat should remain unchanged")

	// Verify temperature was NOT raised
	temp, err := paramsService.GetTemperature(ctx)
	require.NoError(t, err)
	assert.Equal(t, -30, temp, "Temperature should remain unchanged")

	// Verify TR was NOT awarded
	assert.Equal(t, 20, updatedPlayer.TerraformRating, "TR should remain unchanged")
}

func TestConvertHeatToTemperature_TemperatureAlreadyMaxed(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()
	gameID := "test-game"
	playerID := "player-1"

	// Setup repositories
	playerRepo := player.NewRepository(eventBus)
	paramsRepo, err := parameters.NewRepository(gameID, parameters.GlobalParameters{
		Temperature: 8, // Already at maximum
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	paramsService := parameters.NewService(paramsRepo)

	// Create player with heat
	testPlayer := fixtures.NewTestPlayer(
		fixtures.WithID(playerID),
		fixtures.WithTR(20),
		fixtures.WithHeat(8),
	)
	err = playerRepo.Create(ctx, gameID, *testPlayer)
	require.NoError(t, err)

	// Create action
	action := actions.NewConvertHeatToTemperatureAction(
		playerRepo,
		paramsService,
		&mockSessionManager{},
	)

	// Execute conversion
	err = action.Execute(ctx, gameID, playerID)
	assert.NoError(t, err, "Action should succeed even when temperature is maxed")

	// Verify heat was deducted (action still costs heat)
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Equal(t, 0, updatedPlayer.Resources.Heat, "Heat should be depleted")

	// Verify temperature remains at max
	temp, err := paramsService.GetTemperature(ctx)
	require.NoError(t, err)
	assert.Equal(t, 8, temp, "Temperature should remain at maximum")

	// Verify TR was NOT awarded (no temperature change)
	assert.Equal(t, 20, updatedPlayer.TerraformRating, "TR should not increase when temperature already maxed")
}

func TestConvertHeatToTemperature_MultipleConversions(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()
	gameID := "test-game"
	playerID := "player-1"

	// Setup repositories
	playerRepo := player.NewRepository(eventBus)
	paramsRepo, err := parameters.NewRepository(gameID, parameters.GlobalParameters{
		Temperature: -30,
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	paramsService := parameters.NewService(paramsRepo)

	// Create player with enough heat for 3 conversions
	testPlayer := fixtures.NewTestPlayer(
		fixtures.WithID(playerID),
		fixtures.WithTR(20),
		fixtures.WithHeat(24), // 3 x 8 heat
	)
	err = playerRepo.Create(ctx, gameID, *testPlayer)
	require.NoError(t, err)

	// Create action
	action := actions.NewConvertHeatToTemperatureAction(
		playerRepo,
		paramsService,
		&mockSessionManager{},
	)

	// First conversion
	err = action.Execute(ctx, gameID, playerID)
	assert.NoError(t, err)

	temp, _ := paramsService.GetTemperature(ctx)
	assert.Equal(t, -28, temp)

	p, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 16, p.Resources.Heat)
	assert.Equal(t, 21, p.TerraformRating)

	// Second conversion
	err = action.Execute(ctx, gameID, playerID)
	assert.NoError(t, err)

	temp, _ = paramsService.GetTemperature(ctx)
	assert.Equal(t, -26, temp)

	p, _ = playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 8, p.Resources.Heat)
	assert.Equal(t, 22, p.TerraformRating)

	// Third conversion
	err = action.Execute(ctx, gameID, playerID)
	assert.NoError(t, err)

	temp, _ = paramsService.GetTemperature(ctx)
	assert.Equal(t, -24, temp)

	p, _ = playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 0, p.Resources.Heat)
	assert.Equal(t, 23, p.TerraformRating)
}

func TestConvertHeatToTemperature_TRAwarding(t *testing.T) {
	// Verify that +1 TR is awarded per temperature step
	ctx := context.Background()
	eventBus := events.NewEventBus()
	gameID := "test-game"
	playerID := "player-1"

	playerRepo := player.NewRepository(eventBus)
	paramsRepo, err := parameters.NewRepository(gameID, parameters.GlobalParameters{
		Temperature: -10,
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	paramsService := parameters.NewService(paramsRepo)

	testPlayer := fixtures.NewTestPlayer(
		fixtures.WithID(playerID),
		fixtures.WithTR(25), // Start at TR 25
		fixtures.WithHeat(8),
	)
	err = playerRepo.Create(ctx, gameID, *testPlayer)
	require.NoError(t, err)

	action := actions.NewConvertHeatToTemperatureAction(
		playerRepo,
		paramsService,
		&mockSessionManager{},
	)

	// Execute conversion
	err = action.Execute(ctx, gameID, playerID)
	assert.NoError(t, err)

	// Verify TR increased by exactly 1
	p, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Equal(t, 26, p.TerraformRating, "TR should increase by exactly 1 per temperature step")
}

func TestConvertHeatToTemperature_GameRuleCompliance(t *testing.T) {
	// Verify compliance with TERRAFORMING_MARS_RULES.md:
	// - Cost: 8 heat per step
	// - Effect: +1 temperature → +1 TR
	// - Temperature: Each step = 2°C

	ctx := context.Background()
	eventBus := events.NewEventBus()
	gameID := "test-game"
	playerID := "player-1"

	playerRepo := player.NewRepository(eventBus)
	paramsRepo, err := parameters.NewRepository(gameID, parameters.GlobalParameters{
		Temperature: 0,
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	paramsService := parameters.NewService(paramsRepo)

	testPlayer := fixtures.NewTestPlayer(
		fixtures.WithID(playerID),
		fixtures.WithTR(20),
		fixtures.WithHeat(8),
	)
	err = playerRepo.Create(ctx, gameID, *testPlayer)
	require.NoError(t, err)

	action := actions.NewConvertHeatToTemperatureAction(
		playerRepo,
		paramsService,
		&mockSessionManager{},
	)

	// Execute
	err = action.Execute(ctx, gameID, playerID)
	assert.NoError(t, err)

	// Verify cost (8 heat)
	cost := domain.StandardProjectCosts.ConvertHeatToTemperature
	assert.Equal(t, 8, cost.Heat, "Cost should be 8 heat")

	p, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 0, p.Resources.Heat, "8 heat should be deducted")

	// Verify effect (+2°C)
	temp, _ := paramsService.GetTemperature(ctx)
	assert.Equal(t, 2, temp, "Temperature should increase by 2°C")

	// Verify TR award (+1 TR)
	assert.Equal(t, 21, p.TerraformRating, "TR should increase by 1")
}
