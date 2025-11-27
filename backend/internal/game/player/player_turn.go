package player

import "sync"

// Turn manages player turn state
type Turn struct {
	mu               sync.RWMutex
	passed           bool
	availableActions int
	isConnected      bool
}

func newTurn() *Turn {
	return &Turn{
		passed:           false,
		availableActions: 0,
		isConnected:      false,
	}
}

func (t *Turn) Passed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.passed
}

func (t *Turn) AvailableActions() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.availableActions
}

func (t *Turn) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isConnected
}

func (t *Turn) SetPassed(passed bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.passed = passed
}

func (t *Turn) SetAvailableActions(actions int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.availableActions = actions
}

func (t *Turn) SetConnectionStatus(isConnected bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isConnected = isConnected
}

func (t *Turn) ConsumeAction() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.availableActions > 0 {
		t.availableActions--
		return true
	}
	return false
}
