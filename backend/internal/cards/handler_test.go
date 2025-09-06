package cards

import (
	"context"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPlayerService for testing
type MockPlayerService struct {
	mock.Mock
}

func (m *MockPlayerService) ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	args := m.Called(ctx, gameID, playerID, cost)
	return args.Error(0)
}

func TestBaseCardHandler_GetCardID(t *testing.T) {
	handler := &BaseCardHandler{
		CardID: "test-card",
	}

	assert.Equal(t, "test-card", handler.GetCardID())
}

func TestBaseCardHandler_GetRequirements(t *testing.T) {
	requirements := model.CardRequirements{
		MinTemperature: -24,
		MaxTemperature: -14,
		MinOxygen:      2,
		MaxOxygen:      8,
	}

	handler := &BaseCardHandler{
		CardID:       "test-card",
		Requirements: requirements,
	}

	result := handler.GetRequirements()
	assert.Equal(t, requirements, result)
}

func TestBaseCardHandler_RegisterListeners(t *testing.T) {
	handler := &BaseCardHandler{
		CardID: "test-card",
	}

	eventBus := events.NewInMemoryEventBus()

	// Should not return error for default implementation
	err := handler.RegisterListeners(eventBus)
	assert.NoError(t, err)
}

func TestBaseCardHandler_UnregisterListeners(t *testing.T) {
	handler := &BaseCardHandler{
		CardID: "test-card",
	}

	eventBus := events.NewInMemoryEventBus()

	// Should not return error for default implementation
	err := handler.UnregisterListeners(eventBus)
	assert.NoError(t, err)
}

func TestEventCardHandler_Creation(t *testing.T) {
	handler := &EventCardHandler{
		BaseCardHandler: BaseCardHandler{
			CardID: "event-card",
			Requirements: model.CardRequirements{
				MinTemperature: -20,
			},
		},
	}

	assert.Equal(t, "event-card", handler.GetCardID())
	assert.Equal(t, -20, handler.GetRequirements().MinTemperature)
}

func TestEffectCardHandler_Creation(t *testing.T) {
	handler := &EffectCardHandler{
		BaseCardHandler: BaseCardHandler{
			CardID: "effect-card",
			Requirements: model.CardRequirements{
				MinOxygen: 3,
			},
		},
	}

	assert.Equal(t, "effect-card", handler.GetCardID())
	assert.Equal(t, 3, handler.GetRequirements().MinOxygen)
}

func TestActiveCardHandler_Creation(t *testing.T) {
	activationCost := &model.ResourceSet{
		Credits: 5,
		Energy:  2,
	}

	handler := &ActiveCardHandler{
		BaseCardHandler: BaseCardHandler{
			CardID: "active-card",
		},
		ActivationCost: activationCost,
	}

	assert.Equal(t, "active-card", handler.GetCardID())
	assert.Equal(t, activationCost, handler.ActivationCost)
}

func TestActiveCardHandler_CanActivate_WithCost(t *testing.T) {
	mockPlayerService := &MockPlayerService{}
	activationCost := &model.ResourceSet{
		Credits: 5,
		Energy:  2,
	}

	handler := &ActiveCardHandler{
		BaseCardHandler: BaseCardHandler{
			CardID: "active-card",
		},
		ActivationCost: activationCost,
	}

	ctx := &CardHandlerContext{
		Context:       context.Background(),
		Game:          &model.Game{ID: "game1"},
		PlayerID:      "player1",
		PlayerService: mockPlayerService,
	}

	// Test successful validation
	mockPlayerService.On("ValidateResourceCost", ctx.Context, "game1", "player1", *activationCost).Return(nil)

	err := handler.CanActivate(ctx)
	assert.NoError(t, err)

	mockPlayerService.AssertExpectations(t)
}

func TestActiveCardHandler_CanActivate_NoCost(t *testing.T) {
	handler := &ActiveCardHandler{
		BaseCardHandler: BaseCardHandler{
			CardID: "active-card-no-cost",
		},
		ActivationCost: nil,
	}

	ctx := &CardHandlerContext{
		Context:  context.Background(),
		Game:     &model.Game{ID: "game1"},
		PlayerID: "player1",
	}

	// Should return no error when no activation cost is defined
	err := handler.CanActivate(ctx)
	assert.NoError(t, err)
}

func TestActiveCardHandler_Activate_DefaultImplementation(t *testing.T) {
	handler := &ActiveCardHandler{
		BaseCardHandler: BaseCardHandler{
			CardID: "active-card",
		},
	}

	ctx := &CardHandlerContext{
		Context:  context.Background(),
		Game:     &model.Game{ID: "game1"},
		PlayerID: "player1",
	}

	// Default implementation should return nil
	err := handler.Activate(ctx)
	assert.NoError(t, err)
}

func TestCardHandlerContext_Structure(t *testing.T) {
	game := &model.Game{ID: "test-game"}
	card := &model.Card{ID: "test-card"}
	eventBus := events.NewInMemoryEventBus()
	mockPlayerService := &MockPlayerService{}

	ctx := &CardHandlerContext{
		Context:       context.Background(),
		Game:          game,
		PlayerID:      "player1",
		Card:          card,
		EventBus:      eventBus,
		PlayerService: mockPlayerService,
	}

	assert.NotNil(t, ctx.Context)
	assert.Equal(t, game, ctx.Game)
	assert.Equal(t, "player1", ctx.PlayerID)
	assert.Equal(t, card, ctx.Card)
	assert.Equal(t, eventBus, ctx.EventBus)
	assert.Equal(t, mockPlayerService, ctx.PlayerService)
}
