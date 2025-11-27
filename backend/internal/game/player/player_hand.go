package player

import "sync"

// Hand manages player card hand (cards currently held)
type Hand struct {
	mu    sync.RWMutex
	cards []string
}

func newHand() *Hand {
	return &Hand{
		cards: []string{},
	}
}

func (h *Hand) Cards() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	cardsCopy := make([]string, len(h.cards))
	copy(cardsCopy, h.cards)
	return cardsCopy
}

func (h *Hand) CardCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.cards)
}

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

func (h *Hand) SetCards(cards []string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cards == nil {
		h.cards = []string{}
	} else {
		h.cards = make([]string, len(cards))
		copy(h.cards, cards)
	}
}

func (h *Hand) AddCard(cardID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cards = append(h.cards, cardID)
}

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
