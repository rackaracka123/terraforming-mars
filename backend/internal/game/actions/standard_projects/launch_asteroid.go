package standard_projects

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/game/parameters"
	"terraforming-mars-backend/internal/game/resources"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

const (
	// AsteroidCost is the credit cost to launch an asteroid
	AsteroidCost = 14
)

// LaunchAsteroidAction handles the launch asteroid standard project.
// This action orchestrates:
// 1. Validate player can afford 14 credits
// 2. Deduct 14 credits via resources mechanic
// 3. Raise temperature by 1 step via parameters mechanic
// 4. Award TR via parameters mechanic
// 5. Broadcast updated game state
type LaunchAsteroidAction struct {
	playerRepo     repository.PlayerRepository
	gameRepo       repository.GameRepository
	resourcesMech  resources.Service
	parametersMech parameters.Service
	sessionManager session.SessionManager
}

// NewLaunchAsteroidAction creates a new launch asteroid action
func NewLaunchAsteroidAction(
	playerRepo repository.PlayerRepository,
	gameRepo repository.GameRepository,
	resourcesMech resources.Service,
	parametersMech parameters.Service,
	sessionManager session.SessionManager,
) *LaunchAsteroidAction {
	return &LaunchAsteroidAction{
		playerRepo:     playerRepo,
		gameRepo:       gameRepo,
		resourcesMech:  resourcesMech,
		parametersMech: parametersMech,
		sessionManager: sessionManager,
	}
}

// Execute performs the launch asteroid action
func (a *LaunchAsteroidAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🚀 Executing launch asteroid action")

	// 1. Validate player can afford the cost
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	if player.Resources.Credits < AsteroidCost {
		log.Warn("Player cannot afford asteroid",
			zap.Int("cost", AsteroidCost),
			zap.Int("available", player.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", AsteroidCost, player.Resources.Credits)
	}

	// 2. Deduct credits via resources mechanic
	cost := resources.ResourceSet{
		Credits: AsteroidCost,
	}

	if err := a.resourcesMech.PayResourceCost(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct credits", zap.Error(err))
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	log.Info("💰 Credits deducted", zap.Int("amount", AsteroidCost))

	// 3. Check if temperature can be raised
	isMaxed, err := a.parametersMech.IsTemperatureMaxed(ctx, gameID)
	if err != nil {
		log.Error("Failed to check temperature max", zap.Error(err))
		return fmt.Errorf("failed to check temperature: %w", err)
	}

	// 4. Raise temperature by 1 step (awards TR automatically in parameters mechanic)
	if !isMaxed {
		stepsRaised, err := a.parametersMech.RaiseTemperature(ctx, gameID, playerID, 1)
		if err != nil {
			log.Error("Failed to raise temperature", zap.Error(err))
			return fmt.Errorf("failed to raise temperature: %w", err)
		}

		if stepsRaised > 0 {
			log.Info("🌡️ Temperature raised",
				zap.Int("steps", stepsRaised),
				zap.String("effect", "Asteroid impact"))
		} else {
			log.Info("🌡️ Temperature already at maximum")
		}
	} else {
		log.Info("🌡️ Temperature already at maximum, no increase")
	}

	log.Info("✅ Launch asteroid action completed successfully")

	// 5. Broadcast updated game state
	a.sessionManager.Broadcast(gameID)

	return nil
}
