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
	ID               string              `json:"id" ts:"string"`
	Name             string              `json:"name" ts:"string"`
	Type             CardType            `json:"type" ts:"CardType"`
	Cost             int                 `json:"cost" ts:"number"`
	Description      string              `json:"description" ts:"string"`
	Tags             []CardTag           `json:"tags" ts:"CardTag[]"`
	Requirements     CardRequirements    `json:"requirements" ts:"CardRequirements"`
	VictoryPoints    int                 `json:"victoryPoints" ts:"number"`
	Number           string              `json:"number" ts:"string"`                             // Card number (e.g., "#001")
	ProductionEffects *ProductionEffects `json:"productionEffects,omitempty" ts:"ProductionEffects | undefined"` // Production changes
}

// GetStartingCards returns simple starting cards for players to choose from
// This function is deprecated - use CardDataService.GetStartingCardPool() instead
// Kept for backward compatibility during transition
func GetStartingCards() []Card {
	// Return empty slice - actual cards will come from CardDataService
	return []Card{}
}
