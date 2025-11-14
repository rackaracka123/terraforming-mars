package standard_projects

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

const (
	// GreeeneryCost is the credit cost to plant a greenery
	GreeneryCost = 23
)

// PlantGreeneryAction handles the plant greenery standard project.
// This action orchestrates:
// 1. Validate player can afford 23 credits
// 2. Deduct 23 credits via resources mechanic
// 3. Raise oxygen by 1 step via parameters mechanic (awards TR automatically)
// 4. Create tile queue for greenery placement via tiles mechanic
// 5. Process tile queue to prepare tile selection
// 6. Broadcast updated game state
type PlantGreeneryAction struct {
	playerRepo     player.Repository
	gameRepo       game.Repository
	resourcesMech  resources.Service
	parametersMech parameters.Service
	tilesMech      tiles.Service
	sessionManager session.SessionManager
}

// NewPlantGreeneryAction creates a new plant greenery action
func NewPlantGreeneryAction(
	playerRepo player.Repository,
	gameRepo game.Repository,
	resourcesMech resources.Service,
	parametersMech parameters.Service,
	tilesMech tiles.Service,
	sessionManager session.SessionManager,
) *PlantGreeneryAction {
	return &PlantGreeneryAction{
		playerRepo:     playerRepo,
		gameRepo:       gameRepo,
		resourcesMech:  resourcesMech,
		parametersMech: parametersMech,
		tilesMech:      tilesMech,
		sessionManager: sessionManager,
	}
}

// Execute performs the plant greenery action
func (a *PlantGreeneryAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🌱 Executing plant greenery action")

	// 1. Validate player can afford the cost
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	if player.Resources.Credits < GreeneryCost {
		log.Warn("Player cannot afford greenery",
			zap.Int("cost", GreeneryCost),
			zap.Int("available", player.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", GreeneryCost, player.Resources.Credits)
	}

	// 2. Deduct credits via resources mechanic
	cost := resources.ResourceSet{
		Credits: GreeneryCost,
	}

	if err := a.resourcesMech.PayResourceCost(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct credits", zap.Error(err))
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	log.Info("💰 Credits deducted", zap.Int("amount", GreeneryCost))

	// 3. Check if oxygen can be raised
	isMaxed, err := a.parametersMech.IsOxygenMaxed(ctx, gameID)
	if err != nil {
		log.Error("Failed to check oxygen max", zap.Error(err))
		return fmt.Errorf("failed to check oxygen: %w", err)
	}

	// 4. Raise oxygen by 1 step (awards TR automatically in parameters mechanic)
	if !isMaxed {
		stepsRaised, err := a.parametersMech.RaiseOxygen(ctx, gameID, playerID, 1)
		if err != nil {
			log.Error("Failed to raise oxygen", zap.Error(err))
			return fmt.Errorf("failed to raise oxygen: %w", err)
		}

		if stepsRaised > 0 {
			log.Info("🌿 Oxygen raised",
				zap.Int("steps", stepsRaised),
				zap.String("effect", "Greenery photosynthesis"))
		} else {
			log.Info("🌿 Oxygen already at maximum")
		}
	} else {
		log.Info("🌿 Oxygen already at maximum, no increase")
	}

	// 5. Create tile queue for greenery placement
	queueSource := "standard-project-greenery"
	if err := a.playerRepo.CreateTileQueue(ctx, gameID, playerID, queueSource, []string{"greenery"}); err != nil {
		log.Error("Failed to create tile queue", zap.Error(err))
		return fmt.Errorf("failed to create tile queue: %w", err)
	}

	log.Info("📋 Tile queue created for greenery placement")

	// 6. Process tile queue to prepare tile selection
	if err := a.tilesMech.ProcessTileQueue(ctx, gameID, playerID); err != nil {
		log.Error("Failed to process tile queue", zap.Error(err))
		return fmt.Errorf("failed to process tile queue: %w", err)
	}

	log.Info("✅ Plant greenery action completed successfully")

	// 7. Broadcast updated game state
	a.sessionManager.Broadcast(gameID)

	return nil
}
