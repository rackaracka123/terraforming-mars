package select_cards

import (
	"context"

	"terraforming-mars-backend/internal/actions/card_selection"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles select cards action requests
// Delegates all business logic to SelectCardsAction
type Handler struct {
	selectCardsAction *card_selection.SelectCardsAction
	parser            *utils.MessageParser
	errorHandler      *utils.ErrorHandler
	logger            *zap.Logger
}

// NewHandler creates a new select cards handler
func NewHandler(
	selectCardsAction *card_selection.SelectCardsAction,
	parser *utils.MessageParser,
) *Handler {
	return &Handler{
		selectCardsAction: selectCardsAction,
		parser:            parser,
		errorHandler:      utils.NewErrorHandler(),
		logger:            logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Select cards action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🃏 Processing select cards action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Parse the action payload
	var request dto.ActionSelectProductionCardsRequest
	if err := h.parser.ParsePayload(message.Payload, &request); err != nil {
		h.logger.Error("Failed to parse select cards payload",
			zap.Error(err),
			zap.String("player_id", playerID))
		h.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	// Execute the select cards action (handles routing internally)
	if err := h.selectCardsAction.Execute(ctx, gameID, playerID, request.CardIDs); err != nil {
		h.logger.Error("Failed to select cards",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Select cards action completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}
