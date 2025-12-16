package player

import (
	"sync"
)

// Effects manages passive effects from played cards
// Note: RequirementModifiers have been removed - discounts are now calculated on-demand
// via RequirementModifierCalculator during EntityState calculation
type Effects struct {
	mu      sync.RWMutex
	effects []CardEffect
}

// NewEffects creates a new Effects instance
func NewEffects() *Effects {
	return &Effects{
		effects: []CardEffect{},
	}
}

func newEffects() *Effects {
	return NewEffects()
}

func (e *Effects) List() []CardEffect {
	e.mu.RLock()
	defer e.mu.RUnlock()
	effectsCopy := make([]CardEffect, len(e.effects))
	copy(effectsCopy, e.effects)
	return effectsCopy
}

func (e *Effects) SetEffects(effects []CardEffect) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if effects == nil {
		e.effects = []CardEffect{}
	} else {
		e.effects = make([]CardEffect, len(effects))
		copy(e.effects, effects)
	}
}

func (e *Effects) AddEffect(effect CardEffect) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.effects = append(e.effects, effect)
}
