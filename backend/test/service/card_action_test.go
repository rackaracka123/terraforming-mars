package service

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardService_PlayCardAction_Success(t *testing.T) {
	// Setup repositories
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := NewMockCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Create a mock session manager that doesn't actually send messages
	sessionManager := &MockSessionManager{}

	// Create card service
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)

	// Setup test data
	ctx := context.Background()
	playerID := "player1"

	// Create a test game
	gameSettings := model.GameSettings{
		MaxPlayers:      4,
		DevelopmentMode: true,
	}
	game, err := gameRepo.Create(ctx, gameSettings)
	require.NoError(t, err)
	gameID := game.ID

	// Set current turn to the test player
	err = gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	require.NoError(t, err)

	// Create a test player with action and sufficient resources
	testAction := model.PlayerAction{
		CardID:        "space-elevator",
		CardName:      "Space Elevator",
		BehaviorIndex: 1,
		PlayCount:     0,
		Behavior: model.CardBehavior{
			Triggers: []model.Trigger{
				{Type: model.ResourceTriggerManual},
			},
			Inputs: []model.ResourceCondition{
				{Type: model.ResourceSteel, Amount: 1},
			},
			Outputs: []model.ResourceCondition{
				{Type: model.ResourceCredits, Amount: 5},
			},
		},
	}

	player := model.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        model.Resources{Steel: 5, Credits: 10},
		AvailableActions: 2,
		Actions:          []model.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Play the card action
	err = cardService.PlayCardAction(ctx, gameID, playerID, "space-elevator", 1)
	require.NoError(t, err)

	// Verify the player's resources were updated correctly
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)

	// Steel should be reduced by 1
	assert.Equal(t, 4, updatedPlayer.Resources.Steel)
	// Credits should be increased by 5
	assert.Equal(t, 15, updatedPlayer.Resources.Credits)
	// Available actions should be reduced by 1
	assert.Equal(t, 1, updatedPlayer.AvailableActions)
	// Play count should be incremented
	assert.Equal(t, 1, updatedPlayer.Actions[0].PlayCount)
}

func TestCardService_PlayCardAction_InsufficientResources(t *testing.T) {
	// Setup repositories
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := NewMockCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)

	ctx := context.Background()
	playerID := "player1"

	// Create a test game
	gameSettings := model.GameSettings{
		MaxPlayers:      4,
		DevelopmentMode: true,
	}
	game, err := gameRepo.Create(ctx, gameSettings)
	require.NoError(t, err)
	gameID := game.ID

	// Set current turn to the test player
	err = gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	require.NoError(t, err)

	// Create a player with insufficient resources
	testAction := model.PlayerAction{
		CardID:        "space-elevator",
		CardName:      "Space Elevator",
		BehaviorIndex: 1,
		PlayCount:     0,
		Behavior: model.CardBehavior{
			Triggers: []model.Trigger{
				{Type: model.ResourceTriggerManual},
			},
			Inputs: []model.ResourceCondition{
				{Type: model.ResourceSteel, Amount: 5}, // Need 5 steel
			},
			Outputs: []model.ResourceCondition{
				{Type: model.ResourceCredits, Amount: 25},
			},
		},
	}

	player := model.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        model.Resources{Steel: 2, Credits: 10}, // Only have 2 steel
		AvailableActions: 2,
		Actions:          []model.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Attempt to play the card action
	err = cardService.PlayCardAction(ctx, gameID, playerID, "space-elevator", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient steel: need 5, have 2")

	// Verify the player's state was not changed
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)

	assert.Equal(t, 2, updatedPlayer.Resources.Steel)      // Unchanged
	assert.Equal(t, 10, updatedPlayer.Resources.Credits)   // Unchanged
	assert.Equal(t, 2, updatedPlayer.AvailableActions)     // Unchanged
	assert.Equal(t, 0, updatedPlayer.Actions[0].PlayCount) // Unchanged
}

func TestCardService_PlayCardAction_NoAvailableActions(t *testing.T) {
	// Setup repositories
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := NewMockCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)

	ctx := context.Background()
	playerID := "player1"

	// Create a test game
	gameSettings := model.GameSettings{
		MaxPlayers:      4,
		DevelopmentMode: true,
	}
	game, err := gameRepo.Create(ctx, gameSettings)
	require.NoError(t, err)
	gameID := game.ID

	// Set current turn to the test player
	err = gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	require.NoError(t, err)

	// Create a player with no available actions
	testAction := model.PlayerAction{
		CardID:        "space-elevator",
		CardName:      "Space Elevator",
		BehaviorIndex: 1,
		PlayCount:     0,
		Behavior: model.CardBehavior{
			Triggers: []model.Trigger{
				{Type: model.ResourceTriggerManual},
			},
			Inputs: []model.ResourceCondition{
				{Type: model.ResourceSteel, Amount: 1},
			},
			Outputs: []model.ResourceCondition{
				{Type: model.ResourceCredits, Amount: 5},
			},
		},
	}

	player := model.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        model.Resources{Steel: 5, Credits: 10},
		AvailableActions: 0, // No available actions
		Actions:          []model.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Attempt to play the card action
	err = cardService.PlayCardAction(ctx, gameID, playerID, "space-elevator", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no available actions remaining")
}

func TestCardService_PlayCardAction_ActionNotFound(t *testing.T) {
	// Setup repositories
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := NewMockCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)

	ctx := context.Background()
	playerID := "player1"

	// Create a test game
	gameSettings := model.GameSettings{
		MaxPlayers:      4,
		DevelopmentMode: true,
	}
	game, err := gameRepo.Create(ctx, gameSettings)
	require.NoError(t, err)
	gameID := game.ID

	// Set current turn to the test player
	err = gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	require.NoError(t, err)

	// Create a player with no actions
	player := model.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        model.Resources{Steel: 5, Credits: 10},
		AvailableActions: 2,
		Actions:          []model.PlayerAction{}, // No actions
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Attempt to play a non-existent card action
	err = cardService.PlayCardAction(ctx, gameID, playerID, "space-elevator", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "card action not found in player's action list")
}

func TestCardService_PlayCardAction_AlreadyPlayed(t *testing.T) {
	// Setup repositories
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := NewMockCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)

	ctx := context.Background()
	playerID := "player1"

	// Create a test game
	gameSettings := model.GameSettings{
		MaxPlayers:      4,
		DevelopmentMode: true,
	}
	game, err := gameRepo.Create(ctx, gameSettings)
	require.NoError(t, err)
	gameID := game.ID

	// Set current turn to the test player
	err = gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	require.NoError(t, err)

	// Create an action that has already been played this generation
	testAction := model.PlayerAction{
		CardID:        "space-elevator",
		CardName:      "Space Elevator",
		BehaviorIndex: 1,
		PlayCount:     1, // Already played once
		Behavior: model.CardBehavior{
			Triggers: []model.Trigger{
				{Type: model.ResourceTriggerManual},
			},
			Inputs: []model.ResourceCondition{
				{Type: model.ResourceSteel, Amount: 1},
			},
			Outputs: []model.ResourceCondition{
				{Type: model.ResourceCredits, Amount: 5},
			},
		},
	}

	player := model.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        model.Resources{Steel: 5, Credits: 10},
		AvailableActions: 2,
		Actions:          []model.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Attempt to play the card action again
	err = cardService.PlayCardAction(ctx, gameID, playerID, "space-elevator", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action has already been played this generation")
}

func TestCardService_PlayCardAction_ProductionOutputs(t *testing.T) {
	// Setup repositories
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := NewMockCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)

	ctx := context.Background()
	playerID := "player1"

	// Create a test game
	gameSettings := model.GameSettings{
		MaxPlayers:      4,
		DevelopmentMode: true,
	}
	game, err := gameRepo.Create(ctx, gameSettings)
	require.NoError(t, err)
	gameID := game.ID

	// Set current turn to the test player
	err = gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	require.NoError(t, err)

	// Create an action that affects production
	testAction := model.PlayerAction{
		CardID:        "space-mirrors",
		CardName:      "Space Mirrors",
		BehaviorIndex: 0,
		PlayCount:     0,
		Behavior: model.CardBehavior{
			Triggers: []model.Trigger{
				{Type: model.ResourceTriggerManual},
			},
			Inputs: []model.ResourceCondition{
				{Type: model.ResourceCredits, Amount: 7},
			},
			Outputs: []model.ResourceCondition{
				{Type: model.ResourceEnergyProduction, Amount: 1},
			},
		},
	}

	player := model.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        model.Resources{Credits: 20},
		Production:       model.Production{Energy: 2},
		AvailableActions: 2,
		Actions:          []model.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Play the card action
	err = cardService.PlayCardAction(ctx, gameID, playerID, "space-mirrors", 0)
	require.NoError(t, err)

	// Verify the player's resources and production were updated
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)

	// Credits should be reduced by 7
	assert.Equal(t, 13, updatedPlayer.Resources.Credits)
	// Energy production should be increased by 1
	assert.Equal(t, 3, updatedPlayer.Production.Energy)
	// Available actions should be reduced by 1
	assert.Equal(t, 1, updatedPlayer.AvailableActions)
	// Play count should be incremented
	assert.Equal(t, 1, updatedPlayer.Actions[0].PlayCount)
}

func TestCardService_PlayCardAction_TerraformRating(t *testing.T) {
	// Setup repositories
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := NewMockCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)

	ctx := context.Background()
	playerID := "player1"

	// Create a test game
	gameSettings := model.GameSettings{
		MaxPlayers:      4,
		DevelopmentMode: true,
	}
	game, err := gameRepo.Create(ctx, gameSettings)
	require.NoError(t, err)
	gameID := game.ID

	// Set current turn to the test player
	err = gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	require.NoError(t, err)

	// Create an action that affects terraform rating
	testAction := model.PlayerAction{
		CardID:        "equatorial-magnetizer",
		CardName:      "Equatorial Magnetizer",
		BehaviorIndex: 0,
		PlayCount:     0,
		Behavior: model.CardBehavior{
			Triggers: []model.Trigger{
				{Type: model.ResourceTriggerManual},
			},
			Inputs: []model.ResourceCondition{},
			Outputs: []model.ResourceCondition{
				{Type: model.ResourceEnergyProduction, Amount: -1},
				{Type: model.ResourceTR, Amount: 1},
			},
		},
	}

	player := model.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        model.Resources{Credits: 20},
		Production:       model.Production{Energy: 3},
		TerraformRating:  20,
		AvailableActions: 2,
		Actions:          []model.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Play the card action
	err = cardService.PlayCardAction(ctx, gameID, playerID, "equatorial-magnetizer", 0)
	require.NoError(t, err)

	// Verify the player's production and TR were updated
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)

	// Energy production should be reduced by 1
	assert.Equal(t, 2, updatedPlayer.Production.Energy)
	// Terraform rating should be increased by 1
	assert.Equal(t, 21, updatedPlayer.TerraformRating)
	// Available actions should be reduced by 1
	assert.Equal(t, 1, updatedPlayer.AvailableActions)
	// Play count should be incremented
	assert.Equal(t, 1, updatedPlayer.Actions[0].PlayCount)
}
