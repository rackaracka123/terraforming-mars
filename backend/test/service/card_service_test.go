package service

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardService_SelectStartingCards(t *testing.T) {
	// Setup
	// EventBus no longer needed
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)

	ctx := context.Background()

	// Create a test game
	game := model.NewGame("test-game", model.GameSettings{MaxPlayers: 4})
	game.Status = model.GameStatusActive
	game.CurrentPhase = model.GamePhaseStartingCardSelection

	// Create test player with starting credits
	player := model.Player{
		ID:   "player1",
		Name: "Test Player",
		Resources: model.Resources{
			Credits: 40, // Starting credits
		},
		Production: model.Production{
			Credits: 1,
		},
		TerraformRating: 20,
		IsConnected:     true,
		Cards:           []string{},
		PlayedCards:     []string{},
	}

	// Create game using clean architecture
	createdGame, err := gameRepo.Create(ctx, game.Settings)
	require.NoError(t, err)
	gameID := createdGame.ID

	// Set game status and phase using granular updates
	err = gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	require.NoError(t, err)
	err = gameRepo.UpdatePhase(ctx, gameID, model.GamePhaseStartingCardSelection)
	require.NoError(t, err)

	// Add player using clean architecture
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)
	err = gameRepo.AddPlayerID(ctx, gameID, player.ID)
	require.NoError(t, err)

	// Load cards and get real card IDs for testing
	err = cardRepo.LoadCards(context.Background())
	require.NoError(t, err, "Should load card data for testing")

	startingCards, _ := cardRepo.GetStartingCardPool(context.Background())
	require.GreaterOrEqual(t, len(startingCards), 4, "Should have at least 4 starting cards")

	// Use real card IDs from loaded data
	availableCardIDs := []string{
		startingCards[0].ID,
		startingCards[1].ID,
		startingCards[2].ID,
		startingCards[3].ID,
	}
	// Set up player's starting card selection (the cards they can choose from)
	err = playerRepo.SetStartingSelection(ctx, gameID, player.ID, availableCardIDs)
	require.NoError(t, err)

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
			selectedCards: []string{availableCardIDs[0]},
			expectedCost:  3,
			expectedError: false,
		},
		{
			name:          "Select two cards (6 MC total)",
			selectedCards: []string{availableCardIDs[0], availableCardIDs[1]},
			expectedCost:  6,
			expectedError: false,
		},
		{
			name:          "Select three cards (9 MC total)",
			selectedCards: []string{availableCardIDs[0], availableCardIDs[1], availableCardIDs[2]},
			expectedCost:  9,
			expectedError: false,
		},
		{
			name:          "Select four cards (12 MC total)",
			selectedCards: []string{availableCardIDs[0], availableCardIDs[1], availableCardIDs[2], availableCardIDs[3]},
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
			selectedCards: append(availableCardIDs, "extra-card-id"),
			expectedError: true,
			errorMessage:  "invalid card ID: extra-card-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset player state for each test
			resetPlayer := model.Player{
				ID:   "player1",
				Name: "Test Player",
				Resources: model.Resources{
					Credits: 40,
				},
				Production: model.Production{
					Credits: 1,
				},
				TerraformRating: 20,
				IsConnected:     true,
				Cards:           []string{},
				PlayedCards:     []string{},
			}

			// Update player resources using granular update
			err := playerRepo.UpdateResources(ctx, gameID, resetPlayer.ID, resetPlayer.Resources)
			require.NoError(t, err)

			// Clear any existing cards from previous test runs
			currentPlayer, err := playerRepo.GetByID(ctx, gameID, resetPlayer.ID)
			require.NoError(t, err)
			for _, cardID := range currentPlayer.Cards {
				err = playerRepo.RemoveCard(ctx, gameID, resetPlayer.ID, cardID)
				require.NoError(t, err)
			}

			// Reset player's starting selection
			err = playerRepo.SetStartingSelection(ctx, gameID, player.ID, availableCardIDs)
			require.NoError(t, err)

			// Execute
			err = cardService.OnSelectStartingCards(ctx, gameID, player.ID, tt.selectedCards)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)

				// Verify player state after selection
				updatedPlayer, err := playerRepo.GetByID(ctx, gameID, player.ID)
				require.NoError(t, err)

				// Check cards were added to player's hand
				assert.Equal(t, len(tt.selectedCards), len(updatedPlayer.Cards))
				for _, cardID := range tt.selectedCards {
					assert.Contains(t, updatedPlayer.Cards, cardID)
				}

				// Check credits were deducted correctly
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
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)

	ctx := context.Background()

	// Create a test game in starting card selection phase
	createdGame, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 2})
	require.NoError(t, err)
	gameID := createdGame.ID

	// Set game to active status and starting card selection phase
	err = gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	require.NoError(t, err)
	err = gameRepo.UpdatePhase(ctx, gameID, model.GamePhaseStartingCardSelection)
	require.NoError(t, err)

	// Create two test players
	player1 := model.Player{
		ID:   "player1",
		Name: "Test Player 1",
		Resources: model.Resources{
			Credits: 40,
		},
		Production: model.Production{
			Credits: 1,
		},
		TerraformRating: 20,
		IsConnected:     true,
		Cards:           []string{},
		PlayedCards:     []string{},
	}

	player2 := model.Player{
		ID:   "player2",
		Name: "Test Player 2",
		Resources: model.Resources{
			Credits: 40,
		},
		Production: model.Production{
			Credits: 1,
		},
		TerraformRating: 20,
		IsConnected:     true,
		Cards:           []string{},
		PlayedCards:     []string{},
	}

	// Add players to game
	err = playerRepo.Create(ctx, gameID, player1)
	require.NoError(t, err)
	err = playerRepo.Create(ctx, gameID, player2)
	require.NoError(t, err)

	// Load cards and get available starting cards
	err = cardRepo.LoadCards(ctx)
	require.NoError(t, err, "Should load card data for testing")

	availableCards, err := cardService.GetStartingCards(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(availableCards), 2)

	availableCardIDs := make([]string, len(availableCards))
	for i, card := range availableCards {
		availableCardIDs[i] = card.ID
	}

	// Set starting cards for both players
	err = playerRepo.SetStartingSelection(ctx, gameID, player1.ID, availableCardIDs)
	require.NoError(t, err)
	err = playerRepo.SetStartingSelection(ctx, gameID, player2.ID, availableCardIDs)
	require.NoError(t, err)

	// Verify game is in starting card selection phase
	game, err := gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, model.GamePhaseStartingCardSelection, game.CurrentPhase)

	// First player selects starting cards (should NOT trigger phase transition)
	err = cardService.OnSelectStartingCards(ctx, gameID, player1.ID, []string{availableCardIDs[0]})
	require.NoError(t, err)

	// Verify game is still in starting card selection phase
	game, err = gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, model.GamePhaseStartingCardSelection, game.CurrentPhase)

	// Second player selects starting cards (should trigger automatic phase transition)
	err = cardService.OnSelectStartingCards(ctx, gameID, player2.ID, []string{availableCardIDs[1]})
	require.NoError(t, err)

	// Verify game automatically transitioned to action phase
	game, err = gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)
	assert.Equal(t, model.GamePhaseAction, game.CurrentPhase, "Game should automatically transition to action phase when all players complete starting card selection")

	// Verify both players have their selected cards
	updatedPlayer1, err := playerRepo.GetByID(ctx, gameID, player1.ID)
	require.NoError(t, err)
	assert.Contains(t, updatedPlayer1.Cards, availableCardIDs[0])
	assert.Equal(t, 37, updatedPlayer1.Resources.Credits) // 40 - 3 for 1 card

	updatedPlayer2, err := playerRepo.GetByID(ctx, gameID, player2.ID)
	require.NoError(t, err)
	assert.Contains(t, updatedPlayer2.Cards, availableCardIDs[1])
	assert.Equal(t, 37, updatedPlayer2.Resources.Credits) // 40 - 3 for 1 card
}
