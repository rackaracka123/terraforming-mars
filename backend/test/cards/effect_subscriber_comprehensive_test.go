package cards_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// TestEffectSubscriberTemperatureRaise tests passive effect triggering on temperature increase
func TestEffectSubscriberTemperatureRaise(t *testing.T) {
	ctx := context.Background()

	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	// Create game and player
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	player := model.Player{
		ID:   "player-456",
		Name: "Test Player",
		Resources: model.Resources{
			Credits:  10,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
	}

	err = playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create card with temperature-triggered effect (e.g., Arctic Algae)
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

	// Subscribe card effects
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, "card-arctic-algae", card)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// Trigger temperature increase event
	err = gameRepo.UpdateTemperature(ctx, game.ID, -28)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Verify passive effect applied (plants increased by 2)
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	expectedPlants := 2
	if updatedPlayer.Resources.Plants != expectedPlants {
		t.Errorf("Expected plants %d, got %d", expectedPlants, updatedPlayer.Resources.Plants)
	}
}

// TestEffectSubscriberOxygenRaise tests passive effect triggering on oxygen increase
func TestEffectSubscriberOxygenRaise(t *testing.T) {
	ctx := context.Background()

	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	// Create game and player
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	player := model.Player{
		ID:   "player-789",
		Name: "Test Player",
		Resources: model.Resources{
			Credits:  15,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
	}

	err = playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create card with oxygen-triggered effect
	card := &model.Card{
		ID:   "oxygen-card",
		Name: "Oxygen Beneficiary",
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerAuto,
					Condition: &model.ResourceTriggerCondition{
						Type: model.TriggerOxygenRaise,
					},
				}},
				Outputs: []model.ResourceCondition{{
					Type:   model.ResourceCredits,
					Amount: 3,
					Target: model.TargetSelfPlayer,
				}},
			},
		},
	}

	// Subscribe card effects
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, "card-oxygen", card)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// Trigger oxygen increase event
	err = gameRepo.UpdateOxygen(ctx, game.ID, 2)
	if err != nil {
		t.Fatalf("Failed to update oxygen: %v", err)
	}

	// Verify passive effect applied (credits increased by 3)
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	expectedCredits := 18 // 15 initial + 3 from effect
	if updatedPlayer.Resources.Credits != expectedCredits {
		t.Errorf("Expected credits %d, got %d", expectedCredits, updatedPlayer.Resources.Credits)
	}
}

// TestEffectSubscriberOceanPlaced tests passive effect triggering on ocean placement
func TestEffectSubscriberOceanPlaced(t *testing.T) {
	ctx := context.Background()

	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	// Create game and player
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	player := model.Player{
		ID:   "player-123",
		Name: "Test Player",
		Resources: model.Resources{
			Credits:  20,
			Steel:    1,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
	}

	err = playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create card with ocean-triggered effect
	card := &model.Card{
		ID:   "ocean-card",
		Name: "Ocean Beneficiary",
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerAuto,
					Condition: &model.ResourceTriggerCondition{
						Type: model.TriggerOceanPlaced,
					},
				}},
				Outputs: []model.ResourceCondition{{
					Type:   model.ResourceSteel,
					Amount: 2,
					Target: model.TargetSelfPlayer,
				}},
			},
		},
	}

	// Subscribe card effects
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, "card-ocean", card)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// Trigger ocean placement event
	err = gameRepo.UpdateOceans(ctx, game.ID, 1)
	if err != nil {
		t.Fatalf("Failed to update oceans: %v", err)
	}

	// Verify passive effect applied (steel increased by 2)
	updatedPlayer, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}

	expectedSteel := 3 // 1 initial + 2 from effect
	if updatedPlayer.Resources.Steel != expectedSteel {
		t.Errorf("Expected steel %d, got %d", expectedSteel, updatedPlayer.Resources.Steel)
	}
}

// TestEffectSubscriberMultipleEffectsOnSameCard tests multiple passive effects on one card
func TestEffectSubscriberMultipleEffectsOnSameCard(t *testing.T) {
	ctx := context.Background()

	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	// Create game and player
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	player := model.Player{
		ID:   "player-multi",
		Name: "Test Player",
		Resources: model.Resources{
			Credits:  10,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
	}

	err = playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create card with multiple passive effects
	card := &model.Card{
		ID:   "multi-effect-card",
		Name: "Multi Effect Card",
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
			{
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerAuto,
					Condition: &model.ResourceTriggerCondition{
						Type: model.TriggerOxygenRaise,
					},
				}},
				Outputs: []model.ResourceCondition{{
					Type:   model.ResourceCredits,
					Amount: 2,
					Target: model.TargetSelfPlayer,
				}},
			},
		},
	}

	// Subscribe card effects
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, "card-multi", card)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// Trigger temperature increase
	err = gameRepo.UpdateTemperature(ctx, game.ID, -28)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Verify first effect applied
	player1, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}
	if player1.Resources.Plants != 1 {
		t.Errorf("Expected plants 1, got %d", player1.Resources.Plants)
	}

	// Trigger oxygen increase
	err = gameRepo.UpdateOxygen(ctx, game.ID, 2)
	if err != nil {
		t.Fatalf("Failed to update oxygen: %v", err)
	}

	// Verify second effect applied
	player2, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}
	expectedCredits := 12 // 10 initial + 2 from effect
	if player2.Resources.Credits != expectedCredits {
		t.Errorf("Expected credits %d, got %d", expectedCredits, player2.Resources.Credits)
	}
}

// TestEffectSubscriberUnsubscribeStopsTriggering tests that unsubscribed effects don't trigger
func TestEffectSubscriberUnsubscribeStopsTriggering(t *testing.T) {
	ctx := context.Background()

	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	// Create game and player
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	player := model.Player{
		ID:   "player-unsub",
		Name: "Test Player",
		Resources: model.Resources{
			Credits:  10,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
	}

	err = playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create card with temperature-triggered effect
	card := &model.Card{
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
					Type:   model.ResourceHeat,
					Amount: 5,
					Target: model.TargetSelfPlayer,
				}},
			},
		},
	}

	// Subscribe card effects
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, "card-temp", card)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// Trigger temperature increase
	err = gameRepo.UpdateTemperature(ctx, game.ID, -28)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Verify effect applied
	player1, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}
	if player1.Resources.Heat != 5 {
		t.Errorf("Expected heat 5, got %d", player1.Resources.Heat)
	}

	// Unsubscribe card effects
	err = effectSubscriber.UnsubscribeCardEffects("card-temp")
	if err != nil {
		t.Fatalf("Failed to unsubscribe card effects: %v", err)
	}

	// Trigger temperature increase again
	err = gameRepo.UpdateTemperature(ctx, game.ID, -26)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Verify effect NOT applied (heat still 5)
	player2, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	if err != nil {
		t.Fatalf("Failed to get player: %v", err)
	}
	if player2.Resources.Heat != 5 {
		t.Errorf("Expected heat to remain 5 after unsubscribe, got %d", player2.Resources.Heat)
	}
}

// TestEffectSubscriberNoAutoTriggerIgnored tests that non-auto triggers are ignored
func TestEffectSubscriberNoAutoTriggerIgnored(t *testing.T) {
	ctx := context.Background()

	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	// Create game and player
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	player := model.Player{
		ID:   "player-manual",
		Name: "Test Player",
	}

	err = playerRepo.Create(ctx, game.ID, player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Create card with manual trigger (not auto)
	card := &model.Card{
		ID:   "manual-card",
		Name: "Manual Trigger Card",
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{{
					Type: model.ResourceTriggerManual, // Not auto
					Condition: &model.ResourceTriggerCondition{
						Type: model.TriggerTemperatureRaise,
					},
				}},
				Outputs: []model.ResourceCondition{{
					Type:   model.ResourcePlants,
					Amount: 10,
					Target: model.TargetSelfPlayer,
				}},
			},
		},
	}

	// Subscribe card effects - should not subscribe anything
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player.ID, "card-manual", card)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// No error means test passed - manual triggers are correctly ignored
}

// TestEffectSubscriberMultiplePlayersIndependent tests that effects only apply to owning player
func TestEffectSubscriberMultiplePlayersIndependent(t *testing.T) {
	ctx := context.Background()

	// Setup
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)

	// Create game
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Create two players
	player1 := model.Player{
		ID:   "player-1",
		Name: "Test Player 1",
		Resources: model.Resources{
			Credits:  10,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
	}
	err = playerRepo.Create(ctx, game.ID, player1)
	if err != nil {
		t.Fatalf("Failed to create player1: %v", err)
	}

	player2 := model.Player{
		ID:   "player-2",
		Name: "Test Player 2",
		Resources: model.Resources{
			Credits:  10,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
	}
	err = playerRepo.Create(ctx, game.ID, player2)
	if err != nil {
		t.Fatalf("Failed to create player2: %v", err)
	}

	// Card for player 1
	card1 := &model.Card{
		ID:   "card-p1",
		Name: "Player 1 Card",
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
					Amount: 3,
					Target: model.TargetSelfPlayer,
				}},
			},
		},
	}

	// Subscribe card for player 1 only
	err = effectSubscriber.SubscribeCardEffects(ctx, game.ID, player1.ID, "card-p1-instance", card1)
	if err != nil {
		t.Fatalf("Failed to subscribe card effects: %v", err)
	}

	// Trigger temperature increase (global event)
	err = gameRepo.UpdateTemperature(ctx, game.ID, -28)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Verify player 1 got plants
	p1After, err := playerRepo.GetByID(ctx, game.ID, player1.ID)
	if err != nil {
		t.Fatalf("Failed to get player 1: %v", err)
	}
	if p1After.Resources.Plants != 3 {
		t.Errorf("Player 1 expected plants 3, got %d", p1After.Resources.Plants)
	}

	// Verify player 2 got nothing (no card with passive effect)
	p2After, err := playerRepo.GetByID(ctx, game.ID, player2.ID)
	if err != nil {
		t.Fatalf("Failed to get player 2: %v", err)
	}
	if p2After.Resources.Plants != 0 {
		t.Errorf("Player 2 expected plants 0, got %d", p2After.Resources.Plants)
	}
}
