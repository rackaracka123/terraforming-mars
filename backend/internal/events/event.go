package events

import (
	"terraforming-mars-backend/internal/model"
	"time"
)

// Event type constants
const (
	EventTypeGameCreated               = "game.created"
	EventTypeGameDeleted               = "game.deleted"
	EventTypeGameStateChanged          = "game.state_changed"
	EventTypePlayerJoined              = "player.joined"
	EventTypePlayerLeft                = "player.left"
	EventTypeGameStarted               = "game.started"
	EventTypePlayerStartingCardOptions = "player.starting_card_options"
	EventTypeStartingCardSelected      = "starting_card.selected"
	EventTypePhaseChanged              = "game.phase_changed"
	EventTypeGameUpdated               = "game.updated"
	EventTypeCardPlayed                = "card.played"

	// Player Changed
	EventTypePlayerResourcesChanged  = "player.resources_changed"
	EventTypePlayerProductionChanged = "player.production_changed"

	// Global Parameter Changed
	EventTypeTemperatureChanged      = "global-parameters.temperature_changed"
	EventTypeOxygenChanged           = "global-parameters.oxygen_changed"
	EventTypeOceansChanged           = "global-parameters.oceans_changed"
	EventTypeGlobalParametersChanged = "global-parameters.parameters_changed"
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

// Event data types - moved from model package

// GameCreatedEventData represents when a game is created
type GameCreatedEventData struct {
	GameID     string `json:"gameId"`
	MaxPlayers int    `json:"maxPlayers"`
}

// PlayerJoinedEventData represents when a player joins a game
type PlayerJoinedEventData struct {
	GameID     string `json:"gameId"`
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

// GameStartedEventData represents when a game starts
type GameStartedEventData struct {
	GameID      string `json:"gameId"`
	PlayerCount int    `json:"playerCount"`
}

// PlayerStartingCardOptionsEventData represents when starting cards are dealt to a player
type PlayerStartingCardOptionsEventData struct {
	GameID      string   `json:"gameId"`
	PlayerID    string   `json:"playerId"`
	CardOptions []string `json:"cardOptions"`
}

// StartingCardSelectedEventData represents when a player selects starting cards
type StartingCardSelectedEventData struct {
	GameID        string   `json:"gameId"`
	PlayerID      string   `json:"playerId"`
	SelectedCards []string `json:"selectedCards"`
	Cost          int      `json:"cost"`
}

// GameUpdatedEventData represents when a game's state is updated
type GameUpdatedEventData struct {
	GameID string `json:"gameId"`
}

// CardPlayedEventData represents when a player plays a card
type CardPlayedEventData struct {
	GameID   string `json:"gameId"`
	PlayerID string `json:"playerId"`
	CardID   string `json:"cardId"`
}

// PlayerResourcesChangedEventData represents when a player's resources are modified
type PlayerResourcesChangedEventData struct {
	GameID          string          `json:"gameId"`
	PlayerID        string          `json:"playerId"`
	BeforeResources model.Resources `json:"beforeResources"`
	AfterResources  model.Resources `json:"afterResources"`
}

// PlayerProductionChangedEventData represents when a player's production is modified
type PlayerProductionChangedEventData struct {
	GameID           string           `json:"gameId"`
	PlayerID         string           `json:"playerId"`
	BeforeProduction model.Production `json:"beforeProduction"`
	AfterProduction  model.Production `json:"afterProduction"`
}

// GameDeletedEventData represents when a game is deleted
type GameDeletedEventData struct {
	GameID string `json:"gameId"`
}

// GameStateChangedEventData represents when a game's state is changed
type GameStateChangedEventData struct {
	GameID   string      `json:"gameId"`
	OldState *model.Game `json:"oldState"`
	NewState *model.Game `json:"newState"`
}

// PlayerLeftEventData represents when a player leaves a game
type PlayerLeftEventData struct {
	GameID     string `json:"gameId"`
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

// TemperatureChangedEventData represents when global temperature changes
type TemperatureChangedEventData struct {
	GameID         string `json:"gameId"`
	OldTemperature int    `json:"oldTemperature"`
	NewTemperature int    `json:"newTemperature"`
}

// OxygenChangedEventData represents when global oxygen level changes
type OxygenChangedEventData struct {
	GameID    string `json:"gameId"`
	OldOxygen int    `json:"oldOxygen"`
	NewOxygen int    `json:"newOxygen"`
}

// OceansChangedEventData represents when ocean count changes
type OceansChangedEventData struct {
	GameID    string `json:"gameId"`
	OldOceans int    `json:"oldOceans"`
	NewOceans int    `json:"newOceans"`
}

// GlobalParametersChangedEventData represents when any global parameters change
type GlobalParametersChangedEventData struct {
	GameID        string                 `json:"gameId"`
	OldParameters model.GlobalParameters `json:"oldParameters"`
	NewParameters model.GlobalParameters `json:"newParameters"`
}
