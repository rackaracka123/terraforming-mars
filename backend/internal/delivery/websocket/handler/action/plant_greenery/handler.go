package plant_greenery

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles plant greenery standard project action requests
type Handler struct {
	plantGreeneryAction *action.PlantGreeneryAction
}

// NewHandler creates a new plant greenery handler
func NewHandler(plantGreeneryAction *action.PlantGreeneryAction) *Handler {
	return &Handler{
		plantGreeneryAction: plantGreeneryAction,
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := logger.Get().With(
		zap.String("connection_id", connection.ID),
		zap.String("player_id", connection.PlayerID),
		zap.String("game_id", connection.GameID))

	log.Debug("🌱 Processing plant greenery action")

	// Execute the plant greenery action
	err := h.plantGreeneryAction.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute plant greenery action", zap.Error(err))
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": err.Error()},
		}
		return
	}

	log.Info("✅ Plant greenery action completed successfully")
}
