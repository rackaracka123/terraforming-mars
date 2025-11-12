package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/game/parameters"
	"terraforming-mars-backend/internal/game/resources"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

const (
	// BaseHeatForTemperature is the base cost in heat to raise temperature (before discounts)
	BaseHeatForTemperature = 8
)

// ConvertHeatToTemperatureAction handles converting heat to raise temperature
// This action orchestrates:
// - Resources mechanic (deduct heat)
// - Parameters mechanic (raise temperature, award TR)
type ConvertHeatToTemperatureAction struct {
	playerRepo       repository.PlayerRepository
	gameRepo         repository.GameRepository
	resourcesMech    resources.Service
	parametersMech   parameters.Service
	sessionManager   session.SessionManager
}

// NewConvertHeatToTemperatureAction creates a new heat conversion action
func NewConvertHeatToTemperatureAction(
	playerRepo repository.PlayerRepository,
	gameRepo repository.GameRepository,
	resourcesMech resources.Service,
	parametersMech parameters.Service,
	sessionManager session.SessionManager,
) *ConvertHeatToTemperatureAction {
	return &ConvertHeatToTemperatureAction{
		playerRepo:     playerRepo,
		gameRepo:       gameRepo,
		resourcesMech:  resourcesMech,
		parametersMech: parametersMech,
		sessionManager: sessionManager,
	}
}

// Execute performs the heat to temperature conversion
// Steps:
// 1. Validate player has enough heat (considering discounts)
// 2. Deduct heat via resources mechanic
// 3. Raise temperature via parameters mechanic (if not maxed)
// 4. Award TR via parameters mechanic (if temperature was raised)
// 5. Broadcast state
func (a *ConvertHeatToTemperatureAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🔥 Executing heat to temperature conversion action")

	// Get player to calculate required heat
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Calculate required heat (considering discounts from cards)
	requiredHeat := cards.CalculateResourceConversionCost(&player, model.StandardProjectConvertHeatToTemperature, BaseHeatForTemperature)

	// Validate player has enough heat
	if player.Resources.Heat < requiredHeat {
		log.Warn("Player cannot afford heat conversion",
			zap.Int("required", requiredHeat),
			zap.Int("available", player.Resources.Heat))
		return fmt.Errorf("insufficient heat: need %d, have %d", requiredHeat, player.Resources.Heat)
	}

	// Deduct heat via resources mechanic
	cost := model.ResourceSet{
		Heat: requiredHeat,
	}

	if err := a.resourcesMech.PayResourceCost(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct heat", zap.Error(err))
		return fmt.Errorf("failed to deduct heat: %w", err)
	}

	log.Info("💰 Heat deducted", zap.Int("amount", requiredHeat))

	// Check if temperature can be raised
	game, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	if game.GlobalParameters.Temperature >= model.MaxTemperature {
		log.Info("🌡️ Temperature already at maximum, no TR awarded")

		// Still broadcast state even though no parameter changed
		if err := a.sessionManager.Broadcast(gameID); err != nil {
			log.Error("Failed to broadcast game state", zap.Error(err))
		}

		return nil
	}

	// Raise temperature via parameters mechanic
	// Temperature raises by 1 step (2°C), and TR is awarded automatically
	newTemp, err := a.parametersMech.RaiseTemperature(ctx, gameID, playerID, 1)
	if err != nil {
		log.Error("Failed to raise temperature", zap.Error(err))
		return fmt.Errorf("failed to raise temperature: %w", err)
	}

	log.Info("🌡️ Temperature raised", zap.Int("new_temperature", newTemp))

	// Broadcast updated game state
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Don't fail the action, just log
	}

	log.Info("✅ Heat converted to temperature successfully",
		zap.Int("heat_spent", requiredHeat))

	return nil
}
