package events

import (
	"time"

	"terraforming-mars-backend/internal/model"
)

// Event type constants - Simplified event system
const (
	// Core Game Events - Game lifecycle management
	EventTypeGameCreated = "game.created" // When a game is created
	EventTypeGameStarted = "game.started" // When a game transitions from lobby to active
	EventTypeGameUpdated = "game.updated" // When any game state changes (consolidated from GameStateChanged)

	// Player Events - Player lifecycle and state changes
	EventTypePlayerJoined  = "player.joined"  // When a player joins a game
	EventTypePlayerChanged = "player.changed" // When player resources, production, or TR changes (consolidated)

	// Card Events - Card-related game actions
	EventTypeCardSelected = "card.selected" // Player selects starting cards (renamed from StartingCardSelected)
	EventTypeCardPlayed   = "card.played"   // Player plays a card during game

	// Global Events - Terraforming parameters
	EventTypeGlobalParametersChanged = "global.parameters_changed" // Any global parameter changes (consolidated)
)

// Event represents a domain event that can be published and consumed
type Event interface {
	// GetType returns the type of the event
	GetType() string
	// GetGameID returns the game ID this event is associated with
	GetGameID() string
	// GetTimestamp returns when the event occurred
	GetTimestamp() time.Time
	// GetPayload returns the event-specific data
	GetPayload() any
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	Type      string    `json:"type"`
	GameID    string    `json:"gameId"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
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
func (e *BaseEvent) GetPayload() any {
	return e.Payload
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType, gameID string, payload any) BaseEvent {
	return BaseEvent{
		Type:      eventType,
		GameID:    gameID,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}

// Event data types - Simplified payloads for consolidated events

// Core Game Event Data

// GameCreatedEventData represents when a game is created
type GameCreatedEventData struct {
	GameID       string             `json:"gameId"`
	GameSettings model.GameSettings `json:"gameSettings"`
}

// GameStartedEventData represents when a game starts
type GameStartedEventData struct {
	GameID      string `json:"gameId"`
	PlayerCount int    `json:"playerCount"`
}

// GameUpdatedEventData represents when any game state changes (simplified from GameStateChanged)
type GameUpdatedEventData struct {
	GameID string `json:"gameId"`
}

// Player Event Data

// PlayerJoinedEventData represents when a player joins a game
type PlayerJoinedEventData struct {
	GameID     string `json:"gameId"`
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

// PlayerChangedEventData represents when player resources, production, or TR changes (consolidated)
type PlayerChangedEventData struct {
	GameID     string `json:"gameId"`
	PlayerID   string `json:"playerId"`
	ChangeType string `json:"changeType"` // "resources", "production", "terraform_rating"
}

// Card Event Data

// CardSelectedEventData represents when a player selects starting cards (renamed from StartingCardSelected)
type CardSelectedEventData struct {
	GameID        string   `json:"gameId"`
	PlayerID      string   `json:"playerId"`
	SelectedCards []string `json:"selectedCards"`
	Cost          int      `json:"cost"`
}

// CardPlayedEventData represents when a player plays a card
type CardPlayedEventData struct {
	GameID   string `json:"gameId"`
	PlayerID string `json:"playerId"`
	CardID   string `json:"cardId"`
}

// Global Event Data

// GlobalParametersChangedEventData represents when any global parameters change (consolidated from individual parameter events)
type GlobalParametersChangedEventData struct {
	GameID      string   `json:"gameId"`
	ChangeTypes []string `json:"changeTypes"` // ["temperature", "oxygen", "oceans"] - which parameters changed
}
