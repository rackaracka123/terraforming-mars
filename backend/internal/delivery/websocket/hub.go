package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"
)

// Hub maintains the set of active clients and broadcasts messages to clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Clients grouped by game ID for efficient broadcasting
	gameClients map[string][]*Client

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe access to client maps
	mutex sync.RWMutex

	// Game service for handling game logic
	gameService *service.GameService
}

// NewHub creates a new WebSocket hub
func NewHub(gameService *service.GameService) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		gameClients: make(map[string][]*Client),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		gameService: gameService,
	}
}

// Run starts the hub and handles client connections
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client] = true
	log.Printf("Client connected: %s", client.ID)
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)

		// Remove from game clients
		playerID, gameID := client.GetPlayerInfo()
		if gameID != "" {
			h.removeClientFromGame(client, gameID)
		}

		log.Printf("Client disconnected: %s (PlayerID: %s, GameID: %s)", client.ID, playerID, gameID)
	}
}

// removeClientFromGame removes a client from a game's client list
func (h *Hub) removeClientFromGame(client *Client, gameID string) {
	if clients, exists := h.gameClients[gameID]; exists {
		for i, c := range clients {
			if c == client {
				h.gameClients[gameID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}

		// Clean up empty game client lists
		if len(h.gameClients[gameID]) == 0 {
			delete(h.gameClients, gameID)
		}
	}
}

// addClientToGame adds a client to a game's client list
func (h *Hub) addClientToGame(client *Client, gameID string) {
	if _, exists := h.gameClients[gameID]; !exists {
		h.gameClients[gameID] = make([]*Client, 0)
	}
	h.gameClients[gameID] = append(h.gameClients[gameID], client)
}

// broadcastMessage sends a message to all connected clients
func (h *Hub) broadcastMessage(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// BroadcastToGame sends a message to all clients in a specific game
func (h *Hub) BroadcastToGame(gameID string, message *WebSocketMessage) {
	h.mutex.RLock()
	clients, exists := h.gameClients[gameID]
	if !exists {
		h.mutex.RUnlock()
		return
	}

	// Make a copy of the clients slice to avoid holding the lock during message sending
	clientsCopy := make([]*Client, len(clients))
	copy(clientsCopy, clients)
	h.mutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}

	for _, client := range clientsCopy {
		select {
		case client.send <- data:
		default:
			// Client is disconnected, clean up
			h.unregister <- client
		}
	}
}

// handlePlayerConnect processes player connection messages
func (h *Hub) handlePlayerConnect(client *Client, msg *WebSocketMessage) {
	var payload PlayerConnectPayload
	if err := parsePayload(msg.Payload, &payload); err != nil {
		client.sendError("Invalid player connect payload")
		return
	}

	// Get the game
	game, err := h.gameService.GetGame(payload.GameID)
	if err != nil {
		client.sendError("Game not found")
		return
	}

	// Check if player already exists in the game
	var playerID string
	playerExists := false
	for _, player := range game.Players {
		if player.Name == payload.PlayerName {
			playerID = player.ID
			playerExists = true
			break
		}
	}

	// If player doesn't exist, create them
	if !playerExists {
		updatedGame, err := h.gameService.JoinGame(payload.GameID, payload.PlayerName)
		if err != nil {
			client.sendError("Failed to join game: " + err.Error())
			return
		}

		// Find the newly created player
		for _, player := range updatedGame.Players {
			if player.Name == payload.PlayerName {
				playerID = player.ID
				break
			}
		}
		game = updatedGame
	}

	// Set client info
	client.SetPlayerInfo(playerID, payload.GameID)

	// Add client to game clients
	h.mutex.Lock()
	h.addClientToGame(client, payload.GameID)
	h.mutex.Unlock()

	// Send full state to the connecting player
	fullStateMsg := &WebSocketMessage{
		Type: MessageTypeFullState,
		Payload: FullStatePayload{
			Game:     dto.ToGameDto(game),
			PlayerID: playerID,
		},
		GameID: payload.GameID,
	}
	client.SendMessage(fullStateMsg)

	// Broadcast player connected to other clients in the game
	connectedMsg := &WebSocketMessage{
		Type: MessageTypePlayerConnected,
		Payload: PlayerConnectedPayload{
			PlayerID:   playerID,
			PlayerName: payload.PlayerName,
		},
		GameID: payload.GameID,
	}
	h.BroadcastToGame(payload.GameID, connectedMsg)

	log.Printf("Player %s connected to game %s", payload.PlayerName, payload.GameID)
}

// handlePlayAction processes game action messages
func (h *Hub) handlePlayAction(client *Client, msg *WebSocketMessage) {
	var payload PlayActionPayload
	if err := parsePayload(msg.Payload, &payload); err != nil {
		client.sendError("Invalid play action payload")
		return
	}

	playerID, gameID := client.GetPlayerInfo()
	if gameID == "" || playerID == "" {
		client.sendError("Not connected to a game")
		return
	}

	// Apply the action through the service
	game, err := h.gameService.ApplyAction(gameID, playerID, payload.ActionPayload)
	if err != nil {
		client.sendError("Failed to apply action: " + err.Error())
		return
	}

	// Broadcast updated game state to all clients in the game
	h.broadcastGameUpdate(gameID, game)
}

// broadcastGameUpdate broadcasts a game state update to all clients in the game
func (h *Hub) broadcastGameUpdate(gameID string, game *domain.Game) {
	message := &WebSocketMessage{
		Type: MessageTypeGameUpdated,
		Payload: GameUpdatedPayload{
			Game: dto.ToGameDto(game),
		},
		GameID: gameID,
	}

	h.BroadcastToGame(gameID, message)
}

// parsePayload parses a message payload into the target structure
func parsePayload(payload interface{}, target interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
