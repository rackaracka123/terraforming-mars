package repository

import (
	"context"
	"fmt"
	"sync"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"

	"go.uber.org/zap"
)

// GlobalParametersRepository manages terraforming parameters
type GlobalParametersRepository interface {
	// Get global parameters for a game
	Get(ctx context.Context, gameID string) (*model.GlobalParameters, error)

	// Update global parameters
	Update(ctx context.Context, gameID string, params *model.GlobalParameters) error
}

// GlobalParametersRepositoryImpl implements GlobalParametersRepository interface
type GlobalParametersRepositoryImpl struct {
	// Map of gameID -> GlobalParameters
	parameters map[string]*model.GlobalParameters
	mutex      sync.RWMutex
	eventBus   events.EventBus
}

// NewGlobalParametersRepository creates a new global parameters repository
func NewGlobalParametersRepository(eventBus events.EventBus) GlobalParametersRepository {
	return &GlobalParametersRepositoryImpl{
		parameters: make(map[string]*model.GlobalParameters),
		eventBus:   eventBus,
	}
}

// Get retrieves global parameters for a game
func (r *GlobalParametersRepositoryImpl) Get(ctx context.Context, gameID string) (*model.GlobalParameters, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	params, exists := r.parameters[gameID]
	if !exists {
		// Return default parameters if none exist
		defaultParams := &model.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		}
		return defaultParams, nil
	}

	return params, nil
}

// Update updates global parameters for a game
func (r *GlobalParametersRepositoryImpl) Update(ctx context.Context, gameID string, params *model.GlobalParameters) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	if gameID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	if params == nil {
		return fmt.Errorf("global parameters cannot be nil")
	}

	// Validate parameter ranges
	if err := r.validateParameters(params); err != nil {
		log.Error("Invalid global parameters", zap.Error(err))
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Capture old state for events
	var oldParams *model.GlobalParameters
	if existing, exists := r.parameters[gameID]; exists {
		oldParams = &model.GlobalParameters{
			Temperature: existing.Temperature,
			Oxygen:      existing.Oxygen,
			Oceans:      existing.Oceans,
		}
	} else {
		oldParams = &model.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		}
	}

	// Update parameters
	r.parameters[gameID] = &model.GlobalParameters{
		Temperature: params.Temperature,
		Oxygen:      params.Oxygen,
		Oceans:      params.Oceans,
	}

	log.Info("Global parameters updated",
		zap.Int("temperature", params.Temperature),
		zap.Int("oxygen", params.Oxygen),
		zap.Int("oceans", params.Oceans),
	)

	// Publish events for each parameter that changed
	if r.eventBus != nil {
		if oldParams.Temperature != params.Temperature {
			tempEvent := events.NewTemperatureChangedEvent(gameID, oldParams.Temperature, params.Temperature)
			if err := r.eventBus.Publish(ctx, tempEvent); err != nil {
				log.Warn("Failed to publish temperature changed event", zap.Error(err))
			}
		}

		if oldParams.Oxygen != params.Oxygen {
			oxygenEvent := events.NewOxygenChangedEvent(gameID, oldParams.Oxygen, params.Oxygen)
			if err := r.eventBus.Publish(ctx, oxygenEvent); err != nil {
				log.Warn("Failed to publish oxygen changed event", zap.Error(err))
			}
		}

		if oldParams.Oceans != params.Oceans {
			oceansEvent := events.NewOceansChangedEvent(gameID, oldParams.Oceans, params.Oceans)
			if err := r.eventBus.Publish(ctx, oceansEvent); err != nil {
				log.Warn("Failed to publish oceans changed event", zap.Error(err))
			}
		}

		// Also publish a general parameters changed event
		parametersChangedEvent := events.NewGlobalParametersChangedEvent(gameID, *oldParams, *params)
		if err := r.eventBus.Publish(ctx, parametersChangedEvent); err != nil {
			log.Warn("Failed to publish global parameters changed event", zap.Error(err))
		}
	}

	return nil
}

// validateParameters ensures parameters are within valid game ranges
func (r *GlobalParametersRepositoryImpl) validateParameters(params *model.GlobalParameters) error {
	if params.Temperature < -30 || params.Temperature > 8 {
		return fmt.Errorf("temperature must be between -30 and 8, got %d", params.Temperature)
	}

	if params.Oxygen < 0 || params.Oxygen > 14 {
		return fmt.Errorf("oxygen must be between 0 and 14, got %d", params.Oxygen)
	}

	if params.Oceans < 0 || params.Oceans > 9 {
		return fmt.Errorf("oceans must be between 0 and 9, got %d", params.Oceans)
	}

	return nil
}
