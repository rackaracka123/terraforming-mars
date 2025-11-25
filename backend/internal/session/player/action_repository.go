package player

import (
	"context"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/types"
)

var logAction = logger.Get()

// ActionRepository handles player turn actions and forced actions
// Auto-saves changes after every operation
type ActionRepository struct {
	player *Player // Reference to parent player
}

// NewActionRepository creates a new action repository for a specific player
func NewActionRepository(player *Player) *ActionRepository {
	return &ActionRepository{
		player: player,
	}
}

// UpdatePassed updates player passed status for generation
// Auto-saves changes to the player
func (r *ActionRepository) UpdatePassed(ctx context.Context, passed bool) error {
	r.player.Passed = passed

	logAction.Debug("✅ Player passed status updated",
		zap.String("game_id", r.player.GameID),
		zap.String("player_id", r.player.ID),
		zap.Bool("passed", passed))

	return nil
}

// UpdateAvailableActions updates player available actions count
// Auto-saves changes to the player
func (r *ActionRepository) UpdateAvailableActions(ctx context.Context, actions int) error {
	r.player.AvailableActions = actions

	logAction.Debug("✅ Player available actions updated",
		zap.String("game_id", r.player.GameID),
		zap.String("player_id", r.player.ID),
		zap.Int("available_actions", actions))

	return nil
}

// UpdateActions updates player actions list
// Auto-saves changes to the player
func (r *ActionRepository) UpdateActions(ctx context.Context, actions []types.PlayerAction) error {
	r.player.Actions = actions
	return nil
}

// UpdateForcedFirstAction updates player forced first action
// Auto-saves changes to the player
func (r *ActionRepository) UpdateForcedFirstAction(ctx context.Context, action *types.ForcedFirstAction) error {
	r.player.ForcedFirstAction = action
	return nil
}
