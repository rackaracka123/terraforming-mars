package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"terraforming-mars-backend/internal/usecase"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	gameUC     *usecase.GameUseCase
}

// Client represents a websocket client
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	playerID string
	gameID   string
}

// GameMessage represents a websocket message
type GameMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// JoinGameMessage represents a join game request
type JoinGameMessage struct {
	GameID     string `json:"gameId"`
	PlayerName string `json:"playerName"`
}

// SelectCorporationMessage represents a corporation selection
type SelectCorporationMessage struct {
	CorporationID string `json:"corporationId"`
}

// GameActionMessage represents a game action
type GameActionMessage struct {
	GameID string `json:"gameId"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin in development
	},
}

// NewHub creates a new websocket hub
func NewHub(gameUC *usecase.GameUseCase) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		gameUC:     gameUC,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected: %s", client.conn.RemoteAddr())

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected: %s", client.conn.RemoteAddr())
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// ServeWS handles websocket requests from clients
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// BroadcastToGame sends a message to all clients in a specific game
func (h *Hub) BroadcastToGame(gameID string, messageType string, data interface{}) {
	message := GameMessage{
		Type: messageType,
		Data: data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for client := range h.clients {
		if client.gameID == gameID {
			select {
			case client.send <- jsonData:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var gameMsg GameMessage
		if err := json.Unmarshal(message, &gameMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		c.handleMessage(gameMsg)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

// handleMessage processes incoming websocket messages
func (c *Client) handleMessage(msg GameMessage) {
	switch msg.Type {
	case "join-game":
		c.handleJoinGame(msg)
	case "select-corporation":
		c.handleSelectCorporation(msg)
	case "raise-temperature":
		c.handleRaiseTemperature(msg)
	case "skip-action":
		c.handleSkipAction(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleJoinGame processes join game requests
func (c *Client) handleJoinGame(msg GameMessage) {
	var joinData JoinGameMessage
	jsonData, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(jsonData, &joinData); err != nil {
		c.sendError("Invalid join game data")
		return
	}

	// Generate player ID from connection
	playerID := c.conn.RemoteAddr().String()
	c.playerID = playerID
	c.gameID = joinData.GameID

	game, err := c.hub.gameUC.JoinGame(joinData.GameID, playerID, joinData.PlayerName)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	// Send available corporations to the new player
	corporations := c.hub.gameUC.GetAvailableCorporations()
	c.sendMessage("corporations-available", corporations)

	// Broadcast updated game state
	c.hub.BroadcastToGame(joinData.GameID, "game-updated", game)
}

// handleSelectCorporation processes corporation selection
func (c *Client) handleSelectCorporation(msg GameMessage) {
	var selectData SelectCorporationMessage
	jsonData, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(jsonData, &selectData); err != nil {
		c.sendError("Invalid corporation selection data")
		return
	}

	game, err := c.hub.gameUC.SelectCorporation(c.gameID, c.playerID, selectData.CorporationID)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	// Broadcast updated game state
	c.hub.BroadcastToGame(c.gameID, "game-updated", game)
}

// handleRaiseTemperature processes temperature raising actions
func (c *Client) handleRaiseTemperature(msg GameMessage) {
	var actionData GameActionMessage
	jsonData, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(jsonData, &actionData); err != nil {
		c.sendError("Invalid action data")
		return
	}

	game, err := c.hub.gameUC.RaiseTemperature(actionData.GameID, c.playerID)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	// Broadcast updated game state
	c.hub.BroadcastToGame(actionData.GameID, "game-updated", game)
}

// handleSkipAction processes skip action requests
func (c *Client) handleSkipAction(msg GameMessage) {
	var actionData GameActionMessage
	jsonData, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(jsonData, &actionData); err != nil {
		c.sendError("Invalid action data")
		return
	}

	game, err := c.hub.gameUC.SkipAction(actionData.GameID, c.playerID)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	// Broadcast updated game state
	c.hub.BroadcastToGame(actionData.GameID, "game-updated", game)
}

// sendMessage sends a message to this client
func (c *Client) sendMessage(messageType string, data interface{}) {
	message := GameMessage{
		Type: messageType,
		Data: data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case c.send <- jsonData:
	default:
		close(c.send)
		delete(c.hub.clients, c)
	}
}

// sendError sends an error message to this client
func (c *Client) sendError(message string) {
	c.sendMessage("error", map[string]string{"message": message})
}