package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// CardActions handles card-related actions
type CardActions struct {
	cardService service.CardService
	gameService service.GameService
	parser      *utils.MessageParser
	logger      *zap.Logger
}

// NewCardActions creates a new card actions handler
func NewCardActions(cardService service.CardService, gameService service.GameService, parser *utils.MessageParser) *CardActions {
	return &CardActions{
		cardService: cardService,
		gameService: gameService,
		parser:      parser,
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

// SelectProductionCards handles card selection during the game (i.e. after production)
func (ca *CardActions) SelectProductionCards(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionSelectProductionCardsRequest
	if err := ca.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid select card request: %w", err)
	}

	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player selecting production cards",
		zap.Strings("card_ids", request.CardIDs),
		zap.Int("count", len(request.CardIDs)))

	// Process the card selection through CardService
	if err := ca.cardService.SelectProductionCards(ctx, gameID, playerID, request.CardIDs); err != nil {
		log.Error("Failed to select production cards", zap.Error(err))
		return fmt.Errorf("card selection failed: %w", err)
	}

	// Mark player as ready for production phase and check if all players are ready
	updatedGame, err := ca.gameService.ProcessProductionPhaseReady(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to process production phase ready", zap.Error(err))
		return fmt.Errorf("failed to process production phase ready: %w", err)
	}

	log.Info("Player completed production card selection and marked as ready",
		zap.Strings("selected_cards", request.CardIDs),
		zap.String("game_phase", string(updatedGame.CurrentPhase)))

	return nil
}

// PlayCard handles playing a card with payment
func (ca *CardActions) PlayCard(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionPlayCardRequest
	if err := ca.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid play card request: %w", err)
	}

	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player playing card",
		zap.String("card_id", request.CardID),
		zap.Any("payment", request.Payment))

	// Convert DTO payment to model payment
	payment := &model.Payment{
		Credits:  request.Payment.Credits,
		Steel:    request.Payment.Steel,
		Titanium: request.Payment.Titanium,
	}

	// Process the card play through CardService
	if err := ca.cardService.PlayCard(ctx, gameID, playerID, request.CardID, payment); err != nil {
		log.Error("Failed to play card", zap.Error(err))
		return fmt.Errorf("card play failed: %w", err)
	}

	log.Info("Player played card successfully",
		zap.String("card_id", request.CardID))

	return nil
}
