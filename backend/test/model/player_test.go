package model_test

import (
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/types"

	"github.com/stretchr/testify/assert"
)

func TestPlayer_InitialState(t *testing.T) {
	eventBus := events.NewEventBus()
	p := player.NewPlayer(eventBus, "game1", "player1", "Test Player")

	// Test default values using getters
	assert.Equal(t, "player1", p.ID())
	assert.Equal(t, "Test Player", p.Name())
	assert.False(t, p.Corp().HasCorporation())
	assert.Empty(t, p.Hand().Cards())
	assert.Empty(t, p.Hand().PlayedCards())
	assert.Equal(t, 20, p.Resources().TerraformRating()) // Default TR is 20
	assert.False(t, p.Turn().IsConnected())              // Default is disconnected until explicitly set

	// Test zero-value resources and production
	resources := p.Resources().Get()
	assert.Equal(t, 0, resources.Credits)

	production := p.Resources().Production()
	assert.Equal(t, 0, production.Credits)
}

func TestPlayer_ResourcesAndProduction(t *testing.T) {
	eventBus := events.NewEventBus()
	p := player.NewPlayer(eventBus, "game1", "player1", "Test Player")

	// Set resources using setter
	p.Resources().Set(types.Resources{
		Credits:  40,
		Steel:    8,
		Titanium: 3,
		Plants:   12,
		Energy:   6,
		Heat:     15,
	})

	// Set production using setter
	p.Resources().SetProduction(types.Production{
		Credits:  1,
		Steel:    2,
		Titanium: 1,
		Plants:   0,
		Energy:   3,
		Heat:     0,
	})

	// Test resources using getter
	resources := p.Resources().Get()
	assert.Equal(t, 40, resources.Credits)
	assert.Equal(t, 8, resources.Steel)
	assert.Equal(t, 3, resources.Titanium)
	assert.Equal(t, 12, resources.Plants)
	assert.Equal(t, 6, resources.Energy)
	assert.Equal(t, 15, resources.Heat)

	// Test production using getter
	production := p.Resources().Production()
	assert.Equal(t, 1, production.Credits)
	assert.Equal(t, 2, production.Steel)
	assert.Equal(t, 1, production.Titanium)
	assert.Equal(t, 0, production.Plants)
	assert.Equal(t, 3, production.Energy)
	assert.Equal(t, 0, production.Heat)
}

func TestPlayer_SetTerraformRating(t *testing.T) {
	eventBus := events.NewEventBus()
	p := player.NewPlayer(eventBus, "game1", "player1", "Test Player")

	// Initial TR should be 20
	assert.Equal(t, 20, p.Resources().TerraformRating())

	// Set new TR
	p.Resources().SetTerraformRating(25)
	assert.Equal(t, 25, p.Resources().TerraformRating())
}

func TestPlayer_SetConnectionStatus(t *testing.T) {
	eventBus := events.NewEventBus()
	p := player.NewPlayer(eventBus, "game1", "player1", "Test Player")

	// Initially disconnected (default)
	assert.False(t, p.Turn().IsConnected())

	// Connect
	p.Turn().SetConnectionStatus(true)
	assert.True(t, p.Turn().IsConnected())

	// Disconnect
	p.Turn().SetConnectionStatus(false)
	assert.False(t, p.Turn().IsConnected())
}
