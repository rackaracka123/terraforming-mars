package cards

import (
	"terraforming-mars-backend/internal/game/shared"
)

// HasAutoTrigger checks if a behavior has an auto trigger without conditions
func HasAutoTrigger(behavior shared.CardBehavior) bool {
	for _, trigger := range behavior.Triggers {
		if trigger.Type == string(ResourceTriggerAuto) && trigger.Condition == nil {
			return true
		}
	}
	return false
}

// HasManualTrigger checks if a behavior has a manual trigger
func HasManualTrigger(behavior shared.CardBehavior) bool {
	for _, trigger := range behavior.Triggers {
		if trigger.Type == string(ResourceTriggerManual) {
			return true
		}
	}
	return false
}

// HasConditionalTrigger checks if a behavior has an auto trigger with a condition (passive effect)
func HasConditionalTrigger(behavior shared.CardBehavior) bool {
	for _, trigger := range behavior.Triggers {
		if trigger.Type == string(ResourceTriggerAuto) && trigger.Condition != nil {
			return true
		}
	}
	return false
}

// HasCorporationStartTrigger checks if a behavior has a corporation start trigger
func HasCorporationStartTrigger(behavior shared.CardBehavior) bool {
	for _, trigger := range behavior.Triggers {
		if trigger.Type == string(ResourceTriggerAutoCorporationStart) {
			return true
		}
	}
	return false
}

// HasCorporationFirstActionTrigger checks if a behavior has a corporation first action trigger
func HasCorporationFirstActionTrigger(behavior shared.CardBehavior) bool {
	for _, trigger := range behavior.Triggers {
		if trigger.Type == string(ResourceTriggerAutoCorporationFirstAction) {
			return true
		}
	}
	return false
}

// GetImmediateBehaviors returns all behaviors with auto triggers (no conditions)
func GetImmediateBehaviors(card *Card) []shared.CardBehavior {
	var immediate []shared.CardBehavior
	for _, behavior := range card.Behaviors {
		if HasAutoTrigger(behavior) {
			immediate = append(immediate, behavior)
		}
	}
	return immediate
}

// GetManualBehaviors returns all behaviors with manual triggers
func GetManualBehaviors(card *Card) []shared.CardBehavior {
	var manual []shared.CardBehavior
	for _, behavior := range card.Behaviors {
		if HasManualTrigger(behavior) {
			manual = append(manual, behavior)
		}
	}
	return manual
}

// GetPassiveBehaviors returns all behaviors with conditional triggers
func GetPassiveBehaviors(card *Card) []shared.CardBehavior {
	var passive []shared.CardBehavior
	for _, behavior := range card.Behaviors {
		if HasConditionalTrigger(behavior) {
			passive = append(passive, behavior)
		}
	}
	return passive
}
