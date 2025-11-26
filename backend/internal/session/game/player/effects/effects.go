package effects

import (
	"sync"

	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/types"
)

// RequirementModifier represents a discount or lenience that modifies card/standard project requirements
type RequirementModifier struct {
	Amount                int                    // Modifier amount (discount/lenience value)
	AffectedResources     []types.ResourceType   // Resources affected (e.g., ["credits"] for price discount, ["temperature"] for global param)
	CardTarget            *string                // Optional: specific card ID this applies to
	StandardProjectTarget *types.StandardProject // Optional: specific standard project this applies to
}

// Effects manages passive effects from cards and requirement modifiers.
// Thread-safe with its own mutex.
type Effects struct {
	mu                   sync.RWMutex
	effects              []card.PlayerEffect   // Passive effects from played cards
	requirementModifiers []RequirementModifier // Discounts and requirement modifications
}

// NewEffects creates a new Effects component with empty collections.
func NewEffects() *Effects {
	return &Effects{
		effects:              []card.PlayerEffect{},
		requirementModifiers: []RequirementModifier{},
	}
}

// ==================== Getters ====================

// List returns a defensive copy of player effects.
func (e *Effects) List() []card.PlayerEffect {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.effects == nil {
		return []card.PlayerEffect{}
	}
	effectsCopy := make([]card.PlayerEffect, len(e.effects))
	copy(effectsCopy, e.effects)
	return effectsCopy
}

// RequirementModifiers returns a defensive copy of requirement modifiers.
func (e *Effects) RequirementModifiers() []RequirementModifier {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.requirementModifiers == nil {
		return []RequirementModifier{}
	}
	modifiersCopy := make([]RequirementModifier, len(e.requirementModifiers))
	copy(modifiersCopy, e.requirementModifiers)
	return modifiersCopy
}

// ==================== Setters ====================

// SetEffects replaces the entire effects collection.
func (e *Effects) SetEffects(effects []card.PlayerEffect) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if effects == nil {
		e.effects = []card.PlayerEffect{}
		return
	}
	e.effects = make([]card.PlayerEffect, len(effects))
	copy(e.effects, effects)
}

// SetRequirementModifiers replaces the entire requirement modifiers collection.
func (e *Effects) SetRequirementModifiers(modifiers []RequirementModifier) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if modifiers == nil {
		e.requirementModifiers = []RequirementModifier{}
		return
	}
	e.requirementModifiers = make([]RequirementModifier, len(modifiers))
	copy(e.requirementModifiers, modifiers)
}

// ==================== Mutations ====================

// AddEffect adds a new player effect to the collection.
func (e *Effects) AddEffect(effect card.PlayerEffect) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.effects = append(e.effects, effect)
}

// AddRequirementModifier adds a new requirement modifier to the collection.
func (e *Effects) AddRequirementModifier(modifier RequirementModifier) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.requirementModifiers = append(e.requirementModifiers, modifier)
}

// ClearEffects removes all effects.
func (e *Effects) ClearEffects() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.effects = []card.PlayerEffect{}
}

// ClearRequirementModifiers removes all requirement modifiers.
func (e *Effects) ClearRequirementModifiers() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.requirementModifiers = []RequirementModifier{}
}

// ==================== Utilities ====================

// DeepCopy creates a deep copy of the Effects component.
func (e *Effects) DeepCopy() *Effects {
	if e == nil {
		return nil
	}

	effectsCopy := make([]card.PlayerEffect, len(e.effects))
	copy(effectsCopy, e.effects)

	modifiersCopy := make([]RequirementModifier, len(e.requirementModifiers))
	for i, mod := range e.requirementModifiers {
		modifiersCopy[i] = RequirementModifier{
			Amount:                mod.Amount,
			StandardProjectTarget: mod.StandardProjectTarget,
		}
		// Deep copy affected resources
		if mod.AffectedResources != nil {
			modifiersCopy[i].AffectedResources = make([]types.ResourceType, len(mod.AffectedResources))
			copy(modifiersCopy[i].AffectedResources, mod.AffectedResources)
		}
		// Deep copy card target pointer
		if mod.CardTarget != nil {
			cardTargetCopy := *mod.CardTarget
			modifiersCopy[i].CardTarget = &cardTargetCopy
		}
	}

	return &Effects{
		effects:              effectsCopy,
		requirementModifiers: modifiersCopy,
	}
}
