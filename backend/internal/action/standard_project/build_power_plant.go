package standard_project

import (
	baseaction "terraforming-mars-backend/internal/action"
	"context"
	"fmt"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

const (
	// BuildPowerPlantCost is the megacredit cost to build a power plant via standard project
	BuildPowerPlantCost = 11
)

// BuildPowerPlantAction handles the build power plant standard project
// New architecture: Uses only GameRepository + logger
type BuildPowerPlantAction struct {
	baseaction.BaseAction
}

// NewBuildPowerPlantAction creates a new build power plant action
func NewBuildPowerPlantAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *BuildPowerPlantAction {
	return &BuildPowerPlantAction{
		BaseAction: baseaction.NewBaseAction(gameRepo, nil),
	}
}

// Execute performs the build power plant action
func (a *BuildPowerPlantAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("⚡ Building power plant")

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

	// 5. Validate cost (11 M€)
	resources := player.Resources().Get()
	if resources.Credits < BuildPowerPlantCost {
		log.Warn("Insufficient credits for power plant",
			zap.Int("cost", BuildPowerPlantCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildPowerPlantCost, resources.Credits)
	}

	// 6. Deduct cost (publishes ResourcesChangedEvent)
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: -BuildPowerPlantCost,
	})

	resources = player.Resources().Get() // Refresh after update
	log.Info("💰 Deducted power plant cost",
		zap.Int("cost", BuildPowerPlantCost),
		zap.Int("remaining_credits", resources.Credits))

	// 7. Increase energy production by 1 (publishes ProductionChangedEvent)
	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 1,
	})

	production := player.Resources().Production() // Refresh after update
	log.Info("📈 Increased energy production",
		zap.Int("new_energy_production", production.Energy))

	// 8. Consume action (only if not unlimited actions)
	a.ConsumePlayerAction(g, log)

	log.Info("✅ Power plant built successfully",
		zap.Int("new_energy_production", production.Energy),
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
