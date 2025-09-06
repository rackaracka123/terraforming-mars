package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Default server address
	defaultServerAddr = "localhost:3001"

	// CLI tool metadata
	cliVersion = "1.0.0"
	cliName    = "Terraforming Mars CLI"
)

// GameState holds the current game state for display
type GameState struct {
	Player           *model.Player
	Generation       int
	CurrentPhase     string
	GameID           string
	IsConnected      bool
	TotalPlayers     int
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
		// Extract the actual player ID from the payload
		if payload, ok := message.Payload.(map[string]interface{}); ok {
			if playerID, ok := payload["playerId"].(string); ok {
				c.playerID = playerID // Update to the actual player ID from the game
			}
		}

	case dto.MessageTypeError:
		if payload, ok := message.Payload.(map[string]interface{}); ok {
			if msg, ok := payload["message"].(string); ok {
				// Only show errors on the current line, don't add to history
				fmt.Printf("\r%s\n", c.ui.RenderMessage("error", msg))
				fmt.Print(c.ui.RenderPrompt())
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

	case "join":
		c.joinGame(args)

	case "create":
		c.createGame(args)

	case "games":
		c.listGames()

	case "players":
		c.listPlayers()

	case "actions":
		c.showAvailableActions()

	case "send":
		c.sendRawMessage(args)

	case "clear", "cls":
		fmt.Print("\033[2J\033[H") // Clear screen

	default:
		// Check if it's a numbered action selection
		if len(cmd) > 0 && cmd[0] >= '0' && cmd[0] <= '9' {
			c.selectAction(cmd)
		} else {
			fmt.Printf("❓ Unknown command: %s (type 'help' for available commands)\n", cmd)
		}
	}

	return false
}

func (c *CLIClient) showHelp() {
	fmt.Println("📖 Available Commands:")
	fmt.Println("  help, h          - Show this help message")
	fmt.Println("  quit, exit, q    - Exit the CLI")
	fmt.Println("  status, s        - Show connection status")
	fmt.Println("  create <name>    - Create a new game with given name")
	fmt.Println("  join <gameId>    - Join an existing game by ID")
	fmt.Println("  games            - List available games")
	fmt.Println("  players          - List players in current game")
	fmt.Println("  actions          - Show numbered list of available actions")
	fmt.Println("  0-9              - Select action by number (0 = skip)")
	fmt.Println("  send <type>      - Send raw WebSocket message")
	fmt.Println("  clear, cls       - Clear screen")
	fmt.Println()
}

func (c *CLIClient) showStatus() {
	fmt.Printf("🔗 Connection Status:\n")
	fmt.Printf("  Player ID: %s\n", c.playerID)
	fmt.Printf("  Game ID: %s\n", c.gameID)
	fmt.Printf("  Connected: %t\n", c.conn != nil)
	fmt.Println()
}

func (c *CLIClient) joinGame(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Usage: join <gameId>")
		fmt.Println("Example: join f5d085d0-f9e2-47a0-b165-716c6022451b")
		return
	}

	gameID := args[0]

	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayerConnect,
		GameID: gameID,
		Payload: dto.PlayerConnectPayload{
			PlayerName: fmt.Sprintf("CLI-Player-%s", c.playerID[4:]),
			GameID:     gameID,
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		fmt.Printf("❌ Failed to join game: %v\n", err)
		return
	}

	// Set gameID locally since we're attempting to join this game
	c.gameID = gameID
	fmt.Printf("🎮 Joining game: %s\n", gameID)
}

func (c *CLIClient) createGame(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Usage: create <playerName>")
		fmt.Println("Example: create \"Alice\"")
		return
	}

	playerName := strings.Join(args, " ")

	// Generate a game ID automatically
	gameID := fmt.Sprintf("game-%d", time.Now().Unix())

	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayerConnect,
		GameID: gameID,
		Payload: dto.PlayerConnectPayload{
			PlayerName: playerName,
			GameID:     gameID,
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		fmt.Printf("❌ Failed to create game: %v\n", err)
		return
	}

	// Set gameID locally since we're creating and joining this game
	c.gameID = gameID
	fmt.Printf("🎮 Creating game and joining as '%s'\n", playerName)
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
	fmt.Println("🎯 Available Actions:")
	fmt.Println("  0. Skip Action")
	fmt.Println("  1. Raise Temperature (8 heat → +1°C)")
	fmt.Println("  2. Raise Oxygen (14 megacredits → +1%)")
	fmt.Println("  3. Place Ocean (18 megacredits → ocean tile)")
	fmt.Println("  4. Buy Standard Project")
	fmt.Println("  5. Play Card from Hand")
	fmt.Println("  6. Use Corporation Action")
	fmt.Println("  7. Use Card Action")
	fmt.Println("  8. Trade with Colonies")
	fmt.Println("  9. End Turn")
	fmt.Println()
	fmt.Println("💡 Enter the action number (0-9) to perform the action")
}

func (c *CLIClient) selectAction(actionNum string) {
	if c.gameID == "" {
		fmt.Println("❌ Not connected to any game. Use 'connect <game>' first.")
		return
	}

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
		fmt.Printf("❌ Invalid action number: %s (use 0-9)\n", actionNum)
	}
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
		fmt.Printf("\r❌ Failed to skip action: %v\n", err)
		fmt.Print(c.ui.RenderPrompt())
		return
	}
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
		fmt.Printf("\r❌ Failed to raise temperature: %v\n", err)
		fmt.Print(c.ui.RenderPrompt())
		return
	}
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

	// Update current phase
	if phase, ok := gameData["currentPhase"].(string); ok {
		c.gameState.CurrentPhase = phase
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

// refreshDisplay refreshes the status display
func (c *CLIClient) refreshDisplay() {
	c.ui.UpdateGameState(c.gameState)
	fmt.Print("\033[2J\033[H") // Clear screen
	fmt.Println(c.ui.RenderStatus())
	fmt.Println()
}
