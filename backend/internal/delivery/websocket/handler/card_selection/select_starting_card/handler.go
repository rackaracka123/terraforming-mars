package select_starting_card

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Handler handles select starting card action requests
type Handler struct {
	cardService service.CardService
	gameService service.GameService
	parser      *utils.MessageParser
	logger      *zap.Logger
}

// NewHandler creates a new select starting card handler
func NewHandler(cardService service.CardService, gameService service.GameService, parser *utils.MessageParser) *Handler {
	return &Handler{
		cardService: cardService,
		gameService: gameService,
		parser:      parser,
		logger:      logger.Get(),
	}
}

// Handle processes the select starting card action
func (h *Handler) Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	request, err := h.parseRequest(actionRequest)
	if err != nil {
		return err
	}

	h.logCardSelection(gameID, playerID, request.CardIDs)

	if err := h.selectCards(ctx, gameID, playerID, request.CardIDs); err != nil {
		return err
	}

	return h.checkAndAdvancePhase(ctx, gameID, playerID, request.CardIDs)
}

// parseRequest parses and validates the action request
func (h *Handler) parseRequest(actionRequest interface{}) (dto.ActionSelectStartingCardRequest, error) {
	var request dto.ActionSelectStartingCardRequest
	if err := h.parser.ParsePayload(actionRequest, &request); err != nil {
		return request, fmt.Errorf("invalid select starting card request: %w", err)
	}
	return request, nil
}

// logCardSelection logs the card selection attempt
func (h *Handler) logCardSelection(gameID, playerID string, cardIDs []string) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player selecting starting cards",
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)))
}

// selectCards processes the card selection through the service
func (h *Handler) selectCards(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)

	if err := h.cardService.SelectStartingCards(ctx, gameID, playerID, cardIDs); err != nil {
		log.Error("Failed to select starting cards", zap.Error(err))
		return fmt.Errorf("card selection failed: %w", err)
	}

	return nil
}

// checkAndAdvancePhase checks if all players completed selection and advances phase if needed
func (h *Handler) checkAndAdvancePhase(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)

	if h.cardService.IsAllPlayersCardSelectionComplete(ctx, gameID) {
		log.Info("All players completed starting card selection, advancing game phase")

		if err := h.gameService.AdvanceFromCardSelectionPhase(ctx, gameID); err != nil {
			log.Error("Failed to advance game phase", zap.Error(err))
			return fmt.Errorf("failed to advance game phase: %w", err)
		}

		log.Info("Game phase advanced to Action phase")
	}

	log.Info("Player completed starting card selection",
		zap.Strings("selected_cards", cardIDs))

	return nil
}
