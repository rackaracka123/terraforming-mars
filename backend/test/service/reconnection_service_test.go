package service

import (
	"context"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test the core reconnection logic that the WebSocket hub relies on
func TestPlayerReconnection(t *testing.T) {
	// Setup
	eventBus := events.NewInMemoryEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus, playerRepo)
	parametersRepo := repository.NewGlobalParametersRepository(eventBus)

	playerService := service.NewPlayerService(gameRepo, playerRepo)
	gameService := service.NewGameService(gameRepo, playerRepo, parametersRepo)

	ctx := context.Background()

	t.Run("successful reconnection updates connection status", func(t *testing.T) {
		// Create game and add player
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
		require.NoError(t, err)

		// Add player to game (simulates initial connection)
		updatedGame, err := gameService.JoinGame(ctx, game.ID, "TestPlayer")
		require.NoError(t, err)
		require.Len(t, updatedGame.Players, 1)

		player := updatedGame.Players[0]
		assert.Equal(t, "TestPlayer", player.Name)
		assert.Equal(t, model.ConnectionStatusConnected, player.ConnectionStatus)

		// Simulate disconnection
		err = playerService.UpdatePlayerConnectionStatus(ctx, game.ID, player.ID, model.ConnectionStatusDisconnected)
		require.NoError(t, err)

		// Verify player is disconnected
		disconnectedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)
		assert.Equal(t, model.ConnectionStatusDisconnected, disconnectedGame.Players[0].ConnectionStatus)

		// Test reconnection by name (simulates WebSocket reconnection)
		reconnectedPlayer, err := playerService.GetPlayerByName(ctx, game.ID, "TestPlayer")
		require.NoError(t, err)
		assert.Equal(t, player.ID, reconnectedPlayer.ID)
		assert.Equal(t, "TestPlayer", reconnectedPlayer.Name)

		// Update connection status back to connected
		err = playerService.UpdatePlayerConnectionStatus(ctx, game.ID, player.ID, model.ConnectionStatusConnected)
		require.NoError(t, err)

		// Verify successful reconnection
		reconnectedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)
		assert.Equal(t, model.ConnectionStatusConnected, reconnectedGame.Players[0].ConnectionStatus)
	})

	t.Run("reconnection to non-existent game fails", func(t *testing.T) {
		_, err := playerService.GetPlayerByName(ctx, "non-existent-game", "TestPlayer")
		assert.Error(t, err)
	})

	t.Run("reconnection with non-existent player name fails", func(t *testing.T) {
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
		require.NoError(t, err)

		_, err = playerService.GetPlayerByName(ctx, game.ID, "NonExistentPlayer")
		assert.Error(t, err)
	})

	t.Run("reconnection preserves player state", func(t *testing.T) {
		// Create game and add player
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
		require.NoError(t, err)

		updatedGame, err := gameService.JoinGame(ctx, game.ID, "TestPlayer")
		require.NoError(t, err)

		player := &updatedGame.Players[0]

		// Modify player state (simulate gameplay)
		player.Resources.Credits = 100
		player.TerraformRating = 25
		player.Passed = false

		// Update player in repository
		err = playerRepo.UpdatePlayer(ctx, game.ID, player)
		require.NoError(t, err)

		// Simulate disconnection and reconnection
		err = playerService.UpdatePlayerConnectionStatus(ctx, game.ID, player.ID, model.ConnectionStatusDisconnected)
		require.NoError(t, err)

		// Reconnect by name
		reconnectedPlayer, err := playerService.GetPlayerByName(ctx, game.ID, "TestPlayer")
		require.NoError(t, err)

		// Verify all player state is preserved
		assert.Equal(t, player.ID, reconnectedPlayer.ID)
		assert.Equal(t, "TestPlayer", reconnectedPlayer.Name)
		assert.Equal(t, 100, reconnectedPlayer.Resources.Credits)
		assert.Equal(t, 25, reconnectedPlayer.TerraformRating)
		assert.Equal(t, false, reconnectedPlayer.Passed)
	})

	t.Run("multiple players can reconnect independently", func(t *testing.T) {
		// Create game and add multiple players
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
		require.NoError(t, err)

		game, err = gameService.JoinGame(ctx, game.ID, "Player1")
		require.NoError(t, err)
		game, err = gameService.JoinGame(ctx, game.ID, "Player2")
		require.NoError(t, err)

		require.Len(t, game.Players, 2)
		player1 := game.Players[0]
		player2 := game.Players[1]

		// Disconnect both players
		err = playerService.UpdatePlayerConnectionStatus(ctx, game.ID, player1.ID, model.ConnectionStatusDisconnected)
		require.NoError(t, err)
		err = playerService.UpdatePlayerConnectionStatus(ctx, game.ID, player2.ID, model.ConnectionStatusDisconnected)
		require.NoError(t, err)

		// Reconnect Player1 only
		reconnectedPlayer1, err := playerService.GetPlayerByName(ctx, game.ID, "Player1")
		require.NoError(t, err)
		err = playerService.UpdatePlayerConnectionStatus(ctx, game.ID, reconnectedPlayer1.ID, model.ConnectionStatusConnected)
		require.NoError(t, err)

		// Verify connection states
		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		for _, player := range updatedGame.Players {
			if player.Name == "Player1" {
				assert.Equal(t, model.ConnectionStatusConnected, player.ConnectionStatus)
			} else if player.Name == "Player2" {
				assert.Equal(t, model.ConnectionStatusDisconnected, player.ConnectionStatus)
			}
		}

		// Now reconnect Player2
		reconnectedPlayer2, err := playerService.GetPlayerByName(ctx, game.ID, "Player2")
		require.NoError(t, err)
		err = playerService.UpdatePlayerConnectionStatus(ctx, game.ID, reconnectedPlayer2.ID, model.ConnectionStatusConnected)
		require.NoError(t, err)

		// Verify both players are connected
		finalGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		for _, player := range finalGame.Players {
			assert.Equal(t, model.ConnectionStatusConnected, player.ConnectionStatus)
		}
	})
}

// Test reconnection during different game phases
func TestReconnectionDuringDifferentPhases(t *testing.T) {
	// Setup
	eventBus := events.NewInMemoryEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus, playerRepo)
	parametersRepo := repository.NewGlobalParametersRepository(eventBus)

	playerService := service.NewPlayerService(gameRepo, playerRepo)
	gameService := service.NewGameService(gameRepo, playerRepo, parametersRepo)

	ctx := context.Background()

	t.Run("reconnection during lobby phase", func(t *testing.T) {
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
		require.NoError(t, err)

		// Game should be in lobby status
		assert.Equal(t, model.GameStatusLobby, game.Status)

		// Add player
		game, err = gameService.JoinGame(ctx, game.ID, "TestPlayer")
		require.NoError(t, err)
		player := game.Players[0]

		// Disconnect and reconnect
		err = playerService.UpdatePlayerConnectionStatus(ctx, game.ID, player.ID, model.ConnectionStatusDisconnected)
		require.NoError(t, err)

		reconnectedPlayer, err := playerService.GetPlayerByName(ctx, game.ID, "TestPlayer")
		require.NoError(t, err)

		err = playerService.UpdatePlayerConnectionStatus(ctx, game.ID, reconnectedPlayer.ID, model.ConnectionStatusConnected)
		require.NoError(t, err)

		// Verify game is still in lobby and player reconnected successfully
		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)
		assert.Equal(t, model.GameStatusLobby, updatedGame.Status)
		assert.Equal(t, model.ConnectionStatusConnected, updatedGame.Players[0].ConnectionStatus)
	})

	t.Run("reconnection during active game", func(t *testing.T) {
		game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
		require.NoError(t, err)

		// Add player and start game
		game, err = gameService.JoinGame(ctx, game.ID, "TestPlayer")
		require.NoError(t, err)
		player := game.Players[0]

		err = gameService.StartGame(ctx, game.ID, player.ID)
		require.NoError(t, err)

		// Verify game is active
		activeGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)
		assert.Equal(t, model.GameStatusActive, activeGame.Status)

		// Disconnect and reconnect during active game
		err = playerService.UpdatePlayerConnectionStatus(ctx, game.ID, player.ID, model.ConnectionStatusDisconnected)
		require.NoError(t, err)

		reconnectedPlayer, err := playerService.GetPlayerByName(ctx, game.ID, "TestPlayer")
		require.NoError(t, err)

		err = playerService.UpdatePlayerConnectionStatus(ctx, game.ID, reconnectedPlayer.ID, model.ConnectionStatusConnected)
		require.NoError(t, err)

		// Verify game remains active and player reconnected successfully
		finalGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)
		assert.Equal(t, model.GameStatusActive, finalGame.Status)
		assert.Equal(t, model.ConnectionStatusConnected, finalGame.Players[0].ConnectionStatus)
	})
}
