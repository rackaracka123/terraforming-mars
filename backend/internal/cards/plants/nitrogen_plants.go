package plants

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/model"
)

// NitrogenPlantsHandler implements the Nitrogen-Rich Plants card
type NitrogenPlantsHandler struct {
	cards.EffectCardHandler
}

// NewNitrogenPlantsHandler creates a new Nitrogen-Rich Plants card handler
func NewNitrogenPlantsHandler() *NitrogenPlantsHandler {
	return &NitrogenPlantsHandler{
		EffectCardHandler: cards.EffectCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "nitrogen-plants",
				Requirements: model.CardRequirements{},
			},
		},
	}
}

// Play executes the Nitrogen-Rich Plants card effect
func (h *NitrogenPlantsHandler) Play(ctx *cards.CardHandlerContext) error {
	// Gain 1 Plant production
	return ctx.PlayerService.AddProduction(ctx.Context, ctx.Game.ID, ctx.PlayerID, model.ResourceSet{
		Plants: 1,
	})
}
