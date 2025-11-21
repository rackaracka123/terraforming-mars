package card

import "terraforming-mars-backend/internal/model"

// Card represents a game card with complete data for game logic
type Card struct {
	ID              string                        `json:"id"`
	Name            string                        `json:"name"`
	Type            string                        `json:"type"` // "project", "corporation", "prelude"
	Cost            int                           `json:"cost"`
	Description     string                        `json:"description"`
	Pack            string                        `json:"pack"` // "base-game", "future", etc.
	Tags            []model.CardTag               `json:"tags,omitempty"`
	Requirements    []model.Requirement           `json:"requirements,omitempty"`
	Behaviors       []model.CardBehavior          `json:"behaviors,omitempty"`
	ResourceStorage *model.ResourceStorage        `json:"resourceStorage,omitempty"`
	VPConditions    []model.VictoryPointCondition `json:"vpConditions,omitempty"`

	// Corporation-specific fields (nil for non-corporation cards)
	StartingResources  *model.ResourceSet `json:"startingResources,omitempty"`
	StartingProduction *model.ResourceSet `json:"startingProduction,omitempty"`
}

// FromModelCard converts a model.Card to card subdomain Card
func FromModelCard(mc model.Card) Card {
	return Card{
		ID:                 mc.ID,
		Name:               mc.Name,
		Type:               string(mc.Type),
		Cost:               mc.Cost,
		Description:        mc.Description,
		Pack:               mc.Pack,
		Tags:               mc.Tags,
		Requirements:       mc.Requirements,
		Behaviors:          mc.Behaviors,
		ResourceStorage:    mc.ResourceStorage,
		VPConditions:       mc.VPConditions,
		StartingResources:  mc.StartingResources,
		StartingProduction: mc.StartingProduction,
	}
}

// FromModelCards converts multiple model.Card to card subdomain Cards
func FromModelCards(mcs []model.Card) []Card {
	cards := make([]Card, len(mcs))
	for i, mc := range mcs {
		cards[i] = FromModelCard(mc)
	}
	return cards
}

// ToModelCard converts a card subdomain Card to model.Card
func (c *Card) ToModelCard() *model.Card {
	return &model.Card{
		ID:                 c.ID,
		Name:               c.Name,
		Type:               model.CardType(c.Type),
		Cost:               c.Cost,
		Description:        c.Description,
		Pack:               c.Pack,
		Tags:               c.Tags,
		Requirements:       c.Requirements,
		Behaviors:          c.Behaviors,
		ResourceStorage:    c.ResourceStorage,
		VPConditions:       c.VPConditions,
		StartingResources:  c.StartingResources,
		StartingProduction: c.StartingProduction,
	}
}
