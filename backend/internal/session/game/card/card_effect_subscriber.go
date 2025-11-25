package card

import (
	"context"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game/player"
)

// CardEffectSubscriber manages subscriptions for card passive effects to domain events
type CardEffectSubscriber interface {
	// SubscribeCardEffects subscribes all passive effects for a card when it's played
	SubscribeCardEffects(ctx context.Context, p *player.Player, cardID string, card *Card) error

	// UnsubscribeCardEffects unsubscribes all effects for a card (cleanup on card removal)
	UnsubscribeCardEffects(cardID string) error
}

// CardEffectSubscriberImpl implements CardEffectSubscriber
type CardEffectSubscriberImpl struct {
	eventBus *events.EventBusImpl
	cardRepo Repository
	// TODO: Full implementation needs refactoring for new architecture
	// Event handlers need access to session to fetch player/game data when events trigger
	// This requires redesigning how handlers access data - possibly passing session reference
	// or using a different pattern for event-driven effects

	// Track subscription IDs for cleanup
	subscriptions map[string][]events.SubscriptionID // cardID -> list of subscription IDs
}

// NewCardEffectSubscriber creates a new card effect subscriber
func NewCardEffectSubscriber(
	eventBus *events.EventBusImpl,
	cardRepository Repository,
) CardEffectSubscriber {
	return &CardEffectSubscriberImpl{
		eventBus:      eventBus,
		cardRepo:      cardRepository,
		subscriptions: make(map[string][]events.SubscriptionID),
	}
}

// SubscribeCardEffects subscribes passive effects based on card behaviors
// TODO: Full implementation pending architecture refactoring
func (ces *CardEffectSubscriberImpl) SubscribeCardEffects(ctx context.Context, p *player.Player, cardID string, card *Card) error {
	log := logger.WithGameContext(p.GameID, p.ID)
	log.Warn("⚠️  CardEffectSubscriber not yet fully implemented in new architecture - passive card effects will not trigger")
	// Passive effects like "gain 2 MC when any city is placed" won't work until this is implemented
	return nil
}

// UnsubscribeCardEffects unsubscribes all effects for a card
func (ces *CardEffectSubscriberImpl) UnsubscribeCardEffects(cardID string) error {
	subIDs, exists := ces.subscriptions[cardID]
	if !exists {
		return nil // No subscriptions for this card
	}

	log := logger.Get()
	log.Info("🗑️ Unsubscribing card effects (stub)")

	// Unsubscribe all
	for _, subID := range subIDs {
		ces.eventBus.Unsubscribe(subID)
	}

	// Remove from tracking
	delete(ces.subscriptions, cardID)

	return nil
}
