package events

import (
	"context"
)

// EventRepository manages event subscriptions and publications
type EventRepository struct {
	eventBus EventBus
}

// NewEventRepository creates a new event repository
func NewEventRepository(eventBus EventBus) *EventRepository {
	return &EventRepository{
		eventBus: eventBus,
	}
}

// Subscribe registers a listener for events of the specified type
func (r *EventRepository) Subscribe(eventType string, listener EventListener) {
	r.eventBus.Subscribe(eventType, listener)
}

// Publish sends an event to all registered listeners for its type
func (r *EventRepository) Publish(ctx context.Context, event Event) error {
	return r.eventBus.Publish(ctx, event)
}