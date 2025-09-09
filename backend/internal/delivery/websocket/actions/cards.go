package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// CardActions handles card-related actions
type CardActions struct {
	cardService service.CardService
	gameService service.GameService
	parser      *MessageParser
	logger      *zap.Logger
}

// NewCardActions creates a new card actions handler
func NewCardActions(cardService service.CardService, gameService service.GameService) *CardActions {
	return &CardActions{
		cardService: cardService,
		gameService: gameService,
		parser:      NewMessageParser(),
		logger:      logger.Get(),
	}
}

// SelectStartingCards handles starting card selection
func (ca *CardActions) SelectStartingCards(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionSelectStartingCardRequest
	if err := ca.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid select starting card request: %w", err)
	}

	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player selecting starting cards",
		zap.Strings("card_ids", request.CardIDs),
		zap.Int("count", len(request.CardIDs)))

	// Process the card selection through CardService
	if err := ca.cardService.SelectStartingCards(ctx, gameID, playerID, request.CardIDs); err != nil {
		log.Error("Failed to select starting cards", zap.Error(err))
		return fmt.Errorf("card selection failed: %w", err)
	}

	// Check if all players have completed their selection
	if ca.cardService.IsAllPlayersCardSelectionComplete(ctx, gameID) {
		log.Info("All players completed starting card selection, advancing game phase")

		// Advance game phase using proper GameService method
		if err := ca.gameService.AdvanceFromCardSelectionPhase(ctx, gameID); err != nil {
			log.Error("Failed to advance game phase", zap.Error(err))
			return fmt.Errorf("failed to advance game phase: %w", err)
		}

		log.Info("Game phase advanced to Action phase")
	}

	log.Info("Player completed starting card selection",
		zap.Strings("selected_cards", request.CardIDs))

	return nil
}
