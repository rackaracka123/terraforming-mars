package parameters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ============================================================================
// MODELS
// ============================================================================

// GlobalParameters represents the terraforming progress
type GlobalParameters struct {
	Temperature int `json:"temperature" ts:"number"` // Range: -30 to +8°C
	Oxygen      int `json:"oxygen" ts:"number"`      // Range: 0-14%
	Oceans      int `json:"oceans" ts:"number"`      // Range: 0-9
}

// Constants for terraforming limits
const (
	MinTemperature = -30
	MaxTemperature = 8
	MinOxygen      = 0
	MaxOxygen      = 14
	MinOceans      = 0
	MaxOceans      = 9
)

// ============================================================================
// REPOSITORY
// ============================================================================

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
		events.Publish(r.eventBus, events.TemperatureChangedEvent{
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
		events.Publish(r.eventBus, events.OxygenChangedEvent{
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
		events.Publish(r.eventBus, events.OceansChangedEvent{
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
		events.Publish(r.eventBus, events.TemperatureChangedEvent{
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
		events.Publish(r.eventBus, events.OxygenChangedEvent{
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
		events.Publish(r.eventBus, events.OceansChangedEvent{
			GameID:    r.gameID,
			OldValue:  oldOceans,
			NewValue:  oceans,
			ChangedBy: "",
			Timestamp: time.Now(),
		})
	}

	return nil
}

// ============================================================================
// SERVICE
// ============================================================================

// Service handles global parameter operations (temperature, oxygen, oceans).
//
// Scope: Isolated global parameter management
//   - Temperature updates with validation (-30 to +8)
//   - Oxygen updates with validation (0-14%)
//   - Ocean updates with validation (0-9)
//   - Parameter max value enforcement
//
// This feature is ISOLATED and should NOT:
//   - Award TR (handled via events by TR feature)
//   - Call other feature services
//   - Handle tile placement (tiles feature does that)
//   - Manage turn state or phases
//
// Dependencies:
//   - Own repository (independent storage)
//   - EventBus (for publishing parameter change events) [TODO: Phase 5]
type Service interface {
	// Temperature operations
	RaiseTemperature(ctx context.Context, steps int) (int, error)
	GetTemperature(ctx context.Context) (int, error)
	IsTemperatureMaxed(ctx context.Context) (bool, error)
	SetTemperature(ctx context.Context, temperature int) error

	// Oxygen operations
	RaiseOxygen(ctx context.Context, steps int) (int, error)
	GetOxygen(ctx context.Context) (int, error)
	IsOxygenMaxed(ctx context.Context) (bool, error)
	SetOxygen(ctx context.Context, oxygen int) error

	// Ocean operations
	PlaceOcean(ctx context.Context) error
	GetOceans(ctx context.Context) (int, error)
	IsOceansMaxed(ctx context.Context) (bool, error)
	SetOceans(ctx context.Context, oceans int) error

	// Combined operations
	GetGlobalParameters(ctx context.Context) (GlobalParameters, error)
}

// ServiceImpl implements the Global Parameters service
type ServiceImpl struct {
	repo Repository
}

// NewService creates a new Global Parameters service
func NewService(repo Repository) Service {
	return &ServiceImpl{
		repo: repo,
	}
}

// RaiseTemperature raises the global temperature by the specified number of steps.
// Returns the actual number of steps raised (may be less if max is reached).
//
// Note: TR is awarded via events (TemperatureIncreasedEvent)
func (s *ServiceImpl) RaiseTemperature(ctx context.Context, steps int) (int, error) {
	if steps <= 0 {
		return 0, nil
	}

	// Get current temperature for logging
	currentTemp, err := s.repo.GetTemperature(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get current temperature: %w", err)
	}

	// Increase temperature (repo handles capping)
	actualSteps, err := s.repo.IncreaseTemperature(ctx, steps)
	if err != nil {
		return 0, fmt.Errorf("failed to increase temperature: %w", err)
	}

	if actualSteps <= 0 {
		logger.Get().Debug("Temperature already at maximum")
		return 0, nil
	}

	// Get new temperature for logging
	newTemp, _ := s.repo.GetTemperature(ctx)

	logger.Get().Info("🌡️ Temperature raised",
		zap.Int("from", currentTemp),
		zap.Int("to", newTemp),
		zap.Int("steps", actualSteps))

	// TODO Phase 5: Publish TemperatureIncreasedEvent for TR award

	return actualSteps, nil
}

// GetTemperature retrieves the current global temperature
func (s *ServiceImpl) GetTemperature(ctx context.Context) (int, error) {
	temp, err := s.repo.GetTemperature(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get temperature: %w", err)
	}
	return temp, nil
}

// IsTemperatureMaxed checks if temperature is at maximum
func (s *ServiceImpl) IsTemperatureMaxed(ctx context.Context) (bool, error) {
	temp, err := s.GetTemperature(ctx)
	if err != nil {
		return false, err
	}
	return temp >= MaxTemperature, nil
}

// RaiseOxygen raises the global oxygen by the specified number of steps.
// Returns the actual number of steps raised (may be less if max is reached).
//
// Note: TR is awarded via events (OxygenIncreasedEvent)
func (s *ServiceImpl) RaiseOxygen(ctx context.Context, steps int) (int, error) {
	if steps <= 0 {
		return 0, nil
	}

	// Get current oxygen for logging
	currentOxygen, err := s.repo.GetOxygen(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get current oxygen: %w", err)
	}

	// Increase oxygen (repo handles capping)
	actualSteps, err := s.repo.IncreaseOxygen(ctx, steps)
	if err != nil {
		return 0, fmt.Errorf("failed to increase oxygen: %w", err)
	}

	if actualSteps <= 0 {
		logger.Get().Debug("Oxygen already at maximum")
		return 0, nil
	}

	// Get new oxygen for logging
	newOxygen, _ := s.repo.GetOxygen(ctx)

	logger.Get().Info("💨 Oxygen raised",
		zap.Int("from", currentOxygen),
		zap.Int("to", newOxygen),
		zap.Int("steps", actualSteps))

	// TODO Phase 5: Publish OxygenIncreasedEvent for TR award

	return actualSteps, nil
}

// GetOxygen retrieves the current global oxygen level
func (s *ServiceImpl) GetOxygen(ctx context.Context) (int, error) {
	oxygen, err := s.repo.GetOxygen(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get oxygen: %w", err)
	}
	return oxygen, nil
}

// IsOxygenMaxed checks if oxygen is at maximum
func (s *ServiceImpl) IsOxygenMaxed(ctx context.Context) (bool, error) {
	oxygen, err := s.GetOxygen(ctx)
	if err != nil {
		return false, err
	}
	return oxygen >= MaxOxygen, nil
}

// PlaceOcean increments the ocean count by 1.
//
// Note: This method only handles the ocean COUNT, not tile placement (that's tiles feature).
// TR is awarded via events (OceanPlacedEvent).
func (s *ServiceImpl) PlaceOcean(ctx context.Context) error {
	// Get current ocean count for logging
	currentOceans, err := s.repo.GetOceans(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current oceans: %w", err)
	}

	// Check if already at maximum
	if currentOceans >= MaxOceans {
		logger.Get().Debug("Oceans already at maximum")
		return fmt.Errorf("maximum oceans already placed")
	}

	// Increase oceans
	actualCount, err := s.repo.IncreaseOceans(ctx, 1)
	if err != nil {
		return fmt.Errorf("failed to increase oceans: %w", err)
	}

	if actualCount <= 0 {
		return fmt.Errorf("failed to place ocean: already at maximum")
	}

	// Get new ocean count for logging
	newOceans, _ := s.repo.GetOceans(ctx)

	logger.Get().Info("🌊 Ocean placed",
		zap.Int("from", currentOceans),
		zap.Int("to", newOceans))

	// TODO Phase 5: Publish OceanPlacedEvent for TR award

	return nil
}

// GetOceans retrieves the current ocean count
func (s *ServiceImpl) GetOceans(ctx context.Context) (int, error) {
	oceans, err := s.repo.GetOceans(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get oceans: %w", err)
	}
	return oceans, nil
}

// IsOceansMaxed checks if oceans are at maximum
func (s *ServiceImpl) IsOceansMaxed(ctx context.Context) (bool, error) {
	oceans, err := s.GetOceans(ctx)
	if err != nil {
		return false, err
	}
	return oceans >= MaxOceans, nil
}

// GetGlobalParameters retrieves all global parameters
func (s *ServiceImpl) GetGlobalParameters(ctx context.Context) (GlobalParameters, error) {
	params, err := s.repo.Get(ctx)
	if err != nil {
		return GlobalParameters{}, fmt.Errorf("failed to get parameters: %w", err)
	}
	return params, nil
}

// SetTemperature sets the temperature to an absolute value (for admin/testing)
func (s *ServiceImpl) SetTemperature(ctx context.Context, temperature int) error {
	return s.repo.SetTemperature(ctx, temperature)
}

// SetOxygen sets the oxygen to an absolute value (for admin/testing)
func (s *ServiceImpl) SetOxygen(ctx context.Context, oxygen int) error {
	return s.repo.SetOxygen(ctx, oxygen)
}

// SetOceans sets the oceans to an absolute value (for admin/testing)
func (s *ServiceImpl) SetOceans(ctx context.Context, oceans int) error {
	return s.repo.SetOceans(ctx, oceans)
}
