package game

import "sync"

// Turn tracks the active player and available actions for the current turn
// This is game-level turn state (whose turn it is), not per-player state
type Turn struct {
	mu               sync.RWMutex
	playerID         string
	actionsRemaining int // -1 = unlimited, 0 = none, >0 = specific count
}

// NewTurn creates a new turn for the specified player with a specific action count
func NewTurn(playerID string, actionsRemaining int) *Turn {
	return &Turn{
		playerID:         playerID,
		actionsRemaining: actionsRemaining,
	}
}

// PlayerID returns the ID of the player whose turn it is
func (t *Turn) PlayerID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.playerID
}

// ActionsRemaining returns the number of actions remaining in this turn
// Returns -1 for unlimited actions, 0 for no actions, >0 for specific count
func (t *Turn) ActionsRemaining() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.actionsRemaining
}

// SetActionsRemaining sets the number of actions remaining
func (t *Turn) SetActionsRemaining(actions int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.actionsRemaining = actions
}

// ConsumeAction decreases the number of actions remaining
// Returns true if an action was consumed, false if unlimited (-1) or no actions (0)
func (t *Turn) ConsumeAction() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Only consume if player has limited actions remaining (> 0)
	// -1 = unlimited actions (don't consume)
	// 0 = no actions remaining (don't consume)
	if t.actionsRemaining > 0 {
		t.actionsRemaining--
		return true
	}

	return false
}
