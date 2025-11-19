package convert_plants_to_greenery

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles convert plants to greenery action requests
type Handler struct {
	convertPlantsAction *action.ConvertPlantsToGreeneryAction
}

// NewHandler creates a new convert plants to greenery handler
func NewHandler(convertPlantsAction *action.ConvertPlantsToGreeneryAction) *Handler {
	return &Handler{
		convertPlantsAction: convertPlantsAction,
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := logger.Get().With(
		zap.String("connection_id", connection.ID),
		zap.String("player_id", connection.PlayerID),
		zap.String("game_id", connection.GameID))

	log.Debug("🌱 Processing convert plants to greenery action")

	// Execute the convert plants to greenery action
	err := h.convertPlantsAction.Execute(ctx, connection.GameID, connection.PlayerID)
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
