package service_test

import (
	"context"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCardSelectionFlow(t *testing.T) {
	// Setup
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Load card data for testing
	err := cardRepo.LoadCards(context.Background())
	require.NoError(t, err, "Should load card data for testing")

	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber, forcedActionManager)
	gameService := service.NewGameService(gameRepo, playerRepo, cardRepo, cardService.(*service.CardServiceImpl), cardDeckRepo, boardService, sessionManager)

	// EventBus tracking removed - using SessionManager for state updates

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

	// EventBus removed - no event waiting needed, operations are synchronous

	// Verify starting cards are distributed to player
	players, err := playerRepo.ListByGameID(ctx, game.ID)
	require.NoError(t, err)
	require.Len(t, players, 1)

	player := players[0]
	require.NotNil(t, player.SelectStartingCardsPhase, "Player should have starting card selection data")
	require.Len(t, player.SelectStartingCardsPhase.AvailableCards, 10, "Player should have 10 available starting cards")

	t.Logf("✅ Starting cards available to player: %v", player.SelectStartingCardsPhase.AvailableCards)

	// Test card selection - select first 2 cards and first corporation
	selectedCardIDs := player.SelectStartingCardsPhase.AvailableCards[:2]
	selectedCorporationID := player.SelectStartingCardsPhase.AvailableCorporations[0]
	err = cardService.OnSelectStartingCards(ctx, game.ID, playerID, selectedCardIDs, selectedCorporationID)
	require.NoError(t, err)
	t.Logf("✅ Cards selected successfully: %v", selectedCardIDs)

	// Verify player has the selected cards
	player, err = playerRepo.GetByID(ctx, game.ID, playerID)
	require.NoError(t, err)
	require.Equal(t, selectedCardIDs, player.Cards, "Player should have the selected cards")

	t.Log("🎉 Card selection flow completed successfully!")
}
