package model

// PlayerAction represents an action that a player can take, typically from a card with manual triggers
type PlayerAction struct {
	CardID        string       `json:"cardId" ts:"string"`         // ID of the card that provides this action
	CardName      string       `json:"cardName" ts:"string"`       // Name of the card for display purposes
	BehaviorIndex int          `json:"behaviorIndex" ts:"number"`  // Which behavior on the card this action represents
	Behavior      CardBehavior `json:"behavior" ts:"CardBehavior"` // The actual behavior definition with inputs/outputs
	PlayCount     int          `json:"playCount" ts:"number"`      // Number of times this action has been played this generation
}

// DeepCopy creates a deep copy of the PlayerAction
func (pa *PlayerAction) DeepCopy() *PlayerAction {
	if pa == nil {
		return nil
	}

	// Deep copy the behavior
	var behaviorCopy CardBehavior

	// Copy triggers slice
	if pa.Behavior.Triggers != nil {
		behaviorCopy.Triggers = make([]Trigger, len(pa.Behavior.Triggers))
		for i, trigger := range pa.Behavior.Triggers {
			behaviorCopy.Triggers[i] = Trigger{
				Type: trigger.Type,
			}
			// Deep copy condition if it exists
			if trigger.Condition != nil {
				conditionCopy := &ResourceTriggerCondition{
					Type:     trigger.Condition.Type,
					Location: trigger.Condition.Location,
				}
				// Copy affected tags slice
				if trigger.Condition.AffectedTags != nil {
					conditionCopy.AffectedTags = make([]CardTag, len(trigger.Condition.AffectedTags))
					copy(conditionCopy.AffectedTags, trigger.Condition.AffectedTags)
				}
				behaviorCopy.Triggers[i].Condition = conditionCopy
			}
		}
	}

	// Copy inputs slice
	if pa.Behavior.Inputs != nil {
		behaviorCopy.Inputs = make([]ResourceCondition, len(pa.Behavior.Inputs))
		for i, input := range pa.Behavior.Inputs {
			behaviorCopy.Inputs[i] = ResourceCondition{
				Type:       input.Type,
				Amount:     input.Amount,
				Target:     input.Target,
				MaxTrigger: input.MaxTrigger,
			}
			// Copy affected resources slice
			if input.AffectedResources != nil {
				behaviorCopy.Inputs[i].AffectedResources = make([]string, len(input.AffectedResources))
				copy(behaviorCopy.Inputs[i].AffectedResources, input.AffectedResources)
			}
			// Copy affected tags slice
			if input.AffectedTags != nil {
				behaviorCopy.Inputs[i].AffectedTags = make([]CardTag, len(input.AffectedTags))
				copy(behaviorCopy.Inputs[i].AffectedTags, input.AffectedTags)
			}
			// Deep copy per condition if it exists
			if input.Per != nil {
				behaviorCopy.Inputs[i].Per = &PerCondition{
					Type:     input.Per.Type,
					Amount:   input.Per.Amount,
					Location: input.Per.Location,
					Target:   input.Per.Target,
					Tag:      input.Per.Tag,
				}
			}
		}
	}

	// Copy outputs slice
	if pa.Behavior.Outputs != nil {
		behaviorCopy.Outputs = make([]ResourceCondition, len(pa.Behavior.Outputs))
		for i, output := range pa.Behavior.Outputs {
			behaviorCopy.Outputs[i] = ResourceCondition{
				Type:       output.Type,
				Amount:     output.Amount,
				Target:     output.Target,
				MaxTrigger: output.MaxTrigger,
			}
			// Copy affected resources slice
			if output.AffectedResources != nil {
				behaviorCopy.Outputs[i].AffectedResources = make([]string, len(output.AffectedResources))
				copy(behaviorCopy.Outputs[i].AffectedResources, output.AffectedResources)
			}
			// Copy affected tags slice
			if output.AffectedTags != nil {
				behaviorCopy.Outputs[i].AffectedTags = make([]CardTag, len(output.AffectedTags))
				copy(behaviorCopy.Outputs[i].AffectedTags, output.AffectedTags)
			}
			// Deep copy per condition if it exists
			if output.Per != nil {
				behaviorCopy.Outputs[i].Per = &PerCondition{
					Type:     output.Per.Type,
					Amount:   output.Per.Amount,
					Location: output.Per.Location,
					Target:   output.Per.Target,
					Tag:      output.Per.Tag,
				}
			}
		}
	}

	// Copy choices slice
	if pa.Behavior.Choices != nil {
		behaviorCopy.Choices = make([]Choice, len(pa.Behavior.Choices))
		for i, choice := range pa.Behavior.Choices {
			// Copy inputs slice for this choice
			if choice.Inputs != nil {
				behaviorCopy.Choices[i].Inputs = make([]ResourceCondition, len(choice.Inputs))
				copy(behaviorCopy.Choices[i].Inputs, choice.Inputs)
			}
			// Copy outputs slice for this choice
			if choice.Outputs != nil {
				behaviorCopy.Choices[i].Outputs = make([]ResourceCondition, len(choice.Outputs))
				copy(behaviorCopy.Choices[i].Outputs, choice.Outputs)
			}
		}
	}

	return &PlayerAction{
		CardID:        pa.CardID,
		CardName:      pa.CardName,
		BehaviorIndex: pa.BehaviorIndex,
		Behavior:      behaviorCopy,
		PlayCount:     pa.PlayCount,
	}
}
