package actions

import (
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
)

// StandardProjectAsteroidHandler handles standard project asteroid actions
type StandardProjectAsteroidHandler struct{}

// Handle applies the standard project asteroid action
func (h *StandardProjectAsteroidHandler) Handle(game *domain.Game, player *domain.Player, actionPayload dto.ActionPayload) error {
	action := dto.StandardProjectAsteroidAction{Type: actionPayload.Type}
	return h.applyStandardProjectAsteroid(game, player, action)
}

// applyStandardProjectAsteroid applies the standard project asteroid action
func (h *StandardProjectAsteroidHandler) applyStandardProjectAsteroid(game *domain.Game, player *domain.Player, action dto.StandardProjectAsteroidAction) error {
	// Cost: 14 MC, Effect: Raise temperature 1 step
	if player.Resources.Credits < 14 {
		return fmt.Errorf("insufficient credits (need 14, have %d)", player.Resources.Credits)
	}

	if game.GlobalParameters.Temperature >= 8 {
		return fmt.Errorf("temperature already at maximum")
	}

	// Deduct cost
	player.Resources.Credits -= 14

	// Apply effect
	game.GlobalParameters.Temperature += 2 // Each step is 2 degrees

	// Player gains terraform rating
	player.TerraformRating += 1

	return nil
}