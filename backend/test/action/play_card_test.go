package action_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
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
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-space-station")

	// Also add a space-tagged card to hand for modifier calculation
	player.Hand().AddCard("card-space-mirrors")

	// Verify initial state: no effects
	effectsBefore := player.Effects().List()
	testutil.AssertEqual(t, 0, len(effectsBefore), "Should have no effects initially")

	// Play Space Station
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-space-station", payment, nil)
	testutil.AssertNoError(t, err, "Failed to play Space Station")

	// Verify: effect should be registered
	effectsAfter := player.Effects().List()
	testutil.AssertEqual(t, 1, len(effectsAfter), "Should have 1 effect after playing Space Station")
	testutil.AssertEqual(t, "card-space-station", effectsAfter[0].CardID, "Effect should be from Space Station")
	testutil.AssertEqual(t, "Space Station", effectsAfter[0].CardName, "Effect card name should be Space Station")

	// Verify: discounts are calculated on-demand via RequirementModifierCalculator
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	spaceMirrorsCard, err := cardRegistry.GetByID("card-space-mirrors")
	testutil.AssertNoError(t, err, "Space Mirrors card should exist")

	discount := calculator.CalculateCardDiscounts(player, spaceMirrorsCard)
	testutil.AssertEqual(t, 2, discount, "Space Mirrors should have 2 credit discount from Space Station effect")
}

func TestPlayCardAction_ChoiceCardPlantProduction(t *testing.T) {
	// Setup: Create game with player who has Artificial Photosynthesis in hand
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

	// Give player enough credits and add Artificial Photosynthesis to hand
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-artificial-photosynthesis")

	// Verify initial production state
	productionBefore := player.Resources().Production()
	testutil.AssertEqual(t, 0, productionBefore.Plants, "Should have 0 plant production initially")
	testutil.AssertEqual(t, 0, productionBefore.Energy, "Should have 0 energy production initially")

	// Play Artificial Photosynthesis with choice index 0 (plant production +1)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	choiceIndex := 0
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-artificial-photosynthesis", payment, &choiceIndex)
	testutil.AssertNoError(t, err, "Failed to play Artificial Photosynthesis with choice 0")

	// Verify: plant production increased by 1, energy unchanged
	productionAfter := player.Resources().Production()
	testutil.AssertEqual(t, 1, productionAfter.Plants, "Should have 1 plant production after choice 0")
	testutil.AssertEqual(t, 0, productionAfter.Energy, "Should have 0 energy production after choice 0")

	// Verify card was played
	testutil.AssertFalse(t, player.Hand().HasCard("card-artificial-photosynthesis"), "Card should not be in hand")
}

func TestPlayCardAction_ChoiceCardEnergyProduction(t *testing.T) {
	// Setup: Create game with player who has Artificial Photosynthesis in hand
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

	// Give player enough credits and add Artificial Photosynthesis to hand
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-artificial-photosynthesis")

	// Verify initial production state
	productionBefore := player.Resources().Production()
	testutil.AssertEqual(t, 0, productionBefore.Plants, "Should have 0 plant production initially")
	testutil.AssertEqual(t, 0, productionBefore.Energy, "Should have 0 energy production initially")

	// Play Artificial Photosynthesis with choice index 1 (energy production +2)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	choiceIndex := 1
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-artificial-photosynthesis", payment, &choiceIndex)
	testutil.AssertNoError(t, err, "Failed to play Artificial Photosynthesis with choice 1")

	// Verify: energy production increased by 2, plants unchanged
	productionAfter := player.Resources().Production()
	testutil.AssertEqual(t, 0, productionAfter.Plants, "Should have 0 plant production after choice 1")
	testutil.AssertEqual(t, 2, productionAfter.Energy, "Should have 2 energy production after choice 1")

	// Verify card was played
	testutil.AssertFalse(t, player.Hand().HasCard("card-artificial-photosynthesis"), "Card should not be in hand")
}

func TestPlayCardAction_DiscountCalculatedOnDemand(t *testing.T) {
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

	// Give player credits and add Space Station to hand
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-space-station")

	// Play Space Station to register the discount effect
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-space-station", payment, nil)
	testutil.AssertNoError(t, err, "Failed to play Space Station")

	// Verify: effect registered
	effectsAfter := player.Effects().List()
	testutil.AssertEqual(t, 1, len(effectsAfter), "Should have 1 effect")

	// Verify: discounts are calculated on-demand for any space card
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)

	// Get Space Mirrors card (has space tag)
	spaceMirrorsCard, err := cardRegistry.GetByID("card-space-mirrors")
	testutil.AssertNoError(t, err, "Space Mirrors card should exist")

	// Discount should apply regardless of whether card is in hand
	discount := calculator.CalculateCardDiscounts(player, spaceMirrorsCard)
	testutil.AssertEqual(t, 2, discount, "Space Mirrors should have 2 credit discount from Space Station effect")

	// Non-space card should not get discount
	nonSpaceCard, err := cardRegistry.GetByID("card-arctic-algae")
	testutil.AssertNoError(t, err, "Arctic Algae card should exist")

	nonSpaceDiscount := calculator.CalculateCardDiscounts(player, nonSpaceCard)
	testutil.AssertEqual(t, 0, nonSpaceDiscount, "Non-space card should have no discount from Space Station")
}
