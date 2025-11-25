package skip_action

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"

	"go.uber.org/zap"
)

// Handler handles skip action using the action pattern
type Handler struct {
	skipActionAction SkipActionAction
	sessionFactory   session.SessionFactory
	errorHandler     *utils.ErrorHandler
	logger           *zap.Logger
}

// SkipActionAction interface for dependency injection
type SkipActionAction interface {
	Execute(ctx context.Context, sess *session.Session, playerID string) error
}

// NewHandler creates a new skip action handler using action pattern
func NewHandler(skipActionAction SkipActionAction, sessionFactory session.SessionFactory) *Handler {
	return &Handler{
		skipActionAction: skipActionAction,
		sessionFactory:   sessionFactory,
		errorHandler:     utils.NewErrorHandler(),
		logger:           logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Skip action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("⏭️ Processing skip action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Get session for the game
	sess := h.sessionFactory.Get(gameID)
	if sess == nil {
		h.logger.Error("Session not found", zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, "Game session not found")
		return
	}

	// Use action pattern - direct orchestration
	if err := h.skipActionAction.Execute(ctx, sess, playerID); err != nil {
		h.logger.Error("Failed to skip action",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Skip action completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}
