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
	cardDataService := service.NewCardDataService()

	// Load card data for testing
	err := cardDataService.LoadCards()
	require.NoError(t, err, "Should load card data for testing")

	cardDeckRepo := repository.NewCardDeckRepository()
	cardSelectionRepo := repository.NewCardSelectionRepository()
	cardService := service.NewCardService(gameRepo, playerRepo, cardDataService, eventBus, cardDeckRepo, cardSelectionRepo)
	gameService := service.NewGameService(gameRepo, playerRepo, cardService.(*service.CardServiceImpl), eventBus)

	// Track events
	var receivedEvents []events.Event
	eventBus.Subscribe(events.EventTypeCardDealt, func(ctx context.Context, event events.Event) error {
		receivedEvents = append(receivedEvents, event)
		t.Logf("✅ Received event: %s", event.GetType())
		return nil
	})
	t.Logf("📬 Subscribed to event: %s", events.EventTypeCardDealt)

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

	// Wait for the event to be processed by the async worker pool
	maxWaitTime := 100 * time.Millisecond
	waitInterval := 5 * time.Millisecond
	waited := time.Duration(0)

	for len(receivedEvents) == 0 && waited < maxWaitTime {
		time.Sleep(waitInterval)
		waited += waitInterval
	}

	// Verify that starting card options event was published
	require.Len(t, receivedEvents, 1, "Should have received exactly 1 starting card options event after waiting %v", waited)

	event := receivedEvents[0]
	require.Equal(t, events.EventTypeCardDealt, event.GetType())

	payload := event.GetPayload().(events.CardDealtEventData)
	require.Equal(t, game.ID, payload.GameID)
	require.Equal(t, playerID, payload.PlayerID)
	require.Len(t, payload.CardOptions, 4, "Should have received 4 card options")

	t.Logf("✅ Card options received: %v", payload.CardOptions)

	// Test card selection
	selectedCards := payload.CardOptions[:2] // Select first 2 cards
	err = cardService.SelectStartingCards(ctx, game.ID, playerID, selectedCards)
	require.NoError(t, err)
	t.Logf("✅ Cards selected successfully: %v", selectedCards)

	// Verify player has the selected cards
	player, err := playerRepo.GetByID(ctx, game.ID, playerID)
	require.NoError(t, err)
	require.Equal(t, selectedCards, player.Cards, "Player should have the selected cards")

	t.Log("🎉 Card selection flow completed successfully!")
}
