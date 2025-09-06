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
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Default server address
	defaultServerAddr = "localhost:3001"
	
	// CLI tool metadata
	cliVersion = "1.0.0"
	cliName = "Terraforming Mars CLI"
)

type CLIClient struct {
	conn     *websocket.Conn
	playerID string
	gameID   string
	done     chan struct{}
	closed   bool
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
		playerID: "cli-" + uuid.New().String()[:8],
		done:     make(chan struct{}),
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
		fmt.Printf("🔗 Connected to game: %s\n", message.GameID)
	}

	switch message.Type {
	case dto.MessageTypeGameUpdated:
		fmt.Printf("🎮 Game updated (Game ID: %s)\n", message.GameID)
		if payload, ok := message.Payload.(map[string]interface{}); ok {
			if gameData, ok := payload["game"].(map[string]interface{}); ok {
				if players, ok := gameData["players"].([]interface{}); ok {
					fmt.Printf("   Players: %d\n", len(players))
				}
				if phase, ok := gameData["phase"].(string); ok {
					fmt.Printf("   Phase: %s\n", phase)
				}
			}
		}
		
	case dto.MessageTypePlayerConnected:
		fmt.Printf("👤 Player connected\n")
		
	case dto.MessageTypeError:
		fmt.Printf("❌ Error: ")
		if payload, ok := message.Payload.(map[string]interface{}); ok {
			if msg, ok := payload["message"].(string); ok {
				fmt.Printf("%s\n", msg)
			}
		}
		
	case dto.MessageTypeFullState:
		fmt.Printf("📊 Full state received (Game ID: %s)\n", message.GameID)
		c.gameID = message.GameID
		
	case dto.MessageTypeAvailableCards:
		fmt.Printf("🃏 Available cards received\n")
		
	default:
		fmt.Printf("📨 Received: %s\n", message.Type)
	}
}

func (c *CLIClient) commandLoop() {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		// Check if we should exit
		select {
		case <-c.done:
			return
		default:
			// Continue with command processing
		}
		
		fmt.Print("tm> ")
		
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
		fmt.Println("❌ Usage: create <gameName>")
		fmt.Println("Example: create \"Mars Colony Alpha\"")
		return
	}
	
	gameName := strings.Join(args, " ")
	
	// For now, we'll create a game by connecting with a generated ID
	// In a real implementation, this would call a create game API endpoint
	gameID := fmt.Sprintf("game-%s-%d", strings.ReplaceAll(strings.ToLower(gameName), " ", "-"), time.Now().Unix())
	
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayerConnect,
		GameID: gameID,
		Payload: dto.PlayerConnectPayload{
			PlayerName: fmt.Sprintf("CLI-Player-%s", c.playerID[4:]),
			GameID:     gameID,
		},
	}
	
	if err := c.conn.WriteJSON(message); err != nil {
		fmt.Printf("❌ Failed to create game: %v\n", err)
		return
	}
	
	// Set gameID locally since we're creating and joining this game
	c.gameID = gameID
	fmt.Printf("🎮 Creating and joining game '%s' with ID: %s\n", gameName, gameID)
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
	fmt.Println("⏭️  Skipping action...")
	
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
		fmt.Printf("❌ Failed to skip action: %v\n", err)
		return
	}
	
	fmt.Println("✅ Action skipped")
}

func (c *CLIClient) raiseTemperature() {
	fmt.Println("🌡️  Raising temperature...")
	
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: map[string]interface{}{
				"type": "raise-temperature",
			},
		},
	}
	
	if err := c.conn.WriteJSON(message); err != nil {
		fmt.Printf("❌ Failed to raise temperature: %v\n", err)
		return
	}
	
	fmt.Println("✅ Temperature raise requested (costs 8 heat)")
}

func (c *CLIClient) raiseOxygen() {
	fmt.Println("💨 Raising oxygen...")
	fmt.Println("💡 This is a placeholder - oxygen raising not implemented in backend yet")
}

func (c *CLIClient) placeOcean() {
	fmt.Println("🌊 Placing ocean...")
	fmt.Println("💡 This is a placeholder - ocean placement not implemented in backend yet")
}

func (c *CLIClient) buyStandardProject() {
	fmt.Println("🏗️  Buying standard project...")
	fmt.Println("💡 This is a placeholder - standard projects not fully implemented yet")
}

func (c *CLIClient) playCard() {
	fmt.Println("🃏 Playing card from hand...")
	fmt.Println("💡 This is a placeholder - card playing not implemented in backend yet")
}

func (c *CLIClient) useCorporationAction() {
	fmt.Println("🏢 Using corporation action...")
	fmt.Println("💡 This is a placeholder - corporation actions not implemented in backend yet")
}

func (c *CLIClient) useCardAction() {
	fmt.Println("⚡ Using card action...")
	fmt.Println("💡 This is a placeholder - card actions not implemented in backend yet")
}

func (c *CLIClient) tradeWithColonies() {
	fmt.Println("🚀 Trading with colonies...")
	fmt.Println("💡 This is a placeholder - colonies not implemented in backend yet")
}

func (c *CLIClient) endTurn() {
	fmt.Println("🏁 Ending turn...")
	fmt.Println("💡 This is a placeholder - turn management not implemented in backend yet")
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