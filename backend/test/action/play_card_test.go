package action_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestPlayCardAction_DiscountEffectRegistered(t *testing.T) {
	// Setup: Create game with player who has Space Station in hand
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	// Set game to active status and action phase for playing cards
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	// Give player enough credits and add Space Station to hand
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: 100,
	})
	player.Hand().AddCard("card-space-station")

	// Also add a space-tagged card to hand for modifier calculation
	player.Hand().AddCard("card-space-mirrors")

	// Verify initial state: no effects, no modifiers
	effectsBefore := player.Effects().List()
	modifiersBefore := player.Effects().RequirementModifiers()
	testutil.AssertEqual(t, 0, len(effectsBefore), "Should have no effects initially")
	testutil.AssertEqual(t, 0, len(modifiersBefore), "Should have no modifiers initially")

	// Play Space Station
	playCardAction := action.NewPlayCardAction(repo, cardRegistry, logger)
	payment := action.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-space-station", payment)
	testutil.AssertNoError(t, err, "Failed to play Space Station")

	// Verify: effect should be registered
	effectsAfter := player.Effects().List()
	testutil.AssertEqual(t, 1, len(effectsAfter), "Should have 1 effect after playing Space Station")
	testutil.AssertEqual(t, "card-space-station", effectsAfter[0].CardID, "Effect should be from Space Station")
	testutil.AssertEqual(t, "Space Station", effectsAfter[0].CardName, "Effect card name should be Space Station")

	// Verify: requirement modifiers should be calculated for space-mirrors in hand
	modifiersAfter := player.Effects().RequirementModifiers()
	testutil.AssertEqual(t, 1, len(modifiersAfter), "Should have 1 modifier for card-space-mirrors")

	// Verify the modifier details
	modifier := modifiersAfter[0]
	testutil.AssertEqual(t, 2, modifier.Amount, "Discount should be 2")
	testutil.AssertTrue(t, modifier.CardTarget != nil, "Modifier should target a card")
	testutil.AssertEqual(t, "card-space-mirrors", *modifier.CardTarget, "Modifier should target card-space-mirrors")
}

func TestPlayCardAction_DiscountModifierRecalculatedOnHandChange(t *testing.T) {
	// Setup: Create game with Space Station already played (effect registered)
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	// Set game to active status and action phase
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	// Setup global subscribers (normally done in StartGameAction)
	globalSubscriber := action.NewGlobalSubscriber(cardRegistry, logger)
	globalSubscriber.SetupGlobalSubscribers(testGame)

	// Give player credits and add Space Station to hand
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: 100,
	})
	player.Hand().AddCard("card-space-station")

	// Play Space Station to register the discount effect
	playCardAction := action.NewPlayCardAction(repo, cardRegistry, logger)
	payment := action.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-space-station", payment)
	testutil.AssertNoError(t, err, "Failed to play Space Station")

	// Verify: effect registered, but no modifiers yet (no space cards in hand)
	effectsAfter := player.Effects().List()
	testutil.AssertEqual(t, 1, len(effectsAfter), "Should have 1 effect")

	modifiersAfter := player.Effects().RequirementModifiers()
	testutil.AssertEqual(t, 0, len(modifiersAfter), "Should have no modifiers (no space cards in hand)")

	// Now add a space card to hand - this triggers CardHandUpdatedEvent
	// which should recalculate modifiers via GlobalSubscriber
	player.Hand().AddCard("card-space-mirrors")

	// Give the event handler time to process (it's synchronous in tests)
	modifiersAfterAdd := player.Effects().RequirementModifiers()
	testutil.AssertEqual(t, 1, len(modifiersAfterAdd), "Should have 1 modifier after adding space card to hand")

	// Verify the modifier details
	modifier := modifiersAfterAdd[0]
	testutil.AssertEqual(t, 2, modifier.Amount, "Discount should be 2")
	testutil.AssertTrue(t, modifier.CardTarget != nil, "Modifier should target a card")
	testutil.AssertEqual(t, "card-space-mirrors", *modifier.CardTarget, "Modifier should target card-space-mirrors")
}
