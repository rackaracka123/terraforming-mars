package service_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper functions for creating test services
func createTestStandardProjectService() service.StandardProjectService {
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	return service.NewStandardProjectService(gameRepo, playerRepo, sessionManager, tileService)
}

func createTestPlayerService() service.PlayerService {
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	return service.NewPlayerService(gameRepo, playerRepo, sessionManager, boardService, tileService, forcedActionManager, eventBus)
}

func setupStandardProjectServiceTest(t *testing.T) (
	service.StandardProjectService,
	service.GameService,
	service.PlayerService,
	repository.PlayerRepository,
	model.Game,
	string, // playerID
) {
	// Initialize logger for testing
	err := logger.Init(nil)
	require.NoError(t, err)

	// Initialize services
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)

	cardRepo := repository.NewCardRepository()
	// Load card data for testing
	err = cardRepo.LoadCards(context.Background())
	if err != nil {
		t.Fatal("Failed to load card data:", err)
	}

	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber, forcedActionManager)
	gameService := service.NewGameService(gameRepo, playerRepo, cardRepo, cardService.(*service.CardServiceImpl), cardDeckRepo, boardService, sessionManager)
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager, boardService, tileService, forcedActionManager, eventBus)
	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, sessionManager, tileService)

	ctx := context.Background()

	// Create a test game
	game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Add a test player with sufficient resources
	game, err = gameService.JoinGame(ctx, game.ID, "TestPlayer")
	require.NoError(t, err)
	require.Len(t, game.PlayerIDs, 1)

	playerID := game.PlayerIDs[0]

	// Give the player sufficient credits and cards for testing
	updatedResources := model.Resources{
		Credits:  100, // Enough for all standard projects
		Steel:    10,
		Titanium: 10,
		Plants:   10,
		Energy:   10,
		Heat:     10,
	}
	err = playerRepo.UpdateResources(ctx, game.ID, playerID, updatedResources)
	require.NoError(t, err)

	// Add some cards to the player's hand
	updatedGame, err := gameService.GetGame(ctx, game.ID)
	require.NoError(t, err)

	// Add cards to player for testing - using unique card IDs to avoid duplicates across tests
	cardPrefix := make([]byte, 4)
	rand.Read(cardPrefix)
	prefix := hex.EncodeToString(cardPrefix)

	cards := []string{
		fmt.Sprintf("card1-%s", prefix),
		fmt.Sprintf("card2-%s", prefix),
		fmt.Sprintf("card3-%s", prefix),
		fmt.Sprintf("card4-%s", prefix),
		fmt.Sprintf("card5-%s", prefix),
	}
	for _, cardID := range cards {
		err = playerRepo.AddCard(ctx, game.ID, playerID, cardID)
		require.NoError(t, err)
	}

	return standardProjectService, gameService, playerService, playerRepo, updatedGame, playerID
}

func TestStandardProjectService_SellPatents(t *testing.T) {
	standardProjectService, _, _, playerRepo, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	t.Run("Initiate sell patents creates pending card selection", func(t *testing.T) {
		// Initiate sell patents
		err := standardProjectService.InitiateSellPatents(ctx, game.ID, playerID)
		assert.NoError(t, err)

		// Verify pending card selection was created
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)

		require.NotNil(t, updatedPlayer.PendingCardSelection)
		assert.Equal(t, "sell-patents", updatedPlayer.PendingCardSelection.Source)
		assert.Equal(t, 5, len(updatedPlayer.PendingCardSelection.AvailableCards)) // All 5 cards available
		assert.Equal(t, 0, updatedPlayer.PendingCardSelection.MinCards)
		assert.Equal(t, 5, updatedPlayer.PendingCardSelection.MaxCards)

		// Verify costs and rewards
		for _, cardID := range updatedPlayer.PendingCardSelection.AvailableCards {
			assert.Equal(t, 0, updatedPlayer.PendingCardSelection.CardCosts[cardID])
			assert.Equal(t, 1, updatedPlayer.PendingCardSelection.CardRewards[cardID])
		}
	})

	t.Run("Process card selection awards credits and removes cards", func(t *testing.T) {
		initialCredits := 100

		// First initiate sell patents
		err := standardProjectService.InitiateSellPatents(ctx, game.ID, playerID)
		require.NoError(t, err)

		// Get player to find card IDs
		player, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)
		require.True(t, len(player.Cards) >= 3)

		// Select 3 cards to sell (Cards is []string of card IDs)
		cardsToSell := []string{player.Cards[0], player.Cards[1], player.Cards[2]}

		// Process card selection
		err = standardProjectService.ProcessCardSelection(ctx, game.ID, playerID, cardsToSell)
		assert.NoError(t, err)

		// Verify player resources and cards
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)

		assert.Equal(t, initialCredits+3, updatedPlayer.Resources.Credits) // +1 MC per card
		assert.Equal(t, 2, len(updatedPlayer.Cards))                       // 5 - 3 = 2 cards remaining
		assert.Nil(t, updatedPlayer.PendingCardSelection)                  // Pending selection cleared
	})

	t.Run("Cannot initiate when no cards in hand", func(t *testing.T) {
		// Remove all cards from player's hand
		player, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)

		for _, cardID := range player.Cards {
			err = playerRepo.RemoveCardFromHand(ctx, game.ID, playerID, cardID)
			require.NoError(t, err)
		}

		err = standardProjectService.InitiateSellPatents(ctx, game.ID, playerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no cards to sell")
	})

	t.Run("Cannot select more cards than allowed", func(t *testing.T) {
		// Re-add some cards to player (they were removed in previous test)
		err := playerRepo.AddCard(ctx, game.ID, playerID, "test-card-1")
		require.NoError(t, err)
		err = playerRepo.AddCard(ctx, game.ID, playerID, "test-card-2")
		require.NoError(t, err)

		// Initiate sell patents
		err = standardProjectService.InitiateSellPatents(ctx, game.ID, playerID)
		require.NoError(t, err)

		// Get pending selection to verify max cards
		player, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)
		maxCards := player.PendingCardSelection.MaxCards

		// Try to select more cards than max allowed
		invalidCards := make([]string, maxCards+5)
		for i := range invalidCards {
			invalidCards[i] = fmt.Sprintf("card-%d", i)
		}

		err = standardProjectService.ProcessCardSelection(ctx, game.ID, playerID, invalidCards)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must select between")
	})

	t.Run("Can select zero cards when allowed", func(t *testing.T) {
		// Get current credits
		player, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)
		initialCredits := player.Resources.Credits

		// Initiate sell patents
		err = standardProjectService.InitiateSellPatents(ctx, game.ID, playerID)
		require.NoError(t, err)

		// Select no cards (allowed by min=0)
		err = standardProjectService.ProcessCardSelection(ctx, game.ID, playerID, []string{})
		assert.NoError(t, err)

		// Verify nothing changed except pending selection cleared
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)

		assert.Equal(t, initialCredits, updatedPlayer.Resources.Credits) // No change
		assert.Nil(t, updatedPlayer.PendingCardSelection)                // Pending selection cleared
	})
}

func TestStandardProjectService_BuildPowerPlant(t *testing.T) {
	standardProjectService, _, _, playerRepo, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	t.Run("Successful build power plant", func(t *testing.T) {
		initialCredits := 100
		initialEnergyProduction := 0
		expectedCost := 11

		// Execute build power plant
		err := standardProjectService.BuildPowerPlant(ctx, game.ID, playerID)
		assert.NoError(t, err)

		// Verify player resources and production
		// Get updated player directly from repository
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)

		player := updatedPlayer
		assert.Equal(t, initialCredits-expectedCost, player.Resources.Credits)
		assert.Equal(t, initialEnergyProduction+1, player.Production.Energy)
	})

	t.Run("Insufficient credits", func(t *testing.T) {
		// Set player credits to less than cost
		insufficientResources := model.Resources{Credits: 5}
		err := playerRepo.UpdateResources(ctx, game.ID, playerID, insufficientResources)
		require.NoError(t, err)

		err = standardProjectService.BuildPowerPlant(ctx, game.ID, playerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient credits")
	})
}

func TestStandardProjectService_LaunchAsteroid(t *testing.T) {
	standardProjectService, gameService, _, playerRepo, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	t.Run("Successful launch asteroid", func(t *testing.T) {
		initialCredits := 100
		initialTR := 20
		expectedCost := 14

		// Get initial temperature
		initialParams, err := gameService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		initialTemp := initialParams.Temperature

		// Execute launch asteroid
		err = standardProjectService.LaunchAsteroid(ctx, game.ID, playerID)
		assert.NoError(t, err)

		// Verify player resources and TR
		// Get updated player directly from repository
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)

		player := updatedPlayer
		assert.Equal(t, initialCredits-expectedCost, player.Resources.Credits)
		assert.Equal(t, initialTR+1, player.TerraformRating)

		// Verify temperature increase
		updatedParams, err := gameService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		assert.Equal(t, initialTemp+2, updatedParams.Temperature) // Each step = 2°C
	})
}

func TestStandardProjectService_BuildAquifer(t *testing.T) {
	standardProjectService, gameService, _, playerRepo, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	t.Run("Successful build aquifer", func(t *testing.T) {
		initialCredits := 100
		initialTR := 20
		expectedCost := 18

		// Get initial ocean count
		initialParams, err := gameService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		initialOceans := initialParams.Oceans

		// Execute build aquifer (no hex position needed - uses tile queue)
		err = standardProjectService.BuildAquifer(ctx, game.ID, playerID)
		assert.NoError(t, err)

		// Verify player resources and TR
		// Get updated player directly from repository
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)

		player := updatedPlayer
		assert.Equal(t, initialCredits-expectedCost, player.Resources.Credits)
		assert.Equal(t, initialTR+1, player.TerraformRating)

		// Verify ocean count increase
		updatedParams, err := gameService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		assert.Equal(t, initialOceans+1, updatedParams.Oceans)
	})

	// Invalid hex position test removed - hex positions now handled by tile queue system
}

func TestStandardProjectService_PlantGreenery(t *testing.T) {
	standardProjectService, gameService, _, playerRepo, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	t.Run("Successful plant greenery", func(t *testing.T) {
		initialCredits := 100
		initialTR := 20
		expectedCost := 23

		// Get initial oxygen level
		initialParams, err := gameService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		initialOxygen := initialParams.Oxygen

		// Execute plant greenery (no hex position needed - uses tile queue)
		err = standardProjectService.PlantGreenery(ctx, game.ID, playerID)
		assert.NoError(t, err)

		// Verify player resources and TR
		// Get updated player directly from repository
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)

		player := updatedPlayer
		assert.Equal(t, initialCredits-expectedCost, player.Resources.Credits)
		assert.Equal(t, initialTR+1, player.TerraformRating)

		// Verify oxygen level increase
		updatedParams, err := gameService.GetGlobalParameters(ctx, game.ID)
		require.NoError(t, err)
		assert.Equal(t, initialOxygen+1, updatedParams.Oxygen)
	})

	// Invalid hex position test removed - hex positions now handled by tile queue system
}

func TestStandardProjectService_BuildCity(t *testing.T) {
	standardProjectService, _, _, playerRepo, game, playerID := setupStandardProjectServiceTest(t)
	ctx := context.Background()

	t.Run("Successful build city", func(t *testing.T) {
		initialCredits := 100
		initialCreditProduction := 1
		expectedCost := 25

		// Execute build city (no hex position needed - uses tile queue)
		err := standardProjectService.BuildCity(ctx, game.ID, playerID)
		assert.NoError(t, err)

		// Verify player resources and production
		// Get updated player directly from repository
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, playerID)
		require.NoError(t, err)

		player := updatedPlayer
		assert.Equal(t, initialCredits-expectedCost, player.Resources.Credits)
		assert.Equal(t, initialCreditProduction+1, player.Production.Credits)
	})

}

func TestPlayer_CanAffordStandardProject(t *testing.T) {
	player := &model.Player{
		Resources: model.Resources{Credits: 20},
	}

	tests := []struct {
		name     string
		project  model.StandardProject
		expected bool
	}{
		{"Can afford sell patents", model.StandardProjectSellPatents, true}, // 0 cost
		{"Can afford power plant", model.StandardProjectPowerPlant, true},   // 11 cost
		{"Can afford asteroid", model.StandardProjectAsteroid, true},        // 14 cost
		{"Can afford aquifer", model.StandardProjectAquifer, true},          // 18 cost
		{"Cannot afford greenery", model.StandardProjectGreenery, false},    // 23 cost
		{"Cannot afford city", model.StandardProjectCity, false},            // 25 cost
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := createTestPlayerService()
			result := svc.CanAffordStandardProject(player, tt.project)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlayer_HasCardsToSell(t *testing.T) {
	player := &model.Player{
		Cards: []string{"card1", "card2", "card3"},
	}

	tests := []struct {
		name     string
		count    int
		expected bool
	}{
		{"Can sell 1 card", 1, true},
		{"Can sell 3 cards", 3, true},
		{"Cannot sell 4 cards", 4, false},
		{"Cannot sell 0 cards", 0, false},
		{"Cannot sell negative cards", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := createTestPlayerService()
			result := svc.HasCardsToSell(player, tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlayer_GetMaxCardsToSell(t *testing.T) {
	tests := []struct {
		name     string
		cards    []string
		expected int
	}{
		{"No cards", []string{}, 0},
		{"Three cards", []string{"card1", "card2", "card3"}, 3},
		{"Many cards", []string{"c1", "c2", "c3", "c4", "c5", "c6", "c7"}, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &model.Player{Cards: tt.cards}
			svc := createTestPlayerService()
			result := svc.GetMaxCardsToSell(player)
			assert.Equal(t, tt.expected, result)
		})
	}
}
