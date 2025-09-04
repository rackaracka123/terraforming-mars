package power

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/model"
)

// SpaceMirrorsHandler implements the Space Mirrors card
type SpaceMirrorsHandler struct {
	cards.ActiveCardHandler
}

// NewSpaceMirrorsHandler creates a new Space Mirrors card handler
func NewSpaceMirrorsHandler() *SpaceMirrorsHandler {
	return &SpaceMirrorsHandler{
		ActiveCardHandler: cards.ActiveCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "space-mirrors",
				Requirements: model.CardRequirements{},
			},
			ActivationCost: &model.ResourceSet{
				Credits: 7,
			},
		},
	}
}

// Play executes the Space Mirrors card effect (when first played)
func (h *SpaceMirrorsHandler) Play(ctx *cards.CardHandlerContext) error {
	// Active card - no immediate effect when played
	// The effect happens when activated
	return nil
}

// Activate executes the Space Mirrors repeatable action
func (h *SpaceMirrorsHandler) Activate(ctx *cards.CardHandlerContext) error {
	// Check if player can afford the activation cost
	if err := cards.ValidateResourceCost(ctx.Player, *h.ActivationCost); err != nil {
		return err
	}
	
	// Pay the activation cost
	cards.PayResourceCost(ctx.Player, *h.ActivationCost)
	
	// Gain 1 Energy production
	cards.AddProduction(ctx.Player, model.ResourceSet{
		Energy: 1,
	})
	
	return nil
}