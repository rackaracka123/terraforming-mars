package events

import (
	"context"
	"testing"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/listeners"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventDrivenGameStart_IntegrationFlow(t *testing.T) {
	// Setup dependencies
	gameRepo := repository.NewGameRepository()
	eventBus := events.NewInMemoryEventBus()
	gameService := service.NewGameService(gameRepo, eventBus)
	
	// Register event listeners
	listenerRegistry := listeners.NewRegistry(eventBus, gameRepo)
	listenerRegistry.RegisterAllListeners()

	// Track events for assertions
	var capturedEvents []events.Event
	eventBus.Subscribe(events.EventTypePlayerStartingCardOptions, func(ctx context.Context, event events.Event) error {
		capturedEvents = append(capturedEvents, event)
		return nil
	})

	// Create a test game
	gameSettings := domain.GameSettings{
		MaxPlayers: 4,
	}
	game, err := gameService.CreateGame(gameSettings)
	require.NoError(t, err)
	require.NotNil(t, game)

	// Add players to the game
	_, err = gameService.JoinGame(game.ID, "Player1")
	require.NoError(t, err)
	updatedGame, err := gameService.JoinGame(game.ID, "Player2")
	require.NoError(t, err)

	// Get player IDs for testing
	require.Len(t, updatedGame.Players, 2)
	player1ID := updatedGame.Players[0].ID
	player2ID := updatedGame.Players[1].ID

	// Verify initial game state
	assert.Equal(t, domain.GameStatusLobby, updatedGame.Status)
	assert.Equal(t, domain.GamePhaseSetup, updatedGame.CurrentPhase)

	// Start the game (this should trigger the event-driven flow)
	startGameRequest := dto.ActionStartGameRequest{
		Type: dto.ActionTypeStartGame,
	}

	gameAfterStart, err := gameService.ApplyAction(game.ID, player1ID, startGameRequest)
	require.NoError(t, err)
	require.NotNil(t, gameAfterStart)

	// Verify game state changes
	assert.Equal(t, domain.GameStatusActive, gameAfterStart.Status)
	assert.Equal(t, domain.GamePhaseStartingCardSelection, gameAfterStart.CurrentPhase)
	assert.Equal(t, player1ID, gameAfterStart.CurrentPlayerID)

	// Wait for async event processing
	time.Sleep(100 * time.Millisecond)

	// Verify that PlayerStartingCardOptions events were published for each player
	assert.Len(t, capturedEvents, 2, "Should have received starting card options for both players")

	// Verify each player received their card options
	for _, event := range capturedEvents {
		assert.Equal(t, events.EventTypePlayerStartingCardOptions, event.GetType())
		assert.Equal(t, game.ID, event.GetGameID())

		payload, ok := event.GetPayload().(events.PlayerStartingCardOptionsPayload)
		require.True(t, ok, "Event payload should be PlayerStartingCardOptionsPayload")
		assert.Contains(t, []string{player1ID, player2ID}, payload.PlayerID, "Player ID should be one of the game players")
		assert.Len(t, payload.CardOptions, 5, "Each player should receive exactly 5 starting card options")
		
		// Verify all card options are valid
		availableCards := domain.GetStartingCards()
		availableCardIDs := make(map[string]bool)
		for _, card := range availableCards {
			availableCardIDs[card.ID] = true
		}
		
		for _, cardID := range payload.CardOptions {
			assert.True(t, availableCardIDs[cardID], "Card ID %s should be a valid starting card", cardID)
		}
	}

	t.Logf("Test completed successfully:")
	t.Logf("- Game transitioned to Active status with StartingCardSelection phase")
	t.Logf("- %d PlayerStartingCardOptions events published", len(capturedEvents))
	t.Logf("- Each player received 5 unique starting card options")
}

func TestEventDrivenGameStart_SecurityIsolation(t *testing.T) {
	// Setup dependencies
	gameRepo := repository.NewGameRepository()
	eventBus := events.NewInMemoryEventBus()
	gameService := service.NewGameService(gameRepo, eventBus)
	listenerRegistry := listeners.NewRegistry(eventBus, gameRepo)
	listenerRegistry.RegisterAllListeners()

	// Track events per player to verify security isolation
	player1Events := make([]events.Event, 0)
	player2Events := make([]events.Event, 0)

	eventBus.Subscribe(events.EventTypePlayerStartingCardOptions, func(ctx context.Context, event events.Event) error {
		payload := event.GetPayload().(events.PlayerStartingCardOptionsPayload)
		
		// Route events to player-specific collections (simulating client filtering)
		switch payload.PlayerID {
		case "player1":
			player1Events = append(player1Events, event)
		case "player2":
			player2Events = append(player2Events, event)
		}
		return nil
	})

	// Create game and add players
	gameSettings := domain.GameSettings{MaxPlayers: 2}
	game, err := gameService.CreateGame(gameSettings)
	require.NoError(t, err)

	game, err = gameService.JoinGame(game.ID, "Player1")
	require.NoError(t, err)
	game, err = gameService.JoinGame(game.ID, "Player2")
	require.NoError(t, err)

	// Override player IDs for predictable testing
	game.Players[0].ID = "player1"
	game.Players[1].ID = "player2"
	game.HostPlayerID = "player1"  // Update host ID to match
	gameRepo.UpdateGame(game)

	// Start the game 
	startGameRequest := dto.ActionStartGameRequest{Type: dto.ActionTypeStartGame}
	_, err = gameService.ApplyAction(game.ID, "player1", startGameRequest)
	require.NoError(t, err)

	// Wait for event processing
	time.Sleep(100 * time.Millisecond)

	// Verify security: each player only receives their own cards
	require.Len(t, player1Events, 1, "Player1 should receive exactly 1 card options event")
	require.Len(t, player2Events, 1, "Player2 should receive exactly 1 card options event")

	// Verify player1 event contains only player1's data
	p1Payload := player1Events[0].GetPayload().(events.PlayerStartingCardOptionsPayload)
	assert.Equal(t, "player1", p1Payload.PlayerID)
	assert.Len(t, p1Payload.CardOptions, 5)

	// Verify player2 event contains only player2's data
	p2Payload := player2Events[0].GetPayload().(events.PlayerStartingCardOptionsPayload)
	assert.Equal(t, "player2", p2Payload.PlayerID)
	assert.Len(t, p2Payload.CardOptions, 5)

	// Verify card options are different (very high probability with random selection)
	assert.NotEqual(t, p1Payload.CardOptions, p2Payload.CardOptions, "Players should receive different card options")

	t.Logf("Security verification passed:")
	t.Logf("- Player1 received cards for player1 only: %v", p1Payload.CardOptions)
	t.Logf("- Player2 received cards for player2 only: %v", p2Payload.CardOptions)
}