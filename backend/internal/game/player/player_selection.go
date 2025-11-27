package player

import "terraforming-mars-backend/internal/game/shared"

import (
	"sync"

)

// Selection manages player-specific card selection state
type Selection struct {
	mu                       sync.RWMutex
	selectStartingCardsPhase *SelectStartingCardsPhase
	pendingCardSelection     *PendingCardSelection
	pendingCardDrawSelection *PendingCardDrawSelection
}

func newSelection() *Selection {
	return &Selection{}
}

func (s *Selection) GetSelectStartingCardsPhase() *SelectStartingCardsPhase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectStartingCardsPhase
}

func (s *Selection) SetSelectStartingCardsPhase(phase *SelectStartingCardsPhase) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectStartingCardsPhase = phase
}

func (s *Selection) GetPendingCardSelection() *PendingCardSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingCardSelection
}

func (s *Selection) SetPendingCardSelection(selection *PendingCardSelection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingCardSelection = selection
}

func (s *Selection) GetPendingCardDrawSelection() *PendingCardDrawSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingCardDrawSelection
}

func (s *Selection) SetPendingCardDrawSelection(selection *PendingCardDrawSelection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingCardDrawSelection = selection
}

// ==================== Phase State Types ====================

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
	AvailableCards []string
	FreeTakeCount  int
	MaxBuyCount    int
	CardBuyCost    int
	Source         string
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
	BeforeResources shared.Resources
	AfterResources shared.Resources
	EnergyConverted   int
	CreditsIncome     int
}

// PendingTileSelection represents a pending tile placement action
type PendingTileSelection struct {
	TileType       string
	AvailableHexes []string
	Source         string
}

// PendingTileSelectionQueue represents a queue of tile placements
type PendingTileSelectionQueue struct {
	Items  []string
	Source string
}

// ForcedFirstAction represents an action that must be completed as first action
type ForcedFirstAction struct {
	ActionType    string
	CorporationID string
	Source        string
	Completed     bool
	Description   string
}
