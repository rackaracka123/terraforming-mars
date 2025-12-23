package player

import (
	"sync"

	"terraforming-mars-backend/internal/game/shared"
)

// PlayerMilestone represents a milestone with player-specific eligibility state.
// This is a DATA HOLDER - eligibility calculation happens in action package.
// The milestone's global claim state lives in Game.Milestones().
type PlayerMilestone struct {
	milestoneType shared.MilestoneType

	// Calculated state (from action package)
	mu    sync.RWMutex
	state EntityState

	// Event listener cleanup functions
	unsubscribers []func()
}

// NewPlayerMilestone creates a new PlayerMilestone with empty state.
func NewPlayerMilestone(milestoneType shared.MilestoneType) *PlayerMilestone {
	return &PlayerMilestone{
		milestoneType: milestoneType,
		state: EntityState{
			Errors:   []StateError{},
			Cost:     make(map[string]int),
			Metadata: make(map[string]interface{}),
		},
		unsubscribers: make([]func(), 0),
	}
}

// UpdateState updates the calculated state (called by action package after recalculation).
func (pm *PlayerMilestone) UpdateState(newState EntityState) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.state = newState
}

// AddUnsubscriber adds an event listener cleanup function.
func (pm *PlayerMilestone) AddUnsubscriber(unsub func()) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.unsubscribers = append(pm.unsubscribers, unsub)
}

// MilestoneType returns the milestone type.
func (pm *PlayerMilestone) MilestoneType() shared.MilestoneType {
	return pm.milestoneType
}

// State returns a copy of the current calculated state.
func (pm *PlayerMilestone) State() EntityState {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.state
}

// IsAvailable returns true if the milestone can be claimed (no errors).
func (pm *PlayerMilestone) IsAvailable() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.state.Available()
}

// Cleanup unsubscribes all event listeners.
// Called when needed to prevent memory leaks.
func (pm *PlayerMilestone) Cleanup() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, unsub := range pm.unsubscribers {
		unsub()
	}
	pm.unsubscribers = nil
}
