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
	BaseAction
}

// NewBuildCityAction creates a new build city action
func NewBuildCityAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *BuildCityAction {
	return &BuildCityAction{
		BaseAction: BaseAction{
			gameRepo: gameRepo,
			logger:   logger,
		},
	}
}

// Execute performs the build city action
func (a *BuildCityAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "build_city"))
	log.Info("🏢 Building city")

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

	// 8. Process the queue to create PendingTileSelection with available hexes
	if err := g.ProcessNextTile(ctx, playerID); err != nil {
		return fmt.Errorf("failed to process tile queue: %w", err)
	}

	log.Info("🎯 Processed tile queue into pending tile selection")

	// 9. Consume action (only if not unlimited actions)
	a.ConsumePlayerAction(g, log)

	// 10. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//     - g.ProcessNextTile() -> g.SetPendingTileSelection() publishes BroadcastEvent
	//     - player.Resources().Add() publishes ResourcesChangedEvent
	//     - player.Resources().AddProduction() publishes ProductionChangedEvent
	//     Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("✅ City built successfully, tile selection ready",
		zap.Int("new_credit_production", production.Credits),
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
