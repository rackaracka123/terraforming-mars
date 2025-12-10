package player

import (
	"sync"
	"terraforming-mars-backend/internal/events"
	"time"
)

// Hand manages player card hand (cards currently held)
type Hand struct {
	mu          sync.RWMutex
	cards       []string               // Card IDs
	playerCards map[string]*PlayerCard // Cached PlayerCard instances (with state)
	eventBus    *events.EventBusImpl
	gameID      string
	playerID    string
}

func newHand(eventBus *events.EventBusImpl, gameID, playerID string) *Hand {
	return &Hand{
		cards:       []string{},
		playerCards: make(map[string]*PlayerCard),
		eventBus:    eventBus,
		gameID:      gameID,
		playerID:    playerID,
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
	h.cards = append(h.cards, cardID)
	cardsCopy := make([]string, len(h.cards))
	copy(cardsCopy, h.cards)
	h.mu.Unlock()

	// Publish domain events after adding card
	if h.eventBus != nil {
		// Publish CardAddedToHandEvent for passive card effects
		events.Publish(h.eventBus, events.CardAddedToHandEvent{
			GameID:    h.gameID,
			PlayerID:  h.playerID,
			CardID:    cardID,
			Timestamp: time.Now(),
		})

		// Publish CardHandUpdatedEvent with current hand state
		events.Publish(h.eventBus, events.CardHandUpdatedEvent{
			GameID:    h.gameID,
			PlayerID:  h.playerID,
			CardIDs:   cardsCopy,
			Timestamp: time.Now(),
		})

	}
}

func (h *Hand) RemoveCard(cardID string) bool {
	var removed bool
	h.mu.Lock()

	// Clean up event listeners before removing from cache
	if pc, exists := h.playerCards[cardID]; exists {
		pc.Cleanup() // Unsubscribe all event listeners
		delete(h.playerCards, cardID)
	}

	// Remove from card ID list
	for i, id := range h.cards {
		if id == cardID {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			removed = true
			break
		}
	}
	cardsCopy := make([]string, len(h.cards))
	copy(cardsCopy, h.cards)
	h.mu.Unlock()

	// Publish domain events after removing card (only if card was found)
	if removed && h.eventBus != nil {
		// Publish CardHandUpdatedEvent with current hand state
		events.Publish(h.eventBus, events.CardHandUpdatedEvent{
			GameID:    h.gameID,
			PlayerID:  h.playerID,
			CardIDs:   cardsCopy,
			Timestamp: time.Now(),
		})

	}

	return removed
}

// GetPlayerCard returns a cached PlayerCard instance.
// Returns nil if the card is not in the cache (actions must create it).
func (h *Hand) GetPlayerCard(cardID string) (*PlayerCard, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	pc, exists := h.playerCards[cardID]
	return pc, exists
}

// AddPlayerCard adds a PlayerCard to the cache.
// This is called by actions after creating PlayerCard with state and event listeners.
func (h *Hand) AddPlayerCard(cardID string, pc *PlayerCard) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.playerCards[cardID] = pc
}
