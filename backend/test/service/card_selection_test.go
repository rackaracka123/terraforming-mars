package service_test

import (
	"context"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"
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
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, eventBus, cardDeckRepo, sessionManager)
	sessionManager := test.NewMockSessionManager()
	gameService := service.NewGameService(gameRepo, playerRepo, cardService.(*service.CardServiceImpl), eventBus, sessionManager)

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

	// Verify starting cards are distributed to player
	players, err := playerRepo.ListByGameID(ctx, game.ID)
	require.NoError(t, err)
	require.Len(t, players, 1)

	player := players[0]
	require.NotNil(t, player.StartingSelection, "Player should have starting card selection data")
	require.Len(t, player.StartingSelection, 10, "Player should have 10 available starting cards")

	t.Logf("✅ Starting cards available to player: %v", player.StartingSelection)

	// Test card selection - select first 2 cards
	selectedCardIDs := player.StartingSelection[:2]
	err = cardService.SelectStartingCards(ctx, game.ID, playerID, selectedCardIDs)
	require.NoError(t, err)
	t.Logf("✅ Cards selected successfully: %v", selectedCardIDs)

	// Verify player has the selected cards
	player, err = playerRepo.GetByID(ctx, game.ID, playerID)
	require.NoError(t, err)
	require.Equal(t, selectedCardIDs, player.Cards, "Player should have the selected cards")

	t.Log("🎉 Card selection flow completed successfully!")
}
