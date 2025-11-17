package fixtures

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/model"

	"github.com/stretchr/testify/require"
)

// PlayerFixture holds player configuration for testing
type PlayerFixture struct {
	Player model.Player
}

// NewDefaultPlayer creates a player with standard starting resources
func NewDefaultPlayer() model.Player {
	return model.Player{
		ID:              "player1",
		Name:            "Test Player",
		Resources:       model.Resources{Credits: 40},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
	}
}

// NewPlayerWithResources creates a player with custom resources
func NewPlayerWithResources(resources model.Resources) model.Player {
	player := NewDefaultPlayer()
	player.Resources = resources
	return player
}

// NewPlayerWithProduction creates a player with custom production
func NewPlayerWithProduction(production model.Production) model.Player {
	player := NewDefaultPlayer()
	player.Production = production
	return player
}

// NewPlayerWithCorporation creates a player with a specific corporation
func NewPlayerWithCorporation(corporation *model.Card) model.Player {
	player := NewDefaultPlayer()
	player.Corporation = corporation
	return player
}

// NewPlayerWithResourcesAndProduction creates a player with custom resources and production
func NewPlayerWithResourcesAndProduction(resources model.Resources, production model.Production) model.Player {
	player := NewDefaultPlayer()
	player.Resources = resources
	player.Production = production
	return player
}

// AddPlayerToGame creates a player and adds them to an existing game
func AddPlayerToGame(t *testing.T, container *ServiceContainer, gameID string, player model.Player) string {
	ctx := context.Background()
	err := container.PlayerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)
	return player.ID
}

// NewPlayerInGame creates a default player and adds them to a game
func NewPlayerInGame(t *testing.T, container *ServiceContainer, gameID string, playerID string) model.Player {
	ctx := context.Background()
	player := model.Player{
		ID:              playerID,
		Name:            "Test Player " + playerID,
		Resources:       model.Resources{Credits: 40},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
	}
	err := container.PlayerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)
	return player
}

// NewPlayerWithCards creates a player with played cards
func NewPlayerWithCards(playedCards []string) model.Player {
	player := NewDefaultPlayer()
	player.PlayedCards = playedCards
	return player
}
