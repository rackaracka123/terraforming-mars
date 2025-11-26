package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

const (
	// BuildPowerPlantCost is the megacredit cost to build a power plant via standard project
	BuildPowerPlantCost = 11
)

// BuildPowerPlantAction handles the business logic for the build power plant standard project
type BuildPowerPlantAction struct {
	BaseAction
	gameRepo game.Repository
}

// NewBuildPowerPlantAction creates a new build power plant action
func NewBuildPowerPlantAction(
	gameRepo game.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *BuildPowerPlantAction {
	return &BuildPowerPlantAction{
		BaseAction: NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the build power plant action
func (a *BuildPowerPlantAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
	gameID := sess.GetGameID()
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

	// 3. Get session and player
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Validate cost (11 M€)
	if player.Resources.Credits < BuildPowerPlantCost {
		log.Warn("Insufficient credits for power plant",
			zap.Int("cost", BuildPowerPlantCost),
			zap.Int("player_credits", player.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildPowerPlantCost, player.Resources.Credits)
	}

	// 5. Deduct cost using domain method
	player.AddResources(map[types.ResourceType]int{
		types.ResourceCredits: -BuildPowerPlantCost,
	})

	log.Info("💰 Deducted power plant cost",
		zap.Int("cost", BuildPowerPlantCost),
		zap.Int("remaining_credits", player.Resources.Credits))

	// 6. Increase energy production by 1 using domain method
	player.AddProduction(map[types.ResourceType]int{
		types.ResourceEnergy: 1,
	})

	log.Info("📈 Increased energy production",
		zap.Int("new_energy_production", player.Production.Energy))

	// 7. Consume action using domain method
	if player.ConsumeAction() {
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", player.AvailableActions))
	}

	// 8. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Power plant built successfully",
		zap.Int("new_energy_production", player.Production.Energy),
		zap.Int("remaining_credits", player.Resources.Credits))
	return nil
}
