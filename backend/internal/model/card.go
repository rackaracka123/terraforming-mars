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
