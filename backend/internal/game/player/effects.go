package player

import (
	"sync"

	"terraforming-mars-backend/internal/game/cardtypes"
	"terraforming-mars-backend/internal/game/shared"
)

// Effects manages passive effects and requirement modifiers
type Effects struct {
	mu                   sync.RWMutex
	effects              []cardtypes.CardEffect
	requirementModifiers []shared.RequirementModifier
}

// NewEffects creates a new Effects instance
func NewEffects() *Effects {
	return &Effects{
		effects:              []cardtypes.CardEffect{},
		requirementModifiers: []shared.RequirementModifier{},
	}
}

func newEffects() *Effects {
	return NewEffects()
}

func (e *Effects) List() []cardtypes.CardEffect {
	e.mu.RLock()
	defer e.mu.RUnlock()
	effectsCopy := make([]cardtypes.CardEffect, len(e.effects))
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

func (e *Effects) SetEffects(effects []cardtypes.CardEffect) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if effects == nil {
		e.effects = []cardtypes.CardEffect{}
	} else {
		e.effects = make([]cardtypes.CardEffect, len(effects))
		copy(e.effects, effects)
	}
}

func (e *Effects) AddEffect(effect cardtypes.CardEffect) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.effects = append(e.effects, effect)
}
