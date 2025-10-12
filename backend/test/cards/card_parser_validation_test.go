package cards

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
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

// parseCardID parses a card ID and returns components for ordering
type CardIDComponents struct {
	IsNumeric bool
	Prefix    string
	Number    int
	Original  string
}

func parseCardID(id string) CardIDComponents {
	// Check if ID is purely numeric (like "001", "014", "999")
	if matched, _ := regexp.MatchString(`^\d+$`, id); matched {
		num, _ := strconv.Atoi(id)
		return CardIDComponents{
			IsNumeric: true,
			Prefix:    "",
			Number:    num,
			Original:  id,
		}
	}

	// Check if ID has letter prefix (like "P12", "C01", "A99")
	re := regexp.MustCompile(`^([A-Z]+)(\d+)$`)
	if matches := re.FindStringSubmatch(id); matches != nil {
		num, _ := strconv.Atoi(matches[2])
		return CardIDComponents{
			IsNumeric: false,
			Prefix:    matches[1],
			Number:    num,
			Original:  id,
		}
	}

	// Fallback for unexpected format
	return CardIDComponents{
		IsNumeric: false,
		Prefix:    id,
		Number:    0,
		Original:  id,
	}
}

// isCardIDOrderCorrect checks if card1 should come before card2 in the expected order
func isCardIDOrderCorrect(card1, card2 CardIDComponents) bool {
	// Rule: numeric IDs come first ("001" -> "999")
	if card1.IsNumeric && !card2.IsNumeric {
		return true // numeric before prefixed
	}
	if !card1.IsNumeric && card2.IsNumeric {
		return false // prefixed after numeric
	}

	if card1.IsNumeric && card2.IsNumeric {
		// Both numeric: order by number (strictly less, not less-or-equal)
		return card1.Number < card2.Number
	}

	// Both have prefixes: order by prefix first, then by number
	if card1.Prefix != card2.Prefix {
		return card1.Prefix < card2.Prefix // "A01" -> "A99" -> "B01" -> "Z99"
	}

	// Same prefix: order by number (strictly less)
	return card1.Number < card2.Number
}

// TestCardIDOrdering validates that card IDs are in the correct order
func TestCardIDOrdering(t *testing.T) {
	cards := loadCards(t)

	if len(cards) == 0 {
		t.Skip("No cards found to test ordering")
		return
	}

	var violations []string

	// Check if cards are in correct order
	for i := 0; i < len(cards)-1; i++ {
		current := parseCardID(cards[i].ID)
		next := parseCardID(cards[i+1].ID)

		// The current card should come before the next card
		if !isCardIDOrderCorrect(current, next) {
			violations = append(violations,
				fmt.Sprintf("Card order violation at index %d->%d: '%s' should come after '%s'",
					i, i+1, current.Original, next.Original))
		}

		// Also check for duplicates (same ID)
		if current.Original == next.Original {
			violations = append(violations,
				fmt.Sprintf("Duplicate card ID at index %d and %d: '%s'",
					i, i+1, current.Original))
		}
	}

	// Provide some examples of expected ordering in error message
	if len(violations) > 0 {
		t.Errorf("Found %d card ID ordering violations:", len(violations))
		t.Errorf("Expected order: numeric cards first (001->999), then prefixed cards (A01->A99->B01->Z99)")
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}

		// Show first few cards for debugging
		t.Errorf("First 10 card IDs in current order:")
		for i := 0; i < 10 && i < len(cards); i++ {
			components := parseCardID(cards[i].ID)
			t.Errorf("  [%d] %s (numeric: %v, prefix: '%s', number: %d)",
				i, cards[i].ID, components.IsNumeric, components.Prefix, components.Number)
		}

		// Also search for specific problematic cards mentioned by user
		t.Errorf("Searching for specific cards (025, 027, 028, 046, 049):")
		for i, card := range cards {
			if card.ID == "025" || card.ID == "027" || card.ID == "028" || card.ID == "046" || card.ID == "049" {
				components := parseCardID(card.ID)
				t.Errorf("  [%d] %s (numeric: %v, prefix: '%s', number: %d)",
					i, card.ID, components.IsNumeric, components.Prefix, components.Number)
			}
		}
	}
}

// TestCardSequenceIntegrity validates that numeric cards are in proper sequence without gaps
func TestCardSequenceIntegrity(t *testing.T) {
	cards := loadCards(t)

	if len(cards) == 0 {
		t.Skip("No cards found to test sequence")
		return
	}

	// Separate numeric and prefixed cards
	var numericCards []CardIDComponents
	var prefixedCards []CardIDComponents

	for _, card := range cards {
		components := parseCardID(card.ID)
		if components.IsNumeric {
			numericCards = append(numericCards, components)
		} else {
			prefixedCards = append(prefixedCards, components)
		}
	}

	var violations []string

	// Check numeric cards for sequence integrity
	if len(numericCards) > 0 {
		// Sort numeric cards by number to check sequence
		for i := 0; i < len(numericCards)-1; i++ {
			current := numericCards[i].Number
			next := numericCards[i+1].Number

			// Check if numbers are not consecutive (allowing for some gaps, but flagging major jumps)
			if next != current+1 && next > current+1 {
				// Found a gap - check if it's a legitimate skip or a misplacement
				violations = append(violations,
					fmt.Sprintf("Numeric sequence gap: card %03d followed by %03d (missing %03d)",
						current, next, current+1))
			}
		}

		// Also check the actual position of numeric cards in the full list
		t.Logf("Numeric cards found: %d", len(numericCards))
		t.Logf("First numeric card in sequence: %03d", numericCards[0].Number)
		t.Logf("Last numeric card in sequence: %03d", numericCards[len(numericCards)-1].Number)

		// Find where numeric cards actually appear in the main list
		firstNumericIndex := -1
		for i, card := range cards {
			components := parseCardID(card.ID)
			if components.IsNumeric {
				if firstNumericIndex == -1 {
					firstNumericIndex = i
				}
				break // We only need the first one
			}
		}

		if firstNumericIndex > 0 {
			violations = append(violations,
				fmt.Sprintf("Numeric cards should come first, but first numeric card is at index %d", firstNumericIndex))
		}
	}

	if len(violations) > 0 {
		t.Errorf("Found %d sequence integrity violations:", len(violations))
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}
}

// TestCardUniqueFields validates that card IDs and names are unique
func TestCardUniqueFields(t *testing.T) {
	cards := loadCards(t)

	if len(cards) == 0 {
		t.Skip("No cards found to test uniqueness")
		return
	}

	// Track seen IDs and names
	seenIDs := make(map[string][]int)   // ID -> list of indices where it appears
	seenNames := make(map[string][]int) // Name -> list of indices where it appears

	// Check for duplicate IDs and names
	for i, card := range cards {
		// Track IDs
		seenIDs[card.ID] = append(seenIDs[card.ID], i)

		// Track names
		seenNames[card.Name] = append(seenNames[card.Name], i)
	}

	var violations []string

	// Check for duplicate IDs
	for id, indices := range seenIDs {
		if len(indices) > 1 {
			violations = append(violations,
				fmt.Sprintf("Duplicate ID '%s' found at indices: %v", id, indices))
		}
	}

	// Check for duplicate names
	for name, indices := range seenNames {
		if len(indices) > 1 {
			// Get the IDs for better debugging
			var ids []string
			for _, idx := range indices {
				ids = append(ids, cards[idx].ID)
			}
			violations = append(violations,
				fmt.Sprintf("Duplicate name '%s' found at indices %v with IDs %v", name, indices, ids))
		}
	}

	// Report violations
	if len(violations) > 0 {
		t.Errorf("Found %d uniqueness violations:", len(violations))
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}

	// Also report statistics
	t.Logf("Checked %d cards:", len(cards))
	t.Logf("  - Unique IDs: %d", len(seenIDs))
	t.Logf("  - Unique names: %d", len(seenNames))
}

// TestSelfTargetingResourceOutputRequiresStorage validates that cards with self-targeting resource outputs have resourceStorage
func TestSelfTargetingResourceOutputRequiresStorage(t *testing.T) {
	cards := loadCards(t)

	var violations []string

	for _, card := range cards {
		for behaviorIdx, behavior := range card.Behaviors {
			// Check regular outputs
			for outputIdx, output := range behavior.Outputs {
				if requiresResourceStorage(output) && card.ResourceStorage == nil {
					violations = append(violations,
						fmt.Sprintf("Card %s (%s) behavior[%d] output[%d]: has self-targeting %s output but no resourceStorage field",
							card.ID, card.Name, behaviorIdx, outputIdx, output.Type))
				}
			}

			// Also check choices if they exist
			for choiceIdx, choice := range behavior.Choices {
				for outputIdx, output := range choice.Outputs {
					if requiresResourceStorage(output) && card.ResourceStorage == nil {
						violations = append(violations,
							fmt.Sprintf("Card %s (%s) behavior[%d] choice[%d] output[%d]: has self-targeting %s output but no resourceStorage field",
								card.ID, card.Name, behaviorIdx, choiceIdx, outputIdx, output.Type))
					}
				}
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("Found %d cards with self-targeting resource outputs lacking resourceStorage:", len(violations))
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}
}

// requiresResourceStorage checks if a resource output targets the card itself and needs storage
func requiresResourceStorage(output model.ResourceCondition) bool {
	// Check if this is a resource type that can be stored on cards
	resourceTypes := []model.ResourceType{
		model.ResourceMicrobes,
		model.ResourceAnimals,
		model.ResourceFloaters,
		model.ResourceScience,
		model.ResourceAsteroid,
		model.ResourceDisease,
	}

	// Check if the output type matches a storable resource
	for _, resType := range resourceTypes {
		if output.Type == resType && output.Target == model.TargetSelfCard {
			return true
		}
	}

	return false
}

// TestGlobalParameterRequirementLimits validates that card requirements respect global parameter limits
func TestGlobalParameterRequirementLimits(t *testing.T) {
	cards := loadCards(t)

	// Define global parameter limits (inclusive ranges)
	limits := map[model.RequirementType]struct {
		min int
		max int
	}{
		model.RequirementOxygen:      {min: 0, max: 14},
		model.RequirementTemperature: {min: -30, max: 8},
		model.RequirementOceans:      {min: 0, max: 9},
		model.RequirementVenus:       {min: 0, max: 30},
	}

	var violations []string

	for _, card := range cards {
		for reqIdx, requirement := range card.Requirements {
			// Only check requirements that have defined limits
			if limit, exists := limits[requirement.Type]; exists {
				// Check minimum requirement
				if requirement.Min != nil {
					minVal := *requirement.Min
					if minVal < limit.min || minVal > limit.max {
						violations = append(violations,
							fmt.Sprintf("Card %s (%s) requirement[%d]: %s min value %d is outside valid range [%d, %d]",
								card.ID, card.Name, reqIdx, requirement.Type, minVal, limit.min, limit.max))
					}
				}

				// Check maximum requirement
				if requirement.Max != nil {
					maxVal := *requirement.Max
					if maxVal < limit.min || maxVal > limit.max {
						violations = append(violations,
							fmt.Sprintf("Card %s (%s) requirement[%d]: %s max value %d is outside valid range [%d, %d]",
								card.ID, card.Name, reqIdx, requirement.Type, maxVal, limit.min, limit.max))
					}
				}
			}
		}
	}

	if len(violations) > 0 {
		t.Errorf("Found %d global parameter requirement limit violations:", len(violations))
		t.Errorf("Valid ranges: Oxygen [0-14], Temperature [-30 to 8], Oceans [0-9], Venus [0-30]")
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}
}
