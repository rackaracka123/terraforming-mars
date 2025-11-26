package turn

import "sync"

// Turn manages player turn state including passed status, available actions, and connection status.
// Thread-safe with its own mutex.
type Turn struct {
	mu               sync.RWMutex
	passed           bool
	availableActions int
	isConnected      bool
}

// NewTurn creates a new Turn component with default values.
// Players start with no available actions, not passed, and disconnected.
func NewTurn() *Turn {
	return &Turn{
		passed:           false,
		availableActions: 0,
		isConnected:      false,
	}
}

// ==================== Getters ====================

// Passed returns whether the player has passed for the current generation.
func (t *Turn) Passed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.passed
}

// AvailableActions returns the number of actions the player can still take this turn.
func (t *Turn) AvailableActions() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.availableActions
}

// IsConnected returns whether the player is currently connected to the game.
func (t *Turn) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isConnected
}

// HasActions returns true if the player has any available actions remaining.
func (t *Turn) HasActions() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.availableActions > 0
}

// ==================== Setters ====================

// SetPassed sets whether the player has passed for the current generation.
func (t *Turn) SetPassed(passed bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.passed = passed
}

// SetAvailableActions sets the number of actions the player can take.
func (t *Turn) SetAvailableActions(actions int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.availableActions = actions
}

// SetConnectionStatus sets whether the player is currently connected.
func (t *Turn) SetConnectionStatus(isConnected bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isConnected = isConnected
}

// ==================== Mutations ====================

// ConsumeAction decrements the available actions by 1.
// Returns true if an action was consumed, false if no actions were available.
func (t *Turn) ConsumeAction() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.availableActions > 0 {
		t.availableActions--
		return true
	}
	return false
}

// ResetForGeneration resets the turn state for a new generation.
// Clears the passed flag and sets available actions to 0.
func (t *Turn) ResetForGeneration() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.passed = false
	t.availableActions = 0
}

// ==================== Utilities ====================

// DeepCopy creates a deep copy of the Turn component.
func (t *Turn) DeepCopy() *Turn {
	if t == nil {
		return nil
	}
	return &Turn{
		passed:           t.passed,
		availableActions: t.availableActions,
		isConnected:      t.isConnected,
	}
}
