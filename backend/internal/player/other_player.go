package player

// OtherPlayer represents a player from another player's perspective
// Contains public information only - hand cards are hidden but played cards are visible
type OtherPlayer struct {
	ID               string     `json:"id" ts:"string"`
	Name             string     `json:"name" ts:"string"`
	Corporation      string     `json:"corporation" ts:"string"`
	HandCardCount    int        `json:"handCardCount" ts:"number"` // Number of cards in hand (private)
	Resources        Resources  `json:"resources" ts:"Resources"`
	Production       Production `json:"production" ts:"Production"`
	TerraformRating  int        `json:"terraformRating" ts:"number"`
	PlayedCards      []Card     `json:"playedCards" ts:"CardDto[]"` // Played cards are public (Card instances)
	Passed           bool       `json:"passed" ts:"boolean"`
	AvailableActions int        `json:"availableActions" ts:"number"`
	VictoryPoints    int        `json:"victoryPoints" ts:"number"`
	IsConnected      bool       `json:"isConnected" ts:"boolean"`
}

// NewOtherPlayerFromPlayer creates an OtherPlayer from a full Player
// This hides the hand cards but keeps played cards visible
func NewOtherPlayerFromPlayer(p *Player) *OtherPlayer {
	if p == nil {
		return nil
	}

	corporationName := ""
	if p.Corporation != nil {
		corporationName = p.Corporation.Name
	}

	// TODO: Passed and AvailableActions need to be fetched from turn service via Session
	// For now, use default values - these will be populated by the Session/DTO layer
	passed := false
	availableActions := 0

	return &OtherPlayer{
		ID:               p.ID,
		Name:             p.Name,
		Corporation:      corporationName,
		HandCardCount:    len(p.Cards), // Convert hand cards to count
		Resources:        p.Resources,
		Production:       p.Production,
		TerraformRating:  p.TerraformRating,
		PlayedCards:      copyCards(p.PlayedCards), // Copy played card instances (public)
		Passed:           passed,
		AvailableActions: availableActions,
		VictoryPoints:    p.VictoryPoints,
		IsConnected:      p.IsConnected,
	}
}

// DeepCopy creates a deep copy of the OtherPlayer
func (op *OtherPlayer) DeepCopy() *OtherPlayer {
	if op == nil {
		return nil
	}

	// Deep copy played cards slice (Card instances)
	playedCardsCopy := copyCards(op.PlayedCards)

	return &OtherPlayer{
		ID:               op.ID,
		Name:             op.Name,
		Corporation:      op.Corporation,
		HandCardCount:    op.HandCardCount,
		Resources:        op.Resources,  // Resources is a struct, so this is copied by value
		Production:       op.Production, // Production is a struct, so this is copied by value
		TerraformRating:  op.TerraformRating,
		PlayedCards:      playedCardsCopy,
		Passed:           op.Passed,
		AvailableActions: op.AvailableActions,
		VictoryPoints:    op.VictoryPoints,
		IsConnected:      op.IsConnected,
	}
}

// copyCards creates deep copies of Card instances
func copyCards(cards []Card) []Card {
	if cards == nil {
		return nil
	}
	copied := make([]Card, len(cards))
	for i, card := range cards {
		copied[i] = card.DeepCopy()
	}
	return copied
}
