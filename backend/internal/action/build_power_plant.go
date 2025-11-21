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
	// BuildPowerPlantCost is the megacredit cost to build a power plant via standard project
	BuildPowerPlantCost = 11
)

// BuildPowerPlantAction handles the business logic for the build power plant standard project
type BuildPowerPlantAction struct {
	BaseAction
}

// NewBuildPowerPlantAction creates a new build power plant action
func NewBuildPowerPlantAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
) *BuildPowerPlantAction {
	return &BuildPowerPlantAction{
		BaseAction: NewBaseAction(gameRepo, playerRepo, sessionMgr),
	}
}

// Execute performs the build power plant action
func (a *BuildPowerPlantAction) Execute(ctx context.Context, gameID, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("⚡ Building power plant")

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

	// 4. Validate cost (11 M€)
	if p.Resources.Credits < BuildPowerPlantCost {
		log.Warn("Insufficient credits for power plant",
			zap.Int("cost", BuildPowerPlantCost),
			zap.Int("player_credits", p.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildPowerPlantCost, p.Resources.Credits)
	}

	// 5. Deduct cost
	newResources := p.Resources
	newResources.Credits -= BuildPowerPlantCost
	err = a.playerRepo.UpdateResources(ctx, gameID, playerID, newResources)
	if err != nil {
		log.Error("Failed to deduct power plant cost", zap.Error(err))
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("💰 Deducted power plant cost",
		zap.Int("cost", BuildPowerPlantCost),
		zap.Int("remaining_credits", newResources.Credits))

	// 6. Increase energy production by 1
	newProduction := p.Production
	newProduction.Energy++
	err = a.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction)
	if err != nil {
		log.Error("Failed to update production", zap.Error(err))
		return fmt.Errorf("failed to update production: %w", err)
	}

	log.Info("📈 Increased energy production",
		zap.Int("new_energy_production", newProduction.Energy))

	// 7. Consume action (only if not unlimited actions)
	// Refresh player data after production update
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

	log.Info("✅ Power plant built successfully",
		zap.Int("new_energy_production", newProduction.Energy),
		zap.Int("remaining_credits", newResources.Credits))
	return nil
}
