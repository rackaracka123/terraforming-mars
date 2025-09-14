package model

// ConnectionStatus constants removed - now using simple boolean isConnected

// ProductionPhase contains both card selection and production phase state for a player
type ProductionPhase struct {
	AvailableCards    []Card `json:"availableCards" ts:"CardDto[]"`  // Cards available for selection
	SelectionComplete bool   `json:"selectionComplete" ts:"boolean"` // Whether player completed card selection
}

// Player represents a player in the game
type Player struct {
	ID               string     `json:"id" ts:"string"`
	Name             string     `json:"name" ts:"string"`
	Corporation      string     `json:"corporation" ts:"string"`
	Cards            []string   `json:"cards" ts:"string[]"`
	Resources        Resources  `json:"resources" ts:"Resources"`
	Production       Production `json:"production" ts:"Production"`
	TerraformRating  int        `json:"terraformRating" ts:"number"`
	PlayedCards      []string   `json:"playedCards" ts:"string[]"`
	Passed           bool       `json:"passed" ts:"boolean"`
	AvailableActions int        `json:"availableActions" ts:"number"`
	VictoryPoints    int        `json:"victoryPoints" ts:"number"`
	IsConnected      bool       `json:"isConnected" ts:"boolean"`
	// Card selection and production phase - nullable, exists only during selection phase
	ProductionSelection *ProductionPhase `json:"production" ts:"ProductionPhase | null"` // Card selection and production state, null when not selecting
	// Starting card selection - nullable, exists only during starting card selection phase
	StartingSelection []Card `json:"startingSelection" ts:"Card[]"` // Starting cards available for selection (10 cards)
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
	if p.ProductionSelection != nil {
		// Copy available cards slice - copy each card struct
		availableCardsCopy := make([]Card, len(p.ProductionSelection.AvailableCards))
		for i, card := range p.ProductionSelection.AvailableCards {
			availableCardsCopy[i] = card // Card struct copy (should be sufficient as it contains mainly primitives and slices)
		}

		productionSelectionCopy = &ProductionPhase{
			AvailableCards:    availableCardsCopy,
			SelectionComplete: p.ProductionSelection.SelectionComplete,
		}
	}

	// Copy starting selection slice
	var startingSelectionCopy []Card
	if p.StartingSelection != nil {
		startingSelectionCopy = make([]Card, len(p.StartingSelection))
		copy(startingSelectionCopy, p.StartingSelection)
	}

	return &Player{
		ID:                  p.ID,
		Name:                p.Name,
		Corporation:         p.Corporation,
		Cards:               cardsCopy,
		Resources:           p.Resources,  // Resources is a struct, so this is copied by value
		Production:          p.Production, // Production is a struct, so this is copied by value
		TerraformRating:     p.TerraformRating,
		PlayedCards:         playedCardsCopy,
		Passed:              p.Passed,
		AvailableActions:    p.AvailableActions,
		VictoryPoints:       p.VictoryPoints,
		IsConnected:         p.IsConnected,
		ProductionSelection: productionSelectionCopy,
		StartingSelection:   startingSelectionCopy,
	}
}
