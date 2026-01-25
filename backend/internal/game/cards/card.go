package cards

import (
	"terraforming-mars-backend/internal/game/shared"
)

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
	Behaviors       []shared.CardBehavior   `json:"behaviors"`
	ResourceStorage *ResourceStorage        `json:"resourceStorage"`
	VPConditions    []VictoryPointCondition `json:"vpConditions"`

	StartingResources  *shared.ResourceSet `json:"startingResources"`
	StartingProduction *shared.ResourceSet `json:"startingProduction"`
}

// DeepCopy creates a deep copy of the Card
func (c Card) DeepCopy() Card {
	tags := make([]shared.CardTag, len(c.Tags))
	copy(tags, c.Tags)

	requirements := make([]Requirement, len(c.Requirements))
	copy(requirements, c.Requirements)

	behaviors := make([]shared.CardBehavior, len(c.Behaviors))
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
