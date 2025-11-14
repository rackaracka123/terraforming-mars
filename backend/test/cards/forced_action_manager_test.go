package cards

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForcedActionManager_InventrixCardDrawAction(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Initialize repositories
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Load cards for testing
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err)

	// Create forced action manager and subscribe
	manager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	manager.SubscribeToPhaseChanges()

	// Create a test game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	err = gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
	require.NoError(t, err)

	err = gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseStartingCardSelection)
	require.NoError(t, err)

	// Create a test player with Inventrix forced action
	player := model.Player{
		ID:              "player1",
		Name:            "TestPlayer",
		Resources:       model.Resources{Credits: 100},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
		ForcedFirstAction: &model.ForcedFirstAction{
			ActionType:    "card_draw",
			CorporationID: "B05", // Inventrix
			Completed:     false,
			Description:   "Draw 3 cards",
		},
	}
	err = playerRepo.Create(ctx, game.ID, player)
	require.NoError(t, err)

	// Set player as current turn
	err = gameRepo.UpdateCurrentTurn(ctx, game.ID, &player.ID)
	require.NoError(t, err)

	// Create a deck with some cards for the game
	startingCards, err := cardRepo.GetStartingCardPool(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(startingCards), 10, "Need at least 10 starting cards")

	cardsForDeck := make([]model.Card, 10)
	for i := 0; i < 10; i++ {
		cardsForDeck[i] = startingCards[i]
	}

	err = cardDeckRepo.InitializeDeck(ctx, game.ID, cardsForDeck)
	require.NoError(t, err)

	// Publish phase change event to Action phase (this should trigger the forced action)
	// Events are synchronous, so the forced action will trigger immediately
	err = gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseAction)
	require.NoError(t, err)

	// Verify that a PendingCardDrawSelection was created
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, "player1")
	require.NoError(t, err)

	assert.NotNil(t, updatedPlayer.PendingCardDrawSelection, "Expected pending card draw selection to be created")
	if updatedPlayer.PendingCardDrawSelection != nil {
		assert.Equal(t, 3, len(updatedPlayer.PendingCardDrawSelection.AvailableCards), "Expected 3 cards to be drawn")
		assert.Equal(t, "B05", updatedPlayer.PendingCardDrawSelection.Source, "Expected source to be B05 (Inventrix)")
		assert.Equal(t, 3, updatedPlayer.PendingCardDrawSelection.FreeTakeCount, "Expected 3 free cards")
	}
}

func TestForcedActionManager_TharsisCityPlacementAction(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Initialize repositories
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Load cards for testing
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err)

	// Initialize board service to create a valid board
	boardService := service.NewBoardService()

	// Create forced action manager and subscribe
	manager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	manager.SubscribeToPhaseChanges()

	// Create a test game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Initialize board with default tiles
	board := boardService.GenerateDefaultBoard()
	err = gameRepo.UpdateBoard(ctx, game.ID, board)
	require.NoError(t, err)

	err = gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
	require.NoError(t, err)

	err = gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseStartingCardSelection)
	require.NoError(t, err)

	// Create a test player with Tharsis Republic forced action
	player := model.Player{
		ID:              "player1",
		Name:            "TestPlayer",
		Resources:       model.Resources{Credits: 100},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
		ForcedFirstAction: &model.ForcedFirstAction{
			ActionType:    "city_placement",
			CorporationID: "B08", // Tharsis Republic
			Completed:     false,
			Description:   "Place a city tile",
		},
	}
	err = playerRepo.Create(ctx, game.ID, player)
	require.NoError(t, err)

	// Set player as current turn
	err = gameRepo.UpdateCurrentTurn(ctx, game.ID, &player.ID)
	require.NoError(t, err)

	// Publish phase change event to Action phase (events are synchronous)
	err = gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseAction)
	require.NoError(t, err)

	// Verify that a PendingTileSelection was created
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, "player1")
	require.NoError(t, err)

	assert.NotNil(t, updatedPlayer.PendingTileSelection, "Expected pending tile selection to be created")
	if updatedPlayer.PendingTileSelection != nil {
		assert.Equal(t, "city", updatedPlayer.PendingTileSelection.TileType, "Expected tile type to be city")
		assert.Equal(t, "B08", updatedPlayer.PendingTileSelection.Source, "Expected source to be B08 (Tharsis Republic)")
		assert.Greater(t, len(updatedPlayer.PendingTileSelection.AvailableHexes), 0, "Expected available hexes")
	}
}

func TestForcedActionManager_NoActionForNonActionPhase(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Initialize repositories
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Load cards for testing
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err)

	// Create forced action manager and subscribe
	manager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	manager.SubscribeToPhaseChanges()

	// Create a test game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	err = gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
	require.NoError(t, err)

	err = gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseStartingCardSelection)
	require.NoError(t, err)

	// Create a test player with forced action
	player := model.Player{
		ID:              "player1",
		Name:            "TestPlayer",
		Resources:       model.Resources{Credits: 100},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
		ForcedFirstAction: &model.ForcedFirstAction{
			ActionType:    "card_draw",
			CorporationID: "B05", // Inventrix
			Completed:     false,
			Description:   "Draw 3 cards",
		},
	}
	err = playerRepo.Create(ctx, game.ID, player)
	require.NoError(t, err)

	// Set player as current turn
	err = gameRepo.UpdateCurrentTurn(ctx, game.ID, &player.ID)
	require.NoError(t, err)

	// Create a deck with some cards
	startingCards, err := cardRepo.GetStartingCardPool(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(startingCards), 10, "Need at least 10 starting cards")

	cardsForDeck := make([]model.Card, 10)
	for i := 0; i < 10; i++ {
		cardsForDeck[i] = startingCards[i]
	}

	err = cardDeckRepo.InitializeDeck(ctx, game.ID, cardsForDeck)
	require.NoError(t, err)

	// Publish phase change event to Production phase (NOT Action phase)
	err = gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseProductionAndCardDraw)
	require.NoError(t, err)

	// Verify that NO PendingCardDrawSelection was created
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, "player1")
	require.NoError(t, err)

	assert.Nil(t, updatedPlayer.PendingCardDrawSelection, "Expected NO pending card draw selection for non-Action phase")
}

func TestForcedActionManager_NoActionWhenAlreadyCompleted(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Initialize repositories
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Load cards for testing
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err)

	// Create forced action manager and subscribe
	manager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	manager.SubscribeToPhaseChanges()

	// Create a test game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	err = gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
	require.NoError(t, err)

	err = gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseStartingCardSelection)
	require.NoError(t, err)

	// Create a test player with COMPLETED forced action
	player := model.Player{
		ID:              "player1",
		Name:            "TestPlayer",
		Resources:       model.Resources{Credits: 100},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
		ForcedFirstAction: &model.ForcedFirstAction{
			ActionType:    "card_draw",
			CorporationID: "inventrix",
			Completed:     true, // Already completed
			Description:   "Draw 3 cards",
		},
	}
	err = playerRepo.Create(ctx, game.ID, player)
	require.NoError(t, err)

	// Set player as current turn
	err = gameRepo.UpdateCurrentTurn(ctx, game.ID, &player.ID)
	require.NoError(t, err)

	// Create a deck with some cards
	startingCards, err := cardRepo.GetStartingCardPool(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(startingCards), 10, "Need at least 10 starting cards")

	cardsForDeck := make([]model.Card, 10)
	for i := 0; i < 10; i++ {
		cardsForDeck[i] = startingCards[i]
	}

	err = cardDeckRepo.InitializeDeck(ctx, game.ID, cardsForDeck)
	require.NoError(t, err)

	// Publish phase change event to Action phase
	err = gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseAction)
	require.NoError(t, err)

	// Verify that NO PendingCardDrawSelection was created (already completed)
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, "player1")
	require.NoError(t, err)

	assert.Nil(t, updatedPlayer.PendingCardDrawSelection, "Expected NO pending card draw selection when already completed")
}

func TestForcedActionManager_MarkComplete(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Initialize repositories
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Load cards for testing
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err)

	// Create forced action manager
	manager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)

	// Create a test game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Create a test player with forced action
	player := model.Player{
		ID:              "player1",
		Name:            "TestPlayer",
		Resources:       model.Resources{Credits: 100},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
		ForcedFirstAction: &model.ForcedFirstAction{
			ActionType:    "card_draw",
			CorporationID: "B05", // Inventrix
			Completed:     false,
			Description:   "Draw 3 cards",
		},
	}
	err = playerRepo.Create(ctx, game.ID, player)
	require.NoError(t, err)

	// Mark the forced action as complete
	err = manager.MarkComplete(ctx, game.ID, "player1")
	require.NoError(t, err)

	// Verify that the forced action is now marked as completed
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, "player1")
	require.NoError(t, err)

	require.NotNil(t, updatedPlayer.ForcedFirstAction, "Expected forced action to still exist")
	assert.True(t, updatedPlayer.ForcedFirstAction.Completed, "Expected forced action to be marked as completed")
}

func TestForcedActionManager_MarkCompleteWithNoForcedAction(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Initialize repositories
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Load cards for testing
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err)

	// Create forced action manager
	manager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)

	// Create a test game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Create a test player WITHOUT forced action
	player := model.Player{
		ID:                "player1",
		Name:              "TestPlayer",
		Resources:         model.Resources{Credits: 100},
		Production:        model.Production{Credits: 1},
		TerraformRating:   20,
		IsConnected:       true,
		ForcedFirstAction: nil,
	}
	err = playerRepo.Create(ctx, game.ID, player)
	require.NoError(t, err)

	// Try to mark forced action as complete (should not error, just no-op)
	err = manager.MarkComplete(ctx, game.ID, "player1")
	assert.NoError(t, err, "MarkComplete should not error when player has no forced action")

	// Verify player state unchanged
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, "player1")
	require.NoError(t, err)
	assert.Nil(t, updatedPlayer.ForcedFirstAction)
}

func TestForcedActionManager_NoActionWhenPlayerNotCurrentTurn(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Initialize repositories
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Load cards for testing
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err)

	// Create forced action manager and subscribe
	manager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	manager.SubscribeToPhaseChanges()

	// Create a test game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	err = gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
	require.NoError(t, err)

	err = gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseStartingCardSelection)
	require.NoError(t, err)

	// Create two players
	player1 := model.Player{
		ID:              "player1",
		Name:            "Player1",
		Resources:       model.Resources{Credits: 100},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
		ForcedFirstAction: &model.ForcedFirstAction{
			ActionType:    "card_draw",
			CorporationID: "B05", // Inventrix
			Completed:     false,
			Description:   "Draw 3 cards",
		},
	}
	err = playerRepo.Create(ctx, game.ID, player1)
	require.NoError(t, err)

	player2 := model.Player{
		ID:              "player2",
		Name:            "Player2",
		Resources:       model.Resources{Credits: 100},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
	}
	err = playerRepo.Create(ctx, game.ID, player2)
	require.NoError(t, err)

	// Set player2 as current turn (NOT player1 who has the forced action)
	err = gameRepo.UpdateCurrentTurn(ctx, game.ID, &player2.ID)
	require.NoError(t, err)

	// Create a deck
	startingCards, err := cardRepo.GetStartingCardPool(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(startingCards), 10, "Need at least 10 starting cards")

	cardsForDeck := make([]model.Card, 10)
	for i := 0; i < 10; i++ {
		cardsForDeck[i] = startingCards[i]
	}

	err = cardDeckRepo.InitializeDeck(ctx, game.ID, cardsForDeck)
	require.NoError(t, err)

	// Publish phase change event to Action phase
	err = gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseAction)
	require.NoError(t, err)

	// Verify that NO PendingCardDrawSelection was created for player1
	updatedPlayer1, err := playerRepo.GetByID(ctx, game.ID, "player1")
	require.NoError(t, err)

	assert.Nil(t, updatedPlayer1.PendingCardDrawSelection, "Expected NO pending card draw selection when not current turn")
}
