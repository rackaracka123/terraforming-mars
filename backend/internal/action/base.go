package action

import (
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
	sessionGame "terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
)

// BaseAction provides common dependencies and utilities for all actions
// All action implementations should embed this struct to access shared functionality
type BaseAction struct {
	gameRepo   sessionGame.Repository
	playerRepo player.Repository
	sessionMgr session.SessionManager
	logger     *zap.Logger
}

// NewBaseAction creates a new BaseAction with common dependencies
func NewBaseAction(
	gameRepo sessionGame.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
) BaseAction {
	return BaseAction{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
		sessionMgr: sessionMgr,
		logger:     logger.Get(),
	}
}

// InitLogger creates a logger with game and player context
// This should be called at the start of every Execute method
func (b *BaseAction) InitLogger(gameID, playerID string) *zap.Logger {
	return b.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
	)
}

// BroadcastGameState broadcasts the current game state to all connected players
// Errors are logged but not returned as broadcasting failures are non-fatal
func (b *BaseAction) BroadcastGameState(gameID string, log *zap.Logger) {
	err := b.sessionMgr.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Non-fatal - game state was updated, clients may be out of sync temporarily
	}
}

// SendToPlayer sends the current game state to a specific player
// Errors are logged but not returned as broadcasting failures are non-fatal
func (b *BaseAction) SendToPlayer(gameID, playerID string, log *zap.Logger) {
	err := b.sessionMgr.Send(gameID, playerID)
	if err != nil {
		log.Error("Failed to send game state to player", zap.Error(err))
		// Non-fatal - game state was updated, player may be out of sync temporarily
	}
}

// GetGameRepo returns the game repository
func (b *BaseAction) GetGameRepo() sessionGame.Repository {
	return b.gameRepo
}

// GetPlayerRepo returns the player repository
func (b *BaseAction) GetPlayerRepo() player.Repository {
	return b.playerRepo
}

// GetSessionManager returns the session manager
func (b *BaseAction) GetSessionManager() session.SessionManager {
	return b.sessionMgr
}
