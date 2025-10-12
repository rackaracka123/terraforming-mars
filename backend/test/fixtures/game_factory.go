package fixtures

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/model"

	"github.com/stretchr/testify/require"
)

// GameFixture holds a complete game setup for testing
type GameFixture struct {
	GameID    string
	PlayerID  string
	Player    model.Player
	Container *ServiceContainer
}

// NewGameInActionPhase creates a game in the action phase with one player
func NewGameInActionPhase(t *testing.T, container *ServiceContainer) *GameFixture {
	ctx := context.Background()

	// Create game
	game, err := container.GameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)
	gameID := game.ID

	// Set game to active status and action phase
	err = container.GameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	require.NoError(t, err)
	err = container.GameRepo.UpdatePhase(ctx, gameID, model.GamePhaseAction)
	require.NoError(t, err)

	// Create player with default resources
	player := model.Player{
		ID:              "player1",
		Name:            "Test Player",
		Resources:       model.Resources{Credits: 40},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
	}
	err = container.PlayerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	return &GameFixture{
		GameID:    gameID,
		PlayerID:  player.ID,
		Player:    player,
		Container: container,
	}
}

// NewGameInLobby creates a game in lobby status
func NewGameInLobby(t *testing.T, container *ServiceContainer) *GameFixture {
	ctx := context.Background()

	// Create game (defaults to lobby status)
	game, err := container.GameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Create host player
	player := model.Player{
		ID:              "player1",
		Name:            "Host Player",
		Resources:       model.Resources{Credits: 0},
		Production:      model.Production{},
		TerraformRating: 20,
		IsConnected:     true,
	}
	err = container.PlayerRepo.Create(ctx, game.ID, player)
	require.NoError(t, err)

	return &GameFixture{
		GameID:    game.ID,
		PlayerID:  player.ID,
		Player:    player,
		Container: container,
	}
}

// NewGameWithMultiplePlayers creates a game with specified number of players
func NewGameWithMultiplePlayers(t *testing.T, container *ServiceContainer, playerCount int) *GameFixture {
	ctx := context.Background()

	// Create game
	game, err := container.GameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)
	gameID := game.ID

	// Set game to active status and action phase
	err = container.GameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	require.NoError(t, err)
	err = container.GameRepo.UpdatePhase(ctx, gameID, model.GamePhaseAction)
	require.NoError(t, err)

	// Create multiple players
	var firstPlayerID string
	var firstPlayer model.Player

	for i := 0; i < playerCount; i++ {
		player := model.Player{
			ID:              "player" + string(rune('1'+i)),
			Name:            "Test Player " + string(rune('1'+i)),
			Resources:       model.Resources{Credits: 40},
			Production:      model.Production{Credits: 1},
			TerraformRating: 20,
			IsConnected:     true,
		}
		err = container.PlayerRepo.Create(ctx, gameID, player)
		require.NoError(t, err)

		if i == 0 {
			firstPlayerID = player.ID
			firstPlayer = player
		}
	}

	return &GameFixture{
		GameID:    gameID,
		PlayerID:  firstPlayerID,
		Player:    firstPlayer,
		Container: container,
	}
}

// NewGameInStartingCardSelection creates a game in starting card selection phase
func NewGameInStartingCardSelection(t *testing.T, container *ServiceContainer) *GameFixture {
	ctx := context.Background()

	// Create game
	game, err := container.GameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)
	gameID := game.ID

	// Set game to active status and starting card selection phase
	err = container.GameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	require.NoError(t, err)
	err = container.GameRepo.UpdatePhase(ctx, gameID, model.GamePhaseStartingCardSelection)
	require.NoError(t, err)

	// Create player without corporation (will be selected later)
	player := model.Player{
		ID:              "player1",
		Name:            "Test Player",
		Resources:       model.Resources{Credits: 42}, // Standard corporation starting credits
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		Corporation:     nil,
		IsConnected:     true,
	}
	err = container.PlayerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	return &GameFixture{
		GameID:    gameID,
		PlayerID:  player.ID,
		Player:    player,
		Container: container,
	}
}
