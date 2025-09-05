package model

// Player represents a player in the game
type Player struct {
	ID              string     `json:"id" ts:"string"`
	Name            string     `json:"name" ts:"string"`
	Corporation     string     `json:"corporation" ts:"string"`
	Cards           []string   `json:"cards" ts:"string[]"`
	Resources       Resources  `json:"resources" ts:"Resources"`
	Production      Production `json:"production" ts:"Production"`
	TerraformRating int        `json:"terraformRating" ts:"number"`
	IsActive        bool       `json:"isActive" ts:"boolean"`
	PlayedCards     []string   `json:"playedCards" ts:"string[]"`
}

// CanAffordStandardProject checks if the player has enough credits for a standard project
func (p *Player) CanAffordStandardProject(project StandardProject) bool {
	cost, exists := StandardProjectCost[project]
	if !exists {
		return false
	}
	return p.Resources.Credits >= cost
}

// HasCardsToSell checks if the player has enough cards in hand to sell
func (p *Player) HasCardsToSell(count int) bool {
	return len(p.Cards) >= count && count > 0
}

// GetMaxCardsToSell returns the maximum number of cards the player can sell
func (p *Player) GetMaxCardsToSell() int {
	return len(p.Cards)
}