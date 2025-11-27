package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
)

// PlayerReconnectedAction handles the business logic for player reconnection
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type PlayerReconnectedAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewPlayerReconnectedAction creates a new player reconnected action
func NewPlayerReconnectedAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *PlayerReconnectedAction {
	return &PlayerReconnectedAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the player reconnected action
func (a *PlayerReconnectedAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "player_reconnected"),
	)
	log.Info("🔗 Player reconnecting")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. Get player from game
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. Update player connection status to connected
	player.Turn().SetConnectionStatus(true)

	log.Info("✅ Player connection status updated to connected")

	// 4. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - player.Turn().SetConnectionStatus() publishes events
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates
	//
	// NOTE: In old architecture, this action explicitly sent state to reconnected player first,
	// then broadcast to all. In new architecture, the event-driven broadcast handles both.
	// The Broadcaster will send personalized state to all players including the reconnected one.

	log.Info("✅ Player reconnected successfully")
	return nil
}
