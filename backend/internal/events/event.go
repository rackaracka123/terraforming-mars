package events

import "time"

// Event represents a domain event that can be published and consumed
type Event interface {
	// GetType returns the type of the event
	GetType() string
	// GetGameID returns the game ID this event is associated with
	GetGameID() string
	// GetTimestamp returns when the event occurred
	GetTimestamp() time.Time
	// GetPayload returns the event-specific data
	GetPayload() interface{}
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	Type      string      `json:"type"`
	GameID    string      `json:"gameId"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// GetType returns the event type
func (e *BaseEvent) GetType() string {
	return e.Type
}

// GetGameID returns the game ID
func (e *BaseEvent) GetGameID() string {
	return e.GameID
}

// GetTimestamp returns the event timestamp
func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetPayload returns the event payload
func (e *BaseEvent) GetPayload() interface{} {
	return e.Payload
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType, gameID string, payload interface{}) BaseEvent {
	return BaseEvent{
		Type:      eventType,
		GameID:    gameID,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}