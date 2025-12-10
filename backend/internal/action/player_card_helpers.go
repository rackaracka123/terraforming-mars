package action

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
)

// CreateAndCachePlayerCard creates a PlayerCard with event listeners and initial state,
// then caches it in the player's hand. This is called by actions when adding a card to hand.
func CreateAndCachePlayerCard(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) *player.PlayerCard {
	// 1. Create PlayerCard data holder
	pc := player.NewPlayerCard(card)

	// 2. Register event listeners for state recalculation
	registerPlayerCardEventListeners(pc, p, g, cardRegistry)

	// 3. Calculate initial state
	recalculatePlayerCard(pc, p, g, cardRegistry)

	// 4. Cache in Hand
	p.Hand().AddPlayerCard(card.ID, pc)

	return pc
}

// registerPlayerCardEventListeners registers all event listeners on a PlayerCard.
// Stores unsubscribe functions in PlayerCard for cleanup when card is removed.
func registerPlayerCardEventListeners(
	pc *player.PlayerCard,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) {
	eventBus := g.EventBus()

	// When player resources change, recalculate affordability
	subID1 := events.Subscribe(eventBus, func(event events.ResourcesChangedEvent) {
		if event.PlayerID == p.ID() {
			recalculatePlayerCard(pc, p, g, cardRegistry)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID1) })

	// When temperature changes, recalculate requirements
	subID2 := events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		recalculatePlayerCard(pc, p, g, cardRegistry)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID2) })

	// When oxygen changes, recalculate requirements
	subID3 := events.Subscribe(eventBus, func(event events.OxygenChangedEvent) {
		recalculatePlayerCard(pc, p, g, cardRegistry)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID3) })

	// When oceans change, recalculate requirements
	subID4 := events.Subscribe(eventBus, func(event events.OceansChangedEvent) {
		recalculatePlayerCard(pc, p, g, cardRegistry)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID4) })

	// When player effects change (requirement modifiers), recalculate cost
	subID5 := events.Subscribe(eventBus, func(event events.PlayerEffectsChangedEvent) {
		if event.PlayerID == p.ID() {
			recalculatePlayerCard(pc, p, g, cardRegistry)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID5) })

	// When game phase changes, recalculate state (affects phase validation)
	subID6 := events.Subscribe(eventBus, func(event events.GamePhaseChangedEvent) {
		if event.GameID == g.ID() {
			recalculatePlayerCard(pc, p, g, cardRegistry)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID6) })

	// When general game state changes, recalculate availability
	subID7 := events.Subscribe(eventBus, func(event events.GameStateChangedEvent) {
		recalculatePlayerCard(pc, p, g, cardRegistry)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID7) })
}

// recalculatePlayerCard recalculates and updates PlayerCard state.
// Called on initial creation and by event listeners.
func recalculatePlayerCard(
	pc *player.PlayerCard,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) {
	// Type assert card from any to *gamecards.Card
	card, ok := pc.Card().(*gamecards.Card)
	if !ok {
		// Should never happen if architecture is followed correctly
		return
	}
	state := CalculatePlayerCardState(card, p, g, cardRegistry)
	pc.UpdateState(state)
}
