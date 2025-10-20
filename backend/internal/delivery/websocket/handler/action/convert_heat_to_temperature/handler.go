package convert_heat_to_temperature

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Handler handles convert heat to temperature action requests
type Handler struct {
	resourceConversionService service.ResourceConversionService
	errorHandler              *utils.ErrorHandler
	logger                    *zap.Logger
}

// NewHandler creates a new convert heat to temperature handler
func NewHandler(resourceConversionService service.ResourceConversionService) *Handler {
	return &Handler{
		resourceConversionService: resourceConversionService,
		errorHandler:              utils.NewErrorHandler(),
		logger:                    logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Convert heat to temperature action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🔥 client→server: convert heat to temperature action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Execute conversion (no hex position needed for temperature)
	if err := h.resourceConversionService.ConvertHeatToTemperature(ctx, gameID, playerID); err != nil {
		h.logger.Error("Failed to convert heat to temperature",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Convert heat to temperature action processed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}
