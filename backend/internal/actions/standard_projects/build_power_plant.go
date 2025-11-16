package standard_projects

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// BuildPowerPlantAction handles the build power plant standard project.
// This action orchestrates:
// 1. Validate player can afford 11 credits
// 2. Deduct 11 credits
// 3. Increase energy production by 1
// 4. Broadcast updated game state
type BuildPowerPlantAction struct {
	playerRepo     player.Repository
	gameRepo       game.Repository
	sessionManager session.SessionManager
}

// NewBuildPowerPlantAction creates a new build power plant action
func NewBuildPowerPlantAction(
	playerRepo player.Repository,
	gameRepo game.Repository,
	sessionManager session.SessionManager,
) *BuildPowerPlantAction {
	return &BuildPowerPlantAction{
		playerRepo:     playerRepo,
		gameRepo:       gameRepo,
		sessionManager: sessionManager,
	}
}

// Execute performs the build power plant action
func (a *BuildPowerPlantAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("⚡ Executing build power plant action")

	// Get cost from domain
	cost := domain.StandardProjectCosts.PowerPlant

	// Validate player can afford the cost
	canAfford, err := a.playerRepo.CanAfford(ctx, gameID, playerID, cost)
	if err != nil {
		log.Error("Failed to check affordability", zap.Error(err))
		return fmt.Errorf("failed to check affordability: %w", err)
	}

	if !canAfford {
		log.Warn("Player cannot afford power plant",
			zap.Int("cost", cost.Credits))
		return fmt.Errorf("insufficient credits: need %d", cost.Credits)
	}

	// Deduct credits
	if err := a.playerRepo.DeductResources(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct credits", zap.Error(err))
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	log.Info("💰 Credits deducted", zap.Int("amount", cost.Credits))

	// Increase energy production by 1
	production := domain.ResourceSet{
		Energy: 1,
	}

	if err := a.playerRepo.AddProduction(ctx, gameID, playerID, production); err != nil {
		log.Error("Failed to increase energy production", zap.Error(err))
		return fmt.Errorf("failed to increase energy production: %w", err)
	}

	log.Info("⚡ Energy production increased", zap.Int("amount", 1))
	log.Info("✅ Build power plant action completed successfully")

	// Broadcast updated game state
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Don't fail the operation, just log
	}

	return nil
}
