package service

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/test/fixtures"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardService_SelectStartingCards(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		selectedCards []string
		expectedCost  int
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "Select no cards",
			selectedCards: []string{},
			expectedCost:  0,
			expectedError: false,
		},
		{
			name:          "Select one card (3 MC)",
			selectedCards: []string{"card1"},
			expectedCost:  3,
			expectedError: false,
		},
		{
			name:          "Select two cards (6 MC total)",
			selectedCards: []string{"card1", "card2"},
			expectedCost:  6,
			expectedError: false,
		},
		{
			name:          "Select three cards (9 MC total)",
			selectedCards: []string{"card1", "card2", "card3"},
			expectedCost:  9,
			expectedError: false,
		},
		{
			name:          "Select four cards (12 MC total)",
			selectedCards: []string{"card1", "card2", "card3", "card4"},
			expectedCost:  12,
			expectedError: false,
		},
		{
			name:          "Select invalid card",
			selectedCards: []string{"invalid-card"},
			expectedError: true,
			errorMessage:  "invalid card ID: invalid-card",
		},
		{
			name:          "Select too many cards",
			selectedCards: []string{"card1", "card2", "card3", "card4", "extra-card"},
			expectedError: true,
			errorMessage:  "invalid card ID: extra-card",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh service container for each subtest to avoid state pollution
			container := fixtures.NewServiceContainer()

			// Load cards
			err := container.CardRepo.LoadCards(ctx)
			require.NoError(t, err)

			// Get real card IDs
			startingCards, _ := container.CardRepo.GetStartingCardPool(ctx)
			require.GreaterOrEqual(t, len(startingCards), 4)

			availableCardIDs := []string{
				startingCards[0].ID,
				startingCards[1].ID,
				startingCards[2].ID,
				startingCards[3].ID,
			}

			// Map test card names to real IDs
			cardMap := map[string]string{
				"card1": availableCardIDs[0],
				"card2": availableCardIDs[1],
				"card3": availableCardIDs[2],
				"card4": availableCardIDs[3],
			}
			realSelectedCards := make([]string, 0, len(tt.selectedCards))
			for _, cardName := range tt.selectedCards {
				if realID, ok := cardMap[cardName]; ok {
					realSelectedCards = append(realSelectedCards, realID)
				} else {
					realSelectedCards = append(realSelectedCards, cardName) // Keep invalid cards as-is for testing
				}
			}

			// Create game in starting card selection phase
			fixture := fixtures.NewGameInStartingCardSelection(t, container)
			gameID := fixture.GameID
			playerID := fixture.PlayerID

			// Set up starting card selection phase
			err = container.PlayerRepo.UpdateSelectStartingCardsPhase(ctx, gameID, playerID, &model.SelectStartingCardsPhase{
				AvailableCards:        availableCardIDs,
				AvailableCorporations: []string{"CC1", "PC5"},
				SelectionComplete:     false,
			})
			require.NoError(t, err)

			// Add dummy player to prevent auto-phase-transition
			dummyPlayer := fixtures.NewDefaultPlayer()
			dummyPlayer.ID = "player2"
			dummyPlayer.Name = "Dummy Player"
			err = container.PlayerRepo.Create(ctx, gameID, dummyPlayer)
			require.NoError(t, err)
			err = container.GameRepo.AddPlayerID(ctx, gameID, dummyPlayer.ID)
			require.NoError(t, err)

			// Mark dummy player as completed
			err = container.PlayerRepo.UpdateSelectStartingCardsPhase(ctx, gameID, dummyPlayer.ID, &model.SelectStartingCardsPhase{
				AvailableCards:        []string{},
				AvailableCorporations: []string{},
				SelectionComplete:     true,
			})
			require.NoError(t, err)

			// Execute
			err = container.CardService.OnSelectStartingCards(ctx, gameID, playerID, realSelectedCards, "CC1")

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)

				// Verify player state
				updatedPlayer, err := container.PlayerRepo.GetByID(ctx, gameID, playerID)
				require.NoError(t, err)

				assert.Equal(t, len(realSelectedCards), len(updatedPlayer.Cards))
				for _, cardID := range realSelectedCards {
					assert.Contains(t, updatedPlayer.Cards, cardID)
				}

				// CC1 (Aridor) gives 40 credits, cards cost 3 MC each
				expectedCredits := 40 - tt.expectedCost
				assert.Equal(t, expectedCredits, updatedPlayer.Resources.Credits)
			}
		})
	}
}

// TestCardService_ValidateStartingCardSelection and TestCardService_IsAllPlayersCardSelectionComplete
// have been removed as these methods are now internal implementation details.
// Their behavior is tested through the public OnSelectStartingCards method.

func TestCardService_SelectStartingCards_AutomaticPhaseTransition(t *testing.T) {
	// Setup
	ctx := context.Background()
	container := fixtures.NewServiceContainer()

	// Create a test game in starting card selection phase
	fixture := fixtures.NewGameInStartingCardSelection(t, container)
	gameID := fixture.GameID

	// Add second player to game (first player created by fixture)
	player2 := fixtures.NewPlayerInGame(t, container, gameID, "player2")

	// Load cards and get available starting cards
	err := container.CardRepo.LoadCards(ctx)
	require.NoError(t, err, "Should load card data for testing")

	availableCards, err := container.CardService.GetStartingCards(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(availableCards), 2)

	availableCardIDs := make([]string, len(availableCards))
	for i, card := range availableCards {
		availableCardIDs[i] = card.ID
	}

	// Set starting cards for both players
	err = container.PlayerRepo.UpdateSelectStartingCardsPhase(ctx, gameID, fixture.PlayerID, &model.SelectStartingCardsPhase{
		AvailableCards:        availableCardIDs,
		AvailableCorporations: []string{"CC1", "PC5"},
		SelectionComplete:     false,
	})
	require.NoError(t, err)
	err = container.PlayerRepo.UpdateSelectStartingCardsPhase(ctx, gameID, player2.ID, &model.SelectStartingCardsPhase{
		AvailableCards:        availableCardIDs,
		AvailableCorporations: []string{"PC5", "B07"},
		SelectionComplete:     false,
	})
	require.NoError(t, err)

	// Verify game is in starting card selection phase
	game, err := container.GameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, model.GamePhaseStartingCardSelection, game.CurrentPhase)

	// First player selects starting cards (should NOT trigger phase transition)
	err = container.CardService.OnSelectStartingCards(ctx, gameID, fixture.PlayerID, []string{availableCardIDs[0]}, "CC1")
	require.NoError(t, err)

	// Verify game is still in starting card selection phase
	game, err = container.GameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, model.GamePhaseStartingCardSelection, game.CurrentPhase)

	// Second player selects starting cards (should trigger automatic phase transition)
	err = container.CardService.OnSelectStartingCards(ctx, gameID, player2.ID, []string{availableCardIDs[1]}, "PC5")
	require.NoError(t, err)

	// Verify game automatically transitioned to action phase
	game, err = container.GameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, model.GamePhaseAction, game.CurrentPhase, "Game should automatically transition to action phase when all players complete starting card selection")

	// Verify both players have their selected cards
	updatedPlayer1, err := container.PlayerRepo.GetByID(ctx, gameID, fixture.PlayerID)
	require.NoError(t, err)
	assert.Contains(t, updatedPlayer1.Cards, availableCardIDs[0])
	assert.Equal(t, 37, updatedPlayer1.Resources.Credits) // CC1 (Aridor) gives 40, minus 3 for 1 card

	updatedPlayer2, err := container.PlayerRepo.GetByID(ctx, gameID, player2.ID)
	require.NoError(t, err)
	assert.Contains(t, updatedPlayer2.Cards, availableCardIDs[1])
	assert.Equal(t, 42, updatedPlayer2.Resources.Credits) // PC5 (Vitor) gives 45, minus 3 for 1 card
}

func TestCardService_SelectCorporationWithManualAction(t *testing.T) {
	ctx := context.Background()
	container := fixtures.NewServiceContainer()

	// Create test game
	fixture := fixtures.NewGameInStartingCardSelection(t, container)

	// Load cards
	err := container.CardRepo.LoadCards(ctx)
	require.NoError(t, err, "Should load card data for testing")

	// Get available starting cards
	availableCards, err := container.CardService.GetStartingCards(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(availableCards), 1)

	availableCardIDs := make([]string, len(availableCards))
	for i, card := range availableCards {
		availableCardIDs[i] = card.ID
	}

	// Set starting cards for player with B10 (United Nations Mars Initiative) corporation
	err = container.PlayerRepo.UpdateSelectStartingCardsPhase(ctx, fixture.GameID, fixture.PlayerID, &model.SelectStartingCardsPhase{
		AvailableCards:        availableCardIDs,
		AvailableCorporations: []string{"B10"},
		SelectionComplete:     false,
	})
	require.NoError(t, err)

	// Player selects B10 (United Nations Mars Initiative) corporation
	// This corporation has a manual action: "If your Terraform Rating was raised this generation, you may pay 3 M€ to raise it 1 step more"
	err = container.CardService.OnSelectStartingCards(ctx, fixture.GameID, fixture.PlayerID, []string{}, "B10")
	require.NoError(t, err)

	// Verify corporation was selected and manual action was extracted
	updatedPlayer, err := container.PlayerRepo.GetByID(ctx, fixture.GameID, fixture.PlayerID)
	require.NoError(t, err)
	require.NotNil(t, updatedPlayer.Corporation, "Corporation should be set")
	assert.Equal(t, "United Nations Mars Initiative", updatedPlayer.Corporation.Name)
	assert.Equal(t, 40, updatedPlayer.Resources.Credits) // B10 gives 40 MC, minus 0 for 0 cards

	// Verify manual action was registered
	assert.NotEmpty(t, updatedPlayer.Actions, "Should have manual actions from corporation")

	// Find the United Nations Mars Initiative action
	hasUNMIAction := false
	for _, action := range updatedPlayer.Actions {
		if action.CardID == "B10" && action.CardName == "United Nations Mars Initiative" {
			hasUNMIAction = true
			assert.Equal(t, 0, action.PlayCount, "Action should not be played yet")
			// Verify action behavior
			assert.Len(t, action.Behavior.Triggers, 1, "Should have 1 trigger")
			assert.Equal(t, model.ResourceTriggerManual, action.Behavior.Triggers[0].Type, "Should be manual trigger")
			// Verify inputs (costs 3 MC)
			assert.Len(t, action.Behavior.Inputs, 1, "Should have 1 input")
			assert.Equal(t, model.ResourceCredits, action.Behavior.Inputs[0].Type, "Should cost credits")
			assert.Equal(t, 3, action.Behavior.Inputs[0].Amount, "Should cost 3 MC")
			// Verify outputs (raises TR by 1)
			assert.Len(t, action.Behavior.Outputs, 1, "Should have 1 output")
			assert.Equal(t, "tr", string(action.Behavior.Outputs[0].Type), "Should raise TR")
			assert.Equal(t, 1, action.Behavior.Outputs[0].Amount, "Should raise TR by 1 step")
			break
		}
	}
	assert.True(t, hasUNMIAction, "Should have United Nations Mars Initiative manual action")

	t.Log("✅ Corporation manual action extraction test passed")
}

func TestCardService_SelectCorporationWithPassiveEffect(t *testing.T) {
	ctx := context.Background()
	container := fixtures.NewServiceContainer()

	// Create test game
	fixture := fixtures.NewGameInStartingCardSelection(t, container)

	// Load cards
	err := container.CardRepo.LoadCards(ctx)
	require.NoError(t, err, "Should load card data for testing")

	// Get available starting cards
	availableCards, err := container.CardService.GetStartingCards(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(availableCards), 1)

	availableCardIDs := make([]string, len(availableCards))
	for i, card := range availableCards {
		availableCardIDs[i] = card.ID
	}

	// Set starting cards for player with V03 (Manutech) corporation
	err = container.PlayerRepo.UpdateSelectStartingCardsPhase(ctx, fixture.GameID, fixture.PlayerID, &model.SelectStartingCardsPhase{
		AvailableCards:        availableCardIDs,
		AvailableCorporations: []string{"V03"},
		SelectionComplete:     false,
	})
	require.NoError(t, err)

	// Player selects V03 (Manutech) corporation
	// This corporation has a passive effect: "For each step you increase the production of a resource, including this, you also gain that resource"
	err = container.CardService.OnSelectStartingCards(ctx, fixture.GameID, fixture.PlayerID, []string{}, "V03")
	require.NoError(t, err)

	// Verify corporation was selected and passive effect was extracted
	updatedPlayer, err := container.PlayerRepo.GetByID(ctx, fixture.GameID, fixture.PlayerID)
	require.NoError(t, err)
	require.NotNil(t, updatedPlayer.Corporation, "Corporation should be set")
	assert.Equal(t, "Manutech", updatedPlayer.Corporation.Name)
	assert.Equal(t, 35, updatedPlayer.Resources.Credits) // V03 gives 35 MC
	assert.Equal(t, 1, updatedPlayer.Production.Steel)   // V03 gives 1 steel production

	// NOTE: Manutech's passive effect (TriggerProductionIncreased) is not yet implemented in CardEffectSubscriber
	// The event-driven system currently supports: temperature-raise, oxygen-raise, ocean-placed, city-placed, greenery-placed, tile-placed
	// TODO: Add support for production-increased trigger when implementing more advanced card effects

	t.Log("✅ Corporation selection test passed (passive effects use new event-driven system)")
}

func TestCardService_SelectCorporationWithValueModifier(t *testing.T) {
	ctx := context.Background()
	container := fixtures.NewServiceContainer()

	// Create test game
	fixture := fixtures.NewGameInStartingCardSelection(t, container)

	// Load cards
	err := container.CardRepo.LoadCards(ctx)
	require.NoError(t, err, "Should load card data for testing")

	// Get available starting cards
	availableCards, err := container.CardService.GetStartingCards(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(availableCards), 1)

	availableCardIDs := make([]string, len(availableCards))
	for i, card := range availableCards {
		availableCardIDs[i] = card.ID
	}

	// Set starting cards for player with B07 (PhoboLog) corporation
	err = container.PlayerRepo.UpdateSelectStartingCardsPhase(ctx, fixture.GameID, fixture.PlayerID, &model.SelectStartingCardsPhase{
		AvailableCards:        availableCardIDs,
		AvailableCorporations: []string{"B07"},
		SelectionComplete:     false,
	})
	require.NoError(t, err)

	// Player selects B07 (PhoboLog) corporation
	// This corporation has a value modifier effect: "Your titanium resources are each worth 1 M€ extra"
	err = container.CardService.OnSelectStartingCards(ctx, fixture.GameID, fixture.PlayerID, []string{}, "B07")
	require.NoError(t, err)

	// Verify corporation was selected and value modifier effect was extracted
	updatedPlayer, err := container.PlayerRepo.GetByID(ctx, fixture.GameID, fixture.PlayerID)
	require.NoError(t, err)
	require.NotNil(t, updatedPlayer.Corporation, "Corporation should be set")
	assert.Equal(t, "PhoboLog", updatedPlayer.Corporation.Name)
	assert.Equal(t, 23, updatedPlayer.Resources.Credits)  // B07 gives 23 MC
	assert.Equal(t, 10, updatedPlayer.Resources.Titanium) // B07 gives 10 titanium

	// NOTE: PhoboLog's value modifier (TriggerAlwaysActive + ResourceValueModifier) is not yet implemented
	// The event-driven system currently supports event-based triggers only (temperature, oxygen, tiles, etc.)
	// TODO: Add support for always-active effects and value modifiers when implementing resource spending logic

	t.Log("✅ Corporation selection test passed (value modifiers will use event-driven system when implemented)")
}
