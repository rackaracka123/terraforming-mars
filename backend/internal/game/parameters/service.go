package parameters

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// Service handles global parameter operations (temperature, oxygen, oceans).
//
// Scope: Isolated global parameter management mechanic
//   - Temperature updates with validation (-30 to +8)
//   - Oxygen updates with validation (0-14%)
//   - Ocean updates with validation (0-9)
//   - Parameter max value enforcement
//   - TR award calculations
//
// This mechanic is ISOLATED and should NOT:
//   - Call other mechanic services
//   - Handle tile placement (tiles mechanic does that)
//   - Manage turn state or phases
//
// Dependencies:
//   - GameRepository (for reading/updating global parameters)
//   - PlayerRepository (for awarding TR)
type Service interface {
	// Temperature operations
	RaiseTemperature(ctx context.Context, gameID, playerID string, steps int) (int, error)
	GetTemperature(ctx context.Context, gameID string) (int, error)
	IsTemperatureMaxed(ctx context.Context, gameID string) (bool, error)

	// Oxygen operations
	RaiseOxygen(ctx context.Context, gameID, playerID string, steps int) (int, error)
	GetOxygen(ctx context.Context, gameID string) (int, error)
	IsOxygenMaxed(ctx context.Context, gameID string) (bool, error)

	// Ocean operations
	PlaceOcean(ctx context.Context, gameID, playerID string) error
	GetOceans(ctx context.Context, gameID string) (int, error)
	IsOceansMaxed(ctx context.Context, gameID string) (bool, error)

	// Combined operations
	GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error)
}

// ServiceImpl implements the Global Parameters mechanic service
type ServiceImpl struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

// NewService creates a new Global Parameters mechanic service
func NewService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) Service {
	return &ServiceImpl{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
	}
}

// RaiseTemperature raises the global temperature by the specified number of steps.
// Returns the actual number of steps raised (may be less if max is reached).
// Awards TR to the player for each step raised.
func (s *ServiceImpl) RaiseTemperature(ctx context.Context, gameID, playerID string, steps int) (int, error) {
	log := logger.WithGameContext(gameID, playerID)

	if steps <= 0 {
		return 0, nil
	}

	// Get current temperature
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return 0, fmt.Errorf("failed to get game: %w", err)
	}

	currentTemp := game.GlobalParameters.Temperature
	newTemp := currentTemp + (steps * 2) // Each step raises temperature by 2°C

	// Cap at maximum
	actualSteps := steps
	if newTemp > model.MaxTemperature {
		newTemp = model.MaxTemperature
		actualSteps = (model.MaxTemperature - currentTemp) / 2
	}

	if actualSteps <= 0 {
		log.Debug("Temperature already at maximum")
		return 0, nil
	}

	// Update temperature
	if err := s.gameRepo.UpdateTemperature(ctx, gameID, newTemp); err != nil {
		log.Error("Failed to update temperature", zap.Error(err))
		return 0, fmt.Errorf("failed to update temperature: %w", err)
	}

	// Award TR for each step
	if playerID != "" && actualSteps > 0 {
		player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			log.Error("Failed to get player for TR award", zap.Error(err))
			return actualSteps, fmt.Errorf("failed to get player: %w", err)
		}

		newTR := player.TerraformRating + actualSteps
		if err := s.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
			log.Error("Failed to award TR for temperature raise", zap.Error(err))
			return actualSteps, fmt.Errorf("failed to award TR: %w", err)
		}

		log.Info("🌡️ Temperature raised",
			zap.Int("from", currentTemp),
			zap.Int("to", newTemp),
			zap.Int("steps", actualSteps),
			zap.Int("tr_awarded", actualSteps))
	}

	return actualSteps, nil
}

// GetTemperature retrieves the current global temperature
func (s *ServiceImpl) GetTemperature(ctx context.Context, gameID string) (int, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return 0, fmt.Errorf("failed to get game: %w", err)
	}

	return game.GlobalParameters.Temperature, nil
}

// IsTemperatureMaxed checks if temperature is at maximum
func (s *ServiceImpl) IsTemperatureMaxed(ctx context.Context, gameID string) (bool, error) {
	temp, err := s.GetTemperature(ctx, gameID)
	if err != nil {
		return false, err
	}

	return temp >= model.MaxTemperature, nil
}

// RaiseOxygen raises the global oxygen by the specified number of steps.
// Returns the actual number of steps raised (may be less if max is reached).
// Awards TR to the player for each step raised.
func (s *ServiceImpl) RaiseOxygen(ctx context.Context, gameID, playerID string, steps int) (int, error) {
	log := logger.WithGameContext(gameID, playerID)

	if steps <= 0 {
		return 0, nil
	}

	// Get current oxygen
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return 0, fmt.Errorf("failed to get game: %w", err)
	}

	currentOxygen := game.GlobalParameters.Oxygen
	newOxygen := currentOxygen + steps

	// Cap at maximum
	actualSteps := steps
	if newOxygen > model.MaxOxygen {
		newOxygen = model.MaxOxygen
		actualSteps = model.MaxOxygen - currentOxygen
	}

	if actualSteps <= 0 {
		log.Debug("Oxygen already at maximum")
		return 0, nil
	}

	// Update oxygen
	if err := s.gameRepo.UpdateOxygen(ctx, gameID, newOxygen); err != nil {
		log.Error("Failed to update oxygen", zap.Error(err))
		return 0, fmt.Errorf("failed to update oxygen: %w", err)
	}

	// Award TR for each step
	if playerID != "" && actualSteps > 0 {
		player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			log.Error("Failed to get player for TR award", zap.Error(err))
			return actualSteps, fmt.Errorf("failed to get player: %w", err)
		}

		newTR := player.TerraformRating + actualSteps
		if err := s.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
			log.Error("Failed to award TR for oxygen raise", zap.Error(err))
			return actualSteps, fmt.Errorf("failed to award TR: %w", err)
		}

		log.Info("💨 Oxygen raised",
			zap.Int("from", currentOxygen),
			zap.Int("to", newOxygen),
			zap.Int("steps", actualSteps),
			zap.Int("tr_awarded", actualSteps))
	}

	return actualSteps, nil
}

// GetOxygen retrieves the current global oxygen level
func (s *ServiceImpl) GetOxygen(ctx context.Context, gameID string) (int, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return 0, fmt.Errorf("failed to get game: %w", err)
	}

	return game.GlobalParameters.Oxygen, nil
}

// IsOxygenMaxed checks if oxygen is at maximum
func (s *ServiceImpl) IsOxygenMaxed(ctx context.Context, gameID string) (bool, error) {
	oxygen, err := s.GetOxygen(ctx, gameID)
	if err != nil {
		return false, err
	}

	return oxygen >= model.MaxOxygen, nil
}

// PlaceOcean places an ocean tile and increments the ocean count.
// Awards TR to the player.
// Note: This method only handles the ocean COUNT, not tile placement (that's tiles mechanic).
func (s *ServiceImpl) PlaceOcean(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current ocean count
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	currentOceans := game.GlobalParameters.Oceans

	// Check if already at maximum
	if currentOceans >= model.MaxOceans {
		log.Debug("Oceans already at maximum")
		return fmt.Errorf("maximum oceans already placed")
	}

	// Increment ocean count
	newOceans := currentOceans + 1
	if err := s.gameRepo.UpdateOceans(ctx, gameID, newOceans); err != nil {
		log.Error("Failed to update ocean count", zap.Error(err))
		return fmt.Errorf("failed to update oceans: %w", err)
	}

	// Award TR (1 TR per ocean)
	if playerID != "" {
		player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			log.Error("Failed to get player for TR award", zap.Error(err))
			return fmt.Errorf("failed to get player: %w", err)
		}

		newTR := player.TerraformRating + 1
		if err := s.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
			log.Error("Failed to award TR for ocean placement", zap.Error(err))
			return fmt.Errorf("failed to award TR: %w", err)
		}

		log.Info("🌊 Ocean placed",
			zap.Int("from", currentOceans),
			zap.Int("to", newOceans),
			zap.Int("tr_awarded", 1))
	}

	return nil
}

// GetOceans retrieves the current ocean count
func (s *ServiceImpl) GetOceans(ctx context.Context, gameID string) (int, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return 0, fmt.Errorf("failed to get game: %w", err)
	}

	return game.GlobalParameters.Oceans, nil
}

// IsOceansMaxed checks if oceans are at maximum
func (s *ServiceImpl) IsOceansMaxed(ctx context.Context, gameID string) (bool, error) {
	oceans, err := s.GetOceans(ctx, gameID)
	if err != nil {
		return false, err
	}

	return oceans >= model.MaxOceans, nil
}

// GetGlobalParameters retrieves all global parameters
func (s *ServiceImpl) GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return model.GlobalParameters{}, fmt.Errorf("failed to get game: %w", err)
	}

	return game.GlobalParameters, nil
}
