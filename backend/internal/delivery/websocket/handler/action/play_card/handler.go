package play_card

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Handler handles play card action requests
type Handler struct {
	cardService  service.CardService
	parser       *utils.MessageParser
	errorHandler *utils.ErrorHandler
	logger       *zap.Logger
}

// NewHandler creates a new play card handler
func NewHandler(cardService service.CardService, parser *utils.MessageParser) *Handler {
	return &Handler{
		cardService:  cardService,
		parser:       parser,
		errorHandler: utils.NewErrorHandler(),
		logger:       logger.Get(),
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

	// Execute the play card action
	err := h.cardService.OnPlayCard(ctx, gameID, playerID, request.CardID)
	if err != nil {
		h.logger.Warn("Failed to play card",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID),
			zap.String("card_id", request.CardID),
			zap.Error(err))
		h.errorHandler.SendError(connection, "Failed to play card: "+err.Error())
		return
	}

	h.logger.Info("✅ Card played successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("card_id", request.CardID))

	// The CardService will publish game updated events, so we don't need to send a response here
	// The client will receive the updated game state via the game-updated event
}
