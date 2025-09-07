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
	"strconv"
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

	// Track previous game status to detect transitions
	var prevStatus model.GameStatus
	if c.gameState != nil {
		prevStatus = c.gameState.GameStatus
	}

	switch message.Type {
	case dto.MessageTypeGameUpdated:
		c.updateGameStateFromMessage(message)
		
		// Check if game just started (lobby → active)
		if prevStatus == model.GameStatusLobby && c.gameState.GameStatus == model.GameStatusActive {
			c.displayCommandResult("game-started", "🚀 Game Started! Ready for action!\n\n💡 Available actions are displayed above.\n📖 Type 'help' for command details or use numbered actions (0-7).")
		}
		
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
		
		// Ensure connected state is set even if we don't get full game data
		if !c.gameState.IsConnected && c.gameID != "" {
			c.gameState.IsConnected = true
			c.ui.UpdateGameState(c.gameState)
			c.refreshDisplay()
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
		c.showContextualHelp()

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

	case "actions":
		c.displayCommandResult("actions", "💡 Available actions are always displayed above!")

	case "clear", "cls":
		c.ui.ClearCommandOutput()
		c.refreshDisplay()

	// Commands that require being connected to a game
	case "games":
		c.requiresConnection(cmd, func() { c.listGames() })

	case "players":
		c.requiresConnection(cmd, func() { c.listPlayers() })

	case "cards", "hand":
		c.requiresConnection(cmd, func() { c.showCardsInHand() })

	case "play":
		c.requiresConnection(cmd, func() { c.playCardFromHand(args) })

	case "buy":
		c.requiresConnection(cmd, func() { c.buyCards(args) })

	case "convert":
		c.requiresConnection(cmd, func() { c.convertResources(args) })

	case "milestones", "awards":
		c.requiresConnection(cmd, func() { c.showMilestonesAndAwards() })

	case "claim":
		c.requiresConnection(cmd, func() { c.claimMilestoneOrAward(args) })

	case "overview", "summary":
		c.requiresConnection(cmd, func() { c.showGameOverview() })

	case "send":
		c.requiresConnection(cmd, func() { c.sendRawMessage(args) })

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

// requiresConnection checks if connected to a game before executing the command
func (c *CLIClient) requiresConnection(cmd string, fn func()) {
	if c.gameState == nil || !c.gameState.IsConnected || c.gameID == "" {
		c.displayCommandResult(cmd, "❌ This command requires being connected to a game.\n💡 Use 'caj <name>' or 'join <id> <name>' to connect first.")
		return
	}
	fn()
}

// showContextualHelp shows help based on current connection state
func (c *CLIClient) showContextualHelp() {
	var helpText string

	if c.gameState == nil || !c.gameState.IsConnected || c.gameID == "" {
		// Disconnected - show connection commands only
		helpText = `📖 Available Commands (Disconnected):
  help, h          - Show this help message
  quit, exit, q    - Exit the CLI
  status, s        - Show connection status
  caj <name>       - Create and join new game
  join <id> <name> - Join existing game by ID
  clear, cls       - Clear screen
  
💡 Available actions are always shown above. Connect to start playing!`
	} else if c.gameState.GameStatus == model.GameStatusLobby {
		// In lobby
		helpText = `📖 Available Commands (Lobby):
  help, h          - Show this help message
  quit, exit, q    - Exit the CLI
  status, s        - Show connection status
  players          - List players in current game
  overview         - Show detailed game overview
  clear, cls       - Clear screen
  0                - Start game (host only)
  
💡 Available actions are always shown above!`
	} else {
		// Active game - show all commands
		helpText = `📖 Available Commands (Active Game):
  help, h          - Show this help message
  quit, exit, q    - Exit the CLI
  status, s        - Show connection status
  cards, hand      - Show cards in your hand
  play <card_id>   - Play a card from your hand
  buy [amount]     - Buy cards from research deck
  convert <type>   - Convert resources (heat, plants, etc.)
  milestones       - Show available milestones and awards
  claim <name>     - Claim a milestone or award
  overview         - Show detailed game overview
  players          - List players in current game
  0-7              - Select action by number
  clear, cls       - Clear screen
  
💡 Available actions are always shown above!`
	}

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

	// Set gameID locally and mark as connected
	c.gameID = gameID
	c.gameState.GameID = gameID
	c.gameState.IsConnected = true

	// Update UI with current state
	c.ui.UpdateGameState(c.gameState)

	// Give a brief moment for the server to respond with game state
	time.Sleep(100 * time.Millisecond)

	result := fmt.Sprintf("✅ Game created and joined successfully!\n🎮 Game ID: %s\n🎉 Ready to play as '%s'!\n\n💡 Available commands:\n  help - Show help\n  0 - Start game (host only)\n  players - List players\n  status - Show game status", gameID[:8]+"...", playerName)
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

	// Set gameID locally and mark as connected
	c.gameID = gameID
	c.gameState.GameID = gameID
	c.gameState.IsConnected = true

	// Update UI with current state
	c.ui.UpdateGameState(c.gameState)

	result := fmt.Sprintf("✅ Successfully joined game %s\n🎉 Ready to play as '%s'!\n\n💡 Available commands:\n  help - Show help\n  players - List players\n  status - Show game status\n  overview - Show detailed game info", gameID[:min(8, len(gameID))], playerName)
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


func (c *CLIClient) selectAction(actionNum string) {
	// Check connection state first
	if c.gameState == nil || !c.gameState.IsConnected || c.gameID == "" {
		c.displayCommandResult(actionNum, "❌ Not connected to any game. Use 'caj <name>' or 'join <id> <name>' first.")
		return
	}

	// Handle lobby actions
	if c.gameState.GameStatus == model.GameStatusLobby {
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

	// Active game actions (contextual based on game state)
	switch actionNum {
	case "0":
		// End turn / Skip action
		c.endTurn()
	case "1":
		// Convert heat to temperature
		c.convertResources([]string{"heat"})
	case "2":
		// Convert plants to greenery
		c.convertResources([]string{"plants"})
	case "3":
		// Asteroid standard project
		c.executeStandardProject("ASTEROID", "Asteroid")
	case "4":
		// Ocean standard project
		c.executeStandardProject("AQUIFER", "Ocean")
	case "5":
		// Play card from hand
		c.displayCommandResult(actionNum, "💡 Use 'play <card_number>' command to play specific cards.\n    Use 'cards' to view your hand first.")
	case "6":
		// Buy cards
		c.buyCards([]string{"1"}) // Default to 1 card
	case "7":
		// Use corporation action
		c.useCorporationAction()
	default:
		c.displayCommandResult(actionNum, fmt.Sprintf("❌ Invalid action number: %s (use 0-7)", actionNum))
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

	c.displayCommandResult("0", "🚀 Starting game...\n\n⏳ Transitioning from lobby to active game.\n💡 Available actions will be displayed once the game starts.")
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
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type": "standard-project-oxygen",
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("2", fmt.Sprintf("❌ Failed to raise oxygen: %v", err))
		return
	}

	c.displayCommandResult("2", "💨 Attempting to raise oxygen (+1%)")
}

func (c *CLIClient) placeOcean() {
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type": "standard-project-ocean",
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("3", fmt.Sprintf("❌ Failed to place ocean: %v", err))
		return
	}

	c.displayCommandResult("3", "🌊 Attempting to place ocean tile")
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
	c.displayCommandResult("5", "💡 Use 'play <card_number>' command to play specific cards from your hand.\n    Use 'cards' to view your hand first.")
}

func (c *CLIClient) useCorporationAction() {
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type": "corporation-action",
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("6", fmt.Sprintf("❌ Failed to use corporation action: %v", err))
		return
	}

	c.displayCommandResult("6", "🏢 Attempting to use corporation action")
}

func (c *CLIClient) useCardAction() {
	c.displayCommandResult("7", "🃏 Card actions not yet implemented - this would show available card abilities")
}

func (c *CLIClient) tradeWithColonies() {
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type": "trade-colonies",
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("8", fmt.Sprintf("❌ Failed to trade with colonies: %v", err))
		return
	}

	c.displayCommandResult("8", "🚀 Attempting to trade with colonies")
}

func (c *CLIClient) endTurn() {
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type": "end-turn",
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("9", fmt.Sprintf("❌ Failed to end turn: %v", err))
		return
	}

	c.displayCommandResult("9", "⏭️ Turn ended")
}

// showCardsInHand displays all cards currently in the player's hand
func (c *CLIClient) showCardsInHand() {
	if c.gameID == "" {
		c.displayCommandResult("cards", "❌ Not connected to any game. Use 'caj <name>' or 'join <id> <name>' first.")
		return
	}

	if c.gameState == nil || c.gameState.Player == nil {
		c.displayCommandResult("cards", "❌ No player data available.")
		return
	}

	cards := c.gameState.Player.Cards
	if len(cards) == 0 {
		c.displayCommandResult("cards", "🃏 Your hand is empty.\n💡 Use 'buy' command to purchase cards from the research deck.")
		return
	}

	var cardDisplay strings.Builder
	cardDisplay.WriteString(fmt.Sprintf("🃏 Cards in Hand (%d):\n\n", len(cards)))

	for i, cardID := range cards {
		// Enhanced card display with formatting
		cardDisplay.WriteString(fmt.Sprintf("  %d. %-20s 🏷️  %s\n", i+1, cardID, c.getCardTypeIcon(cardID)))
	}

	// Show resource summary for context
	if c.gameState.Player != nil {
		resources := c.gameState.Player.Resources
		cardDisplay.WriteString(fmt.Sprintf("\n💰 Available Resources: %d MC, %d Steel, %d Titanium, %d Plants, %d Energy, %d Heat\n",
			resources.Credits, resources.Steel, resources.Titanium, resources.Plants, resources.Energy, resources.Heat))
	}

	cardDisplay.WriteString("\n💡 Use 'play <card_number>' to play a card")
	c.displayCommandResult("cards", cardDisplay.String())
}

// getCardTypeIcon returns an appropriate icon for the card type (basic categorization)
func (c *CLIClient) getCardTypeIcon(cardID string) string {
	// Basic card type detection based on card ID/name patterns
	cardLower := strings.ToLower(cardID)
	switch {
	case strings.Contains(cardLower, "plant") || strings.Contains(cardLower, "greenery"):
		return "🌱 Plant"
	case strings.Contains(cardLower, "power") || strings.Contains(cardLower, "energy"):
		return "⚡ Power"
	case strings.Contains(cardLower, "space") || strings.Contains(cardLower, "asteroid"):
		return "🚀 Space"
	case strings.Contains(cardLower, "water") || strings.Contains(cardLower, "ocean"):
		return "🌊 Water"
	case strings.Contains(cardLower, "heat") || strings.Contains(cardLower, "temperature"):
		return "🌡️ Heat"
	case strings.Contains(cardLower, "building") || strings.Contains(cardLower, "city"):
		return "🏗️ Building"
	case strings.Contains(cardLower, "science") || strings.Contains(cardLower, "research"):
		return "🔬 Science"
	default:
		return "🃏 Card"
	}
}

// playCardFromHand attempts to play a card from the player's hand
func (c *CLIClient) playCardFromHand(args []string) {
	if c.gameID == "" {
		c.displayCommandResult("play", "❌ Not connected to any game. Use 'caj <name>' or 'join <id> <name>' first.")
		return
	}

	if len(args) == 0 {
		c.displayCommandResult("play", "❌ Usage: play <card_number>\nExample: play 1")
		return
	}

	cardID := args[0]

	// If it's a number, try to get card from hand
	if cardNumber, err := strconv.Atoi(args[0]); err == nil {
		if c.gameState != nil && c.gameState.Player != nil && len(c.gameState.Player.Cards) > 0 {
			if cardNumber >= 1 && cardNumber <= len(c.gameState.Player.Cards) {
				cardID = c.gameState.Player.Cards[cardNumber-1]
			}
		}
	}

	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type":   "play-card",
				"cardId": cardID,
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("play", fmt.Sprintf("❌ Failed to play card: %v", err))
		return
	}

	c.displayCommandResult("play", fmt.Sprintf("🃏 Attempting to play card: %s", cardID))
}

// buyCards purchases cards from the research deck
func (c *CLIClient) buyCards(args []string) {
	if c.gameID == "" {
		c.displayCommandResult("buy", "❌ Not connected to any game. Use 'caj <name>' or 'join <id> <name>' first.")
		return
	}

	// Default to buying 1 card if no amount specified
	cardCount := 1
	if len(args) > 0 {
		var err error
		cardCount, err = strconv.Atoi(args[0])
		if err != nil {
			c.displayCommandResult("buy", "❌ Invalid card count. Must be a number.")
			return
		}
	}

	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type":      "buy-cards",
				"cardCount": cardCount,
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("buy", fmt.Sprintf("❌ Failed to buy cards: %v", err))
		return
	}

	cost := cardCount * 3 // 3 MC per card
	c.displayCommandResult("buy", fmt.Sprintf("💳 Attempting to buy %d card(s) for %d MC", cardCount, cost))
}

// convertResources handles resource conversion actions
func (c *CLIClient) convertResources(args []string) {
	if c.gameID == "" {
		c.displayCommandResult("convert", "❌ Not connected to any game. Use 'caj <name>' or 'join <id> <name>' first.")
		return
	}

	if len(args) == 0 {
		helpText := `💱 Resource Conversion Options:
  convert heat      - Convert 8 heat to raise temperature 1 step
  convert plants    - Convert 8 plants to place greenery tile
  convert energy    - Convert energy to heat at end of generation
  
💡 Usage: convert <resource_type>`
		c.displayCommandResult("convert", helpText)
		return
	}

	resourceType := strings.ToLower(args[0])
	var actionType string
	var description string

	switch resourceType {
	case "heat":
		actionType = "convert-heat-temperature"
		description = "🌡️ Converting 8 heat to raise temperature"
	case "plants":
		actionType = "convert-plants-greenery"
		description = "🌱 Converting 8 plants to place greenery"
	case "energy":
		actionType = "convert-energy-heat"
		description = "⚡ Converting energy to heat"
	default:
		c.displayCommandResult("convert", fmt.Sprintf("❌ Unknown conversion type: %s\nUse: heat, plants, or energy", resourceType))
		return
	}

	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type": actionType,
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("convert", fmt.Sprintf("❌ Failed to convert resources: %v", err))
		return
	}

	c.displayCommandResult("convert", description)
}

// showMilestonesAndAwards displays available milestones and awards
func (c *CLIClient) showMilestonesAndAwards() {
	if c.gameID == "" {
		c.displayCommandResult("milestones", "❌ Not connected to any game. Use 'caj <name>' or 'join <id> <name>' first.")
		return
	}

	milestonesText := `🏆 Milestones (5 VP each, 8 MC to claim):
  • Terraformer - 35 TR or more
  • Mayor - 3+ cities
  • Gardener - 3+ greenery tiles
  • Builder - 8+ building tags
  • Planner - 16+ cards in hand

🥇 Awards (5 VP for 1st, 2 VP for 2nd):
  • Landlord - Most tiles on Mars
  • Banker - Highest MC production
  • Scientist - Most science tags
  • Thermalist - Most heat resource
  • Miner - Most steel and titanium

💡 Use 'claim <milestone/award_name>' to claim them
💡 Backend handles validation and timing`

	c.displayCommandResult("milestones", milestonesText)
}

// claimMilestoneOrAward attempts to claim a milestone or award
func (c *CLIClient) claimMilestoneOrAward(args []string) {
	if c.gameID == "" {
		c.displayCommandResult("claim", "❌ Not connected to any game. Use 'caj <name>' or 'join <id> <name>' first.")
		return
	}

	if len(args) == 0 {
		c.displayCommandResult("claim", "❌ Usage: claim <milestone/award_name>\nExample: claim terraformer")
		return
	}

	name := strings.ToLower(strings.Join(args, " "))

	// Normalize common milestone/award names
	switch name {
	case "terraformer", "mayor", "gardener", "builder", "planner":
		// Valid milestone names
	case "landlord", "banker", "scientist", "thermalist", "miner":
		// Valid award names
	default:
		c.displayCommandResult("claim", fmt.Sprintf("❌ Unknown milestone/award: %s\nUse 'milestones' to see available options", name))
		return
	}

	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type": "claim-milestone-award",
				"name": name,
			},
		},
	}

	if err := c.conn.WriteJSON(message); err != nil {
		c.displayCommandResult("claim", fmt.Sprintf("❌ Failed to claim %s: %v", name, err))
		return
	}

	c.displayCommandResult("claim", fmt.Sprintf("🏆 Attempting to claim: %s", name))
}

// showGameOverview displays a comprehensive game status overview
func (c *CLIClient) showGameOverview() {
	if c.gameID == "" {
		c.displayCommandResult("overview", "❌ Not connected to any game. Use 'caj <name>' or 'join <id> <name>' first.")
		return
	}

	if c.gameState == nil {
		c.displayCommandResult("overview", "❌ No game state available.")
		return
	}

	var overview strings.Builder
	overview.WriteString("📊 Game Overview\n\n")

	// Game Status
	overview.WriteString(fmt.Sprintf("🎮 Game ID: %s\n", c.gameID[:min(8, len(c.gameID))]+"..."))
	overview.WriteString(fmt.Sprintf("📊 Status: %s\n", c.gameState.GameStatus))
	overview.WriteString(fmt.Sprintf("🏃 Phase: %s\n", c.gameState.CurrentPhase))
	overview.WriteString(fmt.Sprintf("🎯 Generation: %d\n", c.gameState.Generation))
	overview.WriteString(fmt.Sprintf("👥 Players: %d\n", c.gameState.TotalPlayers))

	// Global Parameters
	if c.gameState.GlobalParameters != nil {
		overview.WriteString(fmt.Sprintf("\n🌍 Mars Status:\n"))
		overview.WriteString(fmt.Sprintf("  🌡️  Temperature: %d°C\n", c.gameState.GlobalParameters.Temperature))
		overview.WriteString(fmt.Sprintf("  💨 Oxygen: %d%%\n", c.gameState.GlobalParameters.Oxygen))
		overview.WriteString(fmt.Sprintf("  🌊 Oceans: %d\n", c.gameState.GlobalParameters.Oceans))
	}

	// Player Status
	if c.gameState.Player != nil {
		player := c.gameState.Player
		overview.WriteString(fmt.Sprintf("\n👤 Your Status:\n"))
		overview.WriteString(fmt.Sprintf("  📛 Name: %s\n", player.Name))
		overview.WriteString(fmt.Sprintf("  🏢 Corporation: %s\n", player.Corporation))
		overview.WriteString(fmt.Sprintf("  🎯 TR: %d\n", player.TerraformRating))
		overview.WriteString(fmt.Sprintf("  🃏 Hand: %d cards\n", len(player.Cards)))
		overview.WriteString(fmt.Sprintf("  📝 Played: %d cards\n", len(player.PlayedCards)))

		// Resources summary
		r := player.Resources
		overview.WriteString(fmt.Sprintf("  💰 Resources: %d MC, %d Steel, %d Ti, %d Plants, %d Energy, %d Heat\n",
			r.Credits, r.Steel, r.Titanium, r.Plants, r.Energy, r.Heat))

		// Production summary
		p := player.Production
		overview.WriteString(fmt.Sprintf("  🏭 Production: %d MC, %d Steel, %d Ti, %d Plants, %d Energy, %d Heat\n",
			p.Credits, p.Steel, p.Titanium, p.Plants, p.Energy, p.Heat))
	}

	overview.WriteString(fmt.Sprintf("\n💡 Use 'actions' to see available moves"))

	c.displayCommandResult("overview", overview.String())
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
