package game

import (
	"context"
	"time"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
)

// GameGlobalParametersRepository handles global game parameters (temperature, oxygen, oceans, generation)
type GameGlobalParametersRepository struct {
	storage  *GameStorage
	eventBus *events.EventBusImpl
}

// NewGameGlobalParametersRepository creates a new global parameters repository
func NewGameGlobalParametersRepository(storage *GameStorage, eventBus *events.EventBusImpl) *GameGlobalParametersRepository {
	return &GameGlobalParametersRepository{
		storage:  storage,
		eventBus: eventBus,
	}
}

// UpdateTemperature updates the game temperature and publishes event
func (r *GameGlobalParametersRepository) UpdateTemperature(ctx context.Context, gameID string, temperature int) error {
	game, err := r.storage.Get(gameID)
	if err != nil {
		return err
	}

	oldTemp := game.GlobalParameters.Temperature
	game.GlobalParameters.Temperature = temperature

	err = r.storage.Set(gameID, game)
	if err != nil {
		return err
	}

	// Publish TemperatureChangedEvent
	if oldTemp != temperature {
		log := logger.Get().With(
			zap.String("game_id", gameID),
			zap.Int("old_temperature", oldTemp),
			zap.Int("new_temperature", temperature))
		log.Debug("📡 Publishing TemperatureChangedEvent")

		events.Publish(r.eventBus, events.TemperatureChangedEvent{
			GameID:    gameID,
			OldValue:  oldTemp,
			NewValue:  temperature,
			ChangedBy: "", // Can be enhanced to track player who triggered this
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateOxygen updates the game oxygen level and publishes event
func (r *GameGlobalParametersRepository) UpdateOxygen(ctx context.Context, gameID string, oxygen int) error {
	game, err := r.storage.Get(gameID)
	if err != nil {
		return err
	}

	oldOxygen := game.GlobalParameters.Oxygen
	game.GlobalParameters.Oxygen = oxygen

	err = r.storage.Set(gameID, game)
	if err != nil {
		return err
	}

	// Publish OxygenChangedEvent
	if oldOxygen != oxygen {
		log := logger.Get().With(
			zap.String("game_id", gameID),
			zap.Int("old_oxygen", oldOxygen),
			zap.Int("new_oxygen", oxygen))
		log.Debug("📡 Publishing OxygenChangedEvent")

		events.Publish(r.eventBus, events.OxygenChangedEvent{
			GameID:    gameID,
			OldValue:  oldOxygen,
			NewValue:  oxygen,
			ChangedBy: "", // Can be enhanced to track player who triggered this
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateOceans updates the game ocean count and publishes event
func (r *GameGlobalParametersRepository) UpdateOceans(ctx context.Context, gameID string, oceans int) error {
	game, err := r.storage.Get(gameID)
	if err != nil {
		return err
	}

	oldOceans := game.GlobalParameters.Oceans
	game.GlobalParameters.Oceans = oceans

	err = r.storage.Set(gameID, game)
	if err != nil {
		return err
	}

	// Publish OceansChangedEvent
	if oldOceans != oceans {
		log := logger.Get().With(
			zap.String("game_id", gameID),
			zap.Int("old_oceans", oldOceans),
			zap.Int("new_oceans", oceans))
		log.Debug("📡 Publishing OceansChangedEvent")

		events.Publish(r.eventBus, events.OceansChangedEvent{
			GameID:    gameID,
			OldValue:  oldOceans,
			NewValue:  oceans,
			ChangedBy: "", // Can be enhanced to track player who triggered this
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateGeneration updates the game generation counter and publishes event
func (r *GameGlobalParametersRepository) UpdateGeneration(ctx context.Context, gameID string, generation int) error {
	game, err := r.storage.Get(gameID)
	if err != nil {
		return err
	}

	oldGeneration := game.Generation
	game.Generation = generation

	err = r.storage.Set(gameID, game)
	if err != nil {
		return err
	}

	// Publish GenerationAdvancedEvent
	if oldGeneration != generation {
		log := logger.Get().With(
			zap.String("game_id", gameID),
			zap.Int("old_generation", oldGeneration),
			zap.Int("new_generation", generation))
		log.Debug("📡 Publishing GenerationAdvancedEvent")

		events.Publish(r.eventBus, events.GenerationAdvancedEvent{
			GameID:        gameID,
			OldGeneration: oldGeneration,
			NewGeneration: generation,
			Timestamp:     time.Now(),
		})
	}

	return nil
}
