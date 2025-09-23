package handlers

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// GameHandler handles game-level actions (start game)
type GameHandler struct {
	gameService  service.GameService
	errorHandler *utils.ErrorHandler
	logger       *zap.Logger
}

// NewGameHandler creates a new game handler
func NewGameHandler(gameService service.GameService) *GameHandler {
	return &GameHandler{
		gameService:  gameService,
		errorHandler: utils.NewErrorHandler(),
		logger:       logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *GameHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Game action received from unassigned connection",
			zap.String("connection_id", connection.ID),
			zap.String("message_type", string(message.Type)))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	switch message.Type {
	case dto.MessageTypeActionStartGame:
		h.handleStartGame(ctx, gameID, playerID, connection)
	default:
		h.logger.Warn("Unknown game action type", zap.String("type", string(message.Type)))
		h.errorHandler.SendError(connection, "Unknown game action type")
	}
}

func (h *GameHandler) handleStartGame(ctx context.Context, gameID, playerID string, connection *core.Connection) {
	h.logger.Debug("🚀 Processing start game action", zap.String("player_id", playerID))

	if err := h.gameService.StartGame(ctx, gameID, playerID); err != nil {
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