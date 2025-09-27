package test

import (
	"terraforming-mars-backend/internal/delivery/dto"
)

// MockSessionManager provides a no-op implementation for testing
type MockSessionManager struct {
	broadcastCount int
}

func NewMockSessionManager() *MockSessionManager {
	return &MockSessionManager{
		broadcastCount: 0,
	}
}

func (m *MockSessionManager) RegisterSession(playerID, gameID string, sendMessage func(dto.WebSocketMessage)) {
	// No-op for testing
}

func (m *MockSessionManager) UnregisterSession(playerID, gameID string) {
	// No-op for testing
}

func (m *MockSessionManager) Broadcast(gameID string) error {
	// Track broadcast calls for testing
	m.broadcastCount++
	return nil
}

func (m *MockSessionManager) Send(gameID string, playerID string) error {
	// No-op for testing
	return nil
}

// GetBroadcastCount returns the number of times Broadcast was called
func (m *MockSessionManager) GetBroadcastCount() int {
	return m.broadcastCount
}

// ResetBroadcastCount resets the broadcast counter to 0
func (m *MockSessionManager) ResetBroadcastCount() {
	m.broadcastCount = 0
}
