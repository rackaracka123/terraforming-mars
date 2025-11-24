package session

// SessionManager manages game state broadcasting to WebSocket clients for a SPECIFIC game
// Each game has its own SessionManager instance, making it impossible to broadcast to wrong game
// This is an interface defined in the session layer but implemented in the delivery layer
// to maintain clean architecture (session layer should not depend on delivery layer)
type SessionManager interface {
	// Core broadcasting operations - both send complete game state with all data
	// These methods don't need gameID because SessionManager is already bound to a specific game
	Broadcast() error           // Send complete game state to all players in THIS game
	Send(playerID string) error // Send complete game state to specific player in THIS game
	GetGameID() string          // Get the game ID this manager is bound to
}

// SessionManagerFactory creates and manages SessionManager instances per game
// This factory lives in the session layer as an interface
type SessionManagerFactory interface {
	// GetOrCreate returns the SessionManager for a specific game, creating it if needed
	GetOrCreate(gameID string) SessionManager

	// Remove destroys the SessionManager for a game (when game ends)
	Remove(gameID string)
}
