package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"
	playerTypes "terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/types"

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
	resources := player.Resources().Get()
	if resources.Credits < BuildAquiferCost {
		log.Warn("Insufficient credits for aquifer",
			zap.Int("cost", BuildAquiferCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildAquiferCost, resources.Credits)
	}

	// 5. Deduct cost using domain method
	player.Resources().Add(map[types.ResourceType]int{
		types.ResourceCredits: -BuildAquiferCost,
	})

	resources = player.Resources().Get() // Refresh after update
	log.Info("💰 Deducted aquifer cost",
		zap.Int("cost", BuildAquiferCost),
		zap.Int("remaining_credits", resources.Credits))

	// 6. Increase terraform rating (for placing ocean)
	player.Resources().UpdateTerraformRating(1)

	newTR := player.Resources().TerraformRating()
	log.Info("🏆 Increased terraform rating",
		zap.Int("new_tr", newTR))

	// 7. Queue ocean tile for placement on Game (phase state managed by Game)
	queue := &playerTypes.PendingTileSelectionQueue{
		Items:  []string{"ocean"},
		Source: "standard-project-aquifer",
	}
	if err := sess.Game().SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		return fmt.Errorf("failed to queue tile placement: %w", err)
	}

	log.Info("📋 Created tile queue for ocean placement")

	// 8. Consume action using domain method
	if player.Turn().ConsumeAction() {
		availableActions := player.Turn().AvailableActions()
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions))
	}

	// 9. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Aquifer built successfully, ocean tile queued for placement",
		zap.Int("new_terraform_rating", newTR),
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
