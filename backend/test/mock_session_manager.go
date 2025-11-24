package test

import (
	"terraforming-mars-backend/internal/session"
)

// MockSessionManager provides a no-op implementation for testing
// Implements both SessionManager and SessionManagerFactory interfaces
type MockSessionManager struct {
	broadcastCount int
	gameID         string
}

func NewMockSessionManager() *MockSessionManager {
	return &MockSessionManager{
		broadcastCount: 0,
		gameID:         "",
	}
}

// SessionManager interface methods
func (m *MockSessionManager) Broadcast() error {
	// Track broadcast calls for testing
	m.broadcastCount++
	return nil
}

func (m *MockSessionManager) Send(playerID string) error {
	// No-op for testing
	return nil
}

func (m *MockSessionManager) GetGameID() string {
	return m.gameID
}

// SessionManagerFactory interface methods
func (m *MockSessionManager) GetOrCreate(gameID string) session.SessionManager {
	// Return self as session manager (for simple test cases)
	m.gameID = gameID
	return m
}

func (m *MockSessionManager) Remove(gameID string) {
	// No-op for testing
}

// Test helper methods
func (m *MockSessionManager) GetBroadcastCount() int {
	return m.broadcastCount
}

func (m *MockSessionManager) ResetBroadcastCount() {
	m.broadcastCount = 0
}
