package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"
)

// ConvertHeatToTemperatureAction handles converting heat to raise temperature
// This action orchestrates:
// 1. Validation and heat deduction
// 2. Temperature raising
// 3. TR award if temperature increased
type ConvertHeatToTemperatureAction struct {
	playerRepo        player.Repository
	parametersService parameters.Service
	sessionManager    session.SessionManager
}

// NewConvertHeatToTemperatureAction creates a new heat conversion action
func NewConvertHeatToTemperatureAction(
	playerRepo player.Repository,
	parametersService parameters.Service,
	sessionManager session.SessionManager,
) *ConvertHeatToTemperatureAction {
	return &ConvertHeatToTemperatureAction{
		playerRepo:        playerRepo,
		parametersService: parametersService,
		sessionManager:    sessionManager,
	}
}

// Execute performs the heat to temperature conversion
// Steps:
// 1. Get player and resources
// 2. Validate player has enough heat
// 3. Deduct heat from player
// 4. Raise temperature by 1 step (+2°C) if not maxed
// 5. Award TR if temperature was raised
// 6. Broadcast game state
func (a *ConvertHeatToTemperatureAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🔥 Converting heat to temperature")

	// Use the standard project cost from domain
	cost := domain.StandardProjectCosts.ConvertHeatToTemperature

	// Validate player can afford the cost
	canAfford, err := a.playerRepo.CanAfford(ctx, gameID, playerID, cost)
	if err != nil {
		log.Error("Failed to check affordability", zap.Error(err))
		return fmt.Errorf("failed to check affordability: %w", err)
	}

	if !canAfford {
		log.Warn("Player cannot afford heat conversion",
			zap.Int("required_heat", cost.Heat))
		return fmt.Errorf("insufficient heat: need %d", cost.Heat)
	}

	// Deduct heat
	if err := a.playerRepo.DeductResources(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct heat", zap.Error(err))
		return fmt.Errorf("failed to deduct heat: %w", err)
	}

	// Raise temperature by 1 step (+2°C) if not already maxed
	globalParams, err := a.parametersService.GetGlobalParameters(ctx)
	if err != nil {
		log.Error("Failed to get global parameters", zap.Error(err))
		return fmt.Errorf("failed to get global parameters: %w", err)
	}

	temperatureRaised := false
	if globalParams.Temperature < parameters.MaxTemperature {
		// Raise temperature by 1 step (service handles the 2°C increment and max checks)
		newTemperature, err := a.parametersService.RaiseTemperature(ctx, 1)
		if err != nil {
			log.Error("Failed to raise temperature", zap.Error(err))
			return fmt.Errorf("failed to raise temperature: %w", err)
		}

		temperatureRaised = true
		log.Info("🌡️ Temperature raised",
			zap.Int("new_temperature", newTemperature))
	} else {
		log.Info("🌡️ Temperature already at maximum, no TR awarded")
	}

	// Award TR if temperature was raised
	if temperatureRaised {
		// Get current player to update TR
		currentPlayer, err := a.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			log.Error("Failed to get player for TR update", zap.Error(err))
			return fmt.Errorf("failed to get player: %w", err)
		}

		newTR := currentPlayer.TerraformRating + 1
		if err := a.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
			log.Error("Failed to update terraform rating", zap.Error(err))
			return fmt.Errorf("failed to update terraform rating: %w", err)
		}

		log.Info("⭐ Terraform rating increased",
			zap.Int("new_tr", newTR))
	}

	log.Info("✅ Heat converted to temperature successfully",
		zap.Int("heat_spent", cost.Heat))

	return nil
}
