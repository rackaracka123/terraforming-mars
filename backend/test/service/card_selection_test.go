package service_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/production"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"

	"github.com/stretchr/testify/require"
)

func TestCardSelectionFlow(t *testing.T) {
	// Setup
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	cardRepo := game.NewCardRepository()

	// Load card data for testing
	err := cardRepo.LoadCards(context.Background())
	require.NoError(t, err, "Should load card data for testing")

	cardDeckRepo := game.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)

	// Initialize game features
	tilesRepo := tiles.NewRepository(gameRepo, playerRepo)
	tilesBoardAdapter := tiles.NewBoardServiceAdapter(boardService)
	tilesFeature := tiles.NewService(tilesRepo, tilesBoardAdapter, eventBus)

	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tilesFeature, effectSubscriber, forcedActionManager)
	turnRepo := turn.NewRepository(gameRepo, playerRepo)
	turnFeature := turn.NewService(turnRepo)
	productionRepo := production.NewRepository(gameRepo, playerRepo, cardDeckRepo)
	productionFeature := production.NewService(productionRepo)

	_ = service.NewGameService(gameRepo, playerRepo, cardRepo, cardService.(*service.CardServiceImpl), cardDeckRepo, boardService, sessionManager, turnFeature, productionFeature, tilesFeature)
	lobbyService := lobby.NewService(gameRepo, playerRepo, cardRepo, cardService.(*service.CardServiceImpl), cardDeckRepo, boardService, sessionManager)

	// EventBus tracking removed - using SessionManager for state updates

	// Create game
	game, err := lobbyService.CreateGame(ctx, game.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)
	t.Logf("Created game: %s", game.ID)

	// Join player
	updatedGame, err := lobbyService.JoinGame(ctx, game.ID, "TestPlayer")
	require.NoError(t, err)
	require.Len(t, updatedGame.PlayerIDs, 1)
	playerID := updatedGame.PlayerIDs[0]
	t.Logf("Player joined: %s", playerID)

	// Start game (this should distribute starting cards)
	err = lobbyService.StartGame(ctx, game.ID, playerID)
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
