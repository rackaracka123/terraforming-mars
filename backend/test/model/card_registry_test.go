package model_test

import (
	"context"
	"errors"
	"testing"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/cards"

	"github.com/stretchr/testify/assert"
)

// MockCardHandler is a simple card handler for testing
type MockCardHandler struct {
	cards.BaseCardHandler
	playExecuted bool
	playError    error
}

func NewMockCardHandler(cardID string, playError error) *MockCardHandler {
	return &MockCardHandler{
		BaseCardHandler: cards.BaseCardHandler{
			CardID:       cardID,
			Requirements: model.CardRequirements{},
		},
		playError: playError,
	}
}

func (m *MockCardHandler) Play(ctx *cards.CardHandlerContext) error {
	m.playExecuted = true
	return m.playError
}

// ValidatingMockPlayerService actually validates resource costs for testing
type ValidatingMockPlayerService struct {
	game *model.Game
}

func (m *ValidatingMockPlayerService) PayResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	return m.ValidateResourceCost(ctx, gameID, playerID, cost)
}

func (m *ValidatingMockPlayerService) AddResources(ctx context.Context, gameID, playerID string, resources model.ResourceSet) error {
	return nil
}

func (m *ValidatingMockPlayerService) ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	// Find the player
	for _, player := range m.game.Players {
		if player.ID == playerID {
			// Check if player has sufficient resources
			if player.Resources.Credits < cost.Credits {
				return errors.New("insufficient credits")
			}
			if player.Resources.Steel < cost.Steel {
				return errors.New("insufficient steel")
			}
			if player.Resources.Titanium < cost.Titanium {
				return errors.New("insufficient titanium")
			}
			if player.Resources.Plants < cost.Plants {
				return errors.New("insufficient plants")
			}
			if player.Resources.Energy < cost.Energy {
				return errors.New("insufficient energy")
			}
			if player.Resources.Heat < cost.Heat {
				return errors.New("insufficient heat")
			}
			return nil
		}
	}
	return errors.New("player not found")
}

func (m *ValidatingMockPlayerService) AddProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error {
	return nil
}

func (m *ValidatingMockPlayerService) RemoveProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error {
	return nil
}

func (m *ValidatingMockPlayerService) ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement model.ResourceSet) error {
	return nil
}

func TestCardHandlerRegistry(t *testing.T) {
	registry := cards.NewCardHandlerRegistry()

	t.Run("register handler", func(t *testing.T) {
		handler := NewMockCardHandler("test-card", nil)
		
		err := registry.Register(handler)
		assert.NoError(t, err)
		
		// Verify it can be retrieved
		retrieved, err := registry.GetHandler("test-card")
		assert.NoError(t, err)
		assert.Equal(t, handler, retrieved)
	})

	t.Run("register duplicate handler", func(t *testing.T) {
		handler1 := NewMockCardHandler("duplicate-card", nil)
		handler2 := NewMockCardHandler("duplicate-card", nil)
		
		err := registry.Register(handler1)
		assert.NoError(t, err)
		
		// Should fail to register duplicate
		err = registry.Register(handler2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("get non-existent handler", func(t *testing.T) {
		_, err := registry.GetHandler("non-existent-card")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no handler registered")
	})

	t.Run("has handler", func(t *testing.T) {
		handler := NewMockCardHandler("exists-card", nil)
		registry.Register(handler)
		
		assert.True(t, registry.HasHandler("exists-card"))
		assert.False(t, registry.HasHandler("does-not-exist"))
	})

	t.Run("get all registered cards", func(t *testing.T) {
		registry := cards.NewCardHandlerRegistry()
		
		handler1 := NewMockCardHandler("card-1", nil)
		handler2 := NewMockCardHandler("card-2", nil)
		handler3 := NewMockCardHandler("card-3", nil)
		
		registry.Register(handler1)
		registry.Register(handler2)
		registry.Register(handler3)
		
		cardIDs := registry.GetAllRegisteredCards()
		assert.Len(t, cardIDs, 3)
		
		// Convert to map for easier checking
		cardMap := make(map[string]bool)
		for _, cardID := range cardIDs {
			cardMap[cardID] = true
		}
		
		assert.True(t, cardMap["card-1"])
		assert.True(t, cardMap["card-2"])
		assert.True(t, cardMap["card-3"])
	})

	t.Run("unregister all", func(t *testing.T) {
		registry := cards.NewCardHandlerRegistry()
		
		handler1 := NewMockCardHandler("card-1", nil)
		handler2 := NewMockCardHandler("card-2", nil)
		
		registry.Register(handler1)
		registry.Register(handler2)
		
		assert.Len(t, registry.GetAllRegisteredCards(), 2)
		
		registry.UnregisterAll()
		
		assert.Len(t, registry.GetAllRegisteredCards(), 0)
		assert.False(t, registry.HasHandler("card-1"))
		assert.False(t, registry.HasHandler("card-2"))
	})
}

func TestGlobalCardRegistry(t *testing.T) {
	// Clean up before and after test
	cards.GlobalCardRegistry.UnregisterAll()
	defer cards.GlobalCardRegistry.UnregisterAll()
	
	t.Run("register card handler globally", func(t *testing.T) {
		handler := NewMockCardHandler("global-test-card", nil)
		
		err := cards.RegisterCardHandler(handler)
		assert.NoError(t, err)
		
		// Should be able to retrieve via convenience function
		retrieved, err := cards.GetCardHandler("global-test-card")
		assert.NoError(t, err)
		assert.Equal(t, handler, retrieved)
	})
}

func TestBaseCardHandler(t *testing.T) {
	handler := &MockCardHandler{
		BaseCardHandler: cards.BaseCardHandler{
			CardID: "test-base-card",
			Requirements: model.CardRequirements{
				MinTemperature: intPtr(-20),
			},
		},
	}
	
	t.Run("get card id", func(t *testing.T) {
		assert.Equal(t, "test-base-card", handler.GetCardID())
	})
	
	t.Run("get requirements", func(t *testing.T) {
		reqs := handler.GetRequirements()
		assert.NotNil(t, reqs.MinTemperature)
		assert.Equal(t, -20, *reqs.MinTemperature)
	})
	
	t.Run("can play - requirements met", func(t *testing.T) {
		game := createTestGame()
		game.GlobalParameters.Temperature = -15 // Meets min -20 requirement
		playerID := game.Players[0].ID
		mockPlayerService := &MockPlayerService{}
		
		ctx := &cards.CardHandlerContext{
			Context:       context.Background(),
			Game:          game,
			PlayerID:      playerID,
			Card:          &model.Card{ID: "test-base-card"},
			PlayerService: mockPlayerService,
		}
		
		err := handler.CanPlay(ctx)
		assert.NoError(t, err)
	})
	
	t.Run("can play - requirements not met", func(t *testing.T) {
		game := createTestGame()
		game.GlobalParameters.Temperature = -25 // Does not meet min -20 requirement
		playerID := game.Players[0].ID
		mockPlayerService := &MockPlayerService{}
		
		ctx := &cards.CardHandlerContext{
			Context:       context.Background(),
			Game:          game,
			PlayerID:      playerID,
			Card:          &model.Card{ID: "test-base-card"},
			PlayerService: mockPlayerService,
		}
		
		err := handler.CanPlay(ctx)
		assert.Error(t, err)
	})
}

func TestActiveCardHandler(t *testing.T) {
	activationCost := &model.ResourceSet{
		Credits: 5,
		Steel:   2,
	}
	
	handler := &cards.ActiveCardHandler{
		BaseCardHandler: cards.BaseCardHandler{
			CardID:       "test-active-card",
			Requirements: model.CardRequirements{},
		},
		ActivationCost: activationCost,
	}
	
	t.Run("can activate - sufficient resources", func(t *testing.T) {
		player := createTestPlayer()
		player.Resources.Credits = 10
		player.Resources.Steel = 5
		game := createTestGame()
		game.Players[0] = player
		mockPlayerService := &MockPlayerService{}
		
		ctx := &cards.CardHandlerContext{
			Context:       context.Background(),
			Game:          game,
			PlayerID:      player.ID,
			PlayerService: mockPlayerService,
		}
		
		err := handler.CanActivate(ctx)
		assert.NoError(t, err)
	})
	
	t.Run("can activate - insufficient resources", func(t *testing.T) {
		player := &model.Player{
			ID:   "test-player",
			Name: "Test Player",
			Resources: model.Resources{
				Credits: 3, // Less than required 5
				Steel:   1, // Less than required 2
			},
		}
		game := &model.Game{
			ID:           "test-game",
			Status:       model.GameStatusActive,
			CurrentPhase: model.GamePhaseAction,
			Players: []model.Player{*player},
		}
		
		// Create a mock that actually validates resource costs
		mockPlayerService := &ValidatingMockPlayerService{game: game}
		
		ctx := &cards.CardHandlerContext{
			Context:       context.Background(),
			Game:          game,
			PlayerID:      player.ID,
			PlayerService: mockPlayerService,
		}
		
		err := handler.CanActivate(ctx)
		assert.Error(t, err)
	})
	
	t.Run("activate - no activation cost", func(t *testing.T) {
		handlerNoCost := &cards.ActiveCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "no-cost-card",
				Requirements: model.CardRequirements{},
			},
			ActivationCost: nil,
		}
		
		player := createTestPlayer()
		game := createTestGame()
		game.Players[0] = player
		mockPlayerService := &MockPlayerService{}
		
		ctx := &cards.CardHandlerContext{
			Context:       context.Background(),
			Game:          game,
			PlayerID:      player.ID,
			PlayerService: mockPlayerService,
		}
		
		err := handlerNoCost.CanActivate(ctx)
		assert.NoError(t, err)
	})
}