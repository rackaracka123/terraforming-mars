package store

import (
	"fmt"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// InitializeStore creates and configures the application store
func InitializeStore(eventBus events.EventBus) (*Store, error) {
	// Load card registry from assets
	logger.Info("📋 Loading card data from assets...")
	cardRegistry, err := cards.LoadCardsFromAssets()
	if err != nil {
		return nil, fmt.Errorf("failed to load cards: %w", err)
	}

	logger.Info("✅ Cards loaded successfully",
		zap.Int("total_cards", len(cardRegistry.Cards)),
		zap.Int("corporations", len(cardRegistry.Corporations)),
		zap.Int("starting_deck", len(cardRegistry.StartingDeck)))

	// Create middleware stack
	middleware := []Middleware{
		LoggingMiddleware,
		EventMiddleware(eventBus),
		ValidationMiddleware,
	}

	// Create initial application state with card registry
	initialState := NewApplicationState().WithCardRegistry(cardRegistry)

	// Create store with game reducer and middleware
	store := NewStore(GameReducer, eventBus, middleware...)

	// Set the initial state with cards loaded
	store.state = initialState

	return store, nil
}
