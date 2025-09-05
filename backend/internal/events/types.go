package events

import "terraforming-mars-backend/internal/model"

// GameCreatedEvent represents when a game is created
type GameCreatedEvent struct {
	BaseEvent
}

// NewGameCreatedEvent creates a new game created event
func NewGameCreatedEvent(gameID string, maxPlayers int) *GameCreatedEvent {
	payload := GameCreatedEventData{
		GameID:     gameID,
		MaxPlayers: maxPlayers,
	}

	return &GameCreatedEvent{
		BaseEvent: NewBaseEvent(EventTypeGameCreated, gameID, payload),
	}
}

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

// PlayerStartingCardOptionsEvent represents when starting cards are dealt to a player
type PlayerStartingCardOptionsEvent struct {
	BaseEvent
}

// NewPlayerStartingCardOptionsEvent creates a new player starting card options event
func NewPlayerStartingCardOptionsEvent(gameID, playerID string, cardOptions []string) *PlayerStartingCardOptionsEvent {
	payload := PlayerStartingCardOptionsEventData{
		GameID:      gameID,
		PlayerID:    playerID,
		CardOptions: cardOptions,
	}

	return &PlayerStartingCardOptionsEvent{
		BaseEvent: NewBaseEvent(EventTypePlayerStartingCardOptions, gameID, payload),
	}
}

// StartingCardSelectedEvent represents when a player selects starting cards
type StartingCardSelectedEvent struct {
	BaseEvent
}

// NewStartingCardSelectedEvent creates a new starting card selected event
func NewStartingCardSelectedEvent(gameID, playerID string, selectedCards []string, cost int) *StartingCardSelectedEvent {
	payload := StartingCardSelectedEventData{
		GameID:        gameID,
		PlayerID:      playerID,
		SelectedCards: selectedCards,
		Cost:          cost,
	}

	return &StartingCardSelectedEvent{
		BaseEvent: NewBaseEvent(EventTypeStartingCardSelected, gameID, payload),
	}
}

// GameUpdatedEvent represents when a game's state is updated
type GameUpdatedEvent struct {
	BaseEvent
}

// NewGameUpdatedEvent creates a new game updated event
func NewGameUpdatedEvent(gameID string) *GameUpdatedEvent {
	payload := GameUpdatedEventData{
		GameID: gameID,
	}

	return &GameUpdatedEvent{
		BaseEvent: NewBaseEvent(EventTypeGameUpdated, gameID, payload),
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

// PlayerResourcesChangedEvent represents when a player's resources are modified
type PlayerResourcesChangedEvent struct {
	BaseEvent
}

// NewPlayerResourcesChangedEvent creates a new player resources changed event
func NewPlayerResourcesChangedEvent(gameID, playerID string, beforeResources, afterResources model.Resources) *PlayerResourcesChangedEvent {
	payload := PlayerResourcesChangedEventData{
		GameID:          gameID,
		PlayerID:        playerID,
		BeforeResources: beforeResources,
		AfterResources:  afterResources,
	}

	return &PlayerResourcesChangedEvent{
		BaseEvent: NewBaseEvent(EventTypePlayerResourcesChanged, gameID, payload),
	}
}

// PlayerProductionChangedEvent represents when a player's production is modified
type PlayerProductionChangedEvent struct {
	BaseEvent
}

// NewPlayerProductionChangedEvent creates a new player production changed event
func NewPlayerProductionChangedEvent(gameID, playerID string, beforeProduction, afterProduction model.Production) *PlayerProductionChangedEvent {
	payload := PlayerProductionChangedEventData{
		GameID:           gameID,
		PlayerID:         playerID,
		BeforeProduction: beforeProduction,
		AfterProduction:  afterProduction,
	}

	return &PlayerProductionChangedEvent{
		BaseEvent: NewBaseEvent(EventTypePlayerProductionChanged, gameID, payload),
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

// GameStateChangedEvent represents when a game's state is changed
type GameStateChangedEvent struct {
	BaseEvent
}

// NewGameStateChangedEvent creates a new game state changed event
func NewGameStateChangedEvent(gameID string, oldState, newState *model.Game) *GameStateChangedEvent {
	payload := GameStateChangedEventData{
		GameID:   gameID,
		OldState: oldState,
		NewState: newState,
	}

	return &GameStateChangedEvent{
		BaseEvent: NewBaseEvent(EventTypeGameStateChanged, gameID, payload),
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

// TemperatureChangedEvent represents when global temperature changes
type TemperatureChangedEvent struct {
	BaseEvent
}

// NewTemperatureChangedEvent creates a new temperature changed event
func NewTemperatureChangedEvent(gameID string, oldTemperature, newTemperature int) *TemperatureChangedEvent {
	payload := TemperatureChangedEventData{
		GameID:         gameID,
		OldTemperature: oldTemperature,
		NewTemperature: newTemperature,
	}

	return &TemperatureChangedEvent{
		BaseEvent: NewBaseEvent(EventTypeTemperatureChanged, gameID, payload),
	}
}

// OxygenChangedEvent represents when global oxygen level changes
type OxygenChangedEvent struct {
	BaseEvent
}

// NewOxygenChangedEvent creates a new oxygen changed event
func NewOxygenChangedEvent(gameID string, oldOxygen, newOxygen int) *OxygenChangedEvent {
	payload := OxygenChangedEventData{
		GameID:    gameID,
		OldOxygen: oldOxygen,
		NewOxygen: newOxygen,
	}

	return &OxygenChangedEvent{
		BaseEvent: NewBaseEvent(EventTypeOxygenChanged, gameID, payload),
	}
}

// OceansChangedEvent represents when ocean count changes
type OceansChangedEvent struct {
	BaseEvent
}

// NewOceansChangedEvent creates a new oceans changed event
func NewOceansChangedEvent(gameID string, oldOceans, newOceans int) *OceansChangedEvent {
	payload := OceansChangedEventData{
		GameID:    gameID,
		OldOceans: oldOceans,
		NewOceans: newOceans,
	}

	return &OceansChangedEvent{
		BaseEvent: NewBaseEvent(EventTypeOceansChanged, gameID, payload),
	}
}

// GlobalParametersChangedEvent represents when any global parameters change
type GlobalParametersChangedEvent struct {
	BaseEvent
}

// NewGlobalParametersChangedEvent creates a new global parameters changed event
func NewGlobalParametersChangedEvent(gameID string, oldParameters, newParameters model.GlobalParameters) *GlobalParametersChangedEvent {
	payload := GlobalParametersChangedEventData{
		GameID:        gameID,
		OldParameters: oldParameters,
		NewParameters: newParameters,
	}

	return &GlobalParametersChangedEvent{
		BaseEvent: NewBaseEvent(EventTypeGlobalParametersChanged, gameID, payload),
	}
}