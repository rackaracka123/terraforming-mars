package player

import (
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/shared"
)

// PlayerMilestones manages the player's cached milestone eligibility state.
// Global claim state lives in Game.Milestones().
// This component holds per-player eligibility calculations.
type PlayerMilestones struct {
	mu sync.RWMutex

	// Cached PlayerMilestone instances (keyed by MilestoneType)
	milestones map[shared.MilestoneType]*PlayerMilestone

	// Infrastructure
	eventBus *events.EventBusImpl
	gameID   string
	playerID string
}

// newPlayerMilestones creates a new PlayerMilestones component.
func newPlayerMilestones(eventBus *events.EventBusImpl, gameID, playerID string) *PlayerMilestones {
	return &PlayerMilestones{
		milestones: make(map[shared.MilestoneType]*PlayerMilestone),
		eventBus:   eventBus,
		gameID:     gameID,
		playerID:   playerID,
	}
}

// Get returns a cached PlayerMilestone instance.
// Returns nil, false if not cached (actions must create it).
func (pm *PlayerMilestones) Get(milestoneType shared.MilestoneType) (*PlayerMilestone, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	m, exists := pm.milestones[milestoneType]
	return m, exists
}

// Add adds a PlayerMilestone to the cache.
// Called by actions after creating PlayerMilestone with state and event listeners.
func (pm *PlayerMilestones) Add(milestoneType shared.MilestoneType, m *PlayerMilestone) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.milestones[milestoneType] = m
}

// GetAll returns all cached PlayerMilestone instances.
func (pm *PlayerMilestones) GetAll() map[shared.MilestoneType]*PlayerMilestone {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	result := make(map[shared.MilestoneType]*PlayerMilestone, len(pm.milestones))
	for k, v := range pm.milestones {
		result[k] = v
	}
	return result
}

// Cleanup unsubscribes all event listeners.
// Called when player is removed from game.
func (pm *PlayerMilestones) Cleanup() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, m := range pm.milestones {
		m.Cleanup()
	}
	pm.milestones = make(map[shared.MilestoneType]*PlayerMilestone)
}
