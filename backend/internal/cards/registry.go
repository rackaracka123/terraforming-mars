package cards

import (
	"fmt"

	"terraforming-mars-backend/internal/game/cardtypes"
)

// CardRegistry provides lookup functionality for card data
type CardRegistry interface {
	// GetByID retrieves a card by its ID
	GetByID(cardID string) (*cardtypes.Card, error)

	// GetAll returns all cards in the registry
	GetAll() []cardtypes.Card
}

// InMemoryCardRegistry implements CardRegistry with an in-memory map
type InMemoryCardRegistry struct {
	cards map[string]cardtypes.Card
}

// NewInMemoryCardRegistry creates a new card registry from a slice of cards
func NewInMemoryCardRegistry(cards []cardtypes.Card) *InMemoryCardRegistry {
	cardMap := make(map[string]cardtypes.Card, len(cards))
	for _, card := range cards {
		cardMap[card.ID] = card
	}

	return &InMemoryCardRegistry{
		cards: cardMap,
	}
}

// GetByID retrieves a card by its ID, returning a copy to prevent mutation
func (r *InMemoryCardRegistry) GetByID(cardID string) (*cardtypes.Card, error) {
	card, exists := r.cards[cardID]
	if !exists {
		return nil, fmt.Errorf("card not found: %s", cardID)
	}

	// Return a deep copy to prevent external mutation
	cardCopy := card.DeepCopy()
	return &cardCopy, nil
}

// GetAll returns all cards in the registry
func (r *InMemoryCardRegistry) GetAll() []cardtypes.Card {
	cards := make([]cardtypes.Card, 0, len(r.cards))
	for _, card := range r.cards {
		cards = append(cards, card.DeepCopy())
	}
	return cards
}
