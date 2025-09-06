package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayer_CanAffordStandardProject(t *testing.T) {
	player := &Player{
		Resources: Resources{Credits: 20},
	}

	tests := []struct {
		name     string
		project  StandardProject
		expected bool
	}{
		{"Can afford sell patents", StandardProjectSellPatents, true}, // 0 cost
		{"Can afford power plant", StandardProjectPowerPlant, true},   // 11 cost
		{"Can afford asteroid", StandardProjectAsteroid, true},        // 14 cost
		{"Can afford aquifer", StandardProjectAquifer, true},          // 18 cost
		{"Cannot afford greenery", StandardProjectGreenery, false},    // 23 cost
		{"Cannot afford city", StandardProjectCity, false},            // 25 cost
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := player.CanAffordStandardProject(tt.project)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlayer_CanAffordStandardProject_InvalidProject(t *testing.T) {
	player := &Player{
		Resources: Resources{Credits: 100},
	}

	// Test with invalid project (not in cost map)
	result := player.CanAffordStandardProject("INVALID_PROJECT")
	assert.False(t, result)
}

func TestPlayer_HasCardsToSell(t *testing.T) {
	player := &Player{
		Cards: []string{"card1", "card2", "card3"},
	}

	tests := []struct {
		name     string
		count    int
		expected bool
	}{
		{"Can sell 1 card", 1, true},
		{"Can sell 3 cards", 3, true},
		{"Cannot sell 4 cards", 4, false},
		{"Cannot sell 0 cards", 0, false},
		{"Cannot sell negative cards", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := player.HasCardsToSell(tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlayer_HasCardsToSell_EmptyHand(t *testing.T) {
	player := &Player{
		Cards: []string{},
	}

	assert.False(t, player.HasCardsToSell(1))
	assert.False(t, player.HasCardsToSell(0))
}

func TestPlayer_GetMaxCardsToSell(t *testing.T) {
	tests := []struct {
		name     string
		cards    []string
		expected int
	}{
		{"No cards", []string{}, 0},
		{"One card", []string{"card1"}, 1},
		{"Three cards", []string{"card1", "card2", "card3"}, 3},
		{"Many cards", []string{"c1", "c2", "c3", "c4", "c5", "c6", "c7"}, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &Player{Cards: tt.cards}
			result := player.GetMaxCardsToSell()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlayer_InitialState(t *testing.T) {
	player := &Player{
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
