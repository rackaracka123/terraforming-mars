package standard_projects

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

const (
	// CityCost is the credit cost to build a city
	CityCost = 25
)

// BuildCityAction handles the build city standard project.
// This action orchestrates:
// 1. Validate player can afford 25 credits
// 2. Deduct 25 credits via resources mechanic
// 3. Increase credit production by 1 via resources mechanic
// 4. Create tile queue for city placement via tiles mechanic
// 5. Process tile queue to prepare tile selection
// 6. Broadcast updated game state
type BuildCityAction struct {
	playerRepo     player.Repository
	gameRepo       game.Repository
	resourcesMech  resources.Service
	tilesMech      tiles.Service
	sessionManager session.SessionManager
}

// NewBuildCityAction creates a new build city action
func NewBuildCityAction(
	playerRepo player.Repository,
	gameRepo game.Repository,
	resourcesMech resources.Service,
	tilesMech tiles.Service,
	sessionManager session.SessionManager,
) *BuildCityAction {
	return &BuildCityAction{
		playerRepo:     playerRepo,
		gameRepo:       gameRepo,
		resourcesMech:  resourcesMech,
		tilesMech:      tilesMech,
		sessionManager: sessionManager,
	}
}

// Execute performs the build city action
func (a *BuildCityAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🏙️ Executing build city action")

	// 1. Validate player can afford the cost
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	if player.Resources.Credits < CityCost {
		log.Warn("Player cannot afford city",
			zap.Int("cost", CityCost),
			zap.Int("available", player.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", CityCost, player.Resources.Credits)
	}

	// 2. Deduct credits via resources mechanic
	cost := resources.ResourceSet{
		Credits: CityCost,
	}

	if err := a.resourcesMech.PayResourceCost(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct credits", zap.Error(err))
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	log.Info("💰 Credits deducted", zap.Int("amount", CityCost))

	// 3. Increase credit production by 1 (cities provide income)
	production := resources.ResourceSet{
		Credits: 1,
	}

	if err := a.resourcesMech.AddProduction(ctx, gameID, playerID, production); err != nil {
		log.Error("Failed to increase credit production", zap.Error(err))
		return fmt.Errorf("failed to increase credit production: %w", err)
	}

	log.Info("💰 Credit production increased", zap.Int("amount", 1))

	// 4. Create tile queue for city placement
	queueSource := "standard-project-city"
	if err := a.playerRepo.CreateTileQueue(ctx, gameID, playerID, queueSource, []string{"city"}); err != nil {
		log.Error("Failed to create tile queue", zap.Error(err))
		return fmt.Errorf("failed to create tile queue: %w", err)
	}

	log.Info("📋 Tile queue created for city placement")

	// 5. Process tile queue to prepare tile selection
	if err := a.tilesMech.ProcessTileQueue(ctx, gameID, playerID); err != nil {
		log.Error("Failed to process tile queue", zap.Error(err))
		return fmt.Errorf("failed to process tile queue: %w", err)
	}

	log.Info("✅ Build city action completed successfully")

	// 6. Broadcast updated game state
	a.sessionManager.Broadcast(gameID)

	return nil
}
