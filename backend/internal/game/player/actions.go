package player

import (
	"sync"
)

// Actions manages available manual actions
type Actions struct {
	mu      sync.RWMutex
	actions []CardAction
}

// NewActions creates a new Actions instance
func NewActions() *Actions {
	return &Actions{
		actions: []CardAction{},
	}
}

func newActions() *Actions {
	return NewActions()
}

func (a *Actions) List() []CardAction {
	a.mu.RLock()
	defer a.mu.RUnlock()
	actionsCopy := make([]CardAction, len(a.actions))
	copy(actionsCopy, a.actions)
	return actionsCopy
}

func (a *Actions) SetActions(actions []CardAction) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if actions == nil {
		a.actions = []CardAction{}
	} else {
		a.actions = make([]CardAction, len(actions))
		copy(a.actions, actions)
	}
}

func (a *Actions) AddAction(action CardAction) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.actions = append(a.actions, action)
}
