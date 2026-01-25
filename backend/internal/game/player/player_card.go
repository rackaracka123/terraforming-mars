package player

import (
	"sync"
)

// PlayerCard is a player-specific view of a Card with calculated state.
// This is a DATA HOLDER - state calculation happens in action package.
// PlayerCard instances are cached for their entire lifecycle (in hand, selection phase, etc.).
// Note: card field uses any to avoid circular dependency with game/cards package
type PlayerCard struct {
	card any

	mu    sync.RWMutex
	state EntityState

	unsubscribers []func()
}

// NewPlayerCard creates a new PlayerCard with empty state.
// The action package is responsible for:
// - Registering event listeners (tracking unsubscribers)
// - Calculating initial state
// - Adding to Hand cache
// card parameter should be *cards.Card from game/cards package
func NewPlayerCard(card any) *PlayerCard {
	return &PlayerCard{
		card: card,
		state: EntityState{
			Errors:   []StateError{},
			Metadata: make(map[string]interface{}),
		},
		unsubscribers: make([]func(), 0),
	}
}

// UpdateState updates the calculated state (called by action package after recalculation).
func (pc *PlayerCard) UpdateState(newState EntityState) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.state = newState
}

// AddUnsubscriber adds an event listener cleanup function.
// Called by action package when registering event listeners.
func (pc *PlayerCard) AddUnsubscriber(unsub func()) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.unsubscribers = append(pc.unsubscribers, unsub)
}

// Card returns the immutable card data reference (type will be *cards.Card).
// Returns any to avoid circular dependency - caller should type assert if needed.
func (pc *PlayerCard) Card() any {
	return pc.card
}

// State returns a copy of the current calculated state.
func (pc *PlayerCard) State() EntityState {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.state
}

// IsAvailable returns true if the card is available (no errors).
func (pc *PlayerCard) IsAvailable() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.state.Available()
}

// Cleanup unsubscribes all event listeners.
// MUST be called when PlayerCard is removed from hand/selection to prevent memory leaks.
func (pc *PlayerCard) Cleanup() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, unsub := range pc.unsubscribers {
		unsub()
	}
	pc.unsubscribers = nil
}
