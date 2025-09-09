package repository_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimplifiedRepositoryPattern(t *testing.T) {
	// Initialize logger for testing
	err := logger.Init(nil)
	if err != nil {
		t.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Shutdown()

	// Initialize repositories with clean architecture
	eventBus := events.NewInMemoryEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	ctx := context.Background()

	t.Run("Game Repository CRUD Operations", func(t *testing.T) {
		// Create a game
		game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
		require.NoError(t, err)
		assert.Equal(t, model.GameStatusLobby, game.Status)
		assert.Equal(t, 4, game.Settings.MaxPlayers)
		assert.Empty(t, game.PlayerIDs) // Should start with no players
		assert.Equal(t, -30, game.GlobalParameters.Temperature)

		gameID := game.ID

		// Get the game back
		retrievedGame, err := gameRepo.GetByID(ctx, gameID)
		require.NoError(t, err)
		assert.Equal(t, game.ID, retrievedGame.ID)
		assert.Equal(t, game.Status, retrievedGame.Status)

		// Update game status
		err = gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
		require.NoError(t, err)

		// Verify status was updated
		updatedGame, err := gameRepo.GetByID(ctx, gameID)
		require.NoError(t, err)
		assert.Equal(t, model.GameStatusActive, updatedGame.Status)

		// Add a player ID to the game
		err = gameRepo.AddPlayerID(ctx, gameID, "player-1")
		require.NoError(t, err)

		// Verify player ID was added
		gameWithPlayer, err := gameRepo.GetByID(ctx, gameID)
		require.NoError(t, err)
		assert.Contains(t, gameWithPlayer.PlayerIDs, "player-1")
		assert.Equal(t, "player-1", gameWithPlayer.HostPlayerID) // First player becomes host

		// List games
		games, err := gameRepo.List(ctx, "")
		require.NoError(t, err)
		assert.Len(t, games, 1)
		assert.Equal(t, gameID, games[0].ID)

		// Delete the game
		err = gameRepo.Delete(ctx, gameID)
		require.NoError(t, err)

		// Verify game is deleted
		_, err = gameRepo.GetByID(ctx, gameID)
		assert.Error(t, err)
	})

	t.Run("Player Repository CRUD Operations", func(t *testing.T) {
		gameID := "test-game-1"
		player := model.Player{
			ID:              "player-1",
			Name:            "Alice",
			TerraformRating: 20,
			Resources: model.Resources{
				Credits: 45,
			},
			Production: model.Production{
				Credits: 1,
			},
			ConnectionStatus: model.ConnectionStatusConnected,
		}

		// Create a player
		err := playerRepo.Create(ctx, gameID, player)
		require.NoError(t, err)

		// Get the player back
		retrievedPlayer, err := playerRepo.GetByID(ctx, gameID, "player-1")
		require.NoError(t, err)
		assert.Equal(t, "Alice", retrievedPlayer.Name)
		assert.Equal(t, 20, retrievedPlayer.TerraformRating)
		assert.Equal(t, 45, retrievedPlayer.Resources.Credits)

		// Update player resources
		newResources := model.Resources{Credits: 50, Steel: 5}
		err = playerRepo.UpdateResources(ctx, gameID, "player-1", newResources)
		require.NoError(t, err)

		// Verify resources were updated
		updatedPlayer, err := playerRepo.GetByID(ctx, gameID, "player-1")
		require.NoError(t, err)
		assert.Equal(t, 50, updatedPlayer.Resources.Credits)
		assert.Equal(t, 5, updatedPlayer.Resources.Steel)

		// Update terraform rating
		err = playerRepo.UpdateTerraformRating(ctx, gameID, "player-1", 25)
		require.NoError(t, err)

		// Verify TR was updated
		playerWithNewTR, err := playerRepo.GetByID(ctx, gameID, "player-1")
		require.NoError(t, err)
		assert.Equal(t, 25, playerWithNewTR.TerraformRating)

		// List players
		players, err := playerRepo.ListByGameID(ctx, gameID)
		require.NoError(t, err)
		assert.Len(t, players, 1)
		assert.Equal(t, "Alice", players[0].Name)

		// Delete player
		err = playerRepo.Delete(ctx, gameID, "player-1")
		require.NoError(t, err)

		// Verify player is deleted
		_, err = playerRepo.GetByID(ctx, gameID, "player-1")
		assert.Error(t, err)
	})

	t.Run("Global Parameters Update", func(t *testing.T) {
		// Create a game for testing global parameters
		game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 2})
		require.NoError(t, err)

		gameID := game.ID

		// Update global parameters
		newParams := model.GlobalParameters{
			Temperature: -24,
			Oxygen:      2,
			Oceans:      1,
		}
		err = gameRepo.UpdateGlobalParameters(ctx, gameID, newParams)
		require.NoError(t, err)

		// Verify parameters were updated
		updatedGame, err := gameRepo.GetByID(ctx, gameID)
		require.NoError(t, err)
		assert.Equal(t, -24, updatedGame.GlobalParameters.Temperature)
		assert.Equal(t, 2, updatedGame.GlobalParameters.Oxygen)
		assert.Equal(t, 1, updatedGame.GlobalParameters.Oceans)

		// Clean up
		err = gameRepo.Delete(ctx, gameID)
		require.NoError(t, err)
	})
}
