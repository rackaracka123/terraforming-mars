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
	AvailableCards        []string `json:"availableCards" ts:"CardDto[]"`       // Card IDs available for selection
	AvailableCorporations []string `json:"availableCorporations" ts:"string[]"` // Corporation IDs available for selection (2 corporations)
	SelectionComplete     bool     `json:"selectionComplete" ts:"boolean"`      // Whether player completed card selection
}

// PendingTileSelection represents a pending tile placement action
type PendingTileSelection struct {
	TileType       string   `json:"tileType" ts:"string"`         // "city", "greenery", "ocean"
	AvailableHexes []string `json:"availableHexes" ts:"string[]"` // Backend-calculated valid hex coordinates
	Source         string   `json:"source" ts:"string"`           // What triggered this selection (card ID, standard project, etc.)
}

// PendingTileSelectionQueue represents a queue of tile placements to be made
type PendingTileSelectionQueue struct {
	Items  []string `json:"items" ts:"string[]"` // Queue of tile types: ["city", "city", "ocean"]
	Source string   `json:"source" ts:"string"`  // Card ID that triggered all placements
}

// PendingCardSelection represents a pending card selection action (e.g., sell patents, card effects)
type PendingCardSelection struct {
	AvailableCards []string       `json:"availableCards" ts:"string[]"`            // Card IDs player can select from
	CardCosts      map[string]int `json:"cardCosts" ts:"Record<string, number>"`   // Card ID -> cost to select (0 for sell patents, 3 for buying cards)
	CardRewards    map[string]int `json:"cardRewards" ts:"Record<string, number>"` // Card ID -> reward for selecting (1 MC for sell patents)
	Source         string         `json:"source" ts:"string"`                      // What triggered this selection ("sell-patents", card ID, etc.)
	MinCards       int            `json:"minCards" ts:"number"`                    // Minimum cards to select (0 for sell patents)
	MaxCards       int            `json:"maxCards" ts:"number"`                    // Maximum cards to select (hand size for sell patents)
}

// Player represents a player in the game
type Player struct {
	ID                       string                    `json:"id" ts:"string"`
	Name                     string                    `json:"name" ts:"string"`
	Corporation              *Card                     `json:"corporation" ts:"CardDto | null"`
	Cards                    []string                  `json:"cards" ts:"string[]"`
	Resources                Resources                 `json:"resources" ts:"Resources"`
	Production               Production                `json:"production" ts:"Production"`
	TerraformRating          int                       `json:"terraformRating" ts:"number"`
	PlayedCards              []string                  `json:"playedCards" ts:"string[]"`
	Passed                   bool                      `json:"passed" ts:"boolean"`
	AvailableActions         int                       `json:"availableActions" ts:"number"`
	VictoryPoints            int                       `json:"victoryPoints" ts:"number"`
	IsConnected              bool                      `json:"isConnected" ts:"boolean"`
	Actions                  []PlayerAction            `json:"actions" ts:"PlayerAction[]"` // Available actions from played cards with manual triggers
	ProductionPhase          *ProductionPhase          `json:"productionPhase" ts:"ProductionPhase | null"`
	SelectStartingCardsPhase *SelectStartingCardsPhase `json:"selectStartingCardsPhase" ts:"selectStartingCardsPhase | null"`
	// Tile selection - nullable, exists only when player needs to place tiles
	PendingTileSelection      *PendingTileSelection      `json:"pendingTileSelection" ts:"PendingTileSelection | null"`           // Current active tile placement, null when no tiles to place
	PendingTileSelectionQueue *PendingTileSelectionQueue `json:"pendingTileSelectionQueue" ts:"PendingTileSelectionQueue | null"` // Queue of remaining tile placements from cards
	// Card selection - nullable, exists only when player needs to select cards
	PendingCardSelection *PendingCardSelection `json:"pendingCardSelection" ts:"PendingCardSelection | null"` // Current active card selection (sell patents, card effects, etc.)
	// Resource storage - maps card IDs to resource counts stored on those cards
	ResourceStorage map[string]int `json:"resourceStorage" ts:"Record<string, number>"` // Card ID -> resource count
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

		availableCorporationsCopy := make([]string, len(p.SelectStartingCardsPhase.AvailableCorporations))
		copy(availableCorporationsCopy, p.SelectStartingCardsPhase.AvailableCorporations)

		startingSelectionCopy = &SelectStartingCardsPhase{
			AvailableCards:        availableCardsCopy,
			AvailableCorporations: availableCorporationsCopy,
			SelectionComplete:     p.SelectStartingCardsPhase.SelectionComplete,
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

	// Deep copy pending tile selection queue if it exists
	var pendingTileSelectionQueueCopy *PendingTileSelectionQueue
	if p.PendingTileSelectionQueue != nil {
		// Copy items slice
		itemsCopy := make([]string, len(p.PendingTileSelectionQueue.Items))
		copy(itemsCopy, p.PendingTileSelectionQueue.Items)

		pendingTileSelectionQueueCopy = &PendingTileSelectionQueue{
			Items:  itemsCopy,
			Source: p.PendingTileSelectionQueue.Source,
		}
	}

	// Deep copy actions slice
	actionsCopy := make([]PlayerAction, len(p.Actions))
	for i, action := range p.Actions {
		actionsCopy[i] = *action.DeepCopy()
	}

	// Deep copy resource storage map
	resourceStorageCopy := make(map[string]int)
	for cardID, count := range p.ResourceStorage {
		resourceStorageCopy[cardID] = count
	}

	// Deep copy pending card selection if it exists
	var pendingCardSelectionCopy *PendingCardSelection
	if p.PendingCardSelection != nil {
		// Copy available cards slice
		availableCardsCopy := make([]string, len(p.PendingCardSelection.AvailableCards))
		copy(availableCardsCopy, p.PendingCardSelection.AvailableCards)

		// Copy card costs map
		cardCostsCopy := make(map[string]int)
		for cardID, cost := range p.PendingCardSelection.CardCosts {
			cardCostsCopy[cardID] = cost
		}

		// Copy card rewards map
		cardRewardsCopy := make(map[string]int)
		for cardID, reward := range p.PendingCardSelection.CardRewards {
			cardRewardsCopy[cardID] = reward
		}

		pendingCardSelectionCopy = &PendingCardSelection{
			AvailableCards: availableCardsCopy,
			CardCosts:      cardCostsCopy,
			CardRewards:    cardRewardsCopy,
			Source:         p.PendingCardSelection.Source,
			MinCards:       p.PendingCardSelection.MinCards,
			MaxCards:       p.PendingCardSelection.MaxCards,
		}
	}

	// Deep copy corporation if it exists
	var corporationCopy *Card
	if p.Corporation != nil {
		corpCopy := p.Corporation.DeepCopy()
		corporationCopy = &corpCopy
	}

	return &Player{
		ID:                        p.ID,
		Name:                      p.Name,
		Corporation:               corporationCopy,
		Cards:                     cardsCopy,
		Resources:                 p.Resources,  // Resources is a struct, so this is copied by value
		Production:                p.Production, // Production is a struct, so this is copied by value
		TerraformRating:           p.TerraformRating,
		PlayedCards:               playedCardsCopy,
		Passed:                    p.Passed,
		AvailableActions:          p.AvailableActions,
		VictoryPoints:             p.VictoryPoints,
		IsConnected:               p.IsConnected,
		Actions:                   actionsCopy,
		ProductionPhase:           productionSelectionCopy,
		SelectStartingCardsPhase:  startingSelectionCopy,
		PendingTileSelection:      pendingTileSelectionCopy,
		PendingTileSelectionQueue: pendingTileSelectionQueueCopy,
		PendingCardSelection:      pendingCardSelectionCopy,
		ResourceStorage:           resourceStorageCopy,
	}
}
