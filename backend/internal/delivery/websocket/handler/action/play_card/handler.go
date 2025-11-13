package play_card

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/game/actions"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles play card action requests
type Handler struct {
	playCardAction *actions.PlayCardAction
	parser         *utils.MessageParser
	errorHandler   *utils.ErrorHandler
	logger         *zap.Logger
}

// NewHandler creates a new play card handler
func NewHandler(playCardAction *actions.PlayCardAction, parser *utils.MessageParser) *Handler {
	return &Handler{
		playCardAction: playCardAction,
		parser:         parser,
		errorHandler:   utils.NewErrorHandler(),
		logger:         logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Play card action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🎯 Processing play card action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Parse the action payload
	var request dto.ActionPlayCardRequest
	if err := h.parser.ParsePayload(message.Payload, &request); err != nil {
		h.logger.Warn("Failed to parse play card request payload",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID),
			zap.Error(err))
		h.errorHandler.SendError(connection, "Invalid play card request format")
		return
	}

	h.logger.Debug("📋 Play card request parsed",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("card_id", request.CardID))

	// Validate card ID is provided
	if request.CardID == "" {
		h.logger.Warn("Play card request missing card ID",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID))
		h.errorHandler.SendError(connection, "Card ID is required")
		return
	}

	// Convert DTO payment to domain model
	payment := dto.ToCardPayment(request.Payment)

	h.logger.Debug("💰 Payment breakdown",
		zap.String("connection_id", connection.ID),
		zap.String("card_id", request.CardID),
		zap.Int("credits", payment.Credits),
		zap.Int("steel", payment.Steel),
		zap.Int("titanium", payment.Titanium))

	// Execute the play card action with payment, optional choice index, and card storage target
	err := h.playCardAction.Execute(ctx, gameID, playerID, request.CardID, &payment, request.ChoiceIndex, request.CardStorageTarget)
	if err != nil {
		h.logger.Warn("Failed to play card via PlayCardAction",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID),
			zap.String("card_id", request.CardID),
			zap.Error(err))
		h.errorHandler.SendError(connection, "Failed to play card: "+err.Error())
		return
	}

	h.logger.Info("✅ Card played successfully via PlayCardAction",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("card_id", request.CardID))

	// The PlayCardAction will broadcast game state, so we don't need to send a response here
	// The client will receive the updated game state via the game-updated event
}
