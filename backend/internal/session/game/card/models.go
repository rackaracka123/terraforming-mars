package card

import "terraforming-mars-backend/internal/model"

// Card represents a game card
// For Phase 2 (start_game), we just need basic card data for drawing
type Card struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // "project", "corporation", "prelude"
	Pack string `json:"pack"` // "base-game", "future", etc.
}

// FromModelCard converts a model.Card to card subdomain Card
func FromModelCard(mc model.Card) Card {
	return Card{
		ID:   mc.ID,
		Name: mc.Name,
		Type: string(mc.Type),
		Pack: mc.Pack,
	}
}

// FromModelCards converts multiple model.Card to card subdomain Cards
func FromModelCards(mcs []model.Card) []Card {
	cards := make([]Card, len(mcs))
	for i, mc := range mcs {
		cards[i] = FromModelCard(mc)
	}
	return cards
}
