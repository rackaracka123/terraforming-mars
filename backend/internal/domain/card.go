package domain

// CardType represents different types of cards in Terraforming Mars
type CardType string

const (
	CardTypeEffect      CardType = "effect"      // Green cards - immediate effects, production bonuses
	CardTypeActive      CardType = "active"      // Blue cards - ongoing effects, repeatable actions
	CardTypeEvent       CardType = "event"       // Red cards - one-time effects
	CardTypeCorporation CardType = "corporation" // Corporation cards - unique player abilities
)

// Card represents a game card
type Card struct {
	ID          string   `json:"id" ts:"string"`
	Name        string   `json:"name" ts:"string"`
	Type        CardType `json:"type" ts:"CardType"`
	Cost        int      `json:"cost" ts:"number"`
	Description string   `json:"description" ts:"string"`
}

// GetStartingCards returns simple starting cards for players to choose from
func GetStartingCards() []Card {
	return []Card{
		{
			ID:          "investment",
			Name:        "Investment",
			Type:        CardTypeEvent,
			Cost:        5,
			Description: "Gain 1 VP for 5 megacredits",
		},
		{
			ID:          "early-settlement",
			Name:        "Early Settlement", 
			Type:        CardTypeEffect,
			Cost:        8,
			Description: "Gain 1 MC production for 8 megacredits",
		},
		{
			ID:          "research-grant",
			Name:        "Research Grant",
			Type:        CardTypeEvent,
			Cost:        3,
			Description: "Draw 1 additional card next turn for 3 megacredits",
		},
		{
			ID:          "power-plant",
			Name:        "Power Plant",
			Type:        CardTypeEffect,
			Cost:        6,
			Description: "Gain 1 Energy production for 6 megacredits",
		},
		{
			ID:          "heat-generators",
			Name:        "Heat Generators",
			Type:        CardTypeEffect,
			Cost:        4,
			Description: "Gain 1 Heat production for 4 megacredits",
		},
		{
			ID:          "mining-operation",
			Name:        "Mining Operation",
			Type:        CardTypeEvent,
			Cost:        7,
			Description: "Gain 2 Steel for 7 megacredits",
		},
		{
			ID:          "space-mirrors",
			Name:        "Space Mirrors",
			Type:        CardTypeActive,
			Cost:        10,
			Description: "Action: Spend 7 megacredits to gain 1 Energy production",
		},
		{
			ID:          "water-import",
			Name:        "Water Import from Europa",
			Type:        CardTypeEvent,
			Cost:        12,
			Description: "Place 1 ocean tile for 12 megacredits",
		},
		{
			ID:          "nitrogen-plants",
			Name:        "Nitrogen-Rich Plants",
			Type:        CardTypeEffect,
			Cost:        9,
			Description: "Gain 1 Plant production for 9 megacredits",
		},
		{
			ID:          "atmospheric-processors",
			Name:        "Atmospheric Processors",
			Type:        CardTypeEffect,
			Cost:        11,
			Description: "Raise oxygen 1 step for 11 megacredits",
		},
	}
}