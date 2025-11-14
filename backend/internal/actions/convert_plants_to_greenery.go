package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

const (
	// BasePlantsForGreenery is the base cost in plants to place greenery (before discounts)
	BasePlantsForGreenery = 8
)

// ConvertPlantsToGreeneryAction handles converting plants to place greenery
// This action orchestrates:
// - Resources mechanic (deduct plants)
// - Parameters mechanic (raise oxygen, award TR)
// - Tiles mechanic (create tile queue, process placement)
type ConvertPlantsToGreeneryAction struct {
	playerRepo     player.Repository
	gameRepo       game.Repository
	resourcesMech  resources.Service
	parametersMech parameters.Service
	tilesMech      tiles.Service
	sessionManager session.SessionManager
}

// NewConvertPlantsToGreeneryAction creates a new plants conversion action
func NewConvertPlantsToGreeneryAction(
	playerRepo player.Repository,
	gameRepo game.Repository,
	resourcesMech resources.Service,
	parametersMech parameters.Service,
	tilesMech tiles.Service,
	sessionManager session.SessionManager,
) *ConvertPlantsToGreeneryAction {
	return &ConvertPlantsToGreeneryAction{
		playerRepo:     playerRepo,
		gameRepo:       gameRepo,
		resourcesMech:  resourcesMech,
		parametersMech: parametersMech,
		tilesMech:      tilesMech,
		sessionManager: sessionManager,
	}
}

// Execute performs the plants to greenery conversion
// Steps:
// 1. Validate player has enough plants (considering discounts)
// 2. Deduct plants via resources mechanic
// 3. Raise oxygen via parameters mechanic (if not maxed)
// 4. Award TR via parameters mechanic (if oxygen was raised)
// 5. Create tile queue for greenery placement
// 6. Process tile queue to prepare tile selection
// 7. Broadcast state
func (a *ConvertPlantsToGreeneryAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🌱 Executing convert plants to greenery action")

	// Get player to calculate required plants
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Calculate required plants (considering discounts from cards)
	requiredPlants := cards.CalculateResourceConversionCost(&player, types.StandardProjectConvertPlantsToGreenery, BasePlantsForGreenery)

	// Validate player has enough plants
	if player.Resources.Plants < requiredPlants {
		log.Warn("Player cannot afford plants conversion",
			zap.Int("required", requiredPlants),
			zap.Int("available", player.Resources.Plants))
		return fmt.Errorf("insufficient plants: need %d, have %d", requiredPlants, player.Resources.Plants)
	}

	// Deduct plants via resources mechanic
	cost := resources.ResourceSet{
		Plants: requiredPlants,
	}

	if err := a.resourcesMech.PayResourceCost(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct plants", zap.Error(err))
		return fmt.Errorf("failed to deduct plants: %w", err)
	}

	log.Info("💰 Plants deducted", zap.Int("amount", requiredPlants))

	// Check if oxygen can be raised
	isMaxed, err := a.parametersMech.IsOxygenMaxed(ctx, gameID)
	if err != nil {
		log.Error("Failed to check oxygen max", zap.Error(err))
		return fmt.Errorf("failed to check oxygen: %w", err)
	}

	// Raise oxygen by 1 step (awards TR automatically in parameters mechanic)
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

	// Create tile queue for greenery placement
	queueSource := "convert-plants-to-greenery"
	if err := a.playerRepo.CreateTileQueue(ctx, gameID, playerID, queueSource, []string{"greenery"}); err != nil {
		log.Error("Failed to create tile queue", zap.Error(err))
		return fmt.Errorf("failed to create tile queue: %w", err)
	}

	// Process the tile queue to create pending tile selection
	if err := a.tilesMech.ProcessTileQueue(ctx, gameID, playerID); err != nil {
		log.Error("Failed to process tile queue", zap.Error(err))
		return fmt.Errorf("failed to process tile queue: %w", err)
	}

	// Broadcast updated game state
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Don't fail the action, just log
	}

	log.Info("✅ Plants converted to greenery successfully",
		zap.Int("plants_spent", requiredPlants))

	return nil
}
