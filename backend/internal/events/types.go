package events

// Event type constants
const (
	EventTypeGameStarted                = "game.started"
	EventTypePlayerStartingCardOptions  = "player.starting_card_options"
	EventTypeCardPlayed                 = "card.played"
)

// GameStartedEvent represents when a game transitions from lobby to active
type GameStartedEvent struct {
	BaseEvent
}

// GameStartedPayload contains the data for a game started event
type GameStartedPayload struct {
	GameID    string   `json:"gameId"`
	PlayerIDs []string `json:"playerIds"`
}

// NewGameStartedEvent creates a new game started event
func NewGameStartedEvent(gameID string, playerIDs []string) *GameStartedEvent {
	payload := GameStartedPayload{
		GameID:    gameID,
		PlayerIDs: playerIDs,
	}

	return &GameStartedEvent{
		BaseEvent: NewBaseEvent(EventTypeGameStarted, gameID, payload),
	}
}

// PlayerStartingCardOptionsEvent represents when starting card options are dealt to a specific player
type PlayerStartingCardOptionsEvent struct {
	BaseEvent
}

// PlayerStartingCardOptionsPayload contains the card options for a specific player
type PlayerStartingCardOptionsPayload struct {
	GameID       string   `json:"gameId"`
	PlayerID     string   `json:"playerId"`
	CardOptions  []string `json:"cardOptions"`
}

// NewPlayerStartingCardOptionsEvent creates a new player starting card options event for a specific player
func NewPlayerStartingCardOptionsEvent(gameID, playerID string, cardOptions []string) *PlayerStartingCardOptionsEvent {
	payload := PlayerStartingCardOptionsPayload{
		GameID:      gameID,
		PlayerID:    playerID,
		CardOptions: cardOptions,
	}

	return &PlayerStartingCardOptionsEvent{
		BaseEvent: NewBaseEvent(EventTypePlayerStartingCardOptions, gameID, payload),
	}
}

// CardPlayedEvent represents when a player plays a card
type CardPlayedEvent struct {
	BaseEvent
}

// CardPlayedPayload contains the data for a card played event
type CardPlayedPayload struct {
	GameID   string `json:"gameId"`
	PlayerID string `json:"playerId"`
	CardID   string `json:"cardId"`
}

// NewCardPlayedEvent creates a new card played event
func NewCardPlayedEvent(gameID, playerID, cardID string) *CardPlayedEvent {
	payload := CardPlayedPayload{
		GameID:   gameID,
		PlayerID: playerID,
		CardID:   cardID,
	}

	return &CardPlayedEvent{
		BaseEvent: NewBaseEvent(EventTypeCardPlayed, gameID, payload),
	}
}