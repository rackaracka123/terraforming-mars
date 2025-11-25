package sell_patents

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"

	"go.uber.org/zap"
)

// Handler handles sell patents standard project action requests
type Handler struct {
	sellPatentsAction *action.SellPatentsAction
	sessionFactory    session.SessionFactory
}

// NewHandler creates a new sell patents handler
func NewHandler(sellPatentsAction *action.SellPatentsAction, sessionFactory session.SessionFactory) *Handler {
	return &Handler{
		sellPatentsAction: sellPatentsAction,
		sessionFactory:    sessionFactory,
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := logger.Get().With(
		zap.String("connection_id", connection.ID),
		zap.String("player_id", connection.PlayerID),
		zap.String("game_id", connection.GameID))

	log.Debug("🏛️ Processing sell patents action")

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

	// Execute the sell patents action (Phase 1: initiate card selection)
	err := h.sellPatentsAction.Execute(ctx, sess, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute sell patents action", zap.Error(err))
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": err.Error()},
		}
		return
	}

	log.Info("✅ Sell patents action completed successfully - awaiting card selection")
}
