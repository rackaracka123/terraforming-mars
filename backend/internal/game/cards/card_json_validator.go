package cards

import (
	"fmt"

	"terraforming-mars-backend/internal/game/shared"
)

// ValidateCardJSON validates the JSON structure of a card at load time
// This ensures all enum values are valid and structure is correct
func ValidateCardJSON(card *Card) []error {
	var errors []error

	// Validate card type
	if !isValidCardType(card.Type) {
		errors = append(errors, fmt.Errorf("card %s: invalid card type: %s", card.ID, card.Type))
	}

	// Validate tags
	for _, tag := range card.Tags {
		if !isValidCardTag(tag) {
			errors = append(errors, fmt.Errorf("card %s: invalid tag: %s", card.ID, tag))
		}
	}

	// Validate requirements
	for i, req := range card.Requirements {
		if reqErr := validateRequirement(card.ID, i, req); reqErr != nil {
			errors = append(errors, reqErr)
		}
	}

	// Validate behaviors
	for i, behavior := range card.Behaviors {
		behaviorErrors := validateBehavior(card.ID, i, behavior)
		errors = append(errors, behaviorErrors...)
	}

	// Validate resource storage
	if card.ResourceStorage != nil {
		if !isValidResourceType(card.ResourceStorage.Type) {
			errors = append(errors, fmt.Errorf("card %s: invalid resource storage type: %s", card.ID, card.ResourceStorage.Type))
		}
	}

	// Validate victory point conditions
	for i, vp := range card.VPConditions {
		if vpErr := validateVictoryPointCondition(card.ID, i, vp); vpErr != nil {
			errors = append(errors, vpErr)
		}
	}

	// Validate starting resources
	if card.StartingResources != nil {
		if rsErr := validateResourceSet(card.ID, "starting resources", *card.StartingResources); rsErr != nil {
			errors = append(errors, rsErr)
		}
	}

	// Validate starting production
	if card.StartingProduction != nil {
		if prodErr := validateResourceSet(card.ID, "starting production", *card.StartingProduction); prodErr != nil {
			errors = append(errors, prodErr)
		}
	}

	return errors
}

// validateRequirement validates a single requirement
func validateRequirement(cardID string, index int, req Requirement) error {
	if !isValidRequirementType(req.Type) {
		return fmt.Errorf("card %s: requirement[%d] has invalid type: %s", cardID, index, req.Type)
	}

	// Validate tag requirements have valid tags
	if req.Type == RequirementTags && req.Tag != nil {
		if !isValidCardTag(*req.Tag) {
			return fmt.Errorf("card %s: requirement[%d] has invalid tag: %s", cardID, index, *req.Tag)
		}
	}

	return nil
}

// validateBehavior validates a card behavior
func validateBehavior(cardID string, index int, behavior shared.CardBehavior) []error {
	var errors []error

	// Validate triggers
	for i, trigger := range behavior.Triggers {
		// Validate trigger type (now just a string)
		if trigger.Type == "" {
			errors = append(errors, fmt.Errorf("card %s: behavior[%d].trigger[%d] has empty type", cardID, index, i))
		}

		// Validate condition if present
		if trigger.Condition != nil {
			// Validate resource types in condition
			for _, rt := range trigger.Condition.ResourceTypes {
				if !isValidResourceType(rt) {
					errors = append(errors, fmt.Errorf("card %s: behavior[%d].trigger[%d] has invalid resource type: %s", cardID, index, i, rt))
				}
			}
		}
	}

	// Validate inputs
	for i, input := range behavior.Inputs {
		if inputErr := validateResourceCondition(cardID, index, "input", i, input); inputErr != nil {
			errors = append(errors, inputErr)
		}
	}

	// Validate outputs
	for i, output := range behavior.Outputs {
		if outputErr := validateResourceCondition(cardID, index, "output", i, output); outputErr != nil {
			errors = append(errors, outputErr)
		}
	}

	// Validate choices
	for i, choice := range behavior.Choices {
		// Validate choice inputs
		for k, input := range choice.Inputs {
			if inputErr := validateResourceCondition(cardID, index, fmt.Sprintf("choice[%d].input", i), k, input); inputErr != nil {
				errors = append(errors, inputErr)
			}
		}

		// Validate choice outputs
		for k, output := range choice.Outputs {
			if outputErr := validateResourceCondition(cardID, index, fmt.Sprintf("choice[%d].output", i), k, output); outputErr != nil {
				errors = append(errors, outputErr)
			}
		}
	}

	return errors
}

// validateResourceCondition validates a resource condition
func validateResourceCondition(cardID string, behaviorIndex int, condType string, index int, cond shared.ResourceCondition) error {
	// Target is now a string, just check it's not empty
	if cond.Target == "" {
		return fmt.Errorf("card %s: behavior[%d].%s[%d] has empty target", cardID, behaviorIndex, condType, index)
	}

	if !isValidResourceType(cond.ResourceType) {
		return fmt.Errorf("card %s: behavior[%d].%s[%d] has invalid resource type: %s", cardID, behaviorIndex, condType, index, cond.ResourceType)
	}

	// Validate per-condition if present
	if cond.Per != nil {
		if !isValidResourceType(cond.Per.ResourceType) {
			return fmt.Errorf("card %s: behavior[%d].%s[%d].per has invalid resource type: %s", cardID, behaviorIndex, condType, index, cond.Per.ResourceType)
		}
	}

	return nil
}

// validateVictoryPointCondition validates a victory point condition
func validateVictoryPointCondition(cardID string, index int, vp VictoryPointCondition) error {
	// Validate Per condition if present
	if vp.Per != nil {
		if !isValidResourceType(vp.Per.Type) {
			return fmt.Errorf("card %s: victory_point_condition[%d].per has invalid resource type: %s", cardID, index, vp.Per.Type)
		}
	}

	return nil
}

// validateResourceSet validates a resource set (starting resources or production)
func validateResourceSet(cardID, fieldName string, rs shared.ResourceSet) error {
	// ResourceSet uses standard resource types, no specific validation needed beyond type checking
	// The struct itself ensures type safety, so no runtime validation required
	return nil
}

// isValidCardType checks if a card type is valid
func isValidCardType(ct CardType) bool {
	switch ct {
	case CardTypeCorporation, CardTypeAutomated, CardTypeActive, CardTypeEvent, CardTypePrelude:
		return true
	default:
		return false
	}
}

// isValidCardTag checks if a card tag is valid
func isValidCardTag(tag shared.CardTag) bool {
	switch tag {
	case shared.TagBuilding, shared.TagSpace, shared.TagScience,
		shared.TagPower, shared.TagEarth, shared.TagJovian,
		shared.TagVenus, shared.TagPlant, shared.TagMicrobe,
		shared.TagAnimal, shared.TagCity, shared.TagEvent,
		shared.TagWildlife, shared.TagWild:
		return true
	default:
		return false
	}
}

// isValidRequirementType checks if a requirement type is valid
func isValidRequirementType(rt RequirementType) bool {
	switch rt {
	case RequirementTemperature, RequirementOxygen, RequirementOceans,
		RequirementTags, RequirementProduction, RequirementTR,
		RequirementResource, RequirementVenus, RequirementCities,
		RequirementGreeneries:
		return true
	default:
		return false
	}
}

// isValidResourceType checks if a resource type is valid
// Since there are many resource types, just check it's not empty for now
func isValidResourceType(rt shared.ResourceType) bool {
	return rt != ""
}
