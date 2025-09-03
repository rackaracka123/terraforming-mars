package actions

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
)

// ActionHandler defines the interface for handling game actions
type ActionHandler interface {
	Handle(game *domain.Game, player *domain.Player, actionPayload dto.ActionPayload) error
}

// ActionHandlers contains all action handlers
type ActionHandlers struct {
	StandardProjectAsteroid *StandardProjectAsteroidHandler
	RaiseTemperature       *RaiseTemperatureHandler
	SelectCorporation      *SelectCorporationHandler
	SkipAction            *SkipActionHandler
}

// NewActionHandlers creates a new instance of action handlers
func NewActionHandlers() *ActionHandlers {
	return &ActionHandlers{
		StandardProjectAsteroid: &StandardProjectAsteroidHandler{},
		RaiseTemperature:       &RaiseTemperatureHandler{},
		SelectCorporation:      &SelectCorporationHandler{},
		SkipAction:            &SkipActionHandler{},
	}
}