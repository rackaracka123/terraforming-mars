package model_test

import (
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
		player := &game.Players[0]
		
		ctx := &cards.CardHandlerContext{
			Game:   game,
			Player: player,
			Card:   &model.Card{ID: "test-base-card"},
		}
		
		err := handler.CanPlay(ctx)
		assert.NoError(t, err)
	})
	
	t.Run("can play - requirements not met", func(t *testing.T) {
		game := createTestGame()
		game.GlobalParameters.Temperature = -25 // Does not meet min -20 requirement
		player := &game.Players[0]
		
		ctx := &cards.CardHandlerContext{
			Game:   game,
			Player: player,
			Card:   &model.Card{ID: "test-base-card"},
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
		
		ctx := &cards.CardHandlerContext{
			Player: &player,
		}
		
		err := handler.CanActivate(ctx)
		assert.NoError(t, err)
	})
	
	t.Run("can activate - insufficient resources", func(t *testing.T) {
		player := createTestPlayer()
		player.Resources.Credits = 3
		player.Resources.Steel = 1
		
		ctx := &cards.CardHandlerContext{
			Player: &player,
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
		ctx := &cards.CardHandlerContext{
			Player: &player,
		}
		
		err := handlerNoCost.CanActivate(ctx)
		assert.NoError(t, err)
	})
}