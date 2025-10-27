package service

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makePayment creates a credits-only payment for the given card cost
func makePayment(ctx context.Context, cardRepo repository.CardRepository, cardID string) *model.CardPayment {
	card, err := cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		// If card not found, return zero payment (test will fail with proper error)
		return &model.CardPayment{Credits: 0, Steel: 0, Titanium: 0}
	}
	return &model.CardPayment{
		Credits:  card.Cost,
		Steel:    0,
		Titanium: 0,
	}
}

// TestDeepWellHeating_MaxTemperature tests that Deep Well Heating doesn't exceed max temperature
func TestDeepWellHeating_MaxTemperature(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, gameRepo, cardRepo := setupCardTest(t)

	// Set temperature to max (8°C)
	gameRepo.UpdateGlobalParameters(ctx, gameID, model.GlobalParameters{
		Temperature: 8, Oxygen: 0, Oceans: 0,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 15})
	cardID := "003"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	gameBefore, _ := gameRepo.GetByID(ctx, gameID)
	assert.Equal(t, 8, gameBefore.GlobalParameters.Temperature, "Temperature should be at max")

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Deep Well Heating")

	gameAfter, _ := gameRepo.GetByID(ctx, gameID)
	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)

	// Temperature should be clamped at max
	assert.Equal(t, 8, gameAfter.GlobalParameters.Temperature, "Temperature should stay at max (8°C)")
	// Energy production should still increase
	assert.Equal(t, 1, playerAfter.Production.Energy, "Energy production should still increase")

	t.Log("✅ Deep Well Heating max temperature edge case passed")
}

// TestNuclearPower_NegativeCreditsProduction tests Nuclear Power with low credits production (allows negative down to -5)
func TestNuclearPower_NegativeCreditsProduction(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	// Set credits production to 1 (effect is -2, result is -1 which is allowed)
	playerRepo.UpdateProduction(ctx, gameID, playerID, model.Production{
		Credits: 1,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 12})
	cardID := "045"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Nuclear Power with low credits production")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, -1, playerAfter.Production.Credits, "Credits production can go negative (1-2 = -1)")
	assert.Equal(t, 3, playerAfter.Production.Energy, "Energy production should be 3")

	t.Log("✅ Nuclear Power negative credits production edge case passed")
}

// TestNuclearPower_MinimumCreditsProduction tests Nuclear Power with exactly 2 credits production
func TestNuclearPower_MinimumCreditsProduction(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	// Set credits production to exactly 2 (minimum to play)
	playerRepo.UpdateProduction(ctx, gameID, playerID, model.Production{
		Credits: 2,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 12})
	cardID := "045"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Nuclear Power with 2 credits production")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 0, playerAfter.Production.Credits, "Credits production should be 0 (2-2)")
	assert.Equal(t, 3, playerAfter.Production.Energy, "Energy production should be 3")

	t.Log("✅ Nuclear Power minimum credits production edge case passed")
}

// TestCarbonateProcessing_InsufficientEnergyProduction tests that card is blocked with 0 energy production
func TestCarbonateProcessing_InsufficientEnergyProduction(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	// Start with 0 energy production (effect is -1, would go negative which is not allowed)
	playerRepo.UpdateProduction(ctx, gameID, playerID, model.Production{
		Energy: 0,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 8})
	cardID := "043"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.Error(t, err, "Should fail to play Carbonate Processing with 0 energy production")
	assert.Contains(t, err.Error(), "insufficient energy production", "Error should mention insufficient energy production")

	t.Log("✅ Carbonate Processing insufficient energy production validation passed")
}

// TestCarbonateProcessing_MinimumEnergyProduction tests card with exactly 1 energy production
func TestCarbonateProcessing_MinimumEnergyProduction(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	// Start with 1 energy production (minimum to play the card)
	playerRepo.UpdateProduction(ctx, gameID, playerID, model.Production{
		Energy: 1,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 8})
	cardID := "043"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Carbonate Processing with 1 energy production")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 0, playerAfter.Production.Energy, "Energy production should be 0 (1-1)")
	assert.Equal(t, 3, playerAfter.Production.Heat, "Heat production should be 3")

	t.Log("✅ Carbonate Processing minimum energy production edge case passed")
}

// TestFoodFactory_InsufficientPlantsProduction tests Food Factory validation with no plants production
func TestFoodFactory_InsufficientPlantsProduction(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	// Start with 0 plants production (effect is -1, would go negative which is not allowed)
	playerRepo.UpdateProduction(ctx, gameID, playerID, model.Production{
		Plants: 0, Credits: 1,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 15})
	cardID := "041"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.Error(t, err, "Should fail with insufficient plants production")
	assert.Contains(t, err.Error(), "insufficient", "Error should mention insufficient production")

	t.Log("✅ Food Factory insufficient plants production validation passed")
}

// TestFoodFactory_MinimumPlantsProduction tests Food Factory with exactly 1 plants production
func TestFoodFactory_MinimumPlantsProduction(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	// Start with exactly 1 plants production (minimum to play)
	playerRepo.UpdateProduction(ctx, gameID, playerID, model.Production{
		Plants: 1, Credits: 1,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 15})
	cardID := "041"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Food Factory with 1 plants production")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 0, playerAfter.Production.Plants, "Plants production should be 0 (1-1)")
	assert.Equal(t, 5, playerAfter.Production.Credits, "Credits production should be 5 (1+4)")

	t.Log("✅ Food Factory minimum plants production edge case passed")
}

// TestBigAsteroid_MaxTemperature tests Big Asteroid at max temperature
func TestBigAsteroid_MaxTemperature(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, gameRepo, cardRepo := setupCardTest(t)

	// Set temperature near max
	gameRepo.UpdateGlobalParameters(ctx, gameID, model.GlobalParameters{
		Temperature: 7, Oxygen: 0, Oceans: 0,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 30})
	cardID := "011"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Big Asteroid")

	gameAfter, _ := gameRepo.GetByID(ctx, gameID)
	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)

	// Should increase by 1 only (clamped at 8)
	assert.Equal(t, 8, gameAfter.GlobalParameters.Temperature, "Temperature should be clamped at max (8°C)")
	assert.Equal(t, 4, playerAfter.Resources.Titanium, "Should still gain 4 titanium")

	t.Log("✅ Big Asteroid max temperature edge case passed")
}

// TestArchaebacteria_ExactTemperatureRequirement tests Archaebacteria at exactly -18°C
func TestArchaebacteria_ExactTemperatureRequirement(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, gameRepo, cardRepo := setupCardTest(t)

	// Set temperature to exactly -18°C (the requirement boundary)
	gameRepo.UpdateGlobalParameters(ctx, gameID, model.GlobalParameters{
		Temperature: -18, Oxygen: 0, Oceans: 0,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 8})
	cardID := "042"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Archaebacteria at exactly -18°C")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 1, playerAfter.Production.Plants, "Plants production should increase by 1")

	t.Log("✅ Archaebacteria exact temperature requirement edge case passed")
}

// TestMethaneFromTitan_ExactOxygenRequirement tests Methane From Titan at exactly 2% oxygen
func TestMethaneFromTitan_ExactOxygenRequirement(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, gameRepo, cardRepo := setupCardTest(t)

	// Set oxygen to exactly 2% (the requirement boundary)
	gameRepo.UpdateGlobalParameters(ctx, gameID, model.GlobalParameters{
		Temperature: -30, Oxygen: 2, Oceans: 0,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 30})
	cardID := "018"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Methane From Titan at exactly 2% oxygen")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 2, playerAfter.Production.Heat, "Heat production should be 2")
	assert.Equal(t, 2, playerAfter.Production.Plants, "Plants production should be 2")

	t.Log("✅ Methane From Titan exact oxygen requirement edge case passed")
}

// TestLunarBeam_MinimumCreditsProduction tests Lunar Beam with exactly 2 credits production
func TestLunarBeam_MinimumCreditsProduction(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	// Start with exactly 2 credits production (minimum to play)
	playerRepo.UpdateProduction(ctx, gameID, playerID, model.Production{
		Credits: 2,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 15})
	cardID := "030"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Lunar Beam with 2 credits production")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 0, playerAfter.Production.Credits, "Credits production should be 0 (2-2)")
	assert.Equal(t, 2, playerAfter.Production.Heat, "Heat production should be 2")
	assert.Equal(t, 2, playerAfter.Production.Energy, "Energy production should be 2")

	t.Log("✅ Lunar Beam minimum credits production edge case passed")
}

// TestUndergroundCity_MinimumEnergyProduction tests Underground City with exactly 2 energy production
func TestUndergroundCity_MinimumEnergyProduction(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	// Start with exactly 2 energy production (minimum to play)
	playerRepo.UpdateProduction(ctx, gameID, playerID, model.Production{
		Energy: 2,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 20})
	cardID := "032"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Underground City with 2 energy production")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 0, playerAfter.Production.Energy, "Energy production should be 0 (2-2)")
	assert.Equal(t, 2, playerAfter.Production.Steel, "Steel production should be 2")
	assert.NotNil(t, playerAfter.PendingTileSelection, "Should have pending city tile")

	t.Log("✅ Underground City minimum energy production edge case passed")
}

// TestBlackPolarDust_MinimumCreditsProduction tests Black Polar Dust with exactly 2 credits production
func TestBlackPolarDust_MinimumCreditsProduction(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	// Start with exactly 2 credits production (minimum to play)
	playerRepo.UpdateProduction(ctx, gameID, playerID, model.Production{
		Credits: 2,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 20})
	cardID := "022"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Black Polar Dust with 2 credits production")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 0, playerAfter.Production.Credits, "Credits production should be 0 (2-2)")
	assert.Equal(t, 3, playerAfter.Production.Heat, "Heat production should be 3")
	assert.NotNil(t, playerAfter.PendingTileSelection, "Should have pending ocean tile")

	t.Log("✅ Black Polar Dust minimum credits production edge case passed")
}

// TestComet_MinTemperature tests Comet at minimum temperature
func TestComet_MinTemperature(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, gameRepo, cardRepo := setupCardTest(t)

	// Set temperature to minimum
	gameRepo.UpdateGlobalParameters(ctx, gameID, model.GlobalParameters{
		Temperature: -30, Oxygen: 0, Oceans: 0,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 25})
	cardID := "010"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Comet")

	gameAfter, _ := gameRepo.GetByID(ctx, gameID)
	assert.Equal(t, -29, gameAfter.GlobalParameters.Temperature, "Temperature should increase from -30 to -29")

	t.Log("✅ Comet min temperature edge case passed")
}

// TestAsteroid_MinTemperature tests Asteroid at minimum temperature
func TestAsteroid_MinTemperature(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, gameRepo, cardRepo := setupCardTest(t)

	// Set temperature to minimum
	gameRepo.UpdateGlobalParameters(ctx, gameID, model.GlobalParameters{
		Temperature: -30, Oxygen: 0, Oceans: 0,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 16})
	cardID := "009"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Asteroid")

	gameAfter, _ := gameRepo.GetByID(ctx, gameID)
	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)

	assert.Equal(t, -29, gameAfter.GlobalParameters.Temperature, "Temperature should increase from -30 to -29")
	assert.Equal(t, 2, playerAfter.Resources.Titanium, "Should gain 2 titanium")

	t.Log("✅ Asteroid min temperature edge case passed")
}

// TestNitrogenRichAsteroid_ChoiceIndex1 tests Nitrogen-Rich Asteroid with choice 1 (4 plant production)
func TestNitrogenRichAsteroid_ChoiceIndex1(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, gameRepo, cardRepo := setupCardTest(t)

	gameRepo.UpdateGlobalParameters(ctx, gameID, model.GlobalParameters{
		Temperature: -16, Oxygen: 0, Oceans: 0,
	})
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 35})
	cardID := "037"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	// Choose option 1: +4 plant production (requires 3 plant tags, but we test the mechanism)
	choiceIndex := 1
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), &choiceIndex, nil)
	require.NoError(t, err, "Should successfully play Nitrogen-Rich Asteroid with choice 1")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	gameAfter, _ := gameRepo.GetByID(ctx, gameID)

	assert.Equal(t, 4, playerAfter.Production.Plants, "Plant production should increase by 4 with choice 1")
	assert.Equal(t, -15, gameAfter.GlobalParameters.Temperature, "Temperature should increase by 1")

	t.Log("✅ Nitrogen-Rich Asteroid choice 1 edge case passed")
}

// TestImportedHydrogen_Choice0 tests Imported Hydrogen gaining plants
func TestImportedHydrogen_Choice0_GainPlants(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 20, Plants: 5})
	cardID := "019"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	playerBefore, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 5, playerBefore.Resources.Plants, "Initial plants should be 5")

	// Choose option 0: Gain 3 plants
	choiceIndex := 0
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), &choiceIndex, nil)
	require.NoError(t, err, "Should successfully play Imported Hydrogen")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 8, playerAfter.Resources.Plants, "Should gain 3 plants (5+3=8)")
	assert.NotNil(t, playerAfter.PendingTileSelection, "Should have pending ocean tile")

	t.Log("✅ Imported Hydrogen choice 0 edge case passed")
}

// TestReleaseOfInertGases_MaxTR tests Release Of Inert Gases with very high TR
func TestReleaseOfInertGases_HighTR(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	// Set TR to a high value
	playerRepo.UpdateTerraformRating(ctx, gameID, playerID, 50)
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 20})
	cardID := "036"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	playerBefore, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 50, playerBefore.TerraformRating, "Initial TR should be 50")

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Release Of Inert Gases")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 52, playerAfter.TerraformRating, "TR should increase by 2 (50+2=52)")

	t.Log("✅ Release Of Inert Gases high TR edge case passed")
}

// TestAsteroidMining_VictoryPoints tests that victory points are awarded correctly
func TestAsteroidMining_VictoryPoints(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _, cardRepo := setupCardTest(t)

	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 35})
	cardID := "040"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	playerBefore, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 0, playerBefore.VictoryPoints, "Initial VP should be 0")

	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, makePayment(ctx, cardRepo, cardID), nil, nil)
	require.NoError(t, err, "Should successfully play Asteroid Mining")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 2, playerAfter.VictoryPoints, "Should gain 2 VP")
	assert.Equal(t, 2, playerAfter.Production.Titanium, "Titanium production should be 2")

	t.Log("✅ Asteroid Mining victory points edge case passed")
}

// ============================================================================
// CORPORATION TESTS - Testing all 12 base game corporations
// ============================================================================

// TestCorporation_CrediCor tests CrediCor corporation selection
func TestCorporation_CrediCor(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _ := setupCorporationTest(t)

	// Select CrediCor corporation
	err := cardService.OnSelectStartingCards(ctx, gameID, playerID, []string{}, "B01")
	require.NoError(t, err, "Should select CrediCor corporation")

	player, _ := playerRepo.GetByID(ctx, gameID, playerID)

	// Verify corporation is set
	assert.NotNil(t, player.Corporation, "Corporation should be set")
	assert.Equal(t, "B01", player.Corporation.ID, "Corporation should be CrediCor")
	assert.Equal(t, "CrediCor", player.Corporation.Name, "Corporation name should be CrediCor")

	// TODO: Verify starting resources (57 MC) - currently not working correctly in card data
	// TODO: Test passive effect (gain 4 MC when playing card costing 20+ MC) - not implemented

	t.Log("✅ CrediCor corporation selection test passed")
}

// TestCorporation_Ecoline tests Ecoline corporation selection
func TestCorporation_Ecoline(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _ := setupCorporationTest(t)

	// Select Ecoline corporation
	err := cardService.OnSelectStartingCards(ctx, gameID, playerID, []string{}, "B02")
	require.NoError(t, err, "Should select Ecoline corporation")

	player, _ := playerRepo.GetByID(ctx, gameID, playerID)

	// Verify corporation is set
	assert.NotNil(t, player.Corporation, "Corporation should be set")
	assert.Equal(t, "B02", player.Corporation.ID, "Corporation should be Ecoline")
	assert.Equal(t, "Ecoline", player.Corporation.Name, "Corporation name should be Ecoline")

	// TODO: Verify starting resources (36 MC, 3 plants, 2 plant production)
	// TODO: Test reduced greenery cost (7 plants instead of 8)

	t.Log("✅ Ecoline corporation selection test passed")
}

// TestCorporation_AllBaseGame tests that all 12 base game corporations can be selected
func TestCorporation_AllBaseGame(t *testing.T) {
	// Test data for all 12 base game corporations
	corporations := []struct {
		id   string
		name string
	}{
		{"B01", "CrediCor"},
		{"B02", "Ecoline"},
		{"B03", "Helion"},
		{"B04", "Interplanetary Cinematics"},
		{"B05", "Inventrix"},
		{"B06", "Mining Guild"},
		{"B07", "PhoboLog"},
		{"B08", "Tharsis Republic"},
		{"B09", "ThorGate"},
		{"B10", "United Nations Mars Initiative"},
		{"B11", "Saturn Systems"},
		{"B12", "Teractor"},
	}

	for _, corp := range corporations {
		t.Run(corp.name, func(t *testing.T) {
			ctx, gameID, playerID, cardService, playerRepo, _ := setupCorporationTest(t)

			// Select corporation
			err := cardService.OnSelectStartingCards(ctx, gameID, playerID, []string{}, corp.id)
			require.NoError(t, err, "Should select "+corp.name+" corporation")

			player, _ := playerRepo.GetByID(ctx, gameID, playerID)

			// Verify corporation is set correctly
			assert.NotNil(t, player.Corporation, "Corporation should be set")
			assert.Equal(t, corp.id, player.Corporation.ID, "Corporation ID should match")
			assert.Equal(t, corp.name, player.Corporation.Name, "Corporation name should match")

			t.Logf("✅ %s corporation selection test passed", corp.name)
		})
	}

	// TODO: Add tests for starting resources and production for each corporation
	// TODO: Add tests for passive/active corporation abilities
	// TODO: Add tests for corporation-specific game mechanics (e.g., Ecoline greenery discount)
}

// setupCorporationTest sets up a test environment for corporation testing
// Note: This function provides ALL base game corporations (B01-B12) as available for testing purposes
func setupCorporationTest(t *testing.T) (context.Context, string, string, service.CardService, repository.PlayerRepository, repository.CardRepository) {
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

	// Create test game in starting card selection phase
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)
	gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
	gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseStartingCardSelection)

	// Create test player
	player := model.Player{
		ID: "player1", Name: "Test Player",
		Resources:       model.Resources{Credits: 50},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20, IsConnected: true,
	}
	playerRepo.Create(ctx, game.ID, player)

	// Get available cards and corporations for selection
	startingCards, _ := cardRepo.GetStartingCardPool(ctx)
	corporations, _ := cardRepo.GetCorporations(ctx)
	require.GreaterOrEqual(t, len(startingCards), 4, "Should have at least 4 starting cards")
	require.GreaterOrEqual(t, len(corporations), 12, "Should have at least 12 corporations")

	// Set up starting card selection phase for player
	availableCardIDs := []string{
		startingCards[0].ID,
		startingCards[1].ID,
		startingCards[2].ID,
		startingCards[3].ID,
	}

	// For testing purposes, make all base game corporations (B01-B12) available
	// In actual gameplay, only 2 random corporations are provided
	availableCorporations := []string{
		"B01", "B02", "B03", "B04", "B05", "B06",
		"B07", "B08", "B09", "B10", "B11", "B12",
	}

	playerRepo.UpdateSelectStartingCardsPhase(ctx, game.ID, player.ID, &model.SelectStartingCardsPhase{
		AvailableCards:        availableCardIDs,
		AvailableCorporations: availableCorporations,
		SelectionComplete:     false,
	})

	return ctx, game.ID, player.ID, cardService, playerRepo, cardRepo
}
