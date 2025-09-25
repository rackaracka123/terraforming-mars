package test

import (
	"terraforming-mars-backend/internal/delivery/dto"
)

// MockSessionManager provides a no-op implementation for testing
type MockSessionManager struct{}

func NewMockSessionManager() *MockSessionManager {
	return &MockSessionManager{}
}

func (m *MockSessionManager) RegisterSession(playerID, gameID string, sendMessage func(dto.WebSocketMessage)) {
	// No-op for testing
}

func (m *MockSessionManager) UnregisterSession(playerID, gameID string) {
	// No-op for testing
}

func (m *MockSessionManager) Broadcast(gameID string) error {
	// No-op for testing
	return nil
}

func (m *MockSessionManager) Send(gameID string, playerID string) error {
	// No-op for testing
	return nil
}
