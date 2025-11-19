package convert_heat_to_temperature

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles convert heat to temperature action requests
type Handler struct {
	convertHeatAction *action.ConvertHeatToTemperatureAction
}

// NewHandler creates a new convert heat to temperature handler
func NewHandler(convertHeatAction *action.ConvertHeatToTemperatureAction) *Handler {
	return &Handler{
		convertHeatAction: convertHeatAction,
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := logger.Get().With(
		zap.String("connection_id", connection.ID),
		zap.String("player_id", connection.PlayerID),
		zap.String("game_id", connection.GameID))

	log.Debug("🔥 Processing convert heat to temperature action")

	// Execute the convert heat to temperature action
	err := h.convertHeatAction.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute convert heat to temperature action", zap.Error(err))
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": err.Error()},
		}
		return
	}

	log.Info("✅ Convert heat to temperature action completed successfully")
}
