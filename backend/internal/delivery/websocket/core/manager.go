package core

import (
	"sync"
	"terraforming-mars-backend/internal/logger"
	"unsafe"

	"go.uber.org/zap"
)

// Manager handles WebSocket connection lifecycle and organization
type Manager struct {
	// Connection storage
	connections     map[*Connection]bool
	gameConnections map[string]map[*Connection]bool

	// Synchronization
	mu sync.RWMutex
}

// NewManager creates a new connection manager
func NewManager() *Manager {
	return &Manager{
		connections:     make(map[*Connection]bool),
		gameConnections: make(map[string]map[*Connection]bool),
	}
}

// RegisterConnection registers a new connection
func (m *Manager) RegisterConnection(connection *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connections[connection] = true
	logger.Debug("🔗 Client connected to server", zap.String("connection_id", connection.ID))
}

// UnregisterConnection unregisters a connection and handles cleanup
func (m *Manager) UnregisterConnection(connection *Connection) (playerID, gameID string, shouldBroadcast bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.connections[connection]; !exists {
		return "", "", false
	}

	// Remove from main connections
	delete(m.connections, connection)
	connection.CloseSend()

	// Get player info for broadcasting
	playerID, gameID = connection.GetPlayer()
	shouldBroadcast = gameID != "" && playerID != ""

	// Remove from game connections
	if gameConns, exists := m.gameConnections[gameID]; exists {
		delete(gameConns, connection)

		if len(gameConns) == 0 {
			delete(m.gameConnections, gameID)
			logger.Debug("Removed empty game connections map", zap.String("game_id", gameID))
		}
	}

	// Close connection
	connection.Close()

	logger.Debug("⛓️‍💥 Client disconnected from server",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	return playerID, gameID, shouldBroadcast
}

// AddToGame adds a connection to a game group
func (m *Manager) AddToGame(connection *Connection, gameID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.gameConnections[gameID] == nil {
		m.gameConnections[gameID] = make(map[*Connection]bool)
	}
	m.gameConnections[gameID][connection] = true
}

// GetGameConnections returns all connections for a specific game (read-only copy)
func (m *Manager) GetGameConnections(gameID string) map[*Connection]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gameConns := m.gameConnections[gameID]
	if gameConns == nil {
		return nil
	}

	// Return a copy to avoid external mutation
	connections := make(map[*Connection]bool, len(gameConns))
	for conn := range gameConns {
		connections[conn] = true
	}
	return connections
}

// GetConnectionCount returns the total number of registered connections
func (m *Manager) GetConnectionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}

// RemoveExistingPlayerConnection removes any existing connection for the given player
// This is used during reconnection to clean up old connections before adding new ones
// CRITICAL: excludeConnection should be the current connection making the request to avoid cleaning it up
func (m *Manager) RemoveExistingPlayerConnection(playerID, gameID string, excludeConnection *Connection) *Connection {
	m.mu.Lock()
	defer m.mu.Unlock()

	var existingConnection *Connection
	var matchingConnections []*Connection

	logger.Debug("🔍 Starting connection cleanup search",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.String("exclude_connection_id", excludeConnection.ID),
		zap.Uintptr("exclude_connection_ptr", uintptr(unsafe.Pointer(excludeConnection))))

	// Find existing connection for this player, but exclude the current one
	for connection := range m.connections {
		existingPlayerID, existingGameID := connection.GetPlayer()
		if existingPlayerID == playerID && existingGameID == gameID {
			matchingConnections = append(matchingConnections, connection)

			logger.Debug("🔎 Found matching connection",
				zap.String("connection_id", connection.ID),
				zap.Uintptr("connection_ptr", uintptr(unsafe.Pointer(connection))),
				zap.Bool("is_excluded", connection == excludeConnection),
				zap.String("player_id", existingPlayerID),
				zap.String("game_id", existingGameID))

			if connection != excludeConnection {
				existingConnection = connection
				break
			}
		}
	}

	logger.Debug("🔎 Connection search complete",
		zap.Int("total_matching", len(matchingConnections)),
		zap.Bool("found_to_cleanup", existingConnection != nil))

	if existingConnection == nil {
		logger.Debug("🔍 No existing connection to clean up for reconnecting player",
			zap.String("player_id", playerID),
			zap.String("game_id", gameID),
			zap.String("current_connection_id", excludeConnection.ID))
		return nil
	}

	logger.Debug("🧹 Cleaning up existing connection for reconnecting player",
		zap.String("existing_connection_id", existingConnection.ID),
		zap.String("current_connection_id", excludeConnection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.Uintptr("existing_connection_ptr", uintptr(unsafe.Pointer(existingConnection))),
		zap.Uintptr("current_connection_ptr", uintptr(unsafe.Pointer(excludeConnection))))

	// Remove from main connections
	delete(m.connections, existingConnection)
	existingConnection.CloseSend()

	// Remove from game connections
	if gameConns, exists := m.gameConnections[gameID]; exists {
		delete(gameConns, existingConnection)

		if len(gameConns) == 0 {
			delete(m.gameConnections, gameID)
			logger.Debug("Removed empty game connections map after cleanup", zap.String("game_id", gameID))
		}
	}

	// Close connection
	existingConnection.Close()

	logger.Debug("✅ Existing connection cleaned up for reconnecting player",
		zap.String("old_connection_id", existingConnection.ID),
		zap.String("current_connection_id", excludeConnection.ID),
		zap.String("player_id", playerID))

	return existingConnection
}

// CloseAllConnections closes all active connections
func (m *Manager) CloseAllConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Info("🛑 Closing all active connections", zap.Int("connection_count", len(m.connections)))

	for connection := range m.connections {
		connection.Close()
	}

	// Clear the connection maps
	m.connections = make(map[*Connection]bool)
	m.gameConnections = make(map[string]map[*Connection]bool)

	logger.Info("⛓️‍💥 All client connections closed by server")
}

