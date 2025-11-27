package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
)

// SetCorporationAction handles the admin action to set a player's corporation
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
// NOTE: Corporation validation is skipped (admin action with trusted input)
// NOTE: Uses SetID instead of SetCard since we don't have card repository access
type SetCorporationAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewSetCorporationAction creates a new set corporation admin action
func NewSetCorporationAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SetCorporationAction {
	return &SetCorporationAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the set corporation admin action
func (a *SetCorporationAction) Execute(ctx context.Context, gameID string, playerID string, corporationID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "admin_set_corporation"),
		zap.String("corporation_id", corporationID),
	)
	log.Info("🏢 Admin: Setting player corporation")

	// 1. Fetch game from repository
	game, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. Get player from game
	player, err := game.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. Update player corporation
	// NOTE: Corporation validation is skipped - admin actions are trusted to provide valid corporation IDs
	// In production, you might want to validate the corporation ID exists
	player.SetCorporationID(corporationID)

	log.Info("✅ Player corporation updated")

	// 4. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - player.SetCorporationID() publishes events
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("✅ Admin set corporation completed")
	return nil
}
