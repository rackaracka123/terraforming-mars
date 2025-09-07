package websocket

import (
	"context"
	"sync"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// HubMessage represents a message received from a connection
type HubMessage struct {
	Connection *Connection
	Message    dto.WebSocketMessage
}

// Hub maintains active WebSocket connections and handles message routing
type Hub struct {
	// Registered connections
	connections map[*Connection]bool

	// Connections grouped by game ID for efficient broadcasting
	gameConnections map[string]map[*Connection]bool

	// Register requests from connections
	Register chan *Connection

	// Unregister requests from connections
	Unregister chan *Connection

	// Broadcast messages to connections
	Broadcast chan HubMessage

	// Services for handling business logic
	gameService             service.GameService
	playerService           service.PlayerService
	globalParametersService service.GlobalParametersService
	standardProjectService  service.StandardProjectService

	// Synchronization
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewHub creates a new WebSocket hub
func NewHub(gameService service.GameService, playerService service.PlayerService, globalParametersService service.GlobalParametersService, standardProjectService service.StandardProjectService) *Hub {
	return &Hub{
		connections:             make(map[*Connection]bool),
		gameConnections:         make(map[string]map[*Connection]bool),
		Register:                make(chan *Connection, 256),
		Unregister:              make(chan *Connection, 256),
		Broadcast:               make(chan HubMessage, 256),
		gameService:             gameService,
		playerService:           playerService,
		globalParametersService: globalParametersService,
		standardProjectService:  standardProjectService,
		logger:                  logger.Get(),
	}
}

// Run starts the hub and handles connection management
func (h *Hub) Run(ctx context.Context) {
	h.logger.Info("Starting WebSocket hub")
	h.logger.Info("WebSocket hub ready to process messages")

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("WebSocket hub stopping due to context cancellation")
			h.closeAllConnections()
			return

		case connection := <-h.Register:
			h.logger.Info("Processing Register request", zap.String("connection_id", connection.ID))
			h.registerConnection(connection)
			h.logger.Info("Register request processed", zap.String("connection_id", connection.ID))

		case connection := <-h.Unregister:
			h.logger.Info("Processing Unregister request", zap.String("connection_id", connection.ID))
			h.unregisterConnection(connection)
			h.logger.Info("Unregister request processed", zap.String("connection_id", connection.ID))

		case hubMessage := <-h.Broadcast:
			h.logger.Debug("Processing Broadcast message", zap.String("connection_id", hubMessage.Connection.ID))
			h.handleMessage(ctx, hubMessage)
			h.logger.Debug("Broadcast message processed", zap.String("connection_id", hubMessage.Connection.ID))
		}
	}
}

// registerConnection registers a new connection
func (h *Hub) registerConnection(connection *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.connections[connection] = true
	h.logger.Debug("🔗 Client connected to server", zap.String("connection_id", connection.ID))
}

// unregisterConnection unregisters a connection
func (h *Hub) unregisterConnection(connection *Connection) {
	// First, collect info we need while holding the lock
	h.mu.Lock()

	var playerID, gameID string
	var shouldBroadcast bool

	if _, ok := h.connections[connection]; ok {
		// Remove from connections
		delete(h.connections, connection)
		connection.CloseSend()

		// Get player info before releasing lock
		playerID, gameID = connection.GetPlayer()
		shouldBroadcast = gameID != "" && playerID != ""

		// Remove from game connections while still holding lock
		if gameConns, exists := h.gameConnections[gameID]; exists {
			if _, inGame := gameConns[connection]; inGame {
				delete(gameConns, connection)
				h.logger.Debug("Removed connection from game",
					zap.String("connection_id", connection.ID),
					zap.String("game_id", gameID),
					zap.Int("remaining_connections", len(gameConns)))
			} else {
				h.logger.Debug("Connection was not in game connections map",
					zap.String("connection_id", connection.ID),
					zap.String("game_id", gameID))
			}

			if len(gameConns) == 0 {
				delete(h.gameConnections, gameID)
				h.logger.Debug("Removed empty game connections map", zap.String("game_id", gameID))
			}
		}

		// Close the connection properly
		connection.Close()
	}

	h.mu.Unlock()

	if shouldBroadcast {
		// Update player connection status to disconnected
		ctx := context.Background()
		err := h.playerService.UpdatePlayerConnectionStatus(ctx, gameID, playerID, model.ConnectionStatusDisconnected)
		if err != nil {
			h.logger.Error("Failed to update player connection status on disconnect",
				zap.String("player_id", playerID),
				zap.String("game_id", gameID),
				zap.Error(err))
		} else {
			// Get updated game state and broadcast player-disconnected message
			game, err := h.gameService.GetGame(ctx, gameID)
			if err != nil {
				h.logger.Error("Failed to get game for disconnect broadcast",
					zap.String("game_id", gameID),
					zap.Error(err))
			} else {
				// Find the player to get their name
				var playerName string
				for _, player := range game.Players {
					if player.ID == playerID {
						playerName = player.Name
						break
					}
				}

				// Broadcast player-disconnected message to other players in the game
				disconnectedPayload := dto.PlayerDisconnectedPayload{
					PlayerID:   playerID,
					PlayerName: playerName,
					Game:       dto.ToGameDto(game),
				}

				disconnectedMessage := dto.WebSocketMessage{
					Type:    dto.MessageTypePlayerDisconnected,
					Payload: disconnectedPayload,
					GameID:  gameID,
				}

				h.broadcastToGameExcept(gameID, disconnectedMessage, connection)

				h.logger.Info("📢 Player disconnected, broadcasted to other players in game",
					zap.String("player_id", playerID),
					zap.String("player_name", playerName),
					zap.String("game_id", gameID))
			}
		}

		h.logger.Debug("⛓️‍💥 Client disconnected from server",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
	}
}

// addToGame adds a connection to a game group
func (h *Hub) addToGame(connection *Connection, gameID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.gameConnections[gameID] == nil {
		h.gameConnections[gameID] = make(map[*Connection]bool)
	}
	h.gameConnections[gameID][connection] = true
}

// broadcastToGame sends a message to all connections in a game
func (h *Hub) broadcastToGame(gameID string, message dto.WebSocketMessage) {
	h.mu.RLock()
	gameConns := h.gameConnections[gameID]
	h.mu.RUnlock()

	if gameConns == nil {
		return
	}

	for connection := range gameConns {
		connection.SendMessage(message)
	}

	h.logger.Debug("📢 Server broadcasting to game clients",
		zap.String("game_id", gameID),
		zap.String("message_type", string(message.Type)),
		zap.Int("connection_count", len(gameConns)))
}

// broadcastToGameExcept sends a message to all connections in a game except the excluded connection
func (h *Hub) broadcastToGameExcept(gameID string, message dto.WebSocketMessage, excludeConnection *Connection) {
	h.mu.RLock()
	gameConns := h.gameConnections[gameID]
	h.mu.RUnlock()

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

	h.logger.Debug("📢 Server broadcasting to game clients (excluding one)",
		zap.String("game_id", gameID),
		zap.String("message_type", string(message.Type)),
		zap.Int("total_connections", len(gameConns)),
		zap.Int("sent_to_count", sentCount))
}

// sendToConnection sends a message to a specific connection
func (h *Hub) sendToConnection(connection *Connection, message dto.WebSocketMessage) {
	connection.SendMessage(message)

	h.logger.Debug("💬 Server message sent to client",
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)))
}

// closeAllConnections closes all active connections
func (h *Hub) closeAllConnections() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for connection := range h.connections {
		close(connection.Send)
		connection.Conn.Close()
	}

	h.logger.Info("⛓️‍💥 All client connections closed by server")
}
