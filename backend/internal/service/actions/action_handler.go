package actions

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/domain"
)

// ActionHandler defines the interface for handling game actions
type ActionHandler interface {
	Handle(game *domain.Game, player *domain.Player, actionPayload dto.ActionPayload) error
}

// ActionHandlers contains all action handlers
type ActionHandlers struct {
	SelectStartingCards *SelectStartingCardsHandler
	StartGame          *StartGameHandler
}

// NewActionHandlers creates a new instance of action handlers
func NewActionHandlers() *ActionHandlers {
	return &ActionHandlers{
		SelectStartingCards: &SelectStartingCardsHandler{},
		StartGame:          &StartGameHandler{},
	}
}