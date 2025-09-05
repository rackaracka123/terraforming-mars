package websocket

import (
	"context"
	"encoding/json"
	"terraforming-mars-backend/internal/delivery/dto"

	"go.uber.org/zap"
)

// handleMessage processes incoming WebSocket messages and delegates to appropriate handlers
func (h *Hub) handleMessage(ctx context.Context, hubMessage HubMessage) {
	connection := hubMessage.Connection
	message := hubMessage.Message
	
	h.logger.Debug("Processing WebSocket message",
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)))
	
	switch message.Type {
	case dto.MessageTypePlayerConnect:
		h.handlePlayerConnect(ctx, connection, message)
	case dto.MessageTypePlayAction:
		h.handlePlayAction(ctx, connection, message)
	default:
		h.logger.Warn("Unknown message type received",
			zap.String("connection_id", connection.ID),
			zap.String("message_type", string(message.Type)))
		
		h.sendErrorToConnection(connection, "Unknown message type")
	}
}

// handlePlayerConnect handles player connection requests
func (h *Hub) handlePlayerConnect(ctx context.Context, connection *Connection, message dto.WebSocketMessage) {
	var payload dto.PlayerConnectPayload
	if err := h.parseMessagePayload(message.Payload, &payload); err != nil {
		h.logger.Error("Failed to parse player connect payload",
			zap.Error(err),
			zap.String("connection_id", connection.ID))
		h.sendErrorToConnection(connection, "Invalid player connect payload")
		return
	}
	
	// Delegate to service
	game, err := h.gameService.JoinGame(ctx, payload.GameID, payload.PlayerName)
	if err != nil {
		h.logger.Error("Failed to join game via WebSocket",
			zap.Error(err),
			zap.String("connection_id", connection.ID),
			zap.String("game_id", payload.GameID),
			zap.String("player_name", payload.PlayerName))
		h.sendErrorToConnection(connection, "Failed to join game")
		return
	}
	
	// Find the player ID of the newly joined player
	var playerID string
	for _, player := range game.Players {
		if player.Name == payload.PlayerName {
			playerID = player.ID
			break
		}
	}
	
	// Associate connection with player and game
	connection.SetPlayer(playerID, payload.GameID)
	h.addToGame(connection, payload.GameID)
	
	// Send player connected confirmation to the joining player
	h.sendToConnection(connection, dto.WebSocketMessage{
		Type: dto.MessageTypePlayerConnected,
		Payload: dto.PlayerConnectedPayload{
			PlayerID:   playerID,
			PlayerName: payload.PlayerName,
		},
		GameID: payload.GameID,
	})
	
	// Broadcast full game state to all players in the game
	h.broadcastToGame(payload.GameID, dto.WebSocketMessage{
		Type: dto.MessageTypeFullState,
		Payload: dto.FullStatePayload{
			Game:     dto.ToGameDto(game),
			PlayerID: playerID,
		},
		GameID: payload.GameID,
	})
	
	h.logger.Info("Player connected via WebSocket",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", payload.GameID),
		zap.String("player_name", payload.PlayerName))
}

// handlePlayAction handles game action requests
func (h *Hub) handlePlayAction(ctx context.Context, connection *Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.sendErrorToConnection(connection, "You must connect to a game first")
		return
	}
	
	var payload dto.PlayActionPayload
	if err := h.parseMessagePayload(message.Payload, &payload); err != nil {
		h.logger.Error("Failed to parse play action payload",
			zap.Error(err),
			zap.String("connection_id", connection.ID))
		h.sendErrorToConnection(connection, "Invalid action payload")
		return
	}
	
	// TODO: Delegate to appropriate service based on action type
	// For now, just log the action
	h.logger.Info("Action received via WebSocket",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.Any("action", payload.ActionRequest))
	
	// Send acknowledgment back to the player
	h.sendToConnection(connection, dto.WebSocketMessage{
		Type:    dto.MessageTypeError,
		Payload: dto.ErrorPayload{Message: "Action processing not yet implemented"},
		GameID:  gameID,
	})
}

// parseMessagePayload parses a message payload into the given destination
func (h *Hub) parseMessagePayload(payload interface{}, dest interface{}) error {
	// Convert payload to JSON bytes and then unmarshal to the destination
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(payloadBytes, dest)
}

// sendErrorToConnection sends an error message to a connection
func (h *Hub) sendErrorToConnection(connection *Connection, message string) {
	_, gameID := connection.GetPlayer()
	
	errorMessage := dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: dto.ErrorPayload{
			Message: message,
		},
		GameID: gameID,
	}
	
	h.sendToConnection(connection, errorMessage)
}