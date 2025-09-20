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
	Tags            []CardTag               `json:"tags,omitempty" ts:"CardTag[] | undefined"`
	Requirements    []Requirement           `json:"requirements,omitempty" ts:"Requirement[] | undefined"`
	Behaviors       []CardBehavior          `json:"behaviors,omitempty" ts:"CardBehavior[] | undefined"`             // All card behaviors (immediate and repeatable)
	ResourceStorage *ResourceStorage        `json:"resourceStorage,omitempty" ts:"ResourceStorage | undefined"`      // Cards that can hold resources
	VPConditions    []VictoryPointCondition `json:"vpConditions,omitempty" ts:"VictoryPointCondition[] | undefined"` // VP per X conditions
}

// DeepCopy creates a deep copy of the Card struct
func (c *Card) DeepCopy() *Card {
	if c == nil {
		return nil
	}

	newCard := &Card{
		ID:          c.ID,
		Name:        c.Name,
		Type:        c.Type,
		Cost:        c.Cost,
		Description: c.Description,
	}

	// Copy tags slice
	if c.Tags != nil {
		newCard.Tags = make([]CardTag, len(c.Tags))
		copy(newCard.Tags, c.Tags)
	}

	// Copy requirements slice (deep copy each requirement)
	if c.Requirements != nil {
		newCard.Requirements = make([]Requirement, len(c.Requirements))
		copy(newCard.Requirements, c.Requirements)
	}

	// Copy behaviors slice (deep copy each behavior)
	if c.Behaviors != nil {
		newCard.Behaviors = make([]CardBehavior, len(c.Behaviors))
		copy(newCard.Behaviors, c.Behaviors)
	}

	// Copy resource storage pointer
	if c.ResourceStorage != nil {
		newStorage := *c.ResourceStorage
		newCard.ResourceStorage = &newStorage
	}

	// Copy VP conditions slice (deep copy each condition)
	if c.VPConditions != nil {
		newCard.VPConditions = make([]VictoryPointCondition, len(c.VPConditions))
		copy(newCard.VPConditions, c.VPConditions)
	}

	return newCard
}
