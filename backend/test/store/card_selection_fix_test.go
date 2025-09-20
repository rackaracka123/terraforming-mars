package store

import (
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleSelectStartingCards_AddsCardsToPlayer(t *testing.T) {
	// Create a new application state
	state := store.NewApplicationState()

	// Create a test game
	gameID := "test-game-123"
	game := &model.Game{
		ID:           gameID,
		Status:       model.GameStatusActive,
		PlayerIDs:    []string{"player-1"},
		CurrentPhase: model.GamePhaseStartingCardSelection,
		GlobalParameters: model.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
		Generation: 1,
	}
	gameState := store.NewGameState(*game)
	state = state.WithGame(gameID, gameState)

	// Create a test player
	playerID := "player-1"
	player := model.Player{
		ID:   playerID,
		Name: "TestPlayer",
		Resources: model.Resources{
			Credits: 42, // Enough to buy cards
		},
		Cards: []string{}, // Start with no cards
	}
	playerState := store.NewPlayerState(player, gameID)
	state = state.WithPlayer(playerID, playerState)

	// Create a select starting cards action
	selectedCards := []string{"card-1", "card-2", "card-3"}
	cost := len(selectedCards) * 3 // 9 credits total

	payload := store.SelectStartingCardsPayload{
		GameID:        gameID,
		PlayerID:      playerID,
		SelectedCards: selectedCards,
		Cost:          cost,
	}

	action := store.NewAction(
		store.ActionSelectStartingCards,
		payload,
		gameID,
		playerID,
		"test",
	)

	// Apply the action using the GameReducer
	newState, err := store.GameReducer(state, action)
	require.NoError(t, err)

	// Verify the cards were added to the player
	updatedPlayerState, exists := newState.GetPlayer(playerID)
	require.True(t, exists)
	updatedPlayer := updatedPlayerState.Player()

	// Check that the player now has the selected cards
	assert.Equal(t, selectedCards, updatedPlayer.Cards, "Player should have the selected cards")

	// Check that the cost was deducted
	expectedCredits := 42 - cost // 42 - 9 = 33
	assert.Equal(t, expectedCredits, updatedPlayer.Resources.Credits, "Credits should be deducted correctly")

	// Check that startingSelection was cleared
	assert.Nil(t, updatedPlayer.StartingSelection, "StartingSelection should be cleared")
}

func TestHandleSelectStartingCards_HandlesEmptySelection(t *testing.T) {
	// Create a new application state
	state := store.NewApplicationState()

	// Create a test game
	gameID := "test-game-456"
	game := &model.Game{
		ID:           gameID,
		Status:       model.GameStatusActive,
		PlayerIDs:    []string{"player-1"},
		CurrentPhase: model.GamePhaseStartingCardSelection,
		GlobalParameters: model.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
		Generation: 1,
	}
	gameState := store.NewGameState(*game)
	state = state.WithGame(gameID, gameState)

	// Create a test player
	playerID := "player-1"
	player := model.Player{
		ID:   playerID,
		Name: "TestPlayer",
		Resources: model.Resources{
			Credits: 42,
		},
		Cards: []string{}, // Start with no cards
	}
	playerState := store.NewPlayerState(player, gameID)
	state = state.WithPlayer(playerID, playerState)

	// Create a select starting cards action with no cards (skip selection)
	selectedCards := []string{}
	cost := 0 // No cost for skipping

	payload := store.SelectStartingCardsPayload{
		GameID:        gameID,
		PlayerID:      playerID,
		SelectedCards: selectedCards,
		Cost:          cost,
	}

	action := store.NewAction(
		store.ActionSelectStartingCards,
		payload,
		gameID,
		playerID,
		"test",
	)

	// Apply the action using the GameReducer
	newState, err := store.GameReducer(state, action)
	require.NoError(t, err)

	// Verify no cards were added
	updatedPlayerState, exists := newState.GetPlayer(playerID)
	require.True(t, exists)
	updatedPlayer := updatedPlayerState.Player()

	// Check that the player has no cards
	assert.Equal(t, []string{}, updatedPlayer.Cards, "Player should have no cards when skipping selection")

	// Check that no credits were deducted
	assert.Equal(t, 42, updatedPlayer.Resources.Credits, "Credits should not be deducted when skipping selection")

	// Check that startingSelection was cleared
	assert.Nil(t, updatedPlayer.StartingSelection, "StartingSelection should be cleared")
}
