package player

import (
	"context"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/types"
)

var logTurn = logger.Get()

// TurnRepository handles turn state and action management for a specific player
// Auto-saves changes after every operation
type TurnRepository struct {
	player *Player // Reference to parent player
}

// NewTurnRepository creates a new turn repository for a specific player
func NewTurnRepository(player *Player) *TurnRepository {
	return &TurnRepository{
		player: player,
	}
}

// UpdateConnectionStatus updates player connection status
// Auto-saves changes to the player
func (r *TurnRepository) UpdateConnectionStatus(ctx context.Context, isConnected bool) error {
	r.player.IsConnected = isConnected
	return nil
}

// UpdatePassed updates player passed status for generation
// Auto-saves changes to the player
func (r *TurnRepository) UpdatePassed(ctx context.Context, passed bool) error {
	r.player.Passed = passed

	logTurn.Debug("✅ Player passed status updated",
		zap.String("game_id", r.player.GameID),
		zap.String("player_id", r.player.ID),
		zap.Bool("passed", passed))

	return nil
}

// UpdateAvailableActions updates player available actions count
// Auto-saves changes to the player
func (r *TurnRepository) UpdateAvailableActions(ctx context.Context, actions int) error {
	r.player.AvailableActions = actions

	logTurn.Debug("✅ Player available actions updated",
		zap.String("game_id", r.player.GameID),
		zap.String("player_id", r.player.ID),
		zap.Int("available_actions", actions))

	return nil
}

// UpdateActions updates player actions
// Auto-saves changes to the player
func (r *TurnRepository) UpdateActions(ctx context.Context, actions []types.PlayerAction) error {
	r.player.Actions = actions
	return nil
}

// UpdateForcedFirstAction updates player forced first action
// Auto-saves changes to the player
func (r *TurnRepository) UpdateForcedFirstAction(ctx context.Context, action *types.ForcedFirstAction) error {
	r.player.ForcedFirstAction = action
	return nil
}
