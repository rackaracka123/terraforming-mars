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

// CanIncreaseTemperature checks if temperature can be increased by the given steps (business logic from GlobalParameters model)
func (s *GlobalParametersServiceImpl) CanIncreaseTemperature(params *model.GlobalParameters, steps int) bool {
	return steps >= 0 && (params.Temperature < model.MaxTemperature || steps == 0)
}

// CanIncreaseOxygen checks if oxygen can be increased by the given percentage (business logic from GlobalParameters model)
func (s *GlobalParametersServiceImpl) CanIncreaseOxygen(params *model.GlobalParameters, percent int) bool {
	return percent >= 0 && (params.Oxygen < model.MaxOxygen || percent == 0)
}

// CanPlaceOcean checks if oceans can be placed (business logic from GlobalParameters model)
func (s *GlobalParametersServiceImpl) CanPlaceOcean(params *model.GlobalParameters, count int) bool {
	return count >= 0 && (params.Oceans < model.MaxOceans || count == 0)
}

// IsFullyTerraformed checks if all terraforming parameters are at maximum (business logic from GlobalParameters model)
func (s *GlobalParametersServiceImpl) IsFullyTerraformed(params *model.GlobalParameters) bool {
	return params.Temperature == model.MaxTemperature && params.Oxygen == model.MaxOxygen && params.Oceans == model.MaxOceans
}

// GetTerraformingProgress returns the overall terraforming progress as a percentage (business logic from GlobalParameters model)
func (s *GlobalParametersServiceImpl) GetTerraformingProgress(params *model.GlobalParameters) float64 {
	tempProgress := float64(params.Temperature-model.MinTemperature) / float64(model.MaxTemperature-model.MinTemperature)
	oxygenProgress := float64(params.Oxygen-model.MinOxygen) / float64(model.MaxOxygen-model.MinOxygen)
	oceanProgress := float64(params.Oceans-model.MinOceans) / float64(model.MaxOceans-model.MinOceans)

	return (tempProgress + oxygenProgress + oceanProgress) / 3.0 * 100.0
}

// AddResources adds two resource sets together (business logic from Resources model)
func (s *GlobalParametersServiceImpl) AddResources(resources *model.Resources, other model.Resources) model.Resources {
	return model.Resources{
		Credits:  resources.Credits + other.Credits,
		Steel:    resources.Steel + other.Steel,
		Titanium: resources.Titanium + other.Titanium,
		Plants:   resources.Plants + other.Plants,
		Energy:   resources.Energy + other.Energy,
		Heat:     resources.Heat + other.Heat,
	}
}

// SubtractResources subtracts one resource set from another (business logic from Resources model)
func (s *GlobalParametersServiceImpl) SubtractResources(resources *model.Resources, other model.Resources) model.Resources {
	return model.Resources{
		Credits:  resources.Credits - other.Credits,
		Steel:    resources.Steel - other.Steel,
		Titanium: resources.Titanium - other.Titanium,
		Plants:   resources.Plants - other.Plants,
		Energy:   resources.Energy - other.Energy,
		Heat:     resources.Heat - other.Heat,
	}
}

// HasNegativeResources checks if any resource values are negative (business logic from Resources model)
func (s *GlobalParametersServiceImpl) HasNegativeResources(resources *model.Resources) bool {
	return resources.Credits < 0 || resources.Steel < 0 || resources.Titanium < 0 ||
		resources.Plants < 0 || resources.Energy < 0 || resources.Heat < 0
}

// CanAffordResources checks if current resources can afford the given cost (business logic from Resources model)
func (s *GlobalParametersServiceImpl) CanAffordResources(resources *model.Resources, cost model.Resources) bool {
	return resources.Credits >= cost.Credits &&
		resources.Steel >= cost.Steel &&
		resources.Titanium >= cost.Titanium &&
		resources.Plants >= cost.Plants &&
		resources.Energy >= cost.Energy &&
		resources.Heat >= cost.Heat
}

// AddProduction adds two production sets together (business logic from Production model)
func (s *GlobalParametersServiceImpl) AddProduction(production *model.Production, other model.Production) model.Production {
	return model.Production{
		Credits:  production.Credits + other.Credits,
		Steel:    production.Steel + other.Steel,
		Titanium: production.Titanium + other.Titanium,
		Plants:   production.Plants + other.Plants,
		Energy:   production.Energy + other.Energy,
		Heat:     production.Heat + other.Heat,
	}
}

// SubtractProduction subtracts one production set from another (business logic from Production model)
func (s *GlobalParametersServiceImpl) SubtractProduction(production *model.Production, other model.Production) model.Production {
	return model.Production{
		Credits:  production.Credits - other.Credits,
		Steel:    production.Steel - other.Steel,
		Titanium: production.Titanium - other.Titanium,
		Plants:   production.Plants - other.Plants,
		Energy:   production.Energy - other.Energy,
		Heat:     production.Heat - other.Heat,
	}
}
