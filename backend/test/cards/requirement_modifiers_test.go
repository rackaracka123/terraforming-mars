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

// TestTagBasedCardDiscounts tests that discounts correctly apply to cards in hand with matching tags
func TestTagBasedCardDiscounts(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Load cards from JSON (needed for tag filtering)
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err, "Failed to load cards from JSON")

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

	// Create discount card for space tags
	discountCard := &model.Card{
		ID:   "space-tag-discount",
		Name: "Space Tag Discount",
		Type: model.CardTypeActive,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerAuto,
						Condition: &model.ResourceTriggerCondition{
							Type: model.TriggerCardHandUpdated,
						},
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:              model.ResourceDiscount,
						Amount:            -2,
						AffectedTags:      []model.CardTag{model.TagSpace},
						AffectedResources: []string{string(model.ResourceCredits)},
					},
				},
			},
		},
	}

	// Subscribe discount effect
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, discountCard.ID, discountCard)
	require.NoError(t, err)

	// No modifiers yet (no cards in hand)
	player1, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Empty(t, player1.RequirementModifiers, "Should have no modifiers with empty hand")

	// Use real card IDs from the loaded JSON that have space tags
	// "009" (Asteroid) and "010" (Comet) are space-tagged event cards
	spaceCardID1 := "009"  // Asteroid - has space tag
	spaceCardID2 := "010"  // Comet - has space tag
	scienceCardID := "087" // Grass - has science tag, not space

	// Add space-tagged card
	err = playerRepo.AddCard(ctx, gameID, playerID, spaceCardID1)
	require.NoError(t, err)

	// Verify modifier created for space card
	player2, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player2.RequirementModifiers, 1)

	modifier := player2.RequirementModifiers[0]
	assert.Equal(t, -2, modifier.Amount, "Should have -2 MC discount")
	require.NotNil(t, modifier.CardTarget, "Should target specific card")
	assert.Equal(t, spaceCardID1, *modifier.CardTarget)

	// Add another space card
	err = playerRepo.AddCard(ctx, gameID, playerID, spaceCardID2)
	require.NoError(t, err)

	// Should now have 2 modifiers (one per space card)
	player3, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player3.RequirementModifiers, 2, "Should have modifiers for both space cards")

	// Add non-space card (science tag)
	err = playerRepo.AddCard(ctx, gameID, playerID, scienceCardID)
	require.NoError(t, err)

	// Should still have 2 modifiers (science card doesn't match)
	player4, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player4.RequirementModifiers, 2, "Should still have 2 modifiers (science card no discount)")

	// Remove one space card
	err = playerRepo.RemoveCard(ctx, gameID, playerID, spaceCardID1)
	require.NoError(t, err)

	// Should now have 1 modifier
	player5, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player5.RequirementModifiers, 1, "Should have 1 modifier after removing space card")

	t.Log("✅ Tag-based card discounts test passed")
}

// TestCardTypeFiltering tests that discounts correctly filter by card type
func TestCardTypeFiltering(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Load cards from JSON (needed for type filtering)
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err, "Failed to load cards from JSON")

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

	// Create discount card that only applies to event cards
	eventDiscountCard := &model.Card{
		ID:   "event-discount-card",
		Name: "Event Discount Card",
		Type: model.CardTypeActive,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerAuto,
						Condition: &model.ResourceTriggerCondition{
							Type: model.TriggerCardHandUpdated,
						},
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:              model.ResourceDiscount,
						Amount:            -5,
						AffectedCardTypes: []model.CardType{model.CardTypeEvent},
						AffectedResources: []string{string(model.ResourceCredits)},
					},
				},
			},
		},
	}

	// Subscribe discount effect
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, eventDiscountCard.ID, eventDiscountCard)
	require.NoError(t, err)

	// Use real card IDs from the loaded JSON with different types
	eventCardID := "009"     // Asteroid - event type
	automatedCardID := "003" // Deep Well Heating - automated type
	activeCardID := "014"    // Development Center - active type

	// Add event card - should get discount
	err = playerRepo.AddCard(ctx, gameID, playerID, eventCardID)
	require.NoError(t, err)

	player1, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player1.RequirementModifiers, 1, "Event card should get discount")
	assert.Equal(t, -5, player1.RequirementModifiers[0].Amount)
	require.NotNil(t, player1.RequirementModifiers[0].CardTarget)
	assert.Equal(t, eventCardID, *player1.RequirementModifiers[0].CardTarget)

	// Add automated card - should not get discount
	err = playerRepo.AddCard(ctx, gameID, playerID, automatedCardID)
	require.NoError(t, err)

	player2, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player2.RequirementModifiers, 1, "Only event card should have discount")

	// Add active card - should not get discount
	err = playerRepo.AddCard(ctx, gameID, playerID, activeCardID)
	require.NoError(t, err)

	player3, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player3.RequirementModifiers, 1, "Only event card should have discount")

	t.Log("✅ Card type filtering test passed")
}

// TestCardHandUpdateTriggersRecalc tests that adding/removing cards triggers recalculation via CardHandUpdatedEvent
func TestCardHandUpdateTriggersRecalc(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Load cards from JSON (needed for tag filtering)
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err, "Failed to load cards from JSON")

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

	// Create discount card
	discountCard := &model.Card{
		ID:   "building-discount",
		Name: "Building Discount",
		Type: model.CardTypeActive,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerAuto,
						Condition: &model.ResourceTriggerCondition{
							Type: model.TriggerCardHandUpdated,
						},
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:              model.ResourceDiscount,
						Amount:            -3,
						AffectedTags:      []model.CardTag{model.TagBuilding},
						AffectedResources: []string{string(model.ResourceCredits)},
					},
				},
			},
		},
	}

	// Subscribe discount effect
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, discountCard.ID, discountCard)
	require.NoError(t, err)

	// Use real card ID from the loaded JSON with building tag
	buildingCardID := "003" // Deep Well Heating - has building tag

	// Add card via PlayerRepository - should trigger CardHandUpdatedEvent
	err = playerRepo.AddCard(ctx, gameID, playerID, buildingCardID)
	require.NoError(t, err)

	// Verify recalculation happened
	player1, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player1.RequirementModifiers, 1, "Modifier should be created after card added")
	assert.Equal(t, -3, player1.RequirementModifiers[0].Amount)

	// Remove card via PlayerRepository - should trigger CardHandUpdatedEvent
	err = playerRepo.RemoveCard(ctx, gameID, playerID, buildingCardID)
	require.NoError(t, err)

	// Verify recalculation happened (modifier removed)
	player2, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Empty(t, player2.RequirementModifiers, "Modifier should be removed after card removed")

	t.Log("✅ Card hand update triggers recalculation test passed")
}

// TestImmediateEffectsNoModifiers tests that immediate effects don't create requirement modifiers
func TestImmediateEffectsNoModifiers(t *testing.T) {
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
		Effects: []model.PlayerEffect{},
	}
	require.NoError(t, playerRepo.Create(ctx, gameID, player))

	// Create card with immediate resource effect (not a passive effect like discount)
	immediateCard := &model.Card{
		ID:   "immediate-resource-card",
		Name: "Immediate Resource Card",
		Type: model.CardTypeAutomated,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{Type: model.ResourceTriggerAuto}, // No condition = immediate effect
				},
				Outputs: []model.ResourceCondition{
					{
						Type:   model.ResourceCredits,
						Amount: 15,
						Target: model.TargetSelfPlayer,
					},
				},
			},
		},
	}

	// Subscribe effect
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, immediateCard.ID, immediateCard)
	require.NoError(t, err)

	// Verify no requirement modifiers (immediate effect, not passive)
	player1, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Empty(t, player1.RequirementModifiers, "Immediate effects should not create modifiers")

	// Verify effect not added to player effects list
	assert.Empty(t, player1.Effects, "Immediate effects should not be in effects list")

	t.Log("✅ Immediate effects no modifiers test passed")
}

// TestMultipleTagsOnCard tests that cards with multiple tags correctly match discount filters
func TestMultipleTagsOnCard(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Load cards from JSON (needed for tag filtering)
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err, "Failed to load cards from JSON")

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

	// Create discount card for power tags
	powerDiscountCard := &model.Card{
		ID:   "power-discount",
		Name: "Power Discount",
		Type: model.CardTypeActive,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerAuto,
						Condition: &model.ResourceTriggerCondition{
							Type: model.TriggerCardHandUpdated,
						},
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:              model.ResourceDiscount,
						Amount:            -4,
						AffectedTags:      []model.CardTag{model.TagPower},
						AffectedResources: []string{string(model.ResourceCredits)},
					},
				},
			},
		},
	}

	// Subscribe discount effect
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, powerDiscountCard.ID, powerDiscountCard)
	require.NoError(t, err)

	// Use real card ID from JSON with multiple tags including power
	// "117" (Geothermal Power) has power and building tags
	multiTagCardID := "117" // Geothermal Power - has power + building tags

	// Add card to hand
	err = playerRepo.AddCard(ctx, gameID, playerID, multiTagCardID)
	require.NoError(t, err)

	// Verify modifier created (card matches power tag filter)
	player1, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player1.RequirementModifiers, 1, "Card with power tag should get discount")

	modifier := player1.RequirementModifiers[0]
	assert.Equal(t, -4, modifier.Amount)
	require.NotNil(t, modifier.CardTarget)
	assert.Equal(t, multiTagCardID, *modifier.CardTarget)

	// Verify no duplicate modifiers
	modifierCount := 0
	for _, mod := range player1.RequirementModifiers {
		if mod.CardTarget != nil && *mod.CardTarget == multiTagCardID {
			modifierCount++
		}
	}
	assert.Equal(t, 1, modifierCount, "Should have exactly 1 modifier (no duplicates)")

	t.Log("✅ Multiple tags on card test passed")
}

// TestInventrixGlobalParameterLeniencePerCard tests that Inventrix's global parameter lenience
// creates modifiers only for cards with global parameter requirements
func TestInventrixGlobalParameterLeniencePerCard(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Load cards from JSON (needed to test with real cards)
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err, "Failed to load cards from JSON")

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

	// Create Inventrix corporation behavior (global parameter lenience +2)
	inventrixBehavior := &model.Card{
		ID:   "B05",
		Name: "Inventrix",
		Type: model.CardTypeCorporation,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerAuto,
						Condition: &model.ResourceTriggerCondition{
							Type: model.TriggerCardHandUpdated,
						},
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:   model.ResourceGlobalParameterLenience,
						Amount: 2,
						Target: model.TargetSelfPlayer,
					},
				},
			},
		},
	}

	// Subscribe Inventrix effect
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, inventrixBehavior.ID, inventrixBehavior)
	require.NoError(t, err)

	// Step 1: Empty hand - should have no modifiers
	player1, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Empty(t, player1.RequirementModifiers, "Should have no modifiers with empty hand")

	// Step 2: Add Livestock (ID "184") which has oxygen >= 9% requirement
	livestockID := "184"
	err = playerRepo.AddCard(ctx, gameID, playerID, livestockID)
	require.NoError(t, err)

	player2, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player2.RequirementModifiers, 1, "Should have 1 modifier for Livestock")

	// Verify modifier targets Livestock specifically
	modifier := player2.RequirementModifiers[0]
	assert.Equal(t, 2, modifier.Amount, "Should have +2 lenience")
	assert.Equal(t, []model.ResourceType{model.ResourceGlobalParameter}, modifier.AffectedResources)
	require.NotNil(t, modifier.CardTarget, "Should have card target")
	assert.Equal(t, livestockID, *modifier.CardTarget, "Should target Livestock card")

	// Step 3: Add Special Design (ID "206") which has NO global parameter requirements
	specialDesignID := "206"
	err = playerRepo.AddCard(ctx, gameID, playerID, specialDesignID)
	require.NoError(t, err)

	player3, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player3.RequirementModifiers, 1, "Should still have 1 modifier (Special Design has no requirements)")

	// Verify still only targets Livestock
	modifier2 := player3.RequirementModifiers[0]
	require.NotNil(t, modifier2.CardTarget)
	assert.Equal(t, livestockID, *modifier2.CardTarget)

	// Step 4: Add Rad-Chem Factory (ID "205") which also has NO global parameter requirements
	radChemID := "205"
	err = playerRepo.AddCard(ctx, gameID, playerID, radChemID)
	require.NoError(t, err)

	player4, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player4.RequirementModifiers, 1, "Should still have 1 modifier (Rad-Chem has no requirements)")

	// Step 5: Remove Livestock - should have no modifiers again
	err = playerRepo.RemoveCard(ctx, gameID, playerID, livestockID)
	require.NoError(t, err)

	player5, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Empty(t, player5.RequirementModifiers, "Should have no modifiers after removing Livestock")

	// Step 6: Add back Livestock - should recreate modifier
	err = playerRepo.AddCard(ctx, gameID, playerID, livestockID)
	require.NoError(t, err)

	player6, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player6.RequirementModifiers, 1, "Should have 1 modifier again")
	require.NotNil(t, player6.RequirementModifiers[0].CardTarget)
	assert.Equal(t, livestockID, *player6.RequirementModifiers[0].CardTarget)

	t.Log("✅ Inventrix global parameter lenience per-card test passed")
}

// TestShuttlesSpaceTagDiscount tests that Shuttles' discount for space-tagged cards
// creates modifiers only for cards with space tags
func TestShuttlesSpaceTagDiscount(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Load cards from JSON (needed to test with real cards)
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err, "Failed to load cards from JSON")

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

	// Create Shuttles card behavior (discount -2 MC for space-tagged cards)
	shuttlesCard := &model.Card{
		ID:   "166",
		Name: "Shuttles",
		Type: model.CardTypeActive,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerAuto,
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:         model.ResourceDiscount,
						Amount:       -2,
						Target:       model.TargetSelfPlayer,
						AffectedTags: []model.CardTag{model.TagSpace},
					},
				},
			},
		},
	}

	// Subscribe Shuttles effect
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, shuttlesCard.ID, shuttlesCard)
	require.NoError(t, err)

	// Step 1: Empty hand - should have no modifiers
	player1, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Empty(t, player1.RequirementModifiers, "Should have no modifiers with empty hand")

	// Step 2: Add Soletta (ID "203") which has space tag
	solettaID := "203"
	err = playerRepo.AddCard(ctx, gameID, playerID, solettaID)
	require.NoError(t, err)

	player2, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player2.RequirementModifiers, 1, "Should have 1 modifier for Soletta")

	// Verify modifier targets Soletta specifically with -2 MC discount
	modifier := player2.RequirementModifiers[0]
	assert.Equal(t, -2, modifier.Amount, "Should have -2 MC discount")
	assert.Equal(t, []model.ResourceType{model.ResourceCredits}, modifier.AffectedResources)
	require.NotNil(t, modifier.CardTarget, "Should have card target")
	assert.Equal(t, solettaID, *modifier.CardTarget, "Should target Soletta card")

	// Step 3: Add Rad-Chem Factory (ID "205") which has NO space tag (building tag)
	radChemID := "205"
	err = playerRepo.AddCard(ctx, gameID, playerID, radChemID)
	require.NoError(t, err)

	player3, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player3.RequirementModifiers, 1, "Should still have 1 modifier (Rad-Chem has no space tag)")

	// Verify still only targets Soletta
	modifier2 := player3.RequirementModifiers[0]
	require.NotNil(t, modifier2.CardTarget)
	assert.Equal(t, solettaID, *modifier2.CardTarget)

	// Step 4: Add Asteroid (ID "009") which has space tag (event card)
	asteroidID := "009"
	err = playerRepo.AddCard(ctx, gameID, playerID, asteroidID)
	require.NoError(t, err)

	player4, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player4.RequirementModifiers, 2, "Should have 2 modifiers (Soletta + Asteroid)")

	// Verify both space cards have modifiers
	spaceCardTargets := make(map[string]bool)
	for _, mod := range player4.RequirementModifiers {
		require.NotNil(t, mod.CardTarget, "All modifiers should have card targets")
		assert.Equal(t, -2, mod.Amount, "All modifiers should have -2 discount")
		spaceCardTargets[*mod.CardTarget] = true
	}
	assert.True(t, spaceCardTargets[solettaID], "Should have modifier for Soletta")
	assert.True(t, spaceCardTargets[asteroidID], "Should have modifier for Asteroid")

	// Step 5: Remove Soletta - should have 1 modifier for Asteroid
	err = playerRepo.RemoveCard(ctx, gameID, playerID, solettaID)
	require.NoError(t, err)

	player5, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player5.RequirementModifiers, 1, "Should have 1 modifier after removing Soletta")
	require.NotNil(t, player5.RequirementModifiers[0].CardTarget)
	assert.Equal(t, asteroidID, *player5.RequirementModifiers[0].CardTarget)

	// Step 6: Remove all space cards - should have no modifiers
	err = playerRepo.RemoveCard(ctx, gameID, playerID, asteroidID)
	require.NoError(t, err)

	player6, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Empty(t, player6.RequirementModifiers, "Should have no modifiers after removing all space cards")

	t.Log("✅ Shuttles space tag discount test passed")
}

// TestMixedLenienceAndDiscount tests that global parameter lenience and discounts
// work together correctly when both effects are active
func TestMixedLenienceAndDiscount(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Load cards from JSON
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err, "Failed to load cards from JSON")

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

	// Create Inventrix behavior (global parameter lenience +2)
	inventrixCard := &model.Card{
		ID:   "B05",
		Name: "Inventrix",
		Type: model.CardTypeCorporation,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{{Type: model.ResourceTriggerAuto}},
				Outputs: []model.ResourceCondition{
					{
						Type:   model.ResourceGlobalParameterLenience,
						Amount: 2,
						Target: model.TargetSelfPlayer,
					},
				},
			},
		},
	}

	// Create Shuttles behavior (discount -2 MC for space tags)
	shuttlesCard := &model.Card{
		ID:   "166",
		Name: "Shuttles",
		Type: model.CardTypeActive,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{{Type: model.ResourceTriggerAuto}},
				Outputs: []model.ResourceCondition{
					{
						Type:         model.ResourceDiscount,
						Amount:       -2,
						Target:       model.TargetSelfPlayer,
						AffectedTags: []model.CardTag{model.TagSpace},
					},
				},
			},
		},
	}

	// Subscribe both effects
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, inventrixCard.ID, inventrixCard)
	require.NoError(t, err)
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, shuttlesCard.ID, shuttlesCard)
	require.NoError(t, err)

	// Verify empty hand - no modifiers
	player1, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Empty(t, player1.RequirementModifiers, "Should have no modifiers with empty hand")

	// Add Livestock (oxygen requirement, no space tag) - should get lenience only
	livestockID := "184"
	err = playerRepo.AddCard(ctx, gameID, playerID, livestockID)
	require.NoError(t, err)

	player2, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player2.RequirementModifiers, 1, "Should have 1 modifier (lenience for Livestock)")
	assert.Equal(t, 2, player2.RequirementModifiers[0].Amount, "Should be lenience (+2)")
	assert.Equal(t, []model.ResourceType{model.ResourceGlobalParameter}, player2.RequirementModifiers[0].AffectedResources)

	// Add Soletta (space tag, no requirements) - should get discount only
	solettaID := "203"
	err = playerRepo.AddCard(ctx, gameID, playerID, solettaID)
	require.NoError(t, err)

	player3, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player3.RequirementModifiers, 2, "Should have 2 modifiers (lenience + discount)")

	// Verify we have both types
	hasLenience := false
	hasDiscount := false
	for _, mod := range player3.RequirementModifiers {
		if mod.Amount == 2 && len(mod.AffectedResources) == 1 && mod.AffectedResources[0] == model.ResourceGlobalParameter {
			hasLenience = true
			assert.Equal(t, livestockID, *mod.CardTarget, "Lenience should target Livestock")
		}
		if mod.Amount == -2 && len(mod.AffectedResources) == 1 && mod.AffectedResources[0] == model.ResourceCredits {
			hasDiscount = true
			assert.Equal(t, solettaID, *mod.CardTarget, "Discount should target Soletta")
		}
	}
	assert.True(t, hasLenience, "Should have lenience modifier")
	assert.True(t, hasDiscount, "Should have discount modifier")

	// Add card with NO tags and NO requirements - should not add any modifiers
	specialDesignID := "206"
	err = playerRepo.AddCard(ctx, gameID, playerID, specialDesignID)
	require.NoError(t, err)

	player4, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player4.RequirementModifiers, 2, "Should still have 2 modifiers (Special Design doesn't match)")

	// Remove Livestock - should leave only discount for Soletta
	err = playerRepo.RemoveCard(ctx, gameID, playerID, livestockID)
	require.NoError(t, err)

	player5, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player5.RequirementModifiers, 1, "Should have 1 modifier (discount only)")
	assert.Equal(t, -2, player5.RequirementModifiers[0].Amount)
	assert.Equal(t, solettaID, *player5.RequirementModifiers[0].CardTarget)

	// Remove Soletta - should have no modifiers
	err = playerRepo.RemoveCard(ctx, gameID, playerID, solettaID)
	require.NoError(t, err)

	player6, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.Empty(t, player6.RequirementModifiers, "Should have no modifiers after removing all matching cards")

	t.Log("✅ Mixed lenience and discount test passed")
}

// TestDiscountAppliedDuringValidation tests that discounts are correctly applied during card affordability validation
func TestDiscountAppliedDuringValidation(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Load cards from JSON
	err := cardRepo.LoadCards(ctx)
	require.NoError(t, err, "Failed to load cards from JSON")

	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)
	cardManager := cards.NewCardManager(gameRepo, playerRepo, cardRepo, cardDeckRepo, effectSubscriber)

	// Create game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 2})
	require.NoError(t, err)
	gameID := game.ID

	// Create player with 34 credits (not enough for Soletta's 35 MC cost normally)
	playerID := "player-1"
	player := model.Player{
		ID:   playerID,
		Name: "Test Player",
		Resources: model.Resources{
			Credits: 34, // One credit short of Soletta's 35 MC cost
		},
		Cards:   []string{},
		Effects: []model.PlayerEffect{},
	}
	require.NoError(t, playerRepo.Create(ctx, gameID, player))

	// Create Space Station card behavior (discount -2 MC for space tags)
	spaceStationCard := &model.Card{
		ID:   "025",
		Name: "Space Station",
		Type: model.CardTypeActive,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{{Type: model.ResourceTriggerAuto}},
				Outputs: []model.ResourceCondition{
					{
						Type:              model.ResourceDiscount,
						Amount:            -2,
						Target:            model.TargetSelfPlayer,
						AffectedTags:      []model.CardTag{model.TagSpace},
						AffectedResources: []string{string(model.ResourceCredits)},
					},
				},
			},
		},
	}

	// Subscribe Space Station effect
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, spaceStationCard.ID, spaceStationCard)
	require.NoError(t, err)

	// Add Soletta (space-tagged card, 35 MC cost) to player's hand
	solettaID := "203"
	err = playerRepo.AddCard(ctx, gameID, playerID, solettaID)
	require.NoError(t, err)

	// Verify discount modifier was created for Soletta
	player1, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	require.Len(t, player1.RequirementModifiers, 1, "Should have 1 modifier for Soletta")

	modifier := player1.RequirementModifiers[0]
	assert.Equal(t, -2, modifier.Amount, "Should have -2 MC discount")
	require.NotNil(t, modifier.CardTarget, "Should target Soletta")
	assert.Equal(t, solettaID, *modifier.CardTarget)

	// Test 1: Try to play Soletta with only 33 credits payment (should succeed with discount)
	payment := &model.CardPayment{
		Credits:  33, // Effective cost after 2 MC discount: 35 - 2 = 33
		Steel:    0,
		Titanium: 0,
	}

	// This should succeed because discount makes card cost 33 MC
	err = cardManager.CanPlay(ctx, gameID, playerID, solettaID, payment, nil, nil)
	assert.NoError(t, err, "Should be able to play Soletta with 34 credits when discount reduces cost to 33")

	// Test 2: Try to play with only 32 credits (should fail, even with discount)
	payment2 := &model.CardPayment{
		Credits:  32,
		Steel:    0,
		Titanium: 0,
	}

	err = cardManager.CanPlay(ctx, gameID, playerID, solettaID, payment2, nil, nil)
	assert.Error(t, err, "Should not be able to play Soletta with only 32 credits when discounted cost is 33")

	// Test 3: Try to play with 34 credits (should succeed, overpayment is allowed)
	payment3 := &model.CardPayment{
		Credits:  34,
		Steel:    0,
		Titanium: 0,
	}

	err = cardManager.CanPlay(ctx, gameID, playerID, solettaID, payment3, nil, nil)
	assert.NoError(t, err, "Should be able to play Soletta with 34 credits (overpayment allowed)")

	t.Log("✅ Discount applied during validation test passed")
}
