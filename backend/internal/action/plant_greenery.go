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
	// PlantGreeneryStandardProjectCost is the megacredit cost to plant greenery via standard project
	PlantGreeneryStandardProjectCost = 23
)

// PlantGreeneryAction handles the business logic for the plant greenery standard project
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type PlantGreeneryAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewPlantGreeneryAction creates a new plant greenery action
func NewPlantGreeneryAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *PlantGreeneryAction {
	return &PlantGreeneryAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the plant greenery action
func (a *PlantGreeneryAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "plant_greenery"),
	)
	log.Info("🌱 Planting greenery (standard project)")

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

	// 5. BUSINESS LOGIC: Validate cost (23 M€)
	resources := player.Resources().Get()
	if resources.Credits < PlantGreeneryStandardProjectCost {
		log.Warn("Insufficient credits for greenery",
			zap.Int("cost", PlantGreeneryStandardProjectCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", PlantGreeneryStandardProjectCost, resources.Credits)
	}

	// 6. BUSINESS LOGIC: Deduct cost using domain method
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: -PlantGreeneryStandardProjectCost,
	})

	resources = player.Resources().Get() // Refresh after update
	log.Info("💰 Deducted greenery cost",
		zap.Int("cost", PlantGreeneryStandardProjectCost),
		zap.Int("remaining_credits", resources.Credits))

	// 7. Create tile queue with "greenery" type on Game (phase state managed by Game)
	queue := &playerPkg.PendingTileSelectionQueue{
		Items:  []string{"greenery"},
		Source: "standard-project-greenery",
	}
	if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		return fmt.Errorf("failed to queue tile placement: %w", err)
	}

	log.Info("📋 Created tile queue for greenery placement")

	// Note: Terraform rating increase happens when the greenery is placed (via SelectTileAction)
	// Note: Oxygen increase happens when greenery is placed (by SelectTileAction)

	// 8. BUSINESS LOGIC: Consume action (only if not unlimited actions)
	availableActions := player.Turn().AvailableActions()
	if availableActions > 0 {
		player.Turn().SetAvailableActions(availableActions - 1)
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions-1))
	}

	// 9. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - g.SetPendingTileSelectionQueue() publishes BroadcastEvent
	//    - player.Resources().Add() publishes ResourcesChangedEvent
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("✅ Greenery queued successfully, tile awaiting placement",
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
