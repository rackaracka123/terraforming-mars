package select_cards

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// Handler handles select production cards action requests
type Handler struct {
	confirmSellPatentsAction     *action.ConfirmSellPatentsAction
	confirmProductionCardsAction *action.ConfirmProductionCardsAction
	newPlayerRepo                player.Repository
	parser                       *utils.MessageParser
	errorHandler                 *utils.ErrorHandler
	logger                       *zap.Logger
}

// NewHandler creates a new select cards handler
func NewHandler(confirmSellPatentsAction *action.ConfirmSellPatentsAction, confirmProductionCardsAction *action.ConfirmProductionCardsAction, newPlayerRepo player.Repository, parser *utils.MessageParser) *Handler {
	return &Handler{
		confirmSellPatentsAction:     confirmSellPatentsAction,
		confirmProductionCardsAction: confirmProductionCardsAction,
		newPlayerRepo:                newPlayerRepo,
		parser:                       parser,
		errorHandler:                 utils.NewErrorHandler(),
		logger:                       logger.Get(),
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

	// Execute the action
	if err := h.handle(ctx, gameID, playerID, request.CardIDs); err != nil {
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

// handle processes the select production cards action (internal method)
func (h *Handler) handle(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player selecting cards",
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)))

	// Check if player has a pending card selection (e.g., sell patents) using NEW repository
	p, err := h.newPlayerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for pending card selection check", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// If there's a pending card selection, route to ConfirmSellPatentsAction
	if p.PendingCardSelection != nil {
		// Save source before action clears the pending selection
		source := p.PendingCardSelection.Source

		log.Info("Processing pending card selection",
			zap.String("source", source),
			zap.Int("cards_selected", len(cardIDs)))

		// Use ConfirmSellPatentsAction instead of service
		if err := h.confirmSellPatentsAction.Execute(ctx, gameID, playerID, cardIDs); err != nil {
			return err
		}

		log.Info("✅ Pending card selection completed",
			zap.String("source", source))
		return nil
	}

	// Otherwise, handle as production card selection using ConfirmProductionCardsAction
	log.Debug("Processing production card selection")
	if err := h.confirmProductionCardsAction.Execute(ctx, gameID, playerID, cardIDs); err != nil {
		return err
	}

	log.Info("✅ Production card selection completed",
		zap.Strings("selected_cards", cardIDs))

	return nil
}
