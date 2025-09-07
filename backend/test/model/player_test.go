package model_test

import (
	"terraforming-mars-backend/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayer_InitialState(t *testing.T) {
	player := &model.Player{
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
	assert.False(t, player.IsActive)

	// Test zero-value resources and production
	assert.Equal(t, 0, player.Resources.Credits)
	assert.Equal(t, 0, player.Production.Credits)
}

func TestPlayer_DeepCopy(t *testing.T) {
	original := &model.Player{
		ID:          "player1",
		Name:        "Test Player",
		Corporation: "Tharsis Republic",
		Cards:       []string{"card1", "card2"},
		Resources: model.Resources{
			Credits:  50,
			Steel:    10,
			Titanium: 5,
		},
		Production: model.Production{
			Credits: 2,
			Steel:   1,
		},
		TerraformRating: 25,
		IsActive:        true,
		PlayedCards:     []string{"played1", "played2"},
	}

	copy := original.DeepCopy()

	// Should be equal but different pointers
	assert.Equal(t, original.ID, copy.ID)
	assert.Equal(t, original.Name, copy.Name)
	assert.Equal(t, original.Corporation, copy.Corporation)
	assert.Equal(t, original.Cards, copy.Cards)
	assert.Equal(t, original.Resources, copy.Resources)
	assert.Equal(t, original.Production, copy.Production)
	assert.Equal(t, original.TerraformRating, copy.TerraformRating)
	assert.Equal(t, original.IsActive, copy.IsActive)
	assert.Equal(t, original.PlayedCards, copy.PlayedCards)
	assert.NotSame(t, original, copy)

	// Modifying copy should not affect original
	copy.Name = "Modified Player"
	copy.Cards[0] = "modified_card"
	copy.Resources.Credits = 100
	copy.Production.Credits = 5
	copy.PlayedCards[0] = "modified_played"

	assert.Equal(t, "Test Player", original.Name)
	assert.Equal(t, "Modified Player", copy.Name)
	assert.Equal(t, "card1", original.Cards[0])
	assert.Equal(t, "modified_card", copy.Cards[0])
	assert.Equal(t, 50, original.Resources.Credits)
	assert.Equal(t, 100, copy.Resources.Credits)
	assert.Equal(t, 2, original.Production.Credits)
	assert.Equal(t, 5, copy.Production.Credits)
	assert.Equal(t, "played1", original.PlayedCards[0])
	assert.Equal(t, "modified_played", copy.PlayedCards[0])
}

func TestPlayer_ResourcesAndProduction(t *testing.T) {
	player := &model.Player{
		Resources: model.Resources{
			Credits:  40,
			Steel:    8,
			Titanium: 3,
			Plants:   12,
			Energy:   6,
			Heat:     15,
		},
		Production: model.Production{
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