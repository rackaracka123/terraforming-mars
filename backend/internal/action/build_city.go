package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/game/tile"
)

const (
	// BuildCityCost is the megacredit cost to build a city via standard project
	BuildCityCost = 25
)

// BuildCityAction handles the business logic for building a city standard project
type BuildCityAction struct {
	BaseAction
	tileProcessor *tile.Processor
}

// NewBuildCityAction creates a new build city action
func NewBuildCityAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	tileProcessor *tile.Processor,
	sessionMgr session.SessionManager,
) *BuildCityAction {
	return &BuildCityAction{
		BaseAction:    NewBaseAction(gameRepo, playerRepo, sessionMgr),
		tileProcessor: tileProcessor,
	}
}

// Execute performs the build city action
func (a *BuildCityAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🏢 Building city")

	// 1. Validate game is active
	_, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate player exists
	p, err := ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
	if err != nil {
		return err
	}

	// 3. Validate cost (25 M€)
	if p.Resources.Credits < BuildCityCost {
		log.Warn("Insufficient credits for city",
			zap.Int("cost", BuildCityCost),
			zap.Int("player_credits", p.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildCityCost, p.Resources.Credits)
	}

	// 4. Deduct cost
	newResources := p.Resources
	newResources.Credits -= BuildCityCost
	err = a.playerRepo.UpdateResources(ctx, gameID, playerID, newResources)
	if err != nil {
		log.Error("Failed to deduct city cost", zap.Error(err))
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("💰 Deducted city cost",
		zap.Int("cost", BuildCityCost),
		zap.Int("remaining_credits", newResources.Credits))

	// 5. Increase credit production by 1
	newProduction := p.Production
	newProduction.Credits++
	err = a.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction)
	if err != nil {
		log.Error("Failed to update production", zap.Error(err))
		return fmt.Errorf("failed to update production: %w", err)
	}

	log.Info("📈 Increased credit production",
		zap.Int("new_credit_production", newProduction.Credits))

	// 6. Create tile queue with "city" type
	err = a.playerRepo.CreateTileQueue(ctx, gameID, playerID, "standard-project-city", []string{"city"})
	if err != nil {
		log.Error("Failed to create tile queue", zap.Error(err))
		return fmt.Errorf("failed to create tile queue: %w", err)
	}

	log.Info("📋 Created tile queue for city placement")

	// 7. Tile queue processing (now automatic via TileQueueCreatedEvent)
	// No manual call needed - TileProcessor subscribes to events and processes automatically

	// 8. Consume action (only if not unlimited actions)
	// Refresh player data after tile queue creation
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

	log.Info("✅ City built successfully, tile queued for placement",
		zap.Int("new_credit_production", newProduction.Credits),
		zap.Int("remaining_credits", newResources.Credits))
	return nil
}
