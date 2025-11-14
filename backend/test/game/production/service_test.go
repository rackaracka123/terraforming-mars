package production_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/production"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (production.Service, game.Repository, player.Repository, game.CardDeckRepository, string, []string) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardDeckRepo := repository.NewCardDeckRepository()

	productionRepo := production.NewRepository(gameRepo, playerRepo, cardDeckRepo)
	productionService := production.NewService(productionRepo)

	// Create game
	game, err := gameRepo.Create(ctx, game.GameSettings{
		MaxPlayers: 4,
	})
	require.NoError(t, err)
	gameID := game.ID

	// Set game to active status
	err = gameRepo.UpdateStatus(ctx, gameID, game.GameStatusActive)
	require.NoError(t, err)

	// Set initial phase to action
	err = gameRepo.UpdatePhase(ctx, gameID, game.GamePhaseAction)
	require.NoError(t, err)

	// Create 3 players with resources and production
	playerIDs := []string{"player1", "player2", "player3"}
	for _, playerID := range playerIDs {
		player := player.Player{
			ID:               playerID,
			Name:             "Player " + playerID,
			AvailableActions: 2,
			Passed:           false,
			Resources: resources.Resources{
				Credits:  10,
				Steel:    5,
				Titanium: 3,
				Plants:   8,
				Energy:   6,
				Heat:     4,
			},
			Production: resources.Production{
				Credits:  3,
				Steel:    2,
				Titanium: 1,
				Plants:   2,
				Energy:   4,
				Heat:     1,
			},
			TerraformRating: 20,
		}
		err := playerRepo.Create(ctx, gameID, player)
		require.NoError(t, err)
	}

	// Initialize card deck with some test cards
	cards := []model.Card{
		{ID: "card1"}, {ID: "card2"}, {ID: "card3"}, {ID: "card4"},
		{ID: "card5"}, {ID: "card6"}, {ID: "card7"}, {ID: "card8"},
		{ID: "card9"}, {ID: "card10"}, {ID: "card11"}, {ID: "card12"},
	}
	err = cardDeckRepo.InitializeDeck(ctx, gameID, cards)
	require.NoError(t, err)

	// Set current turn to first player
	err = gameRepo.UpdateCurrentTurn(ctx, gameID, &playerIDs[0])
	require.NoError(t, err)

	return productionService, gameRepo, playerRepo, cardDeckRepo, gameID, playerIDs
}

func TestProductionService_ApplyProduction(t *testing.T) {
	productionService, _, playerRepo, _, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	// Get initial state
	player, err := playerRepo.GetByID(ctx, gameID, playerIDs[0])
	require.NoError(t, err)
	initialCredits := player.Resources.Credits
	initialSteel := player.Resources.Steel
	initialEnergy := player.Resources.Energy
	initialHeat := player.Resources.Heat

	// Apply production
	oldResources, newResources, energyConverted, err := productionService.ApplyProduction(ctx, gameID, playerIDs[0])
	assert.NoError(t, err)

	// Verify old resources match initial state
	assert.Equal(t, initialCredits, oldResources.Credits)
	assert.Equal(t, initialEnergy, oldResources.Energy)

	// Verify energy was converted
	assert.Equal(t, initialEnergy, energyConverted)

	// Verify new resources calculated correctly
	// Credits: 10 (old) + 3 (production) + 20 (TR) = 33
	assert.Equal(t, initialCredits+player.Production.Credits+player.TerraformRating, newResources.Credits)

	// Steel: 5 (old) + 2 (production) = 7
	assert.Equal(t, initialSteel+player.Production.Steel, newResources.Steel)

	// Energy: production value only (4)
	assert.Equal(t, player.Production.Energy, newResources.Energy)

	// Heat: 4 (old) + 6 (converted energy) + 1 (production) = 11
	assert.Equal(t, initialHeat+energyConverted+player.Production.Heat, newResources.Heat)

	// Verify player resources were updated
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerIDs[0])
	require.NoError(t, err)
	assert.Equal(t, newResources.Credits, updatedPlayer.Resources.Credits)
	assert.Equal(t, newResources.Steel, updatedPlayer.Resources.Steel)
	assert.Equal(t, newResources.Energy, updatedPlayer.Resources.Energy)
	assert.Equal(t, newResources.Heat, updatedPlayer.Resources.Heat)
}

func TestProductionService_DrawProductionCards(t *testing.T) {
	productionService, _, _, _, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	t.Run("Draw 4 cards", func(t *testing.T) {
		cards, err := productionService.DrawProductionCards(ctx, gameID, playerIDs[0], 4)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(cards), "Should draw exactly 4 cards")
	})

	t.Run("Draw until deck exhausted", func(t *testing.T) {
		// Already drew 4, deck has 12 total, so 8 remaining
		cards, err := productionService.DrawProductionCards(ctx, gameID, playerIDs[0], 10)
		assert.NoError(t, err)
		assert.Equal(t, 8, len(cards), "Should draw remaining 8 cards")
	})

	t.Run("Draw from empty deck", func(t *testing.T) {
		cards, err := productionService.DrawProductionCards(ctx, gameID, playerIDs[0], 4)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(cards), "Should draw no cards from empty deck")
	})
}

func TestProductionService_ResetPlayerForNewGeneration(t *testing.T) {
	productionService, _, playerRepo, _, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	t.Run("Reset for multiplayer", func(t *testing.T) {
		// Mark player as passed with 0 actions
		err := playerRepo.UpdatePassed(ctx, gameID, playerIDs[0], true)
		require.NoError(t, err)
		err = playerRepo.UpdateAvailableActions(ctx, gameID, playerIDs[0], 0)
		require.NoError(t, err)

		// Reset player
		err = productionService.ResetPlayerForNewGeneration(ctx, gameID, playerIDs[0], false)
		assert.NoError(t, err)

		// Verify player reset
		player, err := playerRepo.GetByID(ctx, gameID, playerIDs[0])
		require.NoError(t, err)
		assert.False(t, player.Passed, "Player should not be passed")
		assert.Equal(t, 2, player.AvailableActions, "Player should have 2 actions for multiplayer")
	})

	t.Run("Reset for solo", func(t *testing.T) {
		err := productionService.ResetPlayerForNewGeneration(ctx, gameID, playerIDs[1], true)
		assert.NoError(t, err)

		player, err := playerRepo.GetByID(ctx, gameID, playerIDs[1])
		require.NoError(t, err)
		assert.False(t, player.Passed)
		assert.Equal(t, -1, player.AvailableActions, "Player should have unlimited actions for solo")
	})
}

func TestProductionService_ExecuteProductionPhase(t *testing.T) {
	productionService, gameRepo, playerRepo, _, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	// Get initial game state
	game, err := gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	initialGeneration := game.Generation

	// Get initial player states
	player1, err := playerRepo.GetByID(ctx, gameID, playerIDs[0])
	require.NoError(t, err)
	initialCredits := player1.Resources.Credits
	initialEnergy := player1.Resources.Energy

	// Execute production phase
	err = productionService.ExecuteProductionPhase(ctx, gameID)
	assert.NoError(t, err)

	// Verify generation advanced
	game, err = gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, initialGeneration+1, game.Generation, "Generation should advance by 1")

	// Verify phase updated
	assert.Equal(t, game.GamePhaseProductionAndCardDraw, game.CurrentPhase)

	// Verify current turn set to first player
	require.NotNil(t, game.CurrentTurn)
	assert.Equal(t, playerIDs[0], *game.CurrentTurn)

	// Verify player 1 resources updated
	player1, err = playerRepo.GetByID(ctx, gameID, playerIDs[0])
	require.NoError(t, err)

	// Credits should increase by production + TR
	expectedCredits := initialCredits + player1.Production.Credits + player1.TerraformRating
	assert.Equal(t, expectedCredits, player1.Resources.Credits)

	// Energy should reset to production value
	assert.Equal(t, player1.Production.Energy, player1.Resources.Energy)

	// Heat should increase by old energy + production
	expectedHeat := 4 + initialEnergy + player1.Production.Heat
	assert.Equal(t, expectedHeat, player1.Resources.Heat)

	// Verify player state reset
	assert.False(t, player1.Passed, "Player should not be passed after production")
	assert.Equal(t, 2, player1.AvailableActions, "Player should have 2 actions")

	// Verify production phase data set
	require.NotNil(t, player1.ProductionPhase)
	assert.Equal(t, 4, len(player1.ProductionPhase.AvailableCards), "Should have 4 cards drawn")
	assert.False(t, player1.ProductionPhase.SelectionComplete)
	assert.Equal(t, initialEnergy, player1.ProductionPhase.EnergyConverted)
	assert.Equal(t, player1.Production.Credits+player1.TerraformRating, player1.ProductionPhase.CreditsIncome)
}

func TestProductionService_ExecuteProductionPhaseSolo(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardDeckRepo := repository.NewCardDeckRepository()

	productionRepo := production.NewRepository(gameRepo, playerRepo, cardDeckRepo)
	productionService := production.NewService(productionRepo)

	// Create solo game
	game, err := gameRepo.Create(ctx, game.GameSettings{MaxPlayers: 1})
	require.NoError(t, err)
	gameID := game.ID

	err = gameRepo.UpdateStatus(ctx, gameID, game.GameStatusActive)
	require.NoError(t, err)

	// Create single player
	playerID := "solo-player"
	player := player.Player{
		ID:               playerID,
		Name:             "Solo Player",
		AvailableActions: 2,
		Resources: resources.Resources{
			Credits: 10,
			Energy:  5,
		},
		Production: resources.Production{
			Credits: 2,
			Energy:  3,
		},
		TerraformRating: 20,
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Initialize deck
	cards := []model.Card{{ID: "card1"}, {ID: "card2"}, {ID: "card3"}, {ID: "card4"}}
	err = cardDeckRepo.InitializeDeck(ctx, gameID, cards)
	require.NoError(t, err)

	// Execute production phase
	err = productionService.ExecuteProductionPhase(ctx, gameID)
	assert.NoError(t, err)

	// Verify solo player has unlimited actions
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Equal(t, -1, updatedPlayer.AvailableActions, "Solo player should have unlimited actions")
}

func TestProductionService_ProcessPlayerReady(t *testing.T) {
	productionService, gameRepo, playerRepo, _, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	// Execute production phase first
	err := productionService.ExecuteProductionPhase(ctx, gameID)
	require.NoError(t, err)

	// Verify game is in production phase
	game, err := gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, game.GamePhaseProductionAndCardDraw, game.CurrentPhase)

	t.Run("Not all players ready", func(t *testing.T) {
		// Mark player 1 as ready
		err := playerRepo.SetProductionCardsSelectionComplete(ctx, gameID, playerIDs[0])
		require.NoError(t, err)

		allReady, err := productionService.ProcessPlayerReady(ctx, gameID, playerIDs[0])
		assert.NoError(t, err)
		assert.False(t, allReady, "Not all players should be ready")

		// Verify phase didn't change
		game, err := gameRepo.GetByID(ctx, gameID)
		require.NoError(t, err)
		assert.Equal(t, game.GamePhaseProductionAndCardDraw, game.CurrentPhase)
	})

	t.Run("All players ready", func(t *testing.T) {
		// Mark remaining players as ready
		err := playerRepo.SetProductionCardsSelectionComplete(ctx, gameID, playerIDs[1])
		require.NoError(t, err)
		err = playerRepo.SetProductionCardsSelectionComplete(ctx, gameID, playerIDs[2])
		require.NoError(t, err)

		allReady, err := productionService.ProcessPlayerReady(ctx, gameID, playerIDs[2])
		assert.NoError(t, err)
		assert.True(t, allReady, "All players should be ready")

		// Verify phase advanced to action
		game, err := gameRepo.GetByID(ctx, gameID)
		require.NoError(t, err)
		assert.Equal(t, game.GamePhaseAction, game.CurrentPhase)

		// Verify current turn set to first player
		require.NotNil(t, game.CurrentTurn)
		assert.Equal(t, playerIDs[0], *game.CurrentTurn)
	})
}

func TestProductionService_ExecuteProductionPhaseValidation(t *testing.T) {
	productionService, gameRepo, _, _, gameID, _ := setupTest(t)
	ctx := context.Background()

	t.Run("Non-active game", func(t *testing.T) {
		// Set game to non-active status
		err := gameRepo.UpdateStatus(ctx, gameID, game.GameStatusLobby)
		require.NoError(t, err)

		err = productionService.ExecuteProductionPhase(ctx, gameID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not active")

		// Restore active status
		err = gameRepo.UpdateStatus(ctx, gameID, game.GameStatusActive)
		require.NoError(t, err)
	})
}

func TestProductionService_ProcessPlayerReadyValidation(t *testing.T) {
	productionService, gameRepo, _, _, gameID, playerIDs := setupTest(t)
	ctx := context.Background()

	t.Run("Non-active game", func(t *testing.T) {
		err := gameRepo.UpdateStatus(ctx, gameID, game.GameStatusLobby)
		require.NoError(t, err)

		_, err = productionService.ProcessPlayerReady(ctx, gameID, playerIDs[0])
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not active")

		// Restore active status
		err = gameRepo.UpdateStatus(ctx, gameID, game.GameStatusActive)
		require.NoError(t, err)
	})

	t.Run("Not in production phase", func(t *testing.T) {
		// Game is currently in action phase
		_, err := productionService.ProcessPlayerReady(ctx, gameID, playerIDs[0])
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in production phase")
	})
}
