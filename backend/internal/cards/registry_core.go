package cards

import (
	"fmt"
	"sync"
)

// CardHandlerRegistry manages the registration and lookup of card handlers
type CardHandlerRegistry struct {
	handlers map[string]CardHandler
	mutex    sync.RWMutex
}

// NewCardHandlerRegistry creates a new card handler registry
func NewCardHandlerRegistry() *CardHandlerRegistry {
	return &CardHandlerRegistry{
		handlers: make(map[string]CardHandler),
	}
}

// Register adds a card handler to the registry
func (r *CardHandlerRegistry) Register(handler CardHandler) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	cardID := handler.GetCardID()
	if _, exists := r.handlers[cardID]; exists {
		return fmt.Errorf("handler for card %s already registered", cardID)
	}
	
	r.handlers[cardID] = handler
	return nil
}

// GetHandler retrieves a card handler by card ID
func (r *CardHandlerRegistry) GetHandler(cardID string) (CardHandler, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	handler, exists := r.handlers[cardID]
	if !exists {
		return nil, fmt.Errorf("no handler registered for card %s", cardID)
	}
	
	return handler, nil
}

// HasHandler checks if a handler is registered for the given card ID
func (r *CardHandlerRegistry) HasHandler(cardID string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	_, exists := r.handlers[cardID]
	return exists
}

// GetAllRegisteredCards returns a list of all registered card IDs
func (r *CardHandlerRegistry) GetAllRegisteredCards() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	cardIDs := make([]string, 0, len(r.handlers))
	for cardID := range r.handlers {
		cardIDs = append(cardIDs, cardID)
	}
	
	return cardIDs
}

// UnregisterAll clears all registered handlers (useful for testing)
func (r *CardHandlerRegistry) UnregisterAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.handlers = make(map[string]CardHandler)
}

// Global registry instance
var GlobalCardRegistry = NewCardHandlerRegistry()

// RegisterCardHandler is a convenience function for registering handlers globally
func RegisterCardHandler(handler CardHandler) error {
	return GlobalCardRegistry.Register(handler)
}

// GetCardHandler is a convenience function for getting handlers from the global registry
func GetCardHandler(cardID string) (CardHandler, error) {
	return GlobalCardRegistry.GetHandler(cardID)
}