package action

import (
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
)

// GlobalSubscriber manages game-wide event subscriptions
type GlobalSubscriber struct {
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewGlobalSubscriber creates a new global subscriber
func NewGlobalSubscriber(cardRegistry cards.CardRegistry, logger *zap.Logger) *GlobalSubscriber {
	return &GlobalSubscriber{
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// SetupGlobalSubscribers sets up all global event subscriptions for a game
func (s *GlobalSubscriber) SetupGlobalSubscribers(g *game.Game) {
	s.subscribeRequirementModifierRecalculation(g)
}

// subscribeRequirementModifierRecalculation subscribes to CardHandUpdatedEvent
// to recalculate requirement modifiers when a player's hand changes.
// This is necessary for tag-based and card-type discounts that depend on cards in hand.
func (s *GlobalSubscriber) subscribeRequirementModifierRecalculation(g *game.Game) {
	calculator := gamecards.NewRequirementModifierCalculator(s.cardRegistry)
	log := s.logger.With(zap.String("game_id", g.ID()))

	events.Subscribe(g.EventBus(), func(event events.CardHandUpdatedEvent) {
		// Only process if event is for this game
		if event.GameID != g.ID() {
			return
		}

		// Get the player whose hand changed
		player, err := g.GetPlayer(event.PlayerID)
		if err != nil {
			log.Error("Failed to get player for requirement modifier recalculation",
				zap.String("player_id", event.PlayerID),
				zap.Error(err))
			return
		}

		// Recalculate modifiers based on updated hand
		modifiers := calculator.Calculate(player)
		player.Effects().SetRequirementModifiers(modifiers)

		log.Debug("📊 Recalculated requirement modifiers on hand change",
			zap.String("player_id", event.PlayerID),
			zap.Int("modifier_count", len(modifiers)),
			zap.Int("hand_size", len(event.CardIDs)))
	})

	log.Debug("📬 Subscribed to CardHandUpdatedEvent for requirement modifier recalculation")
}
