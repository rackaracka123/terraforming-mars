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
	}
}