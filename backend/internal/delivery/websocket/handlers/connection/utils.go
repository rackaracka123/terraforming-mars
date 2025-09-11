package connection

import (
	"terraforming-mars-backend/internal/delivery/websocket/core"

	"go.uber.org/zap"
)

// SetupConnection configures the connection for a player
func SetupConnection(connection *core.Connection, manager *core.Manager, gameID, playerID string, isNewPlayer bool, logger *zap.Logger) bool {
	if isNewPlayer {
		// New player - use temporary player ID
		tempPlayerID := "temp-" + connection.ID
		connection.SetPlayer(tempPlayerID, gameID)
		manager.AddToGame(connection, gameID)
		logger.Debug("🔗 Connection set up for new player (temporary)",
			zap.String("connection_id", connection.ID),
			zap.String("temp_player_id", tempPlayerID))
	} else {
		// Existing player - use real player ID
		connection.SetPlayer(playerID, gameID)
		manager.AddToGame(connection, gameID)
		logger.Debug("🔗 Connection set up for existing player",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID))
	}
	return true
}