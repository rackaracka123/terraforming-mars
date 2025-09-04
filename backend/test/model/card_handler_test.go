package model_test

import (
	"terraforming-mars-backend/internal/cards"
	"testing"
	"terraforming-mars-backend/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestCardValidation(t *testing.T) {
	game := createTestGame()
	player := &game.Players[0]

	tests := []struct {
		name         string
		requirements model.CardRequirements
		expectError  bool
		description  string
	}{
		{
			name:         "empty requirements",
			requirements: model.CardRequirements{},
			expectError:  false,
			description:  "card with no requirements should always be playable",
		},
		{
			name: "temperature requirement met",
			requirements: model.CardRequirements{
				MinTemperature: intPtr(-30),
			},
			expectError: false,
			description: "current temp is -30, should meet min -30 requirement",
		},
		{
			name: "temperature requirement not met",
			requirements: model.CardRequirements{
				MinTemperature: intPtr(-20),
			},
			expectError: true,
			description: "current temp is -30, should not meet min -20 requirement",
		},
		{
			name: "oxygen requirement met",
			requirements: model.CardRequirements{
				MaxOxygen: intPtr(5),
			},
			expectError: false,
			description: "current oxygen is 0, should meet max 5 requirement",
		},
		{
			name: "ocean requirement not met",
			requirements: model.CardRequirements{
				MinOceans: intPtr(5),
			},
			expectError: true,
			description: "current oceans is 0, should not meet min 5 requirement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cards.ValidateCardRequirements(game, player, tt.requirements)
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestResourceCostValidation(t *testing.T) {
	player := createTestPlayer()
	
	// Give player some resources
	player.Resources = model.Resources{
		Credits:  10,
		Steel:    5,
		Titanium: 3,
		Plants:   2,
		Energy:   1,
		Heat:     4,
	}

	tests := []struct {
		name        string
		cost        model.ResourceSet
		expectError bool
		description string
	}{
		{
			name: "affordable cost",
			cost: model.ResourceSet{
				Credits: 5,
				Steel:   2,
			},
			expectError: false,
			description: "player has enough resources",
		},
		{
			name: "unaffordable credits",
			cost: model.ResourceSet{
				Credits: 15,
			},
			expectError: true,
			description: "player doesn't have enough credits",
		},
		{
			name: "unaffordable steel",
			cost: model.ResourceSet{
				Steel: 10,
			},
			expectError: true,
			description: "player doesn't have enough steel",
		},
		{
			name: "exactly affordable",
			cost: model.ResourceSet{
				Credits:  10,
				Steel:    5,
				Titanium: 3,
			},
			expectError: false,
			description: "player has exactly enough resources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cards.ValidateResourceCost(&player, tt.cost)
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestPayResourceCost(t *testing.T) {
	player := createTestPlayer()
	
	// Give player some resources
	player.Resources = model.Resources{
		Credits:  10,
		Steel:    5,
		Titanium: 3,
		Plants:   2,
		Energy:   1,
		Heat:     4,
	}

	cost := model.ResourceSet{
		Credits: 3,
		Steel:   2,
		Plants:  1,
	}

	cards.PayResourceCost(&player, cost)

	assert.Equal(t, 7, player.Resources.Credits)
	assert.Equal(t, 3, player.Resources.Steel)
	assert.Equal(t, 1, player.Resources.Plants)
	// Other resources should be unchanged
	assert.Equal(t, 3, player.Resources.Titanium)
	assert.Equal(t, 1, player.Resources.Energy)
	assert.Equal(t, 4, player.Resources.Heat)
}

func TestAddResources(t *testing.T) {
	player := createTestPlayer()
	
	// Start with some resources
	player.Resources = model.Resources{
		Credits: 5,
		Steel:   2,
	}

	toAdd := model.ResourceSet{
		Credits:  3,
		Steel:    1,
		Titanium: 2,
	}

	cards.AddResources(&player, toAdd)

	assert.Equal(t, 8, player.Resources.Credits)
	assert.Equal(t, 3, player.Resources.Steel)
	assert.Equal(t, 2, player.Resources.Titanium)
}

func TestAddProduction(t *testing.T) {
	player := createTestPlayer()
	
	// Start with some production
	player.Production = model.Production{
		Credits: 1,
		Energy:  1,
	}

	toAdd := model.ResourceSet{
		Credits: 2,
		Plants:  1,
	}

	cards.AddProduction(&player, toAdd)

	assert.Equal(t, 3, player.Production.Credits)
	assert.Equal(t, 1, player.Production.Plants)
	// Energy production should be unchanged
	assert.Equal(t, 1, player.Production.Energy)
}

func TestGetPlayerTags(t *testing.T) {
	player := createTestPlayer()
	
	// Player has played some cards
	player.PlayedCards = []string{"power-plant", "heat-generators", "water-import"}
	
	tags := cards.GetPlayerTags(&player)
	
	// Check that we got some tags (exact tags depend on card definitions)
	assert.NotEmpty(t, tags, "Player should have tags from played cards")
	
	// Check that we have the expected tags from the cards
	tagMap := make(map[model.CardTag]bool)
	for _, tag := range tags {
		tagMap[tag] = true
	}
	
	// Power plant and heat generators should give us power tags
	assert.True(t, tagMap[model.TagPower], "Should have power tag from power cards")
	
	// Heat generators, power plant should give us building tags
	assert.True(t, tagMap[model.TagBuilding], "Should have building tag from building cards")
	
	// Water import should give us space tag
	assert.True(t, tagMap[model.TagSpace], "Should have space tag from water import")
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

func createTestGame() *model.Game {
	return &model.Game{
		ID:     "test-game",
		Status: model.GameStatusActive,
		GlobalParameters: model.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
		Players: []model.Player{
			createTestPlayer(),
		},
		CurrentPhase: model.GamePhaseAction,
	}
}

func createTestPlayer() model.Player {
	return model.Player{
		ID:              "test-player",
		Name:            "Test Player",
		TerraformRating: 20,
		Resources: model.Resources{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		Production: model.Production{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		Cards:       []string{},
		PlayedCards: []string{},
	}
}