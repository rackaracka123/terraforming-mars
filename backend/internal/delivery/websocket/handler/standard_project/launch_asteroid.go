package standard_project

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// LaunchAsteroidHandler handles launch asteroid standard project requests
type LaunchAsteroidHandler struct {
	action *action.LaunchAsteroidAction
	logger *zap.Logger
}

// NewLaunchAsteroidHandler creates a new launch asteroid handler
func NewLaunchAsteroidHandler(action *action.LaunchAsteroidAction) *LaunchAsteroidHandler {
	return &LaunchAsteroidHandler{
		action: action,
		logger: logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *LaunchAsteroidHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("☄️ Processing launch asteroid request (migrated)")

	// Use connection's gameID and playerID
	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	// Execute the action
	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute launch asteroid action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Launch asteroid action completed successfully")

	// Send success response
	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "launch-asteroid",
			"success": true,
		},
	}

	connection.Send <- response
	// Note: BroadcastEvent is published by the action, Broadcaster will handle game state updates
}

// sendError sends an error message to the client
func (h *LaunchAsteroidHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
