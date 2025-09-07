package websocket

import (
	"context"
	"sync"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/events"
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
	cardService             service.CardService

	// Event system
	eventBus events.EventBus

	// Synchronization
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewHub creates a new WebSocket hub
func NewHub(gameService service.GameService, playerService service.PlayerService, globalParametersService service.GlobalParametersService, standardProjectService service.StandardProjectService, cardService service.CardService, eventBus events.EventBus) *Hub {
	hub := &Hub{
		connections:             make(map[*Connection]bool),
		gameConnections:         make(map[string]map[*Connection]bool),
		Register:                make(chan *Connection, 256),
		Unregister:              make(chan *Connection, 256),
		Broadcast:               make(chan HubMessage, 256),
		gameService:             gameService,
		playerService:           playerService,
		globalParametersService: globalParametersService,
		standardProjectService:  standardProjectService,
		cardService:             cardService,
		eventBus:                eventBus,
		logger:                  logger.Get(),
	}

	// Subscribe to game state changes
	hub.subscribeToEvents()

	return hub
}

// Run starts the hub and handles connection management
func (h *Hub) Run(ctx context.Context) {
	h.logger.Info("🚀 Starting WebSocket hub")
	h.logger.Info("✅ WebSocket hub ready to process messages")

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("🛑 WebSocket hub stopping due to context cancellation")
			h.closeAllConnections()
			return

		case connection := <-h.Register:
			h.logger.Info("🔗 Processing Register request", zap.String("connection_id", connection.ID))
			h.registerConnection(connection)
			h.logger.Info("✅ Register request processed", zap.String("connection_id", connection.ID))

		case connection := <-h.Unregister:
			h.logger.Info("⛓️‍💥 Processing Unregister request", zap.String("connection_id", connection.ID))
			h.unregisterConnection(connection)
			h.logger.Info("✅ Unregister request processed", zap.String("connection_id", connection.ID))

		case hubMessage := <-h.Broadcast:
			h.logger.Debug("📨 Processing Broadcast message",
				zap.String("connection_id", hubMessage.Connection.ID),
				zap.String("message_type", string(hubMessage.Message.Type)))
			h.handleMessage(ctx, hubMessage)
			h.logger.Debug("✅ Broadcast message processed", zap.String("connection_id", hubMessage.Connection.ID))
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
			// Player disconnection handled by repository layer
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

	h.logger.Debug("📢 Broadcasting to game - checking connections",
		zap.String("game_id", gameID),
		zap.String("message_type", string(message.Type)),
		zap.Bool("has_connections", gameConns != nil),
		zap.Int("connection_count", len(gameConns)))

	if gameConns == nil {
		h.logger.Warn("❌ No connections found for game", zap.String("game_id", gameID))
		return
	}

	if len(gameConns) == 0 {
		h.logger.Warn("❌ Empty connection list for game", zap.String("game_id", gameID))
		return
	}

	sentCount := 0
	for connection := range gameConns {
		h.logger.Debug("📤 Sending message to individual connection",
			zap.String("connection_id", connection.ID),
			zap.String("game_id", gameID))
		connection.SendMessage(message)
		sentCount++
	}

	h.logger.Info("📢 Server broadcasted to game clients",
		zap.String("game_id", gameID),
		zap.String("message_type", string(message.Type)),
		zap.Int("total_connections", len(gameConns)),
		zap.Int("messages_sent", sentCount))
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

// sendToPlayer sends a message to a specific player in a game
func (h *Hub) sendToPlayer(gameID, playerID string, message dto.WebSocketMessage) {
	h.mu.RLock()
	gameConns := h.gameConnections[gameID]
	h.mu.RUnlock()

	if gameConns == nil {
		h.logger.Warn("⚠️ No connections found for game", zap.String("game_id", gameID))
		return
	}

	for conn := range gameConns {
		if conn.PlayerID == playerID {
			h.sendToConnection(conn, message)
			h.logger.Info("📬 Message sent to specific player",
				zap.String("game_id", gameID),
				zap.String("player_id", playerID),
				zap.String("message_type", string(message.Type)))
			return
		}
	}

	h.logger.Warn("⚠️ No connection found for player",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID))
}

// closeAllConnections closes all active connections
func (h *Hub) closeAllConnections() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Info("🛑 Closing all active connections", zap.Int("connection_count", len(h.connections)))

	for connection := range h.connections {
		// Use the proper Close method to ensure all goroutines are signaled
		connection.Close()
	}

	// Clear the connection maps
	h.connections = make(map[*Connection]bool)
	h.gameConnections = make(map[string]map[*Connection]bool)

	h.logger.Info("⛓️‍💥 All client connections closed by server")
}

// subscribeToEvents sets up event listeners for the hub
func (h *Hub) subscribeToEvents() {
	// Subscribe to game state changes to broadcast updates to clients
	h.eventBus.Subscribe(events.EventTypeGameStateChanged, h.handleGameStateChanged)

	// Subscribe to starting card options events to send available cards to players
	h.eventBus.Subscribe(events.EventTypePlayerStartingCardOptions, h.handlePlayerStartingCardOptions)

	h.logger.Info("📡 WebSocket hub subscribed to game state events")
}

// handleGameStateChanged handles game state change events by broadcasting to clients
func (h *Hub) handleGameStateChanged(ctx context.Context, event events.Event) error {
	h.logger.Info("🔄 Event handler called: handleGameStateChanged - running async to prevent deadlock")

	// Run the event handler asynchronously to prevent deadlocks
	go func() {
		// Extract the game state changed event data
		gameStateEvent, ok := event.(*events.GameStateChangedEvent)
		if !ok {
			h.logger.Error("❌ Invalid event type for GameStateChanged handler")
			return
		}

		h.logger.Debug("✅ Event type validated successfully")

		// Get the game ID from the event data
		eventData := gameStateEvent.GetPayload().(events.GameStateChangedEventData)
		gameID := eventData.GameID

		h.logger.Debug("📋 Game ID extracted from event", zap.String("game_id", gameID))

		// Create a new context for this async operation
		asyncCtx := context.Background()

		// Get the current game state to send to clients
		h.logger.Debug("🔍 Getting game state for broadcast (async)", zap.String("game_id", gameID))
		game, err := h.gameService.GetGame(asyncCtx, gameID)
		if err != nil {
			h.logger.Error("❌ Failed to get game for broadcast",
				zap.String("game_id", gameID),
				zap.Error(err))
			return
		}

		h.logger.Debug("✅ Game state retrieved for broadcast", zap.String("game_id", gameID))

		// Convert game to DTO for WebSocket message
		h.logger.Debug("🔄 Converting game to DTO for broadcast")
		gameDTO := dto.ToGameDto(game)
		h.logger.Debug("✅ Game DTO converted successfully")

		// Create game-updated message
		h.logger.Debug("🔄 Creating game-updated message")
		message := dto.WebSocketMessage{
			Type: dto.MessageTypeGameUpdated,
			Payload: dto.GameUpdatedPayload{
				Game: gameDTO,
			},
		}

		h.logger.Debug("✅ Game-updated message created successfully")

		// Broadcast to all clients in the game
		h.logger.Info("📢 Broadcasting game-updated message to clients", zap.String("game_id", gameID))
		h.broadcastToGame(gameID, message)

		h.logger.Info("✅ Game state change broadcasted to clients",
			zap.String("game_id", gameID),
			zap.String("game_phase", string(game.CurrentPhase)),
			zap.String("game_status", string(game.Status)))

		h.logger.Debug("✅ Event handler completed successfully")
	}()

	// Return immediately to prevent blocking the original operation
	h.logger.Debug("✅ Event handler scheduled asynchronously")
	return nil
}

// handlePlayerStartingCardOptions handles when starting cards are dealt to a player
func (h *Hub) handlePlayerStartingCardOptions(ctx context.Context, event events.Event) error {
	h.logger.Info("🃏 Event handler called: handlePlayerStartingCardOptions - running async to prevent deadlock")

	// Create async context to prevent deadlock
	go func() {

		// Parse event payload
		payload := event.GetPayload().(events.PlayerStartingCardOptionsEventData)
		gameID := payload.GameID
		playerID := payload.PlayerID
		cardOptions := payload.CardOptions

		h.logger.Debug("🃏 Processing starting card options for player",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Strings("card_options", cardOptions))

		// Get card details for the options
		allStartingCards := model.GetStartingCards()
		cardMap := make(map[string]model.Card)
		for _, card := range allStartingCards {
			cardMap[card.ID] = card
		}

		// Build card DTOs for the available options
		availableCards := make([]dto.CardDto, 0, len(cardOptions))
		for _, cardID := range cardOptions {
			if card, exists := cardMap[cardID]; exists {
				cardDto := dto.CardDto{
					ID:          card.ID,
					Name:        card.Name,
					Type:        dto.CardType(card.Type),
					Cost:        card.Cost,
					Description: card.Description,
				}
				availableCards = append(availableCards, cardDto)
			}
		}

		// Create available-cards message
		message := dto.WebSocketMessage{
			Type: dto.MessageTypeAvailableCards,
			Payload: dto.AvailableCardsPayload{
				Cards: availableCards,
			},
		}

		// Send to the specific player
		h.sendToPlayer(gameID, playerID, message)

		h.logger.Info("🃏 Available cards sent to player",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Int("cards_count", len(availableCards)))
	}()

	return nil
}
