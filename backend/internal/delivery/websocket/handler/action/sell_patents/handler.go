package sell_patents

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Handler handles sell patents standard project action requests
type Handler struct {
	standardProjectService service.StandardProjectService
	errorHandler           *utils.ErrorHandler
	logger                 *zap.Logger
}

// NewHandler creates a new sell patents handler
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
		h.logger.Warn("Sell patents action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🏛️ Processing sell patents action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Execute the action (no payload needed - initiates card selection)
	if err := h.standardProjectService.InitiateSellPatents(ctx, gameID, playerID); err != nil {
		h.logger.Error("Failed to initiate sell patents",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Sell patents initiated, awaiting card selection",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}
