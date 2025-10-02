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

func TestCardService_PlayCard_BasicValidationFlow(t *testing.T) {
	// Setup
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService)

	ctx := context.Background()

	// Load real cards
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err)

	// Get starting cards to test with
	startingCards, err := cardRepo.GetStartingCardPool(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(startingCards), 1, "Need at least one starting card for testing")

	testCardID := startingCards[0].ID

	// Create test game
	createdGame, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 2})
	require.NoError(t, err)
	gameID := createdGame.ID

	err = gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	require.NoError(t, err)
	err = gameRepo.UpdatePhase(ctx, gameID, model.GamePhaseAction)
	require.NoError(t, err)

	// Create test player with the test card
	player := model.Player{
		ID:   "player1",
		Name: "Test Player",
		Resources: model.Resources{
			Credits: 50, // Enough for most starting cards
		},
		Production: model.Production{
			Credits: 1,
		},
		TerraformRating:  20,
		IsConnected:      true,
		Cards:            []string{testCardID},
		PlayedCards:      []string{},
		AvailableActions: 2,
	}

	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)
	err = gameRepo.AddPlayerID(ctx, gameID, player.ID)
	require.NoError(t, err)

	// Set current turn
	err = gameRepo.UpdateCurrentTurn(ctx, gameID, &player.ID)
	require.NoError(t, err)

	tests := []struct {
		name          string
		cardID        string
		expectedError bool
		errorContains string
		setupFunc     func() error
		cleanupFunc   func() error
	}{
		{
			name:          "Valid card play - all validations pass",
			cardID:        testCardID,
			expectedError: false,
			setupFunc:     func() error { return nil },
			cleanupFunc:   func() error { return nil },
		},
		{
			name:          "Invalid - not player's turn",
			cardID:        testCardID,
			expectedError: true,
			errorContains: "not current player's turn",
			setupFunc: func() error {
				otherPlayer := "other-player"
				return gameRepo.UpdateCurrentTurn(ctx, gameID, &otherPlayer)
			},
			cleanupFunc: func() error {
				return gameRepo.UpdateCurrentTurn(ctx, gameID, &player.ID)
			},
		},
		{
			name:          "Invalid - no available actions",
			cardID:        testCardID,
			expectedError: true,
			errorContains: "no actions available",
			setupFunc: func() error {
				return playerRepo.UpdateAvailableActions(ctx, gameID, player.ID, 0)
			},
			cleanupFunc: func() error {
				return playerRepo.UpdateAvailableActions(ctx, gameID, player.ID, 2)
			},
		},
		{
			name:          "Invalid - card not in hand",
			cardID:        "nonexistent-card",
			expectedError: true,
			errorContains: "player does not have card",
			setupFunc:     func() error { return nil },
			cleanupFunc:   func() error { return nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test-specific state
			err := tt.setupFunc()
			require.NoError(t, err)

			// Execute
			err = cardService.OnPlayCard(ctx, gameID, player.ID, tt.cardID)

			// Cleanup
			cleanupErr := tt.cleanupFunc()
			require.NoError(t, cleanupErr)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)

				// Verify card was played successfully
				updatedPlayer, err := playerRepo.GetByID(ctx, gameID, player.ID)
				require.NoError(t, err)

				// Card should be removed from hand
				assert.NotContains(t, updatedPlayer.Cards, tt.cardID)
				// Card should be in played cards
				assert.Contains(t, updatedPlayer.PlayedCards, tt.cardID)
			}
		})
	}
}

func TestCardService_PlayCard_AffordabilityValidation(t *testing.T) {
	// Setup
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService)

	ctx := context.Background()

	// Load real cards
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err)

	// Get all cards to find one with a cost
	allCards, err := cardRepo.GetAllCards(ctx)
	require.NoError(t, err)

	// Find a card with a reasonable cost for testing
	var expensiveCard *model.Card
	for _, card := range allCards {
		if card.Cost > 30 { // Find a card that costs more than what we'll give the player
			expensiveCard = &card
			break
		}
	}
	require.NotNil(t, expensiveCard, "Need at least one card with cost > 30 for testing")

	// Create test game
	createdGame, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 2})
	require.NoError(t, err)
	gameID := createdGame.ID

	err = gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive)
	require.NoError(t, err)
	err = gameRepo.UpdatePhase(ctx, gameID, model.GamePhaseAction)
	require.NoError(t, err)

	// Create test player with limited credits
	player := model.Player{
		ID:   "player1",
		Name: "Test Player",
		Resources: model.Resources{
			Credits: 20, // Not enough for the expensive card
		},
		Production: model.Production{
			Credits: 1,
		},
		TerraformRating:  20,
		IsConnected:      true,
		Cards:            []string{expensiveCard.ID},
		PlayedCards:      []string{},
		AvailableActions: 2,
	}

	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)
	err = gameRepo.AddPlayerID(ctx, gameID, player.ID)
	require.NoError(t, err)
	err = gameRepo.UpdateCurrentTurn(ctx, gameID, &player.ID)
	require.NoError(t, err)

	// Execute
	err = cardService.OnPlayCard(ctx, gameID, player.ID, expensiveCard.ID)

	// Assert - should fail due to insufficient credits
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot afford to play card")
}
