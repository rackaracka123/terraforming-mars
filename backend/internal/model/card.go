package model

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
