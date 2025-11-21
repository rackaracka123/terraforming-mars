package player

import (
	"terraforming-mars-backend/internal/session/types"
)

// Player is an alias to the unified Player type
type Player = types.Player

// SelectStartingCardsPhase is an alias to the unified type
type SelectStartingCardsPhase = types.SelectStartingCardsPhase

// PendingCardSelection is an alias to the unified type
type PendingCardSelection = types.PendingCardSelection

// NewPlayer creates a new player with default starting values
// This is a compatibility wrapper that delegates to the unified types
func NewPlayer(name string) *Player {
	// Create player using types.Player directly
	player := &types.Player{
		Name:                 name,
		Resources:            types.Resources{},
		Production:           types.Production{},
		TerraformRating:      20,
		IsConnected:          true,
		Passed:               false,
		AvailableActions:     2,
		Cards:                make([]string, 0),
		PlayedCards:          make([]string, 0),
		PaymentSubstitutes:   make([]types.PaymentSubstitute, 0),
		Actions:              make([]types.PlayerAction, 0),
		Effects:              make([]types.PlayerEffect, 0),
		RequirementModifiers: make([]types.RequirementModifier, 0),
		ResourceStorage:      make(map[string]int),
	}
	return player
}
