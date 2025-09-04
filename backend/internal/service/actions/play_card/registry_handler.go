package play_card

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
)

// RegistryBasedPlayCardHandler uses the card handler registry for scalable card management
type RegistryBasedPlayCardHandler struct {
	cardRegistry *cards.CardHandlerRegistry
	eventBus     events.EventBus
}

// NewRegistryBasedPlayCardHandler creates a new registry-based card handler
func NewRegistryBasedPlayCardHandler(cardRegistry *cards.CardHandlerRegistry, eventBus events.EventBus) *RegistryBasedPlayCardHandler {
	return &RegistryBasedPlayCardHandler{
		cardRegistry: cardRegistry,
		eventBus:     eventBus,
	}
}

// Handle applies the play card action using the registry system
func (h *RegistryBasedPlayCardHandler) Handle(game *model.Game, player *model.Player, actionRequest dto.ActionPlayCardRequest) error {
	action := actionRequest.GetAction()
	return h.applyPlayCard(game, player, *action)
}

// applyPlayCard applies the card play action using registered handlers
func (h *RegistryBasedPlayCardHandler) applyPlayCard(game *model.Game, player *model.Player, action dto.PlayCardAction) error {
	// Validate game phase - only allow card play during action phase
	if game.CurrentPhase != model.GamePhaseAction {
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

	// Get the card details from the card data
	availableCards := model.GetStartingCards()
	var cardToPlay *model.Card
	for _, card := range availableCards {
		if card.ID == action.CardID {
			cardToPlay = &card
			break
		}
	}

	if cardToPlay == nil {
		return fmt.Errorf("invalid card ID: %s", action.CardID)
	}

	// Check if player has enough credits for base cost
	if player.Resources.Credits < cardToPlay.Cost {
		return fmt.Errorf("insufficient credits: need %d, have %d", cardToPlay.Cost, player.Resources.Credits)
	}

	// Get the card handler from registry
	cardHandler, err := h.cardRegistry.GetHandler(action.CardID)
	if err != nil {
		return fmt.Errorf("card handler not found: %w", err)
	}

	// Create context for card handler
	ctx := &cards.CardHandlerContext{
		Game:     game,
		Player:   player,
		Card:     cardToPlay,
		EventBus: h.eventBus,
	}

	// Check if card requirements are met and can be played
	if err := cardHandler.CanPlay(ctx); err != nil {
		return fmt.Errorf("card requirements not met: %w", err)
	}

	// Pay for the card
	player.Resources.Credits -= cardToPlay.Cost

	// Apply card effects through the handler
	if err := cardHandler.Play(ctx); err != nil {
		// Refund the cost if effect application fails
		player.Resources.Credits += cardToPlay.Cost
		return fmt.Errorf("failed to apply card effects: %w", err)
	}

	// Remove card from hand
	player.Cards = append(player.Cards[:cardIndex], player.Cards[cardIndex+1:]...)

	// Add card to played cards
	player.PlayedCards = append(player.PlayedCards, action.CardID)

	// Publish card played event
	if h.eventBus != nil {
		cardPlayedEvent := events.NewCardPlayedEvent(game.ID, player.ID, action.CardID)
		h.eventBus.Publish(context.Background(), cardPlayedEvent)
	}

	return nil
}