package player

import (
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/production"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/shared/types"
)

// Type aliases to avoid verbose references
type Card = card.Card
type CardType = card.CardType
type PlayerEffect = card.PlayerEffect
type PlayerAction = card.PlayerAction
type PaymentSubstitute = types.PaymentSubstitute
type Resources = resources.Resources
type Production = resources.Production

// ProductionPhase contains production phase state for display only
// Actual card selection state is managed by production feature service
type ProductionPhase struct {
	AvailableCards    []string            `json:"availableCards" ts:"CardDto[]"`  // Card IDs available for selection
	SelectionComplete bool                `json:"selectionComplete" ts:"boolean"` // Whether player completed card selection
	BeforeResources   resources.Resources `json:"beforeResources" ts:"ResourcesDto"`
	AfterResources    resources.Resources `json:"afterResources" ts:"ResourcesDto"`
	EnergyConverted   int                 `json:"energyConverted" ts:"number"`
	CreditsIncome     int                 `json:"creditsIncome" ts:"number"`
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

// PendingTileSelection is a type alias to avoid circular imports
// The actual definition is in internal/features/tiles/tile_queue_repository.go
type PendingTileSelection = tiles.PendingTileSelection

// PendingTileSelectionQueue is a type alias to avoid circular imports
// The actual definition is in internal/features/tiles/tile_queue_repository.go
type PendingTileSelectionQueue = tiles.PendingTileSelectionQueue

// PendingCardSelection represents a pending card selection action (e.g., sell patents, card effects)
type PendingCardSelection struct {
	AvailableCards []string       `json:"availableCards" ts:"string[]"`            // Card IDs player can select from
	CardCosts      map[string]int `json:"cardCosts" ts:"Record<string, number>"`   // Card ID -> cost to select (0 for sell patents, 3 for buying cards)
	CardRewards    map[string]int `json:"cardRewards" ts:"Record<string, number>"` // Card ID -> reward for selecting (1 MC for sell patents)
	Source         string         `json:"source" ts:"string"`                      // What triggered this selection ("sell-patents", card ID, etc.)
	MinCards       int            `json:"minCards" ts:"number"`                    // Minimum cards to select (0 for sell patents)
	MaxCards       int            `json:"maxCards" ts:"number"`                    // Maximum cards to select (hand size for sell patents)
}

// PendingCardDrawSelection represents a pending card draw/peek/take/buy action from card effects
type PendingCardDrawSelection struct {
	AvailableCards []string `json:"availableCards" ts:"CardDto[]"` // Card IDs shown to player (drawn or peeked)
	FreeTakeCount  int      `json:"freeTakeCount" ts:"number"`     // Number of cards to take for free (mandatory, 0 = optional)
	MaxBuyCount    int      `json:"maxBuyCount" ts:"number"`       // Maximum cards to buy (optional, 0 = no buying allowed)
	CardBuyCost    int      `json:"cardBuyCost" ts:"number"`       // Cost per card when buying (typically 3 MC, 0 if no buying)
	Source         string   `json:"source" ts:"string"`            // Card ID or action that triggered this
}

// ForcedFirstAction represents an action that must be completed as the player's first turn action
// Examples: Tharsis Republic must place a city as their first action
type ForcedFirstAction struct {
	ActionType    string `json:"actionType" ts:"string"`    // Type of action: "city_placement", "card_draw", etc.
	CorporationID string `json:"corporationId" ts:"string"` // Corporation that requires this action
	Completed     bool   `json:"completed" ts:"boolean"`    // Whether the forced action has been completed
	Description   string `json:"description" ts:"string"`   // Human-readable description for UI
}

// Player represents a player in the game with service references to features
type Player struct {
	// Metadata
	ID              string   `json:"id" ts:"string"`
	Name            string   `json:"name" ts:"string"`
	Corporation     *Card    `json:"corporation" ts:"CardDto | null"`
	Cards           []string `json:"cards" ts:"string[]"`
	PlayedCards     []string `json:"playedCards" ts:"string[]"`
	TerraformRating int      `json:"terraformRating" ts:"number"` // Simple field, increases via events
	VictoryPoints   int      `json:"victoryPoints" ts:"number"`
	IsConnected     bool     `json:"isConnected" ts:"boolean"`

	// Card effects and actions
	Effects []PlayerEffect `json:"effects" ts:"PlayerEffect[]"` // Active ongoing passive effects from played cards
	Actions []PlayerAction `json:"actions" ts:"PlayerAction[]"` // Available actions from played cards with manual triggers

	// Selection phases (non-feature state)
	SelectStartingCardsPhase *SelectStartingCardsPhase `json:"selectStartingCardsPhase" ts:"selectStartingCardsPhase | null"`

	// Card selection - nullable, exists only when player needs to select cards
	PendingCardSelection     *PendingCardSelection     `json:"pendingCardSelection" ts:"PendingCardSelection | null"`         // Current active card selection (sell patents, etc.)
	PendingCardDrawSelection *PendingCardDrawSelection `json:"pendingCardDrawSelection" ts:"PendingCardDrawSelection | null"` // Current active card draw/peek selection

	// Forced first action - nullable, exists only when corporation requires specific first turn action
	ForcedFirstAction *ForcedFirstAction `json:"forcedFirstAction" ts:"ForcedFirstAction | null"` // Action that must be taken on first turn (Tharsis city placement, etc.)

	// Resource storage - maps card IDs to resource counts stored on those cards
	ResourceStorage map[string]int `json:"resourceStorage" ts:"Record<string, number>"` // Card ID -> resource count

	// Payment substitutes - alternative resources that can be used as payment for credits
	PaymentSubstitutes []PaymentSubstitute `json:"paymentSubstitutes" ts:"PaymentSubstitute[]"` // Alternative resources usable as payment

	// Feature Services (player-level state management)
	ResourcesService  resources.Service      `json:"-"` // Resources + Production
	ProductionService production.Service     `json:"-"` // Production phase card selection state
	TileQueueService  tiles.TileQueueService `json:"-"` // Tile selection queue
	PlayerTurnService turn.PlayerTurnService `json:"-"` // Passed status + available actions
}

// GetResources retrieves resources via service
func (p *Player) GetResources() (resources.Resources, error) {
	if p.ResourcesService == nil {
		return resources.Resources{}, nil
	}
	return p.ResourcesService.Get(nil)
}

// GetProduction retrieves production via service
func (p *Player) GetProduction() (resources.Production, error) {
	if p.ResourcesService == nil {
		return resources.Production{}, nil
	}
	return p.ResourcesService.GetProduction(nil)
}

// GetPassed retrieves passed status via service
func (p *Player) GetPassed() (bool, error) {
	if p.PlayerTurnService == nil {
		return false, nil
	}
	return p.PlayerTurnService.GetPassed(nil)
}

// GetAvailableActions retrieves available actions via service
func (p *Player) GetAvailableActions() (int, error) {
	if p.PlayerTurnService == nil {
		return 0, nil
	}
	return p.PlayerTurnService.GetAvailableActions(nil)
}

// GetPendingTileSelection retrieves pending tile selection via service
func (p *Player) GetPendingTileSelection() (*tiles.PendingTileSelection, error) {
	if p.TileQueueService == nil {
		return nil, nil
	}
	return p.TileQueueService.GetPendingSelection(nil)
}

// GetPendingTileSelectionQueue retrieves tile selection queue via service
func (p *Player) GetPendingTileSelectionQueue() (*tiles.PendingTileSelectionQueue, error) {
	if p.TileQueueService == nil {
		return nil, nil
	}
	return p.TileQueueService.GetQueue(nil)
}

// GetProductionPhaseState retrieves production phase state via service
func (p *Player) GetProductionPhaseState() (*production.ProductionPhaseState, error) {
	if p.ProductionService == nil {
		return nil, nil
	}
	return p.ProductionService.Get(nil)
}

// GetStartingSelectionCards returns the player's starting card selection, nil if not in that phase
func (p *Player) GetStartingSelectionCards() []string {
	if p.SelectStartingCardsPhase == nil {
		return nil
	}

	return p.SelectStartingCardsPhase.AvailableCards
}

// GetProductionPhaseCards returns the player's production phase card selection via service
func (p *Player) GetProductionPhaseCards() []string {
	if p.ProductionService == nil {
		return nil
	}

	state, err := p.ProductionService.Get(nil)
	if err != nil || state == nil {
		return nil
	}

	return state.AvailableCards
}

// DeepCopy creates a deep copy of the Player
// Note: Services are NOT copied - they should be references to the same instances
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

	// Deep copy resource storage map
	resourceStorageCopy := make(map[string]int)
	for cardID, count := range p.ResourceStorage {
		resourceStorageCopy[cardID] = count
	}

	// Deep copy payment substitutes slice
	var paymentSubstitutesCopy []PaymentSubstitute
	if p.PaymentSubstitutes != nil {
		paymentSubstitutesCopy = make([]PaymentSubstitute, len(p.PaymentSubstitutes))
		copy(paymentSubstitutesCopy, p.PaymentSubstitutes)
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

	// Deep copy pending card draw selection if it exists
	var pendingCardDrawSelectionCopy *PendingCardDrawSelection
	if p.PendingCardDrawSelection != nil {
		// Copy available cards slice
		availableCardsCopy := make([]string, len(p.PendingCardDrawSelection.AvailableCards))
		copy(availableCardsCopy, p.PendingCardDrawSelection.AvailableCards)

		pendingCardDrawSelectionCopy = &PendingCardDrawSelection{
			AvailableCards: availableCardsCopy,
			FreeTakeCount:  p.PendingCardDrawSelection.FreeTakeCount,
			MaxBuyCount:    p.PendingCardDrawSelection.MaxBuyCount,
			CardBuyCost:    p.PendingCardDrawSelection.CardBuyCost,
			Source:         p.PendingCardDrawSelection.Source,
		}
	}

	// Deep copy forced first action if it exists
	var forcedFirstActionCopy *ForcedFirstAction
	if p.ForcedFirstAction != nil {
		forcedFirstActionCopy = &ForcedFirstAction{
			ActionType:    p.ForcedFirstAction.ActionType,
			CorporationID: p.ForcedFirstAction.CorporationID,
			Completed:     p.ForcedFirstAction.Completed,
			Description:   p.ForcedFirstAction.Description,
		}
	}

	// Deep copy corporation if it exists
	var corporationCopy *Card
	if p.Corporation != nil {
		corpCopy := p.Corporation.DeepCopy()
		corporationCopy = &corpCopy
	}

	return &Player{
		ID:                       p.ID,
		Name:                     p.Name,
		Corporation:              corporationCopy,
		Cards:                    cardsCopy,
		TerraformRating:          p.TerraformRating,
		PlayedCards:              playedCardsCopy,
		VictoryPoints:            p.VictoryPoints,
		IsConnected:              p.IsConnected,
		Effects:                  effectsCopy,
		Actions:                  actionsCopy,
		SelectStartingCardsPhase: startingSelectionCopy,
		PendingCardSelection:     pendingCardSelectionCopy,
		PendingCardDrawSelection: pendingCardDrawSelectionCopy,
		ForcedFirstAction:        forcedFirstActionCopy,
		ResourceStorage:          resourceStorageCopy,
		PaymentSubstitutes:       paymentSubstitutesCopy,
		// Services are NOT copied - they're references to the same instances
		ResourcesService:  p.ResourcesService,
		ProductionService: p.ProductionService,
		TileQueueService:  p.TileQueueService,
		PlayerTurnService: p.PlayerTurnService,
	}
}
