package actions

import (
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/domain"
)

// SelectCorporationHandler handles corporation selection actions
type SelectCorporationHandler struct{}

// Handle applies the select corporation action
func (h *SelectCorporationHandler) Handle(game *domain.Game, player *domain.Player, actionPayload dto.ActionPayload) error {
	if actionPayload.CorporationName == nil {
		return fmt.Errorf("corporation name is required for select corporation action")
	}
	action := dto.SelectCorporationAction{
		Type:            actionPayload.Type,
		CorporationName: *actionPayload.CorporationName,
	}
	return h.applySelectCorporation(game, player, action)
}

// applySelectCorporation applies corporation selection
func (h *SelectCorporationHandler) applySelectCorporation(game *domain.Game, player *domain.Player, action dto.SelectCorporationAction) error {
	if player.Corporation != "" {
		return fmt.Errorf("player already has a corporation")
	}

	if action.CorporationName == "" {
		return fmt.Errorf("corporation name cannot be empty")
	}

	// TODO: Validate corporation exists and apply starting resources/production
	player.Corporation = action.CorporationName

	return nil
}