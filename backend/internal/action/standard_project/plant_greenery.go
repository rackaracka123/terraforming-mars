package standard_project

import (
	baseaction "terraforming-mars-backend/internal/action"
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
	baseaction.BaseAction
}

// NewPlantGreeneryAction creates a new plant greenery action
func NewPlantGreeneryAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *PlantGreeneryAction {
	return &PlantGreeneryAction{
		BaseAction: baseaction.NewBaseAction(gameRepo, nil),
	}
}

// Execute performs the plant greenery action
func (a *PlantGreeneryAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "plant_greenery"))
	log.Info("🌱 Planting greenery (standard project)")

	// 1. Fetch game from repository and validate it's active
	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 3. Get player from game
	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
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

	log.Info("📋 Created tile queue for greenery placement (auto-processed by SetPendingTileSelectionQueue)")

	// Note: Terraform rating increase happens when the greenery is placed (via SelectTileAction)
	// Note: Oxygen increase happens when greenery is placed (by SelectTileAction)

	// 9. Consume action (only if not unlimited actions)
	a.ConsumePlayerAction(g, log)

	log.Info("✅ Greenery tile selection ready",
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
