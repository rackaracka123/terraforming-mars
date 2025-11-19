package build_city

import (
	"context"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
)

// Handler handles build city standard project action requests
type Handler struct {
	buildCityAction *action.BuildCityAction
}

// NewHandler creates a new build city handler
func NewHandler(buildCityAction *action.BuildCityAction) *Handler {
	return &Handler{
		buildCityAction: buildCityAction,
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

	// Execute the build city action
	err := h.buildCityAction.Execute(ctx, connection.GameID, connection.PlayerID)
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
