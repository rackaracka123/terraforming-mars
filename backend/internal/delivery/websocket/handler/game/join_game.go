package game

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// JoinGameHandler handles join game requests using the migrated architecture
type JoinGameHandler struct {
	joinGameAction *action.JoinGameAction
	logger         *zap.Logger
}

// NewJoinGameHandler creates a new join game handler for migrated actions
func NewJoinGameHandler(joinGameAction *action.JoinGameAction) *JoinGameHandler {
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

	log.Debug("Parsed join game request",
		zap.String("game_id", gameID),
		zap.String("player_name", playerName),
		zap.String("player_id", playerID))

	// Execute the migrated join game action
	var result *action.JoinGameResult
	var err error

	if playerID != "" {
		result, err = h.joinGameAction.Execute(ctx, gameID, playerName, playerID)
	} else {
		result, err = h.joinGameAction.Execute(ctx, gameID, playerName)
	}

	if err != nil {
		log.Error("Failed to execute join game action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	// Update connection with player and game IDs
	connection.PlayerID = result.PlayerID
	connection.GameID = gameID

	log.Info("✅ Join game action completed successfully",
		zap.String("player_id", result.PlayerID))

	// Send success response with full game state
	response := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayerConnected,
		GameID: gameID,
		Payload: dto.PlayerConnectedPayload{
			PlayerID:   result.PlayerID,
			PlayerName: playerName,
			Game:       result.GameDto,
		},
	}

	connection.Send <- response

	log.Info("📤 Sent player connected response with game state")
	// Note: BroadcastEvent is published by the action, Broadcaster will handle game state updates
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
