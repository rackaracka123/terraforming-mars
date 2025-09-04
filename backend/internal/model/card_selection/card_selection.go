package card_selection

// CardSelection represents the card selection phase data
type CardSelection struct {
	PlayerCardOptions            []PlayerCardOptions `json:"playerCardOptions" ts:"PlayerCardOptions[]"`
	PlayersWhoCompletedSelection []string            `json:"playersWhoCompletedSelection" ts:"string[]"`
}
