package launch_asteroid

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Handler handles launch asteroid standard project action requests
type Handler struct {
	standardProjectService service.StandardProjectService
	errorHandler           *utils.ErrorHandler
	logger                 *zap.Logger
}

// NewHandler creates a new launch asteroid handler
func NewHandler(standardProjectService service.StandardProjectService) *Handler {
	return &Handler{
		standardProjectService: standardProjectService,
		errorHandler:           utils.NewErrorHandler(),
		logger:                 logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Launch asteroid action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🚀 Processing launch asteroid action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Launch asteroid doesn't need payload parsing - it's a simple action
	if err := h.standardProjectService.LaunchAsteroid(ctx, gameID, playerID); err != nil {
		h.logger.Error("Failed to launch asteroid",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Launch asteroid action completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}
