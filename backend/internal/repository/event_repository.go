package repository

import (
	"fmt"
	"sort"
	"sync"
	"terraforming-mars-backend/internal/domain"
)

// EventRepository handles event storage and retrieval
type EventRepository struct {
	events map[string][]domain.GameEvent // gameID -> events
	mutex  sync.RWMutex
}

// NewEventRepository creates a new event repository
func NewEventRepository() *EventRepository {
	return &EventRepository{
		events: make(map[string][]domain.GameEvent),
	}
}

// SaveEvent stores a single event
func (r *EventRepository) SaveEvent(event domain.GameEvent) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.events[event.GameID] == nil {
		r.events[event.GameID] = []domain.GameEvent{}
	}

	// Ensure events are stored in version order
	events := r.events[event.GameID]
	
	// Check for duplicate versions
	for _, existingEvent := range events {
		if existingEvent.Version == event.Version {
			return fmt.Errorf("event with version %d already exists for game %s", event.Version, event.GameID)
		}
	}

	// Insert event in correct position based on version
	insertIndex := len(events)
	for i, existingEvent := range events {
		if existingEvent.Version > event.Version {
			insertIndex = i
			break
		}
	}

	// Insert at the correct position
	events = append(events[:insertIndex], append([]domain.GameEvent{event}, events[insertIndex:]...)...)
	r.events[event.GameID] = events

	return nil
}

// SaveEvents stores multiple events atomically
func (r *EventRepository) SaveEvents(gameID string, events []domain.GameEvent) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.events[gameID] == nil {
		r.events[gameID] = []domain.GameEvent{}
	}

	existingEvents := r.events[gameID]
	
	// Validate that new events have proper version sequence
	for i, event := range events {
		if event.GameID != gameID {
			return fmt.Errorf("event %d has wrong game ID: expected %s, got %s", i, gameID, event.GameID)
		}

		// Check for version conflicts with existing events
		for _, existingEvent := range existingEvents {
			if existingEvent.Version == event.Version {
				return fmt.Errorf("event with version %d already exists for game %s", event.Version, gameID)
			}
		}
	}

	// Add all events
	allEvents := append(existingEvents, events...)
	
	// Sort by version to maintain order
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Version < allEvents[j].Version
	})

	r.events[gameID] = allEvents
	return nil
}

// GetEvents retrieves all events for a game
func (r *EventRepository) GetEvents(gameID string) ([]domain.GameEvent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	events, exists := r.events[gameID]
	if !exists {
		return nil, fmt.Errorf("no events found for game %s", gameID)
	}

	// Return a copy to prevent external modifications
	eventsCopy := make([]domain.GameEvent, len(events))
	copy(eventsCopy, events)
	
	return eventsCopy, nil
}

// GetEventsFromVersion retrieves events starting from a specific version
func (r *EventRepository) GetEventsFromVersion(gameID string, fromVersion int64) ([]domain.GameEvent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	allEvents, exists := r.events[gameID]
	if !exists {
		return nil, fmt.Errorf("no events found for game %s", gameID)
	}

	var filteredEvents []domain.GameEvent
	for _, event := range allEvents {
		if event.Version >= fromVersion {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

// GetEventsToVersion retrieves events up to a specific version
func (r *EventRepository) GetEventsToVersion(gameID string, toVersion int64) ([]domain.GameEvent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	allEvents, exists := r.events[gameID]
	if !exists {
		return nil, fmt.Errorf("no events found for game %s", gameID)
	}

	var filteredEvents []domain.GameEvent
	for _, event := range allEvents {
		if event.Version <= toVersion {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

// GetEventsByType retrieves all events of a specific type for a game
func (r *EventRepository) GetEventsByType(gameID string, eventType domain.GameEventType) ([]domain.GameEvent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	allEvents, exists := r.events[gameID]
	if !exists {
		return nil, fmt.Errorf("no events found for game %s", gameID)
	}

	var filteredEvents []domain.GameEvent
	for _, event := range allEvents {
		if event.Type == eventType {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

// GetEventsByPlayer retrieves all events performed by a specific player
func (r *EventRepository) GetEventsByPlayer(gameID string, playerID string) ([]domain.GameEvent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	allEvents, exists := r.events[gameID]
	if !exists {
		return nil, fmt.Errorf("no events found for game %s", gameID)
	}

	var filteredEvents []domain.GameEvent
	for _, event := range allEvents {
		if event.PlayerID != nil && *event.PlayerID == playerID {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

// GetLatestVersion returns the latest version number for a game
func (r *EventRepository) GetLatestVersion(gameID string) (int64, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	events, exists := r.events[gameID]
	if !exists {
		return 0, fmt.Errorf("no events found for game %s", gameID)
	}

	if len(events) == 0 {
		return 0, nil
	}

	return events[len(events)-1].Version, nil
}

// GameExists checks if a game exists (has any events)
func (r *EventRepository) GameExists(gameID string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	events, exists := r.events[gameID]
	return exists && len(events) > 0
}

// GetGameEventStream returns a complete event stream for a game
func (r *EventRepository) GetGameEventStream(gameID string) (*domain.GameEventStream, error) {
	events, err := r.GetEvents(gameID)
	if err != nil {
		return nil, err
	}

	version, err := r.GetLatestVersion(gameID)
	if err != nil {
		return nil, err
	}

	var createdAt, updatedAt *domain.GameEvent
	if len(events) > 0 {
		createdAt = &events[0]
		updatedAt = &events[len(events)-1]
	}

	return &domain.GameEventStream{
		GameID:    gameID,
		Events:    events,
		Version:   version,
		CreatedAt: createdAt.Timestamp,
		UpdatedAt: updatedAt.Timestamp,
	}, nil
}

// DeleteGame removes all events for a game (for testing/cleanup)
func (r *EventRepository) DeleteGame(gameID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.events, gameID)
	return nil
}

// GetAllGameIDs returns all game IDs that have events
func (r *EventRepository) GetAllGameIDs() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var gameIDs []string
	for gameID := range r.events {
		gameIDs = append(gameIDs, gameID)
	}

	sort.Strings(gameIDs)
	return gameIDs
}