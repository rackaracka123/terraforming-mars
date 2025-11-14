package convert_plants_to_greenery

import (
	"context"

	"terraforming-mars-backend/internal/actions"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles convert plants to greenery action requests
type Handler struct {
	convertPlantsAction *actions.ConvertPlantsToGreeneryAction
	errorHandler        *utils.ErrorHandler
	logger              *zap.Logger
}

// NewHandler creates a new convert plants to greenery handler
func NewHandler(convertPlantsAction *actions.ConvertPlantsToGreeneryAction) *Handler {
	return &Handler{
		convertPlantsAction: convertPlantsAction,
		errorHandler:        utils.NewErrorHandler(),
		logger:              logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Convert plants to greenery action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🌱 client→server: initiate plant conversion",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Execute plant conversion (deducts plants, raises oxygen, creates pending tile selection)
	if err := h.convertPlantsAction.Execute(ctx, gameID, playerID); err != nil {
		h.logger.Error("Failed to initiate plant conversion",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Plant conversion initiated successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}
