package types

// OtherPlayer represents a player from another player's perspective
// Contains public information only - hand cards are hidden but played cards are visible
type OtherPlayer struct {
	ID               string
	Name             string
	Corporation      string
	HandCardCount    int // Number of cards in hand (private)
	Resources        Resources
	Production       Production
	TerraformRating  int
	PlayedCards      []string // Played cards are public
	Passed           bool
	AvailableActions int
	VictoryPoints    int
	IsConnected      bool
}

// NOTE: NewOtherPlayerFromPlayer has been moved to the player package to avoid import cycles

// DeepCopy creates a deep copy of the OtherPlayer
func (op *OtherPlayer) DeepCopy() *OtherPlayer {
	if op == nil {
		return nil
	}

	// Copy played cards slice
	playedCardsCopy := make([]string, len(op.PlayedCards))
	copy(playedCardsCopy, op.PlayedCards)

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
