package service

import (
	"context"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInventrix_ForcedFirstActionFlag tests that selecting Inventrix
// sets the forced first action flag (draw 3 cards on first turn, not immediately)
func TestInventrix_ForcedFirstActionFlag(t *testing.T) {
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
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber, forcedActionManager)

	// Load cards
	require.NoError(t, cardRepo.LoadCards(ctx))

	// Create test game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Initialize board
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

	// Get cards and corporations
	startingCards, _ := cardRepo.GetStartingCardPool(ctx)
	corporations, _ := cardRepo.GetCorporations(ctx)
	require.GreaterOrEqual(t, len(startingCards), 2)
	require.GreaterOrEqual(t, len(corporations), 1)

	// Find Inventrix
	var inventrixID string
	for _, corp := range corporations {
		if corp.ID == "B05" {
			inventrixID = corp.ID
			break
		}
	}
	require.NotEmpty(t, inventrixID, "Inventrix (B05) should be available")

	// Set up starting card selection
	availableCardIDs := []string{startingCards[0].ID, startingCards[1].ID}
	corporationIDs := []string{inventrixID, corporations[0].ID}

	playerRepo.UpdateSelectStartingCardsPhase(ctx, game.ID, player.ID, &model.SelectStartingCardsPhase{
		AvailableCards:        availableCardIDs,
		AvailableCorporations: corporationIDs,
	})

	// Select Inventrix
	selectedCardIDs := []string{startingCards[0].ID}
	err = cardService.OnSelectStartingCards(ctx, game.ID, player.ID, selectedCardIDs, inventrixID)
	require.NoError(t, err, "Should successfully select Inventrix")

	t.Log("✅ Inventrix selected")

	// Verify player has forced first action flag set
	playerAfterSelection, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	require.NoError(t, err)
	require.NotNil(t, playerAfterSelection.ForcedFirstAction, "Player should have forced first action")
	assert.Equal(t, "card_draw", playerAfterSelection.ForcedFirstAction.ActionType, "Action type should be card_draw")
	assert.Equal(t, "B05", playerAfterSelection.ForcedFirstAction.CorporationID, "Should be from Inventrix")
	assert.False(t, playerAfterSelection.ForcedFirstAction.Completed, "Action should not be completed yet")
	assert.Equal(t, "Draw 3 cards", playerAfterSelection.ForcedFirstAction.Description, "Should have correct description")

	t.Logf("🎯 Forced first action set: %s", playerAfterSelection.ForcedFirstAction.Description)

	// Verify the draw-3-cards action was NOT registered as a repeatable manual action
	assert.Empty(t, playerAfterSelection.Actions, "Inventrix should NOT have draw-3-cards in player.Actions (it's a forced first action)")

	t.Log("✅ Draw-3-cards is flagged as forced first action, NOT a repeatable action")

	// Verify game advances to action phase normally
	game, err = gameRepo.GetByID(ctx, game.ID)
	require.NoError(t, err)
	assert.Equal(t, model.GamePhaseAction, game.CurrentPhase, "Game should advance to action phase")

	t.Log("🎉 Inventrix forced first action test passed!")
	t.Log("📝 Note: Card draw execution during first turn is not yet implemented")
}
