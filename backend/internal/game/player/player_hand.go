package player

import (
	"sync"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/playability"
	"time"
)

// Hand manages player card hand (cards currently held)
type Hand struct {
	mu          sync.RWMutex
	cards       []string
	playability map[string]playability.PlayabilityResult
	eventBus    *events.EventBusImpl
	gameID      string
	playerID    string
}

func newHand(eventBus *events.EventBusImpl, gameID, playerID string) *Hand {
	return &Hand{
		cards:       []string{},
		playability: make(map[string]playability.PlayabilityResult),
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

// ==================== Playability Management ====================

// GetPlayability returns the playability result for a specific card
func (h *Hand) GetPlayability(cardID string) playability.PlayabilityResult {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if result, exists := h.playability[cardID]; exists {
		return result
	}
	// Return unplayable result if no cached playability
	return playability.NewPlayabilityResult(false, []playability.ValidationError{
		{
			Type:    playability.ValidationErrorTypeGameState,
			Message: "Playability not calculated",
		},
	})
}

// GetAllPlayability returns playability for all cards in hand
func (h *Hand) GetAllPlayability() map[string]playability.PlayabilityResult {
	h.mu.RLock()
	defer h.mu.RUnlock()
	playabilityCopy := make(map[string]playability.PlayabilityResult, len(h.playability))
	for cardID, result := range h.playability {
		playabilityCopy[cardID] = result
	}
	return playabilityCopy
}

// SetPlayability updates playability for a specific card
// This is called by the DTO layer after calculating playability
func (h *Hand) SetPlayability(cardID string, result playability.PlayabilityResult) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.playability[cardID] = result
}

// ClearStalePlayability removes playability for cards no longer in hand
// This can be called periodically to clean up stale cache entries
func (h *Hand) ClearStalePlayability() {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Create new map with only cards still in hand
	newPlayability := make(map[string]playability.PlayabilityResult)
	for _, cardID := range h.cards {
		if result, exists := h.playability[cardID]; exists {
			newPlayability[cardID] = result
		}
	}
	h.playability = newPlayability
}
