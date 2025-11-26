package card

import (
	"terraforming-mars-backend/internal/session/types"
)

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
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// Card represents a game card
type Card struct {
	ID              string
	Name            string
	Type            CardType
	Cost            int
	Description     string
	Pack            string // Card pack identifier (e.g., "base-game", "corporate-era", "prelude")
	Tags            []types.CardTag
	Requirements    []Requirement           // Within card package
	Behaviors       []CardBehavior          // Within card package
	ResourceStorage *ResourceStorage        // Within card package
	VPConditions    []VictoryPointCondition // Within card package

	// Corporation-specific fields (nil for non-corporation cards)
	StartingResources  *types.ResourceSet // Parsed from first auto behavior (corporations only)
	StartingProduction *types.ResourceSet // Parsed from first auto behavior (corporations only)
}

// DeepCopy creates a deep copy of the Card
func (c Card) DeepCopy() Card {
	// Copy slices
	tags := make([]types.CardTag, len(c.Tags))
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
	var startingResources *types.ResourceSet
	if c.StartingResources != nil {
		rs := *c.StartingResources
		startingResources = &rs
	}

	var startingProduction *types.ResourceSet
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
