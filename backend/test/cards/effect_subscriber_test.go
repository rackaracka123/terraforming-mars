package cards_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
)

// TestCardEffectSubscriberTemperatureIncrease tests passive effects triggered by temperature changes
func TestCardEffectSubscriberTemperatureIncrease(t *testing.T) {
	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)

	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	ctx := context.Background()

	// Create game and player
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	player := model.Player{
		ID:   "player-123",
		Name: "Test Player",
		Resources: model.Resources{
			Credits: 10,
			Plants:  0,
		},
	}

	err = playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create a mock card with passive effect (like Arctic Algae)
	// "When temperature increases, gain 2 plants"
	card := &model.Card{
		ID:   "arctic-algae",
		Name: "Arctic Algae",
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerAuto,
					Condition: &model.ResourceTriggerCondition{
						Type: model.TriggerTemperatureRaise,
					},
				}},
				Outputs: []model.ResourceCondition{{
					Type:   model.ResourcePlants,
					Amount: 2,
					Target: model.TargetSelfPlayer,
				}},
			},
		},
	}

	// Subscribe the card's passive effects
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, card.ID, card)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// Increase temperature (should trigger passive effect)
	err = gameRepo.UpdateTemperature(ctx, game.ID, -28)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Verify player received plants
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if updatedPlayer.Resources.Plants != 2 {
		t.Errorf("Expected player to have 2 plants, got %d", updatedPlayer.Resources.Plants)
	}

	// Increase temperature again (should trigger again)
	err = gameRepo.UpdateTemperature(ctx, game.ID, -26)
	if err != nil {
		t.Fatalf("Failed to update temperature second time: %v", err)
	}

	// Verify player received more plants
	updatedPlayer, err = playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if updatedPlayer.Resources.Plants != 4 {
		t.Errorf("Expected player to have 4 plants after second trigger, got %d", updatedPlayer.Resources.Plants)
	}
}

// TestCardEffectSubscriberOxygenIncrease tests passive effects triggered by oxygen changes
func TestCardEffectSubscriberOxygenIncrease(t *testing.T) {
	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)

	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	ctx := context.Background()

	// Create game and player
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	player := model.Player{
		ID:   "player-123",
		Name: "Test Player",
		Resources: model.Resources{
			Credits: 10,
			Heat:    0,
		},
	}

	err = playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create a mock card with oxygen-triggered passive effect
	// "When oxygen increases, gain 3 heat"
	card := &model.Card{
		ID:   "oxygen-card",
		Name: "Oxygen Card",
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerAuto,
					Condition: &model.ResourceTriggerCondition{
						Type: model.TriggerOxygenRaise,
					},
				}},
				Outputs: []model.ResourceCondition{{
					Type:   model.ResourceHeat,
					Amount: 3,
					Target: model.TargetSelfPlayer,
				}},
			},
		},
	}

	// Subscribe the card's passive effects
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, card.ID, card)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// Increase oxygen (should trigger passive effect)
	err = gameRepo.UpdateOxygen(ctx, game.ID, 2)
	if err != nil {
		t.Fatalf("Failed to update oxygen: %v", err)
	}

	// Verify player received heat
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if updatedPlayer.Resources.Heat != 3 {
		t.Errorf("Expected player to have 3 heat, got %d", updatedPlayer.Resources.Heat)
	}
}

// TestCardEffectSubscriberUnsubscribe tests that unsubscribed effects don't trigger
func TestCardEffectSubscriberUnsubscribe(t *testing.T) {
	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)

	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	ctx := context.Background()

	// Create game and player
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	player := model.Player{
		ID:   "player-123",
		Name: "Test Player",
		Resources: model.Resources{
			Plants: 0,
		},
	}

	err = playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create card with passive effect
	card := &model.Card{
		ID:   "test-card",
		Name: "Test Card",
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerAuto,
					Condition: &model.ResourceTriggerCondition{
						Type: model.TriggerTemperatureRaise,
					},
				}},
				Outputs: []model.ResourceCondition{{
					Type:   model.ResourcePlants,
					Amount: 2,
					Target: model.TargetSelfPlayer,
				}},
			},
		},
	}

	// Subscribe effects
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, card.ID, card)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// Trigger effect once
	err = gameRepo.UpdateTemperature(ctx, game.ID, -28)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Verify effect triggered
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if updatedPlayer.Resources.Plants != 2 {
		t.Errorf("Expected 2 plants before unsubscribe, got %d", updatedPlayer.Resources.Plants)
	}

	// Unsubscribe effects
	err = effectSubscriber.UnsubscribeCardEffects(card.ID)
	if err != nil {
		t.Fatalf("Failed to unsubscribe card effects: %v", err)
	}

	// Trigger temperature change again
	err = gameRepo.UpdateTemperature(ctx, game.ID, -26)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Verify effect did NOT trigger (plants should still be 2)
	updatedPlayer, err = playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if updatedPlayer.Resources.Plants != 2 {
		t.Errorf("Expected 2 plants after unsubscribe (no change), got %d", updatedPlayer.Resources.Plants)
	}
}

// TestCardEffectSubscriberNoPassiveEffects tests cards with no passive effects
func TestCardEffectSubscriberNoPassiveEffects(t *testing.T) {
	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)

	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	ctx := context.Background()

	// Create game and player
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	player := model.Player{
		ID:   "player-123",
		Name: "Test Player",
	}

	err = playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create card with no behaviors
	card := &model.Card{
		ID:   "simple-card",
		Name: "Simple Card",
		// No behaviors array
	}

	// Subscribe effects (should handle gracefully)
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, card.ID, card)
	if err != nil {
		t.Errorf("Subscribe should not error for cards with no passive effects: %v", err)
	}
}

// TestCardEffectSubscriberMultipleCards tests multiple cards with different passive effects
func TestCardEffectSubscriberMultipleCards(t *testing.T) {
	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)

	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	ctx := context.Background()

	// Create game and player
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	player := model.Player{
		ID:   "player-123",
		Name: "Test Player",
		Resources: model.Resources{
			Plants: 0,
			Heat:   0,
		},
	}

	err = playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Card 1: Temperature -> Plants
	card1 := &model.Card{
		ID:   "temp-card",
		Name: "Temperature Card",
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerAuto,
					Condition: &model.ResourceTriggerCondition{
						Type: model.TriggerTemperatureRaise,
					},
				}},
				Outputs: []model.ResourceCondition{{
					Type:   model.ResourcePlants,
					Amount: 1,
					Target: model.TargetSelfPlayer,
				}},
			},
		},
	}

	// Card 2: Oxygen -> Heat
	card2 := &model.Card{
		ID:   "oxygen-card",
		Name: "Oxygen Card",
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerAuto,
					Condition: &model.ResourceTriggerCondition{
						Type: model.TriggerOxygenRaise,
					},
				}},
				Outputs: []model.ResourceCondition{{
					Type:   model.ResourceHeat,
					Amount: 2,
					Target: model.TargetSelfPlayer,
				}},
			},
		},
	}

	// Subscribe both cards
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, card1.ID, card1)
	if err != nil {
		t.Fatalf("Failed to subscribe card1 effects: %v", err)
	}

	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, card2.ID, card2)
	if err != nil {
		t.Fatalf("Failed to subscribe card2 effects: %v", err)
	}

	// Trigger temperature change (should only trigger card1)
	err = gameRepo.UpdateTemperature(ctx, game.ID, -28)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Check resources
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if updatedPlayer.Resources.Plants != 1 {
		t.Errorf("Expected 1 plant from card1, got %d", updatedPlayer.Resources.Plants)
	}
	if updatedPlayer.Resources.Heat != 0 {
		t.Errorf("Expected 0 heat (oxygen not triggered), got %d", updatedPlayer.Resources.Heat)
	}

	// Trigger oxygen change (should only trigger card2)
	err = gameRepo.UpdateOxygen(ctx, game.ID, 2)
	if err != nil {
		t.Fatalf("Failed to update oxygen: %v", err)
	}

	// Check resources again
	updatedPlayer, err = playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	if updatedPlayer.Resources.Plants != 1 {
		t.Errorf("Expected plants to remain 1, got %d", updatedPlayer.Resources.Plants)
	}
	if updatedPlayer.Resources.Heat != 2 {
		t.Errorf("Expected 2 heat from card2, got %d", updatedPlayer.Resources.Heat)
	}
}
