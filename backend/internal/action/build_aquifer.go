package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

const (
	// BuildAquiferCost is the megacredit cost to build an aquifer via standard project
	BuildAquiferCost = 18
)

// BuildAquiferAction handles the business logic for the build aquifer standard project
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type BuildAquiferAction struct {
	BaseAction
}

// NewBuildAquiferAction creates a new build aquifer action
func NewBuildAquiferAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *BuildAquiferAction {
	return &BuildAquiferAction{
		BaseAction: BaseAction{
			gameRepo: gameRepo,
			logger:   logger,
		},
	}
}

// Execute performs the build aquifer action
func (a *BuildAquiferAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "build_aquifer"))
	log.Info("💧 Building aquifer (ocean tile)")

	// 1. Fetch game from repository and validate it's active
	g, err := ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 3. Get player from game
	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	// 5. BUSINESS LOGIC: Validate ocean count hasn't reached maximum
	currentOceans := g.GlobalParameters().Oceans()
	if currentOceans >= 9 { // MaxOceans constant from global_parameters
		log.Warn("Cannot build aquifer: oceans already at maximum",
			zap.Int("current_oceans", currentOceans),
			zap.Int("max_oceans", 9))
		return fmt.Errorf("cannot build aquifer: oceans already at maximum (%d/%d)", currentOceans, 9)
	}

	// 6. BUSINESS LOGIC: Validate cost (18 M€)
	resources := player.Resources().Get()
	if resources.Credits < BuildAquiferCost {
		log.Warn("Insufficient credits for aquifer",
			zap.Int("cost", BuildAquiferCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildAquiferCost, resources.Credits)
	}

	// 7. BUSINESS LOGIC: Deduct cost using domain method
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: -BuildAquiferCost,
	})

	resources = player.Resources().Get() // Refresh after update
	log.Info("💰 Deducted aquifer cost",
		zap.Int("cost", BuildAquiferCost),
		zap.Int("remaining_credits", resources.Credits))

	// 8. BUSINESS LOGIC: Increase terraform rating (for placing ocean)
	player.Resources().UpdateTerraformRating(1)

	newTR := player.Resources().TerraformRating()
	log.Info("🏆 Increased terraform rating",
		zap.Int("new_tr", newTR))

	// 9. Queue ocean tile for placement on Game (phase state managed by Game)
	queue := &playerPkg.PendingTileSelectionQueue{
		Items:  []string{"ocean"},
		Source: "standard-project-aquifer",
	}
	if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		return fmt.Errorf("failed to queue tile placement: %w", err)
	}

	log.Info("📋 Created tile queue for ocean placement (auto-processed by SetPendingTileSelectionQueue)")

	// 10. Consume action (only if not unlimited actions)
	a.ConsumePlayerAction(g, log)

	log.Info("✅ Aquifer built successfully, ocean tile queued for placement",
		zap.Int("new_terraform_rating", newTR),
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
