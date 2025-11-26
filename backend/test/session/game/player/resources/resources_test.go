package resources_test

import (
	"terraforming-mars-backend/internal/session/game/player/resources"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/types"
)

func TestNewResources(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	resources := r.Get()
	if resources.Credits != 0 {
		t.Error("Expected 0 credits initially")
	}

	production := r.Production()
	if production.Credits != 0 {
		t.Error("Expected 0 production initially")
	}

	if r.TerraformRating() != 20 {
		t.Errorf("Expected terraform rating of 20, got %d", r.TerraformRating())
	}

	if r.VictoryPoints() != 0 {
		t.Errorf("Expected 0 victory points, got %d", r.VictoryPoints())
	}

	if len(r.Storage()) != 0 {
		t.Error("Expected empty storage")
	}

	if len(r.PaymentSubstitutes()) != 0 {
		t.Error("Expected empty payment substitutes")
	}
}

func TestResources_Set(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	newResources := types.Resources{
		Credits:  100,
		Steel:    5,
		Titanium: 3,
		Plants:   10,
		Energy:   7,
		Heat:     2,
	}

	r.Set(newResources)

	result := r.Get()
	if result.Credits != 100 {
		t.Errorf("Expected 100 credits, got %d", result.Credits)
	}
	if result.Steel != 5 {
		t.Errorf("Expected 5 steel, got %d", result.Steel)
	}
}

func TestResources_Set_PublishesEvent(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	var receivedEvent events.ResourcesChangedEvent
	events.Subscribe(eventBus, func(e events.ResourcesChangedEvent) {
		receivedEvent = e
	})

	newResources := types.Resources{Credits: 50}
	r.Set(newResources)

	if receivedEvent.GameID != "game-1" {
		t.Error("Expected event with correct game ID")
	}
	if receivedEvent.PlayerID != "player-1" {
		t.Error("Expected event with correct player ID")
	}
	if receivedEvent.Changes["credits"] != 50 {
		t.Errorf("Expected credits change of 50, got %d", receivedEvent.Changes["credits"])
	}
}

func TestResources_Add(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	// Set initial resources
	r.Set(types.Resources{Credits: 100, Steel: 10})

	// Add more resources
	r.Add(map[types.ResourceType]int{
		types.ResourceCredits: 25,
		types.ResourceSteel:   5,
	})

	result := r.Get()
	if result.Credits != 125 {
		t.Errorf("Expected 125 credits, got %d", result.Credits)
	}
	if result.Steel != 15 {
		t.Errorf("Expected 15 steel, got %d", result.Steel)
	}
}

func TestResources_Add_PublishesBatchedEvent(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	var receivedEvent events.ResourcesChangedEvent
	events.Subscribe(eventBus, func(e events.ResourcesChangedEvent) {
		receivedEvent = e
	})

	r.Add(map[types.ResourceType]int{
		types.ResourceCredits: 10,
		types.ResourceSteel:   5,
		types.ResourcePlants:  -3,
	})

	// Verify single event with all changes
	if len(receivedEvent.Changes) != 3 {
		t.Errorf("Expected 3 changes in event, got %d", len(receivedEvent.Changes))
	}
	if receivedEvent.Changes["credits"] != 10 {
		t.Error("Expected credits change in batched event")
	}
	if receivedEvent.Changes["steel"] != 5 {
		t.Error("Expected steel change in batched event")
	}
	if receivedEvent.Changes["plants"] != -3 {
		t.Error("Expected plants change in batched event")
	}
}

func TestResources_SetProduction(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	newProduction := types.Production{
		Credits:  5,
		Steel:    2,
		Titanium: 1,
		Plants:   1,
		Energy:   3,
		Heat:     0,
	}

	r.SetProduction(newProduction)

	result := r.Production()
	if result.Credits != 5 {
		t.Errorf("Expected 5 production credits, got %d", result.Credits)
	}
	if result.Steel != 2 {
		t.Errorf("Expected 2 production steel, got %d", result.Steel)
	}
}

func TestResources_AddProduction(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	r.SetProduction(types.Production{Credits: 5, Steel: 2})

	r.AddProduction(map[types.ResourceType]int{
		types.ResourceCredits: 2,
		types.ResourceSteel:   1,
	})

	result := r.Production()
	if result.Credits != 7 {
		t.Errorf("Expected 7 production credits, got %d", result.Credits)
	}
	if result.Steel != 3 {
		t.Errorf("Expected 3 production steel, got %d", result.Steel)
	}
}

func TestResources_SetTerraformRating(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	r.SetTerraformRating(25)

	if r.TerraformRating() != 25 {
		t.Errorf("Expected terraform rating of 25, got %d", r.TerraformRating())
	}
}

func TestResources_SetTerraformRating_PublishesEvent(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	var receivedEvent events.TerraformRatingChangedEvent
	events.Subscribe(eventBus, func(e events.TerraformRatingChangedEvent) {
		receivedEvent = e
	})

	r.SetTerraformRating(30)

	if receivedEvent.GameID != "game-1" {
		t.Error("Expected event with correct game ID")
	}
	if receivedEvent.PlayerID != "player-1" {
		t.Error("Expected event with correct player ID")
	}
	if receivedEvent.OldRating != 20 {
		t.Errorf("Expected old rating of 20, got %d", receivedEvent.OldRating)
	}
	if receivedEvent.NewRating != 30 {
		t.Errorf("Expected new rating of 30, got %d", receivedEvent.NewRating)
	}
}

func TestResources_UpdateTerraformRating(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	r.UpdateTerraformRating(5)

	if r.TerraformRating() != 25 {
		t.Errorf("Expected terraform rating of 25, got %d", r.TerraformRating())
	}

	r.UpdateTerraformRating(-3)

	if r.TerraformRating() != 22 {
		t.Errorf("Expected terraform rating of 22, got %d", r.TerraformRating())
	}
}

func TestResources_UpdateTerraformRating_PublishesEvent(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	var receivedEvent events.TerraformRatingChangedEvent
	events.Subscribe(eventBus, func(e events.TerraformRatingChangedEvent) {
		receivedEvent = e
	})

	r.UpdateTerraformRating(5)

	if receivedEvent.OldRating != 20 {
		t.Errorf("Expected old rating of 20, got %d", receivedEvent.OldRating)
	}
	if receivedEvent.NewRating != 25 {
		t.Errorf("Expected new rating of 25, got %d", receivedEvent.NewRating)
	}
}

func TestResources_SetVictoryPoints(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	r.SetVictoryPoints(42)

	if r.VictoryPoints() != 42 {
		t.Errorf("Expected 42 victory points, got %d", r.VictoryPoints())
	}
}

func TestResources_SetVictoryPoints_PublishesEvent(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	var receivedEvent events.VictoryPointsChangedEvent
	events.Subscribe(eventBus, func(e events.VictoryPointsChangedEvent) {
		receivedEvent = e
	})

	r.SetVictoryPoints(15)

	if receivedEvent.GameID != "game-1" {
		t.Error("Expected event with correct game ID")
	}
	if receivedEvent.PlayerID != "player-1" {
		t.Error("Expected event with correct player ID")
	}
	if receivedEvent.OldPoints != 0 {
		t.Errorf("Expected old points of 0, got %d", receivedEvent.OldPoints)
	}
	if receivedEvent.NewPoints != 15 {
		t.Errorf("Expected new points of 15, got %d", receivedEvent.NewPoints)
	}
}

func TestResources_AddVictoryPoints(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	r.AddVictoryPoints(10)

	if r.VictoryPoints() != 10 {
		t.Errorf("Expected 10 victory points, got %d", r.VictoryPoints())
	}

	r.AddVictoryPoints(5)

	if r.VictoryPoints() != 15 {
		t.Errorf("Expected 15 victory points, got %d", r.VictoryPoints())
	}
}

func TestResources_SetStorage(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	storage := map[string]int{
		"card-1": 5,
		"card-2": 3,
	}

	r.SetStorage(storage)

	result := r.Storage()
	if len(result) != 2 {
		t.Errorf("Expected 2 storage entries, got %d", len(result))
	}
	if result["card-1"] != 5 {
		t.Error("Expected card-1 to have 5 resources")
	}

	// Verify defensive copy
	storage["card-1"] = 999
	result2 := r.Storage()
	if result2["card-1"] != 5 {
		t.Error("Expected storage to not be affected by external modification")
	}
}

func TestResources_AddToStorage(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	r.AddToStorage("card-1", 3)
	r.AddToStorage("card-1", 2)
	r.AddToStorage("card-2", 5)

	storage := r.Storage()
	if storage["card-1"] != 5 {
		t.Errorf("Expected card-1 to have 5 resources, got %d", storage["card-1"])
	}
	if storage["card-2"] != 5 {
		t.Errorf("Expected card-2 to have 5 resources, got %d", storage["card-2"])
	}
}

func TestResources_RemoveFromStorage(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	r.AddToStorage("card-1", 10)

	// Remove some resources
	removed := r.RemoveFromStorage("card-1", 3)
	if !removed {
		t.Error("Expected RemoveFromStorage to return true")
	}

	storage := r.Storage()
	if storage["card-1"] != 7 {
		t.Errorf("Expected card-1 to have 7 resources, got %d", storage["card-1"])
	}

	// Try to remove more than available
	removed = r.RemoveFromStorage("card-1", 10)
	if removed {
		t.Error("Expected RemoveFromStorage to return false when insufficient resources")
	}

	// Try to remove from non-existent card
	removed = r.RemoveFromStorage("non-existent", 1)
	if removed {
		t.Error("Expected RemoveFromStorage to return false for non-existent card")
	}
}

func TestResources_SetPaymentSubstitutes(t *testing.T) {
	eventBus := events.NewEventBus()
	r := NewResources(eventBus, "game-1", "player-1")

	substitutes := []card.PaymentSubstitute{
		{ResourceType: types.ResourceHeat, ConversionRate: 1},
		{ResourceType: types.ResourceTitanium, ConversionRate: 3},
	}

	r.SetPaymentSubstitutes(substitutes)

	result := r.PaymentSubstitutes()
	if len(result) != 2 {
		t.Errorf("Expected 2 substitutes, got %d", len(result))
	}

	// Verify defensive copy
	substitutes[0].ConversionRate = 999
	result2 := r.PaymentSubstitutes()
	if result2[0].ConversionRate != 1 {
		t.Error("Expected substitutes to not be affected by external modification")
	}
}

func TestResources_DeepCopy(t *testing.T) {
	eventBus := events.NewEventBus()
	original := NewResources(eventBus, "game-1", "player-1")

	original.Set(types.Resources{Credits: 100, Steel: 10})
	original.SetProduction(types.Production{Credits: 5, Steel: 2})
	original.SetTerraformRating(25)
	original.SetVictoryPoints(20)
	original.AddToStorage("card-1", 5)
	original.SetPaymentSubstitutes([]card.PaymentSubstitute{{ResourceType: types.ResourceHeat, ConversionRate: 1}})

	copy := original.DeepCopy()

	// Verify copy has same values
	if copy.Get().Credits != original.Get().Credits {
		t.Error("Expected copy to have same resources")
	}
	if copy.Production().Credits != original.Production().Credits {
		t.Error("Expected copy to have same production")
	}
	if copy.TerraformRating() != original.TerraformRating() {
		t.Error("Expected copy to have same terraform rating")
	}
	if copy.VictoryPoints() != original.VictoryPoints() {
		t.Error("Expected copy to have same victory points")
	}
	if len(copy.Storage()) != len(original.Storage()) {
		t.Error("Expected copy to have same storage")
	}

	// Verify modifying copy doesn't affect original
	copy.Set(types.Resources{Credits: 999})
	if original.Get().Credits != 100 {
		t.Error("Expected original resources to remain unchanged")
	}

	copy.AddToStorage("card-2", 10)
	if len(original.Storage()) != 1 {
		t.Error("Expected original storage to remain unchanged")
	}
}

func TestResources_DeepCopy_Nil(t *testing.T) {
	var r *Resources = nil
	copy := r.DeepCopy()

	if copy != nil {
		t.Error("Expected DeepCopy of nil to return nil")
	}
}
