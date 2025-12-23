package action

import (
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CreateAndCachePlayerAward creates a PlayerAward with event listeners and initial state,
// then caches it in the player's awards. Called during game initialization.
func CreateAndCachePlayerAward(
	awardType shared.AwardType,
	p *player.Player,
	g *game.Game,
) *player.PlayerAward {
	// 1. Create PlayerAward data holder
	pa := player.NewPlayerAward(awardType)

	// 2. Register event listeners for state recalculation
	registerPlayerAwardEventListeners(pa, awardType, p, g)

	// 3. Calculate initial state
	recalculatePlayerAward(pa, awardType, p, g)

	// 4. Cache in Player
	p.Awards().Add(awardType, pa)

	return pa
}

// registerPlayerAwardEventListeners registers all event listeners on a PlayerAward.
func registerPlayerAwardEventListeners(
	pa *player.PlayerAward,
	awardType shared.AwardType,
	p *player.Player,
	g *game.Game,
) {
	eventBus := g.EventBus()

	// When player resources change, recalculate affordability
	subID1 := events.Subscribe(eventBus, func(event events.ResourcesChangedEvent) {
		if event.PlayerID == p.ID() {
			recalculatePlayerAward(pa, awardType, p, g)
		}
	})
	pa.AddUnsubscriber(func() { eventBus.Unsubscribe(subID1) })

	// When any award is funded, recalculate (cost changes, might be this one, or max reached)
	subID2 := events.Subscribe(eventBus, func(event events.AwardFundedEvent) {
		recalculatePlayerAward(pa, awardType, p, g)
	})
	pa.AddUnsubscriber(func() { eventBus.Unsubscribe(subID2) })
}

// recalculatePlayerAward recalculates and updates PlayerAward state.
func recalculatePlayerAward(
	pa *player.PlayerAward,
	awardType shared.AwardType,
	p *player.Player,
	g *game.Game,
) {
	state := CalculateAwardState(awardType, p, g)
	pa.UpdateState(state)
}

// InitializePlayerAwards creates and caches all PlayerAward instances for a player.
// Called when a player joins a game or when the game starts.
func InitializePlayerAwards(
	p *player.Player,
	g *game.Game,
) {
	for _, info := range game.AllAwards {
		CreateAndCachePlayerAward(info.Type, p, g)
	}
}
