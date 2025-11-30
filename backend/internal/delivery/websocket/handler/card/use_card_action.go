package card

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// UseCardActionHandler handles card action execution requests
type UseCardActionHandler struct {
	action *action.UseCardActionAction
	logger *zap.Logger
}

// NewUseCardActionHandler creates a new use card action handler
func NewUseCardActionHandler(action *action.UseCardActionAction) *UseCardActionHandler {
	return &UseCardActionHandler{
		action: action,
		logger: logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *UseCardActionHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🎯 Processing use card action request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	// Parse payload
	payload, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "Invalid payload format")
		return
	}

	// Extract cardId
	cardID, ok := payload["cardId"].(string)
	if !ok || cardID == "" {
		log.Error("Missing or invalid cardId")
		h.sendError(connection, "Missing cardId")
		return
	}

	// Extract behaviorIndex
	behaviorIndexFloat, ok := payload["behaviorIndex"].(float64)
	if !ok {
		log.Error("Missing or invalid behaviorIndex")
		h.sendError(connection, "Missing behaviorIndex")
		return
	}
	behaviorIndex := int(behaviorIndexFloat)

	log = log.With(
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex),
	)

	// Execute the action
	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, cardID, behaviorIndex)
	if err != nil {
		log.Error("Failed to execute use card action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Use card action completed successfully")

	// Send success response
	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":        "card-action",
			"success":       true,
			"cardId":        cardID,
			"behaviorIndex": behaviorIndex,
		},
	}

	connection.Send <- response
}

func (h *UseCardActionHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
