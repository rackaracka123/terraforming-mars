package action

import (
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
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewBuildPowerPlantAction creates a new build power plant action
func NewBuildPowerPlantAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *BuildPowerPlantAction {
	return &BuildPowerPlantAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the build power plant action
func (a *BuildPowerPlantAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
	)
	log.Info("⚡ Building power plant")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %w", err)
	}

	// 2. Validate game is active
	if g.Status() != game.GameStatusActive {
		log.Warn("Game is not active", zap.String("status", string(g.Status())))
		return fmt.Errorf("game is not active: %s", g.Status())
	}

	// 3. Validate it's the player's turn
	currentTurn := g.CurrentTurn()
	if currentTurn == nil || *currentTurn != playerID {
		log.Warn("Not player's turn")
		return fmt.Errorf("not your turn")
	}

	// 4. Get player from game
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
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
		shared.ResourceEnergy: 1,
	})

	production := player.Resources().Production() // Refresh after update
	log.Info("📈 Increased energy production",
		zap.Int("new_energy_production", production.Energy))

	// 8. Consume action
	if player.Turn().ConsumeAction() {
		availableActions := player.Turn().AvailableActions()
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions))
	}

	// 9. NO MANUAL BROADCAST - Events automatically trigger:
	//     - ResourcesChangedEvent → SessionManager → WebSocket broadcast
	//     - ProductionChangedEvent → SessionManager → WebSocket broadcast

	log.Info("✅ Power plant built successfully",
		zap.Int("new_energy_production", production.Energy),
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
