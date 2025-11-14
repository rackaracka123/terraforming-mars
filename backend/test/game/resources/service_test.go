package resources_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (resources.Service, player.Repository, string, string) {
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)

	// Create resources repository and service
	resourcesRepo := resources.NewRepository(playerRepo)
	resourcesService := resources.NewService(resourcesRepo)

	// Create test data
	ctx := context.Background()
	gameID := "test-game-id"
	playerID := "test-player-id"

	// Create a test player
	player := model.Player{
		ID:   playerID,
		Name: "TestPlayer",
		Resources: model.Resources{
			Credits:  50,
			Steel:    10,
			Titanium: 5,
			Plants:   20,
			Energy:   15,
			Heat:     8,
		},
		Production: model.Production{
			Credits:  3,
			Steel:    2,
			Titanium: 1,
			Plants:   4,
			Energy:   3,
			Heat:     2,
		},
	}

	err := playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	return resourcesService, playerRepo, gameID, playerID
}

func TestResourceService_AddResources(t *testing.T) {
	resourcesService, playerRepo, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Add resources
	err := resourcesService.AddResources(ctx, gameID, playerID, resources.ResourceSet{
		Credits: 10,
		Steel:   5,
	})
	assert.NoError(t, err)

	// Verify resources were added
	player, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Equal(t, 60, player.Resources.Credits) // 50 + 10
	assert.Equal(t, 15, player.Resources.Steel)   // 10 + 5
}

func TestResourceService_ValidateResourceCost(t *testing.T) {
	resourcesService, _, gameID, playerID := setupTest(t)
	ctx := context.Background()

	t.Run("Sufficient resources", func(t *testing.T) {
		err := resourcesService.ValidateResourceCost(ctx, gameID, playerID, resources.ResourceSet{
			Credits: 30,
			Steel:   5,
		})
		assert.NoError(t, err)
	})

	t.Run("Insufficient resources", func(t *testing.T) {
		err := resourcesService.ValidateResourceCost(ctx, gameID, playerID, resources.ResourceSet{
			Credits: 100, // Player only has 50
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient resources")
	})
}

func TestResourceService_PayResourceCost(t *testing.T) {
	resourcesService, playerRepo, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Pay cost
	err := resourcesService.PayResourceCost(ctx, gameID, playerID, resources.ResourceSet{
		Credits: 20,
		Steel:   3,
	})
	assert.NoError(t, err)

	// Verify resources were deducted
	player, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Equal(t, 30, player.Resources.Credits) // 50 - 20
	assert.Equal(t, 7, player.Resources.Steel)    // 10 - 3
}

func TestResourceService_AddProduction(t *testing.T) {
	resourcesService, playerRepo, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Add production
	err := resourcesService.AddProduction(ctx, gameID, playerID, resources.ResourceSet{
		Credits: 2,
		Energy:  1,
	})
	assert.NoError(t, err)

	// Verify production was added
	player, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Equal(t, 5, player.Production.Credits) // 3 + 2
	assert.Equal(t, 4, player.Production.Energy)  // 3 + 1
}

func TestResourceService_ValidateProductionRequirement(t *testing.T) {
	resourcesService, _, gameID, playerID := setupTest(t)
	ctx := context.Background()

	t.Run("Sufficient production", func(t *testing.T) {
		err := resourcesService.ValidateProductionRequirement(ctx, gameID, playerID, resources.ResourceSet{
			Credits: 2,
			Steel:   1,
		})
		assert.NoError(t, err)
	})

	t.Run("Insufficient production", func(t *testing.T) {
		err := resourcesService.ValidateProductionRequirement(ctx, gameID, playerID, resources.ResourceSet{
			Credits: 10, // Player only has 3
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient production")
	})
}

func TestResourceService_CanAffordCost(t *testing.T) {
	resourcesService, _, gameID, playerID := setupTest(t)
	ctx := context.Background()

	t.Run("Can afford", func(t *testing.T) {
		canAfford, err := resourcesService.CanAffordCost(ctx, gameID, playerID, 30)
		assert.NoError(t, err)
		assert.True(t, canAfford)
	})

	t.Run("Cannot afford", func(t *testing.T) {
		canAfford, err := resourcesService.CanAffordCost(ctx, gameID, playerID, 100)
		assert.NoError(t, err)
		assert.False(t, canAfford)
	})
}

func TestResourceService_GetPlayerResources(t *testing.T) {
	resourcesService, _, gameID, playerID := setupTest(t)
	ctx := context.Background()

	resources, err := resourcesService.GetPlayerResources(ctx, gameID, playerID)
	assert.NoError(t, err)
	assert.Equal(t, 50, resources.Credits)
	assert.Equal(t, 10, resources.Steel)
	assert.Equal(t, 5, resources.Titanium)
}

func TestResourceService_GetPlayerProduction(t *testing.T) {
	resourcesService, _, gameID, playerID := setupTest(t)
	ctx := context.Background()

	production, err := resourcesService.GetPlayerProduction(ctx, gameID, playerID)
	assert.NoError(t, err)
	assert.Equal(t, 3, production.Credits)
	assert.Equal(t, 2, production.Steel)
	assert.Equal(t, 1, production.Titanium)
}
