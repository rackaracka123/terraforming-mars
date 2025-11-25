package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

const (
	// BuildAquiferCost is the megacredit cost to build an aquifer via standard project
	BuildAquiferCost = 18
)

// BuildAquiferAction handles the business logic for the build aquifer standard project
type BuildAquiferAction struct {
	BaseAction
	gameRepo game.Repository
}

// NewBuildAquiferAction creates a new build aquifer action
func NewBuildAquiferAction(
	gameRepo game.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *BuildAquiferAction {
	return &BuildAquiferAction{
		BaseAction: NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the build aquifer action
func (a *BuildAquiferAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("💧 Building aquifer (ocean tile)")

	// 1. Validate game is active
	g, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 3. Get player from session
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Validate cost (18 M€)
	currentResources, err := player.Resources.Get(ctx)
	if err != nil {
		log.Error("Failed to get player resources", zap.Error(err))
		return fmt.Errorf("failed to get resources: %w", err)
	}

	if currentResources.Credits < BuildAquiferCost {
		log.Warn("Insufficient credits for aquifer",
			zap.Int("cost", BuildAquiferCost),
			zap.Int("player_credits", currentResources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildAquiferCost, currentResources.Credits)
	}

	// 5. Deduct cost
	newResources := currentResources
	newResources.Credits -= BuildAquiferCost
	err = player.Resources.Update(ctx, newResources)
	if err != nil {
		log.Error("Failed to deduct aquifer cost", zap.Error(err))
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("💰 Deducted aquifer cost",
		zap.Int("cost", BuildAquiferCost),
		zap.Int("remaining_credits", newResources.Credits))

	// 6. Increase terraform rating (for placing ocean)
	newTR := player.TerraformRating + 1
	err = player.Resources.UpdateTerraformRating(ctx, newTR)
	if err != nil {
		log.Error("Failed to update terraform rating", zap.Error(err))
		return fmt.Errorf("failed to update terraform rating: %w", err)
	}

	log.Info("🏆 Increased terraform rating",
		zap.Int("old_tr", player.TerraformRating),
		zap.Int("new_tr", newTR))

	// 7. Create tile queue with "ocean" type
	err = player.TileQueue.CreateQueue(ctx, "standard-project-aquifer", []string{"ocean"})
	if err != nil {
		log.Error("Failed to create tile queue", zap.Error(err))
		return fmt.Errorf("failed to create tile queue: %w", err)
	}

	log.Info("📋 Created tile queue for ocean placement")

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

	log.Info("✅ Aquifer built successfully, ocean tile queued for placement",
		zap.Int("new_terraform_rating", newTR),
		zap.Int("remaining_credits", newResources.Credits))
	return nil
}
