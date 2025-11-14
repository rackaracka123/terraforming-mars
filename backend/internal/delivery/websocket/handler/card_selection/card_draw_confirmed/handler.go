package card_draw_confirmed

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/actions/card_selection"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles card draw confirmation requests
type Handler struct {
	confirmCardDrawAction *card_selection.ConfirmCardDrawAction
	parser                *utils.MessageParser
	errorHandler          *utils.ErrorHandler
	logger                *zap.Logger
}

// NewHandler creates a new card draw confirmation handler
func NewHandler(confirmCardDrawAction *card_selection.ConfirmCardDrawAction, parser *utils.MessageParser) *Handler {
	return &Handler{
		confirmCardDrawAction: confirmCardDrawAction,
		parser:                parser,
		errorHandler:          utils.NewErrorHandler(),
		logger:                logger.Get(),
	}
}

// CardDrawConfirmedRequest represents the payload for card draw confirmation
type CardDrawConfirmedRequest struct {
	CardsToTake []string `json:"cardsToTake" ts:"string[]"` // Card IDs to take for free
	CardsToBuy  []string `json:"cardsToBuy" ts:"string[]"`  // Card IDs to buy (pay for each)
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Card draw confirmation received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🃏 Processing card draw confirmation",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Parse the action payload
	var request CardDrawConfirmedRequest
	if err := h.parser.ParsePayload(message.Payload, &request); err != nil {
		h.logger.Error("Failed to parse card draw confirmation payload",
			zap.Error(err),
			zap.String("player_id", playerID))
		h.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	// Execute the card draw confirmation via ConfirmCardDrawAction
	if err := h.confirmCardDrawAction.Execute(ctx, gameID, playerID, request.CardsToTake, request.CardsToBuy); err != nil {
		h.logger.Error("Failed to process card draw confirmation via action",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID),
			zap.Int("cards_to_take_count", len(request.CardsToTake)),
			zap.Int("cards_to_buy_count", len(request.CardsToBuy)))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Card draw confirmation completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.Int("cards_taken", len(request.CardsToTake)),
		zap.Int("cards_bought", len(request.CardsToBuy)))
}
