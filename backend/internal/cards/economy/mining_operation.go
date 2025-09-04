package economy

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/model"
)

// MiningOperationHandler implements the Mining Operation card
type MiningOperationHandler struct {
	cards.EventCardHandler
}

// NewMiningOperationHandler creates a new Mining Operation card handler
func NewMiningOperationHandler() *MiningOperationHandler {
	return &MiningOperationHandler{
		EventCardHandler: cards.EventCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "mining-operation",
				Requirements: model.CardRequirements{},
			},
		},
	}
}

// Play executes the Mining Operation card effect
func (h *MiningOperationHandler) Play(ctx *cards.CardHandlerContext) error {
	// Gain 2 Steel
	cards.AddResources(ctx.Player, model.ResourceSet{
		Steel: 2,
	})
	
	return nil
}