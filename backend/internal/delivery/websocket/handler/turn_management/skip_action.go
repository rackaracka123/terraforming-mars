package turn_management

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// SkipActionHandler handles skip action requests
type SkipActionHandler struct {
	action      *action.SkipActionAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewSkipActionHandler creates a new skip action handler
func NewSkipActionHandler(action *action.SkipActionAction, broadcaster Broadcaster) *SkipActionHandler {
	return &SkipActionHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *SkipActionHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("⏭️ Processing skip action request (migrated)")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute skip action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Skip action completed successfully")

	// Explicitly broadcast game state after action completes
	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("📡 Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "skip-action",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *SkipActionHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
