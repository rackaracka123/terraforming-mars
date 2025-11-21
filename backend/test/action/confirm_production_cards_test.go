package action_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/types"
	"terraforming-mars-backend/test"
)

func TestConfirmProductionCardsAction_Success(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Create repositories
	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)

	// Initialize mock session manager
	sessionMgr := test.NewMockSessionManager()

	// Create action (cardRepo not needed for these tests)
	confirmAction := action.NewConfirmProductionCardsAction(gameRepo, playerRepo, nil, sessionMgr)

	// Setup: Create game
	currentTurn := "player1"
	testGame := &game.Game{
		ID:           "test-game",
		Status:       game.GameStatusActive,
		CurrentPhase: game.GamePhaseProductionAndCardDraw,
		PlayerIDs:    []string{"player1"},
		Generation:   1,
		CurrentTurn:  &currentTurn,
		HostPlayerID: "player1",
	}
	err := gameRepo.Create(ctx, testGame)
	require.NoError(t, err)

	// Setup: Create player with production phase
	testPlayer := player.NewPlayer("TestPlayer")
	testPlayer.ID = "player1"
	testPlayer.Resources = types.Resources{
		Credits: 50, // Enough for 3 cards (9 MC)
	}
	testPlayer.ProductionPhase = &types.ProductionPhase{
		AvailableCards:    []string{"card1", "card2", "card3", "card4"},
		SelectionComplete: false,
		BeforeResources:   testPlayer.Resources,
		AfterResources:    testPlayer.Resources,
	}

	err = playerRepo.Create(ctx, "test-game", testPlayer)
	require.NoError(t, err)

	// Execute: Confirm selection of 2 cards
	selectedCards := []string{"card1", "card3"}
	err = confirmAction.Execute(ctx, "test-game", "player1", selectedCards)
	require.NoError(t, err)

	// Verify: Player has cards in hand
	updatedPlayer, err := playerRepo.GetByID(ctx, "test-game", "player1")
	require.NoError(t, err)
	require.Equal(t, selectedCards, updatedPlayer.Cards)

	// Verify: Credits deducted (2 cards × 3 MC = 6 MC)
	require.Equal(t, 44, updatedPlayer.Resources.Credits)

	// Verify: Phase advanced to Action (all players ready)
	updatedGame, err := gameRepo.GetByID(ctx, "test-game")
	require.NoError(t, err)
	require.Equal(t, game.GamePhaseAction, updatedGame.CurrentPhase)

	// Verify: ProductionPhase cleared (so frontend modal closes)
	finalPlayer, err := playerRepo.GetByID(ctx, "test-game", "player1")
	require.NoError(t, err)
	require.Nil(t, finalPlayer.ProductionPhase, "ProductionPhase should be cleared to close modal")
}

func TestConfirmProductionCardsAction_InsufficientCredits(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)

	sessionMgr := test.NewMockSessionManager()

	confirmAction := action.NewConfirmProductionCardsAction(gameRepo, playerRepo, nil, sessionMgr)

	// Setup game
	currentTurn := "player1"
	testGame := &game.Game{
		ID:           "test-game",
		Status:       game.GameStatusActive,
		CurrentPhase: game.GamePhaseProductionAndCardDraw,
		PlayerIDs:    []string{"player1"},
		Generation:   1,
		CurrentTurn:  &currentTurn,
		HostPlayerID: "player1",
	}
	err := gameRepo.Create(ctx, testGame)
	require.NoError(t, err)

	// Setup player with insufficient credits
	testPlayer := player.NewPlayer("TestPlayer")
	testPlayer.ID = "player1"
	testPlayer.Resources = types.Resources{
		Credits: 5, // Not enough for 3 cards (9 MC needed)
	}
	testPlayer.ProductionPhase = &types.ProductionPhase{
		AvailableCards:    []string{"card1", "card2", "card3", "card4"},
		SelectionComplete: false,
	}

	err = playerRepo.Create(ctx, "test-game", testPlayer)
	require.NoError(t, err)

	// Execute: Try to select 3 cards (requires 9 MC)
	selectedCards := []string{"card1", "card2", "card3"}
	err = confirmAction.Execute(ctx, "test-game", "player1", selectedCards)

	// Verify: Should fail with insufficient credits
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient credits")

	// Verify: Player state unchanged
	updatedPlayer, err := playerRepo.GetByID(ctx, "test-game", "player1")
	require.NoError(t, err)
	require.Empty(t, updatedPlayer.Cards, "No cards should be added on failure")
	require.Equal(t, 5, updatedPlayer.Resources.Credits, "Credits should not be deducted on failure")
	require.False(t, updatedPlayer.ProductionPhase.SelectionComplete, "Selection should not be marked complete on failure")
}

func TestConfirmProductionCardsAction_InvalidCardSelection(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)

	sessionMgr := test.NewMockSessionManager()

	confirmAction := action.NewConfirmProductionCardsAction(gameRepo, playerRepo, nil, sessionMgr)

	// Setup game
	currentTurn := "player1"
	testGame := &game.Game{
		ID:           "test-game",
		Status:       game.GameStatusActive,
		CurrentPhase: game.GamePhaseProductionAndCardDraw,
		PlayerIDs:    []string{"player1"},
		Generation:   1,
		CurrentTurn:  &currentTurn,
		HostPlayerID: "player1",
	}
	err := gameRepo.Create(ctx, testGame)
	require.NoError(t, err)

	// Setup player
	testPlayer := player.NewPlayer("TestPlayer")
	testPlayer.ID = "player1"
	testPlayer.Resources = types.Resources{
		Credits: 50,
	}
	testPlayer.ProductionPhase = &types.ProductionPhase{
		AvailableCards:    []string{"card1", "card2", "card3", "card4"},
		SelectionComplete: false,
	}

	err = playerRepo.Create(ctx, "test-game", testPlayer)
	require.NoError(t, err)

	// Execute: Try to select card not in available list
	selectedCards := []string{"card1", "invalid-card"}
	err = confirmAction.Execute(ctx, "test-game", "player1", selectedCards)

	// Verify: Should fail with invalid card error
	require.Error(t, err)
	require.Contains(t, err.Error(), "not available for selection")
}

func TestConfirmProductionCardsAction_AlreadyComplete(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)

	sessionMgr := test.NewMockSessionManager()

	confirmAction := action.NewConfirmProductionCardsAction(gameRepo, playerRepo, nil, sessionMgr)

	// Setup game
	currentTurn := "player1"
	testGame := &game.Game{
		ID:           "test-game",
		Status:       game.GameStatusActive,
		CurrentPhase: game.GamePhaseProductionAndCardDraw,
		PlayerIDs:    []string{"player1"},
		Generation:   1,
		CurrentTurn:  &currentTurn,
		HostPlayerID: "player1",
	}
	err := gameRepo.Create(ctx, testGame)
	require.NoError(t, err)

	// Setup player with already completed selection
	testPlayer := player.NewPlayer("TestPlayer")
	testPlayer.ID = "player1"
	testPlayer.Resources = types.Resources{
		Credits: 50,
	}
	testPlayer.ProductionPhase = &types.ProductionPhase{
		AvailableCards:    []string{"card1", "card2", "card3", "card4"},
		SelectionComplete: true, // Already complete
	}

	err = playerRepo.Create(ctx, "test-game", testPlayer)
	require.NoError(t, err)

	// Execute: Try to select cards again
	selectedCards := []string{"card1"}
	err = confirmAction.Execute(ctx, "test-game", "player1", selectedCards)

	// Verify: Should fail with already complete error
	require.Error(t, err)
	require.Contains(t, err.Error(), "already complete")
}

func TestConfirmProductionCardsAction_WrongPhase(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)

	sessionMgr := test.NewMockSessionManager()

	confirmAction := action.NewConfirmProductionCardsAction(gameRepo, playerRepo, nil, sessionMgr)

	// Setup game in wrong phase
	currentTurn := "player1"
	testGame := &game.Game{
		ID:           "test-game",
		Status:       game.GameStatusActive,
		CurrentPhase: game.GamePhaseAction, // Wrong phase
		PlayerIDs:    []string{"player1"},
		Generation:   1,
		CurrentTurn:  &currentTurn,
		HostPlayerID: "player1",
	}
	err := gameRepo.Create(ctx, testGame)
	require.NoError(t, err)

	// Setup player
	testPlayer := player.NewPlayer("TestPlayer")
	testPlayer.ID = "player1"
	testPlayer.Resources = types.Resources{
		Credits: 50,
	}

	err = playerRepo.Create(ctx, "test-game", testPlayer)
	require.NoError(t, err)

	// Execute: Try to confirm cards in wrong phase
	selectedCards := []string{"card1"}
	err = confirmAction.Execute(ctx, "test-game", "player1", selectedCards)

	// Verify: Should fail with phase error
	require.Error(t, err)
	require.Contains(t, err.Error(), "game not in production_and_card_draw phase")
}

func TestConfirmProductionCardsAction_MultiplePlayersSync(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)

	sessionMgr := test.NewMockSessionManager()

	confirmAction := action.NewConfirmProductionCardsAction(gameRepo, playerRepo, nil, sessionMgr)

	// Setup game with 2 players
	currentTurn := "player1"
	testGame := &game.Game{
		ID:           "test-game",
		Status:       game.GameStatusActive,
		CurrentPhase: game.GamePhaseProductionAndCardDraw,
		PlayerIDs:    []string{"player1", "player2"},
		Generation:   1,
		CurrentTurn:  &currentTurn,
		HostPlayerID: "player1",
	}
	err := gameRepo.Create(ctx, testGame)
	require.NoError(t, err)

	// Setup player 1
	player1 := player.NewPlayer("Player1")
	player1.ID = "player1"
	player1.Resources = types.Resources{Credits: 50}
	player1.ProductionPhase = &types.ProductionPhase{
		AvailableCards:    []string{"card1", "card2", "card3", "card4"},
		SelectionComplete: false,
	}
	err = playerRepo.Create(ctx, "test-game", player1)
	require.NoError(t, err)

	// Setup player 2
	player2 := player.NewPlayer("Player2")
	player2.ID = "player2"
	player2.Resources = types.Resources{Credits: 50}
	player2.ProductionPhase = &types.ProductionPhase{
		AvailableCards:    []string{"card5", "card6", "card7", "card8"},
		SelectionComplete: false,
	}
	err = playerRepo.Create(ctx, "test-game", player2)
	require.NoError(t, err)

	// Player 1 confirms
	err = confirmAction.Execute(ctx, "test-game", "player1", []string{"card1"})
	require.NoError(t, err)

	// Verify: Phase should NOT advance yet (player 2 not ready)
	g, err := gameRepo.GetByID(ctx, "test-game")
	require.NoError(t, err)
	require.Equal(t, game.GamePhaseProductionAndCardDraw, g.CurrentPhase, "Phase should not advance until all players ready")

	// Player 2 confirms
	err = confirmAction.Execute(ctx, "test-game", "player2", []string{"card5", "card6"})
	require.NoError(t, err)

	// Verify: NOW phase should advance (all players ready)
	g, err = gameRepo.GetByID(ctx, "test-game")
	require.NoError(t, err)
	require.Equal(t, game.GamePhaseAction, g.CurrentPhase, "Phase should advance when all players ready")

	// Verify: ProductionPhase cleared for both players (modal should close for all)
	p1Final, err := playerRepo.GetByID(ctx, "test-game", "player1")
	require.NoError(t, err)
	require.Nil(t, p1Final.ProductionPhase, "Player 1 ProductionPhase should be cleared")

	p2Final, err := playerRepo.GetByID(ctx, "test-game", "player2")
	require.NoError(t, err)
	require.Nil(t, p2Final.ProductionPhase, "Player 2 ProductionPhase should be cleared")
}

func TestConfirmProductionCardsAction_SelectZeroCards(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)

	sessionMgr := test.NewMockSessionManager()

	confirmAction := action.NewConfirmProductionCardsAction(gameRepo, playerRepo, nil, sessionMgr)

	// Setup game
	currentTurn := "player1"
	testGame := &game.Game{
		ID:           "test-game",
		Status:       game.GameStatusActive,
		CurrentPhase: game.GamePhaseProductionAndCardDraw,
		PlayerIDs:    []string{"player1"},
		Generation:   1,
		CurrentTurn:  &currentTurn,
		HostPlayerID: "player1",
	}
	err := gameRepo.Create(ctx, testGame)
	require.NoError(t, err)

	// Setup player
	testPlayer := player.NewPlayer("TestPlayer")
	testPlayer.ID = "player1"
	testPlayer.Resources = types.Resources{
		Credits: 50,
	}
	testPlayer.ProductionPhase = &types.ProductionPhase{
		AvailableCards:    []string{"card1", "card2", "card3", "card4"},
		SelectionComplete: false,
	}

	err = playerRepo.Create(ctx, "test-game", testPlayer)
	require.NoError(t, err)

	// Execute: Select zero cards
	err = confirmAction.Execute(ctx, "test-game", "player1", []string{})
	require.NoError(t, err)

	// Verify: No cards added, no cost
	updatedPlayer, err := playerRepo.GetByID(ctx, "test-game", "player1")
	require.NoError(t, err)
	require.Empty(t, updatedPlayer.Cards)
	require.Equal(t, 50, updatedPlayer.Resources.Credits, "Credits should not be deducted")
	require.Nil(t, updatedPlayer.ProductionPhase, "ProductionPhase should be cleared")
}
