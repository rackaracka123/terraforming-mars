package actions

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
)

// SkipActionHandler handles skip actions
type SkipActionHandler struct{}

// Handle applies the skip action
func (h *SkipActionHandler) Handle(game *domain.Game, player *domain.Player, actionPayload dto.ActionPayload) error {
	action := dto.SkipActionAction{Type: actionPayload.Type}
	return h.applySkipAction(game, player, action)
}

// applySkipAction applies skip action
func (h *SkipActionHandler) applySkipAction(game *domain.Game, player *domain.Player, action dto.SkipActionAction) error {
	// TODO: Implement turn system and move to next player
	return nil
}