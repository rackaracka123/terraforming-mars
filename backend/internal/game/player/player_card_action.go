package player

import (
	"sync"

	"terraforming-mars-backend/internal/game/shared"
)

// PlayerCardAction wraps a card action with usability state.
// This is a DATA HOLDER - usability calculation happens in action package.
type PlayerCardAction struct {
	cardID        string
	behaviorIndex int
	behavior      shared.CardBehavior

	// Persistent state (not calculated, tracked over time)
	timesUsedThisTurn       int
	timesUsedThisGeneration int

	// Calculated state (from action package)
	mu    sync.RWMutex
	state EntityState

	// Event listener cleanup functions
	unsubscribers []func()
}

// NewPlayerCardAction creates a new PlayerCardAction with empty state.
func NewPlayerCardAction(
	cardID string,
	behaviorIndex int,
	behavior shared.CardBehavior,
) *PlayerCardAction {
	return &PlayerCardAction{
		cardID:                  cardID,
		behaviorIndex:           behaviorIndex,
		behavior:                behavior,
		timesUsedThisTurn:       0,
		timesUsedThisGeneration: 0,
		state: EntityState{
			Errors:   []StateError{},
			Metadata: make(map[string]interface{}),
		},
		unsubscribers: make([]func(), 0),
	}
}

// UpdateState updates the calculated state (called by action package after recalculation).
func (pca *PlayerCardAction) UpdateState(newState EntityState) {
	pca.mu.Lock()
	defer pca.mu.Unlock()
	pca.state = newState
}

// AddUnsubscriber adds an event listener cleanup function.
func (pca *PlayerCardAction) AddUnsubscriber(unsub func()) {
	pca.mu.Lock()
	defer pca.mu.Unlock()
	pca.unsubscribers = append(pca.unsubscribers, unsub)
}

// CardID returns the card ID this action belongs to.
func (pca *PlayerCardAction) CardID() string {
	return pca.cardID
}

// BehaviorIndex returns the index of this behavior on the card.
func (pca *PlayerCardAction) BehaviorIndex() int {
	return pca.behaviorIndex
}

// Behavior returns the card behavior definition.
func (pca *PlayerCardAction) Behavior() shared.CardBehavior {
	return pca.behavior
}

// State returns a copy of the current calculated state.
func (pca *PlayerCardAction) State() EntityState {
	pca.mu.RLock()
	defer pca.mu.RUnlock()
	return pca.state
}

// IsAvailable returns true if the action is available (no errors).
func (pca *PlayerCardAction) IsAvailable() bool {
	pca.mu.RLock()
	defer pca.mu.RUnlock()
	return pca.state.Available()
}

// IncrementPlayCount increments both turn and generation play counts.
// Called when the action is used.
func (pca *PlayerCardAction) IncrementPlayCount() {
	pca.mu.Lock()
	defer pca.mu.Unlock()
	pca.timesUsedThisTurn++
	pca.timesUsedThisGeneration++
}

// ResetTurnPlayCount resets the turn play count (called at turn end).
func (pca *PlayerCardAction) ResetTurnPlayCount() {
	pca.mu.Lock()
	defer pca.mu.Unlock()
	pca.timesUsedThisTurn = 0
}

// ResetGenerationPlayCount resets the generation play count (called at generation end).
func (pca *PlayerCardAction) ResetGenerationPlayCount() {
	pca.mu.Lock()
	defer pca.mu.Unlock()
	pca.timesUsedThisGeneration = 0
}

// TimesUsedThisTurn returns the number of times this action has been used this turn.
func (pca *PlayerCardAction) TimesUsedThisTurn() int {
	pca.mu.RLock()
	defer pca.mu.RUnlock()
	return pca.timesUsedThisTurn
}

// TimesUsedThisGeneration returns the number of times this action has been used this generation.
func (pca *PlayerCardAction) TimesUsedThisGeneration() int {
	pca.mu.RLock()
	defer pca.mu.RUnlock()
	return pca.timesUsedThisGeneration
}

// Cleanup unsubscribes all event listeners.
// Called when action is removed to prevent memory leaks.
func (pca *PlayerCardAction) Cleanup() {
	pca.mu.Lock()
	defer pca.mu.Unlock()

	for _, unsub := range pca.unsubscribers {
		unsub()
	}
	pca.unsubscribers = nil
}
