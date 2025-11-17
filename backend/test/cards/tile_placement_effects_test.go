package cards_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCardTilePlacementEffects tests that cards with tile placement effects correctly create tile queues
func TestCardTilePlacementEffects(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Setup repositories
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardDeckRepo := repository.NewCardDeckRepository()
	cardProcessor := cards.NewCardProcessor(gameRepo, playerRepo, cardDeckRepo)

	// Create a game
	gameSettings := model.GameSettings{
		MaxPlayers: 2,
	}
	game, err := gameRepo.Create(ctx, gameSettings)
	require.NoError(t, err)

	// Create a player
	player := model.Player{
		ID:   "test-player",
		Name: "Test Player",
		Resources: model.Resources{
			Credits: 100,
		},
		TerraformRating: 20,
	}
	err = playerRepo.Create(ctx, game.ID, player)
	require.NoError(t, err)

	t.Run("Single City Placement Effect", func(t *testing.T) {
		// Create a card with a single city placement effect
		cardWithCityPlacement := &model.Card{
			ID:   "test-city-card",
			Name: "Test City Card",
			Behaviors: []model.CardBehavior{
				{
					Triggers: []model.Trigger{
						{Type: model.ResourceTriggerAuto},
					},
					Outputs: []model.ResourceCondition{
						{
							Type:   model.ResourceCityPlacement,
							Amount: 1,
							Target: model.TargetNone,
						},
					},
				},
			},
		}

		// Apply card effects
		err := cardProcessor.ApplyCardEffects(ctx, game.ID, player.ID, cardWithCityPlacement, nil, nil)
		require.NoError(t, err)

		// Verify that a tile queue was created
		queue, err := playerRepo.GetPendingTileSelectionQueue(ctx, game.ID, player.ID)
		require.NoError(t, err)
		require.NotNil(t, queue, "Should have a tile queue")
		assert.Equal(t, 1, len(queue.Items), "Queue should have 1 item")
		assert.Equal(t, "city", queue.Items[0])
		assert.Equal(t, cardWithCityPlacement.ID, queue.Source)

		// Clean up for next test
		err = playerRepo.ClearPendingTileSelection(ctx, game.ID, player.ID)
		require.NoError(t, err)
	})

	t.Run("Multiple Tile Placement Effects", func(t *testing.T) {
		// Create a card with multiple tile placement effects
		cardWithMultipleTiles := &model.Card{
			ID:   "test-multi-tile-card",
			Name: "Test Multi-Tile Card",
			Behaviors: []model.CardBehavior{
				{
					Triggers: []model.Trigger{
						{Type: model.ResourceTriggerAuto},
					},
					Outputs: []model.ResourceCondition{
						{
							Type:   model.ResourceOceanPlacement,
							Amount: 2,
							Target: model.TargetNone,
						},
						{
							Type:   model.ResourceCityPlacement,
							Amount: 1,
							Target: model.TargetNone,
						},
					},
				},
			},
		}

		// Apply card effects
		err := cardProcessor.ApplyCardEffects(ctx, game.ID, player.ID, cardWithMultipleTiles, nil, nil)
		require.NoError(t, err)

		// Verify the queue was created with all 3 tiles
		queue, err := playerRepo.GetPendingTileSelectionQueue(ctx, game.ID, player.ID)
		require.NoError(t, err)
		require.NotNil(t, queue, "Should have a tile queue")
		assert.Equal(t, 3, len(queue.Items), "Queue should have 3 items")
		assert.Equal(t, []string{"ocean", "ocean", "city"}, queue.Items)
		assert.Equal(t, cardWithMultipleTiles.ID, queue.Source)

		// Clean up for next test
		err = playerRepo.ClearPendingTileSelection(ctx, game.ID, player.ID)
		require.NoError(t, err)
		err = playerRepo.ClearPendingTileSelectionQueue(ctx, game.ID, player.ID)
		require.NoError(t, err)
	})

	t.Run("Mixed Auto and Manual Triggers", func(t *testing.T) {
		// Create a card with both auto and manual triggers - only auto should create tiles
		cardWithMixedTriggers := &model.Card{
			ID:   "test-mixed-card",
			Name: "Test Mixed Triggers Card",
			Behaviors: []model.CardBehavior{
				{
					Triggers: []model.Trigger{
						{Type: model.ResourceTriggerAuto},
					},
					Outputs: []model.ResourceCondition{
						{
							Type:   model.ResourceGreeneryPlacement,
							Amount: 1,
							Target: model.TargetNone,
						},
					},
				},
				{
					Triggers: []model.Trigger{
						{Type: model.ResourceTriggerManual},
					},
					Outputs: []model.ResourceCondition{
						{
							Type:   model.ResourceCityPlacement,
							Amount: 1,
							Target: model.TargetNone,
						},
					},
				},
			},
		}

		// Apply card effects
		err := cardProcessor.ApplyCardEffects(ctx, game.ID, player.ID, cardWithMixedTriggers, nil, nil)
		require.NoError(t, err)

		// Verify that only the auto-triggered tile placement created a queue
		queue, err := playerRepo.GetPendingTileSelectionQueue(ctx, game.ID, player.ID)
		require.NoError(t, err)
		require.NotNil(t, queue, "Should have a tile queue from auto trigger")
		assert.Equal(t, 1, len(queue.Items), "Queue should have 1 item")
		assert.Equal(t, "greenery", queue.Items[0])
		assert.Equal(t, cardWithMixedTriggers.ID, queue.Source)

		// Verify the manual action was also added to the player
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, player.ID)
		require.NoError(t, err)
		assert.Len(t, updatedPlayer.Actions, 1, "Should have one manual action")
		if len(updatedPlayer.Actions) > 0 {
			assert.Equal(t, cardWithMixedTriggers.ID, updatedPlayer.Actions[0].CardID)
			assert.Equal(t, 1, updatedPlayer.Actions[0].BehaviorIndex)
		}

		// Clean up for next test
		err = playerRepo.ClearPendingTileSelection(ctx, game.ID, player.ID)
		require.NoError(t, err)
		err = playerRepo.ClearPendingTileSelectionQueue(ctx, game.ID, player.ID)
		require.NoError(t, err)
		err = playerRepo.UpdatePlayerActions(ctx, game.ID, player.ID, []model.PlayerAction{})
		require.NoError(t, err)
	})

	t.Run("No Tile Placement Effects", func(t *testing.T) {
		// Create a card with no tile placement effects
		cardWithNoTiles := &model.Card{
			ID:   "test-no-tiles-card",
			Name: "Test No Tiles Card",
			Behaviors: []model.CardBehavior{
				{
					Triggers: []model.Trigger{
						{Type: model.ResourceTriggerAuto},
					},
					Outputs: []model.ResourceCondition{
						{
							Type:   model.ResourceCredits,
							Amount: 5,
							Target: model.TargetSelfPlayer,
						},
					},
				},
			},
		}

		// Apply card effects
		err := cardProcessor.ApplyCardEffects(ctx, game.ID, player.ID, cardWithNoTiles, nil, nil)
		require.NoError(t, err)

		// Verify no queue was created (no tile placement effects)
		queue, err := playerRepo.GetPendingTileSelectionQueue(ctx, game.ID, player.ID)
		require.NoError(t, err)
		assert.Nil(t, queue, "Should not have a tile queue")

		// Verify the credits were still applied
		updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, player.ID)
		require.NoError(t, err)
		assert.Equal(t, 105, updatedPlayer.Resources.Credits, "Credits should have increased by 5")
	})
}
