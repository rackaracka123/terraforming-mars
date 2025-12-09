package player

import (
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/playability"
	"terraforming-mars-backend/internal/game/shared"
)

// CardEffect represents an ongoing effect defined by a card
type CardEffect struct {
	CardID        string
	CardName      string
	BehaviorIndex int
	Behavior      shared.CardBehavior
}

// DeepCopy creates a deep copy of the CardEffect
func (pe *CardEffect) DeepCopy() *CardEffect {
	if pe == nil {
		return nil
	}

	return &CardEffect{
		CardID:        pe.CardID,
		CardName:      pe.CardName,
		BehaviorIndex: pe.BehaviorIndex,
		Behavior:      pe.Behavior.DeepCopy(),
	}
}

// CardAction represents a repeatable manual action defined by a card
// Each CardAction subscribes to relevant events and maintains its own availability state
type CardAction struct {
	CardID        string
	CardName      string
	BehaviorIndex int
	Behavior      shared.CardBehavior
	PlayCount     int

	// Playability state (self-contained)
	availability   playability.ActionPlayabilityResult
	availabilityMu sync.RWMutex
	subscriptionIDs []events.SubscriptionID

	// Validation function to check availability
	checkAvailability func() playability.ActionPlayabilityResult
}

// DeepCopy creates a deep copy of the CardAction
func (pa *CardAction) DeepCopy() *CardAction {
	if pa == nil {
		return nil
	}

	return &CardAction{
		CardID:        pa.CardID,
		CardName:      pa.CardName,
		BehaviorIndex: pa.BehaviorIndex,
		Behavior:      pa.Behavior.DeepCopy(),
		PlayCount:     pa.PlayCount,
		// Note: Don't copy availability, subscriptions, or checkAvailability
		// These are runtime state that should be reinitialized
	}
}

// GetAvailability returns the current availability state
func (pa *CardAction) GetAvailability() playability.ActionPlayabilityResult {
	pa.availabilityMu.RLock()
	defer pa.availabilityMu.RUnlock()
	return pa.availability
}

// recalculate updates the availability state by calling the check function
func (pa *CardAction) recalculate() {
	if pa.checkAvailability == nil {
		return
	}

	newAvailability := pa.checkAvailability()

	pa.availabilityMu.Lock()
	pa.availability = newAvailability
	pa.availabilityMu.Unlock()
}
