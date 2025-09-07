package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"

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
	case dto.MessageTypePlayerReconnect:
		h.handlePlayerReconnect(ctx, connection, message)
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
	h.broadcastToGame(payload.GameID, dto.WebSocketMessage{
		Type: dto.MessageTypePlayerConnected,
		Payload: dto.PlayerConnectedPayload{
			PlayerID:   playerID,
			PlayerName: payload.PlayerName,
			Game:       dto.ToGameDto(game),
		},
		GameID: game.ID,
	})

	h.logger.Info("🎮 Player connected via WebSocket",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", game.ID),
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

	// Handle different action types
	if err := h.processAction(ctx, gameID, playerID, payload.ActionRequest); err != nil {
		h.logger.Error("Failed to process action",
			zap.Error(err),
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.sendErrorToConnection(connection, fmt.Sprintf("Action failed: %v", err))
		return
	}

	// Broadcast updated game state to all players in the game
	game, err := h.gameService.GetGame(ctx, gameID)
	if err != nil {
		h.logger.Error("Failed to get updated game state after action",
			zap.Error(err),
			zap.String("game_id", gameID))
		h.sendErrorToConnection(connection, "Failed to get updated game state")
		return
	}

	h.broadcastToGame(gameID, dto.WebSocketMessage{
		Type: dto.MessageTypeFullState,
		Payload: dto.FullStatePayload{
			Game:     dto.ToGameDto(game),
			PlayerID: playerID,
		},
		GameID: gameID,
	})

	h.logger.Info("✅ Action processed and game state broadcasted",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.String("action_type", fmt.Sprintf("%T", payload.ActionRequest)))
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

// processAction processes different types of game actions
func (h *Hub) processAction(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	// Parse action request into a map to extract action type
	actionBytes, err := json.Marshal(actionRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal action request: %w", err)
	}

	var actionMap map[string]interface{}
	if err := json.Unmarshal(actionBytes, &actionMap); err != nil {
		return fmt.Errorf("failed to unmarshal action request: %w", err)
	}

	actionType, ok := actionMap["type"].(string)
	if !ok {
		return fmt.Errorf("action type not found or invalid")
	}

	// Handle different action types
	switch dto.ActionType(actionType) {
	case dto.ActionTypeStartGame:
		return h.handleStartGame(ctx, gameID, playerID)
	case dto.ActionTypeSellPatents:
		return h.handleSellPatents(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypeBuildPowerPlant:
		return h.handleBuildPowerPlant(ctx, gameID, playerID)
	case dto.ActionTypeLaunchAsteroid:
		return h.handleLaunchAsteroid(ctx, gameID, playerID)
	case dto.ActionTypeBuildAquifer:
		return h.handleBuildAquifer(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypePlantGreenery:
		return h.handlePlantGreenery(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypeBuildCity:
		return h.handleBuildCity(ctx, gameID, playerID, actionRequest)
	default:
		return fmt.Errorf("unsupported action type: %s", actionType)
	}
}

// handleStartGame handles the start game action
func (h *Hub) handleStartGame(ctx context.Context, gameID, playerID string) error {
	return h.gameService.StartGame(ctx, gameID, playerID)
}

//err = s.actionHandlers.StartGame.Handle(game, player, request)

// handleSellPatents handles sell patents standard project
func (h *Hub) handleSellPatents(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionSellPatentsRequest
	if err := h.parseActionRequest(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid sell patents request: %w", err)
	}

	return h.standardProjectService.SellPatents(ctx, gameID, playerID, request.CardCount)
}

// handleBuildPowerPlant handles build power plant standard project
func (h *Hub) handleBuildPowerPlant(ctx context.Context, gameID, playerID string) error {
	return h.standardProjectService.BuildPowerPlant(ctx, gameID, playerID)
}

// handleLaunchAsteroid handles launch asteroid standard project
func (h *Hub) handleLaunchAsteroid(ctx context.Context, gameID, playerID string) error {
	return h.standardProjectService.LaunchAsteroid(ctx, gameID, playerID)
}

// handleBuildAquifer handles build aquifer standard project
func (h *Hub) handleBuildAquifer(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionBuildAquiferRequest
	if err := h.parseActionRequest(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid build aquifer request: %w", err)
	}

	hexPosition := model.HexPosition{
		Q: request.HexPosition.Q,
		R: request.HexPosition.R,
		S: request.HexPosition.S,
	}

	return h.standardProjectService.BuildAquifer(ctx, gameID, playerID, hexPosition)
}

// handlePlantGreenery handles plant greenery standard project
func (h *Hub) handlePlantGreenery(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionPlantGreeneryRequest
	if err := h.parseActionRequest(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid plant greenery request: %w", err)
	}

	hexPosition := model.HexPosition{
		Q: request.HexPosition.Q,
		R: request.HexPosition.R,
		S: request.HexPosition.S,
	}

	return h.standardProjectService.PlantGreenery(ctx, gameID, playerID, hexPosition)
}

// handleBuildCity handles build city standard project
func (h *Hub) handleBuildCity(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionBuildCityRequest
	if err := h.parseActionRequest(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid build city request: %w", err)
	}

	hexPosition := model.HexPosition{
		Q: request.HexPosition.Q,
		R: request.HexPosition.R,
		S: request.HexPosition.S,
	}

	return h.standardProjectService.BuildCity(ctx, gameID, playerID, hexPosition)
}

// parseActionRequest parses an action request into the given destination
func (h *Hub) parseActionRequest(actionRequest interface{}, dest interface{}) error {
	actionBytes, err := json.Marshal(actionRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal action request: %w", err)
	}

	if err := json.Unmarshal(actionBytes, dest); err != nil {
		return fmt.Errorf("failed to unmarshal action request: %w", err)
	}

	return nil
}

// handlePlayerReconnect handles player reconnection requests
func (h *Hub) handlePlayerReconnect(ctx context.Context, connection *Connection, message dto.WebSocketMessage) {
	var payload dto.PlayerReconnectPayload
	if err := h.parseMessagePayload(message.Payload, &payload); err != nil {
		h.logger.Error("Failed to parse player reconnect payload",
			zap.Error(err),
			zap.String("connection_id", connection.ID))
		h.sendErrorToConnection(connection, "Invalid player reconnect payload")
		return
	}

	h.logger.Info("🔄 Player attempting to reconnect",
		zap.String("connection_id", connection.ID),
		zap.String("player_name", payload.PlayerName),
		zap.String("game_id", payload.GameID))

	// Validate game exists
	game, err := h.gameService.GetGame(ctx, payload.GameID)
	if err != nil {
		h.logger.Error("Failed to get game for reconnection",
			zap.String("game_id", payload.GameID),
			zap.String("player_name", payload.PlayerName),
			zap.Error(err))
		h.sendErrorToConnection(connection, "Game does not exist")
		return
	}

	// Find player by name (not ID) for cross-device support
	player, err := h.playerService.GetPlayerByName(ctx, payload.GameID, payload.PlayerName)
	if err != nil {
		h.logger.Error("Player not found for reconnection",
			zap.String("game_id", payload.GameID),
			zap.String("player_name", payload.PlayerName),
			zap.Error(err))
		h.sendErrorToConnection(connection, "Player not found in game")
		return
	}

	// Check if player is already connected elsewhere - disconnect the old connection
	h.mu.Lock()
	for existingConn := range h.connections {
		existingPlayerID, existingGameID := existingConn.GetPlayer()
		if existingPlayerID == player.ID && existingGameID == payload.GameID {
			h.logger.Info("🔀 Disconnecting existing connection for player",
				zap.String("existing_connection_id", existingConn.ID),
				zap.String("new_connection_id", connection.ID),
				zap.String("player_id", player.ID),
				zap.String("player_name", player.Name))

			// Send a message to the old connection about being replaced
			h.sendErrorToConnection(existingConn, "Connection replaced by new device")

			// Let the standard unregister process handle the cleanup
			// This prevents race conditions and ensures proper cleanup
			go func() {
				existingConn.Close() // This will trigger unregisterConnection via ReadPump/WritePump
			}()
			break
		}
	}
	h.mu.Unlock()

	// Associate the connection with the player and game
	connection.SetPlayer(player.ID, payload.GameID)
	h.addToGame(connection, payload.GameID)

	h.logger.Info("Connection ID: {}", zap.String("connection_id", connection.ID))

	// Update player connection status to connected
	err = h.playerService.UpdatePlayerConnectionStatus(ctx, payload.GameID, player.ID, model.ConnectionStatusConnected)
	if err != nil {
		h.logger.Error("Failed to update player connection status on reconnect",
			zap.String("player_id", player.ID),
			zap.String("game_id", payload.GameID),
			zap.Error(err))
	}

	// Get fresh game state after connection status update
	game, err = h.gameService.GetGame(ctx, payload.GameID)
	if err != nil {
		h.logger.Error("Failed to get updated game state after reconnection",
			zap.String("game_id", payload.GameID),
			zap.Error(err))
		h.sendErrorToConnection(connection, "Failed to retrieve game state")
		return
	}

	// Send player-reconnected message to ALL players (including the reconnecting player)
	reconnectedPayload := dto.PlayerReconnectedPayload{
		PlayerID:   player.ID,
		PlayerName: player.Name,
		Game:       dto.ToGameDto(game),
	}

	reconnectedMessage := dto.WebSocketMessage{
		Type:    dto.MessageTypePlayerReconnected,
		Payload: reconnectedPayload,
		GameID:  payload.GameID,
	}

	h.broadcastToGame(payload.GameID, reconnectedMessage)

	h.logger.Info("📢 Player reconnected successfully, broadcasted to game",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", player.ID),
		zap.String("player_name", player.Name),
		zap.String("game_id", payload.GameID))
}
