package economy

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/model"
)

// EarlySettlementHandler implements the Early Settlement card
type EarlySettlementHandler struct {
	cards.EffectCardHandler
}

// NewEarlySettlementHandler creates a new Early Settlement card handler
func NewEarlySettlementHandler() *EarlySettlementHandler {
	return &EarlySettlementHandler{
		EffectCardHandler: cards.EffectCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "early-settlement",
				Requirements: model.CardRequirements{},
			},
		},
	}
}

// Play executes the Early Settlement card effect
func (h *EarlySettlementHandler) Play(ctx *cards.CardHandlerContext) error {
	// Gain 1 MC production
	cards.AddProduction(ctx.Player, model.ResourceSet{
		Credits: 1,
	})
	
	return nil
}