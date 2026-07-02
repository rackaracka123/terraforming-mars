package game

import (
	"context"
	"encoding/json"
	"log/slog"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
)

// SelectDemoChoicesHandler handles demo lobby card selection requests
type SelectDemoChoicesHandler struct {
	action      *gameaction.SelectDemoChoicesAction
	broadcaster Broadcaster
	logger      *slog.Logger
}

// NewSelectDemoChoicesHandler creates a new select demo choices handler
func NewSelectDemoChoicesHandler(action *gameaction.SelectDemoChoicesAction, broadcaster Broadcaster) *SelectDemoChoicesHandler {
	return &SelectDemoChoicesHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *SelectDemoChoicesHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing select demo choices request")

	if connection.GameID == "" || connection.PlayerID == "" {
		h.sendError(connection, "Not connected to a game")
		return
	}

	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		log.Error("Failed to marshal payload", slog.Any("error", err))
		h.sendError(connection, "Invalid payload format")
		return
	}

	var request dto.SelectDemoChoicesRequest
	if err := json.Unmarshal(payloadBytes, &request); err != nil {
		log.Error("Failed to unmarshal payload", slog.Any("error", err))
		h.sendError(connection, "Invalid payload format")
		return
	}

	err = h.action.Execute(ctx, connection.GameID, connection.PlayerID, &request)
	if err != nil {
		log.Error("Failed to execute select demo choices action", slog.Any("error", err))
		h.sendError(connection, err.Error())
		return
	}

	h.broadcaster.BroadcastGameState(connection.GameID, nil)

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "select-demo-choices",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *SelectDemoChoicesHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
