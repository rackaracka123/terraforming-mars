package cards

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"terraforming-mars-backend/internal/model"
)

// loadCards loads the parsed card JSON for testing
func loadCards(t *testing.T) []model.Card {
	t.Helper()

	// Read the generated JSON file
	data, err := os.ReadFile("../../assets/terraforming_mars_cards.json")
	if err != nil {
		t.Fatalf("Failed to read cards JSON: %v", err)
	}

	var cards []model.Card
	if err := json.Unmarshal(data, &cards); err != nil {
		t.Fatalf("Failed to unmarshal cards JSON: %v", err)
	}

	return cards
}

// TestAutoTriggerValidation validates Rule 1: If a behavior.trigger is set to "auto", it must not have any inputs defined
func TestAutoTriggerValidation(t *testing.T) {
	cards := loadCards(t)

	var violations []string

	for _, card := range cards {
		for behaviorIdx, behavior := range card.Behaviors {
			for triggerIdx, trigger := range behavior.Triggers {
				if trigger.Type == model.ResourceTriggerAuto {
					if len(behavior.Inputs) > 0 {
						violations = append(violations,
							fmt.Sprintf("Card %s behavior[%d] trigger[%d]: auto trigger has %d inputs defined",
								card.ID, behaviorIdx, triggerIdx, len(behavior.Inputs)))
					}
				}
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("Found %d auto trigger validation violations:", len(violations))
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}
}

// TestConflictingMinMaxRequirements validates Rule 2: A card may not have conflicting requirements for "min" and "max"
func TestConflictingMinMaxRequirements(t *testing.T) {
	cards := loadCards(t)

	var violations []string

	for _, card := range cards {
		for reqIdx, requirement := range card.Requirements {
			if requirement.Min != nil && requirement.Max != nil {
				minVal := *requirement.Min
				maxVal := *requirement.Max
				if minVal > maxVal {
					violations = append(violations,
						fmt.Sprintf("Card %s requirement[%d]: min value %d is greater than max value %d",
							card.ID, reqIdx, minVal, maxVal))
				}
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("Found %d conflicting min/max requirement violations:", len(violations))
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}
}

// TestRequirementConditionValidation validates Rule 3: A requirement must define a condition, like "min" or "max"
func TestRequirementConditionValidation(t *testing.T) {
	cards := loadCards(t)

	var violations []string

	for _, card := range cards {
		for reqIdx, requirement := range card.Requirements {
			if requirement.Min == nil && requirement.Max == nil {
				violations = append(violations,
					fmt.Sprintf("Card %s requirement[%d]: no condition defined (missing both min and max)",
						card.ID, reqIdx))
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("Found %d requirement condition validation violations:", len(violations))
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}
}

// TestVPConditionPerValidation validates Rule 4: "vpCondition.per" may not be set unless "vpCondition.condition" is "per"
func TestVPConditionPerValidation(t *testing.T) {
	cards := loadCards(t)

	var violations []string

	for _, card := range cards {
		for vpIdx, vpCondition := range card.VPConditions {
			if vpCondition.Per != nil && vpCondition.Condition != model.VPConditionPer {
				violations = append(violations,
					fmt.Sprintf("Card %s vpCondition[%d]: 'per' field is set but condition is '%s' (not 'per')",
						card.ID, vpIdx, string(vpCondition.Condition)))
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("Found %d VP condition per validation violations:", len(violations))
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}
}

// TestBehaviorChoiceValidation validates Rule 5: If a behavior.choice is present, there must be at least 2
func TestBehaviorChoiceValidation(t *testing.T) {
	cards := loadCards(t)

	var violations []string

	for _, card := range cards {
		for behaviorIdx, behavior := range card.Behaviors {
			if len(behavior.Choices) > 0 && len(behavior.Choices) < 2 {
				violations = append(violations,
					fmt.Sprintf("Card %s behavior[%d]: has %d choices but requires at least 2",
						card.ID, behaviorIdx, len(behavior.Choices)))
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("Found %d behavior choice validation violations:", len(violations))
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}
}

// TestDiscountBehaviorValidation validates Rule 6: If a behavior is "discount", it can't have any inputs, outputs, or choices
func TestDiscountBehaviorValidation(t *testing.T) {
	cards := loadCards(t)

	var violations []string

	for _, card := range cards {
		for behaviorIdx, behavior := range card.Behaviors {
			// Check if this is a discount behavior (has outputs with type "discount")
			isDiscountBehavior := false
			for _, output := range behavior.Outputs {
				if output.Type == "discount" {
					isDiscountBehavior = true
					break
				}
			}

			if isDiscountBehavior {
				// Discount behaviors should not have inputs
				if len(behavior.Inputs) > 0 {
					violations = append(violations,
						fmt.Sprintf("Card %s behavior[%d]: discount behavior has %d inputs (should have none)",
							card.ID, behaviorIdx, len(behavior.Inputs)))
				}

				// Discount behaviors should not have choices
				if len(behavior.Choices) > 0 {
					violations = append(violations,
						fmt.Sprintf("Card %s behavior[%d]: discount behavior has %d choices (should have none)",
							card.ID, behaviorIdx, len(behavior.Choices)))
				}

				// Discount behaviors should only have discount outputs
				for outputIdx, output := range behavior.Outputs {
					if output.Type != "discount" {
						violations = append(violations,
							fmt.Sprintf("Card %s behavior[%d] output[%d]: discount behavior has non-discount output type '%s'",
								card.ID, behaviorIdx, outputIdx, output.Type))
					}
				}
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("Found %d discount behavior validation violations:", len(violations))
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}
}
