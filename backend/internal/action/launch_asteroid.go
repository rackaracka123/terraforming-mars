package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
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
	gameRepo game.Repository
}

// NewLaunchAsteroidAction creates a new launch asteroid action
func NewLaunchAsteroidAction(
	gameRepo game.Repository,
	sessionFactory session.SessionFactory,
	sessionMgrFactory session.SessionManagerFactory,
) *LaunchAsteroidAction {
	return &LaunchAsteroidAction{
		BaseAction: NewBaseAction(sessionFactory, sessionMgrFactory),
		gameRepo:   gameRepo,
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

	// 4. Validate cost (14 M€)
	if player.Resources.Credits < LaunchAsteroidCost {
		log.Warn("Insufficient credits for asteroid",
			zap.Int("cost", LaunchAsteroidCost),
			zap.Int("player_credits", player.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", LaunchAsteroidCost, player.Resources.Credits)
	}

	// 5. Deduct cost
	newResources := player.Resources
	newResources.Credits -= LaunchAsteroidCost
	err = player.Resources.Update(ctx, newResources)
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
	newTR := player.TerraformRating + 1
	err = player.Resources.UpdateTerraformRating(ctx, newTR)
	if err != nil {
		log.Error("Failed to update terraform rating", zap.Error(err))
		return fmt.Errorf("failed to update terraform rating: %w", err)
	}

	log.Info("🏆 Increased terraform rating",
		zap.Int("old_tr", player.TerraformRating),
		zap.Int("new_tr", newTR))

	// 8. Consume action (only if not unlimited actions)
	if player.AvailableActions > 0 {
		newActions := player.AvailableActions - 1
		err = player.Action.UpdateAvailableActions(ctx, newActions)
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
