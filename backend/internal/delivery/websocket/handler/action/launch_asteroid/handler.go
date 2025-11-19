package launch_asteroid

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles launch asteroid standard project action requests
type Handler struct {
	launchAsteroidAction *action.LaunchAsteroidAction
}

// NewHandler creates a new launch asteroid handler
func NewHandler(launchAsteroidAction *action.LaunchAsteroidAction) *Handler {
	return &Handler{
		launchAsteroidAction: launchAsteroidAction,
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := logger.Get().With(
		zap.String("connection_id", connection.ID),
		zap.String("player_id", connection.PlayerID),
		zap.String("game_id", connection.GameID))

	log.Debug("🚀 Processing launch asteroid action")

	// Execute the asteroid standard project action
	err := h.launchAsteroidAction.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute launch asteroid action", zap.Error(err))
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": err.Error()},
		}
		return
	}

	log.Info("✅ Launch asteroid action completed successfully")
}
