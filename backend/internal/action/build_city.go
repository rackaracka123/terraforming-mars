package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game/board"
	game "terraforming-mars-backend/internal/session/game/core"
	playerTypes "terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/types"
)

const (
	// BuildCityCost is the megacredit cost to build a city via standard project
	BuildCityCost = 25
)

// BuildCityAction handles the business logic for building a city standard project
type BuildCityAction struct {
	BaseAction
	gameRepo      game.Repository
	tileProcessor *board.Processor
}

// NewBuildCityAction creates a new build city action
func NewBuildCityAction(
	gameRepo game.Repository,
	tileProcessor *board.Processor,
	sessionMgrFactory session.SessionManagerFactory,
) *BuildCityAction {
	return &BuildCityAction{
		BaseAction:    NewBaseAction(sessionMgrFactory),
		gameRepo:      gameRepo,
		tileProcessor: tileProcessor,
	}
}

// Execute performs the build city action
func (a *BuildCityAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("🏢 Building city")

	// 1. Validate game is active
	_, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Get session and player
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. Validate cost (25 M€)
	resources := player.Resources().Get()
	if resources.Credits < BuildCityCost {
		log.Warn("Insufficient credits for city",
			zap.Int("cost", BuildCityCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildCityCost, resources.Credits)
	}

	// 4. Deduct cost using domain method
	player.Resources().Add(map[types.ResourceType]int{
		types.ResourceCredits: -BuildCityCost,
	})

	resources = player.Resources().Get() // Refresh after update
	log.Info("💰 Deducted city cost",
		zap.Int("cost", BuildCityCost),
		zap.Int("remaining_credits", resources.Credits))

	// 5. Increase credit production by 1 using domain method
	player.Resources().AddProduction(map[types.ResourceType]int{
		types.ResourceCredits: 1,
	})

	production := player.Resources().Production() // Refresh after update
	log.Info("📈 Increased credit production",
		zap.Int("new_credit_production", production.Credits))

	// 6. Queue city tile for placement on Game (phase state managed by Game)
	queue := &playerTypes.PendingTileSelectionQueue{
		Items:  []string{"city"},
		Source: "standard-project-city",
	}
	if err := sess.Game().SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		return fmt.Errorf("failed to queue tile placement: %w", err)
	}

	log.Info("📋 Created tile queue for city placement")

	// 7. Tile queue processing (now automatic via TileQueueCreatedEvent)
	// No manual call needed - TileProcessor subscribes to events and processes automatically

	// 8. Consume action using domain method
	if player.Turn().ConsumeAction() {
		availableActions := player.Turn().AvailableActions()
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions))
	}

	// 9. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ City built successfully, tile queued for placement",
		zap.Int("new_credit_production", production.Credits),
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
