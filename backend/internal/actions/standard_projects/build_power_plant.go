package standard_projects

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

const (
	// PowerPlantCost is the credit cost to build a power plant
	PowerPlantCost = 11
)

// BuildPowerPlantAction handles the build power plant standard project.
// This action orchestrates:
// 1. Validate player can afford 11 credits
// 2. Deduct 11 credits via resources mechanic
// 3. Increase energy production by 1 via resources mechanic
// 4. Broadcast updated game state
type BuildPowerPlantAction struct {
	playerRepo     player.Repository
	gameRepo       game.Repository
	resourcesMech  resources.Service
	sessionManager session.SessionManager
}

// NewBuildPowerPlantAction creates a new build power plant action
func NewBuildPowerPlantAction(
	playerRepo player.Repository,
	gameRepo game.Repository,
	resourcesMech resources.Service,
	sessionManager session.SessionManager,
) *BuildPowerPlantAction {
	return &BuildPowerPlantAction{
		playerRepo:     playerRepo,
		gameRepo:       gameRepo,
		resourcesMech:  resourcesMech,
		sessionManager: sessionManager,
	}
}

// Execute performs the build power plant action
func (a *BuildPowerPlantAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("⚡ Executing build power plant action")

	// 1. Validate player can afford the cost
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	playerResources, err := player.GetResources()
	if err != nil {
		log.Error("Failed to get player resources", zap.Error(err))
		return fmt.Errorf("failed to get player resources: %w", err)
	}

	if playerResources.Credits < PowerPlantCost {
		log.Warn("Player cannot afford power plant",
			zap.Int("cost", PowerPlantCost),
			zap.Int("available", playerResources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", PowerPlantCost, playerResources.Credits)
	}

	// 2. Deduct credits via resources mechanic
	cost := resources.ResourceSet{
		Credits: PowerPlantCost,
	}

	if err := player.ResourcesService.PayCost(ctx, cost); err != nil {
		log.Error("Failed to deduct credits", zap.Error(err))
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	log.Info("💰 Credits deducted", zap.Int("amount", PowerPlantCost))

	// 3. Increase energy production by 1
	production := resources.ResourceSet{
		Energy: 1,
	}

	if err := player.ResourcesService.AddProduction(ctx, production); err != nil {
		log.Error("Failed to increase energy production", zap.Error(err))
		return fmt.Errorf("failed to increase energy production: %w", err)
	}

	log.Info("⚡ Energy production increased", zap.Int("amount", 1))
	log.Info("✅ Build power plant action completed successfully")

	// 4. Broadcast updated game state
	a.sessionManager.Broadcast(gameID)

	return nil
}
