package player

import (
	"sync"

	"terraforming-mars-backend/internal/game/shared"
)

// PlayerAward represents an award with player-specific eligibility state.
// This is a DATA HOLDER - eligibility calculation happens in action package.
// The award's global funding state lives in Game.Awards().
type PlayerAward struct {
	awardType shared.AwardType

	// Calculated state (from action package)
	mu    sync.RWMutex
	state EntityState

	// Event listener cleanup functions
	unsubscribers []func()
}

// NewPlayerAward creates a new PlayerAward with empty state.
func NewPlayerAward(awardType shared.AwardType) *PlayerAward {
	return &PlayerAward{
		awardType: awardType,
		state: EntityState{
			Errors:   []StateError{},
			Cost:     make(map[string]int),
			Metadata: make(map[string]interface{}),
		},
		unsubscribers: make([]func(), 0),
	}
}

// UpdateState updates the calculated state (called by action package after recalculation).
func (pa *PlayerAward) UpdateState(newState EntityState) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.state = newState
}

// AddUnsubscriber adds an event listener cleanup function.
func (pa *PlayerAward) AddUnsubscriber(unsub func()) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.unsubscribers = append(pa.unsubscribers, unsub)
}

// AwardType returns the award type.
func (pa *PlayerAward) AwardType() shared.AwardType {
	return pa.awardType
}

// State returns a copy of the current calculated state.
func (pa *PlayerAward) State() EntityState {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.state
}

// IsAvailable returns true if the award can be funded (no errors).
func (pa *PlayerAward) IsAvailable() bool {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.state.Available()
}

// Cleanup unsubscribes all event listeners.
// Called when needed to prevent memory leaks.
func (pa *PlayerAward) Cleanup() {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	for _, unsub := range pa.unsubscribers {
		unsub()
	}
	pa.unsubscribers = nil
}
