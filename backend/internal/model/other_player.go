package model

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
	PlayedCards      []string   `json:"playedCards" ts:"string[]"` // Played cards are public
	Passed           bool       `json:"passed" ts:"boolean"`
	AvailableActions int        `json:"availableActions" ts:"number"`
	VictoryPoints    int        `json:"victoryPoints" ts:"number"`
	IsConnected      bool       `json:"isConnected" ts:"boolean"`
}

// NewOtherPlayerFromPlayer creates an OtherPlayer from a full Player
// This hides the hand cards but keeps played cards visible
func NewOtherPlayerFromPlayer(player *Player) *OtherPlayer {
	if player == nil {
		return nil
	}

	corporationName := ""
	if player.Corporation != nil {
		corporationName = player.Corporation.Name
	}

	// Retrieve resources and production via service
	resources, _ := player.GetResources()
	production, _ := player.GetProduction()
	passed, _ := player.GetPassed()
	availableActions, _ := player.GetAvailableActions()

	return &OtherPlayer{
		ID:               player.ID,
		Name:             player.Name,
		Corporation:      corporationName,
		HandCardCount:    len(player.Cards), // Convert hand cards to count
		Resources:        Resources{
			Credits:  resources.Credits,
			Steel:    resources.Steel,
			Titanium: resources.Titanium,
			Plants:   resources.Plants,
			Energy:   resources.Energy,
			Heat:     resources.Heat,
		},
		Production:       Production{
			Credits:  production.Credits,
			Steel:    production.Steel,
			Titanium: production.Titanium,
			Plants:   production.Plants,
			Energy:   production.Energy,
			Heat:     production.Heat,
		},
		TerraformRating:  player.TerraformRating,
		PlayedCards:      append([]string{}, player.PlayedCards...), // Copy played cards (public)
		Passed:           passed,
		AvailableActions: availableActions,
		VictoryPoints:    player.VictoryPoints,
		IsConnected:      player.IsConnected,
	}
}

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
