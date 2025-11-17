package player

import (
	"github.com/google/uuid"
	"terraforming-mars-backend/internal/model"
)

// SelectStartingCardsPhase represents the starting card selection state
type SelectStartingCardsPhase struct {
	AvailableCards        []string `json:"availableCards"`
	AvailableCorporations []string `json:"availableCorporations"`
	SelectionComplete     bool     `json:"selectionComplete"`
}

// Player represents a player in the game
// For Phase 1-3, we expand to support card selection and corporation
type Player struct {
	ID                       string                    `json:"id"`
	Name                     string                    `json:"name"`
	Resources                model.Resources           `json:"resources"`
	Production               model.Production          `json:"production"`
	TerraformRating          int                       `json:"terraformRating"`
	IsConnected              bool                      `json:"isConnected"`
	SelectStartingCardsPhase *SelectStartingCardsPhase `json:"selectStartingCardsPhase"`
	Cards                    []string                  `json:"cards"`         // Card IDs in hand
	CorporationID            string                    `json:"corporationId"` // Selected corporation
}

// NewPlayer creates a new player with default starting values
func NewPlayer(name string) *Player {
	return &Player{
		ID:   uuid.New().String(),
		Name: name,
		Resources: model.Resources{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		Production: model.Production{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		TerraformRating: 20, // Starting TR
		IsConnected:     true,
		Cards:           make([]string, 0),
		CorporationID:   "",
	}
}
