package tile_selected

import (
	"context"

	"terraforming-mars-backend/internal/actions"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles tile selected requests
type Handler struct {
	selectTileAction *actions.SelectTileAction
	parser           *utils.MessageParser
	errorHandler     *utils.ErrorHandler
	logger           *zap.Logger
}

// NewHandler creates a new tile selected handler
func NewHandler(selectTileAction *actions.SelectTileAction, parser *utils.MessageParser) *Handler {
	return &Handler{
		selectTileAction: selectTileAction,
		parser:           parser,
		errorHandler:     utils.NewErrorHandler(),
		logger:           logger.Get(),
	}
}

// TileSelectedRequest represents the payload for tile selection
type TileSelectedRequest struct {
	Coordinate tiles.HexPosition `json:"coordinate" ts:"HexPosition"`
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Tile selected action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🎯 Processing tile selected action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Parse the action payload
	var request TileSelectedRequest
	if err := h.parser.ParsePayload(message.Payload, &request); err != nil {
		h.logger.Error("Failed to parse tile selected payload",
			zap.Error(err),
			zap.String("player_id", playerID))
		h.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	// Execute the tile selection via SelectTileAction (validates coordinates internally)
	if err := h.selectTileAction.Execute(ctx, gameID, playerID, request.Coordinate); err != nil {
		h.logger.Error("Failed to process tile selection via SelectTileAction",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID),
			zap.Int("q", request.Coordinate.Q),
			zap.Int("r", request.Coordinate.R),
			zap.Int("s", request.Coordinate.S))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Tile selected action completed successfully via SelectTileAction",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.Int("q", request.Coordinate.Q),
		zap.Int("r", request.Coordinate.R),
		zap.Int("s", request.Coordinate.S))
}
