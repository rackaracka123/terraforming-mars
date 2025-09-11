package select_cards

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Handler handles select production cards action requests
type Handler struct {
	cardService service.CardService
	gameService service.GameService
	parser      *utils.MessageParser
	logger      *zap.Logger
}

// NewHandler creates a new select cards handler
func NewHandler(cardService service.CardService, gameService service.GameService, parser *utils.MessageParser) *Handler {
	return &Handler{
		cardService: cardService,
		gameService: gameService,
		parser:      parser,
		logger:      logger.Get(),
	}
}

// Handle processes the select production cards action
func (h *Handler) Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	request, err := h.parseRequest(actionRequest)
	if err != nil {
		return err
	}

	h.logCardSelection(gameID, playerID, request.CardIDs)

	if err := h.selectCards(ctx, gameID, playerID, request.CardIDs); err != nil {
		return err
	}

	return h.markPlayerReady(ctx, gameID, playerID, request.CardIDs)
}

// parseRequest parses and validates the action request
func (h *Handler) parseRequest(actionRequest interface{}) (dto.ActionSelectProductionCardsRequest, error) {
	var request dto.ActionSelectProductionCardsRequest
	if err := h.parser.ParsePayload(actionRequest, &request); err != nil {
		return request, fmt.Errorf("invalid select card request: %w", err)
	}
	return request, nil
}

// logCardSelection logs the card selection attempt
func (h *Handler) logCardSelection(gameID, playerID string, cardIDs []string) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player selecting production cards",
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)))
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
