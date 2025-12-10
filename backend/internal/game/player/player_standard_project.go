package player

import (
	"sync"

	"terraforming-mars-backend/internal/game/shared"
)

// PlayerStandardProject represents a standard project with availability state.
// This is a DATA HOLDER - availability calculation happens in action package.
type PlayerStandardProject struct {
	projectType shared.StandardProject

	// Calculated state (from action package)
	mu    sync.RWMutex
	state EntityState

	// Event listener cleanup functions
	unsubscribers []func()
}

// NewPlayerStandardProject creates a new PlayerStandardProject with empty state.
func NewPlayerStandardProject(projectType shared.StandardProject) *PlayerStandardProject {
	return &PlayerStandardProject{
		projectType: projectType,
		state: EntityState{
			Errors:   []StateError{},
			Metadata: make(map[string]interface{}),
		},
		unsubscribers: make([]func(), 0),
	}
}

// UpdateState updates the calculated state (called by action package after recalculation).
func (psp *PlayerStandardProject) UpdateState(newState EntityState) {
	psp.mu.Lock()
	defer psp.mu.Unlock()
	psp.state = newState
}

// AddUnsubscriber adds an event listener cleanup function.
func (psp *PlayerStandardProject) AddUnsubscriber(unsub func()) {
	psp.mu.Lock()
	defer psp.mu.Unlock()
	psp.unsubscribers = append(psp.unsubscribers, unsub)
}

// ProjectType returns the standard project type.
func (psp *PlayerStandardProject) ProjectType() shared.StandardProject {
	return psp.projectType
}

// State returns a copy of the current calculated state.
func (psp *PlayerStandardProject) State() EntityState {
	psp.mu.RLock()
	defer psp.mu.RUnlock()
	return psp.state
}

// IsAvailable returns true if the project is available (no errors).
func (psp *PlayerStandardProject) IsAvailable() bool {
	psp.mu.RLock()
	defer psp.mu.RUnlock()
	return psp.state.Available()
}

// Cleanup unsubscribes all event listeners.
// Called when needed to prevent memory leaks.
func (psp *PlayerStandardProject) Cleanup() {
	psp.mu.Lock()
	defer psp.mu.Unlock()

	for _, unsub := range psp.unsubscribers {
		unsub()
	}
	psp.unsubscribers = nil
}
