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
	// Pay the activation cost and gain energy production
	if err := ctx.PlayerService.PayResourceCost(ctx.Context, ctx.Game.ID, ctx.PlayerID, *h.ActivationCost); err != nil {
		return err
	}

	// Gain 1 Energy production
	return ctx.PlayerService.AddProduction(ctx.Context, ctx.Game.ID, ctx.PlayerID, model.ResourceSet{
		Energy: 1,
	})
}
