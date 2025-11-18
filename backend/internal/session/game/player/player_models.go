package player

import (
	"github.com/google/uuid"
	"terraforming-mars-backend/internal/model"
)

// SelectStartingCardsPhase represents the starting card selection state
type SelectStartingCardsPhase struct {
	AvailableCards        []string `json:"availableCards"`
	AvailableCorporations []string `json:"availableCorporations"`
}

// Player represents a player in the game
// Expanded to support full game logic including cards, actions, and effects
type Player struct {
	ID                        string                           `json:"id"`
	Name                      string                           `json:"name"`
	Resources                 model.Resources                  `json:"resources"`
	Production                model.Production                 `json:"production"`
	TerraformRating           int                              `json:"terraformRating"`
	IsConnected               bool                             `json:"isConnected"`
	SelectStartingCardsPhase  *SelectStartingCardsPhase        `json:"selectStartingCardsPhase"`
	ProductionPhase           *model.ProductionPhase           `json:"productionPhase"`
	Cards                     []string                         `json:"cards"`                     // Card IDs in hand
	PlayedCards               []string                         `json:"playedCards"`               // Card IDs played to the table
	CorporationID             string                           `json:"corporationId"`             // Selected corporation ID
	Corporation               *model.Card                      `json:"corporation"`               // Full corporation card data
	PaymentSubstitutes        []model.PaymentSubstitute        `json:"paymentSubstitutes"`        // Payment substitutes from cards
	Actions                   []model.PlayerAction             `json:"actions"`                   // Available actions from cards
	ForcedFirstAction         *model.ForcedFirstAction         `json:"forcedFirstAction"`         // Forced first turn action
	RequirementModifiers      []model.RequirementModifier      `json:"requirementModifiers"`      // Requirement modifiers from cards
	Effects                   []model.PlayerEffect             `json:"effects"`                   // Active passive effects from cards
	VictoryPoints             int                              `json:"victoryPoints"`             // Victory points
	PendingTileSelection      *model.PendingTileSelection      `json:"pendingTileSelection"`      // Pending tile selection
	PendingTileSelectionQueue *model.PendingTileSelectionQueue `json:"pendingTileSelectionQueue"` // Pending tile selection queue
	PendingCardDrawSelection  *model.PendingCardDrawSelection  `json:"pendingCardDrawSelection"`  // Pending card draw selection
	ResourceStorage           map[string]int                   `json:"resourceStorage"`           // Card resource storage (animals, microbes, etc.)
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
		TerraformRating:           20, // Starting TR
		IsConnected:               true,
		Cards:                     make([]string, 0),
		PlayedCards:               make([]string, 0),
		CorporationID:             "",
		Corporation:               nil,
		PaymentSubstitutes:        make([]model.PaymentSubstitute, 0),
		Actions:                   make([]model.PlayerAction, 0),
		ForcedFirstAction:         nil,
		RequirementModifiers:      make([]model.RequirementModifier, 0),
		VictoryPoints:             0,
		PendingTileSelection:      nil,
		PendingTileSelectionQueue: nil,
		PendingCardDrawSelection:  nil,
		ResourceStorage:           make(map[string]int),
	}
}
