package events

import "terraforming-mars-backend/internal/model"

// GameCreatedEvent represents when a game is created
type GameCreatedEvent struct {
	BaseEvent
}

// NewGameCreatedEvent creates a new game created event
func NewGameCreatedEvent(gameID string, maxPlayers int) *GameCreatedEvent {
	payload := model.GameCreatedEvent{
		GameID:     gameID,
		MaxPlayers: maxPlayers,
	}

	return &GameCreatedEvent{
		BaseEvent: NewBaseEvent(model.EventTypeGameCreated, gameID, payload),
	}
}

// PlayerJoinedEvent represents when a player joins a game
type PlayerJoinedEvent struct {
	BaseEvent
}

// NewPlayerJoinedEvent creates a new player joined event
func NewPlayerJoinedEvent(gameID, playerID, playerName string) *PlayerJoinedEvent {
	payload := model.PlayerJoinedEvent{
		GameID:     gameID,
		PlayerID:   playerID,
		PlayerName: playerName,
	}

	return &PlayerJoinedEvent{
		BaseEvent: NewBaseEvent(model.EventTypePlayerJoined, gameID, payload),
	}
}

// GameStartedEvent represents when a game starts
type GameStartedEvent struct {
	BaseEvent
}

// NewGameStartedEvent creates a new game started event
func NewGameStartedEvent(gameID string, playerCount int) *GameStartedEvent {
	payload := model.GameStartedEvent{
		GameID:      gameID,
		PlayerCount: playerCount,
	}

	return &GameStartedEvent{
		BaseEvent: NewBaseEvent(model.EventTypeGameStarted, gameID, payload),
	}
}

// PlayerStartingCardOptionsEvent represents when starting cards are dealt to a player
type PlayerStartingCardOptionsEvent struct {
	BaseEvent
}

// NewPlayerStartingCardOptionsEvent creates a new player starting card options event
func NewPlayerStartingCardOptionsEvent(gameID, playerID string, cardOptions []string) *PlayerStartingCardOptionsEvent {
	payload := model.PlayerStartingCardOptionsEvent{
		GameID:      gameID,
		PlayerID:    playerID,
		CardOptions: cardOptions,
	}

	return &PlayerStartingCardOptionsEvent{
		BaseEvent: NewBaseEvent(model.EventTypePlayerStartingCardOptions, gameID, payload),
	}
}

// StartingCardSelectedEvent represents when a player selects starting cards
type StartingCardSelectedEvent struct {
	BaseEvent
}

// NewStartingCardSelectedEvent creates a new starting card selected event
func NewStartingCardSelectedEvent(gameID, playerID string, selectedCards []string, cost int) *StartingCardSelectedEvent {
	payload := model.StartingCardSelectedEvent{
		GameID:        gameID,
		PlayerID:      playerID,
		SelectedCards: selectedCards,
		Cost:          cost,
	}

	return &StartingCardSelectedEvent{
		BaseEvent: NewBaseEvent(model.EventTypeStartingCardSelected, gameID, payload),
	}
}

// GameUpdatedEvent represents when a game's state is updated
type GameUpdatedEvent struct {
	BaseEvent
}

// NewGameUpdatedEvent creates a new game updated event
func NewGameUpdatedEvent(gameID string) *GameUpdatedEvent {
	payload := model.GameUpdatedEvent{
		GameID: gameID,
	}

	return &GameUpdatedEvent{
		BaseEvent: NewBaseEvent(model.EventTypeGameUpdated, gameID, payload),
	}
}

// CardPlayedEvent represents when a player plays a card
type CardPlayedEvent struct {
	BaseEvent
}

// NewCardPlayedEvent creates a new card played event
func NewCardPlayedEvent(gameID, playerID, cardID string) *CardPlayedEvent {
	payload := model.CardPlayedEvent{
		GameID:   gameID,
		PlayerID: playerID,
		CardID:   cardID,
	}

	return &CardPlayedEvent{
		BaseEvent: NewBaseEvent(model.EventTypeCardPlayed, gameID, payload),
	}
}

// PlayerResourcesChangedEvent represents when a player's resources are modified
type PlayerResourcesChangedEvent struct {
	BaseEvent
}

// NewPlayerResourcesChangedEvent creates a new player resources changed event
func NewPlayerResourcesChangedEvent(gameID, playerID string, beforeResources, afterResources model.Resources) *PlayerResourcesChangedEvent {
	payload := model.PlayerResourcesChangedEvent{
		GameID:          gameID,
		PlayerID:        playerID,
		BeforeResources: beforeResources,
		AfterResources:  afterResources,
	}

	return &PlayerResourcesChangedEvent{
		BaseEvent: NewBaseEvent(model.EventTypePlayerResourcesChanged, gameID, payload),
	}
}

// PlayerProductionChangedEvent represents when a player's production is modified
type PlayerProductionChangedEvent struct {
	BaseEvent
}

// NewPlayerProductionChangedEvent creates a new player production changed event
func NewPlayerProductionChangedEvent(gameID, playerID string, beforeProduction, afterProduction model.Production) *PlayerProductionChangedEvent {
	payload := model.PlayerProductionChangedEvent{
		GameID:           gameID,
		PlayerID:         playerID,
		BeforeProduction: beforeProduction,
		AfterProduction:  afterProduction,
	}

	return &PlayerProductionChangedEvent{
		BaseEvent: NewBaseEvent(model.EventTypePlayerProductionChanged, gameID, payload),
	}
}