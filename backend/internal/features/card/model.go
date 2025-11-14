package card

import (
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/shared/types"
)

// Type aliases
type ResourceType = types.ResourceType
type StandardProject = types.StandardProject
type HexPosition = types.HexPosition

// CardType represents different types of cards in Terraforming Mars
type CardType string

const (
	CardTypeAutomated   CardType = "automated"   // Green cards - immediate effects, production bonuses (was "effect")
	CardTypeActive      CardType = "active"      // Blue cards - ongoing effects, repeatable actions
	CardTypeEvent       CardType = "event"       // Red cards - one-time effects
	CardTypeCorporation CardType = "corporation" // Corporation cards - unique player abilities
	CardTypePrelude     CardType = "prelude"     // Prelude cards - setup phase cards
)

// ProductionEffects represents changes to resource production
type ProductionEffects struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// Card represents a game card
type Card struct {
	ID              string                  `json:"id" ts:"string"`
	Name            string                  `json:"name" ts:"string"`
	Type            CardType                `json:"type" ts:"CardType"`
	Cost            int                     `json:"cost" ts:"number"`
	Description     string                  `json:"description" ts:"string"`
	Pack            string                  `json:"pack" ts:"string"` // Card pack identifier (e.g., "base-game", "corporate-era", "prelude")
	Tags            []CardTag               `json:"tags,omitempty" ts:"CardTag[] | undefined"`
	Requirements    []Requirement           `json:"requirements,omitempty" ts:"Requirement[] | undefined"`
	Behaviors       []CardBehavior          `json:"behaviors,omitempty" ts:"CardBehavior[] | undefined"`             // All card behaviors (immediate and repeatable)
	ResourceStorage *ResourceStorage        `json:"resourceStorage,omitempty" ts:"ResourceStorage | undefined"`      // Cards that can hold resources
	VPConditions    []VictoryPointCondition `json:"vpConditions,omitempty" ts:"VictoryPointCondition[] | undefined"` // VP per X conditions

	// Corporation-specific fields (nil for non-corporation cards)
	StartingResources  *ResourceSet `json:"startingResources,omitempty" ts:"ResourceSet | undefined"`  // Parsed from first auto behavior (corporations only)
	StartingProduction *ResourceSet `json:"startingProduction,omitempty" ts:"ResourceSet | undefined"` // Parsed from first auto behavior (corporations only)
}

// DeepCopy creates a deep copy of the Card
func (c Card) DeepCopy() Card {
	// Copy slices
	tags := make([]CardTag, len(c.Tags))
	copy(tags, c.Tags)

	requirements := make([]Requirement, len(c.Requirements))
	copy(requirements, c.Requirements)

	behaviors := make([]CardBehavior, len(c.Behaviors))
	for i, behavior := range c.Behaviors {
		behaviors[i] = behavior.DeepCopy()
	}

	vpConditions := make([]VictoryPointCondition, len(c.VPConditions))
	copy(vpConditions, c.VPConditions)

	// Copy resource storage
	var resourceStorage *ResourceStorage
	if c.ResourceStorage != nil {
		rs := *c.ResourceStorage
		resourceStorage = &rs
	}

	// Copy corporation-specific fields
	var startingResources *ResourceSet
	if c.StartingResources != nil {
		rs := *c.StartingResources
		startingResources = &rs
	}

	var startingProduction *ResourceSet
	if c.StartingProduction != nil {
		sp := *c.StartingProduction
		startingProduction = &sp
	}

	return Card{
		ID:                 c.ID,
		Name:               c.Name,
		Type:               c.Type,
		Cost:               c.Cost,
		Description:        c.Description,
		Pack:               c.Pack,
		Tags:               tags,
		Requirements:       requirements,
		Behaviors:          behaviors,
		ResourceStorage:    resourceStorage,
		VPConditions:       vpConditions,
		StartingResources:  startingResources,
		StartingProduction: startingProduction,
	}
}

// CardTag represents different card categories and attributes
type CardTag string

const (
	TagSpace    CardTag = "space"
	TagEarth    CardTag = "earth"
	TagScience  CardTag = "science"
	TagPower    CardTag = "power"
	TagBuilding CardTag = "building"
	TagMicrobe  CardTag = "microbe"
	TagAnimal   CardTag = "animal"
	TagPlant    CardTag = "plant"
	TagEvent    CardTag = "event"
	TagCity     CardTag = "city"
	TagVenus    CardTag = "venus"
	TagJovian   CardTag = "jovian"
	TagWildlife CardTag = "wildlife"
	TagWild     CardTag = "wild"
)

// ResourceSet is a type alias to avoid circular imports
// The actual definition is in internal/features/resources/service.go
type ResourceSet = resources.ResourceSet

// CardRequirements defines what conditions must be met to play a card
type CardRequirements struct {
	// MinTemperature is the minimum global temperature required (-30 to +8)
	MinTemperature *int `json:"minTemperature,omitempty" ts:"number | undefined"`

	// MaxTemperature is the maximum global temperature allowed (-30 to +8)
	MaxTemperature *int `json:"maxTemperature,omitempty" ts:"number | undefined"`

	// MinOxygen is the minimum oxygen percentage required (0-14)
	MinOxygen *int `json:"minOxygen,omitempty" ts:"number | undefined"`

	// MaxOxygen is the maximum oxygen percentage allowed (0-14)
	MaxOxygen *int `json:"maxOxygen,omitempty" ts:"number | undefined"`

	// MinOceans is the minimum number of ocean tiles required (0-9)
	MinOceans *int `json:"minOceans,omitempty" ts:"number | undefined"`

	// MaxOceans is the maximum number of ocean tiles allowed (0-9)
	MaxOceans *int `json:"maxOceans,omitempty" ts:"number | undefined"`

	// RequiredTags are tags that the player must have from played cards
	RequiredTags []CardTag `json:"requiredTags,omitempty" ts:"CardTag[] | undefined"`

	// RequiredProduction specifies minimum production requirements
	RequiredProduction *ResourceSet `json:"requiredProduction,omitempty" ts:"ResourceSet | undefined"`
}

// TriggerType represents different trigger conditions
type TriggerType string

const (
	TriggerOceanPlaced           TriggerType = "ocean-placed"
	TriggerTemperatureRaise      TriggerType = "temperature-raise"
	TriggerOxygenRaise           TriggerType = "oxygen-raise"
	TriggerCityPlaced            TriggerType = "city-placed"
	TriggerGreeneryPlaced        TriggerType = "greenery-placed"
	TriggerTilePlaced            TriggerType = "tile-placed"
	TriggerCardPlayed            TriggerType = "card-played"
	TriggerStandardProjectPlayed TriggerType = "standard-project-played"
	TriggerTagPlayed             TriggerType = "tag-played"
	TriggerProductionIncreased   TriggerType = "production-increased"
	TriggerPlacementBonusGained  TriggerType = "placement-bonus-gained"
	TriggerAlwaysActive          TriggerType = "always-active"
)

// TerraformingActions represents tile placement actions
type TerraformingActions struct {
	CityPlacement     int `json:"cityPlacement,omitempty" ts:"number"`     // Number of city tiles to place
	OceanPlacement    int `json:"oceanPlacement,omitempty" ts:"number"`    // Number of ocean tiles to place
	GreeneryPlacement int `json:"greeneryPlacement,omitempty" ts:"number"` // Number of greenery tiles to place
}

// RequirementType represents different types of card requirements
type RequirementType string

const (
	RequirementTemperature RequirementType = "temperature" // Global temperature requirement
	RequirementOxygen      RequirementType = "oxygen"      // Global oxygen requirement
	RequirementOceans      RequirementType = "oceans"      // Ocean tiles requirement
	RequirementVenus       RequirementType = "venus"       // Venus terraforming requirement
	RequirementCities      RequirementType = "cities"      // City tiles requirement
	RequirementGreeneries  RequirementType = "greeneries"  // Greenery tiles requirement
	RequirementTags        RequirementType = "tags"        // Tag requirement (e.g., science tags)
	RequirementProduction  RequirementType = "production"  // Production requirement
	RequirementTR          RequirementType = "tr"          // Terraform Rating requirement
	RequirementResource    RequirementType = "resource"    // Resource requirement (e.g., floaters, microbes)
)

// Requirement represents a single card requirement with flexible min/max values
type Requirement struct {
	Type     RequirementType    `json:"type" ts:"RequirementType"`                             // Type of requirement
	Min      *int               `json:"min,omitempty" ts:"number | undefined"`                 // Minimum value required
	Max      *int               `json:"max,omitempty" ts:"number | undefined"`                 // Maximum value allowed
	Location *CardApplyLocation `json:"location,omitempty" ts:"CardApplyLocation | undefined"` // Location constraint (Mars, anywhere, etc.)
	Tag      *CardTag           `json:"tag,omitempty" ts:"CardTag | undefined"`                // For tag requirements: which tag
	Resource *ResourceType      `json:"resource,omitempty" ts:"ResourceType | undefined"`      // For production: which resource
}

// CardBehavior represents any card behavior - both immediate (when played) and repeatable (activated by player)
// The trigger type determines when it executes: auto = immediate, manual = repeatable
type CardBehavior struct {
	Triggers []Trigger           `json:"triggers,omitempty" ts:"Trigger[] | undefined"`          // When/how this action is activated
	Inputs   []ResourceCondition `json:"inputs,omitempty" ts:"ResourceCondition[] | undefined"`  // Resources spent (input side of arrow)
	Outputs  []ResourceCondition `json:"outputs,omitempty" ts:"ResourceCondition[] | undefined"` // Resources gained (output side of arrow)
	Choices  []Choice            `json:"choices,omitempty" ts:"Choice[] | undefined"`            // Player choices between different input/output combinations
}

// DeepCopy creates a deep copy of the CardBehavior
func (cb CardBehavior) DeepCopy() CardBehavior {
	var result CardBehavior

	// Copy triggers slice
	if cb.Triggers != nil {
		result.Triggers = make([]Trigger, len(cb.Triggers))
		for i, trigger := range cb.Triggers {
			result.Triggers[i] = trigger // Trigger is a struct, copied by value
		}
	}

	// Copy inputs slice
	if cb.Inputs != nil {
		result.Inputs = make([]ResourceCondition, len(cb.Inputs))
		for i, input := range cb.Inputs {
			result.Inputs[i] = deepCopyResourceCondition(input)
		}
	}

	// Copy outputs slice
	if cb.Outputs != nil {
		result.Outputs = make([]ResourceCondition, len(cb.Outputs))
		for i, output := range cb.Outputs {
			result.Outputs[i] = deepCopyResourceCondition(output)
		}
	}

	// Copy choices slice
	if cb.Choices != nil {
		result.Choices = make([]Choice, len(cb.Choices))
		for i, choice := range cb.Choices {
			choiceCopy := Choice{}

			// Copy inputs for this choice
			if choice.Inputs != nil {
				choiceCopy.Inputs = make([]ResourceCondition, len(choice.Inputs))
				for j, input := range choice.Inputs {
					choiceCopy.Inputs[j] = deepCopyResourceCondition(input)
				}
			}

			// Copy outputs for this choice
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
	result := rc // Copy struct by value

	// Copy slices within the struct
	if rc.AffectedResources != nil {
		result.AffectedResources = make([]string, len(rc.AffectedResources))
		copy(result.AffectedResources, rc.AffectedResources)
	}

	if rc.AffectedTags != nil {
		result.AffectedTags = make([]CardTag, len(rc.AffectedTags))
		copy(result.AffectedTags, rc.AffectedTags)
	}

	if rc.AffectedStandardProjects != nil {
		result.AffectedStandardProjects = make([]StandardProject, len(rc.AffectedStandardProjects))
		copy(result.AffectedStandardProjects, rc.AffectedStandardProjects)
	}

	return result
}

// ResourceStorage represents a card's ability to hold resources
type ResourceStorage struct {
	Type     ResourceType `json:"type" ts:"ResourceType"`                     // Type of resource stored
	Capacity *int         `json:"capacity,omitempty" ts:"number | undefined"` // Max capacity (if limited)
	Starting int          `json:"starting" ts:"number"`                       // Starting amount
}

// VictoryPointCondition represents a VP condition like "1 VP per jovian tag"
type VictoryPointCondition struct {
	Amount     int             `json:"amount" ts:"number"`                           // VP awarded
	Condition  VPConditionType `json:"condition" ts:"VPConditionType"`               // Type of condition
	MaxTrigger *int            `json:"maxTrigger,omitempty" ts:"number | undefined"` // Max times it can trigger (-1 = unlimited), only for "per" conditions
	Per        *PerCondition   `json:"per,omitempty" ts:"PerCondition | undefined"`  // Per condition details, only for "per" conditions
}

// VPConditionType represents different types of VP conditions
type VPConditionType string

const (
	VPConditionPer   VPConditionType = "per"   // VP per resource/tag
	VPConditionOnce  VPConditionType = "once"  // VP awarded once when condition met
	VPConditionFixed VPConditionType = "fixed" // Fixed VP amount (no condition)
)

// CardApplyLocation represents different locations where card conditions can be evaluated
type CardApplyLocation string

const (
	// CardApplyLocationAnywhere represents no location restriction
	CardApplyLocationAnywhere CardApplyLocation = "anywhere"
	// CardApplyLocationMars represents on Mars only
	CardApplyLocationMars CardApplyLocation = "mars"
)

// DiscountEffect represents cost reductions for playing cards
type DiscountEffect struct {
	Amount      int       `json:"amount" ts:"number"`                        // M€ discount per qualifying tag
	Tags        []CardTag `json:"tags,omitempty" ts:"CardTag[] | undefined"` // Tags that qualify for discount (empty = all cards)
	Description string    `json:"description" ts:"string"`                   // Human readable description
}

// PaymentSubstitute is a type alias to avoid importing player package
type PaymentSubstitute = types.PaymentSubstitute

// PlayerEffect represents ongoing effects that a player has active, aligned with PlayerAction structure
type PlayerEffect struct {
	CardID        string       `json:"cardId" ts:"string"`         // ID of the card that provides this effect
	CardName      string       `json:"cardName" ts:"string"`       // Name of the card for display purposes
	BehaviorIndex int          `json:"behaviorIndex" ts:"number"`  // Which behavior on the card this effect represents
	Behavior      CardBehavior `json:"behavior" ts:"CardBehavior"` // The actual behavior definition with inputs/outputs
	// Note: No PlayCount since effects are ongoing, not per-generation like actions
}

// DeepCopy creates a deep copy of the PlayerEffect
func (pe *PlayerEffect) DeepCopy() *PlayerEffect {
	if pe == nil {
		return nil
	}

	// Deep copy the behavior
	var behaviorCopy CardBehavior

	// Copy triggers slice
	if pe.Behavior.Triggers != nil {
		behaviorCopy.Triggers = make([]Trigger, len(pe.Behavior.Triggers))
		for i, trigger := range pe.Behavior.Triggers {
			behaviorCopy.Triggers[i] = trigger // Trigger is a struct, so it's copied by value
		}
	}

	// Copy inputs slice
	if pe.Behavior.Inputs != nil {
		behaviorCopy.Inputs = make([]ResourceCondition, len(pe.Behavior.Inputs))
		for i, input := range pe.Behavior.Inputs {
			// Deep copy the resource condition
			inputCopy := input // Copy struct by value

			// Copy slices within the struct
			if input.AffectedResources != nil {
				inputCopy.AffectedResources = make([]string, len(input.AffectedResources))
				copy(inputCopy.AffectedResources, input.AffectedResources)
			}

			if input.AffectedTags != nil {
				inputCopy.AffectedTags = make([]CardTag, len(input.AffectedTags))
				copy(inputCopy.AffectedTags, input.AffectedTags)
			}

			behaviorCopy.Inputs[i] = inputCopy
		}
	}

	// Copy outputs slice
	if pe.Behavior.Outputs != nil {
		behaviorCopy.Outputs = make([]ResourceCondition, len(pe.Behavior.Outputs))
		for i, output := range pe.Behavior.Outputs {
			// Deep copy the resource condition
			outputCopy := output // Copy struct by value

			// Copy slices within the struct
			if output.AffectedResources != nil {
				outputCopy.AffectedResources = make([]string, len(output.AffectedResources))
				copy(outputCopy.AffectedResources, output.AffectedResources)
			}

			if output.AffectedTags != nil {
				outputCopy.AffectedTags = make([]CardTag, len(output.AffectedTags))
				copy(outputCopy.AffectedTags, output.AffectedTags)
			}

			behaviorCopy.Outputs[i] = outputCopy
		}
	}

	// Copy choices slice
	if pe.Behavior.Choices != nil {
		behaviorCopy.Choices = make([]Choice, len(pe.Behavior.Choices))
		for i, choice := range pe.Behavior.Choices {
			choiceCopy := Choice{}

			// Copy inputs for this choice
			if choice.Inputs != nil {
				choiceCopy.Inputs = make([]ResourceCondition, len(choice.Inputs))
				for j, input := range choice.Inputs {
					inputCopy := input

					if input.AffectedResources != nil {
						inputCopy.AffectedResources = make([]string, len(input.AffectedResources))
						copy(inputCopy.AffectedResources, input.AffectedResources)
					}

					if input.AffectedTags != nil {
						inputCopy.AffectedTags = make([]CardTag, len(input.AffectedTags))
						copy(inputCopy.AffectedTags, input.AffectedTags)
					}

					choiceCopy.Inputs[j] = inputCopy
				}
			}

			// Copy outputs for this choice
			if choice.Outputs != nil {
				choiceCopy.Outputs = make([]ResourceCondition, len(choice.Outputs))
				for j, output := range choice.Outputs {
					outputCopy := output

					if output.AffectedResources != nil {
						outputCopy.AffectedResources = make([]string, len(output.AffectedResources))
						copy(outputCopy.AffectedResources, output.AffectedResources)
					}

					if output.AffectedTags != nil {
						outputCopy.AffectedTags = make([]CardTag, len(output.AffectedTags))
						copy(outputCopy.AffectedTags, output.AffectedTags)
					}

					choiceCopy.Outputs[j] = outputCopy
				}
			}

			behaviorCopy.Choices[i] = choiceCopy
		}
	}

	return &PlayerEffect{
		CardID:        pe.CardID,
		CardName:      pe.CardName,
		BehaviorIndex: pe.BehaviorIndex,
		Behavior:      behaviorCopy,
	}
}

// TargetType represents different targeting scopes for resource conditions
type TargetType string

const (
	TargetSelfPlayer TargetType = "self-player" // Targets the player who played/owns the card
	TargetSelfCard   TargetType = "self-card"   // Targets the specific card itself
	TargetAnyCard    TargetType = "any-card"    // Targets any card with matching resource storage
	TargetAnyPlayer  TargetType = "any-player"  // Can target any player
	TargetOpponent   TargetType = "opponent"    // Targets opponent players
	TargetNone       TargetType = "none"        // No target (e.g., global parameter changes)
)

// PerCondition represents what to count for conditional resource gains
type PerCondition struct {
	Type     ResourceType       `json:"type" ts:"ResourceType"`                                // What to count (city-tile, ocean-tile, etc.)
	Amount   int                `json:"amount" ts:"number"`                                    // How many of the counted thing per gain
	Location *CardApplyLocation `json:"location,omitempty" ts:"CardApplyLocation | undefined"` // Location constraint (Mars, anywhere, etc.)
	Target   *TargetType        `json:"target,omitempty" ts:"TargetType | undefined"`          // Whose tags/resources to count (self-player, any-player, etc.)
	Tag      *CardTag           `json:"tag,omitempty" ts:"CardTag | undefined"`                // For tag-based VP conditions (jovian tag, science tag, etc.)
}

// Choice represents a single choice option with inputs and outputs
type Choice struct {
	Inputs  []ResourceCondition `json:"inputs,omitempty" ts:"ResourceCondition[] | undefined"`  // Resources spent for this choice
	Outputs []ResourceCondition `json:"outputs,omitempty" ts:"ResourceCondition[] | undefined"` // Resources gained from this choice
}

// ResourceCondition represents a resource amount (input or output)
type ResourceCondition struct {
	Type                     ResourceType      `json:"type" ts:"ResourceType"`                                                // Type of resource
	Amount                   int               `json:"amount" ts:"number"`                                                    // Amount of resource
	Target                   TargetType        `json:"target" ts:"TargetType"`                                                // Target for this resource condition
	AffectedResources        []string          `json:"affectedResources,omitempty" ts:"string[] | undefined"`                 // For defense: resources being protected
	AffectedTags             []CardTag         `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"`                     // For discount: tags qualifying for discount
	AffectedStandardProjects []StandardProject `json:"affectedStandardProjects,omitempty" ts:"StandardProject[] | undefined"` // For discount: standard projects affected
	MaxTrigger               *int              `json:"maxTrigger,omitempty" ts:"number | undefined"`                          // Max times it can trigger (-1 = unlimited), only for "per" conditions
	Per                      *PerCondition     `json:"per,omitempty" ts:"PerCondition | undefined"`                           // For conditional gains: what to count
}

// ResourceTriggerType represents different trigger types for resource exchanges
type ResourceTriggerType string

const (
	ResourceTriggerManual          ResourceTriggerType = "manual"            // Manual activation (actions)
	ResourceTriggerAuto            ResourceTriggerType = "auto"              // Automatic activation (effects, immediate)
	ResourceTriggerAutoFirstAction ResourceTriggerType = "auto-first-action" // Automatic forced first action (corporations only)
)

// MinMaxValue represents a minimum and/or maximum value constraint
type MinMaxValue struct {
	Min *int `json:"min,omitempty" ts:"number | undefined"` // Minimum value (e.g., at least 20)
	Max *int `json:"max,omitempty" ts:"number | undefined"` // Maximum value (e.g., at most 10)
}

// ResourceTriggerCondition represents what triggers an automatic resource exchange
type ResourceTriggerCondition struct {
	Type                   TriggerType                  `json:"type" ts:"TriggerType"`                                                               // What triggers this (onCityPlaced, etc.)
	Location               *CardApplyLocation           `json:"location,omitempty" ts:"CardApplyLocation | undefined"`                               // Where the trigger applies (mars, anywhere)
	AffectedTags           []CardTag                    `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"`                                   // Tags that trigger this effect
	AffectedResources      []string                     `json:"affectedResources,omitempty" ts:"string[] | undefined"`                               // Resource types that trigger this effect (for placement-bonus-gained)
	AffectedCardTypes      []CardType                   `json:"affectedCardTypes,omitempty" ts:"CardType[] | undefined"`                             // Card types that trigger this effect (for card-played triggers: event, automated, active, etc.)
	Target                 *TargetType                  `json:"target,omitempty" ts:"TargetType | undefined"`                                        // Whose actions trigger this (self-player, any-player, etc.)
	RequiredOriginalCost   *MinMaxValue                 `json:"requiredOriginalCost,omitempty" ts:"MinMaxValue | undefined"`                         // Original credit cost requirement (only for card-played/standard-project-played triggers)
	RequiredResourceChange map[ResourceType]MinMaxValue `json:"requiredResourceChange,omitempty" ts:"Record<ResourceType, MinMaxValue> | undefined"` // Min/max requirements for actual resources spent
}

// Trigger represents when and how an action or effect is activated
type Trigger struct {
	Type      ResourceTriggerType       `json:"type" ts:"ResourceTriggerType"`                                 // Manual or auto activation
	Condition *ResourceTriggerCondition `json:"condition,omitempty" ts:"ResourceTriggerCondition | undefined"` // What triggers auto actions
}

// ResourceExchange represents a directional resource trade (input → output)
type ResourceExchange struct {
	Triggers []Trigger           `json:"triggers,omitempty" ts:"Trigger[] | undefined"`          // When/how this exchange is activated
	Inputs   []ResourceCondition `json:"inputs,omitempty" ts:"ResourceCondition[] | undefined"`  // Resources spent (input side of arrow)
	Outputs  []ResourceCondition `json:"outputs,omitempty" ts:"ResourceCondition[] | undefined"` // Resources gained (output side of arrow)
}

// EffectContext provides context about a game event that triggered passive effects
// This allows effects to access information about what triggered them
type EffectContext struct {
	TriggeringPlayerID string        `json:"triggeringPlayerId" ts:"string"`         // Player who caused the event
	TileCoordinate     *HexPosition  `json:"tileCoordinate" ts:"HexPosition | null"` // Coordinate for tile placement events
	CardID             *string       `json:"cardId" ts:"string | null"`              // Card ID for card-played events
	TagType            *CardTag      `json:"tagType" ts:"CardTag | null"`            // Tag type for tag-played events
	TileType           *ResourceType `json:"tileType" ts:"ResourceType | null"`      // Type of tile placed (city, ocean, greenery)
	ParameterChange    *int          `json:"parameterChange" ts:"number | null"`     // Amount of parameter change (temperature, oxygen)
}

// CardSelection represents the card selection phase data
type CardSelection struct {
	PlayerCardOptions            []PlayerCardOptions `json:"playerCardOptions" ts:"PlayerCardOptions[]"`
	PlayersWhoCompletedSelection []string            `json:"playersWhoCompletedSelection" ts:"string[]"`
}

// PlayerCardOptions represents the card options for a specific player
type PlayerCardOptions struct {
	PlayerID    string   `json:"playerId" ts:"string"`
	CardOptions []string `json:"cardOptions" ts:"string[]"`
}

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
