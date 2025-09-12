package repository_test

import (
	"context"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameplayLogic(t *testing.T) {
	// Initialize logger for testing
	err := logger.Init(nil)
	if err != nil {
		t.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Shutdown()

	// Initialize services
	eventBus := events.NewInMemoryEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)

	cardDataService := service.NewCardDataService()
	cardDeckRepo := repository.NewCardDeckRepository()
	cardSelectionRepo := repository.NewCardSelectionRepository()
	cardService := service.NewCardService(gameRepo, playerRepo, cardDataService, eventBus, cardDeckRepo, cardSelectionRepo)
	gameService := service.NewGameService(gameRepo, playerRepo, cardService.(*service.CardServiceImpl), eventBus)
	playerService := service.NewPlayerService(gameRepo, playerRepo)

	ctx := context.Background()

	t.Run("Test Basic Game Flow - Create and Join", func(t *testing.T) {
		// Create a game
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
		assert.NoError(t, err)
		assert.Equal(t, model.GameStatusLobby, game.Status)
		assert.Equal(t, -30, game.GlobalParameters.Temperature) // Mars starting temp
		assert.Equal(t, 0, game.GlobalParameters.Oxygen)
		assert.Equal(t, 0, game.GlobalParameters.Oceans)

		// Join players
		game, err = gameService.JoinGame(ctx, game.ID, "Alice")
		assert.NoError(t, err)
		assert.Len(t, game.PlayerIDs, 1)
		// Get player details separately using clean architecture
		players, err := playerRepo.ListByGameID(ctx, game.ID)
		assert.NoError(t, err)
		assert.Len(t, players, 1)
		assert.Equal(t, "Alice", players[0].Name)
		assert.Equal(t, 20, players[0].TerraformRating)   // Starting TR
		assert.Equal(t, 1, players[0].Production.Credits) // Base production

		game, err = gameService.JoinGame(ctx, game.ID, "Bob")
		assert.NoError(t, err)
		assert.Len(t, game.PlayerIDs, 2)
	})

	t.Run("Test Resource Management", func(t *testing.T) {
		// Create game and add player
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 2})
		assert.NoError(t, err)

		game, err = gameService.JoinGame(ctx, game.ID, "Player1")
		assert.NoError(t, err)
		playerID := game.PlayerIDs[0]

		// Test resource updates
		newResources := model.Resources{
			Credits:  42,
			Steel:    8,
			Titanium: 3,
			Plants:   15,
			Energy:   6,
			Heat:     12,
		}

		err = playerService.UpdatePlayerResources(ctx, game.ID, playerID, newResources)
		assert.NoError(t, err)

		// Verify resources updated using clean architecture
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, playerID)
		assert.NoError(t, err)
		assert.Equal(t, 42, updatedPlayer.Resources.Credits)
		assert.Equal(t, 8, updatedPlayer.Resources.Steel)
		assert.Equal(t, 3, updatedPlayer.Resources.Titanium)
		assert.Equal(t, 15, updatedPlayer.Resources.Plants)
		assert.Equal(t, 6, updatedPlayer.Resources.Energy)
		assert.Equal(t, 12, updatedPlayer.Resources.Heat)
	})

	t.Run("Test Production Management", func(t *testing.T) {
		// Create game and add player
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 2})
		assert.NoError(t, err)

		game, err = gameService.JoinGame(ctx, game.ID, "Producer")
		assert.NoError(t, err)
		playerID := game.PlayerIDs[0]

		// Test production updates
		newProduction := model.Production{
			Credits:  3, // Increased from base 1
			Steel:    2,
			Titanium: 1,
			Plants:   4,
			Energy:   3,
			Heat:     2,
		}

		err = playerService.UpdatePlayerProduction(ctx, game.ID, playerID, newProduction)
		assert.NoError(t, err)

		// Verify production updated using clean architecture
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, playerID)
		assert.NoError(t, err)
		assert.Equal(t, 3, updatedPlayer.Production.Credits)
		assert.Equal(t, 2, updatedPlayer.Production.Steel)
		assert.Equal(t, 1, updatedPlayer.Production.Titanium)
		assert.Equal(t, 4, updatedPlayer.Production.Plants)
		assert.Equal(t, 3, updatedPlayer.Production.Energy)
		assert.Equal(t, 2, updatedPlayer.Production.Heat)
	})

	t.Run("Test Terraforming Progress", func(t *testing.T) {
		// Create game
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 2})
		assert.NoError(t, err)

		// Test temperature increase
		err = gameService.IncreaseTemperature(ctx, game.ID, 3) // 3 steps = 6°C
		assert.NoError(t, err)

		updatedGame, err := gameService.GetGame(ctx, game.ID)
		assert.NoError(t, err)
		assert.Equal(t, -24, updatedGame.GlobalParameters.Temperature) // -30 + 6 = -24

		// Test oxygen increase
		err = gameService.IncreaseOxygen(ctx, game.ID, 5)
		assert.NoError(t, err)

		updatedGame, err = gameService.GetGame(ctx, game.ID)
		assert.NoError(t, err)
		assert.Equal(t, 5, updatedGame.GlobalParameters.Oxygen)

		// Test ocean placement
		err = gameService.PlaceOcean(ctx, game.ID, 2)
		assert.NoError(t, err)

		finalGame, err := gameService.GetGame(ctx, game.ID)
		assert.NoError(t, err)
		assert.Equal(t, 2, finalGame.GlobalParameters.Oceans)
	})

	t.Run("Test Terraforming Limits", func(t *testing.T) {
		// Create game
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 2})
		assert.NoError(t, err)

		// Test maximum temperature (should cap at +8°C)
		err = gameService.IncreaseTemperature(ctx, game.ID, 20) // Way more than needed
		assert.NoError(t, err)

		updatedGame, err := gameService.GetGame(ctx, game.ID)
		assert.NoError(t, err)
		assert.Equal(t, 8, updatedGame.GlobalParameters.Temperature) // Capped at +8

		// Test maximum oxygen (should cap at 14%)
		err = gameService.IncreaseOxygen(ctx, game.ID, 20) // Way more than needed
		assert.NoError(t, err)

		updatedGame, err = gameService.GetGame(ctx, game.ID)
		assert.NoError(t, err)
		assert.Equal(t, 14, updatedGame.GlobalParameters.Oxygen) // Capped at 14

		// Test maximum oceans (should cap at 9)
		err = gameService.PlaceOcean(ctx, game.ID, 15) // Way more than possible
		assert.NoError(t, err)

		finalGame, err := gameService.GetGame(ctx, game.ID)
		assert.NoError(t, err)
		assert.Equal(t, 9, finalGame.GlobalParameters.Oceans) // Capped at 9
	})

	t.Run("Test Game Capacity Limits", func(t *testing.T) {
		// Create game with max 2 players
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 2})
		assert.NoError(t, err)

		// Join 2 players
		game, err = gameService.JoinGame(ctx, game.ID, "Player1")
		assert.NoError(t, err)
		assert.Len(t, game.PlayerIDs, 1)

		game, err = gameService.JoinGame(ctx, game.ID, "Player2")
		assert.NoError(t, err)
		assert.Len(t, game.PlayerIDs, 2)

		// Try to join a third player (should fail)
		_, err = gameService.JoinGame(ctx, game.ID, "Player3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "game is full")
	})
}
