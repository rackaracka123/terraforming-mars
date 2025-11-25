package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

const (
	// PlantGreeneryStandardProjectCost is the megacredit cost to plant greenery via standard project
	PlantGreeneryStandardProjectCost = 23
)

// PlantGreeneryAction handles the business logic for the plant greenery standard project
type PlantGreeneryAction struct {
	BaseAction
	gameRepo game.Repository
}

// NewPlantGreeneryAction creates a new plant greenery action
func NewPlantGreeneryAction(
	gameRepo game.Repository,
	sessionFactory session.SessionFactory,
	sessionMgrFactory session.SessionManagerFactory,
) *PlantGreeneryAction {
	return &PlantGreeneryAction{
		BaseAction: NewBaseAction(sessionFactory, sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the plant greenery action
func (a *PlantGreeneryAction) Execute(ctx context.Context, gameID, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🌱 Planting greenery (standard project)")

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

	// 4. Validate cost (23 M€)
	currentResources, err := player.Resources.Get(ctx)
	if err != nil {
		log.Error("Failed to get player resources", zap.Error(err))
		return fmt.Errorf("failed to get resources: %w", err)
	}

	if currentResources.Credits < PlantGreeneryStandardProjectCost {
		log.Warn("Insufficient credits for greenery",
			zap.Int("cost", PlantGreeneryStandardProjectCost),
			zap.Int("player_credits", currentResources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", PlantGreeneryStandardProjectCost, currentResources.Credits)
	}

	// 5. Deduct cost
	newResources := currentResources
	newResources.Credits -= PlantGreeneryStandardProjectCost
	err = player.Resources.Update(ctx, newResources)
	if err != nil {
		log.Error("Failed to deduct greenery cost", zap.Error(err))
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("💰 Deducted greenery cost",
		zap.Int("cost", PlantGreeneryStandardProjectCost),
		zap.Int("remaining_credits", newResources.Credits))

	// 6. Create tile queue with "greenery" type
	err = player.TileQueue.CreateQueue(ctx, "standard-project-greenery", []string{"greenery"})
	if err != nil {
		log.Error("Failed to create tile queue", zap.Error(err))
		return fmt.Errorf("failed to create tile queue: %w", err)
	}

	log.Info("📋 Created tile queue for greenery placement")

	// Note: Terraform rating increase happens when the greenery is placed (via SelectTileAction)

	// 7. Consume action (only if not unlimited actions)
	if player.AvailableActions > 0 {
		newActions := player.AvailableActions - 1
		err = player.Action.UpdateAvailableActions(ctx, newActions)
		if err != nil {
			log.Error("Failed to consume action", zap.Error(err))
			return fmt.Errorf("failed to consume action: %w", err)
		}
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", newActions))
	}

	// 8. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Greenery queued successfully, tile awaiting placement",
		zap.Int("remaining_credits", newResources.Credits))
	return nil
}
