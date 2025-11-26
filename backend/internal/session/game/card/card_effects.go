package card

import (
	"terraforming-mars-backend/internal/session/types"
)

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
	TriggerCardHandUpdated       TriggerType = "card-hand-updated"      // When player's card hand changes (cards added/removed)
	TriggerPlayerEffectsChanged  TriggerType = "player-effects-changed" // When player's effects list changes
)

// TerraformingActions represents tile placement actions
type TerraformingActions struct {
	CityPlacement     int // Number of city tiles to place
	OceanPlacement    int // Number of ocean tiles to place
	GreeneryPlacement int // Number of greenery tiles to place
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
	Type     RequirementType    // Type of requirement
	Min      *int               // Minimum value required
	Max      *int               // Maximum value allowed
	Location *CardApplyLocation // Location constraint (Mars, anywhere, etc.)
	Tag      *types.CardTag           // For tag requirements: which tag
	Resource *types.ResourceType      // For production: which resource
}

// CardBehavior represents any card behavior - both immediate (when played) and repeatable (activated by player)
// The trigger type determines when it executes: auto = immediate, manual = repeatable
type CardBehavior struct {
	Triggers []Trigger           // When/how this action is activated
	Inputs   []ResourceCondition // Resources spent (input side of arrow)
	Outputs  []ResourceCondition // Resources gained (output side of arrow)
	Choices  []Choice            // Player choices between different input/output combinations
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
		result.AffectedTags = make([]types.CardTag, len(rc.AffectedTags))
		copy(result.AffectedTags, rc.AffectedTags)
	}

	if rc.AffectedStandardProjects != nil {
		result.AffectedStandardProjects = make([]types.StandardProject, len(rc.AffectedStandardProjects))
		copy(result.AffectedStandardProjects, rc.AffectedStandardProjects)
	}

	return result
}

// ResourceStorage represents a card's ability to hold resources
type ResourceStorage struct {
	Type     types.ResourceType // Type of resource stored
	Capacity *int         // Max capacity (if limited)
	Starting int          // Starting amount
}

// VictoryPointCondition represents a VP condition like "1 VP per jovian tag"
type VictoryPointCondition struct {
	Amount     int             // VP awarded
	Condition  VPConditionType // Type of condition
	MaxTrigger *int            // Max times it can trigger (-1 = unlimited), only for "per" conditions
	Per        *PerCondition   // Per condition details, only for "per" conditions
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
	Amount      int       // M€ discount per qualifying tag
	Tags        []types.CardTag // Tags that qualify for discount (empty = all cards)
	Description string    // Human readable description
}

// PaymentSubstitute represents an alternative resource that can be used as payment for credits
// Example: Helion allows using heat as M€ with 1:1 conversion
type PaymentSubstitute struct {
	ResourceType   types.ResourceType // The resource that can be used (e.g., heat)
	ConversionRate int                // How many credits each resource is worth (1 = 1:1)
}

// PlayerEffect represents ongoing effects that a player has active, aligned with PlayerAction structure
type PlayerEffect struct {
	CardID        string       // ID of the card that provides this effect
	CardName      string       // Name of the card for display purposes
	BehaviorIndex int          // Which behavior on the card this effect represents
	Behavior      CardBehavior // The actual behavior definition with inputs/outputs
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
				inputCopy.AffectedTags = make([]types.CardTag, len(input.AffectedTags))
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
				outputCopy.AffectedTags = make([]types.CardTag, len(output.AffectedTags))
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
						inputCopy.AffectedTags = make([]types.CardTag, len(input.AffectedTags))
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
						outputCopy.AffectedTags = make([]types.CardTag, len(output.AffectedTags))
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
	Type     types.ResourceType       // What to count (city-tile, ocean-tile, etc.)
	Amount   int                // How many of the counted thing per gain
	Location *CardApplyLocation // Location constraint (Mars, anywhere, etc.)
	Target   *TargetType        // Whose tags/resources to count (self-player, any-player, etc.)
	Tag      *types.CardTag           // For tag-based VP conditions (jovian tag, science tag, etc.)
}

// Choice represents a single choice option with inputs and outputs
type Choice struct {
	Inputs  []ResourceCondition // Resources spent for this choice
	Outputs []ResourceCondition // Resources gained from this choice
}

// ResourceCondition represents a resource amount (input or output)
type ResourceCondition struct {
	Type                     types.ResourceType      // Type of resource
	Amount                   int               // Amount of resource
	Target                   TargetType        // Target for this resource condition
	AffectedResources        []string          // For defense: resources being protected
	AffectedTags             []types.CardTag         // For discount: tags qualifying for discount
	AffectedCardTypes        []CardType        // For discount/effects: card types qualifying
	AffectedStandardProjects []types.StandardProject // For discount: standard projects affected
	MaxTrigger               *int              // Max times it can trigger (-1 = unlimited), only for "per" conditions
	Per                      *PerCondition     // For conditional gains: what to count
}

// ResourceTriggerType represents different trigger types for resource exchanges
type ResourceTriggerType string

const (
	ResourceTriggerManual                     ResourceTriggerType = "manual"                        // Manual activation (actions)
	ResourceTriggerAuto                       ResourceTriggerType = "auto"                          // Automatic activation (effects, immediate)
	ResourceTriggerAutoCorporationFirstAction ResourceTriggerType = "auto-corporation-first-action" // Automatic forced first action (corporations only)
	ResourceTriggerAutoCorporationStart       ResourceTriggerType = "auto-corporation-start"        // Starting bonuses for corporations (not an effect)
)

// MinMaxValue represents a minimum and/or maximum value constraint
type MinMaxValue struct {
	Min *int // Minimum value (e.g., at least 20)
	Max *int // Maximum value (e.g., at most 10)
}

// ResourceTriggerCondition represents what triggers an automatic resource exchange
type ResourceTriggerCondition struct {
	Type                   TriggerType                  // What triggers this (onCityPlaced, etc.)
	Location               *CardApplyLocation           // Where the trigger applies (mars, anywhere)
	AffectedTags           []types.CardTag                    // Tags that trigger this effect
	AffectedResources      []string                     // Resource types that trigger this effect (for placement-bonus-gained)
	AffectedCardTypes      []CardType                   // Card types that trigger this effect (for card-played triggers: event, automated, active, etc.)
	Target                 *TargetType                  // Whose actions trigger this (self-player, any-player, etc.)
	RequiredOriginalCost   *MinMaxValue                 // Original credit cost requirement (only for card-played/standard-project-played triggers)
	RequiredResourceChange map[types.ResourceType]MinMaxValue // Min/max requirements for actual resources spent
}

// Trigger represents when and how an action or effect is activated
type Trigger struct {
	Type      ResourceTriggerType       // Manual or auto activation
	Condition *ResourceTriggerCondition // What triggers auto actions
}

// ResourceExchange represents a directional resource trade (input → output)
type ResourceExchange struct {
	Triggers []Trigger           // When/how this exchange is activated
	Inputs   []ResourceCondition // Resources spent (input side of arrow)
	Outputs  []ResourceCondition // Resources gained (output side of arrow)
}

// EffectContext provides context about a game event that triggered passive effects
// This allows effects to access information about what triggered them
type EffectContext struct {
	TriggeringPlayerID string              // Player who caused the event
	TileCoordinate     *types.HexPosition  // Coordinate for tile placement events
	CardID             *string             // Card ID for card-played events
	TagType            *types.CardTag      // Tag type for tag-played events
	TileType           *types.ResourceType // Type of tile placed (city, ocean, greenery)
	ParameterChange    *int          // Amount of parameter change (temperature, oxygen)
}
