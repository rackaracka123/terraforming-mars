package model_test

import (
	"context"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper functions for creating pointers
func stringPtr(s string) *string { return &s }
func intPtr(i int) *int         { return &i }

// Helper function to create a test game
func createTestGame() *model.Game {
	return &model.Game{
		ID:           "test-game",
		Status:       model.GameStatusActive,
		CurrentPhase: model.GamePhaseAction,
		Players: []model.Player{
			createTestPlayer(),
		},
		GlobalParameters: model.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
	}
}

// Helper function to create a test player
func createTestPlayer() model.Player {
	return model.Player{
		ID:   "test-player-1",
		Name: "Test Player",
		Resources: model.Resources{
			Credits: 42,
			Steel:   0,
			Titanium: 0,
			Plants:  0,
			Energy:  0,
			Heat:    0,
		},
		Production: model.Production{
			Credits: 1,
			Steel:   0,
			Titanium: 0,
			Plants:  0,
			Energy:  1,
			Heat:    1,
		},
		Cards:        []string{},
		PlayedCards:  []string{},
		TerraformRating: 20,
	}
}

// Mock PlayerService for testing
type MockPlayerService struct{}

func (m *MockPlayerService) PayResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	return nil
}

func (m *MockPlayerService) AddResources(ctx context.Context, gameID, playerID string, resources model.ResourceSet) error {
	return nil
}

func (m *MockPlayerService) ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	return nil
}

func (m *MockPlayerService) AddProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error {
	return nil
}

func (m *MockPlayerService) RemoveProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error {
	return nil
}

func (m *MockPlayerService) ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement model.ResourceSet) error {
	return nil
}

func TestCardValidation(t *testing.T) {
	game := createTestGame()
	playerID := game.Players[0].ID
	mockPlayerService := &MockPlayerService{}

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
			err := cards.ValidateCardRequirements(context.Background(), game, playerID, mockPlayerService, tt.requirements)
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestGetPlayerTags(t *testing.T) {
	game := createTestGame()
	playerID := game.Players[0].ID
	mockPlayerService := &MockPlayerService{}

	// Add some played cards to the player
	game.Players[0].PlayedCards = []string{"power-plant", "early-settlement"}

	tags, err := cards.GetPlayerTags(context.Background(), game, playerID, mockPlayerService)
	assert.NoError(t, err)

	// Should have tags from the played cards
	assert.NotNil(t, tags)
	// The actual tag validation would depend on the card definitions
}

// Integration test for PlayerService
func TestPlayerServiceIntegration(t *testing.T) {
	gameRepo := repository.NewGameRepository()
	eventBus := events.NewInMemoryEventBus()
	eventRepository := events.NewEventRepository(eventBus)
	playerService := service.NewPlayerService(gameRepo, eventBus, eventRepository)

	// Create a game through the repository
	gameSettings := model.GameSettings{MaxPlayers: 4}
	game, err := gameRepo.CreateGame(gameSettings)
	assert.NoError(t, err)
	
	// Add a test player to the game
	testPlayer := createTestPlayer()
	game.Players = append(game.Players, testPlayer)
	err = gameRepo.UpdateGame(game)
	assert.NoError(t, err)

	playerID := game.Players[0].ID

	// Test adding resources
	err = playerService.AddResources(context.Background(), game.ID, playerID, model.ResourceSet{
		Credits: 5,
		Steel:   2,
	})
	assert.NoError(t, err)

	// Test adding production
	err = playerService.AddProduction(context.Background(), game.ID, playerID, model.ResourceSet{
		Credits: 1,
		Plants:  1,
	})
	assert.NoError(t, err)

	// Test paying costs
	err = playerService.PayResourceCost(context.Background(), game.ID, playerID, model.ResourceSet{
		Credits: 3,
	})
	assert.NoError(t, err)
}