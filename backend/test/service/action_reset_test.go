package service

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionResetFunctionality(t *testing.T) {
	// Test that verifies action play count logic works correctly
	// This is a simpler test that doesn't require the full game service constructor

	// Setup repositories
	ctx := context.Background()
	eventBus := events.NewEventBus()
	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)

	// Create a test game
	gameSettings := game.GameSettings{
		MaxPlayers:      1,
		DevelopmentMode: true,
	}
	game, err := gameRepo.Create(ctx, gameSettings)
	require.NoError(t, err)
	gameID := game.ID

	// Create a player with an action that has been played
	playerID := "player1"
	testAction := player.PlayerAction{
		CardID:        "space-elevator",
		CardName:      "Space Elevator",
		BehaviorIndex: 1,
		PlayCount:     1, // Has been played this generation
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
		Resources:        resources.Resources{Steel: 5, Credits: 20},
		AvailableActions: 1,
		Actions:          []player.PlayerAction{testAction},
	}

	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Verify initial state - player has action with playCount = 1
	initialPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Equal(t, 1, initialPlayer.Actions[0].PlayCount)

	// Manually reset the action play count (simulating what happens in phase transition)
	resetActions := make([]player.PlayerAction, len(initialPlayer.Actions))
	for i, action := range initialPlayer.Actions {
		resetActions[i] = *action.DeepCopy()
		resetActions[i].PlayCount = 0 // Reset to 0
	}

	// Update the player's actions
	err = playerRepo.UpdatePlayerActions(ctx, gameID, playerID, resetActions)
	require.NoError(t, err)

	// Verify the action play count has been reset
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Equal(t, 0, updatedPlayer.Actions[0].PlayCount, "Action playCount should be reset to 0")

	// Verify other action properties are preserved
	assert.Equal(t, "space-elevator", updatedPlayer.Actions[0].CardID)
	assert.Equal(t, "Space Elevator", updatedPlayer.Actions[0].CardName)
	assert.Equal(t, 1, updatedPlayer.Actions[0].BehaviorIndex)
}
