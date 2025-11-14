package service

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardService_OnPlayCardAction_Success(t *testing.T) {
	// Setup test data
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Setup repositories
	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	cardRepo := NewMockCardRepository()
	cardDeckRepo := game.NewCardDeckRepository()

	// Create a mock session manager that doesn't actually send messages
	sessionManager := &MockSessionManager{}

	// Create card service
	boardService := service.NewBoardService()
	tilesRepo := tiles.NewRepository(gameRepo, playerRepo)
	tilesBoardAdapter := tiles.NewBoardServiceAdapter(boardService)
	tilesFeature := tiles.NewService(tilesRepo, tilesBoardAdapter, eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tilesFeature, effectSubscriber, forcedActionManager)
	playerID := "player1"

	// Create a test game
	gameSettings := game.GameSettings{
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
	testAction := player.PlayerAction{
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

	player := player.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        resources.Resources{Steel: 5, Credits: 10},
		AvailableActions: 2,
		Actions:          []player.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Play the card action
	err = cardService.OnPlayCardAction(ctx, gameID, playerID, "space-elevator", 1, nil, nil)
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

func TestCardService_OnPlayCardAction_InsufficientResources(t *testing.T) {
	// Setup repositories
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	cardRepo := NewMockCardRepository()
	cardDeckRepo := game.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	boardService := service.NewBoardService()
	tilesRepo := tiles.NewRepository(gameRepo, playerRepo)
	tilesBoardAdapter := tiles.NewBoardServiceAdapter(boardService)
	tilesFeature := tiles.NewService(tilesRepo, tilesBoardAdapter, eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tilesFeature, effectSubscriber, forcedActionManager)

	playerID := "player1"

	// Create a test game
	gameSettings := game.GameSettings{
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
	testAction := player.PlayerAction{
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

	player := player.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        resources.Resources{Steel: 2, Credits: 10}, // Only have 2 steel
		AvailableActions: 2,
		Actions:          []player.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Attempt to play the card action
	err = cardService.OnPlayCardAction(ctx, gameID, playerID, "space-elevator", 1, nil, nil)
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

func TestCardService_OnPlayCardAction_NoAvailableActions(t *testing.T) {
	// Setup repositories
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	cardRepo := NewMockCardRepository()
	cardDeckRepo := game.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	boardService := service.NewBoardService()
	tilesRepo := tiles.NewRepository(gameRepo, playerRepo)
	tilesBoardAdapter := tiles.NewBoardServiceAdapter(boardService)
	tilesFeature := tiles.NewService(tilesRepo, tilesBoardAdapter, eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tilesFeature, effectSubscriber, forcedActionManager)

	playerID := "player1"

	// Create a test game
	gameSettings := game.GameSettings{
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
	testAction := player.PlayerAction{
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

	player := player.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        resources.Resources{Steel: 5, Credits: 10},
		AvailableActions: 0, // No available actions
		Actions:          []player.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Attempt to play the card action
	err = cardService.OnPlayCardAction(ctx, gameID, playerID, "space-elevator", 1, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no available actions remaining")
}

func TestCardService_OnPlayCardAction_ActionNotFound(t *testing.T) {
	// Setup repositories
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	cardRepo := NewMockCardRepository()
	cardDeckRepo := game.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	boardService := service.NewBoardService()
	tilesRepo := tiles.NewRepository(gameRepo, playerRepo)
	tilesBoardAdapter := tiles.NewBoardServiceAdapter(boardService)
	tilesFeature := tiles.NewService(tilesRepo, tilesBoardAdapter, eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tilesFeature, effectSubscriber, forcedActionManager)

	playerID := "player1"

	// Create a test game
	gameSettings := game.GameSettings{
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
	player := player.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        resources.Resources{Steel: 5, Credits: 10},
		AvailableActions: 2,
		Actions:          []player.PlayerAction{}, // No actions
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Attempt to play a non-existent card action
	err = cardService.OnPlayCardAction(ctx, gameID, playerID, "space-elevator", 1, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "card action not found in player's action list")
}

func TestCardService_OnPlayCardAction_AlreadyPlayed(t *testing.T) {
	// Setup repositories
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	cardRepo := NewMockCardRepository()
	cardDeckRepo := game.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	boardService := service.NewBoardService()
	tilesRepo := tiles.NewRepository(gameRepo, playerRepo)
	tilesBoardAdapter := tiles.NewBoardServiceAdapter(boardService)
	tilesFeature := tiles.NewService(tilesRepo, tilesBoardAdapter, eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tilesFeature, effectSubscriber, forcedActionManager)

	playerID := "player1"

	// Create a test game
	gameSettings := game.GameSettings{
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
	testAction := player.PlayerAction{
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

	player := player.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        resources.Resources{Steel: 5, Credits: 10},
		AvailableActions: 2,
		Actions:          []player.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Attempt to play the card action again
	err = cardService.OnPlayCardAction(ctx, gameID, playerID, "space-elevator", 1, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action has already been played this generation")
}

func TestCardService_OnPlayCardAction_ProductionOutputs(t *testing.T) {
	// Setup repositories
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	cardRepo := NewMockCardRepository()
	cardDeckRepo := game.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	boardService := service.NewBoardService()
	tilesRepo := tiles.NewRepository(gameRepo, playerRepo)
	tilesBoardAdapter := tiles.NewBoardServiceAdapter(boardService)
	tilesFeature := tiles.NewService(tilesRepo, tilesBoardAdapter, eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tilesFeature, effectSubscriber, forcedActionManager)

	playerID := "player1"

	// Create a test game
	gameSettings := game.GameSettings{
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
	testAction := player.PlayerAction{
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

	player := player.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        resources.Resources{Credits: 20},
		Production:       resources.Production{Energy: 2},
		AvailableActions: 2,
		Actions:          []player.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Play the card action
	err = cardService.OnPlayCardAction(ctx, gameID, playerID, "space-mirrors", 0, nil, nil)
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

func TestCardService_OnPlayCardAction_TerraformRating(t *testing.T) {
	// Setup repositories
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	cardRepo := NewMockCardRepository()
	cardDeckRepo := game.NewCardDeckRepository()
	sessionManager := &MockSessionManager{}

	boardService := service.NewBoardService()
	tilesRepo := tiles.NewRepository(gameRepo, playerRepo)
	tilesBoardAdapter := tiles.NewBoardServiceAdapter(boardService)
	tilesFeature := tiles.NewService(tilesRepo, tilesBoardAdapter, eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tilesFeature, effectSubscriber, forcedActionManager)

	playerID := "player1"

	// Create a test game
	gameSettings := game.GameSettings{
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
	testAction := player.PlayerAction{
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

	player := player.Player{
		ID:               playerID,
		Name:             "Test Player",
		Resources:        resources.Resources{Credits: 20},
		Production:       resources.Production{Energy: 3},
		TerraformRating:  20,
		AvailableActions: 2,
		Actions:          []player.PlayerAction{testAction},
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Play the card action
	err = cardService.OnPlayCardAction(ctx, gameID, playerID, "equatorial-magnetizer", 0, nil, nil)
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
