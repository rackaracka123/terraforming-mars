package player

import (
	"terraforming-mars-backend/internal/features/resources"
)

// OtherPlayer represents a player from another player's perspective
// Contains public information only - hand cards are hidden but played cards are visible
type OtherPlayer struct {
	ID               string               `json:"id" ts:"string"`
	Name             string               `json:"name" ts:"string"`
	Corporation      string               `json:"corporation" ts:"string"`
	HandCardCount    int                  `json:"handCardCount" ts:"number"` // Number of cards in hand (private)
	Resources        resources.Resources  `json:"resources" ts:"Resources"`
	Production       resources.Production `json:"production" ts:"Production"`
	TerraformRating  int                  `json:"terraformRating" ts:"number"`
	PlayedCards      []string             `json:"playedCards" ts:"string[]"` // Played cards are public
	Passed           bool                 `json:"passed" ts:"boolean"`
	AvailableActions int                  `json:"availableActions" ts:"number"`
	VictoryPoints    int                  `json:"victoryPoints" ts:"number"`
	IsConnected      bool                 `json:"isConnected" ts:"boolean"`
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

	// Retrieve resources and production via service
	res, _ := p.GetResources()
	prod, _ := p.GetProduction()
	passed, _ := p.GetPassed()
	availableActions, _ := p.GetAvailableActions()

	return &OtherPlayer{
		ID:            p.ID,
		Name:          p.Name,
		Corporation:   corporationName,
		HandCardCount: len(p.Cards), // Convert hand cards to count
		Resources: resources.Resources{
			Credits:  res.Credits,
			Steel:    res.Steel,
			Titanium: res.Titanium,
			Plants:   res.Plants,
			Energy:   res.Energy,
			Heat:     res.Heat,
		},
		Production: resources.Production{
			Credits:  prod.Credits,
			Steel:    prod.Steel,
			Titanium: prod.Titanium,
			Plants:   prod.Plants,
			Energy:   prod.Energy,
			Heat:     prod.Heat,
		},
		TerraformRating:  p.TerraformRating,
		PlayedCards:      append([]string{}, p.PlayedCards...), // Copy played cards (public)
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
