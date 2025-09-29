package model

// ProductionPhase contains both card selection and production phase state for a player
type ProductionPhase struct {
	AvailableCards    []string  `json:"availableCards" ts:"CardDto[]"`  // Card IDs available for selection
	SelectionComplete bool      `json:"selectionComplete" ts:"boolean"` // Whether player completed card selection
	BeforeResources   Resources `json:"beforeResources" ts:"ResourcesDto"`
	AfterResources    Resources `json:"afterResources" ts:"ResourcesDto"`
	EnergyConverted   int       `json:"energyConverted" ts:"number"`
	CreditsIncome     int       `json:"creditsIncome" ts:"number"`
}

// DeepCopy creates a deep copy of the ProductionPhase
func (p *ProductionPhase) DeepCopy() *ProductionPhase {
	if p == nil {
		return nil
	}

	return &ProductionPhase{
		AvailableCards:    p.AvailableCards,
		SelectionComplete: p.SelectionComplete,
		BeforeResources:   p.BeforeResources.DeepCopy(),
		AfterResources:    p.AfterResources.DeepCopy(),
		EnergyConverted:   p.EnergyConverted,
		CreditsIncome:     p.CreditsIncome,
	}
}

type SelectStartingCardsPhase struct {
	AvailableCards    []string `json:"availableCards" ts:"CardDto[]"`  // Card IDs available for selection
	SelectionComplete bool     `json:"selectionComplete" ts:"boolean"` // Whether player completed card selection
}

// PendingTileSelection represents a pending tile placement action
type PendingTileSelection struct {
	TileType       string   `json:"tileType" ts:"string"`         // "city", "greenery", "ocean"
	AvailableHexes []string `json:"availableHexes" ts:"string[]"` // Backend-calculated valid hex coordinates
	Source         string   `json:"source" ts:"string"`           // What triggered this selection (card ID, standard project, etc.)
}

// Player represents a player in the game
type Player struct {
	ID                       string                    `json:"id" ts:"string"`
	Name                     string                    `json:"name" ts:"string"`
	Corporation              *string                   `json:"corporation" ts:"string | null"`
	Cards                    []string                  `json:"cards" ts:"string[]"`
	Resources                Resources                 `json:"resources" ts:"Resources"`
	Production               Production                `json:"production" ts:"Production"`
	TerraformRating          int                       `json:"terraformRating" ts:"number"`
	PlayedCards              []string                  `json:"playedCards" ts:"string[]"`
	Passed                   bool                      `json:"passed" ts:"boolean"`
	AvailableActions         int                       `json:"availableActions" ts:"number"`
	VictoryPoints            int                       `json:"victoryPoints" ts:"number"`
	IsConnected              bool                      `json:"isConnected" ts:"boolean"`
	Effects                  []PlayerEffect            `json:"effects" ts:"PlayerEffect[]"` // Active ongoing effects (discounts, special abilities, etc.)
	Actions                  []PlayerAction            `json:"actions" ts:"PlayerAction[]"` // Available actions from played cards with manual triggers
	ProductionPhase          *ProductionPhase          `json:"productionPhase" ts:"ProductionPhase | null"`
	SelectStartingCardsPhase *SelectStartingCardsPhase `json:"selectStartingCardsPhase" ts:"selectStartingCardsPhase | null"`
	// Tile selection - nullable, exists only when player needs to place tiles
	PendingTileSelection *PendingTileSelection `json:"pendingTileSelection" ts:"PendingTileSelection | null"` // Pending tile placement, null when no tiles to place
}

// GetStartingSelectionCards returns the player's starting card selection, nil if not in that phase
func (p *Player) GetStartingSelectionCards() []string {
	if p.SelectStartingCardsPhase == nil {
		return nil
	}

	return p.SelectStartingCardsPhase.AvailableCards
}

// GetProductionPhaseCards returns the player's production phase card selection, nil if not in that phase
func (p *Player) GetProductionPhaseCards() []string {
	if p.ProductionPhase == nil {
		return nil
	}

	return p.ProductionPhase.AvailableCards
}

// DeepCopy creates a deep copy of the Player
func (p *Player) DeepCopy() *Player {
	if p == nil {
		return nil
	}

	// Copy cards slice
	cardsCopy := make([]string, len(p.Cards))
	copy(cardsCopy, p.Cards)

	// Copy played cards slice
	playedCardsCopy := make([]string, len(p.PlayedCards))
	copy(playedCardsCopy, p.PlayedCards)

	// Deep copy production selection if it exists
	var productionSelectionCopy *ProductionPhase
	if p.ProductionPhase != nil {
		// Copy available cards slice
		availableCardsCopy := make([]string, len(p.ProductionPhase.AvailableCards))
		copy(availableCardsCopy, p.ProductionPhase.AvailableCards)

		productionSelectionCopy = &ProductionPhase{
			AvailableCards:    availableCardsCopy,
			SelectionComplete: p.ProductionPhase.SelectionComplete,

			BeforeResources: p.ProductionPhase.BeforeResources.DeepCopy(),
			AfterResources:  p.ProductionPhase.AfterResources.DeepCopy(),
			EnergyConverted: p.ProductionPhase.EnergyConverted,
			CreditsIncome:   p.ProductionPhase.CreditsIncome,
		}
	}

	// Copy starting selection slice
	var startingSelectionCopy *SelectStartingCardsPhase
	if p.SelectStartingCardsPhase != nil {
		availableCardsCopy := make([]string, len(p.SelectStartingCardsPhase.AvailableCards))
		copy(availableCardsCopy, p.SelectStartingCardsPhase.AvailableCards)

		startingSelectionCopy = &SelectStartingCardsPhase{
			AvailableCards:    availableCardsCopy,
			SelectionComplete: p.SelectStartingCardsPhase.SelectionComplete,
		}
	}

	// Deep copy pending tile selection if it exists
	var pendingTileSelectionCopy *PendingTileSelection
	if p.PendingTileSelection != nil {
		// Copy available hexes slice
		availableHexesCopy := make([]string, len(p.PendingTileSelection.AvailableHexes))
		copy(availableHexesCopy, p.PendingTileSelection.AvailableHexes)

		pendingTileSelectionCopy = &PendingTileSelection{
			TileType:       p.PendingTileSelection.TileType,
			AvailableHexes: availableHexesCopy,
			Source:         p.PendingTileSelection.Source,
		}
	}

	// Deep copy effects slice
	effectsCopy := make([]PlayerEffect, len(p.Effects))
	for i, effect := range p.Effects {
		effectsCopy[i] = *effect.DeepCopy()
	}

	// Deep copy actions slice
	actionsCopy := make([]PlayerAction, len(p.Actions))
	for i, action := range p.Actions {
		actionsCopy[i] = *action.DeepCopy()
	}

	return &Player{
		ID:                       p.ID,
		Name:                     p.Name,
		Corporation:              p.Corporation,
		Cards:                    cardsCopy,
		Resources:                p.Resources,  // Resources is a struct, so this is copied by value
		Production:               p.Production, // Production is a struct, so this is copied by value
		TerraformRating:          p.TerraformRating,
		PlayedCards:              playedCardsCopy,
		Passed:                   p.Passed,
		AvailableActions:         p.AvailableActions,
		VictoryPoints:            p.VictoryPoints,
		IsConnected:              p.IsConnected,
		Effects:                  effectsCopy,
		Actions:                  actionsCopy,
		ProductionPhase:          productionSelectionCopy,
		SelectStartingCardsPhase: startingSelectionCopy,
		PendingTileSelection:     pendingTileSelectionCopy,
	}
}
