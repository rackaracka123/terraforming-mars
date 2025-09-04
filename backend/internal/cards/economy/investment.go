package economy

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/model"
)

// InvestmentHandler implements the Investment card
type InvestmentHandler struct {
	cards.EventCardHandler
}

// NewInvestmentHandler creates a new Investment card handler
func NewInvestmentHandler() *InvestmentHandler {
	return &InvestmentHandler{
		EventCardHandler: cards.EventCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "investment",
				Requirements: model.CardRequirements{},
			},
		},
	}
}

// Play executes the Investment card effect
func (h *InvestmentHandler) Play(ctx *cards.CardHandlerContext) error {
	// This card gives 1 VP (already handled by VictoryPoints field in card data)
	// No immediate game state changes needed
	return nil
}