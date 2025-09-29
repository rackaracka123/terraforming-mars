package repository_test

import (
	"context"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"
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

	// Initialize repositories
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()

	cardRepo := repository.NewCardRepository()
	// Load card data for testing
	err = cardRepo.LoadCards(context.Background())
	if err != nil {
		t.Fatal("Failed to load card data:", err)
	}

	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)
	boardService := service.NewBoardService()
	gameService := service.NewGameService(gameRepo, playerRepo, cardRepo, cardService.(*service.CardServiceImpl), cardDeckRepo, boardService, sessionManager)
	_ = service.NewPlayerService(gameRepo, playerRepo, nil)

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

		err = playerRepo.UpdateResources(ctx, game.ID, playerID, newResources)
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

		err = playerRepo.UpdateProduction(ctx, game.ID, playerID, newProduction)
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
