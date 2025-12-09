package action_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestStandardProjects_SelfContained_UpdatesOnResourceChange tests that standard projects update availability automatically
func TestStandardProjects_SelfContained_UpdatesOnResourceChange(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()

	ctx := context.Background()
	players := testGame.GetAllPlayers()
	player := players[0]

	// Set corporation
	player.SetCorporationID("corp-tharsis-republic")

	// Start the game (this registers standard projects)
	startAction := action.NewStartGameAction(repo, cardRegistry, logger)
	err := startAction.Execute(ctx, testGame.ID(), testGame.HostPlayerID())
	testutil.AssertNoError(t, err, "Failed to start game")

	// Get updated game state
	testGame, _ = repo.Get(ctx, testGame.ID())
	players = testGame.GetAllPlayers()
	player = players[0]

	// Give player insufficient credits for Power Plant (needs 11)
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: 5,
	})

	// Get standard project playability
	standardProjects := player.StandardProjects().GetAllAvailability()

	// Find Power Plant project
	var powerPlantAvailable bool
	for _, proj := range standardProjects {
		if proj.ID == "power-plant" {
			powerPlantAvailable = proj.IsAvailable
			break
		}
	}

	// Power Plant should NOT be available (has 5, needs 11)
	testutil.AssertFalse(t, powerPlantAvailable, "Power Plant should not be available with only 5 credits")

	// Add more credits
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: 10, // Now has 15 total
	})

	// Get updated playability (should auto-update via event subscription)
	standardProjects = player.StandardProjects().GetAllAvailability()

	// Find Power Plant project again
	powerPlantAvailable = false
	for _, proj := range standardProjects {
		if proj.ID == "power-plant" {
			powerPlantAvailable = proj.IsAvailable
			break
		}
	}

	// Power Plant should NOW be available (has 15, needs 11)
	testutil.AssertTrue(t, powerPlantAvailable, "Power Plant should be available with 15 credits")
}

// TestStandardProjects_SelfContained_UpdatesOnGlobalParams tests that standard projects update when global parameters change
func TestStandardProjects_SelfContained_UpdatesOnGlobalParams(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()

	ctx := context.Background()
	players := testGame.GetAllPlayers()
	player := players[0]

	// Set corporation
	player.SetCorporationID("corp-tharsis-republic")

	// Start the game
	startAction := action.NewStartGameAction(repo, cardRegistry, logger)
	err := startAction.Execute(ctx, testGame.ID(), testGame.HostPlayerID())
	testutil.AssertNoError(t, err, "Failed to start game")

	// Get updated game state
	testGame, _ = repo.Get(ctx, testGame.ID())
	players = testGame.GetAllPlayers()
	player = players[0]

	// Give player enough credits for Asteroid (needs 14)
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: 20,
	})

	// Initially, Asteroid should be available (temperature not maxed)
	standardProjects := player.StandardProjects().GetAllAvailability()
	var asteroidAvailable bool
	for _, proj := range standardProjects {
		if proj.ID == "asteroid" {
			asteroidAvailable = proj.IsAvailable
			break
		}
	}

	testutil.AssertTrue(t, asteroidAvailable, "Asteroid should be available initially")

	// Max out temperature
	for i := 0; i < 19; i++ { // Temperature goes from -30 to +8 in steps of 2
		testGame.GlobalParameters().IncreaseTemperature(ctx, 1)
	}

	// Get updated playability (should auto-update via event subscription)
	standardProjects = player.StandardProjects().GetAllAvailability()

	asteroidAvailable = true // Reset to check
	for _, proj := range standardProjects {
		if proj.ID == "asteroid" {
			asteroidAvailable = proj.IsAvailable
			break
		}
	}

	// Asteroid should NO LONGER be available (temperature maxed)
	testutil.AssertFalse(t, asteroidAvailable, "Asteroid should not be available when temperature is maxed")
}

// TestStandardProjects_SelfContained_MultipleProjectsIndependent tests that different standard projects maintain independent availability
func TestStandardProjects_SelfContained_MultipleProjectsIndependent(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()

	ctx := context.Background()
	players := testGame.GetAllPlayers()
	player := players[0]

	// Set corporation
	player.SetCorporationID("corp-tharsis-republic")

	// Start the game
	startAction := action.NewStartGameAction(repo, cardRegistry, logger)
	err := startAction.Execute(ctx, testGame.ID(), testGame.HostPlayerID())
	testutil.AssertNoError(t, err, "Failed to start game")

	// Get updated game state
	testGame, _ = repo.Get(ctx, testGame.ID())
	players = testGame.GetAllPlayers()
	player = players[0]

	// Give player 15 credits (enough for Power Plant [11] but not City [25])
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: 15,
	})

	// Get standard project playability
	standardProjects := player.StandardProjects().GetAllAvailability()

	var powerPlantAvailable, cityAvailable bool
	for _, proj := range standardProjects {
		if proj.ID == "power-plant" {
			powerPlantAvailable = proj.IsAvailable
		}
		if proj.ID == "city" {
			cityAvailable = proj.IsAvailable
		}
	}

	// Power Plant should be available (has 15, needs 11)
	testutil.AssertTrue(t, powerPlantAvailable, "Power Plant should be available with 15 credits")

	// City should NOT be available (has 15, needs 25)
	testutil.AssertFalse(t, cityAvailable, "City should not be available with only 15 credits")

	// Add more credits
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: 15, // Now has 30 total
	})

	// Get updated playability
	standardProjects = player.StandardProjects().GetAllAvailability()

	powerPlantAvailable = false
	cityAvailable = false
	for _, proj := range standardProjects {
		if proj.ID == "power-plant" {
			powerPlantAvailable = proj.IsAvailable
		}
		if proj.ID == "city" {
			cityAvailable = proj.IsAvailable
		}
	}

	// Both should now be available
	testutil.AssertTrue(t, powerPlantAvailable, "Power Plant should still be available")
	testutil.AssertTrue(t, cityAvailable, "City should now be available with 30 credits")
}
