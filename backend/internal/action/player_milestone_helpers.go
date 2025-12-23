package action

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CreateAndCachePlayerMilestone creates a PlayerMilestone with event listeners and initial state,
// then caches it in the player's milestones. Called during game initialization.
func CreateAndCachePlayerMilestone(
	milestoneType shared.MilestoneType,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) *player.PlayerMilestone {
	// 1. Create PlayerMilestone data holder
	pm := player.NewPlayerMilestone(milestoneType)

	// 2. Register event listeners for state recalculation
	registerPlayerMilestoneEventListeners(pm, milestoneType, p, g, cardRegistry)

	// 3. Calculate initial state
	recalculatePlayerMilestone(pm, milestoneType, p, g, cardRegistry)

	// 4. Cache in Player
	p.Milestones().Add(milestoneType, pm)

	return pm
}

// registerPlayerMilestoneEventListeners registers all event listeners on a PlayerMilestone.
func registerPlayerMilestoneEventListeners(
	pm *player.PlayerMilestone,
	milestoneType shared.MilestoneType,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) {
	eventBus := g.EventBus()

	// When player resources change, recalculate affordability
	subID1 := events.Subscribe(eventBus, func(event events.ResourcesChangedEvent) {
		if event.PlayerID == p.ID() {
			recalculatePlayerMilestone(pm, milestoneType, p, g, cardRegistry)
		}
	})
	pm.AddUnsubscriber(func() { eventBus.Unsubscribe(subID1) })

	// When terraform rating changes, recalculate Terraformer milestone
	subID2 := events.Subscribe(eventBus, func(event events.TerraformRatingChangedEvent) {
		if event.PlayerID == p.ID() && milestoneType == shared.MilestoneTerraformer {
			recalculatePlayerMilestone(pm, milestoneType, p, g, cardRegistry)
		}
	})
	pm.AddUnsubscriber(func() { eventBus.Unsubscribe(subID2) })

	// When a tile is placed, recalculate Mayor and Gardener milestones
	subID3 := events.Subscribe(eventBus, func(event events.TilePlacedEvent) {
		if event.PlayerID == p.ID() {
			if milestoneType == shared.MilestoneMayor || milestoneType == shared.MilestoneGardener {
				recalculatePlayerMilestone(pm, milestoneType, p, g, cardRegistry)
			}
		}
	})
	pm.AddUnsubscriber(func() { eventBus.Unsubscribe(subID3) })

	// When a card is played, recalculate Builder milestone (building tags)
	subID4 := events.Subscribe(eventBus, func(event events.CardPlayedEvent) {
		if event.PlayerID == p.ID() && milestoneType == shared.MilestoneBuilder {
			recalculatePlayerMilestone(pm, milestoneType, p, g, cardRegistry)
		}
	})
	pm.AddUnsubscriber(func() { eventBus.Unsubscribe(subID4) })

	// When hand changes, recalculate Planner milestone (cards in hand)
	subID5 := events.Subscribe(eventBus, func(event events.CardHandUpdatedEvent) {
		if event.PlayerID == p.ID() && milestoneType == shared.MilestonePlanner {
			recalculatePlayerMilestone(pm, milestoneType, p, g, cardRegistry)
		}
	})
	pm.AddUnsubscriber(func() { eventBus.Unsubscribe(subID5) })

	// When any milestone is claimed, recalculate (might be this one, or max reached)
	subID6 := events.Subscribe(eventBus, func(event events.MilestoneClaimedEvent) {
		recalculatePlayerMilestone(pm, milestoneType, p, g, cardRegistry)
	})
	pm.AddUnsubscriber(func() { eventBus.Unsubscribe(subID6) })
}

// recalculatePlayerMilestone recalculates and updates PlayerMilestone state.
func recalculatePlayerMilestone(
	pm *player.PlayerMilestone,
	milestoneType shared.MilestoneType,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) {
	state := CalculateMilestoneState(milestoneType, p, g, cardRegistry)
	pm.UpdateState(state)
}

// InitializePlayerMilestones creates and caches all PlayerMilestone instances for a player.
// Called when a player joins a game or when the game starts.
func InitializePlayerMilestones(
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) {
	for _, info := range game.AllMilestones {
		CreateAndCachePlayerMilestone(info.Type, p, g, cardRegistry)
	}
}
