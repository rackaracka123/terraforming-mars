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

// TestEcolineStandardProjectDiscount tests that Ecoline's plant discount effect
// automatically recalculates when card hand or effects change
func TestEcolineStandardProjectDiscount(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Create effect subscriber
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)

	// Create game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 2})
	require.NoError(t, err)
	gameID := game.ID

	// Create player
	playerID := "player-1"
	player := model.Player{
		ID:      playerID,
		Name:    "Test Player",
		Cards:   []string{},
		Effects: []model.PlayerEffect{},
	}
	require.NoError(t, playerRepo.Create(ctx, gameID, player))

	// Create Ecoline corporation card with plant discount behavior
	ecolineCard := &model.Card{
		ID:   "ecoline",
		Name: "Ecoline",
		Type: model.CardTypeCorporation,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerAuto,
						Condition: &model.ResourceTriggerCondition{
							Type: model.TriggerPlayerEffectsChanged,
						},
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:                     model.ResourceDiscount,
						Amount:                   1,
						AffectedResources:        []string{string(model.ResourcePlants)},
						AffectedStandardProjects: []model.StandardProject{model.StandardProjectConvertPlantsToGreenery},
					},
				},
			},
		},
	}

	// Subscribe card effects (simulates playing Ecoline corporation)
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, ecolineCard.ID, ecolineCard)
	require.NoError(t, err)

	// Verify that RequirementModifiers were calculated
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)

	// Should have 1 modifier for convert-plants-to-greenery standard project
	require.Len(t, updatedPlayer.RequirementModifiers, 1)
	modifier := updatedPlayer.RequirementModifiers[0]

	assert.Equal(t, 1, modifier.Amount, "Ecoline should provide 1 plant discount")
	assert.Contains(t, modifier.AffectedResources, model.ResourcePlants)
	assert.Nil(t, modifier.CardTarget, "Should not target specific card")
	require.NotNil(t, modifier.StandardProjectTarget, "Should target standard project")
	assert.Equal(t, model.StandardProjectConvertPlantsToGreenery, *modifier.StandardProjectTarget)

	t.Log("✅ Ecoline standard project discount test passed")
}

// TestInventrixGlobalParameterLenience tests that Inventrix's +2 global parameter lenience
// is correctly calculated as a requirement modifier
func TestInventrixGlobalParameterLenience(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)

	// Create game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 2})
	require.NoError(t, err)
	gameID := game.ID

	// Create player
	playerID := "player-1"
	player := model.Player{
		ID:      playerID,
		Name:    "Test Player",
		Cards:   []string{},
		Effects: []model.PlayerEffect{},
	}
	require.NoError(t, playerRepo.Create(ctx, gameID, player))

	// Create Inventrix corporation card with global parameter lenience
	inventrixCard := &model.Card{
		ID:   "inventrix",
		Name: "Inventrix",
		Type: model.CardTypeCorporation,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerAuto,
						Condition: &model.ResourceTriggerCondition{
							Type: model.TriggerPlayerEffectsChanged,
						},
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:              model.ResourceGlobalParameterLenience,
						Amount:            2,
						AffectedResources: []string{string(model.ResourceGlobalParameter)},
					},
				},
			},
		},
	}

	// Subscribe card effects
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, inventrixCard.ID, inventrixCard)
	require.NoError(t, err)

	// Verify RequirementModifiers
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)

	require.Len(t, updatedPlayer.RequirementModifiers, 1)
	modifier := updatedPlayer.RequirementModifiers[0]

	assert.Equal(t, 2, modifier.Amount, "Inventrix should provide +2 lenience")
	assert.Contains(t, modifier.AffectedResources, model.ResourceGlobalParameter)
	assert.Nil(t, modifier.CardTarget, "Should be global modifier")
	assert.Nil(t, modifier.StandardProjectTarget, "Should not target standard project")

	t.Log("✅ Inventrix global parameter lenience test passed")
}

// TestModifierMerging tests that multiple effects targeting the same card/project
// are correctly merged (amounts summed)
func TestModifierMerging(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)

	// Create game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 2})
	require.NoError(t, err)
	gameID := game.ID

	// Create player
	playerID := "player-1"
	player := model.Player{
		ID:      playerID,
		Name:    "Test Player",
		Cards:   []string{},
		Effects: []model.PlayerEffect{},
	}
	require.NoError(t, playerRepo.Create(ctx, gameID, player))

	// Create two cards that both provide discounts for convert-plants-to-greenery
	discountCard1 := &model.Card{
		ID:   "discount-card-1",
		Name: "Discount Card 1",
		Type: model.CardTypeActive,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerAuto,
						Condition: &model.ResourceTriggerCondition{
							Type: model.TriggerPlayerEffectsChanged,
						},
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:                     model.ResourceDiscount,
						Amount:                   1,
						AffectedResources:        []string{string(model.ResourcePlants)},
						AffectedStandardProjects: []model.StandardProject{model.StandardProjectConvertPlantsToGreenery},
					},
				},
			},
		},
	}

	discountCard2 := &model.Card{
		ID:   "discount-card-2",
		Name: "Discount Card 2",
		Type: model.CardTypeActive,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerAuto,
						Condition: &model.ResourceTriggerCondition{
							Type: model.TriggerPlayerEffectsChanged,
						},
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:                     model.ResourceDiscount,
						Amount:                   2,
						AffectedResources:        []string{string(model.ResourcePlants)},
						AffectedStandardProjects: []model.StandardProject{model.StandardProjectConvertPlantsToGreenery},
					},
				},
			},
		},
	}

	// Subscribe first card
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, discountCard1.ID, discountCard1)
	require.NoError(t, err)

	// Verify 1 modifier with amount 1
	player1, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player1.RequirementModifiers, 1)
	assert.Equal(t, 1, player1.RequirementModifiers[0].Amount)

	// Subscribe second card
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, discountCard2.ID, discountCard2)
	require.NoError(t, err)

	// Verify modifiers were merged into 1 with combined amount (1 + 2 = 3)
	player2, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player2.RequirementModifiers, 1, "Should merge into single modifier")

	modifier := player2.RequirementModifiers[0]
	assert.Equal(t, 3, modifier.Amount, "Should sum amounts: 1 + 2 = 3")
	assert.Contains(t, modifier.AffectedResources, model.ResourcePlants)
	require.NotNil(t, modifier.StandardProjectTarget)
	assert.Equal(t, model.StandardProjectConvertPlantsToGreenery, *modifier.StandardProjectTarget)

	t.Log("✅ Modifier merging test passed")
}
