package resources

import (
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/types"
)

// Resources manages player resources, production, scoring, and payment substitutes.
// Thread-safe with its own mutex.
// Publishes events when resources or terraform rating change.
type Resources struct {
	mu sync.RWMutex

	// Core resources
	resources  types.Resources
	production types.Production

	// Scoring
	terraformRating int
	victoryPoints   int

	// Advanced
	resourceStorage    map[string]int // cardID -> resource count
	paymentSubstitutes []card.PaymentSubstitute

	// Event publishing (injected)
	eventBus *events.EventBusImpl
	gameID   string
	playerID string
}

// NewResources creates a new Resources component with default values.
// Starting terraform rating is 20, all other values are 0.
func NewResources(eventBus *events.EventBusImpl, gameID, playerID string) *Resources {
	return &Resources{
		resources: types.Resources{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		production: types.Production{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		terraformRating:    20,
		victoryPoints:      0,
		resourceStorage:    make(map[string]int),
		paymentSubstitutes: []card.PaymentSubstitute{},
		eventBus:           eventBus,
		gameID:             gameID,
		playerID:           playerID,
	}
}

// ==================== Getters ====================

// Get returns a copy of the current resources.
func (r *Resources) Get() types.Resources {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resources
}

// Production returns a copy of the current production.
func (r *Resources) Production() types.Production {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.production
}

// TerraformRating returns the current terraform rating.
func (r *Resources) TerraformRating() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.terraformRating
}

// VictoryPoints returns the current victory points.
func (r *Resources) VictoryPoints() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.victoryPoints
}

// Storage returns a defensive copy of resource storage.
func (r *Resources) Storage() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.resourceStorage == nil {
		return make(map[string]int)
	}
	storageCopy := make(map[string]int, len(r.resourceStorage))
	for k, v := range r.resourceStorage {
		storageCopy[k] = v
	}
	return storageCopy
}

// PaymentSubstitutes returns a defensive copy of payment substitutes.
func (r *Resources) PaymentSubstitutes() []card.PaymentSubstitute {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.paymentSubstitutes == nil {
		return []card.PaymentSubstitute{}
	}
	substitutesCopy := make([]card.PaymentSubstitute, len(r.paymentSubstitutes))
	copy(substitutesCopy, r.paymentSubstitutes)
	return substitutesCopy
}

// ==================== Setters (with Event Publishing) ====================

// Set replaces all resources and publishes a ResourcesChangedEvent with batched changes.
func (r *Resources) Set(resources types.Resources) {
	r.mu.Lock()

	// Calculate delta for event
	oldResources := r.resources
	changes := make(map[string]int)

	if oldResources.Credits != resources.Credits {
		changes["credits"] = resources.Credits - oldResources.Credits
	}
	if oldResources.Steel != resources.Steel {
		changes["steel"] = resources.Steel - oldResources.Steel
	}
	if oldResources.Titanium != resources.Titanium {
		changes["titanium"] = resources.Titanium - oldResources.Titanium
	}
	if oldResources.Plants != resources.Plants {
		changes["plants"] = resources.Plants - oldResources.Plants
	}
	if oldResources.Energy != resources.Energy {
		changes["energy"] = resources.Energy - oldResources.Energy
	}
	if oldResources.Heat != resources.Heat {
		changes["heat"] = resources.Heat - oldResources.Heat
	}

	// Update state
	r.resources = resources

	eventBus := r.eventBus
	gameID := r.gameID
	playerID := r.playerID

	r.mu.Unlock()

	// Publish event AFTER releasing lock
	if eventBus != nil && len(changes) > 0 {
		events.Publish(eventBus, events.ResourcesChangedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			Changes:   changes,
			Timestamp: time.Now(),
		})
	}
}

// SetProduction replaces all production values.
func (r *Resources) SetProduction(production types.Production) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.production = production
}

// SetTerraformRating sets the terraform rating and publishes a TerraformRatingChangedEvent.
func (r *Resources) SetTerraformRating(rating int) {
	r.mu.Lock()

	oldRating := r.terraformRating
	r.terraformRating = rating

	eventBus := r.eventBus
	gameID := r.gameID
	playerID := r.playerID

	r.mu.Unlock()

	// Publish event AFTER releasing lock
	if eventBus != nil && oldRating != rating {
		events.Publish(eventBus, events.TerraformRatingChangedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			OldRating: oldRating,
			NewRating: rating,
			Timestamp: time.Now(),
		})
	}
}

// SetVictoryPoints sets the victory points and publishes a VictoryPointsChangedEvent.
func (r *Resources) SetVictoryPoints(points int) {
	r.mu.Lock()

	oldPoints := r.victoryPoints
	r.victoryPoints = points

	eventBus := r.eventBus
	gameID := r.gameID
	playerID := r.playerID

	r.mu.Unlock()

	// Publish event AFTER releasing lock
	if eventBus != nil && oldPoints != points {
		events.Publish(eventBus, events.VictoryPointsChangedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			OldPoints: oldPoints,
			NewPoints: points,
			Source:    "direct",
			Timestamp: time.Now(),
		})
	}
}

// SetStorage replaces the resource storage map.
func (r *Resources) SetStorage(storage map[string]int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if storage == nil {
		r.resourceStorage = make(map[string]int)
		return
	}
	r.resourceStorage = make(map[string]int, len(storage))
	for k, v := range storage {
		r.resourceStorage[k] = v
	}
}

// SetPaymentSubstitutes replaces the payment substitutes collection.
func (r *Resources) SetPaymentSubstitutes(substitutes []card.PaymentSubstitute) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if substitutes == nil {
		r.paymentSubstitutes = []card.PaymentSubstitute{}
		return
	}
	r.paymentSubstitutes = make([]card.PaymentSubstitute, len(substitutes))
	copy(r.paymentSubstitutes, substitutes)
}

// ==================== Mutations (with Event Publishing) ====================

// Add applies resource changes and publishes a single batched ResourcesChangedEvent.
func (r *Resources) Add(changes map[types.ResourceType]int) {
	r.mu.Lock()

	eventChanges := make(map[string]int)

	for resourceType, delta := range changes {
		if delta == 0 {
			continue
		}

		switch resourceType {
		case types.ResourceCredits:
			r.resources.Credits += delta
			eventChanges["credits"] = delta
		case types.ResourceSteel:
			r.resources.Steel += delta
			eventChanges["steel"] = delta
		case types.ResourceTitanium:
			r.resources.Titanium += delta
			eventChanges["titanium"] = delta
		case types.ResourcePlants:
			r.resources.Plants += delta
			eventChanges["plants"] = delta
		case types.ResourceEnergy:
			r.resources.Energy += delta
			eventChanges["energy"] = delta
		case types.ResourceHeat:
			r.resources.Heat += delta
			eventChanges["heat"] = delta
		}
	}

	eventBus := r.eventBus
	gameID := r.gameID
	playerID := r.playerID

	r.mu.Unlock()

	// Publish single batched event AFTER releasing lock
	if eventBus != nil && len(eventChanges) > 0 {
		events.Publish(eventBus, events.ResourcesChangedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			Changes:   eventChanges,
			Timestamp: time.Now(),
		})
	}
}

// AddProduction applies production changes.
func (r *Resources) AddProduction(changes map[types.ResourceType]int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for resourceType, delta := range changes {
		switch resourceType {
		case types.ResourceCredits:
			r.production.Credits += delta
		case types.ResourceSteel:
			r.production.Steel += delta
		case types.ResourceTitanium:
			r.production.Titanium += delta
		case types.ResourcePlants:
			r.production.Plants += delta
		case types.ResourceEnergy:
			r.production.Energy += delta
		case types.ResourceHeat:
			r.production.Heat += delta
		}
	}
}

// UpdateTerraformRating applies a delta to the terraform rating and publishes an event.
func (r *Resources) UpdateTerraformRating(delta int) {
	if delta == 0 {
		return
	}

	r.mu.Lock()

	oldRating := r.terraformRating
	r.terraformRating += delta
	newRating := r.terraformRating

	eventBus := r.eventBus
	gameID := r.gameID
	playerID := r.playerID

	r.mu.Unlock()

	// Publish event AFTER releasing lock
	if eventBus != nil {
		events.Publish(eventBus, events.TerraformRatingChangedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			OldRating: oldRating,
			NewRating: newRating,
			Timestamp: time.Now(),
		})
	}
}

// AddVictoryPoints applies a delta to the victory points and publishes an event.
func (r *Resources) AddVictoryPoints(delta int) {
	if delta == 0 {
		return
	}

	r.mu.Lock()

	oldPoints := r.victoryPoints
	r.victoryPoints += delta
	newPoints := r.victoryPoints

	eventBus := r.eventBus
	gameID := r.gameID
	playerID := r.playerID

	r.mu.Unlock()

	// Publish event AFTER releasing lock
	if eventBus != nil {
		events.Publish(eventBus, events.VictoryPointsChangedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			OldPoints: oldPoints,
			NewPoints: newPoints,
			Source:    "direct",
			Timestamp: time.Now(),
		})
	}
}

// ==================== Storage Operations ====================

// AddToStorage adds resources to a specific card's storage.
func (r *Resources) AddToStorage(cardID string, amount int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.resourceStorage == nil {
		r.resourceStorage = make(map[string]int)
	}
	r.resourceStorage[cardID] += amount
}

// RemoveFromStorage removes resources from a specific card's storage.
// Returns true if resources were available and removed, false otherwise.
func (r *Resources) RemoveFromStorage(cardID string, amount int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.resourceStorage == nil {
		return false
	}
	current, exists := r.resourceStorage[cardID]
	if !exists || current < amount {
		return false
	}
	r.resourceStorage[cardID] -= amount
	if r.resourceStorage[cardID] == 0 {
		delete(r.resourceStorage, cardID)
	}
	return true
}

// ==================== Utilities ====================

// DeepCopy creates a deep copy of the Resources component.
func (r *Resources) DeepCopy() *Resources {
	if r == nil {
		return nil
	}

	storageCopy := make(map[string]int, len(r.resourceStorage))
	for k, v := range r.resourceStorage {
		storageCopy[k] = v
	}

	substitutesCopy := make([]card.PaymentSubstitute, len(r.paymentSubstitutes))
	copy(substitutesCopy, r.paymentSubstitutes)

	return &Resources{
		resources:          r.resources,
		production:         r.production,
		terraformRating:    r.terraformRating,
		victoryPoints:      r.victoryPoints,
		resourceStorage:    storageCopy,
		paymentSubstitutes: substitutesCopy,
		eventBus:           r.eventBus,
		gameID:             r.gameID,
		playerID:           r.playerID,
	}
}
