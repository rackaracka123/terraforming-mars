package build_city

import (
	"context"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
)

// Handler handles build city standard project action requests
type Handler struct {
	buildCityAction *action.BuildCityAction
	sessionFactory  session.SessionFactory
}

// NewHandler creates a new build city handler
func NewHandler(buildCityAction *action.BuildCityAction, sessionFactory session.SessionFactory) *Handler {
	return &Handler{
		buildCityAction: buildCityAction,
		sessionFactory:  sessionFactory,
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := logger.Get().With(
		zap.String("connection_id", connection.ID),
		zap.String("player_id", connection.PlayerID),
		zap.String("game_id", connection.GameID),
	)
	log.Debug("🏢 Processing build city action")

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

	// Execute the build city action
	err := h.buildCityAction.Execute(ctx, sess, connection.PlayerID)
	if err != nil {
		log.Error("Failed to build city", zap.Error(err))
		// Send error message via channel
		connection.Send <- dto.WebSocketMessage{
			Type: dto.MessageTypeError,
			Payload: map[string]interface{}{
				"message": err.Error(),
			},
		}
		return
	}

	log.Info("✅ City built successfully")
}
