package service_test

import (
	"context"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper functions for creating test services
func createTestStandardProjectService() service.StandardProjectService {
	eventBus := events.NewInMemoryEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus, playerRepo)
	parametersRepo := repository.NewGlobalParametersRepository(eventBus)
	globalParametersService := service.NewGlobalParametersService(gameRepo, parametersRepo)
	return service.NewStandardProjectService(gameRepo, playerRepo, parametersRepo, globalParametersService)
}

func createTestPlayerService() service.PlayerService {
	eventBus := events.NewInMemoryEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus, playerRepo)
	return service.NewPlayerService(gameRepo, playerRepo)
}

func setupStandardProjectServiceTest(t *testing.T) (
	service.StandardProjectService,
	service.GameService,
	service.PlayerService,
	service.GlobalParametersService,
	repository.PlayerRepository,
	*model.Game,
	string, // playerID
) {
	// Initialize logger for testing
	err := logger.Init(nil)
	require.NoError(t, err)

	// Initialize services
	eventBus := events.NewInMemoryEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus, playerRepo)
	parametersRepo := repository.NewGlobalParametersRepository(eventBus)

	cardService := service.NewCardService(gameRepo, playerRepo)
	gameService := service.NewGameService(gameRepo, playerRepo, parametersRepo, cardService.(*service.CardServiceImpl), eventBus)
	playerService := service.NewPlayerService(gameRepo, playerRepo)
	globalParametersService := service.NewGlobalParametersService(gameRepo, parametersRepo)
	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, parametersRepo, globalParametersService)

	ctx := context.Background()

	// Create a test game
	game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Add a test player with sufficient resources
	game, err = gameService.JoinGame(ctx, game.ID, "TestPlayer")
	require.NoError(t, err)
	require.Len(t, game.Players, 1)

	playerID := game.Players[0].ID

	// Give the player sufficient credits and cards for testing
	updatedResources := model.Resources{
		Credits:  100, // Enough for all standard projects
		Steel:    10,
		Titanium: 10,
		Plants:   10,
		Energy:   10,
		Heat:     10,
	}
	err = playerService.UpdatePlayerResources(ctx, game.ID, playerID, updatedResources)
	require.NoError(t, err)

	// Add some cards to the player's hand
	updatedGame, err := gameService.GetGame(ctx, game.ID)
	require.NoError(t, err)

	// Update player to have cards through player repository
	player, err := playerRepo.GetPlayer(ctx, game.ID, playerID)
	require.NoError(t, err)

	player.Cards = []string{"card1", "card2", "card3", "card4", "card5"}
	err = playerRepo.UpdatePlayer(ctx, game.ID, player)
	require.NoError(t, err)

	// Also update the game state to reflect the cards
	for i := range updatedGame.Players {
		if updatedGame.Players[i].ID == playerID {
			updatedGame.Players[i].Cards = []string{"card1", "card2", "card3", "card4", "card5"}
			break
		}
	}
	// Update the game directly through repository instead of removed UpdateGame method
	err = gameRepo.Update(ctx, updatedGame)
	require.NoError(t, err)

	return standardProjectService, gameService, playerService, globalParametersService, playerRepo, updatedGame, playerID
}

func TestStandardProjectService_SellPatents(t *testing.T) {
	standardProjectService, gameService, _, _, _, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	t.Run("Successful sell patents", func(t *testing.T) {
		initialCredits := 100
		cardsToSell := 3

		// Execute sell patents
		err := standardProjectService.SellPatents(ctx, game.ID, playerID, cardsToSell)
		assert.NoError(t, err)

		// Verify player resources and cards
		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		player := updatedGame.Players[0]
		assert.Equal(t, initialCredits+cardsToSell, player.Resources.Credits)
		assert.Equal(t, 2, len(player.Cards)) // 5 - 3 = 2 cards remaining
	})

	t.Run("Cannot sell more cards than available", func(t *testing.T) {
		err := standardProjectService.SellPatents(ctx, game.ID, playerID, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot sell")
	})

	t.Run("Cannot sell zero or negative cards", func(t *testing.T) {
		err := standardProjectService.SellPatents(ctx, game.ID, playerID, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot sell")
	})
}

func TestStandardProjectService_BuildPowerPlant(t *testing.T) {
	standardProjectService, gameService, playerService, _, _, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	t.Run("Successful build power plant", func(t *testing.T) {
		initialCredits := 100
		initialEnergyProduction := 0
		expectedCost := 11

		// Execute build power plant
		err := standardProjectService.BuildPowerPlant(ctx, game.ID, playerID)
		assert.NoError(t, err)

		// Verify player resources and production
		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		player := updatedGame.Players[0]
		assert.Equal(t, initialCredits-expectedCost, player.Resources.Credits)
		assert.Equal(t, initialEnergyProduction+1, player.Production.Energy)
	})

	t.Run("Insufficient credits", func(t *testing.T) {
		// Set player credits to less than cost
		insufficientResources := model.Resources{Credits: 5}
		err := playerService.UpdatePlayerResources(ctx, game.ID, playerID, insufficientResources)
		require.NoError(t, err)

		err = standardProjectService.BuildPowerPlant(ctx, game.ID, playerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient credits")
	})
}

func TestStandardProjectService_LaunchAsteroid(t *testing.T) {
	standardProjectService, gameService, _, globalParametersService, _, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	t.Run("Successful launch asteroid", func(t *testing.T) {
		initialCredits := 100
		initialTR := 20
		expectedCost := 14

		// Get initial temperature
		initialParams, err := globalParametersService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		initialTemp := initialParams.Temperature

		// Execute launch asteroid
		err = standardProjectService.LaunchAsteroid(ctx, game.ID, playerID)
		assert.NoError(t, err)

		// Verify player resources and TR
		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		player := updatedGame.Players[0]
		assert.Equal(t, initialCredits-expectedCost, player.Resources.Credits)
		assert.Equal(t, initialTR+1, player.TerraformRating)

		// Verify temperature increase
		updatedParams, err := globalParametersService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		assert.Equal(t, initialTemp+2, updatedParams.Temperature) // Each step = 2°C
	})
}

func TestStandardProjectService_BuildAquifer(t *testing.T) {
	standardProjectService, gameService, _, globalParametersService, _, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	validHexPosition := model.HexPosition{Q: 1, R: -1, S: 0}

	t.Run("Successful build aquifer", func(t *testing.T) {
		initialCredits := 100
		initialTR := 20
		expectedCost := 18

		// Get initial ocean count
		initialParams, err := globalParametersService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		initialOceans := initialParams.Oceans

		// Execute build aquifer
		err = standardProjectService.BuildAquifer(ctx, game.ID, playerID, validHexPosition)
		assert.NoError(t, err)

		// Verify player resources and TR
		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		player := updatedGame.Players[0]
		assert.Equal(t, initialCredits-expectedCost, player.Resources.Credits)
		assert.Equal(t, initialTR+1, player.TerraformRating)

		// Verify ocean count increase
		updatedParams, err := globalParametersService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		assert.Equal(t, initialOceans+1, updatedParams.Oceans)
	})

	t.Run("Invalid hex position", func(t *testing.T) {
		invalidHexPosition := model.HexPosition{Q: 1, R: 1, S: 1} // Sum != 0

		err := standardProjectService.BuildAquifer(ctx, game.ID, playerID, invalidHexPosition)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid hex position")
	})
}

func TestStandardProjectService_PlantGreenery(t *testing.T) {
	standardProjectService, gameService, _, globalParametersService, _, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	validHexPosition := model.HexPosition{Q: 2, R: -1, S: -1}

	t.Run("Successful plant greenery", func(t *testing.T) {
		initialCredits := 100
		initialTR := 20
		expectedCost := 23

		// Get initial oxygen level
		initialParams, err := globalParametersService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		initialOxygen := initialParams.Oxygen

		// Execute plant greenery
		err = standardProjectService.PlantGreenery(ctx, game.ID, playerID, validHexPosition)
		assert.NoError(t, err)

		// Verify player resources and TR
		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		player := updatedGame.Players[0]
		assert.Equal(t, initialCredits-expectedCost, player.Resources.Credits)
		assert.Equal(t, initialTR+1, player.TerraformRating)

		// Verify oxygen level increase
		updatedParams, err := globalParametersService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		assert.Equal(t, initialOxygen+1, updatedParams.Oxygen)
	})

	t.Run("Invalid hex position", func(t *testing.T) {
		invalidHexPosition := model.HexPosition{Q: 1, R: 2, S: 3} // Sum != 0

		err := standardProjectService.PlantGreenery(ctx, game.ID, playerID, invalidHexPosition)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid hex position")
	})
}

func TestStandardProjectService_BuildCity(t *testing.T) {
	standardProjectService, gameService, _, _, _, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	validHexPosition := model.HexPosition{Q: -2, R: 1, S: 1}

	t.Run("Successful build city", func(t *testing.T) {
		initialCredits := 100
		initialCreditProduction := 1
		expectedCost := 25

		// Execute build city
		err := standardProjectService.BuildCity(ctx, game.ID, playerID, validHexPosition)
		assert.NoError(t, err)

		// Verify player resources and production
		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		player := updatedGame.Players[0]
		assert.Equal(t, initialCredits-expectedCost, player.Resources.Credits)
		assert.Equal(t, initialCreditProduction+1, player.Production.Credits)
	})

	t.Run("Invalid hex position", func(t *testing.T) {
		invalidHexPosition := model.HexPosition{Q: 0, R: 0, S: 1} // Sum != 0

		err := standardProjectService.BuildCity(ctx, game.ID, playerID, invalidHexPosition)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid hex position")
	})
}

func TestHexPosition_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		position model.HexPosition
		expected bool
	}{
		{"Valid position (0,0,0)", model.HexPosition{Q: 0, R: 0, S: 0}, true},
		{"Valid position (1,-1,0)", model.HexPosition{Q: 1, R: -1, S: 0}, true},
		{"Valid position (-2,1,1)", model.HexPosition{Q: -2, R: 1, S: 1}, true},
		{"Invalid position (1,1,1)", model.HexPosition{Q: 1, R: 1, S: 1}, false},
		{"Invalid position (1,0,0)", model.HexPosition{Q: 1, R: 0, S: 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := createTestStandardProjectService()
			result := svc.IsValidHexPosition(&tt.position)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlayer_CanAffordStandardProject(t *testing.T) {
	player := &model.Player{
		Resources: model.Resources{Credits: 20},
	}

	tests := []struct {
		name     string
		project  model.StandardProject
		expected bool
	}{
		{"Can afford sell patents", model.StandardProjectSellPatents, true}, // 0 cost
		{"Can afford power plant", model.StandardProjectPowerPlant, true},   // 11 cost
		{"Can afford asteroid", model.StandardProjectAsteroid, true},        // 14 cost
		{"Can afford aquifer", model.StandardProjectAquifer, true},          // 18 cost
		{"Cannot afford greenery", model.StandardProjectGreenery, false},    // 23 cost
		{"Cannot afford city", model.StandardProjectCity, false},            // 25 cost
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := createTestPlayerService()
			result := svc.CanAffordStandardProject(player, tt.project)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlayer_HasCardsToSell(t *testing.T) {
	player := &model.Player{
		Cards: []string{"card1", "card2", "card3"},
	}

	tests := []struct {
		name     string
		count    int
		expected bool
	}{
		{"Can sell 1 card", 1, true},
		{"Can sell 3 cards", 3, true},
		{"Cannot sell 4 cards", 4, false},
		{"Cannot sell 0 cards", 0, false},
		{"Cannot sell negative cards", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := createTestPlayerService()
			result := svc.HasCardsToSell(player, tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlayer_GetMaxCardsToSell(t *testing.T) {
	tests := []struct {
		name     string
		cards    []string
		expected int
	}{
		{"No cards", []string{}, 0},
		{"Three cards", []string{"card1", "card2", "card3"}, 3},
		{"Many cards", []string{"c1", "c2", "c3", "c4", "c5", "c6", "c7"}, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &model.Player{Cards: tt.cards}
			svc := createTestPlayerService()
			result := svc.GetMaxCardsToSell(player)
			assert.Equal(t, tt.expected, result)
		})
	}
}
