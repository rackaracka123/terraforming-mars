package repository

import (
	"fmt"
	"terraforming-mars-backend/internal/domain"
	"time"
)

// EventSourcedGameRepository implements game repository using event sourcing
type EventSourcedGameRepository struct {
	eventRepo *EventRepository
}

// NewEventSourcedGameRepository creates a new event-sourced game repository
func NewEventSourcedGameRepository(eventRepo *EventRepository) *EventSourcedGameRepository {
	return &EventSourcedGameRepository{
		eventRepo: eventRepo,
	}
}

// GetGame reconstructs the game state from events
func (r *EventSourcedGameRepository) GetGame(gameID string) (*domain.GameState, error) {
	events, err := r.eventRepo.GetEvents(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for game %s: %w", gameID, err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("game %s not found", gameID)
	}

	aggregate, err := domain.NewGameAggregate(gameID, events)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct game state: %w", err)
	}

	return aggregate.GetState(), nil
}

// GetGameAtVersion reconstructs the game state at a specific version
func (r *EventSourcedGameRepository) GetGameAtVersion(gameID string, version int64) (*domain.GameState, error) {
	events, err := r.eventRepo.GetEventsToVersion(gameID, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for game %s up to version %d: %w", gameID, version, err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("game %s not found or no events up to version %d", gameID, version)
	}

	aggregate, err := domain.NewGameAggregate(gameID, events)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct game state at version %d: %w", version, err)
	}

	return aggregate.GetState(), nil
}

// SaveGame is not used in event sourcing - use SaveEvent instead
func (r *EventSourcedGameRepository) SaveGame(game *domain.GameState) error {
	return fmt.Errorf("SaveGame not supported in event-sourced repository - use SaveEvent instead")
}

// SaveEvent saves a single event and returns the updated game state
func (r *EventSourcedGameRepository) SaveEvent(event domain.GameEvent) (*domain.GameState, error) {
	// Get current events
	existingEvents, err := r.eventRepo.GetEvents(event.GameID)
	if err != nil && err.Error() != fmt.Sprintf("no events found for game %s", event.GameID) {
		return nil, fmt.Errorf("failed to get existing events: %w", err)
	}

	// Set the version for the new event
	if len(existingEvents) > 0 {
		lastVersion := existingEvents[len(existingEvents)-1].Version
		event.Version = lastVersion + 1
	} else {
		event.Version = 1
	}

	// Save the event
	if err := r.eventRepo.SaveEvent(event); err != nil {
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	// Reconstruct and return the updated game state
	return r.GetGame(event.GameID)
}

// SaveEvents saves multiple events atomically
func (r *EventSourcedGameRepository) SaveEvents(gameID string, events []domain.GameEvent) (*domain.GameState, error) {
	if len(events) == 0 {
		return r.GetGame(gameID)
	}

	// Get current version
	currentVersion, err := r.eventRepo.GetLatestVersion(gameID)
	if err != nil && err.Error() != fmt.Sprintf("no events found for game %s", gameID) {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	// Set versions for new events
	for i := range events {
		events[i].Version = currentVersion + int64(i) + 1
		events[i].GameID = gameID
		if events[i].Timestamp.IsZero() {
			events[i].Timestamp = time.Now()
		}
	}

	// Save all events
	if err := r.eventRepo.SaveEvents(gameID, events); err != nil {
		return nil, fmt.Errorf("failed to save events: %w", err)
	}

	// Reconstruct and return the updated game state
	return r.GetGame(gameID)
}

// CreateGame creates a new game by saving a GameCreated event
func (r *EventSourcedGameRepository) CreateGame(gameID string, settings domain.GameSettings, createdBy string) (*domain.GameState, error) {
	factory := domain.NewEventFactory(gameID)
	
	event, err := factory.CreateEvent(
		domain.EventTypeGameCreated,
		nil, // No specific player for game creation
		domain.GameCreatedData{
			Settings:   settings,
			CreatedBy:  createdBy,
			MaxPlayers: 4, // Default max players
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create game created event: %w", err)
	}

	return r.SaveEvent(*event)
}

// GetGameHistory returns the complete event history for a game
func (r *EventSourcedGameRepository) GetGameHistory(gameID string) ([]domain.GameEvent, error) {
	return r.eventRepo.GetEvents(gameID)
}

// GetGameHistoryFromVersion returns events from a specific version onwards
func (r *EventSourcedGameRepository) GetGameHistoryFromVersion(gameID string, fromVersion int64) ([]domain.GameEvent, error) {
	return r.eventRepo.GetEventsFromVersion(gameID, fromVersion)
}

// GetPlayerActions returns all actions taken by a specific player
func (r *EventSourcedGameRepository) GetPlayerActions(gameID string, playerID string) ([]domain.GameEvent, error) {
	return r.eventRepo.GetEventsByPlayer(gameID, playerID)
}

// GetEventsByType returns all events of a specific type for a game
func (r *EventSourcedGameRepository) GetEventsByType(gameID string, eventType domain.GameEventType) ([]domain.GameEvent, error) {
	return r.eventRepo.GetEventsByType(gameID, eventType)
}

// GameExists checks if a game exists
func (r *EventSourcedGameRepository) GameExists(gameID string) bool {
	return r.eventRepo.GameExists(gameID)
}

// GetAllGameIDs returns all game IDs
func (r *EventSourcedGameRepository) GetAllGameIDs() []string {
	return r.eventRepo.GetAllGameIDs()
}

// DeleteGame deletes all events for a game
func (r *EventSourcedGameRepository) DeleteGame(gameID string) error {
	return r.eventRepo.DeleteGame(gameID)
}

// GetGameEventStream returns the complete event stream
func (r *EventSourcedGameRepository) GetGameEventStream(gameID string) (*domain.GameEventStream, error) {
	return r.eventRepo.GetGameEventStream(gameID)
}

// ReplayGameToVersion reconstructs the game state up to a specific version
// This is useful for debugging and time-travel functionality
func (r *EventSourcedGameRepository) ReplayGameToVersion(gameID string, version int64) (*domain.GameState, error) {
	events, err := r.eventRepo.GetEventsToVersion(gameID, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get events up to version %d: %w", version, err)
	}

	aggregate, err := domain.NewGameAggregate(gameID, events)
	if err != nil {
		return nil, fmt.Errorf("failed to replay game to version %d: %w", version, err)
	}

	return aggregate.GetState(), nil
}

// GetGameStateDiff compares game state between two versions
func (r *EventSourcedGameRepository) GetGameStateDiff(gameID string, fromVersion, toVersion int64) ([]domain.GameEvent, error) {
	if fromVersion > toVersion {
		return nil, fmt.Errorf("fromVersion (%d) cannot be greater than toVersion (%d)", fromVersion, toVersion)
	}

	allEvents, err := r.eventRepo.GetEvents(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	var diffEvents []domain.GameEvent
	for _, event := range allEvents {
		if event.Version > fromVersion && event.Version <= toVersion {
			diffEvents = append(diffEvents, event)
		}
	}

	return diffEvents, nil
}