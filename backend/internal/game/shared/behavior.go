package shared

// CardBehavior represents card behaviors (immediate and repeatable)
type CardBehavior struct {
	Description                   string                         `json:"description,omitempty" ts:"string | undefined"`
	Triggers                      []Trigger                      `json:"triggers,omitempty"`
	Inputs                        []ResourceCondition            `json:"inputs,omitempty"`
	Outputs                       []ResourceCondition            `json:"outputs,omitempty"`
	Choices                       []Choice                       `json:"choices,omitempty"`
	GenerationalEventRequirements []GenerationalEventRequirement `json:"generationalEventRequirements,omitempty" ts:"GenerationalEventRequirement[] | undefined"`
}

// DeepCopy creates a deep copy of the CardBehavior
func (cb CardBehavior) DeepCopy() CardBehavior {
	var result CardBehavior

	result.Description = cb.Description

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

	if cb.GenerationalEventRequirements != nil {
		result.GenerationalEventRequirements = make([]GenerationalEventRequirement, len(cb.GenerationalEventRequirements))
		for i, req := range cb.GenerationalEventRequirements {
			result.GenerationalEventRequirements[i] = deepCopyGenerationalEventRequirement(req)
		}
	}

	return result
}

// ExtractInputsOutputs extracts the combined inputs and outputs for a behavior,
// optionally incorporating a selected choice. Returns base + choice inputs/outputs.
// If choiceIndex is nil or out of range, only base inputs/outputs are returned.
func (cb CardBehavior) ExtractInputsOutputs(choiceIndex *int) (inputs []ResourceCondition, outputs []ResourceCondition) {
	if len(cb.Inputs) > 0 {
		inputs = make([]ResourceCondition, len(cb.Inputs))
		copy(inputs, cb.Inputs)
	}
	if len(cb.Outputs) > 0 {
		outputs = make([]ResourceCondition, len(cb.Outputs))
		copy(outputs, cb.Outputs)
	}

	if choiceIndex != nil && *choiceIndex >= 0 && *choiceIndex < len(cb.Choices) {
		selectedChoice := cb.Choices[*choiceIndex]

		// Append choice inputs to base inputs
		if len(selectedChoice.Inputs) > 0 {
			inputs = append(inputs, selectedChoice.Inputs...)
		}

		// Append choice outputs to base outputs
		if len(selectedChoice.Outputs) > 0 {
			outputs = append(outputs, selectedChoice.Outputs...)
		}
	}

	return inputs, outputs
}

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

func deepCopyGenerationalEventRequirement(req GenerationalEventRequirement) GenerationalEventRequirement {
	result := GenerationalEventRequirement{
		Event: req.Event,
	}

	if req.Count != nil {
		countCopy := MinMax{}
		if req.Count.Min != nil {
			minCopy := *req.Count.Min
			countCopy.Min = &minCopy
		}
		if req.Count.Max != nil {
			maxCopy := *req.Count.Max
			countCopy.Max = &maxCopy
		}
		result.Count = &countCopy
	}

	if req.Target != nil {
		targetCopy := *req.Target
		result.Target = &targetCopy
	}

	return result
}
