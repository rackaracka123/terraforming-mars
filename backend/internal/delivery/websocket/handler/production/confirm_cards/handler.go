package confirm_cards

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"

	"go.uber.org/zap"
)

// Handler handles confirm production cards action requests
type Handler struct {
	confirmProductionCardsAction *action.ConfirmProductionCardsAction
	sessionFactory               session.SessionFactory
	parser                       *utils.MessageParser
	errorHandler                 *utils.ErrorHandler
	logger                       *zap.Logger
}

// NewHandler creates a new confirm production cards handler
func NewHandler(confirmProductionCardsAction *action.ConfirmProductionCardsAction, sessionFactory session.SessionFactory, parser *utils.MessageParser) *Handler {
	return &Handler{
		confirmProductionCardsAction: confirmProductionCardsAction,
		sessionFactory:               sessionFactory,
		parser:                       parser,
		errorHandler:                 utils.NewErrorHandler(),
		logger:                       logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Confirm production cards action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🃏 Processing confirm production cards action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Parse the action payload
	var request dto.ActionSelectProductionCardsRequest
	if err := h.parser.ParsePayload(message.Payload, &request); err != nil {
		h.logger.Error("Failed to parse confirm production cards payload",
			zap.Error(err),
			zap.String("player_id", playerID))
		h.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	// Execute the action
	if err := h.handle(ctx, gameID, playerID, request.CardIDs); err != nil {
		h.logger.Error("Failed to confirm production cards",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Confirm production cards action completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

// handle processes the confirm production cards action (internal method)
func (h *Handler) handle(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	h.logCardSelection(gameID, playerID, cardIDs)

	return h.confirmCards(ctx, gameID, playerID, cardIDs)
}

// logCardSelection logs the card selection attempt
func (h *Handler) logCardSelection(gameID, playerID string, cardIDs []string) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player confirming production card selection",
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)))
}

// confirmCards processes the card confirmation through the action
func (h *Handler) confirmCards(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get session for the game
	sess := h.sessionFactory.Get(gameID)
	if sess == nil {
		log.Error("Session not found", zap.String("game_id", gameID))
		return fmt.Errorf("session not found: %s", gameID)
	}

	// Execute the action directly - actions are orchestrators
	return h.confirmProductionCardsAction.Execute(ctx, sess, playerID, cardIDs)
}
