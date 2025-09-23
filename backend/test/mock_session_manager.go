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

func (m *MockSessionManager) BroadcastToGame(gameID string, messageType dto.MessageType, payload any) error {
	// No-op for testing
	return nil
}

func (m *MockSessionManager) BroadcastToGameExcept(gameID string, messageType dto.MessageType, payload any, excludePlayerID string) error {
	// No-op for testing
	return nil
}

func (m *MockSessionManager) SendToPlayer(playerID, gameID string, messageType dto.MessageType, payload any) error {
	// No-op for testing
	return nil
}
