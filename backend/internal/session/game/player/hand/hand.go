package hand

import "sync"

// Hand manages player card hand and played cards.
// Thread-safe with its own mutex.
type Hand struct {
	mu          sync.RWMutex
	cards       []string // Card IDs currently in player's hand
	playedCards []string // Card IDs that have been played
}

// NewHand creates a new Hand component with empty card collections.
func NewHand() *Hand {
	return &Hand{
		cards:       []string{},
		playedCards: []string{},
	}
}

// ==================== Getters ====================

// Cards returns a defensive copy of card IDs in the player's hand.
func (h *Hand) Cards() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.cards == nil {
		return []string{}
	}
	cardsCopy := make([]string, len(h.cards))
	copy(cardsCopy, h.cards)
	return cardsCopy
}

// PlayedCards returns a defensive copy of card IDs that have been played.
func (h *Hand) PlayedCards() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.playedCards == nil {
		return []string{}
	}
	playedCopy := make([]string, len(h.playedCards))
	copy(playedCopy, h.playedCards)
	return playedCopy
}

// CardCount returns the number of cards in the player's hand.
func (h *Hand) CardCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.cards)
}

// HasCard returns true if the specified card is in the player's hand.
func (h *Hand) HasCard(cardID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, id := range h.cards {
		if id == cardID {
			return true
		}
	}
	return false
}

// ==================== Setters ====================

// SetCards replaces the entire hand with the provided card IDs.
func (h *Hand) SetCards(cards []string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cards == nil {
		h.cards = []string{}
		return
	}
	h.cards = make([]string, len(cards))
	copy(h.cards, cards)
}

// SetPlayedCards replaces the entire played cards collection.
func (h *Hand) SetPlayedCards(playedCards []string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if playedCards == nil {
		h.playedCards = []string{}
		return
	}
	h.playedCards = make([]string, len(playedCards))
	copy(h.playedCards, playedCards)
}

// ==================== Mutations ====================

// AddCard adds a card to the player's hand.
func (h *Hand) AddCard(cardID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cards = append(h.cards, cardID)
}

// RemoveCard removes a card from the player's hand.
// Returns true if the card was found and removed, false otherwise.
func (h *Hand) RemoveCard(cardID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, id := range h.cards {
		if id == cardID {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			return true
		}
	}
	return false
}

// PlayCard moves a card from the hand to played cards.
// Returns true if the card was found and played, false otherwise.
func (h *Hand) PlayCard(cardID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Remove from hand
	for i, id := range h.cards {
		if id == cardID {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			h.playedCards = append(h.playedCards, cardID)
			return true
		}
	}
	return false
}

// AddPlayedCard adds a card directly to the played cards collection.
// This is used when a card is played without being in hand first (e.g., corporation).
func (h *Hand) AddPlayedCard(cardID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.playedCards = append(h.playedCards, cardID)
}

// ==================== Utilities ====================

// DeepCopy creates a deep copy of the Hand component.
func (h *Hand) DeepCopy() *Hand {
	if h == nil {
		return nil
	}

	cardsCopy := make([]string, len(h.cards))
	copy(cardsCopy, h.cards)

	playedCopy := make([]string, len(h.playedCards))
	copy(playedCopy, h.playedCards)

	return &Hand{
		cards:       cardsCopy,
		playedCards: playedCopy,
	}
}
