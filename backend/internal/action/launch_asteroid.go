package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

const (
	// LaunchAsteroidCost is the megacredit cost to launch an asteroid via standard project
	LaunchAsteroidCost = 14
)

// LaunchAsteroidAction handles the business logic for the launch asteroid standard project
type LaunchAsteroidAction struct {
	BaseAction
}

// NewLaunchAsteroidAction creates a new launch asteroid action
func NewLaunchAsteroidAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *LaunchAsteroidAction {
	return &LaunchAsteroidAction{
		BaseAction: NewBaseAction(gameRepo, playerRepo, sessionMgrFactory),
	}
}

// Execute performs the launch asteroid action
func (a *LaunchAsteroidAction) Execute(ctx context.Context, gameID, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("☄️ Launching asteroid")

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

	// 4. Validate cost (14 M€)
	if p.Resources.Credits < LaunchAsteroidCost {
		log.Warn("Insufficient credits for asteroid",
			zap.Int("cost", LaunchAsteroidCost),
			zap.Int("player_credits", p.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", LaunchAsteroidCost, p.Resources.Credits)
	}

	// 5. Deduct cost
	newResources := p.Resources
	newResources.Credits -= LaunchAsteroidCost
	err = a.playerRepo.UpdateResources(ctx, gameID, playerID, newResources)
	if err != nil {
		log.Error("Failed to deduct asteroid cost", zap.Error(err))
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("💰 Deducted asteroid cost",
		zap.Int("cost", LaunchAsteroidCost),
		zap.Int("remaining_credits", newResources.Credits))

	// 6. Increase temperature by 1 step (2°C)
	if g.GlobalParameters.Temperature < types.MaxTemperature {
		newTemperature := g.GlobalParameters.Temperature + 2 // Each step is 2°C
		if newTemperature > types.MaxTemperature {
			newTemperature = types.MaxTemperature
		}

		err = a.gameRepo.UpdateTemperature(ctx, gameID, newTemperature)
		if err != nil {
			log.Error("Failed to update temperature", zap.Error(err))
			return fmt.Errorf("failed to update temperature: %w", err)
		}

		log.Info("🌡️ Increased temperature",
			zap.Int("old_temperature", g.GlobalParameters.Temperature),
			zap.Int("new_temperature", newTemperature))
	}

	// 7. Increase terraform rating
	newTR := p.TerraformRating + 1
	err = a.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR)
	if err != nil {
		log.Error("Failed to update terraform rating", zap.Error(err))
		return fmt.Errorf("failed to update terraform rating: %w", err)
	}

	log.Info("🏆 Increased terraform rating",
		zap.Int("old_tr", p.TerraformRating),
		zap.Int("new_tr", newTR))

	// 8. Consume action (only if not unlimited actions)
	// Refresh player data
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

	// 9. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Asteroid launched successfully",
		zap.Int("new_terraform_rating", newTR),
		zap.Int("remaining_credits", newResources.Credits))
	return nil
}
