package cards_test

import (
	"context"
	"terraforming-mars-backend/internal/cards"
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

func (m *MockPlayerService) UpdatePlayerResources(ctx context.Context, gameID, playerID string, newResources model.Resources) error {
	args := m.Called(ctx, gameID, playerID, newResources)
	return args.Error(0)
}

func (m *MockPlayerService) UpdatePlayerProduction(ctx context.Context, gameID, playerID string, newProduction model.Production) error {
	args := m.Called(ctx, gameID, playerID, newProduction)
	return args.Error(0)
}

func (m *MockPlayerService) GetPlayer(ctx context.Context, gameID, playerID string) (*model.Player, error) {
	args := m.Called(ctx, gameID, playerID)
	return args.Get(0).(*model.Player), args.Error(1)
}

func (m *MockPlayerService) ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement model.ResourceSet) error {
	args := m.Called(ctx, gameID, playerID, requirement)
	return args.Error(0)
}

func (m *MockPlayerService) ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	args := m.Called(ctx, gameID, playerID, cost)
	return args.Error(0)
}

func (m *MockPlayerService) AddProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error {
	args := m.Called(ctx, gameID, playerID, production)
	return args.Error(0)
}

func (m *MockPlayerService) PayResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	args := m.Called(ctx, gameID, playerID, cost)
	return args.Error(0)
}

func (m *MockPlayerService) AddResources(ctx context.Context, gameID, playerID string, resources model.ResourceSet) error {
	args := m.Called(ctx, gameID, playerID, resources)
	return args.Error(0)
}

func (m *MockPlayerService) UpdatePlayerTR(ctx context.Context, gameID, playerID string, newTR int) error {
	args := m.Called(ctx, gameID, playerID, newTR)
	return args.Error(0)
}

func (m *MockPlayerService) AddPlayerTR(ctx context.Context, gameID, playerID string, trIncrease int) error {
	args := m.Called(ctx, gameID, playerID, trIncrease)
	return args.Error(0)
}

func (m *MockPlayerService) CanAffordStandardProject(player *model.Player, project model.StandardProject) bool {
	args := m.Called(player, project)
	return args.Bool(0)
}

func (m *MockPlayerService) HasCardsToSell(player *model.Player, count int) bool {
	args := m.Called(player, count)
	return args.Bool(0)
}

func (m *MockPlayerService) GetMaxCardsToSell(player *model.Player) int {
	args := m.Called(player)
	return args.Int(0)
}

func TestBaseCardHandler_GetCardID(t *testing.T) {
	handler := &cards.BaseCardHandler{
		CardID: "test-card",
	}

	assert.Equal(t, "test-card", handler.GetCardID())
}

func TestBaseCardHandler_GetRequirements(t *testing.T) {
	minTemp := -24
	maxTemp := -14
	minOxy := 2
	maxOxy := 8
	requirements := model.CardRequirements{
		MinTemperature: &minTemp,
		MaxTemperature: &maxTemp,
		MinOxygen:      &minOxy,
		MaxOxygen:      &maxOxy,
	}

	handler := &cards.BaseCardHandler{
		CardID:       "test-card",
		Requirements: requirements,
	}

	result := handler.GetRequirements()
	assert.Equal(t, requirements, result)
}

func TestBaseCardHandler_RegisterListeners(t *testing.T) {
	handler := &cards.BaseCardHandler{
		CardID: "test-card",
	}

	eventBus := events.NewInMemoryEventBus()

	// Should not return error for default implementation
	err := handler.RegisterListeners(eventBus)
	assert.NoError(t, err)
}

func TestBaseCardHandler_UnregisterListeners(t *testing.T) {
	handler := &cards.BaseCardHandler{
		CardID: "test-card",
	}

	eventBus := events.NewInMemoryEventBus()

	// Should not return error for default implementation
	err := handler.UnregisterListeners(eventBus)
	assert.NoError(t, err)
}

func TestEventCardHandler_Creation(t *testing.T) {
	minTemp := -20
	handler := &cards.EventCardHandler{
		BaseCardHandler: cards.BaseCardHandler{
			CardID: "event-card",
			Requirements: model.CardRequirements{
				MinTemperature: &minTemp,
			},
		},
	}

	assert.Equal(t, "event-card", handler.GetCardID())
	assert.Equal(t, &minTemp, handler.GetRequirements().MinTemperature)
}

func TestEffectCardHandler_Creation(t *testing.T) {
	minOxy := 3
	handler := &cards.EffectCardHandler{
		BaseCardHandler: cards.BaseCardHandler{
			CardID: "effect-card",
			Requirements: model.CardRequirements{
				MinOxygen: &minOxy,
			},
		},
	}

	assert.Equal(t, "effect-card", handler.GetCardID())
	assert.Equal(t, &minOxy, handler.GetRequirements().MinOxygen)
}

func TestActiveCardHandler_Creation(t *testing.T) {
	activationCost := &model.ResourceSet{
		Credits: 5,
		Energy:  2,
	}

	handler := &cards.ActiveCardHandler{
		BaseCardHandler: cards.BaseCardHandler{
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

	handler := &cards.ActiveCardHandler{
		BaseCardHandler: cards.BaseCardHandler{
			CardID: "active-card",
		},
		ActivationCost: activationCost,
	}

	ctx := &cards.CardHandlerContext{
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
	handler := &cards.ActiveCardHandler{
		BaseCardHandler: cards.BaseCardHandler{
			CardID: "active-card-no-cost",
		},
		ActivationCost: nil,
	}

	ctx := &cards.CardHandlerContext{
		Context:  context.Background(),
		Game:     &model.Game{ID: "game1"},
		PlayerID: "player1",
	}

	// Should return no error when no activation cost is defined
	err := handler.CanActivate(ctx)
	assert.NoError(t, err)
}

func TestActiveCardHandler_Activate_DefaultImplementation(t *testing.T) {
	handler := &cards.ActiveCardHandler{
		BaseCardHandler: cards.BaseCardHandler{
			CardID: "active-card",
		},
	}

	ctx := &cards.CardHandlerContext{
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

	ctx := &cards.CardHandlerContext{
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
