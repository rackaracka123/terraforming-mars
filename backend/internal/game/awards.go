package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/shared"
)

// Award constants
const (
	MaxFundedAwards = 3 // Maximum awards that can be funded per game
	AwardFirstVP    = 5 // VP awarded for first place
	AwardSecondVP   = 2 // VP awarded for second place
)

// AwardFundingCosts returns the cost to fund awards based on how many have been funded
// First award: 8 MC, Second: 14 MC, Third: 20 MC
var AwardFundingCosts = []int{8, 14, 20}

// AwardInfo contains display information about an award
type AwardInfo struct {
	Type        shared.AwardType
	Name        string
	Description string
}

// AllAwards returns all available award types with their info
var AllAwards = []AwardInfo{
	{Type: shared.AwardLandlord, Name: "Landlord", Description: "Most tiles on Mars"},
	{Type: shared.AwardBanker, Name: "Banker", Description: "Highest MC production"},
	{Type: shared.AwardScientist, Name: "Scientist", Description: "Most science tags in play"},
	{Type: shared.AwardThermalist, Name: "Thermalist", Description: "Most heat resources"},
	{Type: shared.AwardMiner, Name: "Miner", Description: "Most steel and titanium resources"},
}

// FundedAward represents an award that has been funded by a player
type FundedAward struct {
	Type           shared.AwardType
	FundedByPlayer string
	FundingOrder   int // 0, 1, or 2 (order in which it was funded)
	FundingCost    int
	FundedAt       time.Time
}

// Awards manages the award state for a game
type Awards struct {
	mu       sync.RWMutex
	gameID   string
	funded   []FundedAward
	eventBus *events.EventBusImpl
}

// NewAwards creates a new Awards instance
func NewAwards(gameID string, eventBus *events.EventBusImpl) *Awards {
	return &Awards{
		gameID:   gameID,
		funded:   make([]FundedAward, 0, MaxFundedAwards),
		eventBus: eventBus,
	}
}

// FundedAwards returns a copy of all funded awards
func (a *Awards) FundedAwards() []FundedAward {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]FundedAward, len(a.funded))
	copy(result, a.funded)
	return result
}

// IsFunded returns true if the specified award has been funded
func (a *Awards) IsFunded(awardType shared.AwardType) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, funded := range a.funded {
		if funded.Type == awardType {
			return true
		}
	}
	return false
}

// IsFundedBy returns true if the specified award was funded by the given player
func (a *Awards) IsFundedBy(awardType shared.AwardType, playerID string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, funded := range a.funded {
		if funded.Type == awardType && funded.FundedByPlayer == playerID {
			return true
		}
	}
	return false
}

// CanFundMore returns true if more awards can still be funded (less than 3 funded)
func (a *Awards) CanFundMore() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.funded) < MaxFundedAwards
}

// FundedCount returns the number of awards that have been funded
func (a *Awards) FundedCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.funded)
}

// GetCurrentFundingCost returns the cost to fund the next award
func (a *Awards) GetCurrentFundingCost() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	count := len(a.funded)
	if count >= MaxFundedAwards {
		return 0 // No more can be funded
	}
	return AwardFundingCosts[count]
}

// FundAward funds an award for a player
// Returns an error if the award is already funded or max awards reached
// Publishes AwardFundedEvent after successful funding
func (a *Awards) FundAward(ctx context.Context, awardType shared.AwardType, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	a.mu.Lock()

	if len(a.funded) >= MaxFundedAwards {
		a.mu.Unlock()
		return fmt.Errorf("maximum awards (%d) already funded", MaxFundedAwards)
	}

	for _, funded := range a.funded {
		if funded.Type == awardType {
			a.mu.Unlock()
			return fmt.Errorf("award %s is already funded", awardType)
		}
	}

	fundingCost := AwardFundingCosts[len(a.funded)]

	funded := FundedAward{
		Type:           awardType,
		FundedByPlayer: playerID,
		FundingOrder:   len(a.funded),
		FundingCost:    fundingCost,
		FundedAt:       time.Now(),
	}
	a.funded = append(a.funded, funded)

	a.mu.Unlock()

	if a.eventBus != nil {
		events.Publish(a.eventBus, events.AwardFundedEvent{
			GameID:      a.gameID,
			PlayerID:    playerID,
			AwardType:   string(awardType),
			FundingCost: fundingCost,
			Timestamp:   time.Now(),
		})
	}

	return nil
}

// GetAwardInfo returns the info for a specific award type
func GetAwardInfo(awardType shared.AwardType) (AwardInfo, bool) {
	for _, info := range AllAwards {
		if info.Type == awardType {
			return info, true
		}
	}
	return AwardInfo{}, false
}
