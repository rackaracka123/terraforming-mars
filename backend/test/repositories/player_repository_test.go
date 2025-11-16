package repositories_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/test/fixtures"
)

func TestPlayerRepository_Create_And_GetByID(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(
		fixtures.WithID("player-1"),
		fixtures.WithName("Test Player"),
		fixtures.WithTR(20),
		fixtures.WithCredits(50),
	)

	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Retrieve player
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Equal(t, "player-1", retrieved.ID)
	assert.Equal(t, "Test Player", retrieved.Name)
	assert.Equal(t, 20, retrieved.TerraformRating)
	assert.Equal(t, 50, retrieved.Resources.Credits)
}

func TestPlayerRepository_GetByID_NotFound(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	_, err := repo.GetByID(context.Background(), "game-1", "nonexistent-player")
	assert.Error(t, err)
	assert.Error(t, err)
}

func TestPlayerRepository_Delete(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(fixtures.WithID("player-1"))
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Delete player
	err = repo.Delete(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)

	// Verify player is deleted
	_, err = repo.GetByID(context.Background(), "game-1", "player-1")
	assert.Error(t, err)
}

func TestPlayerRepository_ListByGameID(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	// Create multiple players
	p1 := fixtures.NewTestPlayer(fixtures.WithID("player-1"), fixtures.WithName("Player 1"))
	p2 := fixtures.NewTestPlayer(fixtures.WithID("player-2"), fixtures.WithName("Player 2"))
	p3 := fixtures.NewTestPlayer(fixtures.WithID("player-3"), fixtures.WithName("Player 3"))

	err := repo.Create(context.Background(), "game-1", *p1)
	require.NoError(t, err)
	err = repo.Create(context.Background(), "game-1", *p2)
	require.NoError(t, err)
	err = repo.Create(context.Background(), "game-1", *p3)
	require.NoError(t, err)

	// List all players
	players, err := repo.ListByGameID(context.Background(), "game-1")
	assert.NoError(t, err)
	assert.Len(t, players, 3)

	// Verify all players are in the list
	ids := make(map[string]bool)
	for _, p := range players {
		ids[p.ID] = true
	}
	assert.True(t, ids["player-1"])
	assert.True(t, ids["player-2"])
	assert.True(t, ids["player-3"])
}

func TestPlayerRepository_UpdateResources(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(
		fixtures.WithID("player-1"),
		fixtures.WithResources(10, 5, 3, 8, 4, 2),
	)
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Update resources
	newResources := player.Resources{
		Credits:  25,
		Steel:    10,
		Titanium: 8,
		Plants:   12,
		Energy:   6,
		Heat:     15,
	}
	err = repo.UpdateResources(context.Background(), "game-1", "player-1", newResources)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Equal(t, newResources, retrieved.Resources)
}

func TestPlayerRepository_UpdateProduction(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(fixtures.WithID("player-1"))
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Update production
	newProduction := player.Production{
		Credits:  5,
		Steel:    2,
		Titanium: 1,
		Plants:   3,
		Energy:   4,
		Heat:     2,
	}
	err = repo.UpdateProduction(context.Background(), "game-1", "player-1", newProduction)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Equal(t, newProduction, retrieved.Production)
}

func TestPlayerRepository_UpdateTerraformRating(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(
		fixtures.WithID("player-1"),
		fixtures.WithTR(20),
	)
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Update TR
	err = repo.UpdateTerraformRating(context.Background(), "game-1", "player-1", 25)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Equal(t, 25, retrieved.TerraformRating)
}

func TestPlayerRepository_UpdateVictoryPoints(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(
		fixtures.WithID("player-1"),
		fixtures.WithVP(0),
	)
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Update victory points
	err = repo.UpdateVictoryPoints(context.Background(), "game-1", "player-1", 15)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Equal(t, 15, retrieved.VictoryPoints)
}

func TestPlayerRepository_AddCard(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(fixtures.WithID("player-1"))
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Add a card
	card := player.Card{
		ID:   "card-123",
		Name: "Test Card",
		Type: player.CardType("project"),
	}
	err = repo.AddCard(context.Background(), "game-1", "player-1", card)
	assert.NoError(t, err)

	// Verify card was added
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Len(t, retrieved.Cards, 1)
	assert.Equal(t, "card-123", retrieved.Cards[0].ID)
	assert.Equal(t, "Test Card", retrieved.Cards[0].Name)
}

func TestPlayerRepository_AddCards(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(fixtures.WithID("player-1"))
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Add multiple cards
	cards := []player.Card{
		{ID: "card-1", Name: "Card 1", Type: player.CardType("project")},
		{ID: "card-2", Name: "Card 2", Type: player.CardType("project")},
		{ID: "card-3", Name: "Card 3", Type: player.CardType("project")},
	}
	err = repo.AddCards(context.Background(), "game-1", "player-1", cards)
	assert.NoError(t, err)

	// Verify cards were added
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Len(t, retrieved.Cards, 3)
}

func TestPlayerRepository_AddPlayedCard(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(fixtures.WithID("player-1"))
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Add a played card
	card := player.Card{
		ID:   "card-456",
		Name: "Played Card",
		Type: player.CardType("project"),
	}
	err = repo.AddPlayedCard(context.Background(), "game-1", "player-1", card)
	assert.NoError(t, err)

	// Verify played card was added
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Len(t, retrieved.PlayedCards, 1)
	assert.Equal(t, "card-456", retrieved.PlayedCards[0].ID)
}

func TestPlayerRepository_DeductResources(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(
		fixtures.WithID("player-1"),
		fixtures.WithResources(50, 10, 8, 15, 5, 20),
	)
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Deduct resources
	cost := domain.ResourceSet{
		Credits:  10,
		Steel:    3,
		Titanium: 2,
	}
	err = repo.DeductResources(context.Background(), "game-1", "player-1", cost)
	assert.NoError(t, err)

	// Verify deduction
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Equal(t, 40, retrieved.Resources.Credits)
	assert.Equal(t, 7, retrieved.Resources.Steel)
	assert.Equal(t, 6, retrieved.Resources.Titanium)
	assert.Equal(t, 15, retrieved.Resources.Plants) // Unchanged
}

func TestPlayerRepository_AddResources(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(
		fixtures.WithID("player-1"),
		fixtures.WithResources(10, 5, 3, 8, 4, 2),
	)
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Add resources
	resourcesToAdd := domain.ResourceSet{
		Credits: 20,
		Plants:  5,
		Energy:  3,
	}
	err = repo.AddResources(context.Background(), "game-1", "player-1", resourcesToAdd)
	assert.NoError(t, err)

	// Verify addition
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Equal(t, 30, retrieved.Resources.Credits)
	assert.Equal(t, 13, retrieved.Resources.Plants)
	assert.Equal(t, 7, retrieved.Resources.Energy)
	assert.Equal(t, 5, retrieved.Resources.Steel) // Unchanged
}

func TestPlayerRepository_AddProduction(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(
		fixtures.WithID("player-1"),
		fixtures.WithProduction(5, 2, 1, 3, 4, 2),
	)
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Add production
	productionToAdd := domain.ResourceSet{
		Credits: 3,
		Energy:  2,
		Plants:  1,
	}
	err = repo.AddProduction(context.Background(), "game-1", "player-1", productionToAdd)
	assert.NoError(t, err)

	// Verify addition
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Equal(t, 8, retrieved.Production.Credits)
	assert.Equal(t, 6, retrieved.Production.Energy)
	assert.Equal(t, 4, retrieved.Production.Plants)
	assert.Equal(t, 2, retrieved.Production.Steel) // Unchanged
}

func TestPlayerRepository_CanAfford(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(
		fixtures.WithID("player-1"),
		fixtures.WithResources(20, 10, 5, 8, 4, 15),
	)
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	tests := []struct {
		name      string
		cost      domain.ResourceSet
		canAfford bool
	}{
		{
			name:      "Can afford",
			cost:      domain.ResourceSet{Credits: 10, Steel: 5},
			canAfford: true,
		},
		{
			name:      "Can afford exact amount",
			cost:      domain.ResourceSet{Credits: 20, Steel: 10, Titanium: 5},
			canAfford: true,
		},
		{
			name:      "Cannot afford - insufficient credits",
			cost:      domain.ResourceSet{Credits: 30},
			canAfford: false,
		},
		{
			name:      "Cannot afford - insufficient steel",
			cost:      domain.ResourceSet{Steel: 15},
			canAfford: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canAfford, err := repo.CanAfford(context.Background(), "game-1", "player-1", tt.cost)
			assert.NoError(t, err)
			assert.Equal(t, tt.canAfford, canAfford)
		})
	}
}

func TestPlayerRepository_UpdateResourceStorage(t *testing.T) {
	eventBus := events.NewEventBus()
	repo := player.NewRepository(eventBus)

	p := fixtures.NewTestPlayer(fixtures.WithID("player-1"))
	err := repo.Create(context.Background(), "game-1", *p)
	require.NoError(t, err)

	// Update resource storage
	storage := map[string]int{
		"animals":  5,
		"microbes": 3,
		"floaters": 2,
	}
	err = repo.UpdateResourceStorage(context.Background(), "game-1", "player-1", storage)
	assert.NoError(t, err)

	// Verify storage
	retrieved, err := repo.GetByID(context.Background(), "game-1", "player-1")
	assert.NoError(t, err)
	assert.Equal(t, 5, retrieved.ResourceStorage["animals"])
	assert.Equal(t, 3, retrieved.ResourceStorage["microbes"])
	assert.Equal(t, 2, retrieved.ResourceStorage["floaters"])
}
