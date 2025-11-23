package select_starting_card

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles select starting card action requests
type Handler struct {
	selectStartingCardsAction *action.SelectStartingCardsAction
	parser                    *utils.MessageParser
	errorHandler              *utils.ErrorHandler
	logger                    *zap.Logger
}

// NewHandler creates a new select starting card handler
func NewHandler(selectStartingCardsAction *action.SelectStartingCardsAction, parser *utils.MessageParser) *Handler {
	return &Handler{
		selectStartingCardsAction: selectStartingCardsAction,
		parser:                    parser,
		errorHandler:              utils.NewErrorHandler(),
		logger:                    logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Select starting card action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🃏 Processing select starting card action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Parse the action payload
	var request dto.ActionSelectStartingCardRequest
	if err := h.parser.ParsePayload(message.Payload, &request); err != nil {
		h.logger.Error("Failed to parse select starting card payload",
			zap.Error(err),
			zap.String("player_id", playerID))
		h.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	// Execute the action
	if err := h.handle(ctx, gameID, playerID, request.CardIDs, request.CorporationID); err != nil {
		h.logger.Error("Failed to select starting card",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	// Game state broadcasting is handled automatically by the SessionManager via events

	h.logger.Info("✅ Select starting card action completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

// handle processes the select starting card action (internal method)
func (h *Handler) handle(ctx context.Context, gameID, playerID string, cardIDs []string, corporationID string) error {
	h.logCardSelection(gameID, playerID, cardIDs, corporationID)

	return h.selectCards(ctx, gameID, playerID, cardIDs, corporationID)
}

// logCardSelection logs the card and corporation selection attempt
func (h *Handler) logCardSelection(gameID, playerID string, cardIDs []string, corporationID string) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player selecting starting cards and corporation",
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)),
		zap.String("corporation_id", corporationID))
}

// selectCards processes the card and corporation selection through the action
func (h *Handler) selectCards(ctx context.Context, gameID, playerID string, cardIDs []string, corporationID string) error {
	// Execute the action directly - actions are orchestrators
	return h.selectStartingCardsAction.Execute(ctx, gameID, playerID, cardIDs, corporationID)
}
