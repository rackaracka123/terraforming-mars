package game

// GameStatus represents the current status of the game
type GameStatus string

const (
	GameStatusLobby     GameStatus = "lobby"
	GameStatusActive    GameStatus = "active"
	GameStatusCompleted GameStatus = "completed"
)
