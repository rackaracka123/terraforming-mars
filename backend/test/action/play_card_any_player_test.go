package action_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestPlayCardAction_AsteroidRemovesPlantsFromTargetPlayer(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard("card-asteroid")

	// Give target 5 plants
	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlant: 5,
	})

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 14}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-asteroid", payment, nil, nil, &targetID)
	testutil.AssertNoError(t, err, "Failed to play Asteroid")

	// Target should have 5 - 3 = 2 plants
	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 2, targetResources.Plants, "Target should have 2 plants after Asteroid removes 3")

	// Attacker should have gained 2 titanium from the second behavior
	attackerResources := attacker.Resources().Get()
	testutil.AssertEqual(t, 2, attackerResources.Titanium, "Attacker should have gained 2 titanium")
}

func TestPlayCardAction_AsteroidSoloMode_SkipsTargetPlayer(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
		shared.ResourcePlant:  10,
	})
	player.Hand().AddCard("card-asteroid")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 14}

	// No targetPlayerID = solo mode, skip the any-player output
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-asteroid", payment, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Asteroid in solo mode")

	// Player's plants should be unchanged (the any-player effect does nothing in solo)
	resources := player.Resources().Get()
	testutil.AssertEqual(t, 10, resources.Plants, "Player plants should be unchanged in solo mode")

	// Player should still get titanium from the second behavior
	testutil.AssertEqual(t, 2, resources.Titanium, "Player should have gained 2 titanium")
}

func TestPlayCardAction_AsteroidPartialRemoval(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard("card-asteroid")

	// Give target only 1 plant (less than the 3 Asteroid tries to remove)
	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlant: 1,
	})

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 14}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-asteroid", payment, nil, nil, &targetID)
	testutil.AssertNoError(t, err, "Failed to play Asteroid with partial removal")

	// Target should have 0 plants (had 1, Asteroid removes up to 3)
	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 0, targetResources.Plants, "Target should have 0 plants after partial removal")
}

func TestPlayCardAction_InvalidTargetPlayerID(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	attacker := players[0]
	attacker.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard("card-asteroid")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 14}
	invalidID := "non-existent-player"
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-asteroid", payment, nil, nil, &invalidID)
	testutil.AssertError(t, err, "Should fail with invalid target player ID")
}

func TestPlayCardAction_AsteroidMiningConsortiumDecreasesTargetProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	// Attacker needs titanium production for the requirement
	attacker.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceTitaniumProduction: 1,
	})
	attacker.Hand().AddCard("card-asteroid-mining-consortium")

	// Give target 2 titanium production
	target.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceTitaniumProduction: 2,
	})

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-asteroid-mining-consortium", payment, nil, nil, &targetID)
	testutil.AssertNoError(t, err, "Failed to play Asteroid Mining Consortium")

	// Target's titanium production should decrease by 1 (from 2 to 1)
	targetProduction := target.Resources().Production()
	testutil.AssertEqual(t, 1, targetProduction.Titanium, "Target should have 1 titanium production after decrease")

	// Attacker's titanium production should increase by 1 (from 1 to 2)
	attackerProduction := attacker.Resources().Production()
	testutil.AssertEqual(t, 2, attackerProduction.Titanium, "Attacker should have 2 titanium production after increase")
}

func TestPlayCardAction_HiredRaidersStealsSteelFromTarget(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard("card-hired-raiders")

	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceSteel: 5,
	})

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetID := target.ID()
	choiceIndex := 0
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-hired-raiders", payment, &choiceIndex, nil, &targetID)
	testutil.AssertNoError(t, err, "Failed to play Hired Raiders")

	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 3, targetResources.Steel, "Target should have 3 steel after 2 stolen")

	attackerResources := attacker.Resources().Get()
	testutil.AssertEqual(t, 2, attackerResources.Steel, "Attacker should have gained 2 steel")
}

func TestPlayCardAction_HiredRaidersSoloMode(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
		shared.ResourceSteel:  3,
	})
	player.Hand().AddCard("card-hired-raiders")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	choiceIndex := 0
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-hired-raiders", payment, &choiceIndex, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Hired Raiders in solo mode")

	resources := player.Resources().Get()
	testutil.AssertEqual(t, 3, resources.Steel, "Player steel should be unchanged in solo mode")
}

func TestPlayCardAction_StealPartialAmount(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	target.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, attacker.ID(), 2)

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard("card-hired-raiders")

	// Target has only 1 steel, less than the 2 Hired Raiders tries to steal
	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceSteel: 1,
	})

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetID := target.ID()
	choiceIndex := 0
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-hired-raiders", payment, &choiceIndex, nil, &targetID)
	testutil.AssertNoError(t, err, "Failed to play Hired Raiders with partial steal")

	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 0, targetResources.Steel, "Target should have 0 steel after partial steal")

	attackerResources := attacker.Resources().Get()
	testutil.AssertEqual(t, 1, attackerResources.Steel, "Attacker should have gained only 1 steel")
}
