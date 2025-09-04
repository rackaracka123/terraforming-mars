package actions

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/service/actions/play_card"
	"terraforming-mars-backend/internal/service/actions/select_starting_card"
	"terraforming-mars-backend/internal/service/actions/start_game"
)

// ActionHandler defines the interface for handling game actions
// Note: Individual handlers now use specific request types

// PlayCardHandler interface allows switching between implementations
type PlayCardHandler interface {
	Handle(game *model.Game, player *model.Player, actionRequest dto.ActionPlayCardRequest) error
}

// ActionHandlers contains all action handlers
type ActionHandlers struct {
	SelectStartingCards *select_starting_card.SelectStartingCardsHandler
	StartGame          *start_game.StartGameHandler
	PlayCard           PlayCardHandler
}

// NewActionHandlers creates a new instance of action handlers
func NewActionHandlers(eventBus events.EventBus, cardRegistry *cards.CardHandlerRegistry) *ActionHandlers {
	var playCardHandler PlayCardHandler
	
	// Use registry-based handler if registry is provided, otherwise fall back to legacy
	if cardRegistry != nil {
		playCardHandler = play_card.NewRegistryBasedPlayCardHandler(cardRegistry, eventBus)
	} else {
		playCardHandler = &play_card.PlayCardHandler{}
	}
	
	return &ActionHandlers{
		SelectStartingCards: &select_starting_card.SelectStartingCardsHandler{},
		StartGame:          &start_game.StartGameHandler{EventBus: eventBus},
		PlayCard:           playCardHandler,
	}
}