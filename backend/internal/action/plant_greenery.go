package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

const (
	// PlantGreeneryStandardProjectCost is the megacredit cost to plant greenery via standard project
	PlantGreeneryStandardProjectCost = 23
)

// PlantGreeneryAction handles the business logic for the plant greenery standard project
type PlantGreeneryAction struct {
	BaseAction
}

// NewPlantGreeneryAction creates a new plant greenery action
func NewPlantGreeneryAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
) *PlantGreeneryAction {
	return &PlantGreeneryAction{
		BaseAction: NewBaseAction(gameRepo, playerRepo, sessionMgr),
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

	// 3. Validate player exists
	p, err := ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
	if err != nil {
		return err
	}

	// 4. Validate cost (23 M€)
	if p.Resources.Credits < PlantGreeneryStandardProjectCost {
		log.Warn("Insufficient credits for greenery",
			zap.Int("cost", PlantGreeneryStandardProjectCost),
			zap.Int("player_credits", p.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", PlantGreeneryStandardProjectCost, p.Resources.Credits)
	}

	// 5. Deduct cost
	newResources := p.Resources
	newResources.Credits -= PlantGreeneryStandardProjectCost
	err = a.playerRepo.UpdateResources(ctx, gameID, playerID, newResources)
	if err != nil {
		log.Error("Failed to deduct greenery cost", zap.Error(err))
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("💰 Deducted greenery cost",
		zap.Int("cost", PlantGreeneryStandardProjectCost),
		zap.Int("remaining_credits", newResources.Credits))

	// 6. Create tile queue with "greenery" type
	err = a.playerRepo.CreateTileQueue(ctx, gameID, playerID, "standard-project-greenery", []string{"greenery"})
	if err != nil {
		log.Error("Failed to create tile queue", zap.Error(err))
		return fmt.Errorf("failed to create tile queue: %w", err)
	}

	log.Info("📋 Created tile queue for greenery placement")

	// Note: Terraform rating increase happens when the greenery is placed (via SelectTileAction)

	// 7. Consume action (only if not unlimited actions)
	// Refresh player data after tile queue creation
	p, err = ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
	if err != nil {
		return err
	}

	if p.AvailableActions > 0 {
		newActions := p.AvailableActions - 1
		err = a.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions)
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
