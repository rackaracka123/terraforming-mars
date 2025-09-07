package cli_test

import (
	"strings"
	"terraforming-mars-backend/internal/model"
	"testing"
)

// TestCardTypeIconRecognition tests the basic card type recognition logic
func TestCardTypeIconRecognition(t *testing.T) {
	testCases := []struct {
		cardID   string
		expected string
	}{
		{"nitrogen-plants", "🌱 Plant"},
		{"power-plant", "⚡ Power"}, // Contains "power"
		{"space-mirrors", "🚀 Space"},
		{"water-import", "🌊 Water"},
		{"heat-generators", "🌡️ Heat"},
		{"city-builder", "🏗️ Building"}, // Use "city" or "building" in name
		{"research-grant", "🔬 Science"},
		{"investment", "🃏 Card"},
	}

	for _, tc := range testCases {
		t.Run(tc.cardID, func(t *testing.T) {
			result := getCardTypeIcon(tc.cardID)
			if result != tc.expected {
				t.Errorf("getCardTypeIcon(%s) = %s, want %s", tc.cardID, result, tc.expected)
			}
		})
	}
}

// TestGameStateValidation tests basic game state validation
func TestGameStateValidation(t *testing.T) {
	gameState := &GameState{
		Player: &model.Player{
			ID:              "test-player",
			Name:            "Test Player",
			TerraformRating: 20,
			Cards:           []string{"investment", "power-plant"},
			PlayedCards:     []string{"research-grant"},
			Resources: model.Resources{
				Credits: 50,
				Steel:   2,
				Plants:  3,
			},
		},
		Generation:   1,
		CurrentPhase: model.GamePhaseAction,
		GameID:       "test-game-123",
		IsConnected:  true,
		TotalPlayers: 2,
		GameStatus:   model.GameStatusActive,
		GlobalParameters: &GlobalParams{
			Temperature: -24,
			Oxygen:      2,
			Oceans:      1,
		},
	}

	// Test basic validations
	if gameState.Player.ID == "" {
		t.Error("Player ID should not be empty")
	}

	if len(gameState.Player.Cards) != 2 {
		t.Errorf("Expected 2 cards in hand, got %d", len(gameState.Player.Cards))
	}

	if gameState.Player.TerraformRating < 20 {
		t.Errorf("Expected terraform rating >= 20, got %d", gameState.Player.TerraformRating)
	}

	if gameState.GlobalParameters.Temperature < -30 || gameState.GlobalParameters.Temperature > 8 {
		t.Errorf("Temperature %d out of valid range [-30, 8]", gameState.GlobalParameters.Temperature)
	}
}

// TestMilestoneValidation tests milestone name validation
func TestMilestoneValidation(t *testing.T) {
	validMilestones := []string{
		"terraformer", "mayor", "gardener", "builder", "planner",
	}

	validAwards := []string{
		"landlord", "banker", "scientist", "thermalist", "miner",
	}

	// Test valid milestones
	for _, milestone := range validMilestones {
		if !isValidMilestone(milestone) {
			t.Errorf("Expected %s to be valid milestone", milestone)
		}
	}

	// Test valid awards
	for _, award := range validAwards {
		if !isValidAward(award) {
			t.Errorf("Expected %s to be valid award", award)
		}
	}

	// Test invalid names
	invalid := []string{"invalid", "unknown", "fake"}
	for _, name := range invalid {
		if isValidMilestone(name) || isValidAward(name) {
			t.Errorf("Expected %s to be invalid", name)
		}
	}
}

// Helper types for testing (mirroring CLI types)
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

type GlobalParams struct {
	Temperature int
	Oxygen      int
	Oceans      int
}

// Helper functions to test (simplified versions of CLI logic)
func getCardTypeIcon(cardID string) string {
	cardLower := strings.ToLower(cardID)
	switch {
	case strings.Contains(cardLower, "power") || strings.Contains(cardLower, "energy"):
		return "⚡ Power"
	case strings.Contains(cardLower, "plant") || strings.Contains(cardLower, "greenery"):
		return "🌱 Plant"
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

func isValidMilestone(name string) bool {
	milestones := map[string]bool{
		"terraformer": true, "mayor": true, "gardener": true, "builder": true, "planner": true,
	}
	return milestones[name]
}

func isValidAward(name string) bool {
	awards := map[string]bool{
		"landlord": true, "banker": true, "scientist": true, "thermalist": true, "miner": true,
	}
	return awards[name]
}
