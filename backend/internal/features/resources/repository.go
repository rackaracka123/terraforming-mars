package resources

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
)

// Repository defines granular operations for resource storage
type Repository interface {
	// Get operations
	Get(ctx context.Context) (model.Resources, error)
	GetProduction(ctx context.Context) (model.Production, error)

	// Set operations for bulk updates
	Set(ctx context.Context, resources model.Resources) error
	SetProduction(ctx context.Context, production model.Production) error

	// Granular resource operations
	AddCredits(ctx context.Context, amount int) error
	DeductCredits(ctx context.Context, amount int) error
	GetCredits(ctx context.Context) (int, error)

	AddSteel(ctx context.Context, amount int) error
	DeductSteel(ctx context.Context, amount int) error
	GetSteel(ctx context.Context) (int, error)

	AddTitanium(ctx context.Context, amount int) error
	DeductTitanium(ctx context.Context, amount int) error
	GetTitanium(ctx context.Context) (int, error)

	AddPlants(ctx context.Context, amount int) error
	DeductPlants(ctx context.Context, amount int) error
	GetPlants(ctx context.Context) (int, error)

	AddEnergy(ctx context.Context, amount int) error
	DeductEnergy(ctx context.Context, amount int) error
	GetEnergy(ctx context.Context) (int, error)

	AddHeat(ctx context.Context, amount int) error
	DeductHeat(ctx context.Context, amount int) error
	GetHeat(ctx context.Context) (int, error)

	// Special energy-to-heat conversion
	ConvertEnergyToHeat(ctx context.Context) error

	// Production operations
	IncreaseCreditsProduction(ctx context.Context, amount int) error
	IncreaseSteelProduction(ctx context.Context, amount int) error
	IncreaseTitaniumProduction(ctx context.Context, amount int) error
	IncreasePlantsProduction(ctx context.Context, amount int) error
	IncreaseEnergyProduction(ctx context.Context, amount int) error
	IncreaseHeatProduction(ctx context.Context, amount int) error

	DecreaseCreditsProduction(ctx context.Context, amount int) error
	DecreaseSteelProduction(ctx context.Context, amount int) error
	DecreaseTitaniumProduction(ctx context.Context, amount int) error
	DecreasePlantsProduction(ctx context.Context, amount int) error
	DecreaseEnergyProduction(ctx context.Context, amount int) error
	DecreaseHeatProduction(ctx context.Context, amount int) error
}

// RepositoryImpl implements independent in-memory storage for resources
type RepositoryImpl struct {
	mu         sync.RWMutex
	gameID     string
	playerID   string
	resources  model.Resources
	production model.Production
	eventBus   *events.EventBusImpl
}

// NewRepository creates a new independent resources repository with initial state
func NewRepository(gameID, playerID string, initialResources model.Resources, initialProduction model.Production, eventBus *events.EventBusImpl) Repository {
	return &RepositoryImpl{
		gameID:     gameID,
		playerID:   playerID,
		resources:  initialResources,
		production: initialProduction,
		eventBus:   eventBus,
	}
}

// Get retrieves current resources
func (r *RepositoryImpl) Get(ctx context.Context) (model.Resources, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resources, nil
}

// GetProduction retrieves current production
func (r *RepositoryImpl) GetProduction(ctx context.Context) (model.Production, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.production, nil
}

// Set directly sets resources to specific values (for bulk updates)
func (r *RepositoryImpl) Set(ctx context.Context, resources model.Resources) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store old values for potential event publishing
	oldResources := r.resources
	r.resources = resources

	// Publish events for changed resources if eventBus is available
	if r.eventBus != nil {
		if oldResources.Credits != resources.Credits {
			events.Publish(r.eventBus, events.ResourcesChangedEvent{
				GameID:       r.gameID,
				PlayerID:     r.playerID,
				ResourceType: "credits",
				OldAmount:    oldResources.Credits,
				NewAmount:    resources.Credits,
				Timestamp:    time.Now(),
			})
		}
		if oldResources.Steel != resources.Steel {
			events.Publish(r.eventBus, events.ResourcesChangedEvent{
				GameID:       r.gameID,
				PlayerID:     r.playerID,
				ResourceType: "steel",
				OldAmount:    oldResources.Steel,
				NewAmount:    resources.Steel,
				Timestamp:    time.Now(),
			})
		}
		if oldResources.Titanium != resources.Titanium {
			events.Publish(r.eventBus, events.ResourcesChangedEvent{
				GameID:       r.gameID,
				PlayerID:     r.playerID,
				ResourceType: "titanium",
				OldAmount:    oldResources.Titanium,
				NewAmount:    resources.Titanium,
				Timestamp:    time.Now(),
			})
		}
		if oldResources.Plants != resources.Plants {
			events.Publish(r.eventBus, events.ResourcesChangedEvent{
				GameID:       r.gameID,
				PlayerID:     r.playerID,
				ResourceType: "plants",
				OldAmount:    oldResources.Plants,
				NewAmount:    resources.Plants,
				Timestamp:    time.Now(),
			})
		}
		if oldResources.Energy != resources.Energy {
			events.Publish(r.eventBus, events.ResourcesChangedEvent{
				GameID:       r.gameID,
				PlayerID:     r.playerID,
				ResourceType: "energy",
				OldAmount:    oldResources.Energy,
				NewAmount:    resources.Energy,
				Timestamp:    time.Now(),
			})
		}
		if oldResources.Heat != resources.Heat {
			events.Publish(r.eventBus, events.ResourcesChangedEvent{
				GameID:       r.gameID,
				PlayerID:     r.playerID,
				ResourceType: "heat",
				OldAmount:    oldResources.Heat,
				NewAmount:    resources.Heat,
				Timestamp:    time.Now(),
			})
		}
	}

	return nil
}

// SetProduction directly sets production to specific values (for bulk updates)
func (r *RepositoryImpl) SetProduction(ctx context.Context, production model.Production) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.production = production
	return nil
}

// Credits operations
func (r *RepositoryImpl) AddCredits(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	oldAmount := r.resources.Credits
	r.resources.Credits += amount

	// Publish event if eventBus is available and amount changed
	if r.eventBus != nil && amount > 0 {
		events.Publish(r.eventBus, events.ResourcesChangedEvent{
			GameID:       r.gameID,
			PlayerID:     r.playerID,
			ResourceType: "credits",
			OldAmount:    oldAmount,
			NewAmount:    r.resources.Credits,
			Timestamp:    time.Now(),
		})
	}

	return nil
}

func (r *RepositoryImpl) DeductCredits(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.resources.Credits < amount {
		return fmt.Errorf("insufficient credits: have %d, need %d", r.resources.Credits, amount)
	}

	oldAmount := r.resources.Credits
	r.resources.Credits -= amount

	// Publish event if eventBus is available and amount changed
	if r.eventBus != nil && amount > 0 {
		events.Publish(r.eventBus, events.ResourcesChangedEvent{
			GameID:       r.gameID,
			PlayerID:     r.playerID,
			ResourceType: "credits",
			OldAmount:    oldAmount,
			NewAmount:    r.resources.Credits,
			Timestamp:    time.Now(),
		})
	}

	return nil
}

func (r *RepositoryImpl) GetCredits(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resources.Credits, nil
}

// Steel operations
func (r *RepositoryImpl) AddSteel(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources.Steel += amount
	return nil
}

func (r *RepositoryImpl) DeductSteel(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.resources.Steel < amount {
		return fmt.Errorf("insufficient steel: have %d, need %d", r.resources.Steel, amount)
	}
	r.resources.Steel -= amount
	return nil
}

func (r *RepositoryImpl) GetSteel(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resources.Steel, nil
}

// Titanium operations
func (r *RepositoryImpl) AddTitanium(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources.Titanium += amount
	return nil
}

func (r *RepositoryImpl) DeductTitanium(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.resources.Titanium < amount {
		return fmt.Errorf("insufficient titanium: have %d, need %d", r.resources.Titanium, amount)
	}
	r.resources.Titanium -= amount
	return nil
}

func (r *RepositoryImpl) GetTitanium(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resources.Titanium, nil
}

// Plants operations
func (r *RepositoryImpl) AddPlants(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources.Plants += amount
	return nil
}

func (r *RepositoryImpl) DeductPlants(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.resources.Plants < amount {
		return fmt.Errorf("insufficient plants: have %d, need %d", r.resources.Plants, amount)
	}
	r.resources.Plants -= amount
	return nil
}

func (r *RepositoryImpl) GetPlants(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resources.Plants, nil
}

// Energy operations
func (r *RepositoryImpl) AddEnergy(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources.Energy += amount
	return nil
}

func (r *RepositoryImpl) DeductEnergy(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.resources.Energy < amount {
		return fmt.Errorf("insufficient energy: have %d, need %d", r.resources.Energy, amount)
	}
	r.resources.Energy -= amount
	return nil
}

func (r *RepositoryImpl) GetEnergy(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resources.Energy, nil
}

// Heat operations
func (r *RepositoryImpl) AddHeat(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources.Heat += amount
	return nil
}

func (r *RepositoryImpl) DeductHeat(ctx context.Context, amount int) error {
	if amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.resources.Heat < amount {
		return fmt.Errorf("insufficient heat: have %d, need %d", r.resources.Heat, amount)
	}
	r.resources.Heat -= amount
	return nil
}

func (r *RepositoryImpl) GetHeat(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resources.Heat, nil
}

// ConvertEnergyToHeat converts all energy to heat (production phase)
func (r *RepositoryImpl) ConvertEnergyToHeat(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources.Heat += r.resources.Energy
	r.resources.Energy = 0
	return nil
}

// Production increase operations
func (r *RepositoryImpl) IncreaseCreditsProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.production.Credits += amount
	return nil
}

func (r *RepositoryImpl) IncreaseSteelProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.production.Steel += amount
	return nil
}

func (r *RepositoryImpl) IncreaseTitaniumProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.production.Titanium += amount
	return nil
}

func (r *RepositoryImpl) IncreasePlantsProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.production.Plants += amount
	return nil
}

func (r *RepositoryImpl) IncreaseEnergyProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.production.Energy += amount
	return nil
}

func (r *RepositoryImpl) IncreaseHeatProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.production.Heat += amount
	return nil
}

// Production decrease operations
func (r *RepositoryImpl) DecreaseCreditsProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.production.Credits < amount {
		return fmt.Errorf("cannot decrease credits production below 0")
	}
	r.production.Credits -= amount
	return nil
}

func (r *RepositoryImpl) DecreaseSteelProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.production.Steel < amount {
		return fmt.Errorf("cannot decrease steel production below 0")
	}
	r.production.Steel -= amount
	return nil
}

func (r *RepositoryImpl) DecreaseTitaniumProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.production.Titanium < amount {
		return fmt.Errorf("cannot decrease titanium production below 0")
	}
	r.production.Titanium -= amount
	return nil
}

func (r *RepositoryImpl) DecreasePlantsProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.production.Plants < amount {
		return fmt.Errorf("cannot decrease plants production below 0")
	}
	r.production.Plants -= amount
	return nil
}

func (r *RepositoryImpl) DecreaseEnergyProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.production.Energy < amount {
		return fmt.Errorf("cannot decrease energy production below 0")
	}
	r.production.Energy -= amount
	return nil
}

func (r *RepositoryImpl) DecreaseHeatProduction(ctx context.Context, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.production.Heat < amount {
		return fmt.Errorf("cannot decrease heat production below 0")
	}
	r.production.Heat -= amount
	return nil
}
