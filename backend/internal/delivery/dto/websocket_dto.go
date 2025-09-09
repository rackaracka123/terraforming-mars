package dto

// MessageType represents different types of WebSocket messages
type MessageType string

const (
	// Client -> Server messages
	MessageTypePlayerConnect   MessageType = "player-connect"
	MessageTypePlayerReconnect MessageType = "player-reconnect"
	MessageTypePlayAction      MessageType = "do-action"

	// Server -> Client messages
	MessageTypeGameUpdated            MessageType = "game-updated"
	MessageTypePlayerConnected        MessageType = "player-connected"
	MessageTypePlayerReconnected      MessageType = "player-reconnected"
	MessageTypePlayerDisconnected     MessageType = "player-disconnected"
	MessageTypeError                  MessageType = "error"
	MessageTypeFullState              MessageType = "full-state"
	MessageTypeAvailableCards         MessageType = "available-cards"
	MessageTypeProductionPhaseStarted MessageType = "production-phase-started"
)

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    MessageType `json:"type" ts:"MessageType"`
	Payload interface{} `json:"payload" ts:"any"`
	GameID  string      `json:"gameId,omitempty" ts:"string"`
}

// PlayerConnectPayload contains player connection data
type PlayerConnectPayload struct {
	PlayerName string `json:"playerName" ts:"string"`
	GameID     string `json:"gameId" ts:"string"`
}

// PlayActionPayload contains game action data
type PlayActionPayload struct {
	ActionRequest interface{} `json:"actionRequest" ts:"any"`
}

// GameUpdatedPayload contains updated game state
type GameUpdatedPayload struct {
	Game GameDto `json:"game" ts:"GameDto"`
}

// PlayerConnectedPayload contains data about a newly connected player
type PlayerConnectedPayload struct {
	PlayerID   string  `json:"playerId" ts:"string"`
	PlayerName string  `json:"playerName" ts:"string"`
	Game       GameDto `json:"game" ts:"GameDto"`
}

// ErrorPayload contains error information
type ErrorPayload struct {
	Message string `json:"message" ts:"string"`
	Code    string `json:"code,omitempty" ts:"string"`
}

// FullStatePayload contains the complete game state
type FullStatePayload struct {
	Game     GameDto `json:"game" ts:"GameDto"`
	PlayerID string  `json:"playerId" ts:"string"`
}

// AvailableCardsPayload contains available starting cards
type AvailableCardsPayload struct {
	Cards []CardDto `json:"cards" ts:"CardDto[]"`
}

// PlayerReconnectPayload contains player reconnection data
type PlayerReconnectPayload struct {
	PlayerName string `json:"playerName" ts:"string"`
	GameID     string `json:"gameId" ts:"string"`
}

// PlayerReconnectedPayload contains data about a reconnected player
type PlayerReconnectedPayload struct {
	PlayerID   string  `json:"playerId" ts:"string"`
	PlayerName string  `json:"playerName" ts:"string"`
	Game       GameDto `json:"game" ts:"GameDto"`
}

// PlayerDisconnectedPayload contains data about a disconnected player
type PlayerDisconnectedPayload struct {
	PlayerID   string  `json:"playerId" ts:"string"`
	PlayerName string  `json:"playerName" ts:"string"`
	Game       GameDto `json:"game" ts:"GameDto"`
}

// PlayerProductionData contains production data for a single player
type PlayerProductionData struct {
	PlayerID        string        `json:"playerId" ts:"string"`
	PlayerName      string        `json:"playerName" ts:"string"`
	BeforeResources ResourcesDto  `json:"beforeResources" ts:"ResourcesDto"`
	AfterResources  ResourcesDto  `json:"afterResources" ts:"ResourcesDto"`
	Production      ProductionDto `json:"production" ts:"ProductionDto"`
	TerraformRating int           `json:"terraformRating" ts:"number"`
	EnergyConverted int           `json:"energyConverted" ts:"number"`
	CreditsIncome   int           `json:"creditsIncome" ts:"number"`
}

// ProductionPhaseStartedPayload contains data when production phase begins
type ProductionPhaseStartedPayload struct {
	Generation  int                    `json:"generation" ts:"number"`
	PlayersData []PlayerProductionData `json:"playersData" ts:"PlayerProductionData[]"`
	Game        GameDto                `json:"game" ts:"GameDto"`
}
