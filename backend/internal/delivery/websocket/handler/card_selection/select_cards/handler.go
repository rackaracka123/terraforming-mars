package select_cards

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Handler handles select production cards action requests
type Handler struct {
	cardService  service.CardService
	gameService  service.GameService
	parser       *utils.MessageParser
	errorHandler *utils.ErrorHandler
	logger       *zap.Logger
}

// NewHandler creates a new select cards handler
func NewHandler(cardService service.CardService, gameService service.GameService, parser *utils.MessageParser) *Handler {
	return &Handler{
		cardService:  cardService,
		gameService:  gameService,
		parser:       parser,
		errorHandler: utils.NewErrorHandler(),
		logger:       logger.Get(),
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
	log.Debug("Player selecting production cards",
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)))

	if err := h.selectCards(ctx, gameID, playerID, cardIDs); err != nil {
		return err
	}

	return h.markPlayerReady(ctx, gameID, playerID, cardIDs)
}

// selectCards processes the card selection through the service
func (h *Handler) selectCards(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)

	if err := h.cardService.SelectProductionCards(ctx, gameID, playerID, cardIDs); err != nil {
		log.Error("Failed to select production cards", zap.Error(err))
		return fmt.Errorf("card selection failed: %w", err)
	}

	return nil
}

// markPlayerReady marks the player as ready for production phase
func (h *Handler) markPlayerReady(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)

	updatedGame, err := h.gameService.ProcessProductionPhaseReady(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to process production phase ready", zap.Error(err))
		return fmt.Errorf("failed to process production phase ready: %w", err)
	}

	log.Info("Player completed production card selection and marked as ready",
		zap.Strings("selected_cards", cardIDs),
		zap.String("game_phase", string(updatedGame.CurrentPhase)))

	return nil
}
