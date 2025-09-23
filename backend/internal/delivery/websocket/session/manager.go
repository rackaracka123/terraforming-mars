package session

import (
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// SessionManager manages WebSocket sessions and provides broadcasting capabilities
// This service is used by services to broadcast messages to players
type SessionManager interface {
	// Session management - one session (connection) per player per game
	RegisterSession(playerID, gameID string, sendMessage func(dto.WebSocketMessage))
	UnregisterSession(playerID, gameID string)

	// Core broadcasting operations
	BroadcastToGame(gameID string, messageType dto.MessageType, payload any) error
	BroadcastToGameExcept(gameID string, messageType dto.MessageType, payload any, excludePlayerID string) error
	SendToPlayer(playerID, gameID string, messageType dto.MessageType, payload any) error
}

// SessionManagerImpl implements the SessionManager interface
type SessionManagerImpl struct {
	// Session storage: game -> player -> connection
	sessions map[string]map[string]func(dto.WebSocketMessage)

	// Synchronization
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewSessionManager creates a new session manager
func NewSessionManager() SessionManager {
	return &SessionManagerImpl{
		sessions: make(map[string]map[string]func(dto.WebSocketMessage)),
		logger:   logger.Get(),
	}
}

// RegisterSession registers a new session
func (sm *SessionManagerImpl) RegisterSession(playerID, gameID string, sendMessage func(dto.WebSocketMessage)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Initialize game map if needed
	if sm.sessions[gameID] == nil {
		sm.sessions[gameID] = make(map[string]func(dto.WebSocketMessage))
	}

	// Register the session (replaces any existing session for this player)
	sm.sessions[gameID][playerID] = sendMessage

	sm.logger.Debug("✅ Session registered",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

// UnregisterSession removes a session
func (sm *SessionManagerImpl) UnregisterSession(playerID, gameID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if gameMap, exists := sm.sessions[gameID]; exists {
		delete(gameMap, playerID)
		// Clean up empty game map
		if len(gameMap) == 0 {
			delete(sm.sessions, gameID)
		}
	}

	sm.logger.Debug("❌ Session unregistered",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

// BroadcastToGame broadcasts a message to all players in a game
func (sm *SessionManagerImpl) BroadcastToGame(gameID string, messageType dto.MessageType, payload any) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	gameMap, exists := sm.sessions[gameID]
	if !exists || len(gameMap) == 0 {
		sm.logger.Debug("No sessions to broadcast to",
			zap.String("game_id", gameID),
			zap.String("message_type", string(messageType)))
		return nil
	}

	message := dto.WebSocketMessage{
		Type:    messageType,
		Payload: payload,
		GameID:  gameID,
	}

	sm.logger.Debug("📢 Broadcasting to game",
		zap.String("game_id", gameID),
		zap.String("message_type", string(messageType)),
		zap.Int("player_count", len(gameMap)))

	var lastError error
	successCount := 0

	for playerID, sendFunc := range gameMap {
		sendFunc(message)
		successCount++
		sm.logger.Debug("💬 Message sent to player",
			zap.String("player_id", playerID),
			zap.String("game_id", gameID),
			zap.String("message_type", string(messageType)))
	}

	sm.logger.Debug("📢 Broadcast completed",
		zap.String("game_id", gameID),
		zap.Int("successful_sends", successCount),
		zap.String("message_type", string(messageType)))

	return lastError
}

// BroadcastToGameExcept broadcasts a message to all players in a game except one
func (sm *SessionManagerImpl) BroadcastToGameExcept(gameID string, messageType dto.MessageType, payload any, excludePlayerID string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	gameMap, exists := sm.sessions[gameID]
	if !exists || len(gameMap) == 0 {
		return nil
	}

	message := dto.WebSocketMessage{
		Type:    messageType,
		Payload: payload,
		GameID:  gameID,
	}

	sm.logger.Debug("📢 Broadcasting to game (except player)",
		zap.String("game_id", gameID),
		zap.String("exclude_player_id", excludePlayerID),
		zap.String("message_type", string(messageType)))

	var lastError error
	successCount := 0

	for playerID, sendFunc := range gameMap {
		if playerID != excludePlayerID {
			sendFunc(message)
			successCount++
		}
	}

	sm.logger.Debug("📢 Broadcast completed (except player)",
		zap.String("game_id", gameID),
		zap.Int("successful_sends", successCount),
		zap.String("message_type", string(messageType)))

	return lastError
}

// SendToPlayer sends a message to a specific player
func (sm *SessionManagerImpl) SendToPlayer(playerID, gameID string, messageType dto.MessageType, payload any) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	gameMap, exists := sm.sessions[gameID]
	if !exists {
		return fmt.Errorf("game %s not found", gameID)
	}

	sendFunc, exists := gameMap[playerID]
	if !exists {
		return fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}

	message := dto.WebSocketMessage{
		Type:    messageType,
		Payload: payload,
		GameID:  gameID,
	}

	sendFunc(message)

	sm.logger.Debug("💬 Message sent to player",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.String("message_type", string(messageType)))

	return nil
}