package start_game

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles start game action requests
type Handler struct {
	lobbyService lobby.Service
	errorHandler *utils.ErrorHandler
	logger       *zap.Logger
}

// NewHandler creates a new start game handler
func NewHandler(lobbyService lobby.Service) *Handler {
	return &Handler{
		lobbyService: lobbyService,
		errorHandler: utils.NewErrorHandler(),
		logger:       logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Start game action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🚀 Processing start game action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Start game doesn't need payload parsing - it's a simple action
	if err := h.lobbyService.StartGame(ctx, gameID, playerID); err != nil {
		h.logger.Error("Failed to start game",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Start game action completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}
