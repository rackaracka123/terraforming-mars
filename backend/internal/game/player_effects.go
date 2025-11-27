package game

import (
	"sync"
	"terraforming-mars-backend/internal/game/shared"
)

// Effects manages passive effects and requirement modifiers
type Effects struct {
	mu                   sync.RWMutex
	effects              []PlayerEffect
	requirementModifiers []shared.RequirementModifier
}

// NewEffects creates a new Effects instance
func NewEffects() *Effects {
	return &Effects{
		effects:              []PlayerEffect{},
		requirementModifiers: []shared.RequirementModifier{},
	}
}

func newEffects() *Effects {
	return NewEffects()
}

func (e *Effects) List() []PlayerEffect {
	e.mu.RLock()
	defer e.mu.RUnlock()
	effectsCopy := make([]PlayerEffect, len(e.effects))
	copy(effectsCopy, e.effects)
	return effectsCopy
}

func (e *Effects) RequirementModifiers() []shared.RequirementModifier {
	e.mu.RLock()
	defer e.mu.RUnlock()
	modifiersCopy := make([]shared.RequirementModifier, len(e.requirementModifiers))
	copy(modifiersCopy, e.requirementModifiers)
	return modifiersCopy
}

func (e *Effects) SetEffects(effects []PlayerEffect) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if effects == nil {
		e.effects = []PlayerEffect{}
	} else {
		e.effects = make([]PlayerEffect, len(effects))
		copy(e.effects, effects)
	}
}

func (e *Effects) AddEffect(effect PlayerEffect) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.effects = append(e.effects, effect)
}
