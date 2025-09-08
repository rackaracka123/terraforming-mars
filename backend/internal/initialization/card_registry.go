package initialization

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
)

// RegisterAllCards registers all card handlers with the global registry
func RegisterAllCards() error {
	handlers := getAllCardHandlers()

	for _, handler := range handlers {
		if err := cards.RegisterCardHandler(handler); err != nil {
			return err
		}
	}

	return nil
}

// RegisterCardsWithRegistry registers all card handlers with a specific registry
func RegisterCardsWithRegistry(registry *cards.CardHandlerRegistry) error {
	handlers := getAllCardHandlers()

	for _, handler := range handlers {
		if err := registry.Register(handler); err != nil {
			return err
		}
	}

	return nil
}

// RegisterCardListeners registers event listeners for all cards that need them
// This automatically detects which cards implement the ListenerRegistrar interface
func RegisterCardListeners(eventBus events.EventBus) error {
	// Since we're using mock cards, no listeners to register
	return nil
}

// UnregisterCardListeners cleans up event listeners for all cards
func UnregisterCardListeners(eventBus events.EventBus) error {
	// Since we're using mock cards, no listeners to unregister
	return nil
}

// getAllCardHandlers returns all card handlers for use in registration functions
func getAllCardHandlers() []cards.CardHandler {
	// Since we're using mock cards, no handlers to register
	return []cards.CardHandler{}
}
