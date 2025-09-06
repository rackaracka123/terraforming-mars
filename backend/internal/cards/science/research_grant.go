package science

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/model"
)

// ResearchGrantHandler implements the Research Grant card
type ResearchGrantHandler struct {
	cards.EventCardHandler
}

// NewResearchGrantHandler creates a new Research Grant card handler
func NewResearchGrantHandler() *ResearchGrantHandler {
	return &ResearchGrantHandler{
		EventCardHandler: cards.EventCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "research-grant",
				Requirements: model.CardRequirements{},
			},
		},
	}
}

// Play executes the Research Grant card effect
func (h *ResearchGrantHandler) Play(ctx *cards.CardHandlerContext) error {
	// TODO: Implement card drawing mechanism
	// For now, this is a placeholder - the actual implementation would
	// involve drawing additional cards for the player
	return nil
}
