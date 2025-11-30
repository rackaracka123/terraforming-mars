package game

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// JoinGameHandler handles join game requests using the migrated architecture
type JoinGameHandler struct {
	joinGameAction *action.JoinGameAction
	logger         *zap.Logger
}

// NewJoinGameHandler creates a new join game handler for migrated actions
func NewJoinGameHandler(
	joinGameAction *action.JoinGameAction,
) *JoinGameHandler {
	return &JoinGameHandler{
		joinGameAction: joinGameAction,
		logger:         logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *JoinGameHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🎮 Processing join game request (migrated)")

	// Parse payload
	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "Invalid payload format")
		return
	}

	// Extract fields
	gameID, _ := payloadMap["gameId"].(string)
	playerName, _ := payloadMap["playerName"].(string)
	playerID, _ := payloadMap["playerId"].(string)

	if gameID == "" {
		log.Error("Missing gameId")
		h.sendError(connection, "Missing gameId")
		return
	}

	if playerName == "" {
		log.Error("Missing playerName")
		h.sendError(connection, "Missing playerName")
		return
	}

	// Generate playerID for new players (session-level identifier)
	// PlayerID persists across games and enables reconnection
	if playerID == "" {
		playerID = uuid.New().String()
		log.Debug("Generated new playerID for session",
			zap.String("player_id", playerID))
	}

	log.Debug("Parsed join game request",
		zap.String("game_id", gameID),
		zap.String("player_name", playerName),
		zap.String("player_id", playerID))

	// CRITICAL: Register connection BEFORE executing action
	// This ensures automatic broadcasting works (connection is findable when events fire)
	connection.SetPlayer(playerID, gameID)

	// Execute the migrated join game action with pre-generated playerID
	result, err := h.joinGameAction.Execute(ctx, gameID, playerName, playerID)
	if err != nil {
		log.Error("Failed to execute join game action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Join game action completed successfully",
		zap.String("player_id", result.PlayerID))

	// Send minimal success response confirming join
	// Note: Full game state was already broadcast via automatic broadcasting (PlayerJoinedEvent)
	response := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayerConnected,
		GameID: gameID,
		Payload: map[string]interface{}{
			"playerID":   result.PlayerID,
			"playerName": playerName,
			"success":    true,
		},
	}

	connection.Send <- response
	log.Info("📤 Sent player connected confirmation")
	log.Debug("📡 Automatic broadcast already sent game state to all players")
}

// sendError sends an error message to the client
func (h *JoinGameHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
