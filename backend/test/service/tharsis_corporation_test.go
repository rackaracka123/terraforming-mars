package service

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTharsisRepublic_CityPlacement_ProductionIncrease tests that Tharsis Republic's passive effect
// increases M€ production by 1 when any city tile is placed on Mars
func TestTharsisRepublic_CityPlacement_ProductionIncrease(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Setup repositories and services
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber)
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager, boardService, tileService)
	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, sessionManager, tileService)

	// Load cards
	require.NoError(t, cardRepo.LoadCards(ctx))

	// Create test game in starting card selection phase
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Initialize board with default tiles
	board := boardService.GenerateDefaultBoard()
	gameRepo.UpdateBoard(ctx, game.ID, board)

	gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
	gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseStartingCardSelection)

	// Create test player
	player := model.Player{
		ID:              "player1",
		Name:            "Test Player",
		Resources:       model.Resources{Credits: 100},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
	}
	playerRepo.Create(ctx, game.ID, player)

	// Get available starting cards and corporations
	startingCards, _ := cardRepo.GetStartingCardPool(ctx)
	corporations, _ := cardRepo.GetCorporations(ctx)
	require.GreaterOrEqual(t, len(startingCards), 4, "Should have at least 4 starting cards")
	require.GreaterOrEqual(t, len(corporations), 12, "Should have at least 12 corporations")

	// Find Tharsis Republic (B08)
	var tharsisID string
	for _, corp := range corporations {
		if corp.ID == "B08" {
			tharsisID = corp.ID
			break
		}
	}
	require.NotEmpty(t, tharsisID, "Tharsis Republic (B08) should be available")

	// Set up starting card selection phase for player
	availableCardIDs := []string{
		startingCards[0].ID,
		startingCards[1].ID,
		startingCards[2].ID,
		startingCards[3].ID,
	}
	corporationIDs := []string{tharsisID, corporations[1].ID}

	playerRepo.UpdateSelectStartingCardsPhase(ctx, game.ID, player.ID, &model.SelectStartingCardsPhase{
		AvailableCards:        availableCardIDs,
		AvailableCorporations: corporationIDs,
	})

	// Select Tharsis Republic and some starting cards
	selectedCardIDs := []string{startingCards[0].ID, startingCards[1].ID}
	err = cardService.OnSelectStartingCards(ctx, game.ID, player.ID, selectedCardIDs, tharsisID)
	require.NoError(t, err, "Should successfully select Tharsis Republic")

	// Verify initial production (should still be 1 at this point)
	playerAfterSelection, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, playerAfterSelection.Production.Credits, "Initial M€ production should be 1")

	t.Logf("✅ Tharsis Republic selected, initial production: %d", playerAfterSelection.Production.Credits)

	// Change game phase to action phase so we can place tiles
	gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseAction)
	gameRepo.SetCurrentTurn(ctx, game.ID, &player.ID)

	// Give player sufficient resources for city placement
	playerRepo.UpdateResources(ctx, game.ID, player.ID, model.Resources{Credits: 100})

	// Place a city using standard project (costs 25 M€)
	err = standardProjectService.BuildCity(ctx, game.ID, player.ID)
	require.NoError(t, err, "Should successfully initiate city placement")

	// Player should now have a pending tile selection
	playerAfterBuild, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	require.NoError(t, err)
	require.NotNil(t, playerAfterBuild.PendingTileSelection, "Player should have pending tile selection")

	t.Log("🏗️ City build initiated, selecting tile location...")

	// Select a tile location (0, 0, 0) - center of the map
	err = playerService.OnTileSelected(ctx, game.ID, player.ID, model.HexPosition{Q: 0, R: 0, S: 0})
	require.NoError(t, err, "Should successfully place city tile")

	// Verify production increased by 1 due to Tharsis passive effect
	playerFinal, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	require.NoError(t, err)

	expectedProduction := 2 // 1 (initial) + 1 (from Tharsis city placement effect)
	assert.Equal(t, expectedProduction, playerFinal.Production.Credits,
		"M€ production should increase by 1 when city is placed (Tharsis passive effect)")

	t.Logf("✅ City placed! Final M€ production: %d (expected: %d)", playerFinal.Production.Credits, expectedProduction)
	t.Log("🎉 Tharsis Republic city placement production bonus test passed!")
}

// TestTharsisRepublic_OtherPlayerCityPlacement tests that when another player places a city,
// Tharsis Republic owner gets +1 M€ production BUT NOT the immediate +3 M€ bonus
func TestTharsisRepublic_OtherPlayerCityPlacement(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Setup repositories and services
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber)
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager, boardService, tileService)
	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, sessionManager, tileService)

	// Load cards
	require.NoError(t, cardRepo.LoadCards(ctx))

	// Create test game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Initialize board with default tiles
	board := boardService.GenerateDefaultBoard()
	gameRepo.UpdateBoard(ctx, game.ID, board)

	gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
	gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseStartingCardSelection)

	// Create Player 1 (Tharsis owner)
	player1 := model.Player{
		ID:              "player1",
		Name:            "Tharsis Player",
		Resources:       model.Resources{Credits: 100},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
	}
	playerRepo.Create(ctx, game.ID, player1)

	// Create Player 2 (other player)
	player2 := model.Player{
		ID:              "player2",
		Name:            "Other Player",
		Resources:       model.Resources{Credits: 100},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
	}
	playerRepo.Create(ctx, game.ID, player2)

	// Get available cards
	startingCards, _ := cardRepo.GetStartingCardPool(ctx)
	corporations, _ := cardRepo.GetCorporations(ctx)
	require.GreaterOrEqual(t, len(corporations), 12)

	// Find Tharsis Republic
	var tharsisID string
	for _, corp := range corporations {
		if corp.ID == "B08" {
			tharsisID = corp.ID
			break
		}
	}
	require.NotEmpty(t, tharsisID, "Tharsis Republic should be available")

	// Player 1 selects Tharsis Republic
	availableCardIDs := []string{startingCards[0].ID, startingCards[1].ID}
	corporationIDs := []string{tharsisID, corporations[1].ID}
	playerRepo.UpdateSelectStartingCardsPhase(ctx, game.ID, player1.ID, &model.SelectStartingCardsPhase{
		AvailableCards:        availableCardIDs,
		AvailableCorporations: corporationIDs,
	})
	err = cardService.OnSelectStartingCards(ctx, game.ID, player1.ID, availableCardIDs, tharsisID)
	require.NoError(t, err)

	// Player 2 selects a different corporation
	playerRepo.UpdateSelectStartingCardsPhase(ctx, game.ID, player2.ID, &model.SelectStartingCardsPhase{
		AvailableCards:        availableCardIDs,
		AvailableCorporations: []string{corporations[1].ID, corporations[2].ID},
	})
	err = cardService.OnSelectStartingCards(ctx, game.ID, player2.ID, availableCardIDs, corporations[1].ID)
	require.NoError(t, err)

	t.Log("✅ Player 1 has Tharsis, Player 2 has different corporation")

	// Record initial state of Player 1 (Tharsis owner)
	player1Before, err := playerRepo.GetByID(ctx, game.ID, player1.ID)
	require.NoError(t, err)
	initialProduction := player1Before.Production.Credits
	initialCredits := player1Before.Resources.Credits

	t.Logf("📊 Tharsis player before: Production=%d, Credits=%d", initialProduction, initialCredits)

	// Change to action phase
	gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseAction)
	gameRepo.SetCurrentTurn(ctx, game.ID, &player2.ID)

	// Player 2 builds a city
	err = standardProjectService.BuildCity(ctx, game.ID, player2.ID)
	require.NoError(t, err)

	// Player 2 selects tile location
	err = playerService.OnTileSelected(ctx, game.ID, player2.ID, model.HexPosition{Q: 0, R: 0, S: 0})
	require.NoError(t, err)

	t.Log("🏗️ Player 2 placed a city on Mars")

	// Check Player 1's state after Player 2 placed city
	player1After, err := playerRepo.GetByID(ctx, game.ID, player1.ID)
	require.NoError(t, err)

	// Player 1 should have +1 production (Tharsis passive: any city on Mars)
	expectedProduction := initialProduction + 1
	assert.Equal(t, expectedProduction, player1After.Production.Credits,
		"Tharsis owner should get +1 M€ production when ANY player places city")

	// Player 1 should NOT get the immediate +3 M€ (only triggers when Tharsis player places city)
	assert.Equal(t, initialCredits, player1After.Resources.Credits,
		"Tharsis owner should NOT get immediate +3 M€ when OTHER player places city")

	t.Logf("✅ Tharsis player after: Production=%d (expected %d), Credits=%d (expected %d - no bonus)",
		player1After.Production.Credits, expectedProduction,
		player1After.Resources.Credits, initialCredits)
	t.Log("🎉 Tharsis 'any city' passive effect test passed!")
}
