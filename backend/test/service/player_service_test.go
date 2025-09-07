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
	*model.Game,
) {
	eventBus := events.NewInMemoryEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	parametersRepo := repository.NewGlobalParametersRepository(eventBus)

	playerService := service.NewPlayerService(gameRepo, playerRepo)
	
	cardService := service.NewCardService(gameRepo, playerRepo)
	gameService := service.NewGameService(gameRepo, playerRepo, parametersRepo, cardService.(*service.CardServiceImpl), eventBus)

	ctx := context.Background()
	game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	game, err = gameService.JoinGame(ctx, game.ID, "TestPlayer")
	require.NoError(t, err)

	return playerService, gameService, game
}

func TestPlayerService_UpdatePlayerResources(t *testing.T) {
	playerService, gameService, game := setupPlayerServiceTest(t)
	ctx := context.Background()
	playerID := game.Players[0].ID

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

		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		player := updatedGame.Players[0]
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

		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		player := updatedGame.Players[0]
		assert.Equal(t, 0, player.Resources.Credits)
		assert.Equal(t, 0, player.Resources.Steel)
		assert.Equal(t, 0, player.Resources.Titanium)
		assert.Equal(t, 0, player.Resources.Plants)
		assert.Equal(t, 0, player.Resources.Energy)
		assert.Equal(t, 0, player.Resources.Heat)
	})
}

func TestPlayerService_UpdatePlayerProduction(t *testing.T) {
	playerService, gameService, game := setupPlayerServiceTest(t)
	ctx := context.Background()
	playerID := game.Players[0].ID

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

		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		player := updatedGame.Players[0]
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

		updatedGame, err := gameService.GetGame(ctx, game.ID)
		require.NoError(t, err)

		player := updatedGame.Players[0]
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
	playerID := game.Players[0].ID

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
