package dto

// CreateGameRequest represents the request body for creating a game
type CreateGameRequest struct {
	MaxPlayers int `json:"maxPlayers" binding:"required,min=1,max=5"`
}

// CreateGameResponse represents the response for creating a game
type CreateGameResponse struct {
	Game GameDto `json:"game" ts:"GameDto"`
}

// JoinGameRequest represents the request body for joining a game
type JoinGameRequest struct {
	PlayerName string `json:"playerName" binding:"required,min=1,max=50"`
}

// JoinGameResponse represents the response for joining a game
type JoinGameResponse struct {
	Game     GameDto `json:"game" ts:"GameDto"`
	PlayerID string  `json:"playerId" ts:"string"`
}

// GetGameResponse represents the response for getting a game
type GetGameResponse struct {
	Game GameDto `json:"game" ts:"GameDto"`
}

// ListGamesResponse represents the response for listing games
type ListGamesResponse struct {
	Games []GameDto `json:"games" ts:"GameDto[]"`
}

// UpdateResourcesRequest represents the request body for updating player resources
type UpdateResourcesRequest struct {
	Resources ResourcesDto `json:"resources" ts:"ResourcesDto"`
}

// GetPlayerResponse represents the response for getting a player
type GetPlayerResponse struct {
	Player PlayerDto `json:"player" ts:"PlayerDto"`
}

// UpdatePlayerResourcesResponse represents the response for updating player resources
type UpdatePlayerResourcesResponse struct {
	Player PlayerDto `json:"player" ts:"PlayerDto"`
}

// ListCardsResponse represents the response for listing cards with pagination
type ListCardsResponse struct {
	Cards      []CardDto `json:"cards" ts:"CardDto[]"`
	TotalCount int       `json:"totalCount" ts:"number"`
	Offset     int       `json:"offset" ts:"number"`
	Limit      int       `json:"limit" ts:"number"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" ts:"string"`
	Code    string `json:"code,omitempty" ts:"string"`
	Details string `json:"details,omitempty" ts:"string"`
}
