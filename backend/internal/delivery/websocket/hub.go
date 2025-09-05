package websocket

import (
	"context"
	"sync"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"
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
	playerService          service.PlayerService
	globalParametersService service.GlobalParametersService
	
	// Synchronization
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewHub creates a new WebSocket hub
func NewHub(gameService service.GameService, playerService service.PlayerService, globalParametersService service.GlobalParametersService) *Hub {
	return &Hub{
		connections:             make(map[*Connection]bool),
		gameConnections:         make(map[string]map[*Connection]bool),
		Register:                make(chan *Connection),
		Unregister:              make(chan *Connection),
		Broadcast:               make(chan HubMessage),
		gameService:             gameService,
		playerService:          playerService,
		globalParametersService: globalParametersService,
		logger:                  logger.Get(),
	}
}

// Run starts the hub and handles connection management
func (h *Hub) Run(ctx context.Context) {
	h.logger.Info("Starting WebSocket hub")
	
	for {
		select {
		case <-ctx.Done():
			h.logger.Info("WebSocket hub stopping due to context cancellation")
			h.closeAllConnections()
			return
			
		case connection := <-h.Register:
			h.registerConnection(connection)
			
		case connection := <-h.Unregister:
			h.unregisterConnection(connection)
			
		case hubMessage := <-h.Broadcast:
			h.handleMessage(ctx, hubMessage)
		}
	}
}

// registerConnection registers a new connection
func (h *Hub) registerConnection(connection *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.connections[connection] = true
	h.logger.Info("Connection registered", zap.String("connection_id", connection.ID))
}

// unregisterConnection unregisters a connection
func (h *Hub) unregisterConnection(connection *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if _, ok := h.connections[connection]; ok {
		// Remove from connections
		delete(h.connections, connection)
		close(connection.Send)
		
		// Remove from game connections if assigned
		playerID, gameID := connection.GetPlayer()
		if gameID != "" {
			if gameConns, exists := h.gameConnections[gameID]; exists {
				delete(gameConns, connection)
				if len(gameConns) == 0 {
					delete(h.gameConnections, gameID)
				}
			}
		}
		
		h.logger.Info("Connection unregistered",
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
	
	h.logger.Debug("Message broadcast to game",
		zap.String("game_id", gameID),
		zap.String("message_type", string(message.Type)),
		zap.Int("connection_count", len(gameConns)))
}

// sendToConnection sends a message to a specific connection
func (h *Hub) sendToConnection(connection *Connection, message dto.WebSocketMessage) {
	connection.SendMessage(message)
	
	h.logger.Debug("Message sent to connection",
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
	
	h.logger.Info("All connections closed")
}