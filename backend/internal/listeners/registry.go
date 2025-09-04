package listeners

import (
	"context"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/listeners/card_effects"
	"terraforming-mars-backend/internal/listeners/starting_cards"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Registry manages the registration of all event listeners
type Registry struct {
	eventBus     *events.InMemoryEventBus
	gameRepo     *repository.GameRepository
	cardRegistry *cards.CardHandlerRegistry
}

// NewRegistry creates a new listener registry
func NewRegistry(eventBus *events.InMemoryEventBus, gameRepo *repository.GameRepository, cardRegistry *cards.CardHandlerRegistry) *Registry {
	return &Registry{
		eventBus:     eventBus,
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
	}
}

// RegisterAllListeners registers all feature listeners with the event bus
func (r *Registry) RegisterAllListeners() {
	log := logger.Get()
	log.Info("Registering all event listeners")

	// Register starting cards feature listeners
	r.registerStartingCardsListeners()
	
	// Register card effects listeners
	r.registerCardEffectsListeners()

	log.Info("All event listeners registered successfully")
}

// registerStartingCardsListeners registers listeners for the starting cards feature
func (r *Registry) registerStartingCardsListeners() {
	startingCardsListener := starting_cards.NewListener(r.gameRepo, r.eventBus)
	
	// Register for game started events to deal starting cards
	r.eventBus.Subscribe(events.EventTypeGameStarted, func(ctx context.Context, event events.Event) error {
		return startingCardsListener.OnGameStarted(ctx, event)
	})

	log := logger.Get()
	log.Info("Starting cards listeners registered", 
		zap.String("feature", "starting_cards"),
		zap.Int("listeners", 1),
	)
}

// registerCardEffectsListeners registers listeners for card effects and interactions
func (r *Registry) registerCardEffectsListeners() {
	cardEffectsListener := card_effects.NewCardEffectsListener(r.gameRepo, r.cardRegistry)
	
	// Register for card played events to handle card interactions
	r.eventBus.Subscribe(events.EventTypeCardPlayed, func(ctx context.Context, event events.Event) error {
		return cardEffectsListener.HandleCardPlayed(event)
	})

	log := logger.Get()
	log.Info("Card effects listeners registered", 
		zap.String("feature", "card_effects"),
		zap.Int("listeners", 1),
	)
}