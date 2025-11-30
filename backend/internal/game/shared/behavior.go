package shared

// ==================== Card Behaviors ====================

// CardBehavior represents card behaviors (immediate and repeatable)
type CardBehavior struct {
	Triggers []Trigger           `json:"triggers,omitempty"`
	Inputs   []ResourceCondition `json:"inputs,omitempty"`
	Outputs  []ResourceCondition `json:"outputs,omitempty"`
	Choices  []Choice            `json:"choices,omitempty"`
}

// DeepCopy creates a deep copy of the CardBehavior
func (cb CardBehavior) DeepCopy() CardBehavior {
	var result CardBehavior

	if cb.Triggers != nil {
		result.Triggers = make([]Trigger, len(cb.Triggers))
		for i, trigger := range cb.Triggers {
			result.Triggers[i] = trigger
		}
	}

	if cb.Inputs != nil {
		result.Inputs = make([]ResourceCondition, len(cb.Inputs))
		for i, input := range cb.Inputs {
			result.Inputs[i] = deepCopyResourceCondition(input)
		}
	}

	if cb.Outputs != nil {
		result.Outputs = make([]ResourceCondition, len(cb.Outputs))
		for i, output := range cb.Outputs {
			result.Outputs[i] = deepCopyResourceCondition(output)
		}
	}

	if cb.Choices != nil {
		result.Choices = make([]Choice, len(cb.Choices))
		for i, choice := range cb.Choices {
			choiceCopy := Choice{}

			if choice.Inputs != nil {
				choiceCopy.Inputs = make([]ResourceCondition, len(choice.Inputs))
				for j, input := range choice.Inputs {
					choiceCopy.Inputs[j] = deepCopyResourceCondition(input)
				}
			}

			if choice.Outputs != nil {
				choiceCopy.Outputs = make([]ResourceCondition, len(choice.Outputs))
				for j, output := range choice.Outputs {
					choiceCopy.Outputs[j] = deepCopyResourceCondition(output)
				}
			}

			result.Choices[i] = choiceCopy
		}
	}

	return result
}

// deepCopyResourceCondition creates a deep copy of a ResourceCondition
func deepCopyResourceCondition(rc ResourceCondition) ResourceCondition {
	result := rc

	if rc.AffectedResources != nil {
		result.AffectedResources = make([]string, len(rc.AffectedResources))
		copy(result.AffectedResources, rc.AffectedResources)
	}

	if rc.AffectedTags != nil {
		result.AffectedTags = make([]CardTag, len(rc.AffectedTags))
		copy(result.AffectedTags, rc.AffectedTags)
	}

	if rc.AffectedCardTypes != nil {
		result.AffectedCardTypes = make([]string, len(rc.AffectedCardTypes))
		copy(result.AffectedCardTypes, rc.AffectedCardTypes)
	}

	if rc.AffectedStandardProjects != nil {
		result.AffectedStandardProjects = make([]StandardProject, len(rc.AffectedStandardProjects))
		copy(result.AffectedStandardProjects, rc.AffectedStandardProjects)
	}

	if rc.Per != nil {
		perCopy := *rc.Per
		result.Per = &perCopy
	}

	return result
}
