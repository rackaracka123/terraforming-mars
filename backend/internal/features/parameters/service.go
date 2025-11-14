package parameters

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

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

	// Oxygen operations
	RaiseOxygen(ctx context.Context, steps int) (int, error)
	GetOxygen(ctx context.Context) (int, error)
	IsOxygenMaxed(ctx context.Context) (bool, error)

	// Ocean operations
	PlaceOcean(ctx context.Context) error
	GetOceans(ctx context.Context) (int, error)
	IsOceansMaxed(ctx context.Context) (bool, error)

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
