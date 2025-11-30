package player

import (
	"sync"

	"terraforming-mars-backend/internal/game/shared"
)

// Effects manages passive effects and requirement modifiers
type Effects struct {
	mu                   sync.RWMutex
	effects              []CardEffect
	requirementModifiers []shared.RequirementModifier
}

// NewEffects creates a new Effects instance
func NewEffects() *Effects {
	return &Effects{
		effects:              []CardEffect{},
		requirementModifiers: []shared.RequirementModifier{},
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

func (e *Effects) RequirementModifiers() []shared.RequirementModifier {
	e.mu.RLock()
	defer e.mu.RUnlock()
	modifiersCopy := make([]shared.RequirementModifier, len(e.requirementModifiers))
	copy(modifiersCopy, e.requirementModifiers)
	return modifiersCopy
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
