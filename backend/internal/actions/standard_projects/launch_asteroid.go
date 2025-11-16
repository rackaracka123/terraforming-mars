package standard_projects

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// LaunchAsteroidAction handles the launch asteroid standard project.
// This action orchestrates:
// 1. Validate player can afford 14 credits
// 2. Deduct 14 credits
// 3. Raise temperature by 1 step
// 4. Award TR if temperature was raised
// 5. Broadcast updated game state
type LaunchAsteroidAction struct {
	playerRepo        player.Repository
	gameRepo          game.Repository
	parametersService parameters.Service
	sessionManager    session.SessionManager
}

// NewLaunchAsteroidAction creates a new launch asteroid action
func NewLaunchAsteroidAction(
	playerRepo player.Repository,
	gameRepo game.Repository,
	parametersService parameters.Service,
	sessionManager session.SessionManager,
) *LaunchAsteroidAction {
	return &LaunchAsteroidAction{
		playerRepo:        playerRepo,
		gameRepo:          gameRepo,
		parametersService: parametersService,
		sessionManager:    sessionManager,
	}
}

// Execute performs the launch asteroid action
func (a *LaunchAsteroidAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🚀 Executing launch asteroid action")

	// Get cost from domain
	cost := domain.StandardProjectCosts.Asteroid

	// Validate player can afford the cost
	canAfford, err := a.playerRepo.CanAfford(ctx, gameID, playerID, cost)
	if err != nil {
		log.Error("Failed to check affordability", zap.Error(err))
		return fmt.Errorf("failed to check affordability: %w", err)
	}

	if !canAfford {
		log.Warn("Player cannot afford asteroid",
			zap.Int("cost", cost.Credits))
		return fmt.Errorf("insufficient credits: need %d", cost.Credits)
	}

	// Deduct credits
	if err := a.playerRepo.DeductResources(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct credits", zap.Error(err))
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	log.Info("💰 Credits deducted", zap.Int("amount", cost.Credits))

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
		log.Info("🌡️ Temperature raised by asteroid impact",
			zap.Int("new_temperature", newTemperature))
	} else {
		log.Info("🌡️ Temperature already at maximum, no TR awarded")
	}

	// Award TR if temperature was raised
	if temperatureRaised {
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

	log.Info("✅ Launch asteroid action completed successfully")

	// Broadcast updated game state
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Don't fail the operation, just log
	}

	return nil
}
