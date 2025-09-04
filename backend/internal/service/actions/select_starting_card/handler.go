package select_starting_card

import (
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
)

// SelectStartingCardsHandler handles starting card selection actions
type SelectStartingCardsHandler struct{}

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

	// Get available starting cards
	availableCards := model.GetStartingCards()
	availableCardMap := make(map[string]model.Card)
	for _, card := range availableCards {
		availableCardMap[card.ID] = card
	}
	
	// Validate all selected cards exist and calculate total cost (3 MC per card)
	const costPerCard = 3
	
	for _, cardID := range action.CardIDs {
		_, exists := availableCardMap[cardID]
		if !exists {
			return fmt.Errorf("invalid starting card ID: %s", cardID)
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
		game.CurrentPhase = model.GamePhaseCorporationSelection
	}

	return nil
}