package game

import (
	"sync"
)

// Actions manages available manual actions
type Actions struct {
	mu      sync.RWMutex
	actions []PlayerAction
}

func newActions() *Actions {
	return &Actions{
		actions: []PlayerAction{},
	}
}

func (a *Actions) List() []PlayerAction {
	a.mu.RLock()
	defer a.mu.RUnlock()
	actionsCopy := make([]PlayerAction, len(a.actions))
	copy(actionsCopy, a.actions)
	return actionsCopy
}

func (a *Actions) SetActions(actions []PlayerAction) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if actions == nil {
		a.actions = []PlayerAction{}
	} else {
		a.actions = make([]PlayerAction, len(actions))
		copy(a.actions, actions)
	}
}

func (a *Actions) AddAction(action PlayerAction) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.actions = append(a.actions, action)
}
