package player

import (
	"sync"
	"terraforming-mars-backend/internal/events"
	"time"
)

// PlayedCards manages all cards a player has played, including corporation
type PlayedCards struct {
	mu       sync.RWMutex
	cards    []string // Includes ALL played cards (corporation + project cards)
	eventBus *events.EventBusImpl
	gameID   string
	playerID string
}

func newPlayedCards(eventBus *events.EventBusImpl, gameID, playerID string) *PlayedCards {
	return &PlayedCards{
		cards:    []string{},
		eventBus: eventBus,
		gameID:   gameID,
		playerID: playerID,
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
// Parameters:
//   - cardID: The unique identifier of the card
//   - cardName: The display name of the card
//   - cardType: The type of card (event, automated, active, corporation, prelude)
func (pc *PlayedCards) AddCard(cardID, cardName, cardType string) {
	pc.mu.Lock()
	pc.cards = append(pc.cards, cardID)
	pc.mu.Unlock()

	// Publish domain events after adding card
	if pc.eventBus != nil {
		// Publish CardPlayedEvent for passive card effects and game logging
		events.Publish(pc.eventBus, events.CardPlayedEvent{
			GameID:    pc.gameID,
			PlayerID:  pc.playerID,
			CardID:    cardID,
			CardName:  cardName,
			CardType:  cardType,
			Timestamp: time.Now(),
		})

		// Publish broadcast event to trigger client updates
		events.Publish(pc.eventBus, events.BroadcastEvent{
			GameID:    pc.gameID,
			PlayerIDs: []string{pc.playerID},
		})
	}
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
