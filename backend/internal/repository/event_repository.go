package repository

import (
	"context"

	"terraforming-mars-backend/internal/events"
)

// EventRepository manages event subscriptions and publications
type EventRepository struct {
	eventBus events.EventBus
}

// NewEventRepository creates a new event repository
func NewEventRepository(eventBus events.EventBus) *EventRepository {
	return &EventRepository{
		eventBus: eventBus,
	}
}

// Subscribe registers a listener for events of the specified type
func (r *EventRepository) Subscribe(eventType string, listener events.EventListener) {
	r.eventBus.Subscribe(eventType, listener)
}

// Publish sends an event to all registered listeners for its type
func (r *EventRepository) Publish(ctx context.Context, event events.Event) error {
	return r.eventBus.Publish(ctx, event)
}

