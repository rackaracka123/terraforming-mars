package service

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardService_SelectStartingCards(t *testing.T) {
	// Setup
	eventBus := events.NewInMemoryEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	cardSelectionRepo := repository.NewCardSelectionRepository()
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, eventBus, cardDeckRepo, cardSelectionRepo)

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
		IsActive:        true,
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
	availableCards := []string{
		startingCards[0].ID,
		startingCards[1].ID,
		startingCards[2].ID,
		startingCards[3].ID,
	}
	cardServiceImpl := cardService.(*service.CardServiceImpl)
	cardServiceImpl.StorePlayerCardOptions(gameID, player.ID, availableCards)

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
			selectedCards: []string{availableCards[0]},
			expectedCost:  3,
			expectedError: false,
		},
		{
			name:          "Select two cards (6 MC total)",
			selectedCards: []string{availableCards[0], availableCards[1]},
			expectedCost:  6,
			expectedError: false,
		},
		{
			name:          "Select three cards (9 MC total)",
			selectedCards: []string{availableCards[0], availableCards[1], availableCards[2]},
			expectedCost:  9,
			expectedError: false,
		},
		{
			name:          "Select four cards (12 MC total)",
			selectedCards: []string{availableCards[0], availableCards[1], availableCards[2], availableCards[3]},
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
			selectedCards: []string{"investment", "early-settlement", "research-grant", "power-plant", "extra-card"},
			expectedError: true,
			errorMessage:  "cannot select more than 4 cards",
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
				IsActive:        true,
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

			// Reset card service selection status
			cardServiceImpl.StorePlayerCardOptions(gameID, player.ID, availableCards)

			// Execute
			err = cardService.SelectStartingCards(ctx, gameID, player.ID, tt.selectedCards)

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

func TestCardService_ValidateStartingCardSelection(t *testing.T) {
	// Setup
	eventBus := events.NewInMemoryEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	cardSelectionRepo := repository.NewCardSelectionRepository()
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, eventBus, cardDeckRepo, cardSelectionRepo)

	ctx := context.Background()

	// Load cards and get real card IDs for testing
	err := cardRepo.LoadCards(context.Background())
	require.NoError(t, err, "Should load card data for testing")

	startingCards, _ := cardRepo.GetStartingCardPool(context.Background())
	require.GreaterOrEqual(t, len(startingCards), 4, "Should have at least 4 starting cards")

	// Store starting card options using real card IDs
	availableCards := []string{
		startingCards[0].ID,
		startingCards[1].ID,
		startingCards[2].ID,
		startingCards[3].ID,
	}

	// Get a card that exists but is NOT in the player's options
	var cardNotInOptions string
	for _, card := range startingCards {
		found := false
		for _, available := range availableCards {
			if card.ID == available {
				found = true
				break
			}
		}
		if !found {
			cardNotInOptions = card.ID
			break
		}
	}

	cardServiceImpl := cardService.(*service.CardServiceImpl)
	cardServiceImpl.StorePlayerCardOptions("test-game", "player1", availableCards)

	tests := []struct {
		name          string
		cardIDs       []string
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "Valid selection from options",
			cardIDs:       []string{availableCards[0], availableCards[1]},
			expectedError: false,
		},
		{
			name:          "Empty selection is valid",
			cardIDs:       []string{},
			expectedError: false,
		},
		{
			name:          "Card not in player's options",
			cardIDs:       []string{cardNotInOptions}, // Real card but not in player's options
			expectedError: true,
			errorMessage:  "not in player's available options",
		},
		{
			name:          "Too many cards",
			cardIDs:       []string{availableCards[0], availableCards[1], availableCards[2], availableCards[3], "extra"},
			expectedError: true,
			errorMessage:  "cannot select more than 4 cards, got 5",
		},
		{
			name:          "Non-existent card ID",
			cardIDs:       []string{"fake-card-id"},
			expectedError: true,
			errorMessage:  "invalid card ID: fake-card-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset selection status
			cardServiceImpl.StorePlayerCardOptions("test-game", "player1", availableCards)

			err := cardService.ValidateStartingCardSelection(ctx, "test-game", "player1", tt.cardIDs)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardService_IsAllPlayersCardSelectionComplete(t *testing.T) {
	// Setup
	eventBus := events.NewInMemoryEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	cardSelectionRepo := repository.NewCardSelectionRepository()
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, eventBus, cardDeckRepo, cardSelectionRepo)

	ctx := context.Background()

	// Create a test game with multiple players
	game := model.NewGame("test-game", model.GameSettings{MaxPlayers: 4})
	game.Status = model.GameStatusActive
	game.CurrentPhase = model.GamePhaseStartingCardSelection

	// Create test players
	player1 := model.Player{
		ID:              "player1",
		Name:            "Player 1",
		Resources:       model.Resources{Credits: 40},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsActive:        true,
	}

	player2 := model.Player{
		ID:              "player2",
		Name:            "Player 2",
		Resources:       model.Resources{Credits: 40},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsActive:        true,
	}

	// Create game using clean architecture
	createdGame, err := gameRepo.Create(ctx, game.Settings)
	require.NoError(t, err)
	gameID := createdGame.ID

	// Set game status and phase
	err = gameRepo.UpdateStatus(ctx, gameID, game.Status)
	require.NoError(t, err)
	err = gameRepo.UpdatePhase(ctx, gameID, game.CurrentPhase)
	require.NoError(t, err)

	// Add players using clean architecture
	err = playerRepo.Create(ctx, gameID, player1)
	require.NoError(t, err)
	err = gameRepo.AddPlayerID(ctx, gameID, player1.ID)
	require.NoError(t, err)
	err = playerRepo.Create(ctx, gameID, player2)
	require.NoError(t, err)
	err = gameRepo.AddPlayerID(ctx, gameID, player2.ID)
	require.NoError(t, err)

	// Load cards and get real card IDs for testing
	err = cardRepo.LoadCards(context.Background())
	require.NoError(t, err, "Should load card data for testing")

	startingCards, _ := cardRepo.GetStartingCardPool(context.Background())
	require.GreaterOrEqual(t, len(startingCards), 4, "Should have at least 4 starting cards")

	cardServiceImpl := cardService.(*service.CardServiceImpl)

	// Test: No players have selection data
	complete := cardService.IsAllPlayersCardSelectionComplete(ctx, gameID)
	assert.False(t, complete)

	// Setup card options for both players using real card IDs
	availableCards := []string{
		startingCards[0].ID,
		startingCards[1].ID,
		startingCards[2].ID,
		startingCards[3].ID,
	}
	cardServiceImpl.StorePlayerCardOptions(gameID, player1.ID, availableCards)
	cardServiceImpl.StorePlayerCardOptions(gameID, player2.ID, availableCards)

	// Test: No players have completed selection
	complete = cardService.IsAllPlayersCardSelectionComplete(ctx, gameID)
	assert.False(t, complete)

	// Test: Only one player completed selection
	err = cardService.SelectStartingCards(ctx, gameID, player1.ID, []string{availableCards[0]})
	require.NoError(t, err)

	complete = cardService.IsAllPlayersCardSelectionComplete(ctx, gameID)
	assert.False(t, complete)

	// Test: All players completed selection
	err = cardService.SelectStartingCards(ctx, gameID, player2.ID, []string{availableCards[1]})
	require.NoError(t, err)

	complete = cardService.IsAllPlayersCardSelectionComplete(ctx, gameID)
	assert.True(t, complete)
}
