package player

import "terraforming-mars-backend/internal/game/shared"

import (
	"sync"
	"terraforming-mars-backend/internal/events"
)

// Selection manages player-specific card selection state
type Selection struct {
	mu                       sync.RWMutex
	selectStartingCardsPhase *SelectStartingCardsPhase
	pendingCardSelection     *PendingCardSelection
	pendingCardDrawSelection *PendingCardDrawSelection
	eventBus                 *events.EventBusImpl
	gameID                   string
	playerID                 string
}

func newSelection(eventBus *events.EventBusImpl, gameID, playerID string) *Selection {
	return &Selection{
		eventBus: eventBus,
		gameID:   gameID,
		playerID: playerID,
	}
}

func (s *Selection) GetSelectStartingCardsPhase() *SelectStartingCardsPhase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectStartingCardsPhase
}

func (s *Selection) SetSelectStartingCardsPhase(phase *SelectStartingCardsPhase) {
	s.mu.Lock()
	s.selectStartingCardsPhase = phase
	s.mu.Unlock()

	if s.eventBus != nil {
	}
}

func (s *Selection) GetPendingCardSelection() *PendingCardSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingCardSelection
}

func (s *Selection) SetPendingCardSelection(selection *PendingCardSelection) {
	s.mu.Lock()
	s.pendingCardSelection = selection
	s.mu.Unlock()

	if s.eventBus != nil {
	}
}

func (s *Selection) GetPendingCardDrawSelection() *PendingCardDrawSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingCardDrawSelection
}

func (s *Selection) SetPendingCardDrawSelection(selection *PendingCardDrawSelection) {
	s.mu.Lock()
	s.pendingCardDrawSelection = selection
	s.mu.Unlock()

	if s.eventBus != nil {
	}
}

// PendingCardSelection represents a pending card selection
type PendingCardSelection struct {
	AvailableCards []string
	MinCards       int
	MaxCards       int
	CardCosts      map[string]int
	CardRewards    map[string]int
	Source         string
}

// PendingCardDrawSelection represents a pending card draw/peek/take/buy action
type PendingCardDrawSelection struct {
	AvailableCards      []string
	FreeTakeCount       int
	MaxBuyCount         int
	CardBuyCost         int
	Source              string
	SourceCardID        string // Card that triggered this selection (for card actions)
	SourceBehaviorIndex int    // Behavior index of the card action
}

// SelectStartingCardsPhase represents the starting cards selection phase state
type SelectStartingCardsPhase struct {
	AvailableCards        []string
	AvailableCorporations []string
	SelectionComplete     bool
}

// ProductionPhase represents the production phase state for a player
type ProductionPhase struct {
	AvailableCards    []string
	SelectionComplete bool
	BeforeResources   shared.Resources
	AfterResources    shared.Resources
	EnergyConverted   int
	CreditsIncome     int
}

// TileCompletionCallback stores info about what to call when tile placement completes
type TileCompletionCallback struct {
	Type string
	Data map[string]interface{}
}

// PendingTileSelection represents a pending tile placement action
type PendingTileSelection struct {
	TileType       string
	AvailableHexes []string
	Source         string
	OnComplete     *TileCompletionCallback
}

// PendingTileSelectionQueue represents a queue of tile placements
type PendingTileSelectionQueue struct {
	Items      []string
	Source     string
	OnComplete *TileCompletionCallback
}

// ForcedFirstAction represents an action that must be completed as first action
type ForcedFirstAction struct {
	ActionType    string
	CorporationID string
	Source        string
	Completed     bool
	Description   string
}
