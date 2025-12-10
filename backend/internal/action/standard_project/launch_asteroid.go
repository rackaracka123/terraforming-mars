package standard_project

import (
	baseaction "terraforming-mars-backend/internal/action"
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

const (
	// LaunchAsteroidCost is the megacredit cost to launch an asteroid via standard project
	LaunchAsteroidCost = 14
)

// LaunchAsteroidAction handles the business logic for the launch asteroid standard project
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type LaunchAsteroidAction struct {
	baseaction.BaseAction
}

// NewLaunchAsteroidAction creates a new launch asteroid action
func NewLaunchAsteroidAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *LaunchAsteroidAction {
	return &LaunchAsteroidAction{
		BaseAction: baseaction.NewBaseAction(gameRepo, nil),
	}
}

// Execute performs the launch asteroid action
func (a *LaunchAsteroidAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "launch_asteroid"))
	log.Info("☄️ Launching asteroid")

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

	// 5. BUSINESS LOGIC: Validate cost (14 M€)
	resources := player.Resources().Get()
	if resources.Credits < LaunchAsteroidCost {
		log.Warn("Insufficient credits for asteroid",
			zap.Int("cost", LaunchAsteroidCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", LaunchAsteroidCost, resources.Credits)
	}

	// 6. BUSINESS LOGIC: Deduct cost using domain method
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: -LaunchAsteroidCost,
	})

	resources = player.Resources().Get() // Refresh after update
	log.Info("💰 Deducted asteroid cost",
		zap.Int("cost", LaunchAsteroidCost),
		zap.Int("remaining_credits", resources.Credits))

	// 7. BUSINESS LOGIC: Increase temperature by 1 step
	oldTemp := g.GlobalParameters().Temperature()
	stepsRaised, err := g.GlobalParameters().IncreaseTemperature(ctx, 1)
	if err != nil {
		log.Error("Failed to increase temperature", zap.Error(err))
		return fmt.Errorf("failed to increase temperature: %w", err)
	}
	newTemp := g.GlobalParameters().Temperature()

	if stepsRaised > 0 {
		log.Info("🌡️ Increased temperature",
			zap.Int("old_temperature", oldTemp),
			zap.Int("new_temperature", newTemp),
			zap.Int("steps_raised", stepsRaised))
	}

	// 8. BUSINESS LOGIC: Increase terraform rating (only if temperature actually increased)
	if stepsRaised > 0 {
		oldTR := player.Resources().TerraformRating()
		player.Resources().UpdateTerraformRating(1)
		newTR := player.Resources().TerraformRating()

		log.Info("🏆 Increased terraform rating",
			zap.Int("old_tr", oldTR),
			zap.Int("new_tr", newTR))
	}

	// 9. Consume action (only if not unlimited actions)
	a.ConsumePlayerAction(g, log)

	log.Info("✅ Asteroid launched successfully",
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
