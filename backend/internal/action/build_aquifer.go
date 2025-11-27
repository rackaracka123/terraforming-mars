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
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewBuildAquiferAction creates a new build aquifer action
func NewBuildAquiferAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *BuildAquiferAction {
	return &BuildAquiferAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the build aquifer action
func (a *BuildAquiferAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "build_aquifer"),
	)
	log.Info("💧 Building aquifer (ocean tile)")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. BUSINESS LOGIC: Validate game is active
	if g.Status() != game.GameStatusActive {
		log.Warn("Game is not active", zap.String("status", string(g.Status())))
		return fmt.Errorf("game is not active: %s", g.Status())
	}

	// 3. BUSINESS LOGIC: Validate it's the player's turn
	currentTurn := g.CurrentTurn()
	if currentTurn == nil || *currentTurn != playerID {
		var turnPlayerID string
		if currentTurn != nil {
			turnPlayerID = *currentTurn
		}
		log.Warn("Not player's turn",
			zap.String("current_turn_player", turnPlayerID),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("not your turn")
	}

	// 4. Get player from game
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 5. BUSINESS LOGIC: Validate cost (18 M€)
	resources := player.Resources().Get()
	if resources.Credits < BuildAquiferCost {
		log.Warn("Insufficient credits for aquifer",
			zap.Int("cost", BuildAquiferCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildAquiferCost, resources.Credits)
	}

	// 6. BUSINESS LOGIC: Deduct cost using domain method
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: -BuildAquiferCost,
	})

	resources = player.Resources().Get() // Refresh after update
	log.Info("💰 Deducted aquifer cost",
		zap.Int("cost", BuildAquiferCost),
		zap.Int("remaining_credits", resources.Credits))

	// 7. BUSINESS LOGIC: Increase terraform rating (for placing ocean)
	player.Resources().UpdateTerraformRating(1)

	newTR := player.Resources().TerraformRating()
	log.Info("🏆 Increased terraform rating",
		zap.Int("new_tr", newTR))

	// 8. Queue ocean tile for placement on Game (phase state managed by Game)
	queue := &playerPkg.PendingTileSelectionQueue{
		Items:  []string{"ocean"},
		Source: "standard-project-aquifer",
	}
	if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		return fmt.Errorf("failed to queue tile placement: %w", err)
	}

	log.Info("📋 Created tile queue for ocean placement")

	// 9. BUSINESS LOGIC: Consume action using domain method
	if player.Turn().ConsumeAction() {
		availableActions := player.Turn().AvailableActions()
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions))
	}

	// 10. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - player.Resources().Add() publishes ResourcesChangedEvent
	//    - player.Resources().UpdateTerraformRating() publishes TerraformRatingChangedEvent
	//    - g.SetPendingTileSelectionQueue() publishes BroadcastEvent
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("✅ Aquifer built successfully, ocean tile queued for placement",
		zap.Int("new_terraform_rating", newTR),
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
