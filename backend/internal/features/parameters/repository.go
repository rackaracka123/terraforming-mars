package parameters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
)

// Repository defines granular operations for global parameters storage
type Repository interface {
	// Get current parameters
	Get(ctx context.Context) (GlobalParameters, error)

	// Granular temperature operations
	IncreaseTemperature(ctx context.Context, steps int) (int, error)
	GetTemperature(ctx context.Context) (int, error)
	SetTemperature(ctx context.Context, temperature int) error

	// Granular oxygen operations
	IncreaseOxygen(ctx context.Context, steps int) (int, error)
	GetOxygen(ctx context.Context) (int, error)
	SetOxygen(ctx context.Context, oxygen int) error

	// Granular ocean operations
	IncreaseOceans(ctx context.Context, count int) (int, error)
	GetOceans(ctx context.Context) (int, error)
	SetOceans(ctx context.Context, oceans int) error
}

// RepositoryImpl implements independent in-memory storage for global parameters
type RepositoryImpl struct {
	mu       sync.RWMutex
	gameID   string
	params   GlobalParameters
	eventBus *events.EventBusImpl
}

// NewRepository creates a new independent parameters repository with initial state
func NewRepository(gameID string, initialParams GlobalParameters, eventBus *events.EventBusImpl) (Repository, error) {
	// Validate initial parameters
	if initialParams.Temperature < -30 || initialParams.Temperature > MaxTemperature {
		return nil, fmt.Errorf("invalid temperature: %d", initialParams.Temperature)
	}
	if initialParams.Oxygen < 0 || initialParams.Oxygen > MaxOxygen {
		return nil, fmt.Errorf("invalid oxygen: %d", initialParams.Oxygen)
	}
	if initialParams.Oceans < 0 || initialParams.Oceans > MaxOceans {
		return nil, fmt.Errorf("invalid oceans: %d", initialParams.Oceans)
	}

	return &RepositoryImpl{
		gameID:   gameID,
		params:   initialParams,
		eventBus: eventBus,
	}, nil
}

// Get retrieves current global parameters
func (r *RepositoryImpl) Get(ctx context.Context) (GlobalParameters, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.params, nil
}

// IncreaseTemperature increases temperature by steps (each step = 2°C)
// Returns actual steps increased (capped at maximum)
func (r *RepositoryImpl) IncreaseTemperature(ctx context.Context, steps int) (int, error) {
	if steps <= 0 {
		return 0, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	currentTemp := r.params.Temperature
	newTemp := currentTemp + (steps * 2) // Each step = 2°C

	// Cap at maximum
	actualSteps := steps
	if newTemp > MaxTemperature {
		newTemp = MaxTemperature
		actualSteps = (MaxTemperature - currentTemp) / 2
	}

	if actualSteps <= 0 {
		return 0, nil
	}

	r.params.Temperature = newTemp

	// Publish event if eventBus is available
	if r.eventBus != nil && currentTemp != newTemp {
		events.Publish(r.eventBus, TemperatureChangedEvent{
			GameID:    r.gameID,
			OldValue:  currentTemp,
			NewValue:  newTemp,
			ChangedBy: "", // Will be set by service layer if needed
			Timestamp: time.Now(),
		})
	}

	return actualSteps, nil
}

// GetTemperature retrieves current temperature
func (r *RepositoryImpl) GetTemperature(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.params.Temperature, nil
}

// IncreaseOxygen increases oxygen by steps
// Returns actual steps increased (capped at maximum)
func (r *RepositoryImpl) IncreaseOxygen(ctx context.Context, steps int) (int, error) {
	if steps <= 0 {
		return 0, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	currentOxygen := r.params.Oxygen
	newOxygen := currentOxygen + steps

	// Cap at maximum
	actualSteps := steps
	if newOxygen > MaxOxygen {
		newOxygen = MaxOxygen
		actualSteps = MaxOxygen - currentOxygen
	}

	if actualSteps <= 0 {
		return 0, nil
	}

	r.params.Oxygen = newOxygen

	// Publish event if eventBus is available
	if r.eventBus != nil && currentOxygen != newOxygen {
		events.Publish(r.eventBus, OxygenChangedEvent{
			GameID:    r.gameID,
			OldValue:  currentOxygen,
			NewValue:  newOxygen,
			ChangedBy: "",
			Timestamp: time.Now(),
		})
	}

	return actualSteps, nil
}

// GetOxygen retrieves current oxygen level
func (r *RepositoryImpl) GetOxygen(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.params.Oxygen, nil
}

// IncreaseOceans increases ocean count
// Returns actual count increased (capped at maximum)
func (r *RepositoryImpl) IncreaseOceans(ctx context.Context, count int) (int, error) {
	if count <= 0 {
		return 0, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	currentOceans := r.params.Oceans
	newOceans := currentOceans + count

	// Cap at maximum
	actualCount := count
	if newOceans > MaxOceans {
		newOceans = MaxOceans
		actualCount = MaxOceans - currentOceans
	}

	if actualCount <= 0 {
		return 0, nil
	}

	r.params.Oceans = newOceans

	// Publish event if eventBus is available
	if r.eventBus != nil && currentOceans != newOceans {
		events.Publish(r.eventBus, OceansChangedEvent{
			GameID:    r.gameID,
			OldValue:  currentOceans,
			NewValue:  newOceans,
			ChangedBy: "",
			Timestamp: time.Now(),
		})
	}

	return actualCount, nil
}

// GetOceans retrieves current ocean count
func (r *RepositoryImpl) GetOceans(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.params.Oceans, nil
}

// SetTemperature directly sets temperature to a specific value (for initialization or bulk updates)
func (r *RepositoryImpl) SetTemperature(ctx context.Context, temperature int) error {
	// Validate range
	if temperature < -30 || temperature > MaxTemperature {
		return fmt.Errorf("temperature must be between -30 and %d, got %d", MaxTemperature, temperature)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldTemp := r.params.Temperature
	r.params.Temperature = temperature

	// Publish event if changed and eventBus available
	if r.eventBus != nil && oldTemp != temperature {
		events.Publish(r.eventBus, TemperatureChangedEvent{
			GameID:    r.gameID,
			OldValue:  oldTemp,
			NewValue:  temperature,
			ChangedBy: "",
			Timestamp: time.Now(),
		})
	}

	return nil
}

// SetOxygen directly sets oxygen to a specific value (for initialization or bulk updates)
func (r *RepositoryImpl) SetOxygen(ctx context.Context, oxygen int) error {
	// Validate range
	if oxygen < 0 || oxygen > MaxOxygen {
		return fmt.Errorf("oxygen must be between 0 and %d, got %d", MaxOxygen, oxygen)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldOxygen := r.params.Oxygen
	r.params.Oxygen = oxygen

	// Publish event if changed and eventBus available
	if r.eventBus != nil && oldOxygen != oxygen {
		events.Publish(r.eventBus, OxygenChangedEvent{
			GameID:    r.gameID,
			OldValue:  oldOxygen,
			NewValue:  oxygen,
			ChangedBy: "",
			Timestamp: time.Now(),
		})
	}

	return nil
}

// SetOceans directly sets oceans to a specific value (for initialization or bulk updates)
func (r *RepositoryImpl) SetOceans(ctx context.Context, oceans int) error {
	// Validate range
	if oceans < 0 || oceans > MaxOceans {
		return fmt.Errorf("oceans must be between 0 and %d, got %d", MaxOceans, oceans)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldOceans := r.params.Oceans
	r.params.Oceans = oceans

	// Publish event if changed and eventBus available
	if r.eventBus != nil && oldOceans != oceans {
		events.Publish(r.eventBus, OceansChangedEvent{
			GameID:    r.gameID,
			OldValue:  oldOceans,
			NewValue:  oceans,
			ChangedBy: "",
			Timestamp: time.Now(),
		})
	}

	return nil
}
