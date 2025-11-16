package repositories_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
)

func TestGameRepository_Create(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	settings := game.GameSettings{
		DevelopmentMode: false,
		CardPacks:       []string{"base"},
	}

	g, err := repo.Create(context.Background(), settings)
	require.NoError(t, err)

	// Verify game was created
	assert.NotEmpty(t, g.ID)
	assert.Equal(t, game.GameStatusLobby, g.Status)
	assert.Equal(t, game.GamePhaseWaitingForGameStart, g.CurrentPhase)
	assert.NotZero(t, g.CreatedAt)
	assert.Equal(t, settings, g.Settings)
	assert.Empty(t, g.PlayerIDs)
	assert.Equal(t, 1, g.Generation)
}

func TestGameRepository_GetByID_Success(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	// Create a game
	settings := game.GameSettings{DevelopmentMode: true, CardPacks: []string{"base", "corporate"}}
	created, err := repo.Create(context.Background(), settings)
	require.NoError(t, err)

	// Retrieve the game
	retrieved, err := repo.GetByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Settings, retrieved.Settings)
	assert.Equal(t, created.Status, retrieved.Status)
}

func TestGameRepository_GetByID_NotFound(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	_, err := repo.GetByID(context.Background(), "nonexistent-game")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "game not found")
}

func TestGameRepository_GetByID_EmptyID(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	_, err := repo.GetByID(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestGameRepository_Delete(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	// Create a game
	created, err := repo.Create(context.Background(), game.GameSettings{})
	require.NoError(t, err)

	// Delete the game
	err = repo.Delete(context.Background(), created.ID)
	assert.NoError(t, err)

	// Verify game is deleted
	_, err = repo.GetByID(context.Background(), created.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "game not found")
}

func TestGameRepository_UpdateStatus(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	// Create a game
	created, err := repo.Create(context.Background(), game.GameSettings{})
	require.NoError(t, err)
	assert.Equal(t, game.GameStatusLobby, created.Status)

	// Update status to active
	err = repo.UpdateStatus(context.Background(), created.ID, game.GameStatusActive)
	assert.NoError(t, err)

	// Verify status was updated
	retrieved, err := repo.GetByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.Equal(t, game.GameStatusActive, retrieved.Status)
}

func TestGameRepository_UpdatePhase(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	// Create a game
	created, err := repo.Create(context.Background(), game.GameSettings{})
	require.NoError(t, err)
	assert.Equal(t, game.GamePhaseWaitingForGameStart, created.CurrentPhase)

	// Update phase
	err = repo.UpdatePhase(context.Background(), created.ID, game.GamePhaseAction)
	assert.NoError(t, err)

	// Verify phase was updated
	retrieved, err := repo.GetByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.Equal(t, game.GamePhaseAction, retrieved.CurrentPhase)
}

func TestGameRepository_AddPlayerID(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	// Create a game
	created, err := repo.Create(context.Background(), game.GameSettings{})
	require.NoError(t, err)
	assert.Empty(t, created.PlayerIDs)

	// Add player IDs
	err = repo.AddPlayerID(context.Background(), created.ID, "player-1")
	assert.NoError(t, err)

	err = repo.AddPlayerID(context.Background(), created.ID, "player-2")
	assert.NoError(t, err)

	// Verify players were added
	retrieved, err := repo.GetByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.Len(t, retrieved.PlayerIDs, 2)
	assert.Contains(t, retrieved.PlayerIDs, "player-1")
	assert.Contains(t, retrieved.PlayerIDs, "player-2")
}

func TestGameRepository_AddPlayerID_Duplicate(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	// Create a game
	created, err := repo.Create(context.Background(), game.GameSettings{})
	require.NoError(t, err)

	// Add player
	err = repo.AddPlayerID(context.Background(), created.ID, "player-1")
	assert.NoError(t, err)

	// Try to add same player again - should return error
	err = repo.AddPlayerID(context.Background(), created.ID, "player-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in game")
}

func TestGameRepository_RemovePlayerID(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	// Create a game with players
	created, err := repo.Create(context.Background(), game.GameSettings{})
	require.NoError(t, err)

	err = repo.AddPlayerID(context.Background(), created.ID, "player-1")
	require.NoError(t, err)
	err = repo.AddPlayerID(context.Background(), created.ID, "player-2")
	require.NoError(t, err)

	// Remove a player
	err = repo.RemovePlayerID(context.Background(), created.ID, "player-1")
	assert.NoError(t, err)

	// Verify player was removed
	retrieved, err := repo.GetByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.Len(t, retrieved.PlayerIDs, 1)
	assert.NotContains(t, retrieved.PlayerIDs, "player-1")
	assert.Contains(t, retrieved.PlayerIDs, "player-2")
}

func TestGameRepository_SetHostPlayer(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	// Create a game
	created, err := repo.Create(context.Background(), game.GameSettings{})
	require.NoError(t, err)

	// Set host player
	err = repo.SetHostPlayer(context.Background(), created.ID, "player-1")
	assert.NoError(t, err)

	// Verify host was set
	retrieved, err := repo.GetByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "player-1", retrieved.HostPlayerID)
}

func TestGameRepository_UpdateGeneration(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	// Create a game
	created, err := repo.Create(context.Background(), game.GameSettings{})
	require.NoError(t, err)
	assert.Equal(t, 1, created.Generation)

	// Update generation
	err = repo.UpdateGeneration(context.Background(), created.ID, 5)
	assert.NoError(t, err)

	// Verify generation was updated
	retrieved, err := repo.GetByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.Equal(t, 5, retrieved.Generation)
}

func TestGameRepository_List(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	// Create multiple games with different statuses
	_, err := repo.Create(context.Background(), game.GameSettings{})
	require.NoError(t, err)

	activeGame, err := repo.Create(context.Background(), game.GameSettings{})
	require.NoError(t, err)
	err = repo.UpdateStatus(context.Background(), activeGame.ID, game.GameStatusActive)
	require.NoError(t, err)

	completedGame, err := repo.Create(context.Background(), game.GameSettings{})
	require.NoError(t, err)
	err = repo.UpdateStatus(context.Background(), completedGame.ID, game.GameStatusCompleted)
	require.NoError(t, err)

	// List all games
	allGames, err := repo.List(context.Background(), "")
	assert.NoError(t, err)
	assert.Len(t, allGames, 3)

	// List active games
	activeGames, err := repo.List(context.Background(), string(game.GameStatusActive))
	assert.NoError(t, err)
	assert.Len(t, activeGames, 1)
	assert.Equal(t, game.GameStatusActive, activeGames[0].Status)

	// List lobby games
	lobbyGames, err := repo.List(context.Background(), string(game.GameStatusLobby))
	assert.NoError(t, err)
	assert.Len(t, lobbyGames, 1)
	assert.Equal(t, game.GameStatusLobby, lobbyGames[0].Status)
}

func TestGameRepository_MultipleGames_Independence(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := game.NewRepository(eventBus)

	// Create two games
	game1, err := repo.Create(context.Background(), game.GameSettings{DevelopmentMode: true})
	require.NoError(t, err)

	game2, err := repo.Create(context.Background(), game.GameSettings{DevelopmentMode: false})
	require.NoError(t, err)

	// Update game1 status
	err = repo.UpdateStatus(context.Background(), game1.ID, game.GameStatusActive)
	require.NoError(t, err)

	// Update game2 phase
	err = repo.UpdatePhase(context.Background(), game2.ID, game.GamePhaseStartingCardSelection)
	require.NoError(t, err)

	// Verify game1 is not affected by game2 changes
	retrieved1, err := repo.GetByID(context.Background(), game1.ID)
	assert.NoError(t, err)
	assert.Equal(t, game.GameStatusActive, retrieved1.Status)
	assert.Equal(t, game.GamePhaseWaitingForGameStart, retrieved1.CurrentPhase) // Should still be lobby phase

	// Verify game2 is not affected by game1 changes
	retrieved2, err := repo.GetByID(context.Background(), game2.ID)
	assert.NoError(t, err)
	assert.Equal(t, game.GameStatusLobby, retrieved2.Status) // Should still be lobby status
	assert.Equal(t, game.GamePhaseStartingCardSelection, retrieved2.CurrentPhase)
}
