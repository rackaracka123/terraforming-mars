package skip_action

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Handler handles skip action requests
type Handler struct {
	gameService  service.GameService
	errorHandler *utils.ErrorHandler
	logger       *zap.Logger
}

// NewHandler creates a new skip action handler
func NewHandler(gameService service.GameService) *Handler {
	return &Handler{
		gameService:  gameService,
		errorHandler: utils.NewErrorHandler(),
		logger:       logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Skip action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("⏭️ Processing skip action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Skip action doesn't need payload parsing - it's a simple action
	if err := h.handle(ctx, gameID, playerID); err != nil {
		h.logger.Error("Failed to skip action",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Skip action completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

// handle processes the skip action (internal method)
func (h *Handler) handle(ctx context.Context, gameID, playerID string) error {
	// Let the service handle all validation and business logic
	return h.gameService.SkipPlayerTurn(ctx, gameID, playerID)
}
