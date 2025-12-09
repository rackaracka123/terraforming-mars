package player

import (
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/playability"
)

// Actions manages available manual actions
// Each action is self-contained and maintains its own availability via event subscriptions
type Actions struct {
	mu      sync.RWMutex
	actions []*CardAction

	// Infrastructure for event subscriptions
	eventBus *events.EventBusImpl
	playerID string
}

// NewActions creates a new Actions instance
func NewActions() *Actions {
	return &Actions{
		actions: []*CardAction{},
	}
}

func newActions() *Actions {
	return NewActions()
}

// SetInfrastructure injects event bus and player ID for action subscriptions
// Must be called before AddAction if actions need to subscribe to events
func (a *Actions) SetInfrastructure(eventBus *events.EventBusImpl, playerID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.eventBus = eventBus
	a.playerID = playerID
}

func (a *Actions) List() []*CardAction {
	a.mu.RLock()
	defer a.mu.RUnlock()
	actionsCopy := make([]*CardAction, len(a.actions))
	copy(actionsCopy, a.actions)
	return actionsCopy
}

func (a *Actions) SetActions(actions []*CardAction) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if actions == nil {
		a.actions = []*CardAction{}
	} else {
		a.actions = make([]*CardAction, len(actions))
		copy(a.actions, actions)
	}
}

// AddAction adds an action and subscribes it to relevant events
func (a *Actions) AddAction(action *CardAction) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.actions = append(a.actions, action)

	// Subscribe action to events if infrastructure is set up
	if a.eventBus != nil && a.playerID != "" {
		action.Subscribe(a.eventBus, a.playerID)
	}
}

// ResetPlayCounts resets the PlayCount for all actions to 0
// Called at the start of each new generation
func (a *Actions) ResetPlayCounts() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.actions {
		a.actions[i].PlayCount = 0
	}
}

// ==================== Playability Management ====================

// GetPlayability returns playability for all actions
// Each action maintains its own availability via event subscriptions
func (a *Actions) GetPlayability() []playability.ActionPlayabilityResult {
	a.mu.RLock()
	defer a.mu.RUnlock()

	results := make([]playability.ActionPlayabilityResult, len(a.actions))
	for i, action := range a.actions {
		results[i] = action.GetAvailability()
	}
	return results
}

// GetPlayabilityForAction returns playability for a specific action by index
func (a *Actions) GetPlayabilityForAction(actionIndex int) playability.ActionPlayabilityResult {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if actionIndex >= 0 && actionIndex < len(a.actions) {
		return a.actions[actionIndex].GetAvailability()
	}

	// Return unaffordable result if index out of bounds
	return playability.NewActionPlayabilityResult()
}
