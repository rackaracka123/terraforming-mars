package tile_selected

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// Handler handles tile selected requests
type Handler struct {
	selectTileAction *action.SelectTileAction
	sessionFactory   session.SessionFactory
	parser           *utils.MessageParser
	errorHandler     *utils.ErrorHandler
	logger           *zap.Logger
}

// NewHandler creates a new tile selected handler
func NewHandler(selectTileAction *action.SelectTileAction, sessionFactory session.SessionFactory, parser *utils.MessageParser) *Handler {
	return &Handler{
		selectTileAction: selectTileAction,
		sessionFactory:   sessionFactory,
		parser:           parser,
		errorHandler:     utils.NewErrorHandler(),
		logger:           logger.Get(),
	}
}

// TileSelectedRequest represents the payload for tile selection
type TileSelectedRequest struct {
	Coordinate types.HexPosition `json:"coordinate"`
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

	// Validate hex coordinates (must satisfy q + r + s = 0)
	if request.Coordinate.Q+request.Coordinate.R+request.Coordinate.S != 0 {
		h.logger.Error("Invalid hex coordinates",
			zap.String("player_id", playerID),
			zap.Int("q", request.Coordinate.Q),
			zap.Int("r", request.Coordinate.R),
			zap.Int("s", request.Coordinate.S))
		h.errorHandler.SendError(connection, "invalid hex coordinates: q+r+s must equal 0")
		return
	}

	// Get session for the game
	sess := h.sessionFactory.Get(gameID)
	if sess == nil {
		h.logger.Error("Session not found", zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": session not found")
		return
	}

	// Execute the tile selection using NEW action pattern
	if err := h.selectTileAction.Execute(ctx, sess, playerID, request.Coordinate.Q, request.Coordinate.R, request.Coordinate.S); err != nil {
		h.logger.Error("Failed to process tile selection",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID),
			zap.Int("q", request.Coordinate.Q),
			zap.Int("r", request.Coordinate.R),
			zap.Int("s", request.Coordinate.S))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Tile selected action completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.Int("q", request.Coordinate.Q),
		zap.Int("r", request.Coordinate.R),
		zap.Int("s", request.Coordinate.S))
}
