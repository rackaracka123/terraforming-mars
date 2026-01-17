package player

import (
	"sync"
	"terraforming-mars-backend/internal/game/shared"
	"time"

	"terraforming-mars-backend/internal/events"
)

// PlayerResources manages player resources, production, scoring
type PlayerResources struct {
	mu                 sync.RWMutex
	resources          shared.Resources
	production         shared.Production
	terraformRating    int
	victoryPoints      int
	resourceStorage    map[string]int
	paymentSubstitutes []shared.PaymentSubstitute
	valueModifiers     map[shared.ResourceType]int // e.g., {"titanium": 1, "steel": 1} for cards like Phobolog, Advanced Alloys
	eventBus           *events.EventBusImpl
	gameID             string
	playerID           string
}

func newResources(eventBus *events.EventBusImpl, gameID, playerID string) *PlayerResources {
	return &PlayerResources{
		resources:          shared.Resources{},
		production:         shared.Production{},
		terraformRating:    20,
		victoryPoints:      0,
		resourceStorage:    make(map[string]int),
		paymentSubstitutes: []shared.PaymentSubstitute{},
		valueModifiers:     make(map[shared.ResourceType]int),
		eventBus:           eventBus,
		gameID:             gameID,
		playerID:           playerID,
	}
}

func (r *PlayerResources) Get() shared.Resources {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resources
}

func (r *PlayerResources) Production() shared.Production {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.production
}

func (r *PlayerResources) TerraformRating() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.terraformRating
}

func (r *PlayerResources) VictoryPoints() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.victoryPoints
}

func (r *PlayerResources) Storage() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	storageCopy := make(map[string]int, len(r.resourceStorage))
	for k, v := range r.resourceStorage {
		storageCopy[k] = v
	}
	return storageCopy
}

// Base payment values (also defined in cards.SteelValue/TitaniumValue)
const (
	baseSteelValue    = 2
	baseTitaniumValue = 3
)

// PaymentSubstitutes returns all payment substitutes including steel/titanium with dynamic values.
// Steel and titanium are always included with their effective values (base + modifiers from cards like Phobolog).
// Additional substitutes (like heat for Helion) are appended.
func (r *PlayerResources) PaymentSubstitutes() []shared.PaymentSubstitute {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Always include steel and titanium as payment substitutes with dynamic values
	substitutes := []shared.PaymentSubstitute{
		{ResourceType: shared.ResourceSteel, ConversionRate: baseSteelValue + r.valueModifiers[shared.ResourceSteel]},
		{ResourceType: shared.ResourceTitanium, ConversionRate: baseTitaniumValue + r.valueModifiers[shared.ResourceTitanium]},
	}

	// Add any additional substitutes (like heat for Helion)
	substitutes = append(substitutes, r.paymentSubstitutes...)

	return substitutes
}

func (r *PlayerResources) AddPaymentSubstitute(resourceType shared.ResourceType, conversionRate int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paymentSubstitutes = append(r.paymentSubstitutes, shared.PaymentSubstitute{
		ResourceType:   resourceType,
		ConversionRate: conversionRate,
	})
}

// ValueModifiers returns a copy of the value modifiers map
func (r *PlayerResources) ValueModifiers() map[shared.ResourceType]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	modifiersCopy := make(map[shared.ResourceType]int, len(r.valueModifiers))
	for k, v := range r.valueModifiers {
		modifiersCopy[k] = v
	}
	return modifiersCopy
}

// AddValueModifier adds a value modifier for a resource type (e.g., titanium +1 from Phobolog)
func (r *PlayerResources) AddValueModifier(resourceType shared.ResourceType, amount int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.valueModifiers == nil {
		r.valueModifiers = make(map[shared.ResourceType]int)
	}
	r.valueModifiers[resourceType] += amount
}

// GetValueModifier returns the total value modifier for a resource type
func (r *PlayerResources) GetValueModifier(resourceType shared.ResourceType) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.valueModifiers[resourceType]
}

func (r *PlayerResources) Set(resources shared.Resources) {
	r.mu.Lock()
	r.resources = resources
	r.mu.Unlock()

	// Publish event
	if r.eventBus != nil {
		// Publish ResourcesChangedEvent for state synchronization
		// Note: Changes map is empty since this is a full replacement (not a delta)
		events.Publish(r.eventBus, events.ResourcesChangedEvent{
			GameID:    r.gameID,
			PlayerID:  r.playerID,
			Changes:   make(map[string]int),
			Timestamp: time.Now(),
		})
	}
}

func (r *PlayerResources) SetProduction(production shared.Production) {
	// Store old values before modification
	r.mu.Lock()
	oldProduction := r.production
	r.production = production
	newProduction := r.production
	r.mu.Unlock()

	// Publish domain events for each resource type
	if r.eventBus != nil {
		resourceTypes := []struct {
			name     string
			oldValue int
			newValue int
		}{
			{"credits", oldProduction.Credits, newProduction.Credits},
			{"steel", oldProduction.Steel, newProduction.Steel},
			{"titanium", oldProduction.Titanium, newProduction.Titanium},
			{"plants", oldProduction.Plants, newProduction.Plants},
			{"energy", oldProduction.Energy, newProduction.Energy},
			{"heat", oldProduction.Heat, newProduction.Heat},
		}

		for _, rt := range resourceTypes {
			events.Publish(r.eventBus, events.ProductionChangedEvent{
				GameID:        r.gameID,
				PlayerID:      r.playerID,
				ResourceType:  rt.name,
				OldProduction: rt.oldValue,
				NewProduction: rt.newValue,
				Timestamp:     time.Now(),
			})
		}
	}
}

func (r *PlayerResources) SetTerraformRating(tr int) {
	r.mu.Lock()
	oldRating := r.terraformRating
	r.terraformRating = tr
	newRating := r.terraformRating
	r.mu.Unlock()

	// Publish domain event
	if r.eventBus != nil {
		events.Publish(r.eventBus, events.TerraformRatingChangedEvent{
			GameID:    r.gameID,
			PlayerID:  r.playerID,
			OldRating: oldRating,
			NewRating: newRating,
			Timestamp: time.Now(),
		})
	}
}

func (r *PlayerResources) SetVictoryPoints(vp int) {
	r.mu.Lock()
	oldPoints := r.victoryPoints
	r.victoryPoints = vp
	newPoints := r.victoryPoints
	r.mu.Unlock()

	// Publish domain event
	if r.eventBus != nil {
		events.Publish(r.eventBus, events.VictoryPointsChangedEvent{
			GameID:    r.gameID,
			PlayerID:  r.playerID,
			OldPoints: oldPoints,
			NewPoints: newPoints,
			Source:    "direct", // Direct setter without specific source context
			Timestamp: time.Now(),
		})
	}
}

func (r *PlayerResources) Add(changes map[shared.ResourceType]int) {
	r.mu.Lock()
	for resourceType, amount := range changes {
		switch resourceType {
		case shared.ResourceCredit:
			r.resources.Credits += amount
		case shared.ResourceSteel:
			r.resources.Steel += amount
		case shared.ResourceTitanium:
			r.resources.Titanium += amount
		case shared.ResourcePlant:
			r.resources.Plants += amount
		case shared.ResourceEnergy:
			r.resources.Energy += amount
		case shared.ResourceHeat:
			r.resources.Heat += amount
		}
	}
	r.mu.Unlock()

	// Publish domain events
	if r.eventBus != nil {
		// Convert changes map to string keys for event
		changesMap := make(map[string]int, len(changes))
		for resourceType, amount := range changes {
			changesMap[string(resourceType)] = amount
		}

		// Publish ResourcesChangedEvent for passive card effects
		events.Publish(r.eventBus, events.ResourcesChangedEvent{
			GameID:    r.gameID,
			PlayerID:  r.playerID,
			Changes:   changesMap,
			Timestamp: time.Now(),
		})

	}
}

func (r *PlayerResources) AddProduction(changes map[shared.ResourceType]int) {
	// Store old values before modifications
	r.mu.Lock()
	oldProduction := r.production
	for resourceType, amount := range changes {
		switch resourceType {
		case shared.ResourceCreditProduction:
			r.production.Credits += amount
			if r.production.Credits < shared.MinCreditProduction {
				r.production.Credits = shared.MinCreditProduction
			}
		case shared.ResourceSteelProduction:
			r.production.Steel += amount
			if r.production.Steel < shared.MinOtherProduction {
				r.production.Steel = shared.MinOtherProduction
			}
		case shared.ResourceTitaniumProduction:
			r.production.Titanium += amount
			if r.production.Titanium < shared.MinOtherProduction {
				r.production.Titanium = shared.MinOtherProduction
			}
		case shared.ResourcePlantProduction:
			r.production.Plants += amount
			if r.production.Plants < shared.MinOtherProduction {
				r.production.Plants = shared.MinOtherProduction
			}
		case shared.ResourceEnergyProduction:
			r.production.Energy += amount
			if r.production.Energy < shared.MinOtherProduction {
				r.production.Energy = shared.MinOtherProduction
			}
		case shared.ResourceHeatProduction:
			r.production.Heat += amount
			if r.production.Heat < shared.MinOtherProduction {
				r.production.Heat = shared.MinOtherProduction
			}
		}
	}
	newProduction := r.production
	r.mu.Unlock()

	// Publish domain events
	if r.eventBus != nil {
		// Publish ProductionChangedEvent for each resource type that changed
		for resourceType := range changes {
			var oldValue, newValue int
			resourceName := string(resourceType)

			switch resourceType {
			case shared.ResourceCreditProduction:
				oldValue = oldProduction.Credits
				newValue = newProduction.Credits
				resourceName = "credits"
			case shared.ResourceSteelProduction:
				oldValue = oldProduction.Steel
				newValue = newProduction.Steel
				resourceName = "steel"
			case shared.ResourceTitaniumProduction:
				oldValue = oldProduction.Titanium
				newValue = newProduction.Titanium
				resourceName = "titanium"
			case shared.ResourcePlantProduction:
				oldValue = oldProduction.Plants
				newValue = newProduction.Plants
				resourceName = "plants"
			case shared.ResourceEnergyProduction:
				oldValue = oldProduction.Energy
				newValue = newProduction.Energy
				resourceName = "energy"
			case shared.ResourceHeatProduction:
				oldValue = oldProduction.Heat
				newValue = newProduction.Heat
				resourceName = "heat"
			}

			events.Publish(r.eventBus, events.ProductionChangedEvent{
				GameID:        r.gameID,
				PlayerID:      r.playerID,
				ResourceType:  resourceName,
				OldProduction: oldValue,
				NewProduction: newValue,
				Timestamp:     time.Now(),
			})
		}

	}
}

func (r *PlayerResources) UpdateTerraformRating(delta int) {
	r.mu.Lock()
	oldRating := r.terraformRating
	r.terraformRating += delta
	newRating := r.terraformRating
	r.mu.Unlock()

	// Publish domain events
	if r.eventBus != nil {
		// Publish TerraformRatingChangedEvent for passive card effects
		events.Publish(r.eventBus, events.TerraformRatingChangedEvent{
			GameID:    r.gameID,
			PlayerID:  r.playerID,
			OldRating: oldRating,
			NewRating: newRating,
			Timestamp: time.Now(),
		})

	}
}

// AddToStorage adds resources to a specific card's storage
// cardID is the card ID, amount is the number of resources to add
func (r *PlayerResources) AddToStorage(cardID string, amount int) {
	r.mu.Lock()
	if r.resourceStorage == nil {
		r.resourceStorage = make(map[string]int)
	}
	r.resourceStorage[cardID] += amount
	r.mu.Unlock()
}

// GetCardStorage returns the amount of resources stored on a specific card
func (r *PlayerResources) GetCardStorage(cardID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resourceStorage[cardID]
}

// RemoveCardStorage removes the storage entry for a specific card
func (r *PlayerResources) RemoveCardStorage(cardID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.resourceStorage, cardID)
}

// ClearPaymentSubstitutes removes all non-standard payment substitutes
// (keeps steel and titanium which are always available)
func (r *PlayerResources) ClearPaymentSubstitutes() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paymentSubstitutes = []shared.PaymentSubstitute{}
}

// ClearValueModifiers resets all value modifiers to zero
func (r *PlayerResources) ClearValueModifiers() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.valueModifiers = make(map[shared.ResourceType]int)
}
