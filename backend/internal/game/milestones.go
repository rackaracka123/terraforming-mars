package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/shared"
)

// Milestone constants
const (
	MilestoneClaimCost   = 8 // MC cost to claim a milestone
	MaxClaimedMilestones = 3 // Maximum milestones that can be claimed per game
	MilestoneVP          = 5 // VP awarded for each claimed milestone
)

// MilestoneInfo contains display information about a milestone
type MilestoneInfo struct {
	Type        shared.MilestoneType
	Name        string
	Description string
	Requirement int // The numeric requirement to claim
}

// AllMilestones returns all available milestone types with their info
var AllMilestones = []MilestoneInfo{
	{Type: shared.MilestoneTerraformer, Name: "Terraformer", Description: "Have a Terraform Rating of at least 35", Requirement: 35},
	{Type: shared.MilestoneMayor, Name: "Mayor", Description: "Own at least 3 city tiles", Requirement: 3},
	{Type: shared.MilestoneGardener, Name: "Gardener", Description: "Own at least 3 greenery tiles", Requirement: 3},
	{Type: shared.MilestoneBuilder, Name: "Builder", Description: "Have at least 8 building tags in play", Requirement: 8},
	{Type: shared.MilestonePlanner, Name: "Planner", Description: "Have at least 16 cards in hand", Requirement: 16},
}

// ClaimedMilestone represents a milestone that has been claimed by a player
type ClaimedMilestone struct {
	Type       shared.MilestoneType
	PlayerID   string
	Generation int
	ClaimedAt  time.Time
}

// Milestones manages the milestone state for a game
type Milestones struct {
	mu       sync.RWMutex
	gameID   string
	claimed  []ClaimedMilestone
	eventBus *events.EventBusImpl
}

// NewMilestones creates a new Milestones instance
func NewMilestones(gameID string, eventBus *events.EventBusImpl) *Milestones {
	return &Milestones{
		gameID:   gameID,
		claimed:  make([]ClaimedMilestone, 0, MaxClaimedMilestones),
		eventBus: eventBus,
	}
}

// ================== Getters ==================

// ClaimedMilestones returns a copy of all claimed milestones
func (m *Milestones) ClaimedMilestones() []ClaimedMilestone {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]ClaimedMilestone, len(m.claimed))
	copy(result, m.claimed)
	return result
}

// IsClaimed returns true if the specified milestone has been claimed
func (m *Milestones) IsClaimed(milestoneType shared.MilestoneType) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, claimed := range m.claimed {
		if claimed.Type == milestoneType {
			return true
		}
	}
	return false
}

// IsClaimedBy returns true if the specified milestone was claimed by the given player
func (m *Milestones) IsClaimedBy(milestoneType shared.MilestoneType, playerID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, claimed := range m.claimed {
		if claimed.Type == milestoneType && claimed.PlayerID == playerID {
			return true
		}
	}
	return false
}

// CanClaimMore returns true if more milestones can still be claimed (less than 3 claimed)
func (m *Milestones) CanClaimMore() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.claimed) < MaxClaimedMilestones
}

// ClaimedCount returns the number of milestones that have been claimed
func (m *Milestones) ClaimedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.claimed)
}

// GetClaimedByPlayer returns all milestones claimed by a specific player
func (m *Milestones) GetClaimedByPlayer(playerID string) []ClaimedMilestone {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []ClaimedMilestone
	for _, claimed := range m.claimed {
		if claimed.PlayerID == playerID {
			result = append(result, claimed)
		}
	}
	return result
}

// ================== Mutators ==================

// ClaimMilestone claims a milestone for a player
// Returns an error if the milestone is already claimed or max milestones reached
// Publishes MilestoneClaimedEvent after successful claim
func (m *Milestones) ClaimMilestone(ctx context.Context, milestoneType shared.MilestoneType, playerID string, generation int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()

	// Check if max milestones already claimed
	if len(m.claimed) >= MaxClaimedMilestones {
		m.mu.Unlock()
		return fmt.Errorf("maximum milestones (%d) already claimed", MaxClaimedMilestones)
	}

	// Check if this milestone is already claimed
	for _, claimed := range m.claimed {
		if claimed.Type == milestoneType {
			m.mu.Unlock()
			return fmt.Errorf("milestone %s is already claimed", milestoneType)
		}
	}

	// Claim the milestone
	claimed := ClaimedMilestone{
		Type:       milestoneType,
		PlayerID:   playerID,
		Generation: generation,
		ClaimedAt:  time.Now(),
	}
	m.claimed = append(m.claimed, claimed)

	m.mu.Unlock()

	// Publish event after releasing lock
	if m.eventBus != nil {
		events.Publish(m.eventBus, events.MilestoneClaimedEvent{
			GameID:        m.gameID,
			PlayerID:      playerID,
			MilestoneType: string(milestoneType),
			Timestamp:     time.Now(),
		})
	}

	return nil
}

// GetMilestoneInfo returns the info for a specific milestone type
func GetMilestoneInfo(milestoneType shared.MilestoneType) (MilestoneInfo, bool) {
	for _, info := range AllMilestones {
		if info.Type == milestoneType {
			return info, true
		}
	}
	return MilestoneInfo{}, false
}
