package power

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/model"
)

// PowerPlantHandler implements the Power Plant card
type PowerPlantHandler struct {
	cards.EffectCardHandler
}

// NewPowerPlantHandler creates a new Power Plant card handler
func NewPowerPlantHandler() *PowerPlantHandler {
	return &PowerPlantHandler{
		EffectCardHandler: cards.EffectCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "power-plant",
				Requirements: model.CardRequirements{},
			},
		},
	}
}

// Play executes the Power Plant card effect
func (h *PowerPlantHandler) Play(ctx *cards.CardHandlerContext) error {
	// Gain 1 Energy production
	return ctx.PlayerService.AddProduction(ctx.Context, ctx.Game.ID, ctx.PlayerID, model.ResourceSet{
		Energy: 1,
	})
}