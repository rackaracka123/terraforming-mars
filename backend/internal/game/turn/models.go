package turn

// GameStatus represents the status of a game
type GameStatus string

const (
	GameStatusActive GameStatus = "active"
)

// Game represents the minimal game state needed for turn management
type Game struct {
	CurrentTurn *string
	Status      GameStatus
	PlayerIDs   []string
}

// Player represents the minimal player state needed for turn management
type Player struct {
	ID               string
	Passed           bool
	AvailableActions int
}
