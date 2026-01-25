package player

import (
	"sync"

	"terraforming-mars-backend/internal/events"
)

// Effects manages passive effects from played cards
// Note: RequirementModifiers have been removed - discounts are now calculated on-demand
// via RequirementModifierCalculator during EntityState calculation
type Effects struct {
	mu            sync.RWMutex
	effects       []CardEffect
	subscriptions map[string][]events.SubscriptionID
	eventBus      *events.EventBusImpl
}

// NewEffects creates a new Effects instance
func NewEffects(eventBus *events.EventBusImpl) *Effects {
	return &Effects{
		effects:       []CardEffect{},
		subscriptions: make(map[string][]events.SubscriptionID),
		eventBus:      eventBus,
	}
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

// RegisterSubscription tracks an event subscription for a card so it can be unsubscribed later
func (e *Effects) RegisterSubscription(cardID string, subID events.SubscriptionID) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.subscriptions[cardID] = append(e.subscriptions[cardID], subID)
}

// RemoveEffectsByCardID removes all effects from a specific card and unsubscribes from events
func (e *Effects) RemoveEffectsByCardID(cardID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	filtered := make([]CardEffect, 0, len(e.effects))
	for _, effect := range e.effects {
		if effect.CardID != cardID {
			filtered = append(filtered, effect)
		}
	}
	e.effects = filtered

	if subs, exists := e.subscriptions[cardID]; exists {
		for _, subID := range subs {
			e.eventBus.Unsubscribe(subID)
		}
		delete(e.subscriptions, cardID)
	}
}
