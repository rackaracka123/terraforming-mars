package build_aquifer

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"

	"go.uber.org/zap"
)

// Handler handles build aquifer standard project action requests
type Handler struct {
	buildAquiferAction *action.BuildAquiferAction
	sessionFactory     session.SessionFactory
}

// NewHandler creates a new build aquifer handler
func NewHandler(buildAquiferAction *action.BuildAquiferAction, sessionFactory session.SessionFactory) *Handler {
	return &Handler{
		buildAquiferAction: buildAquiferAction,
		sessionFactory:     sessionFactory,
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := logger.Get().With(
		zap.String("connection_id", connection.ID),
		zap.String("player_id", connection.PlayerID),
		zap.String("game_id", connection.GameID))

	log.Debug("🌊 Processing build aquifer action")

	// Get session for the game
	sess := h.sessionFactory.Get(connection.GameID)
	if sess == nil {
		log.Error("Session not found")
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": "Game session not found"},
		}
		return
	}

	// Execute the build aquifer action
	err := h.buildAquiferAction.Execute(ctx, sess, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute build aquifer action", zap.Error(err))
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": err.Error()},
		}
		return
	}

	log.Info("✅ Build aquifer action completed successfully")
}
