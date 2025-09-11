package broadcast

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"

	"go.uber.org/zap"
)

// SystemMessages handles basic message broadcasting
type SystemMessages struct {
	manager *core.Manager
	logger  *zap.Logger
}

// NewSystemMessages creates a new system messages broadcaster
func NewSystemMessages(manager *core.Manager, logger *zap.Logger) *SystemMessages {
	return &SystemMessages{
		manager: manager,
		logger:  logger,
	}
}

// BroadcastToGame sends a message to all connections in a game
func (sm *SystemMessages) BroadcastToGame(gameID string, message dto.WebSocketMessage) {
	gameConns := sm.manager.GetGameConnections(gameID)

	if gameConns == nil || len(gameConns) == 0 {
		sm.logger.Warn("❌ No connections found for game", zap.String("game_id", gameID))
		return
	}

	sentCount := 0
	for connection := range gameConns {
		playerID, _ := connection.GetPlayer()
		sm.logger.Debug("📤 Sending message to individual connection",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID),
			zap.String("message_type", string(message.Type)))

		connection.SendMessage(message)
		sentCount++
	}

	sm.logger.Info("📢 Server broadcasted to game clients",
		zap.String("game_id", gameID),
		zap.String("message_type", string(message.Type)),
		zap.Int("messages_sent", sentCount))
}

// BroadcastToGameExcept sends a message to all connections in a game except the excluded connection
func (sm *SystemMessages) BroadcastToGameExcept(gameID string, message dto.WebSocketMessage, excludeConnection *core.Connection) {
	gameConns := sm.manager.GetGameConnections(gameID)

	if gameConns == nil {
		return
	}

	sentCount := 0
	for connection := range gameConns {
		if connection != excludeConnection {
			connection.SendMessage(message)
			sentCount++
		}
	}

	sm.logger.Debug("📢 Server broadcasting to game clients (excluding one)",
		zap.String("game_id", gameID),
		zap.String("message_type", string(message.Type)),
		zap.Int("sent_to_count", sentCount))
}

// SendToConnection sends a message to a specific connection
func (sm *SystemMessages) SendToConnection(connection *core.Connection, message dto.WebSocketMessage) {
	connection.SendMessage(message)

	sm.logger.Debug("💬 Server message sent to client",
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)))
}
