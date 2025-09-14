package service_test

import (
	"context"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCardSelectionFlow(t *testing.T) {
	// Setup
	ctx := context.Background()
	eventBus := events.NewInMemoryEventBus()

	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Load card data for testing
	err := cardRepo.LoadCards(context.Background())
	require.NoError(t, err, "Should load card data for testing")

	cardDeckRepo := repository.NewCardDeckRepository()
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, eventBus, cardDeckRepo)
	gameService := service.NewGameService(gameRepo, playerRepo, cardService.(*service.CardServiceImpl), eventBus)

	// Track game updated events (consolidated state)
	var receivedEvents []events.Event
	eventBus.Subscribe(events.EventTypeGameUpdated, func(ctx context.Context, event events.Event) error {
		receivedEvents = append(receivedEvents, event)
		t.Logf("✅ Received event: %s", event.GetType())
		return nil
	})
	t.Logf("📬 Subscribed to event: %s", events.EventTypeGameUpdated)

	// Create game
	game, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)
	t.Logf("Created game: %s", game.ID)

	// Join player
	updatedGame, err := gameService.JoinGame(ctx, game.ID, "TestPlayer")
	require.NoError(t, err)
	require.Len(t, updatedGame.PlayerIDs, 1)
	playerID := updatedGame.PlayerIDs[0]
	t.Logf("Player joined: %s", playerID)

	// Start game (this should distribute starting cards)
	err = gameService.StartGame(ctx, game.ID, playerID)
	require.NoError(t, err)
	t.Log("Game started")

	// Wait for game updated event to be processed
	maxWaitTime := 100 * time.Millisecond
	waitInterval := 5 * time.Millisecond
	waited := time.Duration(0)

	for len(receivedEvents) == 0 && waited < maxWaitTime {
		time.Sleep(waitInterval)
		waited += waitInterval
	}

	// Verify that game updated event was published
	require.GreaterOrEqual(t, len(receivedEvents), 1, "Should have received at least 1 game updated event after waiting %v", waited)

	// Verify card selection is now stored in player's production field
	players, err := playerRepo.ListByGameID(ctx, game.ID)
	require.NoError(t, err)
	require.Len(t, players, 1)

	player := players[0]
	require.NotNil(t, player.ProductionSelection, "Player should have production selection data")
	require.Len(t, player.ProductionSelection.AvailableCards, 4, "Player should have 4 available starting cards")
	require.False(t, player.ProductionSelection.SelectionComplete, "Player should not have completed selection yet")

	t.Logf("✅ Card options available in player production: %v", player.ProductionSelection.AvailableCards)

	// Test card selection - extract card IDs from Card objects
	selectedCardObjects := player.ProductionSelection.AvailableCards[:2] // Select first 2 cards
	selectedCards := make([]string, len(selectedCardObjects))
	for i, card := range selectedCardObjects {
		selectedCards[i] = card.ID
	}
	err = cardService.SelectStartingCards(ctx, game.ID, playerID, selectedCards)
	require.NoError(t, err)
	t.Logf("✅ Cards selected successfully: %v", selectedCards)

	// Verify player has the selected cards
	player, err = playerRepo.GetByID(ctx, game.ID, playerID)
	require.NoError(t, err)
	require.Equal(t, selectedCards, player.Cards, "Player should have the selected cards")

	t.Log("🎉 Card selection flow completed successfully!")
}
