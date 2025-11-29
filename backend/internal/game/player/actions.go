package player

import (
	"sync"

	"terraforming-mars-backend/internal/game/cardtypes"
)

// Actions manages available manual actions
type Actions struct {
	mu      sync.RWMutex
	actions []cardtypes.CardAction
}

// NewActions creates a new Actions instance
func NewActions() *Actions {
	return &Actions{
		actions: []cardtypes.CardAction{},
	}
}

func newActions() *Actions {
	return NewActions()
}

func (a *Actions) List() []cardtypes.CardAction {
	a.mu.RLock()
	defer a.mu.RUnlock()
	actionsCopy := make([]cardtypes.CardAction, len(a.actions))
	copy(actionsCopy, a.actions)
	return actionsCopy
}

func (a *Actions) SetActions(actions []cardtypes.CardAction) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if actions == nil {
		a.actions = []cardtypes.CardAction{}
	} else {
		a.actions = make([]cardtypes.CardAction, len(actions))
		copy(a.actions, actions)
	}
}

func (a *Actions) AddAction(action cardtypes.CardAction) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.actions = append(a.actions, action)
}
