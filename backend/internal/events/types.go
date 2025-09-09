package events

import "terraforming-mars-backend/internal/model"

// Core Game Events

// GameCreatedEvent represents when a game is created
type GameCreatedEvent struct {
	BaseEvent
}

// NewGameCreatedEvent creates a new game created event
func NewGameCreatedEvent(gameID string, settings model.GameSettings) *GameCreatedEvent {
	payload := GameCreatedEventData{
		GameID:       gameID,
		GameSettings: settings,
	}

	return &GameCreatedEvent{
		BaseEvent: NewBaseEvent(EventTypeGameCreated, gameID, payload),
	}
}

// GameStartedEvent represents when a game starts
type GameStartedEvent struct {
	BaseEvent
}

// NewGameStartedEvent creates a new game started event
func NewGameStartedEvent(gameID string, playerCount int) *GameStartedEvent {
	payload := GameStartedEventData{
		GameID:      gameID,
		PlayerCount: playerCount,
	}

	return &GameStartedEvent{
		BaseEvent: NewBaseEvent(EventTypeGameStarted, gameID, payload),
	}
}

// GameUpdatedEvent represents when any game state changes (consolidated from GameStateChangedEvent)
type GameUpdatedEvent struct {
	BaseEvent
}

// NewGameUpdatedEvent creates a new game updated event (simplified payload)
func NewGameUpdatedEvent(gameID string) *GameUpdatedEvent {
	payload := GameUpdatedEventData{
		GameID: gameID,
	}

	return &GameUpdatedEvent{
		BaseEvent: NewBaseEvent(EventTypeGameUpdated, gameID, payload),
	}
}

// GameDeletedEvent represents when a game is deleted
type GameDeletedEvent struct {
	BaseEvent
}

// NewGameDeletedEvent creates a new game deleted event
func NewGameDeletedEvent(gameID string) *GameDeletedEvent {
	payload := GameDeletedEventData{
		GameID: gameID,
	}

	return &GameDeletedEvent{
		BaseEvent: NewBaseEvent(EventTypeGameDeleted, gameID, payload),
	}
}

// Player Events

// PlayerJoinedEvent represents when a player joins a game
type PlayerJoinedEvent struct {
	BaseEvent
}

// NewPlayerJoinedEvent creates a new player joined event
func NewPlayerJoinedEvent(gameID, playerID, playerName string) *PlayerJoinedEvent {
	payload := PlayerJoinedEventData{
		GameID:     gameID,
		PlayerID:   playerID,
		PlayerName: playerName,
	}

	return &PlayerJoinedEvent{
		BaseEvent: NewBaseEvent(EventTypePlayerJoined, gameID, payload),
	}
}

// PlayerLeftEvent represents when a player leaves a game
type PlayerLeftEvent struct {
	BaseEvent
}

// NewPlayerLeftEvent creates a new player left event
func NewPlayerLeftEvent(gameID, playerID, playerName string) *PlayerLeftEvent {
	payload := PlayerLeftEventData{
		GameID:     gameID,
		PlayerID:   playerID,
		PlayerName: playerName,
	}

	return &PlayerLeftEvent{
		BaseEvent: NewBaseEvent(EventTypePlayerLeft, gameID, payload),
	}
}

// PlayerChangedEvent represents when player resources, production, or TR changes (consolidated event)
type PlayerChangedEvent struct {
	BaseEvent
}

// NewPlayerChangedEvent creates a new consolidated player changed event
func NewPlayerChangedEvent(gameID, playerID, changeType string) *PlayerChangedEvent {
	payload := PlayerChangedEventData{
		GameID:     gameID,
		PlayerID:   playerID,
		ChangeType: changeType,
	}

	return &PlayerChangedEvent{
		BaseEvent: NewBaseEvent(EventTypePlayerChanged, gameID, payload),
	}
}

// Convenience constructors for specific player change types
func NewPlayerResourcesChangedEvent(gameID, playerID string) *PlayerChangedEvent {
	return NewPlayerChangedEvent(gameID, playerID, "resources")
}

func NewPlayerProductionChangedEvent(gameID, playerID string) *PlayerChangedEvent {
	return NewPlayerChangedEvent(gameID, playerID, "production")
}

func NewPlayerTRChangedEvent(gameID, playerID string) *PlayerChangedEvent {
	return NewPlayerChangedEvent(gameID, playerID, "terraform_rating")
}

// Card Events

// CardDealtEvent represents when starting cards are dealt to a player (renamed from PlayerStartingCardOptionsEvent)
type CardDealtEvent struct {
	BaseEvent
}

// NewCardDealtEvent creates a new card dealt event
func NewCardDealtEvent(gameID, playerID string, cardOptions []string) *CardDealtEvent {
	payload := CardDealtEventData{
		GameID:      gameID,
		PlayerID:    playerID,
		CardOptions: cardOptions,
	}

	return &CardDealtEvent{
		BaseEvent: NewBaseEvent(EventTypeCardDealt, gameID, payload),
	}
}

// CardSelectedEvent represents when a player selects starting cards (renamed from StartingCardSelectedEvent)
type CardSelectedEvent struct {
	BaseEvent
}

// NewCardSelectedEvent creates a new card selected event
func NewCardSelectedEvent(gameID, playerID string, selectedCards []string, cost int) *CardSelectedEvent {
	payload := CardSelectedEventData{
		GameID:        gameID,
		PlayerID:      playerID,
		SelectedCards: selectedCards,
		Cost:          cost,
	}

	return &CardSelectedEvent{
		BaseEvent: NewBaseEvent(EventTypeCardSelected, gameID, payload),
	}
}

// CardPlayedEvent represents when a player plays a card
type CardPlayedEvent struct {
	BaseEvent
}

// NewCardPlayedEvent creates a new card played event
func NewCardPlayedEvent(gameID, playerID, cardID string) *CardPlayedEvent {
	payload := CardPlayedEventData{
		GameID:   gameID,
		PlayerID: playerID,
		CardID:   cardID,
	}

	return &CardPlayedEvent{
		BaseEvent: NewBaseEvent(EventTypeCardPlayed, gameID, payload),
	}
}

// Global Events

// GlobalParametersChangedEvent represents when any global parameters change (consolidated from individual parameter events)
type GlobalParametersChangedEvent struct {
	BaseEvent
}

// NewGlobalParametersChangedEvent creates a new global parameters changed event
func NewGlobalParametersChangedEvent(gameID string, changeTypes []string) *GlobalParametersChangedEvent {
	payload := GlobalParametersChangedEventData{
		GameID:      gameID,
		ChangeTypes: changeTypes,
	}

	return &GlobalParametersChangedEvent{
		BaseEvent: NewBaseEvent(EventTypeGlobalParametersChanged, gameID, payload),
	}
}
