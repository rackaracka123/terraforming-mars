package websocket

import (
	"context"
	"encoding/json"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/store"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WebSocketActionDispatcher handles WebSocket messages and dispatches actions to the store
type WebSocketActionDispatcher struct {
	store   *store.Store
	manager *core.Manager
}

// NewWebSocketActionDispatcher creates a new WebSocket action dispatcher
func NewWebSocketActionDispatcher(appStore *store.Store, manager *core.Manager) *WebSocketActionDispatcher {
	return &WebSocketActionDispatcher{
		store:   appStore,
		manager: manager,
	}
}

// HandleMessage processes WebSocket messages and dispatches appropriate actions
func (d *WebSocketActionDispatcher) HandleMessage(ctx context.Context, connection *core.Connection, message interface{}) {
	// Convert message to WebSocketMessage
	wsMessage, ok := message.(dto.WebSocketMessage)
	if !ok {
		logger.Warn("❓ Invalid message type received")
		return
	}
	logger.Debug("🎯 Dispatching WebSocket message to store",
		zap.String("message_type", string(wsMessage.Type)),
		zap.String("connection_id", connection.ID))

	switch wsMessage.Type {
	case dto.MessageTypePlayerConnect:
		d.handlePlayerConnect(ctx, connection, wsMessage)
	case "join-game":
		d.handleJoinGame(ctx, connection, wsMessage)
	case dto.MessageTypeActionStartGame:
		d.handleStartGame(ctx, connection, wsMessage)
	case dto.MessageTypeActionSkipAction:
		d.handleSkipAction(ctx, connection, wsMessage)
	case dto.MessageTypeActionSellPatents:
		d.handleStandardProject(ctx, connection, wsMessage)
	case dto.MessageTypeActionSelectStartingCard:
		d.handleSelectStartingCards(ctx, connection, wsMessage)
	case dto.MessageTypeActionBuildPowerPlant, dto.MessageTypeActionLaunchAsteroid, dto.MessageTypeActionBuildAquifer,
		dto.MessageTypeActionPlantGreenery, dto.MessageTypeActionBuildCity:
		d.handleStandardProject(ctx, connection, wsMessage)
	default:
		logger.Warn("❓ Unknown message type", zap.String("type", string(wsMessage.Type)))
	}
}

func (d *WebSocketActionDispatcher) handlePlayerConnect(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	// Parse player connect payload
	var payload struct {
		GameID     string `json:"gameId"`
		PlayerName string `json:"playerName,omitempty"`
		PlayerID   string `json:"playerId,omitempty"`
	}

	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		logger.Error("Failed to marshal player connect payload", zap.Error(err))
		return
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		logger.Error("Failed to parse player connect payload", zap.Error(err))
		return
	}

	// Validate required fields
	if payload.PlayerName == "" {
		logger.Error("PlayerName required for player connection")
		return
	}

	// Check if this is a reconnection (frontend provides existing playerId)
	var playerID string
	var isReconnection bool
	if payload.PlayerID != "" {
		// Frontend provided playerId - validate it exists in the game
		_, gameExists := d.store.GetGame(payload.GameID)
		if !gameExists {
			logger.Error("Game not found for reconnection attempt",
				zap.String("game_id", payload.GameID),
				zap.String("player_id", payload.PlayerID))
			return
		}

		playerState, playerExists := d.store.GetPlayer(payload.PlayerID)
		if !playerExists || playerState.GameID() != payload.GameID {
			// Invalid playerID or player not in this game - treat as new player
			logger.Warn("⚠️ Invalid playerId for reconnection, creating new player",
				zap.String("provided_player_id", payload.PlayerID),
				zap.String("player_name", payload.PlayerName),
				zap.String("game_id", payload.GameID))
			playerID = uuid.New().String()
			isReconnection = false
		} else {
			// Valid reconnection
			playerID = payload.PlayerID
			isReconnection = true
			logger.Info("🔄 Valid player reconnection",
				zap.String("player_id", playerID),
				zap.String("player_name", payload.PlayerName),
				zap.String("game_id", payload.GameID))
		}
	} else {
		// No playerId provided - create new player
		playerID = uuid.New().String()
		isReconnection = false
		logger.Info("🆕 Creating new player",
			zap.String("player_id", playerID),
			zap.String("player_name", payload.PlayerName),
			zap.String("game_id", payload.GameID))
	}
	playerName := payload.PlayerName

	// Create join game action
	action := store.JoinGameAction(payload.GameID, playerID, playerName, "websocket")

	// Dispatch to store
	if err := d.store.Dispatch(ctx, action); err != nil {
		logger.Error("Failed to dispatch player connect action", zap.Error(err))
		return
	}

	// Update connection with player info
	connection.SetPlayer(playerID, payload.GameID)

	// Add connection to the game's connection pool
	d.manager.AddToGame(connection, payload.GameID)

	// Send immediate response to the connecting client
	d.sendPlayerConnectionResponse(connection, playerID, playerName, payload.GameID, isReconnection)

	logger.Info("✅ Player connected via reducer",
		zap.String("player_id", playerID),
		zap.String("game_id", payload.GameID),
		zap.String("player_name", playerName))
}

func (d *WebSocketActionDispatcher) handleJoinGame(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	// Parse join game payload
	var payload struct {
		GameID     string `json:"gameId"`
		PlayerName string `json:"playerName"`
	}

	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		logger.Error("Failed to marshal payload", zap.Error(err))
		return
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		logger.Error("Failed to parse join game payload", zap.Error(err))
		return
	}

	// Generate player ID
	playerID := uuid.New().String()

	// Create join game action
	action := store.JoinGameAction(payload.GameID, playerID, payload.PlayerName, "websocket")

	// Dispatch to store
	if err := d.store.Dispatch(ctx, action); err != nil {
		logger.Error("Failed to dispatch join game action", zap.Error(err))
		return
	}

	// Update connection with player info
	connection.SetPlayer(playerID, payload.GameID)

	logger.Info("✅ Player joined game via reducer",
		zap.String("player_id", playerID),
		zap.String("game_id", payload.GameID),
		zap.String("player_name", payload.PlayerName))
}

func (d *WebSocketActionDispatcher) handleStartGame(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		logger.Warn("Start game received from unassigned connection")
		return
	}

	// Create start game action
	action := store.StartGameAction(gameID, playerID, "websocket")

	// Dispatch to store
	if err := d.store.Dispatch(ctx, action); err != nil {
		logger.Error("Failed to dispatch start game action", zap.Error(err))
		return
	}

	logger.Info("✅ Game started via reducer",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

func (d *WebSocketActionDispatcher) handleSkipAction(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		logger.Warn("Skip action received from unassigned connection")
		return
	}

	// Create skip action using DTO structure
	action := store.SkipActionAction(gameID, playerID, "websocket")

	// Dispatch to store
	if err := d.store.Dispatch(ctx, action); err != nil {
		logger.Error("Failed to dispatch skip action", zap.Error(err))
		return
	}

	logger.Info("✅ Action skipped via reducer",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

func (d *WebSocketActionDispatcher) handleStandardProject(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		logger.Warn("Standard project received from unassigned connection")
		return
	}

	// Create standard project action using message payload directly
	action := store.NewAction(store.ActionType(message.Type), message.Payload, gameID, playerID, "websocket")

	// Dispatch to store
	if err := d.store.Dispatch(ctx, action); err != nil {
		logger.Error("Failed to dispatch standard project action", zap.Error(err))
		return
	}

	logger.Info("🏗️ Standard project executed via reducer",
		zap.String("action_type", string(message.Type)),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

func (d *WebSocketActionDispatcher) handleSelectStartingCards(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		logger.Warn("Select starting cards received from unassigned connection")
		return
	}

	// Parse the frontend payload (just cardIds)
	var frontendPayload struct {
		CardIds []string `json:"cardIds"`
	}

	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		logger.Error("Failed to marshal select starting cards payload", zap.Error(err))
		return
	}

	if err := json.Unmarshal(payloadBytes, &frontendPayload); err != nil {
		logger.Error("Failed to parse select starting cards payload", zap.Error(err))
		return
	}

	// Calculate cost: all cards cost 3 MC each
	cost := len(frontendPayload.CardIds) * 3

	// Create proper backend payload with derived values
	backendPayload := store.SelectStartingCardsPayload{
		GameID:        gameID,
		PlayerID:      playerID,
		SelectedCards: frontendPayload.CardIds,
		Cost:          cost,
	}

	// Create select starting cards action with proper payload
	action := store.NewAction(store.ActionSelectStartingCards, backendPayload, gameID, playerID, "websocket")

	// Dispatch to store
	if err := d.store.Dispatch(ctx, action); err != nil {
		logger.Error("Failed to dispatch select starting cards action", zap.Error(err))
		return
	}

	logger.Info("🃏 Starting cards selected via reducer",
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.Int("card_count", len(frontendPayload.CardIds)),
		zap.Int("total_cost", cost))
}

// sendPlayerConnectionResponse sends the response message to the connecting client
func (d *WebSocketActionDispatcher) sendPlayerConnectionResponse(connection *core.Connection, playerID, playerName, gameID string, isReconnection bool) {
	// Get the current game state from store
	gameState, exists := d.store.GetGame(gameID)
	if !exists {
		logger.Error("Game not found when sending connection response", zap.String("game_id", gameID))
		return
	}

	// Convert game to DTO for response
	gameDto := dto.ToGameDtoBasic(gameState.Game())

	// Send appropriate message based on connection type
	if isReconnection {
		// Send player-reconnected message
		payload := dto.PlayerReconnectedPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
			Game:       gameDto,
		}
		message := dto.WebSocketMessage{
			Type:    dto.MessageTypePlayerReconnected,
			Payload: payload,
			GameID:  gameID,
		}
		connection.SendMessage(message)
		logger.Debug("📤 Sent player-reconnected response", zap.String("player_id", playerID))
	} else {
		// Send player-connected message
		payload := dto.PlayerConnectedPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
			Game:       gameDto,
		}
		message := dto.WebSocketMessage{
			Type:    dto.MessageTypePlayerConnected,
			Payload: payload,
			GameID:  gameID,
		}
		connection.SendMessage(message)
		logger.Debug("📤 Sent player-connected response", zap.String("player_id", playerID))
	}
}
