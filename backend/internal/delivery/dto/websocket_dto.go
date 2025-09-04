package dto

// MessageType represents different types of WebSocket messages
type MessageType string

const (
	// Client -> Server messages
	MessageTypePlayerConnect MessageType = "player-connect"
	MessageTypePlayAction    MessageType = "do-action"

	// Server -> Client messages
	MessageTypeGameUpdated     MessageType = "game-updated"
	MessageTypePlayerConnected MessageType = "player-connected"
	MessageTypeError           MessageType = "error"
	MessageTypeFullState       MessageType = "full-state"
	MessageTypeAvailableCards  MessageType = "available-cards"
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
	PlayerID   string `json:"playerId" ts:"string"`
	PlayerName string `json:"playerName" ts:"string"`
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
