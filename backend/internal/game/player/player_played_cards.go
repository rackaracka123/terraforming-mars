package player

import "sync"

// PlayedCards manages all cards a player has played, including corporation
type PlayedCards struct {
	mu    sync.RWMutex
	cards []string // Includes ALL played cards (corporation + project cards)
}

func newPlayedCards() *PlayedCards {
	return &PlayedCards{
		cards: []string{},
	}
}

// Cards returns a copy of all played cards
func (pc *PlayedCards) Cards() []string {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	cardsCopy := make([]string, len(pc.cards))
	copy(cardsCopy, pc.cards)
	return cardsCopy
}

// Contains checks if a specific card has been played
func (pc *PlayedCards) Contains(cardID string) bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	for _, id := range pc.cards {
		if id == cardID {
			return true
		}
	}
	return false
}

// AddCard adds a card to played cards (used for both corporation and project cards)
func (pc *PlayedCards) AddCard(cardID string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.cards = append(pc.cards, cardID)
}

// RemoveCard removes a card from played cards (if it exists)
func (pc *PlayedCards) RemoveCard(cardID string) bool {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	for i, id := range pc.cards {
		if id == cardID {
			pc.cards = append(pc.cards[:i], pc.cards[i+1:]...)
			return true
		}
	}
	return false
}

// SetCards replaces all played cards (used for initialization/loading)
func (pc *PlayedCards) SetCards(cards []string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	if cards == nil {
		pc.cards = []string{}
	} else {
		pc.cards = make([]string, len(cards))
		copy(pc.cards, cards)
	}
}

// Count returns the number of played cards
func (pc *PlayedCards) Count() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return len(pc.cards)
}
