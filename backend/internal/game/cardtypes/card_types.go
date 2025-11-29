package cardtypes

import (
	"fmt"

	"terraforming-mars-backend/internal/game/shared"
)

// ==================== Card Type ====================

// CardType represents different types of cards in Terraforming Mars
type CardType string

const (
	CardTypeAutomated   CardType = "automated"   // Green cards - immediate effects, production bonuses
	CardTypeActive      CardType = "active"      // Blue cards - ongoing effects, repeatable actions
	CardTypeEvent       CardType = "event"       // Red cards - one-time effects
	CardTypeCorporation CardType = "corporation" // Corporation cards - unique player abilities
	CardTypePrelude     CardType = "prelude"     // Prelude cards - setup phase cards
)

// Card represents a game card
type Card struct {
	ID              string                  `json:"id"`
	Name            string                  `json:"name"`
	Type            CardType                `json:"type"`
	Cost            int                     `json:"cost"`
	Description     string                  `json:"description"`
	Pack            string                  `json:"pack"`
	Tags            []shared.CardTag        `json:"tags"`
	Requirements    []Requirement           `json:"requirements"`
	Behaviors       []CardBehavior          `json:"behaviors"`
	ResourceStorage *ResourceStorage        `json:"resourceStorage"`
	VPConditions    []VictoryPointCondition `json:"vpConditions"`

	// Corporation-specific fields (nil for non-corporation cards)
	StartingResources  *shared.ResourceSet `json:"startingResources"`
	StartingProduction *shared.ResourceSet `json:"startingProduction"`
}

// DeepCopy creates a deep copy of the Card
func (c Card) DeepCopy() Card {
	tags := make([]shared.CardTag, len(c.Tags))
	copy(tags, c.Tags)

	requirements := make([]Requirement, len(c.Requirements))
	copy(requirements, c.Requirements)

	behaviors := make([]CardBehavior, len(c.Behaviors))
	for i, behavior := range c.Behaviors {
		behaviors[i] = behavior.DeepCopy()
	}

	vpConditions := make([]VictoryPointCondition, len(c.VPConditions))
	copy(vpConditions, c.VPConditions)

	var resourceStorage *ResourceStorage
	if c.ResourceStorage != nil {
		rs := *c.ResourceStorage
		resourceStorage = &rs
	}

	var startingResources *shared.ResourceSet
	if c.StartingResources != nil {
		rs := *c.StartingResources
		startingResources = &rs
	}

	var startingProduction *shared.ResourceSet
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

// ==================== Card Requirements ====================

// RequirementType represents different types of card requirements
type RequirementType string

const (
	RequirementTemperature RequirementType = "temperature"
	RequirementOxygen      RequirementType = "oxygen"
	RequirementOceans      RequirementType = "oceans"
	RequirementVenus       RequirementType = "venus"
	RequirementCities      RequirementType = "cities"
	RequirementGreeneries  RequirementType = "greeneries"
	RequirementTags        RequirementType = "tags"
	RequirementProduction  RequirementType = "production"
	RequirementTR          RequirementType = "tr"
	RequirementResource    RequirementType = "resource"
)

// Requirement represents a single card requirement
type Requirement struct {
	Type     RequirementType      `json:"type"`
	Min      *int                 `json:"min,omitempty"`
	Max      *int                 `json:"max,omitempty"`
	Location *CardApplyLocation   `json:"location,omitempty"`
	Tag      *shared.CardTag      `json:"tag,omitempty"`
	Resource *shared.ResourceType `json:"resource,omitempty"`
}

// CardApplyLocation represents different locations
type CardApplyLocation string

const (
	CardApplyLocationAnywhere CardApplyLocation = "anywhere"
	CardApplyLocationMars     CardApplyLocation = "mars"
)

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
		result.AffectedTags = make([]shared.CardTag, len(rc.AffectedTags))
		copy(result.AffectedTags, rc.AffectedTags)
	}

	if rc.AffectedStandardProjects != nil {
		result.AffectedStandardProjects = make([]shared.StandardProject, len(rc.AffectedStandardProjects))
		copy(result.AffectedStandardProjects, rc.AffectedStandardProjects)
	}

	return result
}

// Choice represents a player choice option
type Choice struct {
	Inputs  []ResourceCondition `json:"inputs,omitempty"`
	Outputs []ResourceCondition `json:"outputs,omitempty"`
}

// ==================== Triggers ====================

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
	TriggerCardHandUpdated       TriggerType = "card-hand-updated"
	TriggerPlayerEffectsChanged  TriggerType = "player-effects-changed"
)

// ResourceTriggerType represents trigger types for resource exchanges
type ResourceTriggerType string

const (
	ResourceTriggerManual                     ResourceTriggerType = "manual"
	ResourceTriggerAuto                       ResourceTriggerType = "auto"
	ResourceTriggerAutoCorporationFirstAction ResourceTriggerType = "auto-corporation-first-action"
	ResourceTriggerAutoCorporationStart       ResourceTriggerType = "auto-corporation-start"
)

// Trigger represents when and how an action or effect is activated
type Trigger struct {
	Type      ResourceTriggerType       `json:"type"`
	Condition *ResourceTriggerCondition `json:"condition,omitempty"`
}

// MinMaxValue represents a min/max value constraint
type MinMaxValue struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

// ResourceTriggerCondition represents what triggers an automatic resource exchange
type ResourceTriggerCondition struct {
	Type                   TriggerType                         `json:"type"`
	Location               *CardApplyLocation                  `json:"location,omitempty"`
	AffectedTags           []shared.CardTag                    `json:"affectedTags,omitempty"`
	AffectedResources      []string                            `json:"affectedResources,omitempty"`
	AffectedCardTypes      []CardType                          `json:"affectedCardTypes,omitempty"`
	Target                 *TargetType                         `json:"target,omitempty"`
	RequiredOriginalCost   *MinMaxValue                        `json:"requiredOriginalCost,omitempty"`
	RequiredResourceChange map[shared.ResourceType]MinMaxValue `json:"requiredResourceChange,omitempty"`
}

// ==================== Resource Conditions ====================

// TargetType represents different targeting scopes
type TargetType string

const (
	TargetSelfPlayer TargetType = "self-player"
	TargetSelfCard   TargetType = "self-card"
	TargetAnyCard    TargetType = "any-card"
	TargetAnyPlayer  TargetType = "any-player"
	TargetOpponent   TargetType = "opponent"
	TargetNone       TargetType = "none"
)

// ResourceCondition represents a resource amount (input or output)
type ResourceCondition struct {
	Type                     shared.ResourceType      `json:"type"`
	Amount                   int                      `json:"amount"`
	Target                   TargetType               `json:"target"`
	AffectedResources        []string                 `json:"affectedResources,omitempty"`
	AffectedTags             []shared.CardTag         `json:"affectedTags,omitempty"`
	AffectedCardTypes        []CardType               `json:"affectedCardTypes,omitempty"`
	AffectedStandardProjects []shared.StandardProject `json:"affectedStandardProjects,omitempty"`
	MaxTrigger               *int                     `json:"maxTrigger,omitempty"`
	Per                      *PerCondition            `json:"per,omitempty"`
}

// PerCondition represents what to count for conditional resource gains
type PerCondition struct {
	Type     shared.ResourceType `json:"type"`
	Amount   int                 `json:"amount"`
	Location *CardApplyLocation  `json:"location,omitempty"`
	Target   *TargetType         `json:"target,omitempty"`
	Tag      *shared.CardTag     `json:"tag,omitempty"`
}

// ==================== Card Storage and VP ====================

// ResourceStorage represents a card's ability to hold resources
type ResourceStorage struct {
	Type     shared.ResourceType `json:"type"`
	Capacity *int                `json:"capacity,omitempty"`
	Starting int                 `json:"starting"`
}

// VictoryPointCondition represents a VP condition
type VictoryPointCondition struct {
	Amount     int             `json:"amount"`
	Condition  VPConditionType `json:"condition"`
	MaxTrigger *int            `json:"maxTrigger,omitempty"`
	Per        *PerCondition   `json:"per,omitempty"`
}

// VPConditionType represents different types of VP conditions
type VPConditionType string

const (
	VPConditionPer   VPConditionType = "per"
	VPConditionOnce  VPConditionType = "once"
	VPConditionFixed VPConditionType = "fixed"
)

// ==================== Card Effects and Actions ====================

// CardEffect represents an ongoing effect defined by a card
type CardEffect struct {
	CardID        string
	CardName      string
	BehaviorIndex int
	Behavior      CardBehavior
}

// DeepCopy creates a deep copy of the CardEffect
func (pe *CardEffect) DeepCopy() *CardEffect {
	if pe == nil {
		return nil
	}

	return &CardEffect{
		CardID:        pe.CardID,
		CardName:      pe.CardName,
		BehaviorIndex: pe.BehaviorIndex,
		Behavior:      pe.Behavior.DeepCopy(),
	}
}

// CardAction represents a repeatable manual action defined by a card
type CardAction struct {
	CardID        string
	CardName      string
	BehaviorIndex int
	Behavior      CardBehavior
	PlayCount     int
}

// DeepCopy creates a deep copy of the CardAction
func (pa *CardAction) DeepCopy() *CardAction {
	if pa == nil {
		return nil
	}

	return &CardAction{
		CardID:        pa.CardID,
		CardName:      pa.CardName,
		BehaviorIndex: pa.BehaviorIndex,
		Behavior:      pa.Behavior.DeepCopy(),
		PlayCount:     pa.PlayCount,
	}
}

// ==================== Card Payment ====================

// shared.PaymentSubstitute represents an alternative resource for payment

// CardPayment represents how a player is paying for a card
type CardPayment struct {
	Credits     int
	Steel       int
	Titanium    int
	Substitutes map[shared.ResourceType]int
}

// Payment method constants
const (
	SteelValue    = 2
	TitaniumValue = 3
)

// TotalValue calculates the total MC value of this payment
func (p CardPayment) TotalValue(playerSubstitutes []shared.PaymentSubstitute) int {
	total := p.Credits + (p.Steel * SteelValue) + (p.Titanium * TitaniumValue)

	if p.Substitutes != nil {
		for resourceType, amount := range p.Substitutes {
			for _, sub := range playerSubstitutes {
				if sub.ResourceType == resourceType {
					total += amount * sub.ConversionRate
					break
				}
			}
		}
	}

	return total
}

// Validate checks if the payment is valid
func (p CardPayment) Validate() error {
	if p.Credits < 0 {
		return fmt.Errorf("payment credits cannot be negative: %d", p.Credits)
	}
	if p.Steel < 0 {
		return fmt.Errorf("payment steel cannot be negative: %d", p.Steel)
	}
	if p.Titanium < 0 {
		return fmt.Errorf("payment titanium cannot be negative: %d", p.Titanium)
	}

	if p.Substitutes != nil {
		for resourceType, amount := range p.Substitutes {
			if amount < 0 {
				return fmt.Errorf("payment substitute %s cannot be negative: %d", resourceType, amount)
			}
		}
	}

	return nil
}

// CanAfford checks if a player has sufficient resources for this payment
func (p CardPayment) CanAfford(playerResources shared.Resources) error {
	if playerResources.Credits < p.Credits {
		return fmt.Errorf("insufficient credits: need %d, have %d", p.Credits, playerResources.Credits)
	}
	if playerResources.Steel < p.Steel {
		return fmt.Errorf("insufficient steel: need %d, have %d", p.Steel, playerResources.Steel)
	}
	if playerResources.Titanium < p.Titanium {
		return fmt.Errorf("insufficient titanium: need %d, have %d", p.Titanium, playerResources.Titanium)
	}

	if p.Substitutes != nil {
		for resourceType, amount := range p.Substitutes {
			var available int
			switch resourceType {
			case shared.ResourceHeat:
				available = playerResources.Heat
			case shared.ResourceEnergy:
				available = playerResources.Energy
			case shared.ResourcePlants:
				available = playerResources.Plants
			default:
				return fmt.Errorf("unsupported payment substitute resource type: %s", resourceType)
			}

			if available < amount {
				return fmt.Errorf("insufficient %s: need %d, have %d", resourceType, amount, available)
			}
		}
	}

	return nil
}

// CoversCardCost checks if this payment covers the card cost
func (p CardPayment) CoversCardCost(cardCost int, allowSteel, allowTitanium bool, playerSubstitutes []shared.PaymentSubstitute) error {
	if err := p.Validate(); err != nil {
		return err
	}

	if p.Steel > 0 && !allowSteel {
		return fmt.Errorf("card does not have building tag, cannot use steel")
	}
	if p.Titanium > 0 && !allowTitanium {
		return fmt.Errorf("card does not have space tag, cannot use titanium")
	}

	if p.Substitutes != nil {
		for resourceType := range p.Substitutes {
			found := false
			for _, sub := range playerSubstitutes {
				if sub.ResourceType == resourceType {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("player cannot use %s as payment substitute", resourceType)
			}
		}
	}

	totalValue := p.TotalValue(playerSubstitutes)
	if totalValue < cardCost {
		return fmt.Errorf("payment insufficient: card costs %d MC, payment provides %d MC", cardCost, totalValue)
	}

	return nil
}

// ==================== Other Card Types ====================

// DiscountEffect represents cost reductions for playing cards
type DiscountEffect struct {
	Amount      int
	Tags        []shared.CardTag
	Description string
}

// shared.RequirementModifier represents a discount or lenience that modifies requirements

// ProductionEffects represents changes to resource production
type ProductionEffects struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// TerraformingActions represents tile placement actions
type TerraformingActions struct {
	CityPlacement     int
	OceanPlacement    int
	GreeneryPlacement int
}

// EffectContext provides context about a game event that triggered passive effects
type EffectContext struct {
	TriggeringPlayerID string
	TileCoordinate     *shared.HexPosition
	CardID             *string
	TagType            *shared.CardTag
	TileType           *shared.ResourceType
	ParameterChange    *int
}
