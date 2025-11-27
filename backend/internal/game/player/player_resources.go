package player

import (
	"terraforming-mars-backend/internal/game/shared"
	"sync"

	"terraforming-mars-backend/internal/events"
)

// PlayerResources manages player resources, production, scoring
type PlayerResources struct {
	mu                 sync.RWMutex
	resources shared.Resources
	production shared.Production
	terraformRating    int
	victoryPoints      int
	resourceStorage    map[string]int
	paymentSubstitutes []shared.PaymentSubstitute
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

func (r *PlayerResources) PaymentSubstitutes() []shared.PaymentSubstitute {
	r.mu.RLock()
	defer r.mu.RUnlock()
	substitutesCopy := make([]shared.PaymentSubstitute, len(r.paymentSubstitutes))
	copy(substitutesCopy, r.paymentSubstitutes)
	return substitutesCopy
}

func (r *PlayerResources) Set(resources shared.Resources) {
	r.mu.Lock()
	r.resources = resources
	r.mu.Unlock()

	// Publish event
	if r.eventBus != nil {
		events.Publish(r.eventBus, events.BroadcastEvent{
			GameID:    r.gameID,
			PlayerIDs: []string{r.playerID},
		})
	}
}

func (r *PlayerResources) SetProduction(production shared.Production) {
	r.mu.Lock()
	r.production = production
	r.mu.Unlock()
}

func (r *PlayerResources) SetTerraformRating(tr int) {
	r.mu.Lock()
	r.terraformRating = tr
	r.mu.Unlock()
}

func (r *PlayerResources) SetVictoryPoints(vp int) {
	r.mu.Lock()
	r.victoryPoints = vp
	r.mu.Unlock()
}

func (r *PlayerResources) Add(changes map[shared.ResourceType]int) {
	r.mu.Lock()
	for resourceType, amount := range changes {
		switch resourceType {
		case shared.ResourceCredits:
			r.resources.Credits += amount
		case shared.ResourceSteel:
			r.resources.Steel += amount
		case shared.ResourceTitanium:
			r.resources.Titanium += amount
		case shared.ResourcePlants:
			r.resources.Plants += amount
		case shared.ResourceEnergy:
			r.resources.Energy += amount
		case shared.ResourceHeat:
			r.resources.Heat += amount
		}
	}
	r.mu.Unlock()

	// Publish event
	if r.eventBus != nil {
		events.Publish(r.eventBus, events.BroadcastEvent{
			GameID:    r.gameID,
			PlayerIDs: []string{r.playerID},
		})
	}
}

func (r *PlayerResources) AddProduction(changes map[shared.ResourceType]int) {
	r.mu.Lock()
	for resourceType, amount := range changes {
		switch resourceType {
		case shared.ResourceCreditsProduction:
			r.production.Credits += amount
		case shared.ResourceSteelProduction:
			r.production.Steel += amount
		case shared.ResourceTitaniumProduction:
			r.production.Titanium += amount
		case shared.ResourcePlantsProduction:
			r.production.Plants += amount
		case shared.ResourceEnergyProduction:
			r.production.Energy += amount
		case shared.ResourceHeatProduction:
			r.production.Heat += amount
		}
	}
	r.mu.Unlock()

	// Publish event
	if r.eventBus != nil {
		events.Publish(r.eventBus, events.BroadcastEvent{
			GameID:    r.gameID,
			PlayerIDs: []string{r.playerID},
		})
	}
}

func (r *PlayerResources) UpdateTerraformRating(delta int) {
	r.mu.Lock()
	r.terraformRating += delta
	r.mu.Unlock()

	// Publish event
	if r.eventBus != nil {
		events.Publish(r.eventBus, events.BroadcastEvent{
			GameID:    r.gameID,
			PlayerIDs: []string{r.playerID},
		})
	}
}
