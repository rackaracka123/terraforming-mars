package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

const (
	// BasePlantsForGreenery is the base cost in plants to convert to greenery (before card discounts)
	BasePlantsForGreenery = 8
)

// ConvertPlantsToGreeneryAction handles the business logic for converting plants to greenery tile
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
// NOTE: Still uses old gamePackage.CalculateResourceConversionCost until card effects fully migrated
type ConvertPlantsToGreeneryAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewConvertPlantsToGreeneryAction creates a new convert plants to greenery action
func NewConvertPlantsToGreeneryAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *ConvertPlantsToGreeneryAction {
	return &ConvertPlantsToGreeneryAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the convert plants to greenery action
func (a *ConvertPlantsToGreeneryAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "convert_plants_to_greenery"),
	)
	log.Info("🌱 Converting plants to greenery")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. BUSINESS LOGIC: Validate game is active
	if g.Status() != game.GameStatusActive {
		log.Warn("Game is not active", zap.String("status", string(g.Status())))
		return fmt.Errorf("game is not active: %s", g.Status())
	}

	// 3. BUSINESS LOGIC: Validate it's the player's turn
	currentTurn := g.CurrentTurn()
	if currentTurn == nil || currentTurn.PlayerID() != playerID {
		var turnPlayerID string
		if currentTurn != nil {
			turnPlayerID = currentTurn.PlayerID()
		}
		log.Warn("Not player's turn",
			zap.String("current_turn_player", turnPlayerID),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("not your turn")
	}

	// 4. Get player from game
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 5. BUSINESS LOGIC: Calculate required plants (with card discount effects)
	// TODO: Reimplement card discount effects when card system is migrated
	requiredPlants := BasePlantsForGreenery
	log.Debug("💰 Calculated plants cost",
		zap.Int("base_cost", BasePlantsForGreenery),
		zap.Int("final_cost", requiredPlants))

	// 6. BUSINESS LOGIC: Validate player has enough plants
	resources := player.Resources().Get()
	if resources.Plants < requiredPlants {
		log.Warn("Player cannot afford plants conversion",
			zap.Int("required", requiredPlants),
			zap.Int("available", resources.Plants))
		return fmt.Errorf("insufficient plants: need %d, have %d", requiredPlants, resources.Plants)
	}

	// 7. BUSINESS LOGIC: Deduct plants using domain method
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlants: -requiredPlants,
	})

	resources = player.Resources().Get() // Refresh after update
	log.Info("🌿 Deducted plants",
		zap.Int("plants_spent", requiredPlants),
		zap.Int("remaining_plants", resources.Plants))

	// 8. Create tile queue with "greenery" type on Game (phase state managed by Game)
	queue := &playerPkg.PendingTileSelectionQueue{
		Items:  []string{"greenery"},
		Source: "convert-plants-to-greenery",
	}
	if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		return fmt.Errorf("failed to queue tile placement: %w", err)
	}

	log.Info("📋 Created tile queue for greenery placement")

	// Note: Terraform rating increase and oxygen increase happen when the greenery is placed (via SelectTileAction)

	// 9. BUSINESS LOGIC: Consume action (only if not unlimited actions)
	availableActions := player.Turn().AvailableActions()
	if availableActions > 0 {
		player.Turn().SetAvailableActions(availableActions - 1)
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions-1))
	}

	// 10. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - player.Resources().Add() publishes ResourcesChangedEvent
	//    - g.SetPendingTileSelectionQueue() publishes BroadcastEvent
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("✅ Plants converted successfully, greenery tile queued for placement",
		zap.Int("plants_spent", requiredPlants))
	return nil
}
