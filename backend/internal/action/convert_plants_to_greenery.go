package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game/card"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

const (
	// BasePlantsForGreenery is the base cost in plants to convert to greenery (before card discounts)
	BasePlantsForGreenery = 8
)

// ConvertPlantsToGreeneryAction handles the business logic for converting plants to greenery tile
type ConvertPlantsToGreeneryAction struct {
	BaseAction
	gameRepo game.Repository
}

// NewConvertPlantsToGreeneryAction creates a new convert plants to greenery action
func NewConvertPlantsToGreeneryAction(
	gameRepo game.Repository,
	sessionFactory session.SessionFactory,
	sessionMgrFactory session.SessionManagerFactory,
) *ConvertPlantsToGreeneryAction {
	return &ConvertPlantsToGreeneryAction{
		BaseAction: NewBaseAction(sessionFactory, sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the convert plants to greenery action
func (a *ConvertPlantsToGreeneryAction) Execute(ctx context.Context, gameID, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🌱 Converting plants to greenery")

	// 1. Validate game is active
	g, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 3. Get session and player
	sess := a.sessionFactory.Get(gameID)
	if sess == nil {
		log.Error("Game session not found")
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Calculate required plants (with card discount effects)
	requiredPlants := card.CalculateResourceConversionCost(player.Player, types.StandardProjectConvertPlantsToGreenery, BasePlantsForGreenery)
	log.Debug("💰 Calculated plants cost",
		zap.Int("base_cost", BasePlantsForGreenery),
		zap.Int("final_cost", requiredPlants))

	// 5. Validate player has enough plants
	currentResources, err := player.Resources.Get(ctx)
	if err != nil {
		log.Error("Failed to get player resources", zap.Error(err))
		return fmt.Errorf("failed to get resources: %w", err)
	}

	if currentResources.Plants < requiredPlants {
		log.Warn("Player cannot afford plants conversion",
			zap.Int("required", requiredPlants),
			zap.Int("available", currentResources.Plants))
		return fmt.Errorf("insufficient plants: need %d, have %d", requiredPlants, currentResources.Plants)
	}

	// 6. Deduct plants
	newResources := currentResources
	newResources.Plants -= requiredPlants
	err = player.Resources.Update(ctx, newResources)
	if err != nil {
		log.Error("Failed to deduct plants", zap.Error(err))
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("🌿 Deducted plants",
		zap.Int("plants_spent", requiredPlants),
		zap.Int("remaining_plants", newResources.Plants))

	// 7. Create tile queue with "greenery" type
	err = player.TileQueue.CreateQueue(ctx, "convert-plants-to-greenery", []string{"greenery"})
	if err != nil {
		log.Error("Failed to create tile queue", zap.Error(err))
		return fmt.Errorf("failed to create tile queue: %w", err)
	}

	log.Info("📋 Created tile queue for greenery placement")

	// Note: Terraform rating increase and oxygen increase happen when the greenery is placed (via SelectTileAction)

	// 8. Consume action (only if not unlimited actions)
	if player.AvailableActions > 0 {
		newActions := player.AvailableActions - 1
		err = player.Action.UpdateAvailableActions(ctx, newActions)
		if err != nil {
			log.Error("Failed to consume action", zap.Error(err))
			return fmt.Errorf("failed to consume action: %w", err)
		}
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", newActions))
	}

	// 9. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Plants converted successfully, greenery tile queued for placement",
		zap.Int("plants_spent", requiredPlants))
	return nil
}
