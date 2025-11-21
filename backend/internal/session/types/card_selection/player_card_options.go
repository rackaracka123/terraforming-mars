package card_selection

// PlayerCardOptions represents the card options for a specific player
type PlayerCardOptions struct {
	PlayerID    string   `json:"playerId" ts:"string"`
	CardOptions []string `json:"cardOptions" ts:"string[]"`
}
