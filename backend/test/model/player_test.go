package model_test

import (
	"terraforming-mars-backend/internal/session/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayer_InitialState(t *testing.T) {
	player := &types.Player{
		ID:   "player1",
		Name: "Test Player",
	}

	// Test default values
	assert.Equal(t, "player1", player.ID)
	assert.Equal(t, "Test Player", player.Name)
	assert.Empty(t, player.Corporation)
	assert.Empty(t, player.Cards)
	assert.Empty(t, player.PlayedCards)
	assert.Equal(t, 0, player.TerraformRating)
	assert.False(t, player.IsConnected)

	// Test zero-value resources and production
	assert.Equal(t, 0, player.Resources.Credits)
	assert.Equal(t, 0, player.Production.Credits)
}

func TestPlayer_ResourcesAndProduction(t *testing.T) {
	player := &types.Player{
		Resources: types.Resources{
			Credits:  40,
			Steel:    8,
			Titanium: 3,
			Plants:   12,
			Energy:   6,
			Heat:     15,
		},
		Production: types.Production{
			Credits:  1,
			Steel:    2,
			Titanium: 1,
			Plants:   0,
			Energy:   3,
			Heat:     0,
		},
	}

	// Test resources
	assert.Equal(t, 40, player.Resources.Credits)
	assert.Equal(t, 8, player.Resources.Steel)
	assert.Equal(t, 3, player.Resources.Titanium)
	assert.Equal(t, 12, player.Resources.Plants)
	assert.Equal(t, 6, player.Resources.Energy)
	assert.Equal(t, 15, player.Resources.Heat)

	// Test production
	assert.Equal(t, 1, player.Production.Credits)
	assert.Equal(t, 2, player.Production.Steel)
	assert.Equal(t, 1, player.Production.Titanium)
	assert.Equal(t, 0, player.Production.Plants)
	assert.Equal(t, 3, player.Production.Energy)
	assert.Equal(t, 0, player.Production.Heat)
}
