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

func setupPlayerServiceTest(t *testing.T) (
	service.PlayerService,
	service.GameService,
	model.Game,
) {
	eventBus := events.NewInMemoryEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	playerService := service.NewPlayerService(gameRepo, playerRepo)

	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	cardSelectionRepo := repository.NewCardSelectionRepository()
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, eventBus, cardDeckRepo, cardSelectionRepo)
	gameService := service.NewGameService(gameRepo, playerRepo, cardService.(*service.CardServiceImpl), eventBus)

	ctx := context.Background()
	game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	game, err = gameService.JoinGame(ctx, game.ID, "TestPlayer")
	require.NoError(t, err)

	return playerService, gameService, game
}

func TestPlayerService_UpdatePlayerResources(t *testing.T) {
	playerService, _, game := setupPlayerServiceTest(t)
	ctx := context.Background()
	playerID := game.PlayerIDs[0]

	t.Run("Valid resource update", func(t *testing.T) {
		newResources := model.Resources{
			Credits:  50,
			Steel:    10,
			Titanium: 5,
			Plants:   20,
			Energy:   15,
			Heat:     8,
		}

		err := playerService.UpdatePlayerResources(ctx, game.ID, playerID, newResources)
		assert.NoError(t, err)

		updatedPlayer, err := playerService.GetPlayer(ctx, game.ID, playerID)
		require.NoError(t, err)

		player := updatedPlayer
		assert.Equal(t, 50, player.Resources.Credits)
		assert.Equal(t, 10, player.Resources.Steel)
		assert.Equal(t, 5, player.Resources.Titanium)
		assert.Equal(t, 20, player.Resources.Plants)
		assert.Equal(t, 15, player.Resources.Energy)
		assert.Equal(t, 8, player.Resources.Heat)
	})

	t.Run("Invalid game ID", func(t *testing.T) {
		newResources := model.Resources{Credits: 30}
		err := playerService.UpdatePlayerResources(ctx, "invalid-game-id", playerID, newResources)
		assert.Error(t, err)
	})

	t.Run("Invalid player ID", func(t *testing.T) {
		newResources := model.Resources{Credits: 30}
		err := playerService.UpdatePlayerResources(ctx, game.ID, "invalid-player-id", newResources)
		assert.Error(t, err)
	})

	t.Run("Zero resources", func(t *testing.T) {
		newResources := model.Resources{} // All zeros
		err := playerService.UpdatePlayerResources(ctx, game.ID, playerID, newResources)
		assert.NoError(t, err)

		updatedPlayer, err := playerService.GetPlayer(ctx, game.ID, playerID)
		require.NoError(t, err)

		player := updatedPlayer
		assert.Equal(t, 0, player.Resources.Credits)
		assert.Equal(t, 0, player.Resources.Steel)
		assert.Equal(t, 0, player.Resources.Titanium)
		assert.Equal(t, 0, player.Resources.Plants)
		assert.Equal(t, 0, player.Resources.Energy)
		assert.Equal(t, 0, player.Resources.Heat)
	})
}

func TestPlayerService_UpdatePlayerProduction(t *testing.T) {
	playerService, _, game := setupPlayerServiceTest(t)
	ctx := context.Background()
	playerID := game.PlayerIDs[0]

	t.Run("Valid production update", func(t *testing.T) {
		newProduction := model.Production{
			Credits:  3,
			Steel:    2,
			Titanium: 1,
			Plants:   4,
			Energy:   3,
			Heat:     2,
		}

		err := playerService.UpdatePlayerProduction(ctx, game.ID, playerID, newProduction)
		assert.NoError(t, err)

		updatedPlayer, err := playerService.GetPlayer(ctx, game.ID, playerID)
		require.NoError(t, err)

		player := updatedPlayer
		assert.Equal(t, 3, player.Production.Credits)
		assert.Equal(t, 2, player.Production.Steel)
		assert.Equal(t, 1, player.Production.Titanium)
		assert.Equal(t, 4, player.Production.Plants)
		assert.Equal(t, 3, player.Production.Energy)
		assert.Equal(t, 2, player.Production.Heat)
	})

	t.Run("Negative production values", func(t *testing.T) {
		newProduction := model.Production{
			Credits:  -1,
			Steel:    -2,
			Titanium: 1,
			Plants:   2,
			Energy:   -1,
			Heat:     0,
		}

		err := playerService.UpdatePlayerProduction(ctx, game.ID, playerID, newProduction)
		assert.NoError(t, err)

		updatedPlayer, err := playerService.GetPlayer(ctx, game.ID, playerID)
		require.NoError(t, err)

		player := updatedPlayer
		assert.Equal(t, -1, player.Production.Credits)
		assert.Equal(t, -2, player.Production.Steel)
		assert.Equal(t, 1, player.Production.Titanium)
		assert.Equal(t, 2, player.Production.Plants)
		assert.Equal(t, -1, player.Production.Energy)
		assert.Equal(t, 0, player.Production.Heat)
	})

	t.Run("Invalid game ID", func(t *testing.T) {
		newProduction := model.Production{Credits: 5}
		err := playerService.UpdatePlayerProduction(ctx, "invalid-game-id", playerID, newProduction)
		assert.Error(t, err)
	})

	t.Run("Invalid player ID", func(t *testing.T) {
		newProduction := model.Production{Credits: 5}
		err := playerService.UpdatePlayerProduction(ctx, game.ID, "invalid-player-id", newProduction)
		assert.Error(t, err)
	})
}

func TestPlayerService_GetPlayer(t *testing.T) {
	playerService, _, game := setupPlayerServiceTest(t)
	ctx := context.Background()
	playerID := game.PlayerIDs[0]

	t.Run("Get existing player", func(t *testing.T) {
		player, err := playerService.GetPlayer(ctx, game.ID, playerID)
		assert.NoError(t, err)
		assert.NotNil(t, player)
		assert.Equal(t, playerID, player.ID)
		assert.Equal(t, "TestPlayer", player.Name)
	})

	t.Run("Invalid game ID", func(t *testing.T) {
		_, err := playerService.GetPlayer(ctx, "invalid-game-id", playerID)
		assert.Error(t, err)
	})

	t.Run("Invalid player ID", func(t *testing.T) {
		_, err := playerService.GetPlayer(ctx, game.ID, "invalid-player-id")
		assert.Error(t, err)
	})
}

func TestPlayerService_UpdatePlayerConnectionStatus(t *testing.T) {
	playerService, _, game := setupPlayerServiceTest(t)
	ctx := context.Background()
	playerID := game.PlayerIDs[0]

	t.Run("Update to connected status", func(t *testing.T) {
		err := playerService.UpdatePlayerConnectionStatus(ctx, game.ID, playerID, model.ConnectionStatusConnected)
		assert.NoError(t, err)

		// Verify the status was updated
		player, err := playerService.GetPlayer(ctx, game.ID, playerID)
		assert.NoError(t, err)
		assert.Equal(t, model.ConnectionStatusConnected, player.ConnectionStatus)
	})

	t.Run("Update to disconnected status", func(t *testing.T) {
		err := playerService.UpdatePlayerConnectionStatus(ctx, game.ID, playerID, model.ConnectionStatusDisconnected)
		assert.NoError(t, err)

		// Verify the status was updated
		player, err := playerService.GetPlayer(ctx, game.ID, playerID)
		assert.NoError(t, err)
		assert.Equal(t, model.ConnectionStatusDisconnected, player.ConnectionStatus)
	})

	t.Run("Invalid game ID", func(t *testing.T) {
		err := playerService.UpdatePlayerConnectionStatus(ctx, "invalid-game-id", playerID, model.ConnectionStatusConnected)
		assert.Error(t, err)
	})

	t.Run("Invalid player ID", func(t *testing.T) {
		err := playerService.UpdatePlayerConnectionStatus(ctx, game.ID, "invalid-player-id", model.ConnectionStatusConnected)
		assert.Error(t, err)
	})
}

func TestPlayerService_FindPlayerByName(t *testing.T) {
	playerService, gameService, game := setupPlayerServiceTest(t)
	ctx := context.Background()

	// Add another player to the game for testing
	game, err := gameService.JoinGame(ctx, game.ID, "TestPlayer2")
	require.NoError(t, err)

	t.Run("Find existing player by name", func(t *testing.T) {
		player, err := playerService.GetPlayerByName(ctx, game.ID, "TestPlayer")
		assert.NoError(t, err)
		assert.NotNil(t, player)
		assert.Equal(t, "TestPlayer", player.Name)
		assert.Equal(t, game.PlayerIDs[0], player.ID)
	})

	t.Run("Find second player by name", func(t *testing.T) {
		player, err := playerService.GetPlayerByName(ctx, game.ID, "TestPlayer2")
		assert.NoError(t, err)
		assert.NotNil(t, player)
		assert.Equal(t, "TestPlayer2", player.Name)
		assert.Equal(t, game.PlayerIDs[1], player.ID)
	})

	t.Run("Player not found", func(t *testing.T) {
		_, err := playerService.GetPlayerByName(ctx, game.ID, "NonexistentPlayer")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Invalid game ID", func(t *testing.T) {
		_, err := playerService.GetPlayerByName(ctx, "invalid-game-id", "TestPlayer")
		assert.Error(t, err)
	})

	t.Run("Empty player name", func(t *testing.T) {
		_, err := playerService.GetPlayerByName(ctx, game.ID, "")
		assert.Error(t, err)
	})
}
