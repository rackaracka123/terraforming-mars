package selection

import (
	"sync"

	"terraforming-mars-backend/internal/session/types"
)

// ProductionPhase represents the production phase state for a player
// Note: This is managed by Game, not by Player.Selection
type ProductionPhase struct {
	AvailableCards    []string        // Cards available for selection
	SelectionComplete bool            // Whether player has completed their selection
	BeforeResources   types.Resources // Resources before production was applied
	AfterResources    types.Resources // Resources after production was applied
	EnergyConverted   int             // Amount of energy converted to heat
	CreditsIncome     int             // Total credits income (production + TR)
}

// SelectStartingCardsPhase represents the starting cards selection phase state
type SelectStartingCardsPhase struct {
	AvailableCards        []string // Cards available for selection
	AvailableCorporations []string // Corporations available for selection
	SelectionComplete     bool     // Whether player has completed their selection
}

// PendingCardSelection represents a pending card selection for a player
type PendingCardSelection struct {
	AvailableCards []string       // Cards available for selection
	MinCards       int            // Minimum number of cards to select
	MaxCards       int            // Maximum number of cards to select
	CardCosts      map[string]int // Cost per card
	CardRewards    map[string]int // Reward per card (for sell patents)
	Source         string         // Source of this selection (e.g., "sell-patents", "trade")
}

// PendingCardDrawSelection represents a pending card draw/peek/take/buy action from card effects
type PendingCardDrawSelection struct {
	AvailableCards []string // card.Card IDs shown to player (drawn or peeked)
	FreeTakeCount  int      // Number of cards to take for free (mandatory, 0 = optional)
	MaxBuyCount    int      // Maximum cards to buy (optional, 0 = no buying allowed)
	CardBuyCost    int      // Cost per card when buying (typically 3 MC, 0 if no buying)
	Source         string   // card.Card ID or action that triggered this
}

// Selection manages player-specific card selection state.
// This includes temporary state that exists during specific game phases.
// Thread-safe with its own mutex.
type Selection struct {
	mu sync.RWMutex

	// Card selection states
	selectStartingCardsPhase *SelectStartingCardsPhase
	pendingCardSelection     *PendingCardSelection
	pendingCardDrawSelection *PendingCardDrawSelection
}

// NewSelection creates a new Selection component with no active selection states.
func NewSelection() *Selection {
	return &Selection{}
}

// ==================== Starting Cards Selection ====================

// GetSelectStartingCardsPhase returns the starting cards selection phase state.
func (s *Selection) GetSelectStartingCardsPhase() *SelectStartingCardsPhase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectStartingCardsPhase
}

// SetSelectStartingCardsPhase sets the starting cards selection phase state.
func (s *Selection) SetSelectStartingCardsPhase(phase *SelectStartingCardsPhase) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectStartingCardsPhase = phase
}

// ==================== Card Selection ====================

// GetPendingCardSelection returns the pending card selection state.
func (s *Selection) GetPendingCardSelection() *PendingCardSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingCardSelection
}

// SetPendingCardSelection sets the pending card selection state.
func (s *Selection) SetPendingCardSelection(selection *PendingCardSelection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingCardSelection = selection
}

// GetPendingCardDrawSelection returns the pending card draw selection state.
func (s *Selection) GetPendingCardDrawSelection() *PendingCardDrawSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingCardDrawSelection
}

// SetPendingCardDrawSelection sets the pending card draw selection state.
func (s *Selection) SetPendingCardDrawSelection(selection *PendingCardDrawSelection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingCardDrawSelection = selection
}

// ==================== Utilities ====================

// ClearAll clears all card selection states.
func (s *Selection) ClearAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectStartingCardsPhase = nil
	s.pendingCardSelection = nil
	s.pendingCardDrawSelection = nil
}

// DeepCopy creates a deep copy of the Selection component.
func (s *Selection) DeepCopy() *Selection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newSelection := &Selection{}

	if s.selectStartingCardsPhase != nil {
		ssp := *s.selectStartingCardsPhase
		newSelection.selectStartingCardsPhase = &ssp
	}

	if s.pendingCardSelection != nil {
		pcs := *s.pendingCardSelection
		newSelection.pendingCardSelection = &pcs
	}

	if s.pendingCardDrawSelection != nil {
		pcds := *s.pendingCardDrawSelection
		newSelection.pendingCardDrawSelection = &pcds
	}

	return newSelection
}
