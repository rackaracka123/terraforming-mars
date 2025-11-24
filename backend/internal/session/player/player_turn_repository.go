package player

import (
	"context"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/types"
)

var logTurn = logger.Get()

// PlayerTurnRepository handles player turn state and action management
type PlayerTurnRepository struct {
	storage *PlayerStorage
}

// NewPlayerTurnRepository creates a new player turn repository
func NewPlayerTurnRepository(storage *PlayerStorage) *PlayerTurnRepository {
	return &PlayerTurnRepository{
		storage: storage,
	}
}

// UpdateConnectionStatus updates player connection status
func (r *PlayerTurnRepository) UpdateConnectionStatus(ctx context.Context, gameID string, playerID string, isConnected bool) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.IsConnected = isConnected

	return r.storage.Set(gameID, playerID, p)
}

// UpdatePassed updates player passed status for generation
func (r *PlayerTurnRepository) UpdatePassed(ctx context.Context, gameID string, playerID string, passed bool) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.Passed = passed

	logTurn.Debug("✅ Player passed status updated",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Bool("passed", passed))

	return r.storage.Set(gameID, playerID, p)
}

// UpdateAvailableActions updates player available actions count
func (r *PlayerTurnRepository) UpdateAvailableActions(ctx context.Context, gameID string, playerID string, actions int) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.AvailableActions = actions

	logTurn.Debug("✅ Player available actions updated",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Int("available_actions", actions))

	return r.storage.Set(gameID, playerID, p)
}

// UpdatePlayerActions updates player actions
func (r *PlayerTurnRepository) UpdatePlayerActions(ctx context.Context, gameID string, playerID string, actions []types.PlayerAction) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.Actions = actions

	return r.storage.Set(gameID, playerID, p)
}

// UpdateForcedFirstAction updates player forced first action
func (r *PlayerTurnRepository) UpdateForcedFirstAction(ctx context.Context, gameID string, playerID string, action *types.ForcedFirstAction) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.ForcedFirstAction = action

	return r.storage.Set(gameID, playerID, p)
}
