package model

// GameCreatedEvent represents when a game is created
type GameCreatedEvent struct {
	GameID     string `json:"gameId"`
	MaxPlayers int    `json:"maxPlayers"`
}

// PlayerJoinedEvent represents when a player joins a game
type PlayerJoinedEvent struct {
	GameID     string `json:"gameId"`
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

// GameStartedEvent represents when a game starts
type GameStartedEvent struct {
	GameID      string `json:"gameId"`
	PlayerCount int    `json:"playerCount"`
}

// PlayerStartingCardOptionsEvent represents when starting cards are dealt to a player
type PlayerStartingCardOptionsEvent struct {
	GameID      string   `json:"gameId"`
	PlayerID    string   `json:"playerId"`
	CardOptions []string `json:"cardOptions"`
}

// StartingCardSelectedEvent represents when a player selects starting cards
type StartingCardSelectedEvent struct {
	GameID        string   `json:"gameId"`
	PlayerID      string   `json:"playerId"`
	SelectedCards []string `json:"selectedCards"`
	Cost          int      `json:"cost"`
}

// GameUpdatedEvent represents when a game's state is updated
type GameUpdatedEvent struct {
	GameID string `json:"gameId"`
}

// CardPlayedEvent represents when a player plays a card
type CardPlayedEvent struct {
	GameID   string `json:"gameId"`
	PlayerID string `json:"playerId"`
	CardID   string `json:"cardId"`
}

// PlayerResourcesChangedEvent represents when a player's resources are modified
type PlayerResourcesChangedEvent struct {
	GameID         string    `json:"gameId"`
	PlayerID       string    `json:"playerId"`
	BeforeResources Resources `json:"beforeResources"`
	AfterResources  Resources `json:"afterResources"`
}

// PlayerProductionChangedEvent represents when a player's production is modified
type PlayerProductionChangedEvent struct {
	GameID           string     `json:"gameId"`
	PlayerID         string     `json:"playerId"`
	BeforeProduction Production `json:"beforeProduction"`
	AfterProduction  Production `json:"afterProduction"`
}
