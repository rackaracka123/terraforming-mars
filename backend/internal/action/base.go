package action

import (
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"

	"go.uber.org/zap"
)

// BaseAction provides common dependencies and utilities for all actions
// All action implementations should embed this struct to access shared functionality
type BaseAction struct {
	sessionFactory    session.SessionFactory
	sessionMgrFactory session.SessionManagerFactory
	logger            *zap.Logger
}

// NewBaseAction creates a new BaseAction with common dependencies
func NewBaseAction(
	sessionFactory session.SessionFactory,
	sessionMgrFactory session.SessionManagerFactory,
) BaseAction {
	return BaseAction{
		sessionFactory:    sessionFactory,
		sessionMgrFactory: sessionMgrFactory,
		logger:            logger.Get(),
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

// GetLogger returns the base logger for actions that don't have game/player context
func (b *BaseAction) GetLogger() *zap.Logger {
	return b.logger
}

// BroadcastGameState broadcasts the current game state to all connected players
// Gets the session-specific broadcaster for this game from the factory
// Errors are logged but not returned as broadcasting failures are non-fatal
func (b *BaseAction) BroadcastGameState(gameID string, log *zap.Logger) {
	sessionMgr := b.sessionMgrFactory.GetOrCreate(gameID)
	err := sessionMgr.Broadcast()
	if err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Non-fatal - game state was updated, clients may be out of sync temporarily
	}
}

// SendToPlayer sends the current game state to a specific player
// Gets the session-specific broadcaster for this game from the factory
// Errors are logged but not returned as broadcasting failures are non-fatal
func (b *BaseAction) SendToPlayer(gameID, playerID string, log *zap.Logger) {
	sessionMgr := b.sessionMgrFactory.GetOrCreate(gameID)
	err := sessionMgr.Send(playerID)
	if err != nil {
		log.Error("Failed to send game state to player", zap.Error(err))
		// Non-fatal - game state was updated, player may be out of sync temporarily
	}
}

// GetSessionFactory returns the session factory
func (b *BaseAction) GetSessionFactory() session.SessionFactory {
	return b.sessionFactory
}

// GetSessionManagerFactory returns the session manager factory
func (b *BaseAction) GetSessionManagerFactory() session.SessionManagerFactory {
	return b.sessionMgrFactory
}
