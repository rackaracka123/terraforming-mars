package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

	// Panel styles
	panelStyle = baseStyle.
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
	state *GameState
}

// NewUI creates a new UI instance
func NewUI() *UI {
	return &UI{
		state: &GameState{},
	}
}

// UpdateGameState updates the current game state
func (ui *UI) UpdateGameState(state *GameState) {
	ui.state = state
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

	// Join sections horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, sections...)
}

// renderDisconnectedStatus shows when not connected
func (ui *UI) renderDisconnectedStatus() string {
	content := headerStyle.Render("🔌 Disconnected") + "\n" +
		inactiveStyle.Render("Connect to a game to see status")

	return panelStyle.
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
	lines = append(lines, fmt.Sprintf("Generation: %s",
		resourceValueStyle.Render(fmt.Sprintf("%d", ui.state.Generation))))
	lines = append(lines, fmt.Sprintf("Phase: %s",
		productionStyle.Render(ui.state.CurrentPhase)))
	lines = append(lines, fmt.Sprintf("Players: %s",
		resourceValueStyle.Render(fmt.Sprintf("%d", ui.state.TotalPlayers))))

	if ui.state.GameID != "" {
		gameIDShort := ui.state.GameID
		if len(gameIDShort) > 8 {
			gameIDShort = gameIDShort[:8] + "..."
		}
		lines = append(lines, fmt.Sprintf("Game ID: %s",
			baseStyle.Foreground(mutedColor).Render(gameIDShort)))
	}

	content := title + "\n" + strings.Join(lines, "\n")
	return panelStyle.Render(content)
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
	return panelStyle.Render(content)
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
	return panelStyle.Render(content)
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
	return panelStyle.Render(content)
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
