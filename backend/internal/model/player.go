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