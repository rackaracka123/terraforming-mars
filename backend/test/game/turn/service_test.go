package turn_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (turn.Service, game.Repository, player.Repository, string, []string) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)

	turnRepo := turn.NewRepository(gameRepo, playerRepo)
	turnService := turn.NewService(turnRepo)

	// Create game
	game, err := gameRepo.Create(ctx, game.GameSettings{
		MaxPlayers: 4,
	})
	require.NoError(t, err)
	gameID := game.ID

	// Set game to active status
	err = gameRepo.UpdateStatus(ctx, gameID, game.GameStatusActive)
	require.NoError(t, err)

	// Create 3 players
	playerIDs := []string{"player1", "player2", "player3"}
	for _, playerID := range playerIDs {
		player := player.Player{
			ID:               playerID,
			Name:             "Player " + playerID,
			AvailableActions: 2,
			Passed:           false,
			Resources: resources.Resources{
				Credits: 50,
			},
			TerraformRating: 20,
		}
		err := playerRepo.Create(ctx, gameID, player)
		require.NoError(t, err)
	}

	// Set current turn to first player
	err = gameRepo.UpdateCurrentTurn(ctx, gameID, &playerIDs[0])
	require.NoError(t, err)

	return turnService, gameRepo, playerRepo, gameID, playerIDs
}

func TestTurnService_SkipTurn(t *testing.T) {
	turnService, gameRepo, _, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	// Skip turn for first player (should advance to player 2)
	generationEnded, err := turnService.SkipTurn(ctx, gameID, playerIDs[0])
	assert.NoError(t, err)
	assert.False(t, generationEnded, "Generation should not end after one skip")

	// Verify turn advanced to player 2
	game, err := gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	require.NotNil(t, game.CurrentTurn)
	assert.Equal(t, playerIDs[1], *game.CurrentTurn)
}

func TestTurnService_PassTurn(t *testing.T) {
	turnService, _, playerRepo, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	// Player 1 passes (has full 2 actions available)
	generationEnded, err := turnService.SkipTurn(ctx, gameID, playerIDs[0])
	assert.NoError(t, err)
	assert.False(t, generationEnded)

	// Verify player 1 is marked as passed
	player1, err := playerRepo.GetByID(ctx, gameID, playerIDs[0])
	require.NoError(t, err)
	assert.True(t, player1.Passed, "Player should be marked as passed")
}

func TestTurnService_SkipVsPass(t *testing.T) {
	turnService, _, playerRepo, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	t.Run("Skip with 1 action left", func(t *testing.T) {
		// Set player 1 to have 1 action (used one)
		err := playerRepo.UpdateAvailableActions(ctx, gameID, playerIDs[0], 1)
		require.NoError(t, err)

		generationEnded, err := turnService.SkipTurn(ctx, gameID, playerIDs[0])
		assert.NoError(t, err)
		assert.False(t, generationEnded)

		// Verify player 1 is NOT marked as passed (they skipped, not passed)
		player1, err := playerRepo.GetByID(ctx, gameID, playerIDs[0])
		require.NoError(t, err)
		assert.False(t, player1.Passed, "Player should NOT be passed after skipping with 1 action")
	})

	t.Run("Pass with 2 actions", func(t *testing.T) {
		// Reset player 2 to have 2 actions
		err := playerRepo.UpdateAvailableActions(ctx, gameID, playerIDs[1], 2)
		require.NoError(t, err)

		generationEnded, err := turnService.SkipTurn(ctx, gameID, playerIDs[1])
		assert.NoError(t, err)
		assert.False(t, generationEnded)

		// Verify player 2 IS marked as passed
		player2, err := playerRepo.GetByID(ctx, gameID, playerIDs[1])
		require.NoError(t, err)
		assert.True(t, player2.Passed, "Player should be passed with 2 actions")
	})
}

func TestTurnService_LastActivePlayerUnlimitedActions(t *testing.T) {
	turnService, _, playerRepo, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	// Player 1 passes (2 actions)
	_, err := turnService.SkipTurn(ctx, gameID, playerIDs[0])
	require.NoError(t, err)

	// Player 2 passes (2 actions) - should grant player 3 unlimited actions
	_, err = turnService.SkipTurn(ctx, gameID, playerIDs[1])
	require.NoError(t, err)

	// Verify player 3 has unlimited actions (-1)
	player3, err := playerRepo.GetByID(ctx, gameID, playerIDs[2])
	require.NoError(t, err)
	assert.Equal(t, -1, player3.AvailableActions, "Last active player should have unlimited actions")
}

func TestTurnService_GenerationEnd(t *testing.T) {
	turnService, _, _, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	// All players pass
	_, err := turnService.SkipTurn(ctx, gameID, playerIDs[0])
	require.NoError(t, err)

	_, err = turnService.SkipTurn(ctx, gameID, playerIDs[1])
	require.NoError(t, err)

	// Last player passes - generation should end
	generationEnded, err := turnService.SkipTurn(ctx, gameID, playerIDs[2])
	assert.NoError(t, err)
	assert.True(t, generationEnded, "Generation should end when all players pass")
}

func TestTurnService_GenerationEndNoActions(t *testing.T) {
	turnService, _, playerRepo, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	// Set all players to have 0 actions
	for _, playerID := range playerIDs {
		err := playerRepo.UpdateAvailableActions(ctx, gameID, playerID, 0)
		require.NoError(t, err)
	}

	// Check generation ended
	ended, err := turnService.IsGenerationEnded(ctx, gameID)
	assert.NoError(t, err)
	assert.True(t, ended, "Generation should end when all players have 0 actions")
}

func TestTurnService_ValidateCurrentPlayer(t *testing.T) {
	turnService, _, playerRepo, gameID, playerIDs := setupTest(t)
	ctx := context.Background()
	_ = playerRepo // Used in setup

	t.Run("Valid current player", func(t *testing.T) {
		err := turnService.ValidateCurrentPlayer(ctx, gameID, playerIDs[0])
		assert.NoError(t, err)
	})

	t.Run("Invalid current player", func(t *testing.T) {
		err := turnService.ValidateCurrentPlayer(ctx, gameID, playerIDs[1])
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only the current player")
	})
}

func TestTurnService_AdvanceToNextPlayer(t *testing.T) {
	turnService, gameRepo, _, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	// Advance from player 1 to player 2
	err := turnService.AdvanceToNextPlayer(ctx, gameID)
	assert.NoError(t, err)

	game, err := gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	require.NotNil(t, game.CurrentTurn)
	assert.Equal(t, playerIDs[1], *game.CurrentTurn)

	// Advance from player 2 to player 3
	err = turnService.AdvanceToNextPlayer(ctx, gameID)
	assert.NoError(t, err)

	game, err = gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	require.NotNil(t, game.CurrentTurn)
	assert.Equal(t, playerIDs[2], *game.CurrentTurn)

	// Advance from player 3 back to player 1 (wrap around)
	err = turnService.AdvanceToNextPlayer(ctx, gameID)
	assert.NoError(t, err)

	game, err = gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	require.NotNil(t, game.CurrentTurn)
	assert.Equal(t, playerIDs[0], *game.CurrentTurn)
}

func TestTurnService_AdvanceSkipsPassedPlayers(t *testing.T) {
	turnService, gameRepo, playerRepo, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	// Mark player 2 as passed
	err := playerRepo.UpdatePassed(ctx, gameID, playerIDs[1], true)
	require.NoError(t, err)

	// Advance from player 1 - should skip player 2 and go to player 3
	err = turnService.AdvanceToNextPlayer(ctx, gameID)
	assert.NoError(t, err)

	game, err := gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	require.NotNil(t, game.CurrentTurn)
	assert.Equal(t, playerIDs[2], *game.CurrentTurn, "Should skip passed player")
}

func TestTurnService_GetNextPlayer(t *testing.T) {
	turnService, _, playerRepo, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	// Get next player from player 1
	nextPlayerID, err := turnService.GetNextPlayer(ctx, gameID)
	assert.NoError(t, err)
	assert.Equal(t, playerIDs[1], nextPlayerID)

	// Mark player 2 as passed and get next player
	err = playerRepo.UpdatePassed(ctx, gameID, playerIDs[1], true)
	require.NoError(t, err)

	nextPlayerID, err = turnService.GetNextPlayer(ctx, gameID)
	assert.NoError(t, err)
	assert.Equal(t, playerIDs[2], nextPlayerID, "Should skip passed player")
}

func TestTurnService_IsGenerationEnded(t *testing.T) {
	turnService, _, playerRepo, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	t.Run("Not ended - players have actions", func(t *testing.T) {
		ended, err := turnService.IsGenerationEnded(ctx, gameID)
		assert.NoError(t, err)
		assert.False(t, ended)
	})

	t.Run("Not ended - one player has unlimited actions", func(t *testing.T) {
		// Mark first two as passed, third has unlimited
		err := playerRepo.UpdatePassed(ctx, gameID, playerIDs[0], true)
		require.NoError(t, err)
		err = playerRepo.UpdatePassed(ctx, gameID, playerIDs[1], true)
		require.NoError(t, err)
		err = playerRepo.UpdateAvailableActions(ctx, gameID, playerIDs[2], -1)
		require.NoError(t, err)

		ended, err := turnService.IsGenerationEnded(ctx, gameID)
		assert.NoError(t, err)
		assert.False(t, ended, "Should not end with one player having unlimited actions")
	})

	t.Run("Ended - all passed", func(t *testing.T) {
		// Mark all as passed
		for _, playerID := range playerIDs {
			err := playerRepo.UpdatePassed(ctx, gameID, playerID, true)
			require.NoError(t, err)
		}

		ended, err := turnService.IsGenerationEnded(ctx, gameID)
		assert.NoError(t, err)
		assert.True(t, ended)
	})
}
