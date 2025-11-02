package cards_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// TestImmediateEffectsNotAddedToPlayerEffects verifies that immediate effects
// (like gaining resources or production when playing a card) are NOT added to
// the player's effects list, only passive effects should be there
func TestImmediateEffectsNotAddedToPlayerEffects(t *testing.T) {
	ctx := context.Background()

	// Setup repositories
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)

	// Create test game and player
	playerID := "player-1"
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 2})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}
	gameID := game.ID

	player := model.Player{
		ID:   playerID,
		Name: "Test Player",
		Resources: model.Resources{
			Credits: 50,
		},
		Effects: []model.PlayerEffect{}, // Start with empty effects
	}
	if err := playerRepo.Create(ctx, gameID, player); err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create a card with IMMEDIATE effects only (heat production)
	cardWithImmediateEffect := &model.Card{
		ID:   "soletta",
		Name: "Soletta",
		Type: model.CardTypeAutomated,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{Type: model.ResourceTriggerAuto}, // Auto trigger without condition = immediate
				},
				Outputs: []model.ResourceCondition{
					{
						Type:   model.ResourceHeatProduction, // Immediate effect: gain heat production
						Amount: 7,
						Target: model.TargetSelfPlayer,
					},
				},
			},
		},
	}

	// Subscribe card effects
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, cardWithImmediateEffect.ID, cardWithImmediateEffect)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// Verify player effects list is EMPTY (immediate effects should NOT be added)
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if len(updatedPlayer.Effects) != 0 {
		t.Errorf("Expected 0 effects in player's effects list (immediate effects should not be added), got %d", len(updatedPlayer.Effects))
		for i, effect := range updatedPlayer.Effects {
			t.Logf("Effect %d: CardID=%s, CardName=%s, Outputs=%+v", i, effect.CardID, effect.CardName, effect.Behavior.Outputs)
		}
	}
}

// TestPassiveEffectsAddedToPlayerEffects verifies that passive effects
// (like discounts or value modifiers) ARE correctly added to the player's effects list
func TestPassiveEffectsAddedToPlayerEffects(t *testing.T) {
	ctx := context.Background()

	// Setup repositories
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)

	// Create test game and player
	playerID := "player-1"
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 2})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}
	gameID := game.ID

	player := model.Player{
		ID:      playerID,
		Name:    "Test Player",
		Effects: []model.PlayerEffect{}, // Start with empty effects
	}
	if err := playerRepo.Create(ctx, gameID, player); err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create a card with PASSIVE effect (discount)
	cardWithPassiveEffect := &model.Card{
		ID:   "discount-card",
		Name: "Discount Card",
		Type: model.CardTypeActive,
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{Type: model.ResourceTriggerAuto}, // Auto trigger without condition
				},
				Outputs: []model.ResourceCondition{
					{
						Type:              model.ResourceDiscount, // Passive effect: discount
						Amount:            -2,
						Target:            model.TargetSelfPlayer,
						AffectedTags:      []model.CardTag{model.TagSpace},
						AffectedResources: []string{"credits"},
					},
				},
			},
		},
	}

	// Subscribe card effects
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, cardWithPassiveEffect.ID, cardWithPassiveEffect)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// Verify player effects list has the passive effect
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if len(updatedPlayer.Effects) != 1 {
		t.Errorf("Expected 1 effect in player's effects list (passive effect should be added), got %d", len(updatedPlayer.Effects))
	}

	if len(updatedPlayer.Effects) > 0 {
		effect := updatedPlayer.Effects[0]
		if effect.CardID != "discount-card" {
			t.Errorf("Expected effect from 'discount-card', got CardID=%s", effect.CardID)
		}
		if len(effect.Behavior.Outputs) == 0 || effect.Behavior.Outputs[0].Type != model.ResourceDiscount {
			t.Errorf("Expected discount output type, got %+v", effect.Behavior.Outputs)
		}
	}
}

// TestDiscountEffectAppliedToExistingCards tests the scenario where a player:
// 1. Has cards in hand (Soletta with space tag, Immigration Shuttles with space tag)
// 2. Plays Shuttles card (which gives -2 MC discount for space-tagged cards)
// 3. Expects modifiers to be created immediately for the existing space-tagged cards
func TestDiscountEffectAppliedToExistingCards(t *testing.T) {
	ctx := context.Background()

	// Setup infrastructure
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameRepo := repository.NewGameRepository(eventBus)
	cardRepo := repository.NewCardRepository()

	// Load real cards from JSON
	err := cardRepo.LoadCards(ctx)
	if err != nil {
		t.Fatalf("Failed to load cards from JSON: %v", err)
	}

	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)

	// Create game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 2})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}
	gameID := game.ID

	// Create player
	playerID := "player-1"
	player := model.Player{
		ID:      playerID,
		Name:    "Test Player",
		Cards:   []string{},
		Effects: []model.PlayerEffect{},
	}
	if err := playerRepo.Create(ctx, gameID, player); err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Step 1: Add cards to player's hand BEFORE playing Shuttles
	// Soletta (ID: 203) - has space tag
	err = playerRepo.AddCard(ctx, gameID, playerID, "203")
	if err != nil {
		t.Fatalf("Failed to add Soletta to hand: %v", err)
	}

	// Immigration Shuttles (ID: 198) - has earth and space tags
	err = playerRepo.AddCard(ctx, gameID, playerID, "198")
	if err != nil {
		t.Fatalf("Failed to add Immigration Shuttles to hand: %v", err)
	}

	// Verify player has 2 cards in hand
	player1, err := playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}
	if len(player1.Cards) != 2 {
		t.Fatalf("Player should have 2 cards in hand, got %d", len(player1.Cards))
	}
	if len(player1.RequirementModifiers) != 0 {
		t.Fatalf("Should have 0 modifiers before playing Shuttles, got %d", len(player1.RequirementModifiers))
	}

	// Step 2: Play Shuttles card (ID: 166) which gives -2 MC discount for space-tagged cards
	// Create Shuttles card behavior matching the real card
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

	// Subscribe Shuttles effects
	err = effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, shuttlesCard.ID, shuttlesCard)
	if err != nil {
		t.Fatalf("Failed to subscribe Shuttles effects: %v", err)
	}

	// Step 3: Verify that modifiers were created immediately for existing cards in hand
	player2, err := playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		t.Fatalf("Failed to get player after subscribing Shuttles: %v", err)
	}

	// Should have 2 modifiers (one for Soletta, one for Immigration Shuttles)
	if len(player2.RequirementModifiers) != 2 {
		t.Fatalf("Should have 2 modifiers after playing Shuttles, got %d", len(player2.RequirementModifiers))
	}

	// Verify modifiers have correct properties
	foundSoletta := false
	foundImmigrationShuttles := false

	for _, mod := range player2.RequirementModifiers {
		if mod.Amount != -2 {
			t.Errorf("Discount amount should be -2, got %d", mod.Amount)
		}
		if len(mod.AffectedResources) != 1 || mod.AffectedResources[0] != model.ResourceCredits {
			t.Errorf("Should affect credits, got %v", mod.AffectedResources)
		}
		if mod.CardTarget == nil {
			t.Error("Should target specific cards")
			continue
		}
		if *mod.CardTarget == "203" {
			foundSoletta = true
		} else if *mod.CardTarget == "198" {
			foundImmigrationShuttles = true
		}
	}

	if !foundSoletta {
		t.Error("Should have modifier for Soletta (203)")
	}
	if !foundImmigrationShuttles {
		t.Error("Should have modifier for Immigration Shuttles (198)")
	}

	t.Log("✅ Discount effect applied to existing cards test passed")
}
