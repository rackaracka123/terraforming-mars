package cards

import "terraforming-mars-backend/internal/model"

// MockCardDatabase provides a simple, centralized source of card data
// This is temporary for testing and should be replaced with proper card loading later
var MockCardDatabase = map[string]model.Card{
	"water-import": {
		ID:          "water-import",
		Name:        "Water Import From Europa",
		Type:        model.CardTypeEffect,
		Cost:        12,
		Description: "Increase your plant production 2 steps and place an ocean tile.",
	},
	"atmospheric-processors": {
		ID:          "atmospheric-processors",
		Name:        "Atmospheric Processors",
		Type:        model.CardTypeEffect,
		Cost:        18,
		Description: "Increase your heat production 3 steps.",
	},
	"heat-generators": {
		ID:          "heat-generators",
		Name:        "Heat Generators",
		Type:        model.CardTypeEffect,
		Cost:        6,
		Description: "Increase your heat production 2 steps.",
	},
	"power-plant": {
		ID:          "power-plant",
		Name:        "Power Plant",
		Type:        model.CardTypeEffect,
		Cost:        4,
		Description: "Increase your energy production 1 step.",
	},
	"mining-operation": {
		ID:          "mining-operation",
		Name:        "Mining Operation",
		Type:        model.CardTypeEffect,
		Cost:        5,
		Description: "Increase your steel production 2 steps.",
	},
	"nitrogen-plants": {
		ID:          "nitrogen-plants",
		Name:        "Nitrogen-Rich Asteroid",
		Type:        model.CardTypeEvent,
		Cost:        8,
		Description: "Increase your plant production 2 steps and TR 2 steps.",
	},
	"investment": {
		ID:          "investment",
		Name:        "Investment Loan",
		Type:        model.CardTypeEvent,
		Cost:        3,
		Description: "Increase your MC production 1 step but decrease your MC 1 step.",
	},
	"early-settlement": {
		ID:          "early-settlement",
		Name:        "Early Settlement",
		Type:        model.CardTypeEffect,
		Cost:        5,
		Description: "Increase your plant production 1 step.",
	},
	"space-mirrors": {
		ID:          "space-mirrors",
		Name:        "Space Mirrors",
		Type:        model.CardTypeEffect,
		Cost:        3,
		Description: "Increase your energy production 1 step.",
	},
}

// GetCardByID returns card data from the mock database
func GetCardByID(cardID string) (model.Card, bool) {
	card, exists := MockCardDatabase[cardID]
	return card, exists
}

// GetCardsByIDs returns multiple cards from the mock database
func GetCardsByIDs(cardIDs []string) []model.Card {
	cards := make([]model.Card, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		if card, exists := MockCardDatabase[cardID]; exists {
			cards = append(cards, card)
		}
	}
	return cards
}
