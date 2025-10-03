package build_city

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Handler handles build city standard project action requests
type Handler struct {
	standardProjectService service.StandardProjectService
	parser                 *utils.MessageParser
	errorHandler           *utils.ErrorHandler
	logger                 *zap.Logger
}

// NewHandler creates a new build city handler
func NewHandler(standardProjectService service.StandardProjectService, parser *utils.MessageParser) *Handler {
	return &Handler{
		standardProjectService: standardProjectService,
		parser:                 parser,
		errorHandler:           utils.NewErrorHandler(),
		logger:                 logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Build city action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🏢 Processing build city action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Execute the action (no hex position needed, tile queue will be created)
	if err := h.standardProjectService.BuildCity(ctx, gameID, playerID); err != nil {
		h.logger.Error("Failed to build city",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Build city action completed successfully, tile queued for placement",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}
