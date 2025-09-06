package service

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// GlobalParametersService handles global terraforming parameters
type GlobalParametersService interface {
	// Update global terraforming parameters
	UpdateGlobalParameters(ctx context.Context, gameID string, newParams model.GlobalParameters) error

	// Get current global parameters
	GetGlobalParameters(ctx context.Context, gameID string) (*model.GlobalParameters, error)

	// Increment temperature by specified steps
	IncreaseTemperature(ctx context.Context, gameID string, steps int) error

	// Increment oxygen by specified steps
	IncreaseOxygen(ctx context.Context, gameID string, steps int) error

	// Place ocean tiles
	PlaceOcean(ctx context.Context, gameID string, count int) error
}

// GlobalParametersServiceImpl implements GlobalParametersService interface
type GlobalParametersServiceImpl struct {
	gameRepo       repository.GameRepository
	parametersRepo repository.GlobalParametersRepository
}

// NewGlobalParametersService creates a new GlobalParametersService instance
func NewGlobalParametersService(gameRepo repository.GameRepository, parametersRepo repository.GlobalParametersRepository) GlobalParametersService {
	return &GlobalParametersServiceImpl{
		gameRepo:       gameRepo,
		parametersRepo: parametersRepo,
	}
}

// UpdateGlobalParameters updates global terraforming parameters
func (s *GlobalParametersServiceImpl) UpdateGlobalParameters(ctx context.Context, gameID string, newParams model.GlobalParameters) error {
	log := logger.WithGameContext(gameID, "")

	log.Info("Updating global parameters via GlobalParametersService",
		zap.Int("temperature", newParams.Temperature),
		zap.Int("oxygen", newParams.Oxygen),
		zap.Int("oceans", newParams.Oceans))

	// Update through ParametersRepository (this will publish events)
	if err := s.parametersRepo.Update(ctx, gameID, &newParams); err != nil {
		log.Error("Failed to update global parameters", zap.Error(err))
		return fmt.Errorf("failed to update parameters: %w", err)
	}

	// Also need to update the game state to keep the main Game entity in sync
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for parameter update", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Update game global parameters
	game.GlobalParameters = newParams

	// Update game state
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game after global parameter change", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Global parameters updated",
		zap.Int("temperature", newParams.Temperature),
		zap.Int("oxygen", newParams.Oxygen),
		zap.Int("oceans", newParams.Oceans))

	return nil
}

// GetGlobalParameters gets current global parameters
func (s *GlobalParametersServiceImpl) GetGlobalParameters(ctx context.Context, gameID string) (*model.GlobalParameters, error) {
	return s.parametersRepo.Get(ctx, gameID)
}

// IncreaseTemperature increases temperature by specified steps
func (s *GlobalParametersServiceImpl) IncreaseTemperature(ctx context.Context, gameID string, steps int) error {
	log := logger.WithGameContext(gameID, "")

	// Get current parameters
	params, err := s.parametersRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get global parameters", zap.Error(err))
		return fmt.Errorf("failed to get parameters: %w", err)
	}

	// Calculate new temperature (max +8°C)
	newTemp := params.Temperature + (steps * 2) // Each step = 2°C
	if newTemp > 8 {
		newTemp = 8
	}

	// Update parameters
	updatedParams := *params
	updatedParams.Temperature = newTemp

	return s.UpdateGlobalParameters(ctx, gameID, updatedParams)
}

// IncreaseOxygen increases oxygen by specified steps
func (s *GlobalParametersServiceImpl) IncreaseOxygen(ctx context.Context, gameID string, steps int) error {
	log := logger.WithGameContext(gameID, "")

	// Get current parameters
	params, err := s.parametersRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get global parameters", zap.Error(err))
		return fmt.Errorf("failed to get parameters: %w", err)
	}

	// Calculate new oxygen level (max 14%)
	newOxygen := params.Oxygen + steps
	if newOxygen > 14 {
		newOxygen = 14
	}

	// Update parameters
	updatedParams := *params
	updatedParams.Oxygen = newOxygen

	return s.UpdateGlobalParameters(ctx, gameID, updatedParams)
}

// PlaceOcean places ocean tiles
func (s *GlobalParametersServiceImpl) PlaceOcean(ctx context.Context, gameID string, count int) error {
	log := logger.WithGameContext(gameID, "")

	// Get current parameters
	params, err := s.parametersRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get global parameters", zap.Error(err))
		return fmt.Errorf("failed to get parameters: %w", err)
	}

	// Calculate new ocean count (max 9 oceans)
	newOceans := params.Oceans + count
	if newOceans > 9 {
		newOceans = 9
	}

	// Update parameters
	updatedParams := *params
	updatedParams.Oceans = newOceans

	return s.UpdateGlobalParameters(ctx, gameID, updatedParams)
}
