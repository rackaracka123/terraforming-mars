package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"terraforming-mars-backend/internal/model"
)

const (
	colID           = 0
	colName         = 1
	colCost         = 2
	colSortOrder    = 3
	colDeck         = 4
	colType         = 5
	colExpansion    = 6
	colCategories   = 7
	colMax          = 8
	colRequireNum   = 9
	colRequireWhat  = 10
	colRequirements = 11

	colBuilding = 12
	colSpace    = 13
	colCity     = 14
	colPower    = 15
	colPlant    = 16
	colMicrobe  = 17
	colAnimal   = 18
	colScience  = 19
	colEarth    = 20
	colJovian   = 21
	colVenus    = 22
	colNone     = 23
	colWild     = 24

	colVPAmount              = 33
	colVPPer                 = 34
	behaviorID               = 0
	behaviorName             = 1
	behaviorRows             = 2
	behaviorType             = 3
	behaviorOption           = 4
	behaviorSort             = 5
	behaviorExp              = 6
	behaviorTrigger          = 7
	behaviorRestrictions     = 8
	behaviorN                = 9
	behaviorDetails          = 10
	behaviorMegacreditProd   = 11
	behaviorSteelProd        = 12
	behaviorTitaniumProd     = 13
	behaviorPlantProd        = 14
	behaviorEnergyProd       = 15
	behaviorHeatProd         = 16
	behaviorProduction       = 17
	behaviorMegacredits      = 18
	behaviorSteel            = 19
	behaviorTitanium         = 20
	behaviorPlants           = 21
	behaviorEnergy           = 22
	behaviorHeat             = 23
	behaviorInventory        = 24
	behaviorCards            = 25
	behaviorCardDetails      = 26
	behaviorCardResources    = 27
	behaviorResourceType     = 28
	behaviorWhere            = 29
	behaviorCardsCol         = 30
	behaviorCityTile         = 31
	behaviorSpecialTile      = 32
	behaviorGreeneryTile     = 33
	behaviorOceanTile        = 34
	behaviorTileRestrictions = 35
	behaviorTilesOnMars      = 36
	behaviorRaiseTemp        = 37
	behaviorRaiseOxygen      = 38
	behaviorRaiseVenus       = 39
	behaviorRaiseTR          = 40
	behaviorTotalTR          = 41
	behaviorCityNotMars      = 42
	behaviorColony           = 43
	behaviorTradeFleet       = 44
	behaviorDelegate         = 45
	behaviorCommunity        = 46
	behaviorAndAlso          = 47
	behaviorOther            = 48
	behaviorTextForm         = 49
	behaviorColor            = 50
)

// BehaviorData represents a behavior from behaviors.csv
type BehaviorData struct {
	ID           string
	Name         string
	BehaviorType string
	Option       string
	Trigger      string
	NEquals      string

	MegacreditProd    int
	MegacreditProdRaw string
	SteelProd         int
	SteelProdRaw      string
	TitaniumProd      int
	TitaniumProdRaw   string
	PlantProd         int
	PlantProdRaw      string
	EnergyProd        int
	EnergyProdRaw     string
	HeatProd          int
	HeatProdRaw       string

	Megacredits    int
	MegacreditsRaw string
	Steel          int
	SteelRaw       string
	Titanium       int
	TitaniumRaw    string
	Plants         int
	PlantsRaw      string
	Energy         int
	EnergyRaw      string
	Heat           int
	HeatRaw        string

	CityTile     int
	OceanTile    int
	GreeneryTile int

	RaiseTemp   int
	RaiseOxygen int
	RaiseVenus  int
	RaiseTR     int

	Cards         int
	CardResources string
	ResourceType  string
	Where         string

	TextForm string
}

// compareCardIDs compares two card IDs for sorting
// Rules: numeric cards (001-999) come first, then prefixed cards (A01-Z99) alphabetically
func compareCardIDs(id1, id2 string) bool {
	// Parse both IDs
	isNumeric1 := isNumericID(id1)
	isNumeric2 := isNumericID(id2)

	// Rule 1: Numeric cards come before prefixed cards
	if isNumeric1 && !isNumeric2 {
		return true
	}
	if !isNumeric1 && isNumeric2 {
		return false
	}

	if isNumeric1 && isNumeric2 {
		// Both numeric: compare as integers
		num1, _ := strconv.Atoi(id1)
		num2, _ := strconv.Atoi(id2)
		return num1 < num2
	}

	// Both prefixed: extract prefix and number for comparison
	prefix1, num1 := parseCardID(id1)
	prefix2, num2 := parseCardID(id2)

	// Compare by prefix first
	if prefix1 != prefix2 {
		return prefix1 < prefix2
	}

	// Same prefix: compare by number
	return num1 < num2
}

// isNumericID checks if a card ID is purely numeric
func isNumericID(id string) bool {
	_, err := strconv.Atoi(id)
	return err == nil
}

// parseCardID extracts prefix and number from a prefixed card ID (like "P12" -> ("P", 12))
func parseCardID(id string) (string, int) {
	if isNumericID(id) {
		num, _ := strconv.Atoi(id)
		return "", num
	}

	// Find where digits start
	i := 0
	for i < len(id) && (id[i] < '0' || id[i] > '9') {
		i++
	}

	if i == len(id) {
		// No digits found, treat whole string as prefix
		return id, 0
	}

	prefix := id[:i]
	numStr := id[i:]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return id, 0
	}

	return prefix, num
}

// sortBehaviors canonicalizes the order of behaviors for deterministic output.
// Sorts by: trigger type (alphabetically), then first input type, then first output type.
func sortBehaviors(behaviors []model.CardBehavior) {
	sort.Slice(behaviors, func(i, j int) bool {
		bI := behaviors[i]
		bJ := behaviors[j]

		// Compare by first trigger type (alphabetically)
		triggerI := getTriggerSortKey(bI)
		triggerJ := getTriggerSortKey(bJ)
		if triggerI != triggerJ {
			return triggerI < triggerJ
		}

		// Compare by first input type (alphabetically)
		inputI := getInputSortKey(bI)
		inputJ := getInputSortKey(bJ)
		if inputI != inputJ {
			return inputI < inputJ
		}

		// Compare by first output type (alphabetically)
		outputI := getOutputSortKey(bI)
		outputJ := getOutputSortKey(bJ)
		return outputI < outputJ
	})
}

// getTriggerSortKey returns the first trigger type as a string for sorting, or empty string if no triggers
func getTriggerSortKey(behavior model.CardBehavior) string {
	if len(behavior.Triggers) > 0 {
		return string(behavior.Triggers[0].Type)
	}
	return ""
}

// getInputSortKey returns the first input type as a string for sorting, or empty string if no inputs
func getInputSortKey(behavior model.CardBehavior) string {
	if len(behavior.Inputs) > 0 {
		return string(behavior.Inputs[0].Type)
	}
	return ""
}

// getOutputSortKey returns the first output type as a string for sorting, or empty string if no outputs
func getOutputSortKey(behavior model.CardBehavior) string {
	if len(behavior.Outputs) > 0 {
		return string(behavior.Outputs[0].Type)
	}
	return ""
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run parse_cards_enhanced.go <output.json>")
		os.Exit(1)
	}

	outputPath := os.Args[1]

	// Parse behaviors.csv first
	behaviorsMap, err := parseBehaviorsCSV()
	if err != nil {
		fmt.Printf("Error parsing behaviors.csv: %v\n", err)
		os.Exit(1)
	}

	// Parse cards.csv
	cards, err := parseCardsCSV(behaviorsMap)
	if err != nil {
		fmt.Printf("Error parsing cards.csv: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed %d cards with enhanced behavior data\n", len(cards))

	// Sort cards by ID for proper ordering (numeric first, then prefixed)
	sort.Slice(cards, func(i, j int) bool {
		return compareCardIDs(cards[i].ID, cards[j].ID)
	})
	fmt.Printf("Cards sorted by ID\n")

	output, err := json.MarshalIndent(cards, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(outputPath, output, 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Enhanced cards written to %s\n", outputPath)
}

func parseBehaviorsCSV() (map[string][]BehaviorData, error) {
	csvPath := filepath.Join("assets", "behaviors.csv")
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	behaviorsMap := make(map[string][]BehaviorData)
	lineNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading behaviors CSV line %d: %v\n", lineNum, err)
			continue
		}

		lineNum++

		// Skip header lines and empty/summary lines
		if lineNum <= 5 || len(record) < 10 || record[behaviorID] == "" ||
			strings.HasPrefix(record[behaviorID], "these totals") {
			continue
		}

		behavior := parseBehaviorFromRecord(record)
		if behavior != nil {
			behaviorsMap[behavior.ID] = append(behaviorsMap[behavior.ID], *behavior)
		}
	}

	fmt.Printf("Parsed behaviors for %d cards\n", len(behaviorsMap))
	return behaviorsMap, nil
}

func parseBehaviorFromRecord(record []string) *BehaviorData {
	if len(record) < 20 {
		return nil
	}

	behavior := &BehaviorData{
		ID:           strings.TrimSpace(record[behaviorID]),
		Name:         strings.TrimSpace(record[behaviorName]),
		BehaviorType: strings.TrimSpace(record[behaviorType]),
		Option:       strings.TrimSpace(record[behaviorOption]),
		Trigger:      strings.TrimSpace(record[behaviorTrigger]),
		NEquals:      strings.TrimSpace(record[behaviorN]),
	}

	// Parse production changes
	behavior.MegacreditProd = parseIntFromRecord(record, behaviorMegacreditProd)
	behavior.MegacreditProdRaw = strings.TrimSpace(record[behaviorMegacreditProd])
	behavior.SteelProd = parseIntFromRecord(record, behaviorSteelProd)
	behavior.SteelProdRaw = strings.TrimSpace(record[behaviorSteelProd])
	behavior.TitaniumProd = parseIntFromRecord(record, behaviorTitaniumProd)
	behavior.TitaniumProdRaw = strings.TrimSpace(record[behaviorTitaniumProd])
	behavior.PlantProd = parseIntFromRecord(record, behaviorPlantProd)
	behavior.PlantProdRaw = strings.TrimSpace(record[behaviorPlantProd])
	behavior.EnergyProd = parseIntFromRecord(record, behaviorEnergyProd)
	behavior.EnergyProdRaw = strings.TrimSpace(record[behaviorEnergyProd])
	behavior.HeatProd = parseIntFromRecord(record, behaviorHeatProd)
	behavior.HeatProdRaw = strings.TrimSpace(record[behaviorHeatProd])

	// Parse resource gains
	behavior.Megacredits = parseIntFromRecord(record, behaviorMegacredits)
	behavior.MegacreditsRaw = strings.TrimSpace(record[behaviorMegacredits])
	behavior.Steel = parseIntFromRecord(record, behaviorSteel)
	behavior.SteelRaw = strings.TrimSpace(record[behaviorSteel])
	behavior.Titanium = parseIntFromRecord(record, behaviorTitanium)
	behavior.TitaniumRaw = strings.TrimSpace(record[behaviorTitanium])
	behavior.Plants = parseIntFromRecord(record, behaviorPlants)
	behavior.PlantsRaw = strings.TrimSpace(record[behaviorPlants])
	behavior.Energy = parseIntFromRecord(record, behaviorEnergy)
	behavior.EnergyRaw = strings.TrimSpace(record[behaviorEnergy])
	behavior.Heat = parseIntFromRecord(record, behaviorHeat)
	behavior.HeatRaw = strings.TrimSpace(record[behaviorHeat])

	// Parse tile placements
	behavior.CityTile = parseIntFromRecord(record, behaviorCityTile)
	behavior.OceanTile = parseIntFromRecord(record, behaviorOceanTile)
	behavior.GreeneryTile = parseIntFromRecord(record, behaviorGreeneryTile)

	// Parse terraforming
	behavior.RaiseTemp = parseIntFromRecord(record, behaviorRaiseTemp)
	behavior.RaiseOxygen = parseIntFromRecord(record, behaviorRaiseOxygen)
	behavior.RaiseVenus = parseIntFromRecord(record, behaviorRaiseVenus)
	behavior.RaiseTR = parseIntFromRecord(record, behaviorRaiseTR)

	// Parse cards
	behavior.Cards = parseIntFromRecord(record, behaviorCards)

	// Parse text fields
	if len(record) > behaviorCardResources {
		behavior.CardResources = strings.TrimSpace(record[behaviorCardResources])
	}
	if len(record) > behaviorResourceType {
		behavior.ResourceType = strings.TrimSpace(record[behaviorResourceType])
	}
	if len(record) > behaviorWhere {
		behavior.Where = strings.TrimSpace(record[behaviorWhere])
	}
	if len(record) > behaviorTextForm {
		behavior.TextForm = strings.TrimSpace(record[behaviorTextForm])
	}

	return behavior
}

func parseIntFromRecord(record []string, colIndex int) int {
	if len(record) <= colIndex {
		return 0
	}

	value := strings.TrimSpace(record[colIndex])
	if value == "" || value == "0" {
		return 0
	}

	// Handle negative values
	if strings.HasPrefix(value, "-") {
		if num, err := strconv.Atoi(value); err == nil {
			return num
		}
	}

	// Handle positive values
	if num, err := strconv.Atoi(value); err == nil {
		return num
	}

	return 0
}

func parseCardsCSV(behaviorsMap map[string][]BehaviorData) ([]model.Card, error) {
	csvPath := filepath.Join("assets", "cards.csv")
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	var cards []model.Card
	lineNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading cards CSV line %d: %v\n", lineNum, err)
			continue
		}

		lineNum++

		// Skip header lines and empty/summary lines
		if lineNum <= 4 || len(record) < 10 || record[colID] == "" || strings.HasPrefix(record[colID], "these totals") {
			continue
		}

		card, err := parseCardFromRecord(record, behaviorsMap, lineNum)
		if err != nil {
			fmt.Printf("Error parsing card at line %d: %v\n", lineNum, err)
			continue
		}

		if card != nil {
			cards = append(cards, *card)
		}
	}

	return cards, nil
}

func parseCardFromRecord(record []string, behaviorsMap map[string][]BehaviorData, lineNum int) (*model.Card, error) {
	if len(record) < 20 {
		return nil, fmt.Errorf("insufficient columns")
	}

	card := &model.Card{}

	// Basic info
	card.ID = strings.TrimSpace(record[colID])
	card.Name = strings.TrimSpace(record[colName])

	// Parse cost
	if costStr := strings.TrimSpace(record[colCost]); costStr != "" {
		if cost, err := strconv.Atoi(costStr); err == nil {
			card.Cost = cost
		}
	}

	// Parse card type
	card.Type = parseCardType(record[colType])

	// Parse tags and sort them for consistent output
	tags := parseTags(record)
	if len(tags) > 0 {
		// Sort tags by string value for deterministic output
		sort.Slice(tags, func(i, j int) bool {
			return string(tags[i]) < string(tags[j])
		})
		card.Tags = tags
	}

	// Parse requirements
	requirements := parseRequirements(record)
	if requirements != nil {
		card.Requirements = requirements
	}

	// Parse basic victory points from cards.csv - create fixed VP condition if > 0
	// But only if there's no "per" condition that would make it a sophisticated VP condition
	if len(record) > 37 && record[37] != "" { // VP column in cards.csv
		// Check if there's a "per" condition (column 38) - if so, skip basic parsing
		hasPerCondition := len(record) > 38 && strings.TrimSpace(record[38]) != ""

		if !hasPerCondition {
			if vp, err := strconv.Atoi(strings.TrimSpace(record[37])); err == nil && vp > 0 {
				// Create a fixed VP condition instead of using the old VictoryPoints field
				fixedVP := model.VictoryPointCondition{
					Amount:    vp,
					Condition: model.VPConditionFixed,
					// For fixed VP, don't include MaxTrigger or Per
				}
				card.VPConditions = append(card.VPConditions, fixedVP)
			}
		}
	}

	// Parse description from cards.csv
	card.Description = parseDescription(record)

	// Parse resource storage from cards.csv "resource type held" column (column 36)
	if len(record) > 36 && strings.TrimSpace(record[36]) != "" {
		resourceTypeStr := strings.ToLower(strings.TrimSpace(record[36]))
		resourceStorage := createResourceStorageFromType(resourceTypeStr)
		if resourceStorage != nil {
			card.ResourceStorage = resourceStorage
		}
	}

	// Now enhance with behaviors.csv data
	if behaviors, exists := behaviorsMap[card.ID]; exists {
		enhanceCardWithBehaviors(card, behaviors, record)
	} else {
		// For cards without behaviors.csv data, parse effects from description
		parseEffectsFromDescription(card)
	}

	return card, nil
}

// ChoiceBehaviorData represents a grouped choice behavior
type ChoiceBehaviorData struct {
	BehaviorData
	IsChoice bool
	Choices  []BehaviorData
}

// processChoiceBehaviors groups behaviors with options (A, B, C) into choice behaviors
func processChoiceBehaviors(behaviors []BehaviorData) []ChoiceBehaviorData {
	// Group behaviors by BehaviorType
	behaviorGroups := make(map[string][]BehaviorData)

	for _, behavior := range behaviors {
		key := behavior.BehaviorType
		behaviorGroups[key] = append(behaviorGroups[key], behavior)
	}

	var result []ChoiceBehaviorData

	// Sort behavior group keys for deterministic output
	var groupKeys []string
	for key := range behaviorGroups {
		groupKeys = append(groupKeys, key)
	}
	sort.Strings(groupKeys)

	for _, key := range groupKeys {
		group := behaviorGroups[key]
		if hasChoiceOptions(group) {
			// Create a choice behavior
			base := group[0] // Use first as base
			choiceBehavior := ChoiceBehaviorData{
				BehaviorData: base,
				IsChoice:     true,
				Choices:      group,
			}
			result = append(result, choiceBehavior)
		} else {
			// For "0 Immediate" behaviors without choices, combine them into one behavior
			if len(group) > 1 && group[0].BehaviorType == "0 Immediate" {
				// Combine multiple immediate behaviors into one
				combinedBehavior := ChoiceBehaviorData{
					BehaviorData: group[0], // Use first as base
					IsChoice:     false,
					Choices:      group, // Store all behaviors for combining
				}
				result = append(result, combinedBehavior)
			} else {
				// Add all behaviors from this group as regular behaviors
				for _, behavior := range group {
					result = append(result, ChoiceBehaviorData{
						BehaviorData: behavior,
						IsChoice:     false,
					})
				}
			}
		}
	}

	return result
}

// parseWhereToAffectedTags converts Where field values to CardTag slice for location-based targeting.
//
// Converts location strings from behaviors.csv Where column to appropriate CardTag arrays
// for targeting cards with specific tags. Used when card effects should only apply to
// cards with certain location tags (e.g., "add 2 animals to any Venus card").
//
// Examples:
//
//	parseWhereToAffectedTags("Venus") -> []CardTag{TagVenus}
//	parseWhereToAffectedTags("jovian") -> []CardTag{TagJovian}
//	parseWhereToAffectedTags("any") -> nil
//	parseWhereToAffectedTags("") -> nil
func parseWhereToAffectedTags(where string) []model.CardTag {
	whereLower := strings.ToLower(strings.TrimSpace(where))
	switch whereLower {
	case "venus":
		return []model.CardTag{model.TagVenus}
	case "jovian":
		return []model.CardTag{model.TagJovian}
	case "earth":
		return []model.CardTag{model.TagEarth}
	default:
		return nil
	}
}

// hasChoiceOptions checks if a group of behaviors represents choice options (A, B, C)
func hasChoiceOptions(behaviors []BehaviorData) bool {
	if len(behaviors) <= 1 {
		return false
	}

	optionCount := 0
	for _, behavior := range behaviors {
		if behavior.Option != "" {
			optionCount++
		}
	}

	return optionCount > 1 && optionCount == len(behaviors)
}

// DEPRECATED: enhanceEffectTriggerActions - replaced by createTriggeredEffectAction

func enhanceCardWithBehaviors(card *model.Card, behaviors []BehaviorData, record []string) {
	var cardBehaviors []model.CardBehavior
	var resourceStorage *model.ResourceStorage
	var victoryConditions []model.VictoryPointCondition

	// Group behaviors by type and handle choices
	processedBehaviors := processChoiceBehaviors(behaviors)

	for _, behavior := range processedBehaviors {
		if behavior.IsChoice {
			// Handle choice behavior - check if it's a triggered choice
			if behavior.BehaviorType == "4 Effect/Trigger" {
				cardBehavior := createTriggeredChoiceBehavior(behavior, card)
				if cardBehavior != nil {
					cardBehaviors = append(cardBehaviors, *cardBehavior)
				}
			} else {
				// Handle immediate choice behavior
				cardBehavior := createChoiceBehavior(behavior, card)
				if cardBehavior != nil {
					cardBehaviors = append(cardBehaviors, *cardBehavior)
				}
			}
		} else {
			// Handle regular behavior
			switch behavior.BehaviorType {
			case "0 Immediate":
				// Handle both single and combined immediate behaviors
				if len(behavior.Choices) > 1 {
					// Multiple immediate behaviors to combine
					cardBehavior := createCombinedImmediateBehavior(behavior.Choices)
					if cardBehavior != nil {
						cardBehaviors = append(cardBehaviors, *cardBehavior)
					}
				} else {
					// Single immediate behavior
					cardBehavior := createImmediateBehavior(behavior.BehaviorData, false)
					if cardBehavior != nil {
						cardBehaviors = append(cardBehaviors, *cardBehavior)
					}
				}

			case "1 Immediate/Attack":
				// Create a new immediate behavior for attack effects
				cardBehavior := createAttackBehavior(behavior.BehaviorData)
				if cardBehavior != nil {
					cardBehaviors = append(cardBehaviors, *cardBehavior)
				}

			case "2 Action", "3 Action/First":
				cardBehavior := createActionBehavior(behavior.BehaviorData)
				if cardBehavior != nil {
					cardBehaviors = append(cardBehaviors, *cardBehavior)
				}

			case "4 Effect/Trigger":
				// Create a separate behavior for triggered effects
				cardBehavior := createTriggeredEffectBehavior(behavior.BehaviorData)
				if cardBehavior != nil {
					cardBehaviors = append(cardBehaviors, *cardBehavior)
				}

			case "5 Effect/Discount":
				// Create a separate behavior for discount effects
				cardBehavior := createDiscountBehavior(behavior.BehaviorData)
				if cardBehavior != nil {
					cardBehaviors = append(cardBehaviors, *cardBehavior)
				}

			case "6 Effect/Resource":
				// Check if this has a trigger condition - if so, treat as triggered effect
				if behavior.BehaviorData.Trigger != "" {
					cardBehavior := createTriggeredEffectBehavior(behavior.BehaviorData)
					if cardBehavior != nil {
						cardBehaviors = append(cardBehaviors, *cardBehavior)
					}
				} else {
					// Convert resource value modifiers to unified behaviors
					cardBehavior := createValueModifierBehavior(behavior.BehaviorData)
					if cardBehavior != nil {
						cardBehaviors = append(cardBehaviors, *cardBehavior)
					}
				}

			case "8 Effect/Ambient":
				// Check if this is a defense effect that should use unified system
				if defenseBehavior := createDefenseBehavior(behavior.BehaviorData); defenseBehavior != nil {
					cardBehaviors = append(cardBehaviors, *defenseBehavior)
				} else if isVenusLenienceEffect(behavior.BehaviorData) {
					// Check if this is a Venus lenience effect that should use unified system
					cardBehavior := createVenusLenienceBehavior(behavior.BehaviorData)
					if cardBehavior != nil {
						cardBehaviors = append(cardBehaviors, *cardBehavior)
					}
				} else if isGlobalParameterLenienceEffect(behavior.BehaviorData) {
					// Check if this is a global parameter lenience effect that should use unified system
					cardBehavior := createGlobalParameterLenienceBehavior(behavior.BehaviorData)
					if cardBehavior != nil {
						cardBehaviors = append(cardBehaviors, *cardBehavior)
					}
				} else {
					// Other ongoing effects - skip for now since we don't plan to support influence immediately
					// Log a message about unsupported effect type
					fmt.Printf("Skipping unsupported effect type for card %s: %s\n", behavior.ID, behavior.BehaviorType)
				}
			}

			// Resource storage is now parsed from cards.csv preserve column, not behaviors
		}
	}

	// Set the enhanced data on the card
	if len(cardBehaviors) > 0 {
		// Sort behaviors for deterministic output
		sortBehaviors(cardBehaviors)
		card.Behaviors = cardBehaviors
	}
	if resourceStorage != nil {
		card.ResourceStorage = resourceStorage
	}
	if len(victoryConditions) > 0 {
		card.VPConditions = victoryConditions
	}

	// Parse VP conditions from cards.csv first (most authoritative)
	vpConditions := parseVictoryConditionsFromCardsCSV(record)

	// DISABLED: behavior-based VP parsing can create incorrect VP conditions
	// VP conditions should only come from explicit VP columns in cards.csv
	// Cards with CardResources for immediate effects should not get VP conditions
	// if len(vpConditions) == 0 && len(card.VPConditions) == 0 {
	//     vpConditions = parseVictoryConditionsFromBehaviors(behaviors, record)
	// }

	// Note: Removed risky text parsing fallback - all VP data should come from CSV columns

	// Only set VP conditions if we found some and don't already have them
	if len(vpConditions) > 0 && len(card.VPConditions) == 0 {
		card.VPConditions = vpConditions
	}
}

// DEPRECATED: enhanceImmediateActions - replaced by createImmediateAction

// DEPRECATED: enhanceAttackActions - replaced by createAttackAction

// parseTriggerCondition parses trigger conditions from behaviors.csv "Where" column
func parseTriggerCondition(whereText string) *model.ResourceTriggerCondition {
	if whereText == "" {
		return nil
	}

	whereText = strings.ToLower(strings.TrimSpace(whereText))

	switch {
	case strings.Contains(whereText, "city"):
		triggerType := model.TriggerCityPlaced
		location := model.CardApplyLocationAnywhere
		if strings.Contains(whereText, "mars") {
			location = model.CardApplyLocationMars
		}
		return &model.ResourceTriggerCondition{
			Type:     triggerType,
			Location: &location,
		}
	case strings.Contains(whereText, "ocean"):
		triggerType := model.TriggerOceanPlaced
		location := model.CardApplyLocationAnywhere
		if strings.Contains(whereText, "mars") {
			location = model.CardApplyLocationMars
		}
		return &model.ResourceTriggerCondition{
			Type:     triggerType,
			Location: &location,
		}
	case strings.Contains(whereText, "greenery"):
		triggerType := model.TriggerTilePlaced
		location := model.CardApplyLocationAnywhere
		if strings.Contains(whereText, "mars") {
			location = model.CardApplyLocationMars
		}
		return &model.ResourceTriggerCondition{
			Type:     triggerType,
			Location: &location,
		}
	case strings.Contains(whereText, "temperature"):
		return &model.ResourceTriggerCondition{
			Type: model.TriggerTemperatureRaise,
		}
	case strings.Contains(whereText, "oxygen"):
		return &model.ResourceTriggerCondition{
			Type: model.TriggerOxygenRaise,
		}
	}

	return nil
}

func createResourceExchange(behavior BehaviorData, isAttack bool, triggerType *model.ResourceTriggerType, triggerCondition *model.ResourceTriggerCondition, isImmediateEffect bool) *model.ResourceExchange {
	// Create triggers list - will be added at the end
	var triggers []model.Trigger
	// Helper function to determine target based on resource type and context
	getTarget := func(resourceType model.ResourceType, isInput bool) model.TargetType {
		if isAttack {
			return model.TargetAnyPlayer // Attack actions target other players
		}

		// Determine target based on resource type
		switch resourceType {
		case model.ResourceCityPlacement, model.ResourceOceanPlacement, model.ResourceGreeneryPlacement,
			model.ResourceTemperature, model.ResourceOxygen, model.ResourceVenus, model.ResourceTR:
			return model.TargetNone // Board actions don't target players
		case model.ResourceCreditsProduction, model.ResourceSteelProduction, model.ResourceTitaniumProduction,
			model.ResourcePlantsProduction, model.ResourceEnergyProduction, model.ResourceHeatProduction:
			return model.TargetSelfPlayer // Production changes target the player
		case model.ResourceGlobalParameterLenience:
			return model.TargetSelfPlayer // Global parameter lenience applies to the player
		case model.ResourceVenusLenience:
			return model.TargetSelfPlayer // Venus lenience applies to the player
		case model.ResourceDefense:
			return model.TargetSelfCard // Defense effects apply to the specific card
		default:
			return model.TargetSelfPlayer // Resource/production/card actions target the player
		}
	}

	// Check for conditional/variable outputs first (like Martian Rails with "N" megacredits)
	conditionalOutput := parseConditionalOutput(behavior)
	hasConditionalOutput := conditionalOutput != nil

	// Resource exchange (unified system: resources, cards, terraforming, global parameters, attacks)
	if behavior.Megacredits != 0 || behavior.Steel != 0 || behavior.Titanium != 0 ||
		behavior.Plants != 0 || behavior.Energy != 0 || behavior.Heat != 0 ||
		behavior.MegacreditProd != 0 || behavior.SteelProd != 0 || behavior.TitaniumProd != 0 ||
		behavior.PlantProd != 0 || behavior.EnergyProd != 0 || behavior.HeatProd != 0 ||
		behavior.Cards > 0 || strings.Contains(behavior.TextForm, " of ") ||
		hasConditionalOutput || // Include conditional outputs
		behavior.CityTile != 0 || behavior.OceanTile != 0 || behavior.GreeneryTile != 0 || // Terraforming actions
		behavior.RaiseTemp != 0 || behavior.RaiseOxygen != 0 || behavior.RaiseTR != 0 || // Global parameters
		behavior.CardResources != "" { // Card resources (like floaters, microbes, animals)

		var inputs []model.ResourceCondition
		var outputs []model.ResourceCondition

		// Add conditional output if we found one
		if conditionalOutput != nil {
			outputs = append(outputs, *conditionalOutput)
		}

		// Convert resource values to ResourceCondition objects
		resourceInputs := []struct {
			value        int
			resourceType model.ResourceType
		}{
			{behavior.Megacredits, model.ResourceCredits},
			{behavior.Steel, model.ResourceSteel},
			{behavior.Titanium, model.ResourceTitanium},
			{behavior.Plants, model.ResourcePlants},
			{behavior.Energy, model.ResourceEnergy},
			{behavior.Heat, model.ResourceHeat},
		}

		for _, res := range resourceInputs {
			if res.value > 0 {
				// Fixed output (not conditional)
				condition := model.ResourceCondition{
					Type:   res.resourceType,
					Amount: res.value,
					Target: getTarget(res.resourceType, false),
				}
				outputs = append(outputs, condition)
			} else if res.value < 0 {
				// Handle negative values: for immediate effects, treat as negative outputs
				// For non-immediate effects, treat as positive inputs (cost)
				if isImmediateEffect {
					// Immediate effects: negative values are negative outputs (like "lose 3 M€")
					outputs = append(outputs, model.ResourceCondition{
						Type:   res.resourceType,
						Amount: res.value, // Keep negative amount
						Target: getTarget(res.resourceType, false),
					})
				} else {
					// Non-immediate effects: negative values are positive inputs (cost)
					inputs = append(inputs, model.ResourceCondition{
						Type:   res.resourceType,
						Amount: -res.value, // Convert to positive
						Target: getTarget(res.resourceType, true),
					})
				}
			}
		}

		// Handle production resources (for attacks and regular production changes)
		productionInputs := []struct {
			value        int
			resourceType model.ResourceType
		}{
			{behavior.MegacreditProd, model.ResourceCreditsProduction},
			{behavior.SteelProd, model.ResourceSteelProduction},
			{behavior.TitaniumProd, model.ResourceTitaniumProduction},
			{behavior.PlantProd, model.ResourcePlantsProduction},
			{behavior.EnergyProd, model.ResourceEnergyProduction},
			{behavior.HeatProd, model.ResourceHeatProduction},
		}

		for _, prod := range productionInputs {
			if prod.value != 0 {
				// Production changes (both positive and negative) are outputs
				amount := prod.value
				target := getTarget(prod.resourceType, false)

				// For attacks: positive values should be negative since attacks decrease opponent's production
				if isAttack && prod.value > 0 {
					amount = -prod.value
				}

				outputs = append(outputs, model.ResourceCondition{
					Type:   prod.resourceType,
					Amount: amount,
					Target: target,
				})
			}
		}

		// Handle tile placements
		if behavior.CityTile > 0 {
			outputs = append(outputs, model.ResourceCondition{
				Type:   model.ResourceCityPlacement,
				Amount: behavior.CityTile,
				Target: model.TargetNone,
			})
		}
		if behavior.OceanTile > 0 {
			outputs = append(outputs, model.ResourceCondition{
				Type:   model.ResourceOceanPlacement,
				Amount: behavior.OceanTile,
				Target: model.TargetNone,
			})
		}
		if behavior.GreeneryTile > 0 {
			outputs = append(outputs, model.ResourceCondition{
				Type:   model.ResourceGreeneryPlacement,
				Amount: behavior.GreeneryTile,
				Target: model.TargetNone,
			})
		}

		// Global parameters are handled in the globalParams loop below to avoid duplicates

		// Handle card operations
		// Check for "X of Y" pattern in TextForm (e.g., "Draw 2 of 4 cards")
		if strings.Contains(behavior.TextForm, " of ") {
			// Extract numbers from patterns like "Draw 2 of 4 cards"
			re := regexp.MustCompile(`(\d+)\s+of\s+(\d+)`)
			matches := re.FindStringSubmatch(behavior.TextForm)
			if len(matches) == 3 {
				if take, takeErr := strconv.Atoi(matches[1]); takeErr == nil && take > 0 {
					if peek, peekErr := strconv.Atoi(matches[2]); peekErr == nil && peek > 0 {
						if take == peek {
							// Same amount for take and peek = simple draw
							outputs = append(outputs, model.ResourceCondition{
								Type:   model.ResourceCardDraw,
								Amount: take,

								Target: getTarget(model.ResourceCardDraw, false),
							})
						} else {
							// Different amounts - use separate take and peek
							outputs = append(outputs, model.ResourceCondition{
								Type:   model.ResourceCardTake,
								Amount: take,

								Target: getTarget(model.ResourceCardTake, false),
							})
							outputs = append(outputs, model.ResourceCondition{
								Type:   model.ResourceCardPeek,
								Amount: peek,

								Target: getTarget(model.ResourceCardPeek, false),
							})
						}
					}
				}
			}
		} else if behavior.Cards > 0 {
			// Simple card draw without "X of Y" pattern
			outputs = append(outputs, model.ResourceCondition{
				Type:   model.ResourceCardDraw,
				Amount: behavior.Cards,

				Target: getTarget(model.ResourceCardDraw, false),
			})
		}

		if behavior.CardResources != "" && behavior.ResourceType != "" {
			resourceType := parseResourceType(behavior.ResourceType)
			target := model.TargetAny
			if behavior.Where == "self" || behavior.Where == "here" {
				target = model.TargetSelfCard
			} else if behavior.Where == "any" {
				target = model.TargetAny
			}

			cardResourceAmount := parseResourceAmount(behavior.CardResources)
			affectedTags := parseWhereToAffectedTags(behavior.Where)
			outputs = append(outputs, model.ResourceCondition{
				Type:         resourceType,
				Amount:       cardResourceAmount,
				Target:       target,
				AffectedTags: affectedTags,
			})
		}
		globalParams := []struct {
			value        int
			resourceType model.ResourceType
		}{
			{behavior.RaiseTemp, model.ResourceTemperature},
			{behavior.RaiseOxygen, model.ResourceOxygen},
			{behavior.RaiseTR, model.ResourceTR},
		}

		for _, param := range globalParams {
			if param.value > 0 {
				outputs = append(outputs, model.ResourceCondition{
					Type:   param.resourceType,
					Amount: param.value,

					Target: getTarget(param.resourceType, false),
				})
			}
		}

		// Create trigger if specified
		if triggerType != nil {
			trigger := model.Trigger{
				Type: *triggerType,
			}
			if triggerCondition != nil {
				trigger.Condition = triggerCondition
			}
			triggers = append(triggers, trigger)
		}

		// Return ResourceExchange if we have inputs or outputs
		if len(inputs) > 0 || len(outputs) > 0 {
			exchange := &model.ResourceExchange{
				Triggers: triggers,
				Inputs:   inputs,
				Outputs:  outputs,
			}
			return exchange
		}
	}

	return nil
}

func createActionBehavior(behavior BehaviorData) *model.CardBehavior {
	// Use the shared ResourceExchange creation logic with manual trigger (actions)
	manualTrigger := model.ResourceTriggerManual
	resourceExchange := createResourceExchange(behavior, false, &manualTrigger, nil, false)

	if resourceExchange != nil {
		return &model.CardBehavior{
			Triggers: resourceExchange.Triggers,
			Inputs:   resourceExchange.Inputs,
			Outputs:  resourceExchange.Outputs,
		}
	}

	// If no action content, return nil so it's not added to the card
	return nil
}

// DEPRECATED: enhanceDiscountActions - replaced by createDiscountAction

// parseNBasedConditionalOutput parses conditional outputs using "N" mechanism in resource columns
func parseNBasedConditionalOutput(behavior BehaviorData) *model.ResourceCondition {
	// Check if NEquals is empty - no conditional output
	if behavior.NEquals == "" {
		return nil
	}

	// Find which resource column has "N" value
	var resourceType model.ResourceType

	// Check each resource column for "N" value
	if behavior.MegacreditsRaw == "N" {
		resourceType = model.ResourceCredits
	} else if behavior.PlantsRaw == "N" {
		resourceType = model.ResourcePlants
	} else if behavior.EnergyRaw == "N" {
		resourceType = model.ResourceEnergy
	} else if behavior.SteelRaw == "N" {
		resourceType = model.ResourceSteel
	} else if behavior.TitaniumRaw == "N" {
		resourceType = model.ResourceTitanium
	} else if behavior.HeatRaw == "N" {
		resourceType = model.ResourceHeat
	} else if behavior.MegacreditProdRaw == "N" {
		resourceType = model.ResourceCreditsProduction
	} else if behavior.SteelProdRaw == "N" {
		resourceType = model.ResourceSteelProduction
	} else if behavior.TitaniumProdRaw == "N" {
		resourceType = model.ResourceTitaniumProduction
	} else if behavior.PlantProdRaw == "N" {
		resourceType = model.ResourcePlantsProduction
	} else if behavior.EnergyProdRaw == "N" {
		resourceType = model.ResourceEnergyProduction
	} else if behavior.HeatProdRaw == "N" {
		resourceType = model.ResourceHeatProduction
	} else {
		// No "N" found in resource or production columns
		return nil
	}

	// Parse the "N equals" column to understand what N represents
	nEqualsLower := strings.ToLower(behavior.NEquals)

	// Parse the per condition type and location
	var perType model.ResourceType
	var location *model.CardApplyLocation
	var target *model.TargetType // For specifying whose tags/resources to count
	var perAmount int = 1        // Default: per 1 of something
	var tagType *model.CardTag   // For tag-based conditions
	var maxTrigger *int          // For MAX limitations like "MAX 4"

	// Parse different patterns in "N equals"
	switch {
	case strings.Contains(nEqualsLower, "any cities") || strings.Contains(nEqualsLower, "any city"):
		perType = model.ResourceCityTile
		locationVal := model.CardApplyLocationAnywhere
		location = &locationVal

	case strings.Contains(nEqualsLower, "city on mars") || strings.Contains(nEqualsLower, "cities on mars"):
		perType = model.ResourceCityTile
		locationVal := model.CardApplyLocationMars
		location = &locationVal

	case strings.Contains(nEqualsLower, "ocean"):
		perType = model.ResourceOceanTile
		locationVal := model.CardApplyLocationAnywhere
		location = &locationVal

	case strings.Contains(nEqualsLower, "greenery"):
		perType = model.ResourceGreeneryTile
		locationVal := model.CardApplyLocationAnywhere
		location = &locationVal

	case strings.Contains(nEqualsLower, "tag"):
		// Handle tag-based conditions like "Earth tags"
		perType = model.ResourceTag
		perAmount = 1 // 1 resource per 1 tag
		// Location for tags is always "anywhere" (tags don't have physical board location)
		locationVal := model.CardApplyLocationAnywhere
		location = &locationVal
		// Check if it should target all players (ANY tags) or just self-player (own tags)
		if strings.Contains(nEqualsLower, "any ") {
			// "any science tags" = count all players' tags
			targetVal := model.TargetAny
			target = &targetVal
		} else {
			// "science tags" (without "any") = count only self-player's tags
			targetVal := model.TargetSelfPlayer
			target = &targetVal
		}

		// Parse the specific tag type
		switch {
		case strings.Contains(nEqualsLower, "earth"):
			tag := model.TagEarth
			tagType = &tag
		case strings.Contains(nEqualsLower, "jovian"):
			tag := model.TagJovian
			tagType = &tag
		case strings.Contains(nEqualsLower, "venus"):
			tag := model.TagVenus
			tagType = &tag
		case strings.Contains(nEqualsLower, "science"):
			tag := model.TagScience
			tagType = &tag
		case strings.Contains(nEqualsLower, "space"):
			tag := model.TagSpace
			tagType = &tag
		case strings.Contains(nEqualsLower, "building"):
			tag := model.TagBuilding
			tagType = &tag
		case strings.Contains(nEqualsLower, "power"):
			tag := model.TagPower
			tagType = &tag
		case strings.Contains(nEqualsLower, "plant"):
			tag := model.TagPlant
			tagType = &tag
		case strings.Contains(nEqualsLower, "microbe"):
			tag := model.TagMicrobe
			tagType = &tag
		case strings.Contains(nEqualsLower, "animal"):
			tag := model.TagAnimal
			tagType = &tag
		case strings.Contains(nEqualsLower, "city"):
			tag := model.TagCity
			tagType = &tag
		// Add default case for unknown tag types
		default:
			// Unknown tag type, return nil
			return nil
		}

	case strings.Contains(nEqualsLower, "floaters here"):
		perType = model.ResourceFloaters
		perAmount = 1
		targetVal := model.TargetSelfCard
		target = &targetVal
		locationVal := model.CardApplyLocationAnywhere
		location = &locationVal

	default:
		// Unknown per condition
		return nil
	}

	// Parse MAX limitations like "MAX 4"
	if strings.Contains(nEqualsLower, "max ") {
		// Extract the number after "max "
		maxRegex := regexp.MustCompile(`max\s+(\d+)`)
		if matches := maxRegex.FindStringSubmatch(nEqualsLower); len(matches) > 1 {
			if maxVal, err := strconv.Atoi(matches[1]); err == nil {
				maxTrigger = &maxVal
			}
		}
	}

	// Create the conditional output
	return &model.ResourceCondition{
		Type:       resourceType,
		Amount:     1, // Amount gained per condition (N means 1 per)
		Target:     model.TargetSelfPlayer,
		MaxTrigger: maxTrigger,
		Per: &model.PerCondition{
			Type:     perType,
			Amount:   perAmount,
			Location: location,
			Target:   target,
			Tag:      tagType,
		},
	}
}

// parseConditionalOutput parses conditional resource gains like "1 M€ per city on Mars"
func parseConditionalOutput(behavior BehaviorData) *model.ResourceCondition {
	// First check for "N" mechanism in resource columns
	nCondition := parseNBasedConditionalOutput(behavior)
	if nCondition != nil {
		return nCondition
	}

	// Fallback to text-based parsing for backwards compatibility
	textLower := strings.ToLower(behavior.TextForm)
	if !strings.Contains(textLower, "per") || behavior.Where == "" {
		return nil
	}

	// Parse the resource type from TextForm (e.g., "1 M€ per city" -> credits)
	var resourceType model.ResourceType
	var amount int = 1 // Default amount gained per condition

	switch {
	case strings.Contains(textLower, "m€") || strings.Contains(textLower, "credit") || strings.Contains(textLower, "M€"):
		resourceType = model.ResourceCredits
	case strings.Contains(textLower, "energy"):
		resourceType = model.ResourceEnergy
	case strings.Contains(textLower, "plant"):
		resourceType = model.ResourcePlants
	case strings.Contains(textLower, "steel"):
		resourceType = model.ResourceSteel
	case strings.Contains(textLower, "titanium"):
		resourceType = model.ResourceTitanium
	case strings.Contains(textLower, "heat"):
		resourceType = model.ResourceHeat
	default:
		return nil // Unknown resource type
	}

	// Parse the "per" condition from Where field and TextForm
	whereLower := strings.ToLower(behavior.Where)
	var perType model.ResourceType
	var location *model.CardApplyLocation

	switch {
	case strings.Contains(whereLower, "city"):
		perType = model.ResourceCityTile
	case strings.Contains(whereLower, "ocean"):
		perType = model.ResourceOceanTile
	case strings.Contains(whereLower, "greenery"):
		perType = model.ResourceGreeneryTile
	default:
		return nil // Unknown per condition
	}

	// Parse location constraint
	if strings.Contains(whereLower, "mars") {
		locationVal := model.CardApplyLocationMars
		location = &locationVal
	} else {
		locationVal := model.CardApplyLocationAnywhere
		location = &locationVal
	}

	// Create the conditional output with per condition
	return &model.ResourceCondition{
		Type:   resourceType,
		Amount: amount,
		Target: model.TargetSelfPlayer,
		Per: &model.PerCondition{
			Type:     perType,
			Amount:   1, // 1 means "per 1 city" etc.
			Location: location,
		},
	}
}

func createResourceStorage(behavior BehaviorData) *model.ResourceStorage {
	// DEPRECATED: Resource storage should be parsed from cards.csv preserve column, not behaviors
	// This function is kept for backward compatibility but should not be used
	return nil
}

// createResourceStorageFromType creates ResourceStorage from "resource type held" CSV values.
//
// Converts resource type strings from cards.csv "resource type held" column to
// ResourceStorage objects that define what types of resources a card can hold.
//
// Examples:
//
//	createResourceStorageFromType("animal") -> &ResourceStorage{Type: ResourceAnimals, Starting: 0}
//	createResourceStorageFromType("floater") -> &ResourceStorage{Type: ResourceFloaters, Starting: 0}
//	createResourceStorageFromType("unknown") -> nil
func createResourceStorageFromType(resourceTypeStr string) *model.ResourceStorage {
	var resourceType model.ResourceType

	switch resourceTypeStr {
	case "microbe":
		resourceType = model.ResourceMicrobes
	case "animal":
		resourceType = model.ResourceAnimals
	case "floater":
		resourceType = model.ResourceFloaters
	case "science":
		resourceType = model.ResourceScience
	case "asteroid":
		resourceType = model.ResourceAsteroid
	case "disease":
		resourceType = model.ResourceDisease
	default:
		return nil
	}

	return &model.ResourceStorage{
		Type:     resourceType,
		Starting: 0,
	}
}

// createCombinedImmediateBehavior creates a single immediate behavior from multiple immediate behaviors
func createCombinedImmediateBehavior(behaviors []BehaviorData) *model.CardBehavior {
	var allOutputs []model.ResourceCondition

	// Extract outputs from all behaviors and combine them
	for _, behavior := range behaviors {
		// Use the same logic as createImmediateBehavior, but just extract outputs
		autoTrigger := model.ResourceTriggerAuto
		resourceExchange := createResourceExchange(behavior, false, &autoTrigger, nil, true)
		if resourceExchange != nil {
			allOutputs = append(allOutputs, resourceExchange.Outputs...)
		}
	}

	if len(allOutputs) == 0 {
		return nil
	}

	// Create auto trigger
	autoTrigger := model.ResourceTriggerAuto
	trigger := model.Trigger{
		Type: autoTrigger,
	}

	return &model.CardBehavior{
		Triggers: []model.Trigger{trigger},
		Inputs:   []model.ResourceCondition{},
		Outputs:  allOutputs,
	}
}

// isVenusLenienceEffect checks if a behavior represents a Venus lenience effect
func isVenusLenienceEffect(behavior BehaviorData) bool {
	text := strings.ToLower(behavior.TextForm)
	return strings.Contains(text, "weaken venus reqt") ||
		strings.Contains(text, "venus requirement") ||
		strings.Contains(text, "venus terraforming requirement")
}

// createVenusLenienceAction creates an immediate action for Venus parameter lenience
func createVenusLenienceBehavior(behavior BehaviorData) *model.CardBehavior {
	// Parse the lenience amount from the text
	lenienceAmount := 2 // Default for Morning Star Inc
	text := strings.ToLower(behavior.TextForm)

	// Try to extract specific lenience amount from text
	if strings.Contains(text, "2 steps") {
		lenienceAmount = 2
	}

	// Create ResourceExchange with output "venus-lenience"
	autoTrigger := model.ResourceTriggerAuto

	// Create trigger for immediate auto effect
	trigger := model.Trigger{
		Type: autoTrigger,
		// No trigger condition needed for immediate effects
	}

	// Output: venus lenience effect
	venusLenienceOutput := model.ResourceCondition{
		Type:   model.ResourceVenusLenience,
		Amount: lenienceAmount, // +/- steps for Venus requirements

		Target: model.TargetSelfPlayer,
	}

	exchange := &model.ResourceExchange{
		Triggers: []model.Trigger{trigger},
		Outputs:  []model.ResourceCondition{venusLenienceOutput},
	}

	return &model.CardBehavior{
		Triggers: exchange.Triggers,
		Inputs:   exchange.Inputs,
		Outputs:  exchange.Outputs,
	}
}

func parseTriggerType(behaviorType, textForm string) *model.TriggerType {
	text := strings.ToLower(textForm)

	if strings.Contains(text, "city") && strings.Contains(text, "placed") {
		trigger := model.TriggerCityPlaced
		return &trigger
	}
	if strings.Contains(text, "ocean") && strings.Contains(text, "placed") {
		trigger := model.TriggerOceanPlaced
		return &trigger
	}
	if strings.Contains(text, "temperature") && strings.Contains(text, "raised") {
		trigger := model.TriggerTemperatureRaise
		return &trigger
	}
	if strings.Contains(text, "oxygen") && strings.Contains(text, "raised") {
		trigger := model.TriggerOxygenRaise
		return &trigger
	}
	if strings.Contains(text, "card") && strings.Contains(text, "played") {
		trigger := model.TriggerCardPlayed
		return &trigger
	}

	return nil
}

// createDefenseAction creates a defense immediate action for protection effects
func createDefenseBehavior(behavior BehaviorData) *model.CardBehavior {
	// Check for defense text patterns in behavior TextForm
	textLower := strings.ToLower(behavior.TextForm)

	// Defense patterns: "opponents can't remove P/M/A" or "animals may not be removed from this card"
	if strings.Contains(textLower, "opponents can't remove") ||
		strings.Contains(textLower, "opponents may not remove") ||
		strings.Contains(textLower, "may not be removed from this card") {

		var protectedTypes []string

		// Check for specific resource types in the behavior text
		// Protected Habitats uses "P/M/A" abbreviation
		if strings.Contains(textLower, "p/m/a") {
			protectedTypes = append(protectedTypes, "plants", "microbes", "animals")
		} else if strings.Contains(textLower, "animals may not be removed") {
			// Pets pattern: "animals may not be removed from this card"
			protectedTypes = append(protectedTypes, "animals")
		} else {
			// Individual checks for other potential cards
			if strings.Contains(textLower, "p") {
				protectedTypes = append(protectedTypes, "plants")
			}
			if strings.Contains(textLower, "m") {
				protectedTypes = append(protectedTypes, "microbes")
			}
			if strings.Contains(textLower, "a") {
				protectedTypes = append(protectedTypes, "animals")
			}
		}

		if len(protectedTypes) > 0 {
			defenseCondition := model.ResourceCondition{
				Type:              model.ResourceDefense,
				Amount:            1, // Defense is active (binary)
				Target:            model.TargetSelfCard,
				AffectedResources: protectedTypes,
			}

			// Create trigger for immediate auto effect
			autoTrigger := model.ResourceTriggerAuto
			trigger := model.Trigger{
				Type: autoTrigger,
				// No trigger condition needed for immediate defense effects
			}

			exchange := &model.ResourceExchange{
				Triggers: []model.Trigger{trigger},
				Outputs:  []model.ResourceCondition{defenseCondition},
			}

			return &model.CardBehavior{
				Triggers: exchange.Triggers,
				Inputs:   exchange.Inputs,
				Outputs:  exchange.Outputs,
			}
		}
	}

	return nil
}

// Reuse existing functions from the original parser
func parseCardType(typeStr string) model.CardType {
	typeStr = strings.ToLower(strings.TrimSpace(typeStr))
	switch {
	case strings.Contains(typeStr, "corp"):
		return model.CardTypeCorporation
	case strings.Contains(typeStr, "prelude"):
		return model.CardTypePrelude
	case strings.Contains(typeStr, "event"):
		return model.CardTypeEvent
	case strings.Contains(typeStr, "active"):
		return model.CardTypeActive
	default:
		return model.CardTypeAutomated
	}
}

func parseTags(record []string) []model.CardTag {
	var tags []model.CardTag

	tagMap := map[int]model.CardTag{
		colBuilding: model.TagBuilding,
		colSpace:    model.TagSpace,
		colCity:     model.TagCity,
		colPower:    model.TagPower,
		colPlant:    model.TagPlant,
		colMicrobe:  model.TagMicrobe,
		colAnimal:   model.TagAnimal,
		colScience:  model.TagScience,
		colEarth:    model.TagEarth,
		colJovian:   model.TagJovian,
		colVenus:    model.TagVenus,
		colWild:     model.TagWild,
	}

	// Sort tag map keys for deterministic output
	var tagCols []int
	for col := range tagMap {
		tagCols = append(tagCols, col)
	}
	sort.Ints(tagCols)

	for _, col := range tagCols {
		tag := tagMap[col]
		if len(record) > col && record[col] != "" && record[col] != "0" {
			count := 1
			if num, err := strconv.Atoi(strings.TrimSpace(record[col])); err == nil && num > 1 {
				count = num
			}
			for i := 0; i < count; i++ {
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

func parseRequirements(record []string) []model.Requirement {
	if len(record) <= colRequireWhat {
		return nil
	}

	requireWhat := strings.ToLower(strings.TrimSpace(record[colRequireWhat]))
	requireNumStr := strings.TrimSpace(record[colRequireNum])

	if requireNumStr == "" {
		return nil
	}

	requireNum, err := strconv.Atoi(requireNumStr)
	if err != nil {
		return nil
	}

	var requirements []model.Requirement
	isMax := strings.Contains(record[colMax], "max")

	switch {
	case strings.Contains(requireWhat, "ocean"):
		requirement := model.Requirement{
			Type: model.RequirementOceans,
		}
		if isMax {
			requirement.Max = &requireNum
		} else {
			requirement.Min = &requireNum
		}
		requirements = append(requirements, requirement)
	case strings.Contains(requireWhat, "oxygen") || strings.Contains(requireWhat, "%"):
		requirement := model.Requirement{
			Type: model.RequirementOxygen,
		}
		if isMax {
			requirement.Max = &requireNum
		} else {
			requirement.Min = &requireNum
		}
		requirements = append(requirements, requirement)
	case strings.Contains(requireWhat, "degrees") || strings.Contains(requireWhat, "°c"):
		requirement := model.Requirement{
			Type: model.RequirementTemperature,
		}
		if isMax {
			requirement.Max = &requireNum
		} else {
			requirement.Min = &requireNum
		}
		requirements = append(requirements, requirement)
	case strings.Contains(requireWhat, "production"):
		// Parse production requirements
		var resourceType model.ResourceType
		switch {
		case strings.Contains(requireWhat, "titanium"):
			resourceType = model.ResourceTitaniumProduction
		case strings.Contains(requireWhat, "steel"):
			resourceType = model.ResourceSteelProduction
		case strings.Contains(requireWhat, "plant"):
			resourceType = model.ResourcePlantsProduction
		case strings.Contains(requireWhat, "energy"):
			resourceType = model.ResourceEnergyProduction
		case strings.Contains(requireWhat, "heat"):
			resourceType = model.ResourceHeatProduction
		case strings.Contains(requireWhat, "m€") || strings.Contains(requireWhat, "credit"):
			resourceType = model.ResourceCreditsProduction
		}
		requirement := model.Requirement{
			Type:     model.RequirementProduction,
			Min:      &requireNum,
			Resource: &resourceType,
		}
		requirements = append(requirements, requirement)

	case strings.Contains(requireWhat, "tag"):
		// Parse tag requirements like "science tag" or "set of e/j/v tags"
		var tag model.CardTag
		switch {
		case strings.Contains(requireWhat, "science"):
			tag = model.TagScience
		case strings.Contains(requireWhat, "building"):
			tag = model.TagBuilding
		case strings.Contains(requireWhat, "space"):
			tag = model.TagSpace
		case strings.Contains(requireWhat, "power"):
			tag = model.TagPower
		case strings.Contains(requireWhat, "plant"):
			tag = model.TagPlant
		case strings.Contains(requireWhat, "microbe"):
			tag = model.TagMicrobe
		case strings.Contains(requireWhat, "animal"):
			tag = model.TagAnimal
		case strings.Contains(requireWhat, "earth"):
			tag = model.TagEarth
		case strings.Contains(requireWhat, "jovian"):
			tag = model.TagJovian
		case strings.Contains(requireWhat, "venus"):
			tag = model.TagVenus
		case strings.Contains(requireWhat, "city"):
			tag = model.TagCity
		}

		// For special cases like "e/j/v" tags, create multiple requirements
		if strings.Contains(requireWhat, "e/j/v") || strings.Contains(requireWhat, "earth/jovian/venus") {
			// Set of E/J/V tags means you need requireNum of each type
			for _, t := range []model.CardTag{model.TagEarth, model.TagJovian, model.TagVenus} {
				tagCopy := t
				requirement := model.Requirement{
					Type: model.RequirementTags,
					Min:  &requireNum,
					Tag:  &tagCopy,
				}
				requirements = append(requirements, requirement)
			}
		} else if tag != "" {
			requirement := model.Requirement{
				Type: model.RequirementTags,
				Min:  &requireNum,
				Tag:  &tag,
			}
			requirements = append(requirements, requirement)
		}

	case strings.Contains(requireWhat, "floater"):
		// Floater resource requirement (e.g., "Requires that you have 5 floaters")
		resourceType := model.ResourceFloaters
		requirement := model.Requirement{
			Type:     model.RequirementResource,
			Min:      &requireNum,
			Resource: &resourceType,
		}
		requirements = append(requirements, requirement)

	case strings.Contains(requireWhat, "microbe"):
		// Microbe resource requirement
		resourceType := model.ResourceMicrobes
		requirement := model.Requirement{
			Type:     model.RequirementResource,
			Min:      &requireNum,
			Resource: &resourceType,
		}
		requirements = append(requirements, requirement)

	case strings.Contains(requireWhat, "animal"):
		// Animal resource requirement
		resourceType := model.ResourceAnimals
		requirement := model.Requirement{
			Type:     model.RequirementResource,
			Min:      &requireNum,
			Resource: &resourceType,
		}
		requirements = append(requirements, requirement)

	case strings.Contains(requireWhat, "science"):
		// Science resource requirement
		resourceType := model.ResourceScience
		requirement := model.Requirement{
			Type:     model.RequirementResource,
			Min:      &requireNum,
			Resource: &resourceType,
		}
		requirements = append(requirements, requirement)
	}

	// Return requirements if we parsed any
	if len(requirements) > 0 {
		return requirements
	}
	return nil
}

func appendRequiredTag(tags []model.CardTag, tag model.CardTag, count int) []model.CardTag {
	// Add the tag 'count' number of times to represent the requirement
	for i := 0; i < count; i++ {
		tags = append(tags, tag)
	}
	return tags
}

func parseEffectsFromDescription(card *model.Card) {
	// Legacy global parameter lenience parsing removed - now handled by unified behaviors system
}

func parseDescription(record []string) string {
	// Try full text first (assuming it's around column 42)
	if len(record) > 42 && record[42] != "" {
		return strings.TrimSpace(record[42])
	}

	// Fallback to earlier columns
	var parts []string
	if len(record) > 40 && record[40] != "" {
		parts = append(parts, strings.TrimSpace(record[40]))
	}
	if len(record) > 39 && record[39] != "" {
		parts = append(parts, strings.TrimSpace(record[39]))
	}

	return strings.Join(parts, " ")
}

func parseVictoryConditionsFromBehaviors(behaviors []BehaviorData, record []string) []model.VictoryPointCondition {
	var conditions []model.VictoryPointCondition
	processedResourceTypes := make(map[string]bool) // Track which resource types we've already processed

	for _, behavior := range behaviors {
		// Look for VP conditions in card resources
		if behavior.CardResources != "" && behavior.ResourceType != "" {
			// Skip if we already processed this resource type for this card
			if processedResourceTypes[behavior.ResourceType] {
				continue
			}
			processedResourceTypes[behavior.ResourceType] = true
			// Parse VP base from cards.csv VP amount column, not CardResources
			vpBase := 1 // Default
			if len(record) > colVPAmount {
				if amount := parseIntFromString(record[colVPAmount]); amount > 0 {
					vpBase = amount
				}
			}

			// Convert resource type string to ResourceType
			resourceType := parseResourceTypeFromString(behavior.ResourceType)
			if resourceType == nil {
				continue
			}

			// Determine VP per resource requirement from description patterns
			per := 1 // Default: 1 VP per 1 resource

			// Check for patterns like "1 VP per 2 animals"
			desc := strings.ToLower(behavior.TextForm)
			if strings.Contains(desc, "vp per 2") {
				per = 2
			} else if strings.Contains(desc, "vp per 3") {
				per = 3
			} else if strings.Contains(desc, "vp per 4") {
				per = 4
			}

			// Determine what to count from based on "Where" column
			target := model.TargetSelfCard // Default - count from this card
			if strings.ToLower(behavior.Where) == "any" {
				target = model.TargetAny // Count globally
			}

			// Check cards.csv "per" column for maxTrigger
			maxTrigger := -1 // Default: unlimited
			if len(record) > colVPPer {
				vpPer := strings.ToLower(strings.TrimSpace(record[colVPPer]))
				if strings.Contains(vpPer, "first resource only") || strings.Contains(vpPer, "once") {
					maxTrigger = 1
				}
			}

			// Convert target to location
			var location model.CardApplyLocation
			if target == model.TargetAny {
				location = model.CardApplyLocationAnywhere
			} else {
				location = model.CardApplyLocationAnywhere // Default for self-card resources
			}

			condition := model.VictoryPointCondition{
				Amount:     vpBase,
				Condition:  model.VPConditionPer,
				MaxTrigger: &maxTrigger,
				Per: &model.PerCondition{
					Type:     *resourceType,
					Amount:   per,
					Location: &location,
				},
			}

			conditions = append(conditions, condition)
		}
	}

	return conditions
}

func parseIntFromString(s string) int {
	if val, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return val
	}
	return 0
}

func parseResourceTypeFromString(s string) *model.ResourceType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "animal":
		r := model.ResourceAnimals
		return &r
	case "microbe":
		r := model.ResourceMicrobes
		return &r
	case "floater":
		r := model.ResourceFloaters
		return &r
	case "fighter":
		r := model.ResourceAsteroid
		return &r
	case "science":
		r := model.ResourceScience
		return &r
	default:
		return nil
	}
}

func parseVictoryConditionsFromCardsCSV(record []string) []model.VictoryPointCondition {
	var conditions []model.VictoryPointCondition

	// Look for resource type, VP amount, and per condition in cards.csv
	colResourceTypeCSV := 36 // Resource type like "science"
	colVPAmountCSV := 37     // VP amount like "3"
	colVPPerCSV := 38        // Per condition like "first resource only"

	if len(record) <= colVPPerCSV {
		return conditions
	}

	resourceTypeStr := strings.ToLower(strings.TrimSpace(record[colResourceTypeCSV]))
	vpAmountStr := strings.TrimSpace(record[colVPAmountCSV])
	vpPerStr := strings.ToLower(strings.TrimSpace(record[colVPPerCSV]))

	if vpAmountStr == "" {
		return conditions
	}

	vpAmount := parseIntFromString(vpAmountStr)
	if vpAmount <= 0 {
		return conditions
	}

	// Check if this is a tag-based VP condition or resource-based
	var resourceType *model.ResourceType
	if resourceTypeStr != "" {
		resourceType = parseResourceTypeFromString(resourceTypeStr)
		if resourceType == nil {
			return conditions
		}
	}

	// If "per" column is empty, it's a fixed VP condition, not "per resource"
	if vpPerStr == "" {
		condition := model.VictoryPointCondition{
			Amount:    vpAmount,
			Condition: model.VPConditionFixed,
			// For fixed VP, don't include MaxTrigger or Per
		}
		conditions = append(conditions, condition)
		return conditions
	}

	// Otherwise, it's a "per something" condition
	maxTrigger := -1 // Default: unlimited
	if strings.Contains(vpPerStr, "first resource only") || strings.Contains(vpPerStr, "once") {
		maxTrigger = 1
	}

	var condition model.VictoryPointCondition

	// Check if this is a tag-based condition
	if strings.Contains(vpPerStr, "tag") {
		// For tag-based VP conditions like "1 VP per jovian tag"
		// Parse tag type from vpPerStr
		var tagType model.CardTag
		switch {
		case strings.Contains(vpPerStr, "jovian"):
			tagType = model.TagJovian
		case strings.Contains(vpPerStr, "earth"):
			tagType = model.TagEarth
		case strings.Contains(vpPerStr, "venus"):
			tagType = model.TagVenus
		case strings.Contains(vpPerStr, "science"):
			tagType = model.TagScience
		case strings.Contains(vpPerStr, "space"):
			tagType = model.TagSpace
		case strings.Contains(vpPerStr, "building"):
			tagType = model.TagBuilding
		case strings.Contains(vpPerStr, "power"):
			tagType = model.TagPower
		case strings.Contains(vpPerStr, "plant"):
			tagType = model.TagPlant
		case strings.Contains(vpPerStr, "microbe"):
			tagType = model.TagMicrobe
		case strings.Contains(vpPerStr, "animal"):
			tagType = model.TagAnimal
		case strings.Contains(vpPerStr, "city"):
			tagType = model.TagCity
		default:
			// Unknown tag type, skip this condition
			return conditions
		}

		// Create tag-based VP condition
		condition = model.VictoryPointCondition{
			Amount:     vpAmount,
			Condition:  model.VPConditionPer,
			MaxTrigger: &maxTrigger,
			Per: &model.PerCondition{
				Type:     model.ResourceTag,
				Amount:   1, // 1 VP per 1 tag (standard for tag-based VP)
				Location: func() *model.CardApplyLocation { loc := model.CardApplyLocationAnywhere; return &loc }(),
				Tag:      &tagType,
			},
		}
	} else if resourceType != nil {
		// Resource-based VP condition
		// Extract the resource amount from vpPerStr (e.g., "4 resources here" -> 4)
		per := vpAmount // Default fallback
		if matches := regexp.MustCompile(`\b(\d+)\b`).FindString(vpPerStr); matches != "" {
			if num, err := strconv.Atoi(matches); err == nil {
				per = num
			}
		}

		// Check if the condition refers to resources on this card ("resource here" or "resources here")
		var target *model.TargetType
		if strings.Contains(strings.ToLower(vpPerStr), "here") {
			t := model.TargetSelfCard
			target = &t
		}

		condition = model.VictoryPointCondition{
			Amount:     vpAmount, // VP amount from column 37
			Condition:  model.VPConditionPer,
			MaxTrigger: &maxTrigger,
			Per: &model.PerCondition{
				Type:     *resourceType,
				Amount:   per,                                                                                       // Resources needed per VP
				Location: func() *model.CardApplyLocation { loc := model.CardApplyLocationAnywhere; return &loc }(), // Default to anywhere, could be refined based on context
				Target:   target,                                                                                    // Set to self-card when "here" is detected
			},
		}
	} else {
		// Check for tile-based VP conditions (cities, oceans, greeneries, colonies)
		var tileType model.ResourceType
		var perAmount int = 1 // Default amount per VP

		// Extract number from vpPerStr if present (e.g., "any 3 cities" -> 3)
		if matches := regexp.MustCompile(`\b(\d+)\b`).FindString(vpPerStr); matches != "" {
			if num, err := strconv.Atoi(matches); err == nil {
				perAmount = num
			}
		}

		switch {
		case strings.Contains(vpPerStr, "cities") || strings.Contains(vpPerStr, "city"):
			tileType = model.ResourceCityTile
		case strings.Contains(vpPerStr, "oceans") || strings.Contains(vpPerStr, "ocean"):
			tileType = model.ResourceOceanTile
		case strings.Contains(vpPerStr, "greeneries") || strings.Contains(vpPerStr, "greenery"):
			tileType = model.ResourceGreeneryTile
		case strings.Contains(vpPerStr, "colonies") || strings.Contains(vpPerStr, "colony"):
			tileType = model.ResourceColonyTile
		default:
			// Unknown per condition, skip
			return conditions
		}

		// Determine location
		location := model.CardApplyLocationAnywhere
		if strings.Contains(vpPerStr, "mars") {
			location = model.CardApplyLocationMars
		}

		condition = model.VictoryPointCondition{
			Amount:     vpAmount, // VP amount (e.g., 1 VP)
			Condition:  model.VPConditionPer,
			MaxTrigger: &maxTrigger,
			Per: &model.PerCondition{
				Type:     tileType,
				Amount:   perAmount, // Amount needed per VP (e.g., 3 cities)
				Location: &location,
			},
		}
	}

	conditions = append(conditions, condition)
	return conditions
}

// isGlobalParameterLenienceEffect checks if a behavior represents a global parameter lenience effect
func isGlobalParameterLenienceEffect(behavior BehaviorData) bool {
	text := strings.ToLower(behavior.TextForm)
	return strings.Contains(text, "global requirements") ||
		strings.Contains(text, "global parameter requirement") ||
		strings.Contains(text, "weaken global reqt")
}

// createImmediateBehavior creates an immediate behavior (non-triggered) from behavior data
func createImmediateBehavior(behavior BehaviorData, isAttack bool) *model.CardBehavior {
	autoTrigger := model.ResourceTriggerAuto
	resourceExchange := createResourceExchange(behavior, isAttack, &autoTrigger, nil, true)
	if resourceExchange != nil {
		return &model.CardBehavior{
			Triggers: resourceExchange.Triggers,
			Inputs:   resourceExchange.Inputs,
			Outputs:  resourceExchange.Outputs,
		}
	}
	return nil
}

// createAttackBehavior creates an immediate behavior for attack effects
func createAttackBehavior(behavior BehaviorData) *model.CardBehavior {
	return createImmediateBehavior(behavior, true)
}

// createTriggeredEffectBehavior creates a behavior for triggered effects
func createTriggeredEffectBehavior(behavior BehaviorData) *model.CardBehavior {
	// Parse trigger condition from Trigger column ("ANY city", "ANY ocean", etc.)
	triggerCondition := parseTriggerCondition(behavior.Trigger)

	// Create ResourceExchange with trigger condition
	// The trigger condition is now handled by the new trigger system in ResourceExchange
	autoTrigger := model.ResourceTriggerAuto
	resourceExchange := createResourceExchange(behavior, false, &autoTrigger, triggerCondition, true)

	if resourceExchange != nil {
		return &model.CardBehavior{
			Triggers: resourceExchange.Triggers,
			Inputs:   resourceExchange.Inputs,
			Outputs:  resourceExchange.Outputs,
		}
	}
	return nil
}

// createDiscountBehavior creates a behavior for discount effects using unified system
func createDiscountBehavior(behavior BehaviorData) *model.CardBehavior {
	// Get the discount amount from Megacredits column (positive value represents discount)
	if behavior.Megacredits <= 0 {
		return nil
	}

	// Parse tags from the Trigger field (same logic as old createDiscountEffect)
	var affectedTags []model.CardTag
	tagMapping := map[string]model.CardTag{
		"space tag":    model.TagSpace,
		"power tag":    model.TagPower,
		"building tag": model.TagBuilding,
		"science tag":  model.TagScience,
		"earth tag":    model.TagEarth,
		"microbe tag":  model.TagMicrobe,
		"animal tag":   model.TagAnimal,
		"plant tag":    model.TagPlant,
		"event tag":    model.TagEvent,
		"city tag":     model.TagCity,
		"venus tag":    model.TagVenus,
		"jovian tag":   model.TagJovian,
	}

	if behavior.Trigger != "" && behavior.Trigger != "play card" {
		// Make trigger comparison case-insensitive
		triggerLower := strings.ToLower(behavior.Trigger)
		if tag, exists := tagMapping[triggerLower]; exists {
			affectedTags = []model.CardTag{tag}
		}
	}
	// If Trigger is "play card" or empty, affectedTags remains empty (applies to all cards)

	// Create discount resource condition
	discountCondition := model.ResourceCondition{
		Type:         model.ResourceDiscount,
		Amount:       behavior.Megacredits,
		Target:       model.TargetSelfPlayer,
		AffectedTags: affectedTags,
	}

	// Create ongoing trigger (discount applies to future card plays)
	trigger := model.Trigger{
		Type: model.ResourceTriggerAuto, // Discount is applied automatically when playing cards
	}

	resourceExchange := &model.ResourceExchange{
		Triggers: []model.Trigger{trigger},
		Outputs:  []model.ResourceCondition{discountCondition},
	}

	return &model.CardBehavior{
		Triggers: resourceExchange.Triggers,
		Inputs:   resourceExchange.Inputs,
		Outputs:  resourceExchange.Outputs,
	}
}

// createValueModifierBehavior creates a CardBehavior with value-modifier output from Effect/Resource behaviors
func createValueModifierBehavior(behavior BehaviorData) *model.CardBehavior {
	// Check if this is a value modifier behavior: Megacredits should be "N" and NEquals should have resource info
	if behavior.MegacreditsRaw != "N" || behavior.NEquals == "" {
		return nil
	}

	// Parse affected resources from NEquals field (e.g., "steel spent", "titanium spent")
	var affectedResources []string
	nEqualsLower := strings.ToLower(behavior.NEquals)

	if strings.Contains(nEqualsLower, "steel spent") {
		affectedResources = append(affectedResources, "steel")
	}
	if strings.Contains(nEqualsLower, "titanium spent") {
		affectedResources = append(affectedResources, "titanium")
	}

	// If no recognized resources found, skip this behavior
	if len(affectedResources) == 0 {
		return nil
	}

	// Create auto trigger for ongoing value modifier effects
	autoTrigger := model.ResourceTriggerAuto
	trigger := model.Trigger{
		Type: autoTrigger,
	}

	// Create value modifier output
	valueModifierOutput := model.ResourceCondition{
		Type:              model.ResourceValueModifier,
		Amount:            1, // +1 value increase
		Target:            model.TargetSelfPlayer,
		AffectedResources: affectedResources,
	}

	return &model.CardBehavior{
		Triggers: []model.Trigger{trigger},
		Inputs:   []model.ResourceCondition{},
		Outputs:  []model.ResourceCondition{valueModifierOutput},
	}
}

// extractInputsAndOutputsFromChoice extracts and separates inputs/outputs from a choice
// Inputs are only for active cards where negative values represent spending/activation costs
// For non-active cards, all effects (including negative) are outputs representing "what happens"
func extractInputsAndOutputsFromChoice(choice BehaviorData, isActiveCard bool, isAttack bool) (inputs []model.ResourceCondition, outputs []model.ResourceCondition) {
	// Get all effects
	allEffects := extractAllEffectsFromChoice(choice, isAttack)

	// For active cards: negative = inputs (spending), positive = outputs (gaining)
	// For non-active cards: everything is outputs (including decreases/negative amounts)
	for _, effect := range allEffects {
		if effect.Amount < 0 && isActiveCard {
			// Negative amount = input (cost/spending) - only for active cards
			// Convert to positive amount for the input
			inputEffect := effect
			inputEffect.Amount = -effect.Amount
			inputs = append(inputs, inputEffect)
		} else if effect.Amount != 0 {
			// All non-zero amounts are outputs (for non-active cards or positive amounts)
			outputs = append(outputs, effect)
		}
		// Ignore zero amounts
	}

	return inputs, outputs
}

// extractAllEffectsFromChoice extracts all possible effects from a single choice option
func extractAllEffectsFromChoice(choice BehaviorData, isAttack bool) []model.ResourceCondition {
	var effects []model.ResourceCondition

	// Check for N-based conditional outputs first
	if nCondition := parseNBasedConditionalOutput(choice); nCondition != nil {
		effects = append(effects, *nCondition)
	}

	// Basic resources
	resourceTypes := []struct {
		value        int
		resourceType model.ResourceType
	}{
		{choice.Megacredits, model.ResourceCredits},
		{choice.Steel, model.ResourceSteel},
		{choice.Titanium, model.ResourceTitanium},
		{choice.Plants, model.ResourcePlants},
		{choice.Energy, model.ResourceEnergy},
		{choice.Heat, model.ResourceHeat},
	}

	for _, res := range resourceTypes {
		if res.value != 0 {
			// Use attack-aware targeting and amount negation
			target := model.TargetSelfPlayer
			amount := res.value
			if isAttack {
				target = model.TargetAnyPlayer
				amount = -res.value // Negate for attacks (removing resources)
			}
			effects = append(effects, model.ResourceCondition{
				Type:   res.resourceType,
				Amount: amount,
				Target: target,
			})
		}
	}

	// Card resources (like floaters, microbes, animals)
	if choice.CardResources != "" && choice.ResourceType != "" {
		resourceType := parseResourceType(choice.ResourceType)
		target := model.TargetAny // Default for card resources
		if choice.Where == "self" || choice.Where == "here" {
			target = model.TargetSelfCard
		} else if choice.Where == "any" {
			target = model.TargetAny // "any" means any card for card resources
		}

		// For attacks, update targeting and negate amounts
		amount := parseResourceAmount(choice.CardResources)
		if isAttack {
			target = model.TargetAny // Keep "any" for card resources in attacks
			amount = -amount         // Negate for attacks (removing resources)
		}

		affectedTags := parseWhereToAffectedTags(choice.Where)
		effects = append(effects, model.ResourceCondition{
			Type:         resourceType,
			Amount:       amount,
			Target:       target,
			AffectedTags: affectedTags,
		})
	}

	// Terraforming actions
	if choice.RaiseTemp > 0 {
		effects = append(effects, model.ResourceCondition{
			Type:   model.ResourceTemperature,
			Amount: choice.RaiseTemp,
			Target: model.TargetNone,
		})
	}
	if choice.RaiseOxygen > 0 {
		effects = append(effects, model.ResourceCondition{
			Type:   model.ResourceOxygen,
			Amount: choice.RaiseOxygen,
			Target: model.TargetNone,
		})
	}
	if choice.RaiseVenus > 0 {
		effects = append(effects, model.ResourceCondition{
			Type:   model.ResourceVenus,
			Amount: choice.RaiseVenus,
			Target: model.TargetNone,
		})
	}

	// Tile placements
	if choice.CityTile > 0 {
		effects = append(effects, model.ResourceCondition{
			Type:   model.ResourceCityPlacement,
			Amount: choice.CityTile,
			Target: model.TargetNone,
		})
	}
	if choice.OceanTile > 0 {
		effects = append(effects, model.ResourceCondition{
			Type:   model.ResourceOceanPlacement,
			Amount: choice.OceanTile,
			Target: model.TargetNone,
		})
	}
	if choice.GreeneryTile > 0 {
		effects = append(effects, model.ResourceCondition{
			Type:   model.ResourceGreeneryPlacement,
			Amount: choice.GreeneryTile,
			Target: model.TargetNone,
		})
	}

	// Cards
	if choice.Cards > 0 {
		effects = append(effects, model.ResourceCondition{
			Type:   model.ResourceCardDraw,
			Amount: choice.Cards,
			Target: model.TargetSelfPlayer,
		})
	}

	// Production resources
	productionTypes := []struct {
		value        int
		resourceType model.ResourceType
	}{
		{choice.MegacreditProd, model.ResourceCreditsProduction},
		{choice.SteelProd, model.ResourceSteelProduction},
		{choice.TitaniumProd, model.ResourceTitaniumProduction},
		{choice.PlantProd, model.ResourcePlantsProduction},
		{choice.EnergyProd, model.ResourceEnergyProduction},
		{choice.HeatProd, model.ResourceHeatProduction},
	}

	for _, prod := range productionTypes {
		if prod.value != 0 {
			effects = append(effects, model.ResourceCondition{
				Type:   prod.resourceType,
				Amount: prod.value,
				Target: model.TargetSelfPlayer,
			})
		}
	}

	return effects
}

// separateCommonAndDifferentEffects identifies common effects vs choice-specific effects
func separateCommonAndDifferentEffects(allChoiceEffects [][]model.ResourceCondition) ([]model.ResourceCondition, [][]model.ResourceCondition) {
	if len(allChoiceEffects) == 0 {
		return nil, nil
	}

	// Find effects that appear in ALL choices
	var commonEffects []model.ResourceCondition
	var choiceSpecificEffects [][]model.ResourceCondition

	// For each effect in the first choice, check if it appears in all other choices
	for _, effect := range allChoiceEffects[0] {
		isCommon := true
		for i := 1; i < len(allChoiceEffects); i++ {
			if !containsEffect(allChoiceEffects[i], effect) {
				isCommon = false
				break
			}
		}
		if isCommon {
			commonEffects = append(commonEffects, effect)
		}
	}

	// Extract choice-specific effects (everything that's not common)
	for _, choiceEffects := range allChoiceEffects {
		var specificEffects []model.ResourceCondition
		for _, effect := range choiceEffects {
			if !containsEffect(commonEffects, effect) {
				specificEffects = append(specificEffects, effect)
			}
		}
		choiceSpecificEffects = append(choiceSpecificEffects, specificEffects)
	}

	return commonEffects, choiceSpecificEffects
}

// containsEffect checks if an effect list contains a specific effect
func containsEffect(effects []model.ResourceCondition, target model.ResourceCondition) bool {
	for _, effect := range effects {
		if effect.Type == target.Type && effect.Amount == target.Amount && effect.Target == target.Target {
			return true
		}
	}
	return false
}

// hasNonEmptyChoices checks if any choice has at least one effect
func hasNonEmptyChoices(choices [][]model.ResourceCondition) bool {
	for _, choice := range choices {
		if len(choice) > 0 {
			return true
		}
	}
	return false
}

// createChoiceBehavior creates a CardBehavior with choice outputs from multiple option behaviors
func createChoiceBehavior(choiceBehavior ChoiceBehaviorData, card *model.Card) *model.CardBehavior {
	if len(choiceBehavior.Choices) == 0 {
		return nil
	}

	// Create trigger based on behavior type
	var triggerType model.ResourceTriggerType
	if choiceBehavior.BehaviorType == "2 Action" {
		// Action cards should have manual triggers (player-activated)
		triggerType = model.ResourceTriggerManual
	} else {
		// Immediate effects should have auto triggers
		triggerType = model.ResourceTriggerAuto
	}
	trigger := model.Trigger{
		Type: triggerType,
	}

	var outputs []model.ResourceCondition

	// Detect if this is an attack behavior
	isAttack := choiceBehavior.BehaviorType == "1 Immediate/Attack"

	// Extract all effects from each choice option first
	allChoiceEffects := make([][]model.ResourceCondition, len(choiceBehavior.Choices))
	for i, choice := range choiceBehavior.Choices {
		allChoiceEffects[i] = extractAllEffectsFromChoice(choice, isAttack)
	}

	// Identify common effects that appear in ALL choices
	var commonEffects []model.ResourceCondition
	var choiceSpecificEffects [][]model.ResourceCondition

	if len(allChoiceEffects) > 0 {
		commonEffects, choiceSpecificEffects = separateCommonAndDifferentEffects(allChoiceEffects)
	}

	// Convert choice-specific effects to Choice structs with proper input/output separation
	var choices []model.Choice
	if len(choiceSpecificEffects) > 0 && hasNonEmptyChoices(choiceSpecificEffects) {
		for _, choiceEffects := range choiceSpecificEffects {
			// Use the choice-specific effects (with common effects already removed)
			// We need to separate inputs and outputs from these specific effects
			var inputs, outputs []model.ResourceCondition
			isActiveCard := card.Type == "active"

			// For active cards: negative amounts = inputs, positive = outputs
			// For non-active cards: all effects are outputs
			for _, effect := range choiceEffects {
				if effect.Amount < 0 && isActiveCard {
					// Negative amount = input (cost/spending) - only for active cards
					inputEffect := effect
					inputEffect.Amount = -effect.Amount
					inputs = append(inputs, inputEffect)
				} else if effect.Amount != 0 {
					// All non-zero amounts are outputs (for non-active cards or positive amounts)
					outputs = append(outputs, effect)
				}
			}

			choice := model.Choice{
				Inputs:  inputs,
				Outputs: outputs,
			}
			choices = append(choices, choice)
		}
	}

	// Add common effects as regular outputs
	outputs = append(outputs, commonEffects...)

	// Return behavior with choices at the top level
	return &model.CardBehavior{
		Triggers: []model.Trigger{trigger},
		Inputs:   []model.ResourceCondition{},
		Outputs:  outputs,
		Choices:  choices,
	}
}

// createTriggeredChoiceBehavior creates a CardBehavior with choice outputs for triggered effects
func createTriggeredChoiceBehavior(choiceBehavior ChoiceBehaviorData, card *model.Card) *model.CardBehavior {
	if len(choiceBehavior.Choices) == 0 {
		return nil
	}

	// Parse trigger condition from the Trigger column (not Where column)
	var triggerCondition *model.ResourceTriggerCondition

	triggerText := strings.ToLower(strings.TrimSpace(choiceBehavior.Trigger))
	if strings.Contains(triggerText, "science tag") {
		triggerCondition = &model.ResourceTriggerCondition{
			Type:         model.TriggerTagPlayed,
			AffectedTags: []model.CardTag{model.TagScience},
		}
	} else {
		// Fallback to the existing parseTriggerCondition function
		triggerCondition = parseTriggerCondition(choiceBehavior.Trigger)
	}

	// Create auto trigger with the parsed condition
	autoTrigger := model.ResourceTriggerAuto
	trigger := model.Trigger{
		Type:      autoTrigger,
		Condition: triggerCondition,
	}

	var outputs []model.ResourceCondition

	// Detect if this is an attack behavior
	isAttack := choiceBehavior.BehaviorType == "1 Immediate/Attack"

	// Extract all effects from each choice using general parsing
	var choices [][]model.ResourceCondition

	for _, choice := range choiceBehavior.Choices {
		choiceActions := extractAllEffectsFromChoice(choice, isAttack)
		if len(choiceActions) > 0 {
			choices = append(choices, choiceActions)
		}
	}

	// Convert choices to Choice structs with proper input/output separation
	var behaviorChoices []model.Choice
	for i, choiceActions := range choices {
		// Use the original choice data to get proper input/output separation
		var inputs, outputs []model.ResourceCondition
		if i < len(choiceBehavior.Choices) {
			isActiveCard := card.Type == "active"
			inputs, outputs = extractInputsAndOutputsFromChoice(choiceBehavior.Choices[i], isActiveCard, isAttack)
		} else {
			// Fallback: if we can't match to original choice, put everything as outputs
			outputs = choiceActions
		}

		choice := model.Choice{
			Inputs:  inputs,
			Outputs: outputs,
		}
		behaviorChoices = append(behaviorChoices, choice)
	}

	return &model.CardBehavior{
		Triggers: []model.Trigger{trigger},
		Inputs:   []model.ResourceCondition{},
		Outputs:  outputs,
		Choices:  behaviorChoices,
	}
}

// parseResourceType converts string resource type to ResourceType
func parseResourceType(resourceType string) model.ResourceType {
	switch strings.ToLower(resourceType) {
	case "microbe":
		return model.ResourceMicrobes
	case "animal":
		return model.ResourceAnimals
	case "floater":
		return model.ResourceFloaters
	case "science":
		return model.ResourceScience
	case "asteroid":
		return model.ResourceAsteroid
	case "disease":
		return model.ResourceDisease
	default:
		return model.ResourceCredits // fallback
	}
}

// parseResourceAmount parses amount from card resources string
func parseResourceAmount(cardResources string) int {
	// Extract number from string like "3" or "2"
	amount, err := strconv.Atoi(strings.TrimSpace(cardResources))
	if err != nil {
		return 1 // fallback
	}
	return amount
}

// createGlobalParameterLenienceAction creates an immediate action for global parameter lenience
func createGlobalParameterLenienceBehavior(behavior BehaviorData) *model.CardBehavior {
	// Parse the lenience amount from the text
	lenienceAmount := 2 // Default for Adaptation Technology
	text := strings.ToLower(behavior.TextForm)

	// Try to extract specific lenience amount from text
	if strings.Contains(text, "+2 or -2") || strings.Contains(text, "2 steps") {
		lenienceAmount = 2
	}

	// Create ResourceExchange with output "global-parameter-lenience" and trigger for card played
	autoTrigger := model.ResourceTriggerAuto
	cardPlayedTrigger := model.TriggerCardPlayed

	triggerCondition := &model.ResourceTriggerCondition{
		Type: cardPlayedTrigger,
	}

	// Create trigger for when card is played
	trigger := model.Trigger{
		Type:      autoTrigger,
		Condition: triggerCondition,
	}

	// No input needed for global parameter lenience - it's an ongoing effect

	// Output: global parameter lenience effect
	lenienceOutput := model.ResourceCondition{
		Type:   model.ResourceGlobalParameterLenience,
		Amount: lenienceAmount,

		Target: model.TargetSelfPlayer, // The lenience applies to the player who played the card
	}

	resourceExchange := &model.ResourceExchange{
		Triggers: []model.Trigger{trigger},
		Inputs:   []model.ResourceCondition{}, // No inputs for global parameter lenience
		Outputs:  []model.ResourceCondition{lenienceOutput},
	}

	return &model.CardBehavior{
		Triggers: resourceExchange.Triggers,
		Inputs:   resourceExchange.Inputs,
		Outputs:  resourceExchange.Outputs,
	}
}

// DEPRECATED: enhanceGlobalParameterLenienceActions - replaced by createGlobalParameterLenienceAction
