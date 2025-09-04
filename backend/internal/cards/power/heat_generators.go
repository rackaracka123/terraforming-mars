package power

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/model"
)

// HeatGeneratorsHandler implements the Heat Generators card
type HeatGeneratorsHandler struct {
	cards.EffectCardHandler
}

// NewHeatGeneratorsHandler creates a new Heat Generators card handler
func NewHeatGeneratorsHandler() *HeatGeneratorsHandler {
	return &HeatGeneratorsHandler{
		EffectCardHandler: cards.EffectCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "heat-generators",
				Requirements: model.CardRequirements{},
			},
		},
	}
}

// Play executes the Heat Generators card effect
func (h *HeatGeneratorsHandler) Play(ctx *cards.CardHandlerContext) error {
	// Gain 1 Heat production
	cards.AddProduction(ctx.Player, model.ResourceSet{
		Heat: 1,
	})
	
	return nil
}