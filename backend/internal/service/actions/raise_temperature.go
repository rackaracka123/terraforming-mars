package actions

import (
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/domain"
)

// RaiseTemperatureHandler handles raise temperature actions
type RaiseTemperatureHandler struct{}

// Handle applies the raise temperature action
func (h *RaiseTemperatureHandler) Handle(game *domain.Game, player *domain.Player, actionPayload dto.ActionPayload) error {
	if actionPayload.HeatAmount == nil {
		return fmt.Errorf("heat amount is required for raise temperature action")
	}
	action := dto.RaiseTemperatureAction{
		Type:       actionPayload.Type,
		HeatAmount: *actionPayload.HeatAmount,
	}
	return h.applyRaiseTemperature(game, player, action)
}

// applyRaiseTemperature applies heat to raise temperature
func (h *RaiseTemperatureHandler) applyRaiseTemperature(game *domain.Game, player *domain.Player, action dto.RaiseTemperatureAction) error {
	if action.HeatAmount < 8 {
		return fmt.Errorf("need at least 8 heat to raise temperature")
	}

	if player.Resources.Heat < action.HeatAmount {
		return fmt.Errorf("insufficient heat (need %d, have %d)", action.HeatAmount, player.Resources.Heat)
	}

	if game.GlobalParameters.Temperature >= 8 {
		return fmt.Errorf("temperature already at maximum")
	}

	// Spend 8 heat to raise temperature 1 step
	steps := action.HeatAmount / 8
	player.Resources.Heat -= steps * 8
	game.GlobalParameters.Temperature += steps * 2

	// Cap at maximum
	if game.GlobalParameters.Temperature > 8 {
		game.GlobalParameters.Temperature = 8
	}

	// Player gains terraform rating for each step
	player.TerraformRating += steps

	return nil
}