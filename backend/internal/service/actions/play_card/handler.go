package play_card

import (
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/domain"
)

// PlayCardHandler handles card play actions
type PlayCardHandler struct{}

// Handle applies the play card action
func (h *PlayCardHandler) Handle(game *domain.Game, player *domain.Player, actionRequest dto.ActionPlayCardRequest) error {
	action := actionRequest.GetAction()
	return h.applyPlayCard(game, player, *action)
}

// applyPlayCard applies the card play action
func (h *PlayCardHandler) applyPlayCard(game *domain.Game, player *domain.Player, action dto.PlayCardAction) error {
	// Validate game phase - only allow card play during action phase
	if game.CurrentPhase != domain.GamePhaseAction {
		return fmt.Errorf("cannot play cards outside of action phase")
	}

	// Check if card exists in player's hand
	cardIndex := -1
	for i, cardID := range player.Cards {
		if cardID == action.CardID {
			cardIndex = i
			break
		}
	}
	
	if cardIndex == -1 {
		return fmt.Errorf("card %s not found in player's hand", action.CardID)
	}

	// Get the card details to check cost
	availableCards := domain.GetStartingCards()
	var cardToPlay *domain.Card
	for _, card := range availableCards {
		if card.ID == action.CardID {
			cardToPlay = &card
			break
		}
	}

	if cardToPlay == nil {
		return fmt.Errorf("invalid card ID: %s", action.CardID)
	}

	// Check if player has enough credits
	if player.Resources.Credits < cardToPlay.Cost {
		return fmt.Errorf("insufficient credits: need %d, have %d", cardToPlay.Cost, player.Resources.Credits)
	}

	// Pay for the card
	player.Resources.Credits -= cardToPlay.Cost

	// Apply card effects based on card ID
	if err := h.applyCardEffects(game, player, *cardToPlay); err != nil {
		// Refund the cost if effect application fails
		player.Resources.Credits += cardToPlay.Cost
		return fmt.Errorf("failed to apply card effects: %w", err)
	}

	// Remove card from hand
	player.Cards = append(player.Cards[:cardIndex], player.Cards[cardIndex+1:]...)

	// Add card to played cards
	player.PlayedCards = append(player.PlayedCards, action.CardID)

	return nil
}

// applyCardEffects delegates to specific card handlers
func (h *PlayCardHandler) applyCardEffects(game *domain.Game, player *domain.Player, card domain.Card) error {
	switch card.ID {
	case "early-settlement":
		return h.handleEarlySettlement(game, player)
	case "power-plant":
		return h.handlePowerPlant(game, player)
	case "heat-generators":
		return h.handleHeatGenerators(game, player)
	case "mining-operation":
		return h.handleMiningOperation(game, player)
	case "space-mirrors":
		return h.handleSpaceMirrors(game, player)
	case "water-import":
		return h.handleWaterImport(game, player)
	case "nitrogen-plants":
		return h.handleNitrogenPlants(game, player)
	case "atmospheric-processors":
		return h.handleAtmosphericProcessors(game, player)
	default:
		return fmt.Errorf("card %s is not yet implemented", card.ID)
	}
}

// Helper functions for common operations
func (h *PlayCardHandler) increaseProduction(player *domain.Player, resource string, amount int) {
	switch resource {
	case "credits":
		player.Production.Credits += amount
	case "steel":
		player.Production.Steel += amount
	case "titanium":
		player.Production.Titanium += amount
	case "plants":
		player.Production.Plants += amount
	case "energy":
		player.Production.Energy += amount
	case "heat":
		player.Production.Heat += amount
	}
}

func (h *PlayCardHandler) increaseResource(player *domain.Player, resource string, amount int) {
	switch resource {
	case "credits":
		player.Resources.Credits += amount
	case "steel":
		player.Resources.Steel += amount
	case "titanium":
		player.Resources.Titanium += amount
	case "plants":
		player.Resources.Plants += amount
	case "energy":
		player.Resources.Energy += amount
	case "heat":
		player.Resources.Heat += amount
	}
}

func (h *PlayCardHandler) increaseGlobalParameter(game *domain.Game, parameter string, amount int) error {
	switch parameter {
	case "temperature":
		if game.GlobalParameters.Temperature < 8 {
			game.GlobalParameters.Temperature += amount
			if game.GlobalParameters.Temperature > 8 {
				game.GlobalParameters.Temperature = 8
			}
		}
	case "oxygen":
		if game.GlobalParameters.Oxygen < 14 {
			game.GlobalParameters.Oxygen += amount
			if game.GlobalParameters.Oxygen > 14 {
				game.GlobalParameters.Oxygen = 14
			}
		}
	case "oceans":
		if game.GlobalParameters.Oceans < 9 {
			game.GlobalParameters.Oceans += amount
			if game.GlobalParameters.Oceans > 9 {
				game.GlobalParameters.Oceans = 9
			}
		}
	default:
		return fmt.Errorf("unknown global parameter: %s", parameter)
	}
	return nil
}

// Individual card handlers
func (h *PlayCardHandler) handleEarlySettlement(game *domain.Game, player *domain.Player) error {
	h.increaseProduction(player, "credits", 1)
	return nil
}

func (h *PlayCardHandler) handlePowerPlant(game *domain.Game, player *domain.Player) error {
	h.increaseProduction(player, "energy", 1)
	return nil
}

func (h *PlayCardHandler) handleHeatGenerators(game *domain.Game, player *domain.Player) error {
	h.increaseProduction(player, "heat", 1)
	return nil
}

func (h *PlayCardHandler) handleMiningOperation(game *domain.Game, player *domain.Player) error {
	h.increaseResource(player, "steel", 2)
	return nil
}

func (h *PlayCardHandler) handleSpaceMirrors(game *domain.Game, player *domain.Player) error {
	// Active card - no immediate effect, just moves to played cards
	return nil
}

func (h *PlayCardHandler) handleWaterImport(game *domain.Game, player *domain.Player) error {
	return h.increaseGlobalParameter(game, "oceans", 1)
}

func (h *PlayCardHandler) handleNitrogenPlants(game *domain.Game, player *domain.Player) error {
	h.increaseProduction(player, "plants", 1)
	return nil
}

func (h *PlayCardHandler) handleAtmosphericProcessors(game *domain.Game, player *domain.Player) error {
	return h.increaseGlobalParameter(game, "oxygen", 1)
}