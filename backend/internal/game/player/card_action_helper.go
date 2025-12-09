package player

import (
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/playability"
	"terraforming-mars-backend/internal/game/shared"
)

// ActionAvailabilityChecker defines a function that checks if an action is available
type ActionAvailabilityChecker func() playability.ActionPlayabilityResult

// NewCardAction creates a new card action with availability checking
func NewCardAction(
	cardID string,
	cardName string,
	behaviorIndex int,
	behavior shared.CardBehavior,
	checkFunc ActionAvailabilityChecker,
) *CardAction {
	return &CardAction{
		CardID:            cardID,
		CardName:          cardName,
		BehaviorIndex:     behaviorIndex,
		Behavior:          behavior,
		PlayCount:         0,
		availability:      playability.NewActionPlayabilityResult(),
		subscriptionIDs:   []events.SubscriptionID{},
		checkAvailability: checkFunc,
	}
}

// Subscribe registers event handlers for this card action
// Card actions subscribe to resource changes to update affordability
func (ca *CardAction) Subscribe(eventBus *events.EventBusImpl, playerID string) {
	// All manual actions need to subscribe to resource changes
	// because they require resources to execute
	subID := events.Subscribe(eventBus, func(event events.ResourcesChangedEvent) {
		if event.PlayerID == playerID {
			ca.recalculate()
		}
	})
	ca.subscriptionIDs = append(ca.subscriptionIDs, subID)

	// Check for specific input/output types that require additional event subscriptions
	for _, input := range ca.Behavior.Inputs {
		switch input.ResourceType {
		case shared.ResourceHeat:
			// Subscribe to temperature changes if action uses heat
			// (Heat may be restricted if temperature is maxed)
			tempSubID := events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
				ca.recalculate()
			})
			ca.subscriptionIDs = append(ca.subscriptionIDs, tempSubID)
		}
	}

	// Initial calculation
	ca.recalculate()
}
