package player

import (
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/shared"
)

// PlayerAwards manages the player's cached award eligibility state.
// Global funding state lives in Game.Awards().
// This component holds per-player eligibility calculations.
type PlayerAwards struct {
	mu sync.RWMutex

	// Cached PlayerAward instances (keyed by AwardType)
	awards map[shared.AwardType]*PlayerAward

	// Infrastructure
	eventBus *events.EventBusImpl
	gameID   string
	playerID string
}

// newPlayerAwards creates a new PlayerAwards component.
func newPlayerAwards(eventBus *events.EventBusImpl, gameID, playerID string) *PlayerAwards {
	return &PlayerAwards{
		awards:   make(map[shared.AwardType]*PlayerAward),
		eventBus: eventBus,
		gameID:   gameID,
		playerID: playerID,
	}
}

// Get returns a cached PlayerAward instance.
// Returns nil, false if not cached (actions must create it).
func (pa *PlayerAwards) Get(awardType shared.AwardType) (*PlayerAward, bool) {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	a, exists := pa.awards[awardType]
	return a, exists
}

// Add adds a PlayerAward to the cache.
// Called by actions after creating PlayerAward with state and event listeners.
func (pa *PlayerAwards) Add(awardType shared.AwardType, a *PlayerAward) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.awards[awardType] = a
}

// GetAll returns all cached PlayerAward instances.
func (pa *PlayerAwards) GetAll() map[shared.AwardType]*PlayerAward {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	result := make(map[shared.AwardType]*PlayerAward, len(pa.awards))
	for k, v := range pa.awards {
		result[k] = v
	}
	return result
}

// Cleanup unsubscribes all event listeners.
// Called when player is removed from game.
func (pa *PlayerAwards) Cleanup() {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	for _, a := range pa.awards {
		a.Cleanup()
	}
	pa.awards = make(map[shared.AwardType]*PlayerAward)
}
