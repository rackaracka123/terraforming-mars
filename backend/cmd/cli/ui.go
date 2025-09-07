package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"terraforming-mars-backend/internal/model"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// UI styling constants
var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#06B6D4") // Cyan
	accentColor    = lipgloss.Color("#10B981") // Green
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	textColor      = lipgloss.Color("#F8FAFC") // Light gray
	mutedColor     = lipgloss.Color("#94A3B8") // Muted gray

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Panel styles (base - will be modified dynamically)
	basePanelStyle = baseStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			Margin(1, 0)

	headerStyle = baseStyle.
			Foreground(primaryColor).
			Bold(true).
			Align(lipgloss.Center)

	// Resource styles
	resourceStyle = baseStyle.
			Padding(0, 1)

	resourceValueStyle = baseStyle.
				Bold(true).
				Foreground(accentColor)

	productionStyle = baseStyle.
			Foreground(secondaryColor)

	// Status styles
	activeStyle = baseStyle.
			Foreground(accentColor).
			Bold(true)

	inactiveStyle = baseStyle.
			Foreground(mutedColor)
)

// UI manages the terminal UI display
type UI struct {
	state         *GameState
	lastCommand   string
	lastResult    string
	commandOutput []string
	termWidth     int
	termHeight    int
}

// NewUI creates a new UI instance
func NewUI() *UI {
	ui := &UI{
		state:         &GameState{},
		commandOutput: make([]string, 0),
	}
	ui.updateTerminalSize()
	return ui
}

// updateTerminalSize detects and updates the terminal dimensions
func (ui *UI) updateTerminalSize() {
	// Try to get terminal size from stdout first
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Try stderr if stdout fails
		width, height, err = term.GetSize(int(os.Stderr.Fd()))
	}
	if err != nil {
		// Try stdin if both stdout and stderr fail
		width, height, err = term.GetSize(int(os.Stdin.Fd()))
	}

	if err != nil {
		// Check environment variables as last resort
		if cols := os.Getenv("COLUMNS"); cols != "" {
			if w, parseErr := strconv.Atoi(cols); parseErr == nil {
				ui.termWidth = w
			} else {
				ui.termWidth = 80
			}
		} else {
			ui.termWidth = 80
		}

		if lines := os.Getenv("LINES"); lines != "" {
			if h, parseErr := strconv.Atoi(lines); parseErr == nil {
				ui.termHeight = h
			} else {
				ui.termHeight = 24
			}
		} else {
			ui.termHeight = 24
		}
	} else {
		ui.termWidth = width
		ui.termHeight = height
	}

	// Ensure minimum width
	if ui.termWidth < 40 {
		ui.termWidth = 40
	}
}

// getPanelStyle returns a panel style appropriate for current terminal size
func (ui *UI) getPanelStyle() lipgloss.Style {
	style := basePanelStyle

	// For horizontal layout, limit panel width to fit multiple panels
	if ui.termWidth >= 80 {
		maxPanels := 4 // default to 4 panels
		if ui.state != nil && ui.state.Player != nil && len(ui.state.Player.PlayedCards) > 0 {
			maxPanels = 5 // add played cards panel
		}
		maxPanelWidth := (ui.termWidth - 8) / maxPanels // panels with some margin
		style = style.Width(maxPanelWidth)
	}

	return style
}

// UpdateGameState updates the current game state
func (ui *UI) UpdateGameState(state *GameState) {
	ui.state = state
}

// SetLastCommand sets the last command and its result for display
func (ui *UI) SetLastCommand(command, result string) {
	ui.lastCommand = command
	ui.lastResult = result
}

// ClearCommandOutput clears the command output area
func (ui *UI) ClearCommandOutput() {
	ui.lastCommand = ""
	ui.lastResult = ""
}

// RenderStatus renders the complete status display
func (ui *UI) RenderStatus() string {
	if ui.state == nil || !ui.state.IsConnected {
		return ui.renderDisconnectedStatus()
	}

	sections := []string{
		ui.renderGameInfo(),
		ui.renderPlayerResources(),
		ui.renderPlayerProduction(),
		ui.renderGlobalParameters(),
	}

	// Add played cards panel if player has played cards
	if ui.state.Player != nil && len(ui.state.Player.PlayedCards) > 0 {
		sections = append(sections, ui.renderPlayedCards())
	}

	// Always use horizontal layout for better space utilization
	// Only stack vertically for very narrow terminals (< 80 chars)
	if ui.termWidth < 80 {
		return strings.Join(sections, "\n")
	} else {
		return lipgloss.JoinHorizontal(lipgloss.Top, sections...)
	}
}

// RenderFullDisplay renders the complete display with status and command areas
func (ui *UI) RenderFullDisplay() string {
	// Update terminal size in case it changed
	ui.updateTerminalSize()

	var parts []string

	// Status area at the top
	parts = append(parts, ui.RenderStatus())

	// Separator line using terminal width
	separator := strings.Repeat("─", ui.termWidth)
	parts = append(parts, baseStyle.Foreground(mutedColor).Render(separator))

	// Always show available actions/help
	parts = append(parts, ui.renderContextualActions())

	// Another separator before command output
	if ui.lastCommand != "" || ui.lastResult != "" {
		parts = append(parts, baseStyle.Foreground(mutedColor).Render(separator))
		parts = append(parts, ui.renderCommandArea())
	}

	return strings.Join(parts, "\n")
}

// renderCommandArea renders the last command and its result
func (ui *UI) renderCommandArea() string {
	var lines []string

	if ui.lastCommand != "" {
		commandLine := baseStyle.Foreground(primaryColor).Render("tm> ") +
			baseStyle.Render(ui.lastCommand)
		lines = append(lines, commandLine)
	}

	if ui.lastResult != "" {
		lines = append(lines, ui.lastResult)
	}

	return strings.Join(lines, "\n")
}

// renderDisconnectedStatus shows when not connected
func (ui *UI) renderDisconnectedStatus() string {
	content := headerStyle.Render("🔌 Disconnected") + "\n" +
		inactiveStyle.Render("Connect to a game to see status")

	return ui.getPanelStyle().
		BorderForeground(warningColor).
		Render(content)
}

// renderGameInfo renders basic game information
func (ui *UI) renderGameInfo() string {
	if ui.state == nil {
		return ""
	}

	title := headerStyle.Render("🎮 Game Status")

	var lines []string
	lines = append(lines, "")

	// Show game status with appropriate styling
	statusText := string(ui.state.GameStatus)
	var statusStyle lipgloss.Style
	switch ui.state.GameStatus {
	case model.GameStatusLobby:
		statusStyle = baseStyle.Foreground(warningColor)
		statusText = "🔄 Lobby"
	case model.GameStatusActive:
		statusStyle = baseStyle.Foreground(accentColor)
		statusText = "🎮 Active"
	case model.GameStatusCompleted:
		statusStyle = baseStyle.Foreground(mutedColor)
		statusText = "✅ Complete"
	default:
		statusStyle = baseStyle.Foreground(mutedColor)
	}
	lines = append(lines, fmt.Sprintf("Status: %s", statusStyle.Render(statusText)))

	lines = append(lines, fmt.Sprintf("Generation: %s",
		resourceValueStyle.Render(fmt.Sprintf("%d", ui.state.Generation))))
	lines = append(lines, fmt.Sprintf("Phase: %s",
		productionStyle.Render(string(ui.state.CurrentPhase))))
	lines = append(lines, fmt.Sprintf("Players: %s",
		resourceValueStyle.Render(fmt.Sprintf("%d", ui.state.TotalPlayers))))

	// Show host status if player is available
	if ui.state.Player != nil && ui.state.HostPlayerID != "" {
		if ui.state.Player.ID == ui.state.HostPlayerID {
			lines = append(lines, fmt.Sprintf("Role: %s",
				activeStyle.Render("👑 Host")))
		} else {
			lines = append(lines, fmt.Sprintf("Role: %s",
				baseStyle.Foreground(mutedColor).Render("👤 Player")))
		}
	}

	if ui.state.GameID != "" {
		gameIDShort := ui.state.GameID
		if len(gameIDShort) > 8 {
			gameIDShort = gameIDShort[:8] + "..."
		}
		lines = append(lines, fmt.Sprintf("Game ID: %s",
			baseStyle.Foreground(mutedColor).Render(gameIDShort)))
	}

	content := title + "\n" + strings.Join(lines, "\n")
	return ui.getPanelStyle().Render(content)
}

// renderPlayerResources renders the player's resources
func (ui *UI) renderPlayerResources() string {
	if ui.state == nil || ui.state.Player == nil {
		return ""
	}

	title := headerStyle.Render("💰 Resources")

	resources := ui.state.Player.Resources

	var lines []string
	lines = append(lines, "")
	lines = append(lines, ui.formatResourceLine("Credits", "💳", resources.Credits))
	lines = append(lines, ui.formatResourceLine("Steel", "🔩", resources.Steel))
	lines = append(lines, ui.formatResourceLine("Titanium", "🔗", resources.Titanium))
	lines = append(lines, ui.formatResourceLine("Plants", "🌱", resources.Plants))
	lines = append(lines, ui.formatResourceLine("Energy", "⚡", resources.Energy))
	lines = append(lines, ui.formatResourceLine("Heat", "🌡️", resources.Heat))

	// Terraform Rating
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("TR: %s",
		activeStyle.Render(fmt.Sprintf("%d", ui.state.Player.TerraformRating))))

	content := title + "\n" + strings.Join(lines, "\n")
	return ui.getPanelStyle().Render(content)
}

// renderPlayerProduction renders the player's production
func (ui *UI) renderPlayerProduction() string {
	if ui.state == nil || ui.state.Player == nil {
		return ""
	}

	title := headerStyle.Render("🏭 Production")

	production := ui.state.Player.Production

	var lines []string
	lines = append(lines, "")
	lines = append(lines, ui.formatProductionLine("Credits", "💳", production.Credits))
	lines = append(lines, ui.formatProductionLine("Steel", "🔩", production.Steel))
	lines = append(lines, ui.formatProductionLine("Titanium", "🔗", production.Titanium))
	lines = append(lines, ui.formatProductionLine("Plants", "🌱", production.Plants))
	lines = append(lines, ui.formatProductionLine("Energy", "⚡", production.Energy))
	lines = append(lines, ui.formatProductionLine("Heat", "🌡️", production.Heat))

	content := title + "\n" + strings.Join(lines, "\n")
	return ui.getPanelStyle().Render(content)
}

// renderGlobalParameters renders the global terraforming parameters
func (ui *UI) renderGlobalParameters() string {
	if ui.state == nil || ui.state.GlobalParameters == nil {
		return ""
	}

	title := headerStyle.Render("🌍 Mars Status")

	params := ui.state.GlobalParameters

	var lines []string
	lines = append(lines, "")
	lines = append(lines, ui.formatGlobalParam("Temperature", "🌡️", params.Temperature, "°C"))
	lines = append(lines, ui.formatGlobalParam("Oxygen", "💨", params.Oxygen, "%"))
	lines = append(lines, ui.formatGlobalParam("Oceans", "🌊", params.Oceans, ""))

	content := title + "\n" + strings.Join(lines, "\n")
	return ui.getPanelStyle().Render(content)
}

// renderPlayedCards renders the player's played cards for reference
func (ui *UI) renderPlayedCards() string {
	if ui.state == nil || ui.state.Player == nil {
		return ""
	}

	title := headerStyle.Render("🃏 Played Cards")

	playedCards := ui.state.Player.PlayedCards
	if len(playedCards) == 0 {
		content := title + "\n\n" + baseStyle.Foreground(mutedColor).Render("No cards played yet")
		return ui.getPanelStyle().Render(content)
	}

	var lines []string
	lines = append(lines, "")

	// Show first few played cards
	maxDisplay := 5
	for i, cardID := range playedCards {
		if i >= maxDisplay {
			lines = append(lines, baseStyle.Foreground(mutedColor).Render(fmt.Sprintf("... and %d more", len(playedCards)-maxDisplay)))
			break
		}
		lines = append(lines, fmt.Sprintf("• %s", baseStyle.Render(cardID)))
	}

	content := title + "\n" + strings.Join(lines, "\n")
	return ui.getPanelStyle().Render(content)
}

// formatResourceLine formats a resource line with icon and value
func (ui *UI) formatResourceLine(name, icon string, value int) string {
	nameFormatted := resourceStyle.Render(fmt.Sprintf("%s %s:", icon, name))
	valueFormatted := resourceValueStyle.Render(fmt.Sprintf("%d", value))
	return fmt.Sprintf("%-12s %s", nameFormatted, valueFormatted)
}

// formatProductionLine formats a production line with icon and value
func (ui *UI) formatProductionLine(name, icon string, value int) string {
	nameFormatted := resourceStyle.Render(fmt.Sprintf("%s %s:", icon, name))
	var valueFormatted string
	if value > 0 {
		valueFormatted = productionStyle.Render(fmt.Sprintf("+%d", value))
	} else if value < 0 {
		valueFormatted = baseStyle.Foreground(errorColor).Render(fmt.Sprintf("%d", value))
	} else {
		valueFormatted = baseStyle.Foreground(mutedColor).Render("0")
	}
	return fmt.Sprintf("%-12s %s", nameFormatted, valueFormatted)
}

// formatGlobalParam formats a global parameter line
func (ui *UI) formatGlobalParam(name, icon string, value int, unit string) string {
	nameFormatted := resourceStyle.Render(fmt.Sprintf("%s %s:", icon, name))
	valueStr := fmt.Sprintf("%d%s", value, unit)
	valueFormatted := resourceValueStyle.Render(valueStr)
	return fmt.Sprintf("%-12s %s", nameFormatted, valueFormatted)
}

// ClearScreen clears the terminal screen
func (ui *UI) ClearScreen() {
	fmt.Print("\033[2J\033[H")
}

// RenderPrompt renders the command prompt with consistent styling
func (ui *UI) RenderPrompt() string {
	return baseStyle.Foreground(primaryColor).Render("tm> ")
}

// RenderMessage renders a status message with appropriate styling
func (ui *UI) RenderMessage(msgType string, message string) string {
	var style lipgloss.Style
	var icon string

	switch msgType {
	case "success":
		style = baseStyle.Foreground(accentColor)
		icon = "✅"
	case "error":
		style = baseStyle.Foreground(errorColor)
		icon = "❌"
	case "warning":
		style = baseStyle.Foreground(warningColor)
		icon = "⚠️"
	case "info":
		style = baseStyle.Foreground(secondaryColor)
		icon = "ℹ️"
	default:
		style = baseStyle
		icon = "📨"
	}

	return style.Render(fmt.Sprintf("%s %s", icon, message))
}

// renderContextualActions displays available actions based on current game state
func (ui *UI) renderContextualActions() string {
	if ui.state == nil {
		return ui.renderDisconnectedActions()
	}

	if !ui.state.IsConnected || ui.state.GameID == "" {
		return ui.renderDisconnectedActions()
	} else if ui.state.GameStatus == model.GameStatusLobby {
		return ui.renderLobbyActions()
	} else {
		return ui.renderActiveGameActions()
	}
}

// renderDisconnectedActions shows connection options when not connected
func (ui *UI) renderDisconnectedActions() string {
	title := headerStyle.Render("🔌 Available Commands")
	
	content := `
• caj <name>       - Create and join new game
• join <id> <name> - Join existing game by ID
• help             - Show detailed help
• status           - Show connection status
• quit             - Exit CLI

💡 Connect to a game with 'caj <name>' to start playing!`

	return basePanelStyle.
		BorderForeground(warningColor).
		Width(ui.termWidth - 4).
		Render(title + content)
}

// renderLobbyActions shows lobby-specific actions
func (ui *UI) renderLobbyActions() string {
	title := headerStyle.Render("🎯 Lobby Actions")
	
	var content string
	if ui.state.Player != nil && ui.state.Player.ID == ui.state.HostPlayerID {
		// Host in lobby
		content = `
• 0 - Start Game
• help, status, players, overview, quit

💡 You are the host! Press 0 to start when ready.`
	} else {
		// Non-host in lobby
		content = `
⏳ Waiting for host to start the game...

• help, status, players, overview, quit

💡 The host will start the game when all players are ready.`
	}

	return basePanelStyle.
		BorderForeground(secondaryColor).
		Width(ui.termWidth - 4).
		Render(title + content)
}

// renderActiveGameActions shows game actions based on current state
func (ui *UI) renderActiveGameActions() string {
	if ui.state.Player == nil {
		return ""
	}

	title := headerStyle.Render("🎯 Available Actions")
	
	player := ui.state.Player
	var content strings.Builder
	
	content.WriteString("\n• 0 - End Turn / Skip Action")
	
	// Resource conversions
	if player.Resources.Heat >= 8 {
		content.WriteString(fmt.Sprintf("\n• 1 - Convert Heat → Temp ✓ (have %d)", player.Resources.Heat))
	} else {
		content.WriteString(fmt.Sprintf("\n• 1 - Convert Heat → Temp ✗ (need 8, have %d)", player.Resources.Heat))
	}
	
	if player.Resources.Plants >= 8 {
		content.WriteString(fmt.Sprintf("\n• 2 - Convert Plants → Greenery ✓ (have %d)", player.Resources.Plants))
	} else {
		content.WriteString(fmt.Sprintf("\n• 2 - Convert Plants → Greenery ✗ (need 8, have %d)", player.Resources.Plants))
	}
	
	// Standard projects
	if player.Resources.Credits >= 14 {
		content.WriteString(fmt.Sprintf("\n• 3 - Asteroid Project ✓ (have %d MC)", player.Resources.Credits))
	} else {
		content.WriteString(fmt.Sprintf("\n• 3 - Asteroid Project ✗ (need 14 MC, have %d)", player.Resources.Credits))
	}
	
	if player.Resources.Credits >= 18 {
		content.WriteString(fmt.Sprintf("\n• 4 - Ocean Project ✓ (have %d MC)", player.Resources.Credits))
	} else {
		content.WriteString(fmt.Sprintf("\n• 4 - Ocean Project ✗ (need 18 MC, have %d)", player.Resources.Credits))
	}
	
	// Cards
	if len(player.Cards) > 0 {
		content.WriteString(fmt.Sprintf("\n• 5 - Play Card ✓ (%d available)", len(player.Cards)))
	} else {
		content.WriteString("\n• 5 - Play Card ✗ (no cards)")
	}
	
	if player.Resources.Credits >= 3 {
		content.WriteString(fmt.Sprintf("\n• 6 - Buy Cards ✓ (have %d MC)", player.Resources.Credits))
	} else {
		content.WriteString(fmt.Sprintf("\n• 6 - Buy Cards ✗ (need 3 MC, have %d)", player.Resources.Credits))
	}
	
	// Corporation
	if player.Corporation != "" {
		content.WriteString(fmt.Sprintf("\n• 7 - Corporation Action [%s]", player.Corporation))
	} else {
		content.WriteString("\n• 7 - Corporation Action [none]")
	}
	
	content.WriteString("\n\n💡 Type number (0-7) or use commands: cards, buy, convert, etc.")

	return basePanelStyle.
		BorderForeground(accentColor).
		Width(ui.termWidth - 4).
		Render(title + content.String())
}
