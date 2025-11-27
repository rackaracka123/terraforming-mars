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
	// BuildCityCost is the megacredit cost to build a city via standard project
	BuildCityCost = 25
)

// BuildCityAction handles the business logic for building a city standard project
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type BuildCityAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewBuildCityAction creates a new build city action
func NewBuildCityAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *BuildCityAction {
	return &BuildCityAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the build city action
func (a *BuildCityAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "build_city"),
	)
	log.Info("🏢 Building city")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. Validate game is active
	if g.Status() != game.GameStatusActive {
		log.Warn("Game is not active", zap.String("status", string(g.Status())))
		return fmt.Errorf("game is not active: %s", g.Status())
	}

	// 3. Get player from game
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. BUSINESS LOGIC: Validate cost (25 M€)
	resources := player.Resources().Get()
	if resources.Credits < BuildCityCost {
		log.Warn("Insufficient credits for city",
			zap.Int("cost", BuildCityCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildCityCost, resources.Credits)
	}

	// 5. BUSINESS LOGIC: Deduct cost using domain method
	// Player.game.Resources() is already encapsulated - no changes needed
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: -BuildCityCost,
	})

	resources = player.Resources().Get() // Refresh after update
	log.Info("💰 Deducted city cost",
		zap.Int("cost", BuildCityCost),
		zap.Int("remaining_credits", resources.Credits))

	// 6. BUSINESS LOGIC: Increase credit production by 1 using domain method
	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCredits: 1,
	})

	production := player.Resources().Production() // Refresh after update
	log.Info("📈 Increased credit production",
		zap.Int("new_credit_production", production.Credits))

	// 7. Queue city tile for placement on Game (phase state managed by Game)
	queue := &playerPkg.PendingTileSelectionQueue{
		Items:  []string{"city"},
		Source: "standard-project-city",
	}
	if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		return fmt.Errorf("failed to queue tile placement: %w", err)
	}

	log.Info("📋 Created tile queue for city placement")

	// 8. BUSINESS LOGIC: Consume action using domain method
	if player.Turn().ConsumeAction() {
		availableActions := player.Turn().AvailableActions()
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions))
	}

	// 9. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - g.SetPendingTileSelectionQueue() publishes BroadcastEvent
	//    - player.Resources().Add() publishes ResourcesChangedEvent
	//    - player.Resources().AddProduction() publishes ProductionChangedEvent
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("✅ City built successfully, tile queued for placement",
		zap.Int("new_credit_production", production.Credits),
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
