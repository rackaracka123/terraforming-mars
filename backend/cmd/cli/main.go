package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Default server address
	defaultServerAddr = "localhost:3001"

	// CLI tool metadata
	cliVersion = "1.0.0"
	cliName    = "Terraforming Mars CLI"

	// HTTP API base URL
	httpAPIBase = "http://localhost:3001/api/v1"
)

// GameState holds the current game state for display
type GameState struct {
	Player           *model.Player
	Generation       int
	CurrentPhase     model.GamePhase
	GameID           string
	IsConnected      bool
	TotalPlayers     int
	GameStatus       model.GameStatus
	HostPlayerID     string
	GlobalParameters *GlobalParams
}

// GlobalParams represents the global parameters for display
type GlobalParams struct {
	Temperature int
	Oxygen      int
	Oceans      int
}

type CLIClient struct {
	conn      *websocket.Conn
	playerID  string
	gameID    string
	done      chan struct{}
	closed    bool
	ui        *UI
	gameState *GameState
}

func main() {
	fmt.Printf("%s v%s\n", cliName, cliVersion)
	fmt.Println("Interactive CLI for Terraforming Mars backend")
	fmt.Println("Type 'help' for available commands or 'quit' to exit")
	fmt.Println()

	// Get server address from args or use default
	serverAddr := defaultServerAddr
	if len(os.Args) > 1 {
		serverAddr = os.Args[1]
	}

	client := &CLIClient{
		playerID:  "cli-" + uuid.New().String()[:8],
		done:      make(chan struct{}),
		ui:        NewUI(),
		gameState: &GameState{},
	}

	// Connect to server
	if err := client.connect(serverAddr); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.conn.Close()

	fmt.Printf("✅ Connected to server at %s\n", serverAddr)
	fmt.Printf("🔧 CLI Player ID: %s\n\n", client.playerID)

	// Set up signal handling
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Set up window resize signal handling
	winResize := make(chan os.Signal, 1)
	signal.Notify(winResize, syscall.SIGWINCH)

	// Start message reader goroutine
	go client.readMessages()

	// Handle interrupt signal in goroutine
	go func() {
		<-interrupt
		fmt.Println("\n🛑 Shutting down CLI...")

		if !client.closed {
			client.closed = true
			close(client.done)
		}

		// Close connection gracefully
		client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		time.Sleep(time.Second)
		os.Exit(0)
	}()

	// Handle window resize signal in goroutine
	go func() {
		for {
			select {
			case <-winResize:
				// Terminal was resized, refresh the display
				client.refreshDisplay()
			case <-client.done:
				return
			}
		}
	}()

	// Run interactive command loop in main thread
	client.commandLoop()
}

func (c *CLIClient) connect(serverAddr string) error {
	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/ws"}

	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}

	return nil
}

func (c *CLIClient) readMessages() {
	// Don't close done channel here - only on real disconnect or user quit
	for {
		select {
		case <-c.done:
			return
		default:
			var message dto.WebSocketMessage
			err := c.conn.ReadJSON(&message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("❌ WebSocket error: %v\n", err)
					// Only close on actual connection errors
					if !c.closed {
						c.closed = true
						close(c.done)
					}
				}
				return
			}

			c.handleMessage(message)
		}
	}
}

func (c *CLIClient) handleMessage(message dto.WebSocketMessage) {
	// Set gameID if provided in any message
	if message.GameID != "" && c.gameID == "" {
		c.gameID = message.GameID
		c.gameState.GameID = message.GameID
		c.gameState.IsConnected = true
		c.ui.UpdateGameState(c.gameState)
		c.refreshDisplay()
	}

	switch message.Type {
	case dto.MessageTypeGameUpdated:
		c.updateGameStateFromMessage(message)
		c.refreshDisplay()

	case dto.MessageTypePlayerConnected:
		// Extract the actual player ID and game state from the payload
		if payload, ok := message.Payload.(map[string]interface{}); ok {
			if playerID, ok := payload["playerId"].(string); ok {
				c.playerID = playerID // Update to the actual player ID from the game
			}

			// Parse the nested game data to update full game state
			if gameData, ok := payload["game"].(map[string]interface{}); ok {
				c.parseGameData(gameData)
				c.gameID = message.GameID
				c.gameState.GameID = message.GameID
				c.gameState.IsConnected = true
				c.ui.UpdateGameState(c.gameState)
				c.refreshDisplay()
			}
		}

	case dto.MessageTypeError:
		if payload, ok := message.Payload.(map[string]interface{}); ok {
			if msg, ok := payload["message"].(string); ok {
				errorMsg := c.ui.RenderMessage("error", msg)
				c.ui.SetLastCommand("", errorMsg)
				c.refreshDisplay()
			}
		}

	case dto.MessageTypeFullState:
		c.updateGameStateFromMessage(message)
		c.gameID = message.GameID
		c.gameState.GameID = message.GameID
		c.gameState.IsConnected = true
		c.ui.UpdateGameState(c.gameState)
		c.refreshDisplay()
	}
}

func (c *CLIClient) commandLoop() {
	reader := bufio.NewReader(os.Stdin)

	// Initial display refresh
	c.refreshDisplay()

	for {
		// Check if we should exit
		select {
		case <-c.done:
			return
		default:
			// Continue with command processing
		}

		fmt.Print(c.ui.RenderPrompt())

		// Read line input
		line, err := reader.ReadString('\n')
		if err != nil {
			// Check if we're done due to EOF or closed channel
			select {
			case <-c.done:
				return
			default:
				// EOF reached or other error
				return
			}
		}

		command := strings.TrimSpace(line)
		if command == "" {
			continue
		}

		if c.processCommand(command) {
			return // quit command
		}
	}
}

func (c *CLIClient) processCommand(command string) bool {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "help", "h":
		c.showHelp()

	case "quit", "exit", "q":
		fmt.Println("👋 Goodbye!")
		if !c.closed {
			c.closed = true
			close(c.done)
		}
		return true

	case "status", "s":
		c.showStatus()

	case "caj":
		c.createAndJoinGame(args)

	case "join", "j":
		c.joinExistingGame(args)

	case "games":
		c.listGames()

	case "players":
		c.listPlayers()

	case "actions":
		c.showAvailableActions()

	case "send":
		c.sendRawMessage(args)

	case "clear", "cls":
		c.ui.ClearCommandOutput()
		c.refreshDisplay()

	default:
		// Check if it's a numbered action selection
		if len(cmd) > 0 && cmd[0] >= '0' && cmd[0] <= '9' {
			c.selectAction(cmd)
		} else {
			c.displayCommandResult(cmd, fmt.Sprintf("❓ Unknown command: %s (type 'help' for available commands)", cmd))
		}
	}

	return false
}

func (c *CLIClient) showHelp() {
	helpText := `📖 Available Commands:
  help, h          - Show this help message
  quit, exit, q    - Exit the CLI
  status, s        - Show connection status
  caj <name>       - Create and join game with player name
  join <id> <name> - Join existing game by ID with player name
  games            - List available games
  players          - List players in current game
  actions          - Show numbered list of available actions
  0-9              - Select action by number (0 = skip)
  send <type>      - Send raw WebSocket message
  clear, cls       - Clear screen`

	c.displayCommandResult("help", helpText)
}

func (c *CLIClient) showStatus() {
	statusText := fmt.Sprintf(`🔗 Connection Status:
  Player ID: %s
  Game ID: %s
  Connected: %t`, c.playerID, c.gameID, c.conn != nil)

	c.displayCommandResult("status", statusText)
}

func (c *CLIClient) createAndJoinGame(args []string) {
	if len(args) == 0 {
		c.displayCommandResult("caj", "❌ Usage: caj <playerName>\nExample: caj \"Alice\"")
		return
	}

	playerName := strings.Join(args, " ")

	// Step 1: Create game via HTTP API
	gameID, err := c.createGameViaHTTP()
	if err != nil {
		c.displayCommandResult("caj", fmt.Sprintf("❌ Failed to create game: %v", err))
		return
	}

	// Step 2: Join the created game via WebSocket
	err = c.joinGameViaWebSocket(gameID, playerName)
	if err != nil {
		c.displayCommandResult("caj", fmt.Sprintf("❌ Failed to join game: %v", err))
		return
	}

	// Set gameID locally
	c.gameID = gameID
	
	// Give a brief moment for the server to respond with game state
	time.Sleep(100 * time.Millisecond)
	
	result := fmt.Sprintf("✅ Game created and joined successfully!\n🎮 Game ID: %s\n🎉 Ready to play as '%s'!", gameID[:8]+"...", playerName)
	c.displayCommandResult("caj "+playerName, result)
}

func (c *CLIClient) joinExistingGame(args []string) {
	if len(args) < 2 {
		c.displayCommandResult("join", "❌ Usage: join <gameID> <playerName>\nExample: join abc123 \"Alice\"")
		return
	}

	gameID := args[0]
	playerName := strings.Join(args[1:], " ")

	// Join the game via WebSocket
	err := c.joinGameViaWebSocket(gameID, playerName)
	if err != nil {
		c.displayCommandResult("join", fmt.Sprintf("❌ Failed to join game: %v", err))
		return
	}

	// Set gameID locally
	c.gameID = gameID
	result := fmt.Sprintf("✅ Successfully joined game %s\n🎉 Ready to play as '%s'!", gameID[:min(8, len(gameID))], playerName)
	c.displayCommandResult("join "+gameID+" "+playerName, result)
}

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// createGameViaHTTP creates a game using the HTTP API and returns the game ID
func (c *CLIClient) createGameViaHTTP() (string, error) {
	// Create request payload
	requestBody := dto.CreateGameRequest{
		MaxPlayers: 4,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP POST request
	resp, err := http.Post(httpAPIBase+"/games", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var response dto.CreateGameResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Game.ID, nil
}

// joinGameViaWebSocket joins an existing game via WebSocket
func (c *CLIClient) joinGameViaWebSocket(gameID, playerName string) error {
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayerConnect,
		GameID: gameID,
		Payload: dto.PlayerConnectPayload{
			PlayerName: playerName,
			GameID:     gameID,
		},
	}

	return c.conn.WriteJSON(message)
}

func (c *CLIClient) listGames() {
	fmt.Println("📋 Games: This would list available games (not implemented in backend yet)")
}

func (c *CLIClient) listPlayers() {
	if c.gameID == "" {
		fmt.Println("❌ Not connected to any game")
		return
	}

	fmt.Printf("👥 Players in game %s:\n", c.gameID)
	fmt.Printf("   • %s (you)\n", fmt.Sprintf("CLI-Player-%s", c.playerID[4:]))
	fmt.Println("\n💡 Other players will appear here when they join the game")
}

func (c *CLIClient) showAvailableActions() {
	var actionsText string


	// Check if we're in lobby and current player is host
	if c.gameState != nil && c.gameState.GameStatus == model.GameStatusLobby {
		if c.gameState.Player != nil && c.gameState.Player.ID == c.gameState.HostPlayerID {
			// Host in lobby - show start game action
			actionsText = `🎯 Lobby Actions (Host):
  0. Start Game (transition from lobby to active game)
  
💡 As the host, you can start the game when ready!`
		} else {
			// Non-host in lobby
			actionsText = `🎯 Lobby Actions:
  Waiting for host to start the game...
  
💡 The host will start the game when ready.`
		}
	} else {
		// Active game actions
		actionsText = `🎯 Available Actions:
  0. Skip Action
  1. Raise Temperature (8 heat → +1°C)
  2. Raise Oxygen (14 megacredits → +1%)
  3. Place Ocean (18 megacredits → ocean tile)
  4. Buy Standard Project
  5. Play Card from Hand
  6. Use Corporation Action
  7. Use Card Action
  8. Trade with Colonies
  9. End Turn

💡 Enter the action number (0-9) to perform the action`
	}

	c.displayCommandResult("actions", actionsText)
}

func (c *CLIClient) selectAction(actionNum string) {
	if c.gameID == "" {
		c.displayCommandResult(actionNum, "❌ Not connected to any game. Use 'caj <name>' or 'join <id> <name>' first.")
		return
	}

	// Handle lobby actions differently from active game actions
	if c.gameState != nil && c.gameState.GameStatus == model.GameStatusLobby {
		switch actionNum {
		case "0":
			// Start game action in lobby (host only)
			if c.gameState.Player != nil && c.gameState.Player.ID == c.gameState.HostPlayerID {
				c.startGame()
			} else {
				c.displayCommandResult(actionNum, "❌ Only the host can start the game.")
			}
		default:
			c.displayCommandResult(actionNum, "❌ Invalid action for lobby. Only action 0 (Start Game) is available for the host.")
		}
		return
	}

	// Active game actions
	switch actionNum {
	case "0":
		c.skipAction()
	case "1":
		c.raiseTemperature()
	case "2":
		c.raiseOxygen()
	case "3":
		c.placeOcean()
	case "4":
		c.buyStandardProject()
	case "5":
		c.playCard()
	case "6":
		c.useCorporationAction()
	case "7":
		c.useCardAction()
	case "8":
		c.tradeWithColonies()
	case "9":
		c.endTurn()
	default:
		c.displayCommandResult(actionNum, fmt.Sprintf("❌ Invalid action number: %s (use 0-9)", actionNum))
	}
}

func (c *CLIClient) startGame() {
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type": "start-game",
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("0", fmt.Sprintf("❌ Failed to start game: %v", err))
		return
	}

	c.displayCommandResult("0", "🚀 Starting game...")
}

func (c *CLIClient) skipAction() {
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type": "skip-action",
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("0", fmt.Sprintf("❌ Failed to skip action: %v", err))
		return
	}

	c.displayCommandResult("0", "✅ Action skipped")
}

func (c *CLIClient) raiseTemperature() {
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type": "launch-asteroid",
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("1", fmt.Sprintf("❌ Failed to raise temperature: %v", err))
		return
	}

	c.displayCommandResult("1", "🌡️ Asteroid launched to raise temperature")
}

func (c *CLIClient) raiseOxygen() {
	// Placeholder - not implemented in backend yet
}

func (c *CLIClient) placeOcean() {
	// Placeholder - not implemented in backend yet
}

func (c *CLIClient) buyStandardProject() {
	if c.gameID == "" {
		fmt.Println("❌ Not connected to any game. Use 'connect <game>' first.")
		return
	}

	// Show available standard projects
	fmt.Println("🏗️ Standard Projects Available:")
	fmt.Println("  1. Sell Patents - 1 M€ per card")
	fmt.Println("  2. Power Plant - 11 M€ (Energy production +1)")
	fmt.Println("  3. Asteroid - 14 M€ (Temperature +1, TR +1)")
	fmt.Println("  4. Aquifer - 18 M€ (Ocean +1, TR +1)")
	fmt.Println("  5. Greenery - 23 M€ (Oxygen +1, TR +1)")
	fmt.Println("  6. City - 25 M€ (Credit production +1)")
	fmt.Println("  0. Cancel")
	fmt.Print("Select project (0-6): ")

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "0":
		fmt.Println("❌ Cancelled standard project selection")
		return
	case "1":
		c.executeStandardProject("SELL_PATENTS", "Sell Patents")
	case "2":
		c.executeStandardProject("POWER_PLANT", "Power Plant")
	case "3":
		c.executeStandardProject("ASTEROID", "Asteroid")
	case "4":
		c.executeStandardProject("AQUIFER", "Aquifer")
	case "5":
		c.executeStandardProject("GREENERY", "Greenery")
	case "6":
		c.executeStandardProject("CITY", "City")
	default:
		fmt.Printf("❌ Invalid choice: %s\n", choice)
	}
}

func (c *CLIClient) playCard() {
	// Placeholder - not implemented in backend yet
}

func (c *CLIClient) useCorporationAction() {
	// Placeholder - not implemented in backend yet
}

func (c *CLIClient) useCardAction() {
	// Placeholder - not implemented in backend yet
}

func (c *CLIClient) tradeWithColonies() {
	// Placeholder - not implemented in backend yet
}

func (c *CLIClient) endTurn() {
	// Placeholder - not implemented in backend yet
}

// executeStandardProject executes a standard project and shows status
func (c *CLIClient) executeStandardProject(projectType, projectName string) {
	fmt.Printf("🔨 Executing: %s", projectName)

	var message dto.WebSocketMessage

	// Build the message based on project type
	switch projectType {
	case "SELL_PATENTS":
		// For sell patents, ask how many cards to sell
		fmt.Print("\nHow many cards to sell? ")
		var cardCount int
		fmt.Scanln(&cardCount)
		if cardCount <= 0 {
			fmt.Printf(" ❌ Failed\n")
			return
		}

		message = dto.WebSocketMessage{
			Type:   dto.MessageTypePlayAction,
			GameID: c.gameID,
			Payload: dto.PlayActionPayload{
				ActionRequest: map[string]interface{}{
					"type":      "sell-patents",
					"cardCount": cardCount,
				},
			},
		}

	case "POWER_PLANT":
		message = dto.WebSocketMessage{
			Type:   dto.MessageTypePlayAction,
			GameID: c.gameID,
			Payload: dto.PlayActionPayload{
				ActionRequest: map[string]interface{}{
					"type": "build-power-plant",
				},
			},
		}

	case "ASTEROID":
		message = dto.WebSocketMessage{
			Type:   dto.MessageTypePlayAction,
			GameID: c.gameID,
			Payload: dto.PlayActionPayload{
				ActionRequest: map[string]interface{}{
					"type": "launch-asteroid",
				},
			},
		}

	case "AQUIFER":
		// For hex placement projects, ask for position
		fmt.Print("\nEnter hex position (q r s): ")
		var q, r, s int
		fmt.Scanf("%d %d %d", &q, &r, &s)
		if q+r+s != 0 {
			fmt.Printf(" ❌ Failed - Invalid hex position\n")
			return
		}

		message = dto.WebSocketMessage{
			Type:   dto.MessageTypePlayAction,
			GameID: c.gameID,
			Payload: dto.PlayActionPayload{
				ActionRequest: map[string]interface{}{
					"type": "build-aquifer",
					"hexPosition": map[string]interface{}{
						"q": q,
						"r": r,
						"s": s,
					},
				},
			},
		}

	case "GREENERY":
		// For hex placement projects, ask for position
		fmt.Print("\nEnter hex position (q r s): ")
		var q, r, s int
		fmt.Scanf("%d %d %d", &q, &r, &s)
		if q+r+s != 0 {
			fmt.Printf(" ❌ Failed - Invalid hex position\n")
			return
		}

		message = dto.WebSocketMessage{
			Type:   dto.MessageTypePlayAction,
			GameID: c.gameID,
			Payload: dto.PlayActionPayload{
				ActionRequest: map[string]interface{}{
					"type": "plant-greenery",
					"hexPosition": map[string]interface{}{
						"q": q,
						"r": r,
						"s": s,
					},
				},
			},
		}

	case "CITY":
		// For hex placement projects, ask for position
		fmt.Print("\nEnter hex position (q r s): ")
		var q, r, s int
		fmt.Scanf("%d %d %d", &q, &r, &s)
		if q+r+s != 0 {
			fmt.Printf(" ❌ Failed - Invalid hex position\n")
			return
		}

		message = dto.WebSocketMessage{
			Type:   dto.MessageTypePlayAction,
			GameID: c.gameID,
			Payload: dto.PlayActionPayload{
				ActionRequest: map[string]interface{}{
					"type": "build-city",
					"hexPosition": map[string]interface{}{
						"q": q,
						"r": r,
						"s": s,
					},
				},
			},
		}
	}

	// Send the message
	if err := c.conn.WriteJSON(message); err != nil {
		fmt.Printf(" ❌ Failed - %v\n", err)
		return
	}

	fmt.Printf(" ✅ Success\n")
}

func (c *CLIClient) sendRawMessage(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Usage: send <message_type> [payload_json]")
		fmt.Println("Example: send player-connect")
		return
	}

	messageType := dto.MessageType(args[0])
	var payload interface{}

	if len(args) > 1 {
		payloadStr := strings.Join(args[1:], " ")
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			fmt.Printf("❌ Invalid JSON payload: %v\n", err)
			return
		}
	}

	message := dto.WebSocketMessage{
		Type:    messageType,
		GameID:  c.gameID,
		Payload: payload,
	}

	if err := c.conn.WriteJSON(message); err != nil {
		fmt.Printf("❌ Failed to send message: %v\n", err)
		return
	}

	fmt.Printf("📤 Sent message: %s\n", messageType)
}

// updateGameStateFromMessage updates the game state from a WebSocket message
func (c *CLIClient) updateGameStateFromMessage(message dto.WebSocketMessage) {
	if payload, ok := message.Payload.(map[string]interface{}); ok {
		if gameData, ok := payload["game"].(map[string]interface{}); ok {
			c.parseGameData(gameData)
		}
	}
}

// parseGameData parses game data from the message payload
func (c *CLIClient) parseGameData(gameData map[string]interface{}) {
	// Update generation
	if generation, ok := gameData["generation"].(float64); ok {
		c.gameState.Generation = int(generation)
	}

	// Update current phase using backend types
	if phase, ok := gameData["currentPhase"].(string); ok {
		c.gameState.CurrentPhase = model.GamePhase(phase)
	}

	// Update game status using backend types
	if status, ok := gameData["status"].(string); ok {
		c.gameState.GameStatus = model.GameStatus(status)
	}

	// Update host player ID
	if hostPlayerID, ok := gameData["hostPlayerId"].(string); ok {
		c.gameState.HostPlayerID = hostPlayerID
	}

	// Update total players
	if players, ok := gameData["players"].([]interface{}); ok {
		c.gameState.TotalPlayers = len(players)

		// Find current player
		for _, playerInterface := range players {
			if playerMap, ok := playerInterface.(map[string]interface{}); ok {
				if playerID, ok := playerMap["id"].(string); ok && playerID == c.playerID {
					c.parsePlayerData(playerMap)
					break
				}
			}
		}
	}

	// Update global parameters
	if globalParams, ok := gameData["globalParameters"].(map[string]interface{}); ok {
		if c.gameState.GlobalParameters == nil {
			c.gameState.GlobalParameters = &GlobalParams{}
		}

		if temp, ok := globalParams["temperature"].(float64); ok {
			c.gameState.GlobalParameters.Temperature = int(temp)
		}
		if oxygen, ok := globalParams["oxygen"].(float64); ok {
			c.gameState.GlobalParameters.Oxygen = int(oxygen)
		}
		if oceans, ok := globalParams["oceans"].(float64); ok {
			c.gameState.GlobalParameters.Oceans = int(oceans)
		}
	}
}

// parsePlayerData parses player data from the message
func (c *CLIClient) parsePlayerData(playerData map[string]interface{}) {
	if c.gameState.Player == nil {
		c.gameState.Player = &model.Player{}
	}

	// Parse basic player info
	if id, ok := playerData["id"].(string); ok {
		c.gameState.Player.ID = id
	}
	if name, ok := playerData["name"].(string); ok {
		c.gameState.Player.Name = name
	}
	if corp, ok := playerData["corporation"].(string); ok {
		c.gameState.Player.Corporation = corp
	}
	if tr, ok := playerData["terraformRating"].(float64); ok {
		c.gameState.Player.TerraformRating = int(tr)
	}
	if active, ok := playerData["isActive"].(bool); ok {
		c.gameState.Player.IsActive = active
	}

	// Parse resources
	if resources, ok := playerData["resources"].(map[string]interface{}); ok {
		c.parseResources(resources, &c.gameState.Player.Resources)
	}

	// Parse production
	if production, ok := playerData["production"].(map[string]interface{}); ok {
		c.parseProduction(production, &c.gameState.Player.Production)
	}

	// Parse cards
	if cards, ok := playerData["cards"].([]interface{}); ok {
		c.gameState.Player.Cards = make([]string, len(cards))
		for i, card := range cards {
			if cardStr, ok := card.(string); ok {
				c.gameState.Player.Cards[i] = cardStr
			}
		}
	}

	// Parse played cards
	if playedCards, ok := playerData["playedCards"].([]interface{}); ok {
		c.gameState.Player.PlayedCards = make([]string, len(playedCards))
		for i, card := range playedCards {
			if cardStr, ok := card.(string); ok {
				c.gameState.Player.PlayedCards[i] = cardStr
			}
		}
	}
}

// parseResources parses resources from the message
func (c *CLIClient) parseResources(resourcesData map[string]interface{}, resources *model.Resources) {
	if credits, ok := resourcesData["credits"].(float64); ok {
		resources.Credits = int(credits)
	}
	if steel, ok := resourcesData["steel"].(float64); ok {
		resources.Steel = int(steel)
	}
	if titanium, ok := resourcesData["titanium"].(float64); ok {
		resources.Titanium = int(titanium)
	}
	if plants, ok := resourcesData["plants"].(float64); ok {
		resources.Plants = int(plants)
	}
	if energy, ok := resourcesData["energy"].(float64); ok {
		resources.Energy = int(energy)
	}
	if heat, ok := resourcesData["heat"].(float64); ok {
		resources.Heat = int(heat)
	}
}

// parseProduction parses production from the message
func (c *CLIClient) parseProduction(productionData map[string]interface{}, production *model.Production) {
	if credits, ok := productionData["credits"].(float64); ok {
		production.Credits = int(credits)
	}
	if steel, ok := productionData["steel"].(float64); ok {
		production.Steel = int(steel)
	}
	if titanium, ok := productionData["titanium"].(float64); ok {
		production.Titanium = int(titanium)
	}
	if plants, ok := productionData["plants"].(float64); ok {
		production.Plants = int(plants)
	}
	if energy, ok := productionData["energy"].(float64); ok {
		production.Energy = int(energy)
	}
	if heat, ok := productionData["heat"].(float64); ok {
		production.Heat = int(heat)
	}
}

// refreshDisplay refreshes the complete display
func (c *CLIClient) refreshDisplay() {
	c.ui.UpdateGameState(c.gameState)
	fmt.Print("\033[2J\033[H") // Clear screen and move cursor to home
	fmt.Println(c.ui.RenderFullDisplay())
	fmt.Println() // Add space before prompt
}

// displayCommandResult displays a command and its result, then refreshes
func (c *CLIClient) displayCommandResult(command, result string) {
	c.ui.SetLastCommand(command, result)
	c.refreshDisplay()
}
