package actions

import (
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
)

// SelectStartingCardsHandler handles starting card selection actions
type SelectStartingCardsHandler struct{}

// Handle applies the select starting card action
func (h *SelectStartingCardsHandler) Handle(game *domain.Game, player *domain.Player, actionPayload dto.ActionPayload) error {
	action := dto.SelectStartingCardAction{
		Type:    actionPayload.Type,
		CardIDs: actionPayload.CardIDs,
	}
	return h.applySelectStartingCard(game, player, action)
}

// applySelectStartingCard applies starting card selection
func (h *SelectStartingCardsHandler) applySelectStartingCard(game *domain.Game, player *domain.Player, action dto.SelectStartingCardAction) error {
	// Validate game phase
	if game.CurrentPhase != domain.GamePhaseStartingCardSelection {
		return fmt.Errorf("not in starting card selection phase")
	}

	// Check if player has already selected cards
	if len(player.Cards) > 0 {
		return fmt.Errorf("starting cards already selected")
	}

	// Get available starting cards
	availableCards := domain.GetStartingCards()
	availableCardMap := make(map[string]domain.Card)
	for _, card := range availableCards {
		availableCardMap[card.ID] = card
	}
	
	// Validate all selected cards exist and calculate total cost
	totalCost := 0
	
	for _, cardID := range action.CardIDs {
		card, exists := availableCardMap[cardID]
		if !exists {
			return fmt.Errorf("invalid starting card ID: %s", cardID)
		}
		totalCost += card.Cost
	}

	// Check if player has enough credits
	if player.Resources.Credits < totalCost {
		return fmt.Errorf("insufficient credits: need %d, have %d", totalCost, player.Resources.Credits)
	}

	// Pay for the cards
	player.Resources.Credits -= totalCost

	// Add cards to player's hand
	player.Cards = action.CardIDs

	// Check if all players have selected their starting cards
	allSelected := true
	for i := range game.Players {
		if len(game.Players[i].Cards) == 0 {
			allSelected = false
			break
		}
	}

	// If all players have selected, move to next phase
	if allSelected {
		game.CurrentPhase = domain.GamePhaseCorporationSelection
	}

	return nil
}