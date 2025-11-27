package game

import "sync"

// ActionType represents the type of action a player can take
type ActionType string

// Turn tracks the active player and available actions for the current turn
// This is game-level turn state (whose turn it is), not per-player state
type Turn struct {
	mu               sync.RWMutex
	playerID         string
	availableActions []ActionType
	actionsRemaining int
}

// NewTurn creates a new turn for the specified player
func NewTurn(playerID string, availableActions []ActionType) *Turn {
	return &Turn{
		playerID:         playerID,
		availableActions: availableActions,
		actionsRemaining: len(availableActions),
	}
}

// PlayerID returns the ID of the player whose turn it is
func (t *Turn) PlayerID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.playerID
}

// AvailableActions returns the list of available action types
func (t *Turn) AvailableActions() []ActionType {
	t.mu.RLock()
	defer t.mu.RUnlock()
	// Return a copy to prevent external modification
	actions := make([]ActionType, len(t.availableActions))
	copy(actions, t.availableActions)
	return actions
}

// ActionsRemaining returns the number of actions remaining in this turn
func (t *Turn) ActionsRemaining() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.actionsRemaining
}

// CanPerformAction checks if the specified action type is available
func (t *Turn) CanPerformAction(actionType ActionType) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, action := range t.availableActions {
		if action == actionType {
			return true
		}
	}
	return false
}

// DecrementActions decreases the number of actions remaining
func (t *Turn) DecrementActions() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.actionsRemaining > 0 {
		t.actionsRemaining--
	}
}

// SetActions updates the available actions and resets the remaining count
func (t *Turn) SetActions(actions []ActionType) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.availableActions = make([]ActionType, len(actions))
	copy(t.availableActions, actions)
	t.actionsRemaining = len(actions)
}
