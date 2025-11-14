package repository_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
)

// TestGameRepositoryPublishesTemperatureChangedEvent tests that UpdateTemperature publishes events
func TestGameRepositoryPublishesTemperatureChangedEvent(t *testing.T) {
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)

	// Track event
	var eventReceived bool
	var receivedEvent events.TemperatureChangedEvent

	_ = events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		eventReceived = true
		receivedEvent = event
	})

	// Create a game
	ctx := context.Background()
	game, err := gameRepo.Create(ctx, game.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Update temperature
	err = gameRepo.UpdateTemperature(ctx, game.ID, -26)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Verify event was published
	if !eventReceived {
		t.Error("Temperature changed event was not published")
	}

	if receivedEvent.GameID != game.ID {
		t.Errorf("Expected GameID %s, got %s", game.ID, receivedEvent.GameID)
	}
	if receivedEvent.OldValue != -30 {
		t.Errorf("Expected old temperature -30, got %d", receivedEvent.OldValue)
	}
	if receivedEvent.NewValue != -26 {
		t.Errorf("Expected new temperature -26, got %d", receivedEvent.NewValue)
	}
}

// TestGameRepositoryPublishesOxygenChangedEvent tests that UpdateOxygen publishes events
func TestGameRepositoryPublishesOxygenChangedEvent(t *testing.T) {
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)

	// Track event
	var eventReceived bool
	var receivedEvent events.OxygenChangedEvent

	_ = events.Subscribe(eventBus, func(event events.OxygenChangedEvent) {
		eventReceived = true
		receivedEvent = event
	})

	// Create a game
	ctx := context.Background()
	game, err := gameRepo.Create(ctx, game.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Update oxygen
	err = gameRepo.UpdateOxygen(ctx, game.ID, 4)
	if err != nil {
		t.Fatalf("Failed to update oxygen: %v", err)
	}

	// Verify event was published
	if !eventReceived {
		t.Error("Oxygen changed event was not published")
	}

	if receivedEvent.GameID != game.ID {
		t.Errorf("Expected GameID %s, got %s", game.ID, receivedEvent.GameID)
	}
	if receivedEvent.OldValue != 0 {
		t.Errorf("Expected old oxygen 0, got %d", receivedEvent.OldValue)
	}
	if receivedEvent.NewValue != 4 {
		t.Errorf("Expected new oxygen 4, got %d", receivedEvent.NewValue)
	}
}

// TestGameRepositoryPublishesOceansChangedEvent tests that UpdateOceans publishes events
func TestGameRepositoryPublishesOceansChangedEvent(t *testing.T) {
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)

	// Track event
	var eventReceived bool
	var receivedEvent events.OceansChangedEvent

	_ = events.Subscribe(eventBus, func(event events.OceansChangedEvent) {
		eventReceived = true
		receivedEvent = event
	})

	// Create a game
	ctx := context.Background()
	game, err := gameRepo.Create(ctx, game.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Update oceans
	err = gameRepo.UpdateOceans(ctx, game.ID, 3)
	if err != nil {
		t.Fatalf("Failed to update oceans: %v", err)
	}

	// Verify event was published
	if !eventReceived {
		t.Error("Oceans changed event was not published")
	}

	if receivedEvent.GameID != game.ID {
		t.Errorf("Expected GameID %s, got %s", game.ID, receivedEvent.GameID)
	}
	if receivedEvent.OldValue != 0 {
		t.Errorf("Expected old oceans 0, got %d", receivedEvent.OldValue)
	}
	if receivedEvent.NewValue != 3 {
		t.Errorf("Expected new oceans 3, got %d", receivedEvent.NewValue)
	}
}

// TestPlayerRepositoryPublishesResourcesChangedEvent tests that UpdateResources publishes events
func TestPlayerRepositoryPublishesResourcesChangedEvent(t *testing.T) {
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)

	// Track events
	var eventsReceived []events.ResourcesChangedEvent

	_ = events.Subscribe(eventBus, func(event events.ResourcesChangedEvent) {
		eventsReceived = append(eventsReceived, event)
	})

	// Create a player
	ctx := context.Background()
	player := player.Player{
		ID:   "player-123",
		Name: "Test Player",
		Resources: resources.Resources{
			Credits: 10,
			Plants:  2,
		},
	}

	err := playerRepo.Create(ctx, "game-123", player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Update resources
	newResources := resources.Resources{
		Credits: 15,
		Plants:  5,
		Steel:   3,
	}

	err = playerRepo.UpdateResources(ctx, "game-123", "player-123", newResources)
	if err != nil {
		t.Fatalf("Failed to update resources: %v", err)
	}

	// Verify events were published (should get events for credits, plants, and steel)
	if len(eventsReceived) != 3 {
		t.Errorf("Expected 3 resource change events, got %d", len(eventsReceived))
	}

	// Check that we got events for the right resource types
	resourceTypes := make(map[string]bool)
	for _, event := range eventsReceived {
		resourceTypes[event.ResourceType] = true

		if event.GameID != "game-123" {
			t.Errorf("Wrong GameID in event: %s", event.GameID)
		}
		if event.PlayerID != "player-123" {
			t.Errorf("Wrong PlayerID in event: %s", event.PlayerID)
		}
	}

	expectedTypes := []string{"credits", "plants", "steel"}
	for _, expectedType := range expectedTypes {
		if !resourceTypes[expectedType] {
			t.Errorf("Missing event for resource type: %s", expectedType)
		}
	}
}

// TestPlayerRepositoryPublishesTerraformRatingChangedEvent tests TR event publishing
func TestPlayerRepositoryPublishesTerraformRatingChangedEvent(t *testing.T) {
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)

	// Track event
	var eventReceived bool
	var receivedEvent player.TerraformRatingChangedEvent

	_ = events.Subscribe(eventBus, func(event player.TerraformRatingChangedEvent) {
		eventReceived = true
		receivedEvent = event
	})

	// Create a player
	ctx := context.Background()
	player := player.Player{
		ID:              "player-123",
		Name:            "Test Player",
		TerraformRating: 20,
	}

	err := playerRepo.Create(ctx, "game-123", player)
	if err != nil {
		t.Fatalf("Failed to create player: %v", err)
	}

	// Update terraform rating
	err = playerRepo.UpdateTerraformRating(ctx, "game-123", "player-123", 25)
	if err != nil {
		t.Fatalf("Failed to update terraform rating: %v", err)
	}

	// Verify event was published
	if !eventReceived {
		t.Error("Terraform rating changed event was not published")
	}

	if receivedEvent.GameID != "game-123" {
		t.Errorf("Expected GameID game-123, got %s", receivedEvent.GameID)
	}
	if receivedEvent.PlayerID != "player-123" {
		t.Errorf("Expected PlayerID player-123, got %s", receivedEvent.PlayerID)
	}
	if receivedEvent.OldRating != 20 {
		t.Errorf("Expected old rating 20, got %d", receivedEvent.OldRating)
	}
	if receivedEvent.NewRating != 25 {
		t.Errorf("Expected new rating 25, got %d", receivedEvent.NewRating)
	}
}

// TestRepositoryDoesNotPublishWhenNoChange tests that events aren't published when values don't change
func TestRepositoryDoesNotPublishWhenNoChange(t *testing.T) {
	eventBus := events.NewEventBus()
	gameRepo := repository.NewGameRepository(eventBus)

	// Track events
	var tempEventCount int

	_ = events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		tempEventCount++
	})

	// Create a game
	ctx := context.Background()
	game, err := gameRepo.Create(ctx, game.GameSettings{MaxPlayers: 4})
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Update temperature to same value
	err = gameRepo.UpdateTemperature(ctx, game.ID, -30)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Verify no event was published (temperature didn't actually change)
	if tempEventCount != 0 {
		t.Errorf("Expected 0 events when value doesn't change, got %d", tempEventCount)
	}

	// Now actually change it
	err = gameRepo.UpdateTemperature(ctx, game.ID, -28)
	if err != nil {
		t.Fatalf("Failed to update temperature: %v", err)
	}

	// Should have exactly 1 event now
	if tempEventCount != 1 {
		t.Errorf("Expected 1 event after change, got %d", tempEventCount)
	}
}
