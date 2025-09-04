package select_starting_card

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	
	"go.uber.org/zap"
)

// SelectStartingCardsHandler handles starting card selection actions
type SelectStartingCardsHandler struct{
	cardSelectionRepo *repository.CardSelectionRepository
	eventRepository   *events.EventRepository
}

// NewSelectStartingCardsHandler creates a new select starting cards handler
func NewSelectStartingCardsHandler(cardSelectionRepo *repository.CardSelectionRepository, eventRepository *events.EventRepository) *SelectStartingCardsHandler {
	return &SelectStartingCardsHandler{
		cardSelectionRepo: cardSelectionRepo,
		eventRepository:   eventRepository,
	}
}

// Handle applies the select starting card action
func (h *SelectStartingCardsHandler) Handle(game *model.Game, player *model.Player, actionRequest dto.ActionSelectStartingCardRequest) error {
	action := actionRequest.GetAction()
	return h.applySelectStartingCard(game, player, *action)
}

// applySelectStartingCard applies starting card selection
func (h *SelectStartingCardsHandler) applySelectStartingCard(game *model.Game, player *model.Player, action dto.SelectStartingCardAction) error {
	// Validate game phase
	if game.CurrentPhase != model.GamePhaseStartingCardSelection {
		return fmt.Errorf("not in starting card selection phase")
	}

	// Check if player has already selected cards
	if len(player.Cards) > 0 {
		return fmt.Errorf("starting cards already selected")
	}

	// Get the player's available starting card options
	playerCardOptions, err := h.cardSelectionRepo.GetPlayerStartingCardOptions(game.ID, player.ID)
	if err != nil {
		return fmt.Errorf("failed to get player card options: %w", err)
	}

	// Validate all selected cards are from the player's available options
	availableCardSet := make(map[string]bool)
	for _, cardID := range playerCardOptions {
		availableCardSet[cardID] = true
	}
	
	// Validate all selected cards exist in player's options and calculate total cost (3 MC per card)
	const costPerCard = 3
	
	for _, cardID := range action.CardIDs {
		if !availableCardSet[cardID] {
			return fmt.Errorf("invalid starting card ID: %s (not in player's available options)", cardID)
		}
	}
	
	totalCost := len(action.CardIDs) * costPerCard

	// Check if player has enough credits
	if player.Resources.Credits < totalCost {
		return fmt.Errorf("insufficient credits: need %d, have %d", totalCost, player.Resources.Credits)
	}

	// Pay for the cards
	player.Resources.Credits -= totalCost

	// Add cards to player's hand
	player.Cards = action.CardIDs

	// Publish starting card selected event
	if h.eventRepository != nil {
		log := logger.WithGameContext(game.ID, player.ID)
		cardSelectedEvent := events.NewStartingCardSelectedEvent(game.ID, player.ID, action.CardIDs, totalCost)
		if err := h.eventRepository.Publish(context.Background(), cardSelectedEvent); err != nil {
			log.Warn("Failed to publish starting card selected event", zap.Error(err))
		}
	}

	// Mark player as having completed their starting card selection
	if err := h.cardSelectionRepo.MarkPlayerCompletedStartingCardSelection(game.ID, player.ID); err != nil {
		return fmt.Errorf("failed to mark player selection complete: %w", err)
	}

	// Check if all players have completed their starting card selection
	allSelected, err := h.cardSelectionRepo.AllPlayersCompletedStartingCardSelection(game.ID, len(game.Players))
	if err != nil {
		return fmt.Errorf("failed to check if all players selected: %w", err)
	}

	// If all players have selected, clean up selection data and move to next phase
	if allSelected {
		// Clean up the starting card selection data since it's no longer needed
		if err := h.cardSelectionRepo.DeleteStartingCardSelection(game.ID); err != nil {
			return fmt.Errorf("failed to cleanup starting card selection: %w", err)
		}
		game.CurrentPhase = model.GamePhaseCorporationSelection
	}

	return nil
}