package parameters_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/parameters"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (parameters.Service, string, string) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err)

	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber, forcedActionManager).(*service.CardServiceImpl)
	lobbyService := lobby.NewService(gameRepo, playerRepo, cardRepo, cardService, cardDeckRepo, boardService, sessionManager)

	parametersRepo := parameters.NewRepository(gameRepo, playerRepo)
	parametersService := parameters.NewService(parametersRepo)

	// Create game and player
	game, err := lobbyService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	game, err = lobbyService.JoinGame(ctx, game.ID, "TestPlayer")
	require.NoError(t, err)

	playerID := game.PlayerIDs[0]

	return parametersService, game.ID, playerID
}

func TestParametersService_RaiseTemperature(t *testing.T) {
	parametersService, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Raise temperature by 2 steps (4°C)
	actualSteps, err := parametersService.RaiseTemperature(ctx, gameID, playerID, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, actualSteps)

	// Verify temperature was raised
	temp, err := parametersService.GetTemperature(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, -26, temp) // -30 + 4 = -26
}

func TestParametersService_RaiseTemperatureToMax(t *testing.T) {
	parametersService, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Try to raise temperature by 100 steps (should cap at max)
	actualSteps, err := parametersService.RaiseTemperature(ctx, gameID, playerID, 100)
	assert.NoError(t, err)
	assert.Equal(t, 19, actualSteps) // (8 - (-30)) / 2 = 19 steps

	// Verify temperature is at maximum
	temp, err := parametersService.GetTemperature(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, model.MaxTemperature, temp)

	// Verify it's maxed
	isMaxed, err := parametersService.IsTemperatureMaxed(ctx, gameID)
	assert.NoError(t, err)
	assert.True(t, isMaxed)
}

func TestParametersService_RaiseOxygen(t *testing.T) {
	parametersService, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Raise oxygen by 3 steps
	actualSteps, err := parametersService.RaiseOxygen(ctx, gameID, playerID, 3)
	assert.NoError(t, err)
	assert.Equal(t, 3, actualSteps)

	// Verify oxygen was raised
	oxygen, err := parametersService.GetOxygen(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, 3, oxygen)
}

func TestParametersService_RaiseOxygenToMax(t *testing.T) {
	parametersService, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Try to raise oxygen by 100 steps (should cap at max)
	actualSteps, err := parametersService.RaiseOxygen(ctx, gameID, playerID, 100)
	assert.NoError(t, err)
	assert.Equal(t, model.MaxOxygen, actualSteps) // Should cap at 14

	// Verify oxygen is at maximum
	oxygen, err := parametersService.GetOxygen(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, model.MaxOxygen, oxygen)

	// Verify it's maxed
	isMaxed, err := parametersService.IsOxygenMaxed(ctx, gameID)
	assert.NoError(t, err)
	assert.True(t, isMaxed)
}

func TestParametersService_PlaceOcean(t *testing.T) {
	parametersService, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Place an ocean
	err := parametersService.PlaceOcean(ctx, gameID, playerID)
	assert.NoError(t, err)

	// Verify ocean count increased
	oceans, err := parametersService.GetOceans(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, 1, oceans)
}

func TestParametersService_PlaceOceanToMax(t *testing.T) {
	parametersService, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Place 9 oceans
	for i := 0; i < model.MaxOceans; i++ {
		err := parametersService.PlaceOcean(ctx, gameID, playerID)
		assert.NoError(t, err)
	}

	// Verify ocean count is at maximum
	oceans, err := parametersService.GetOceans(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, model.MaxOceans, oceans)

	// Verify it's maxed
	isMaxed, err := parametersService.IsOceansMaxed(ctx, gameID)
	assert.NoError(t, err)
	assert.True(t, isMaxed)

	// Try to place another ocean (should fail)
	err = parametersService.PlaceOcean(ctx, gameID, playerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum oceans")
}

func TestParametersService_GetGlobalParameters(t *testing.T) {
	parametersService, gameID, _ := setupTest(t)
	ctx := context.Background()

	params, err := parametersService.GetGlobalParameters(ctx, gameID)
	assert.NoError(t, err)
	assert.Equal(t, -30, params.Temperature)
	assert.Equal(t, 0, params.Oxygen)
	assert.Equal(t, 0, params.Oceans)
}
