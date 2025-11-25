package convert_plants_to_greenery

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"

	"go.uber.org/zap"
)

// Handler handles convert plants to greenery action requests
type Handler struct {
	convertPlantsAction *action.ConvertPlantsToGreeneryAction
	sessionFactory      session.SessionFactory
}

// NewHandler creates a new convert plants to greenery handler
func NewHandler(convertPlantsAction *action.ConvertPlantsToGreeneryAction, sessionFactory session.SessionFactory) *Handler {
	return &Handler{
		convertPlantsAction: convertPlantsAction,
		sessionFactory:      sessionFactory,
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := logger.Get().With(
		zap.String("connection_id", connection.ID),
		zap.String("player_id", connection.PlayerID),
		zap.String("game_id", connection.GameID))

	log.Debug("🌱 Processing convert plants to greenery action")

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

	// Execute the convert plants to greenery action
	err := h.convertPlantsAction.Execute(ctx, sess, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute convert plants to greenery action", zap.Error(err))
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": err.Error()},
		}
		return
	}

	log.Info("✅ Convert plants to greenery action completed successfully")
}
