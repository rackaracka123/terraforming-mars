package actions

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/interfaces"
	"terraforming-mars-backend/internal/repository"
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
func NewActionHandlers(eventBus events.EventBus, eventRepository *events.EventRepository, cardRegistry *cards.CardHandlerRegistry, cardSelectionRepo *repository.CardSelectionRepository, playerService interfaces.PlayerService) *ActionHandlers {
	var playCardHandler PlayCardHandler
	
	// Use registry-based handler if registry is provided, otherwise fall back to legacy
	if cardRegistry != nil {
		playCardHandler = play_card.NewRegistryBasedPlayCardHandler(cardRegistry, eventBus, playerService)
	} else {
		playCardHandler = &play_card.PlayCardHandler{}
	}
	
	return &ActionHandlers{
		SelectStartingCards: select_starting_card.NewSelectStartingCardsHandler(cardSelectionRepo, eventRepository),
		StartGame:          start_game.NewStartGameHandler(cardSelectionRepo, eventRepository),
		PlayCard:           playCardHandler,
	}
}