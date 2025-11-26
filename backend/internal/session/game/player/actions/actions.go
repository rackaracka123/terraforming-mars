package actions

import (
	"sync"

	"terraforming-mars-backend/internal/session/game/card"
)

// PlayerAction represents an action that a player can take, typically from a card with manual triggers
type PlayerAction struct {
	CardID        string            // ID of the card that provides this action
	CardName      string            // Name of the card for display purposes
	BehaviorIndex int               // Which behavior on the card this action represents
	Behavior      card.CardBehavior // The actual behavior definition with inputs/outputs
	PlayCount     int               // Number of times this action has been played this generation
}

// Actions manages available actions that a player can take from cards.
// Thread-safe with its own mutex.
type Actions struct {
	mu      sync.RWMutex
	actions []PlayerAction // Manual actions available from cards
}

// NewActions creates a new Actions component with an empty collection.
func NewActions() *Actions {
	return &Actions{
		actions: []PlayerAction{},
	}
}

// ==================== Getters ====================

// List returns a defensive copy of available actions.
func (a *Actions) List() []PlayerAction {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.actions == nil {
		return []PlayerAction{}
	}
	actionsCopy := make([]PlayerAction, len(a.actions))
	copy(actionsCopy, a.actions)
	return actionsCopy
}

// ==================== Setters ====================

// SetActions replaces the entire actions collection.
func (a *Actions) SetActions(actions []PlayerAction) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if actions == nil {
		a.actions = []PlayerAction{}
		return
	}
	a.actions = make([]PlayerAction, len(actions))
	copy(a.actions, actions)
}

// ==================== Mutations ====================

// AddAction adds a new action to the collection.
func (a *Actions) AddAction(action PlayerAction) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.actions = append(a.actions, action)
}

// ClearActions removes all actions.
func (a *Actions) ClearActions() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.actions = []PlayerAction{}
}

// ==================== Utilities ====================

// DeepCopy creates a deep copy of the Actions component.
func (a *Actions) DeepCopy() *Actions {
	if a == nil {
		return nil
	}

	actionsCopy := make([]PlayerAction, len(a.actions))
	copy(actionsCopy, a.actions)

	return &Actions{
		actions: actionsCopy,
	}
}
