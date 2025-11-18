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

// Handler handles skip action requests (LEGACY - uses OLD service)
type Handler struct {
	gameService  service.GameService
	errorHandler *utils.ErrorHandler
	logger       *zap.Logger
}

// NewHandler creates a new skip action handler (LEGACY)
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

	h.logger.Debug("⏭️ Processing skip action [LEGACY]",
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

	h.logger.Info("✅ Skip action completed successfully [LEGACY]",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

// handle processes the skip action (internal method)
func (h *Handler) handle(ctx context.Context, gameID, playerID string) error {
	// Let the service handle all validation and business logic
	return h.gameService.SkipPlayerTurn(ctx, gameID, playerID)
}

// ========== NEW ARCHITECTURE: Action Pattern ==========

// ActionHandler handles skip action using the new action pattern
type ActionHandler struct {
	skipActionAction SkipActionAction
	errorHandler     *utils.ErrorHandler
	logger           *zap.Logger
}

// SkipActionAction interface for dependency injection
type SkipActionAction interface {
	Execute(ctx context.Context, gameID string, playerID string) error
}

// NewHandlerWithAction creates a new skip action handler using action pattern
func NewHandlerWithAction(skipActionAction SkipActionAction) *ActionHandler {
	return &ActionHandler{
		skipActionAction: skipActionAction,
		errorHandler:     utils.NewErrorHandler(),
		logger:           logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *ActionHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Skip action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("⏭️ Processing skip action [NEW ARCHITECTURE]",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Use action pattern - direct orchestration
	if err := h.skipActionAction.Execute(ctx, gameID, playerID); err != nil {
		h.logger.Error("Failed to skip action via action",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Skip action completed successfully [NEW ARCHITECTURE]",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}
