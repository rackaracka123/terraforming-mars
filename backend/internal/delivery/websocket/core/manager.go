package core

import (
	"sync"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Manager handles WebSocket connection lifecycle and organization
type Manager struct {
	// Connection storage
	connections     map[*Connection]bool
	gameConnections map[string]map[*Connection]bool

	// Synchronization
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewManager creates a new connection manager
func NewManager() *Manager {
	return &Manager{
		connections:     make(map[*Connection]bool),
		gameConnections: make(map[string]map[*Connection]bool),
		logger:          logger.Get(),
	}
}

// RegisterConnection registers a new connection
func (m *Manager) RegisterConnection(connection *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connections[connection] = true
	m.logger.Debug("🔗 Client connected to server", zap.String("connection_id", connection.ID))
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
			m.logger.Debug("Removed empty game connections map", zap.String("game_id", gameID))
		}
	}

	// Close connection
	connection.Close()

	m.logger.Debug("⛓️‍💥 Client disconnected from server",
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

// FindConnectionByPlayer finds an existing connection for the given player
func (m *Manager) FindConnectionByPlayer(playerID, gameID string) *Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for connection := range m.connections {
		existingPlayerID, existingGameID := connection.GetPlayer()
		if existingPlayerID == playerID && existingGameID == gameID {
			return connection
		}
	}
	return nil
}

// CloseAllConnections closes all active connections
func (m *Manager) CloseAllConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("🛑 Closing all active connections", zap.Int("connection_count", len(m.connections)))

	for connection := range m.connections {
		connection.Close()
	}

	// Clear the connection maps
	m.connections = make(map[*Connection]bool)
	m.gameConnections = make(map[string]map[*Connection]bool)

	m.logger.Info("⛓️‍💥 All client connections closed by server")
}