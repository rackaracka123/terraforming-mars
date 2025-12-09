package game

import (
	"sync"
	"time"

	"terraforming-mars-backend/internal/game/playability"
)

// PlayabilityState stores cached playability information for all players in a game.
// This state is maintained by PlayabilityManager which subscribes to domain events
// and automatically updates playability when game state changes.
type PlayabilityState struct {
	mu           sync.RWMutex
	PlayerStates map[string]*PlayerPlayabilityState
}

// NewPlayabilityState creates a new empty playability state
func NewPlayabilityState() *PlayabilityState {
	return &PlayabilityState{
		PlayerStates: make(map[string]*PlayerPlayabilityState),
	}
}

// PlayerPlayabilityState contains all playability information for a single player
type PlayerPlayabilityState struct {
	HandPlayability   map[string]CardPlayabilityResult // cardID -> playability result
	ActionPlayability []ActionPlayabilityResult        // indexed by action position
	LastUpdated       time.Time                        // when this state was last calculated
}

// NewPlayerPlayabilityState creates a new empty player playability state
func NewPlayerPlayabilityState() *PlayerPlayabilityState {
	return &PlayerPlayabilityState{
		HandPlayability:   make(map[string]CardPlayabilityResult),
		ActionPlayability: []ActionPlayabilityResult{},
		LastUpdated:       time.Now(),
	}
}

// CardPlayabilityResult stores playability information for a card in hand
type CardPlayabilityResult struct {
	CardID     string
	IsPlayable bool
	Errors     []playability.ValidationError
}

// ActionPlayabilityResult stores playability information for a card action
type ActionPlayabilityResult struct {
	ActionIndex         int
	IsAffordable        bool
	PlayableChoices     []int
	ChoicePlayabilities []playability.ChoicePlayability
	Errors              []playability.ValidationError
}

// GetPlayerState returns the playability state for a player (read-only)
func (ps *PlayabilityState) GetPlayerState(playerID string) *PlayerPlayabilityState {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.PlayerStates[playerID]
}

// SetPlayerState updates the playability state for a player
func (ps *PlayabilityState) SetPlayerState(playerID string, state *PlayerPlayabilityState) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.PlayerStates[playerID] = state
}

// InvalidatePlayer marks a player's playability state as needing recalculation
func (ps *PlayabilityState) InvalidatePlayer(playerID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.PlayerStates, playerID)
}

// InvalidateAll clears all playability state (used when multiple players affected)
func (ps *PlayabilityState) InvalidateAll() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.PlayerStates = make(map[string]*PlayerPlayabilityState)
}
