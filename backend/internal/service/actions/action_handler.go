package actions

import (
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/service/actions/play_card"
	"terraforming-mars-backend/internal/service/actions/select_starting_card"
	"terraforming-mars-backend/internal/service/actions/start_game"
)

// ActionHandler defines the interface for handling game actions
// Note: Individual handlers now use specific request types

// ActionHandlers contains all action handlers
type ActionHandlers struct {
	SelectStartingCards *select_starting_card.SelectStartingCardsHandler
	StartGame          *start_game.StartGameHandler
	PlayCard           *play_card.PlayCardHandler
}

// NewActionHandlers creates a new instance of action handlers
func NewActionHandlers(eventBus events.EventBus) *ActionHandlers {
	return &ActionHandlers{
		SelectStartingCards: &select_starting_card.SelectStartingCardsHandler{},
		StartGame:          &start_game.StartGameHandler{EventBus: eventBus},
		PlayCard:           &play_card.PlayCardHandler{},
	}
}