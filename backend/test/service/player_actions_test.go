package service

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardService_OnPlayCard_WithManualTriggers_AddsActions(t *testing.T) {
	// Setup test data
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Setup repositories
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := NewMockCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Create a mock session manager that doesn't actually send messages
	sessionManager := &MockSessionManager{}

	// Create card service
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber, forcedActionManager)
	playerID := "player1"

	// Create a test game
	gameSettings := model.GameSettings{
		MaxPlayers:      4,
		DevelopmentMode: true,
	}
	game, err := gameRepo.Create(ctx, gameSettings)
	require.NoError(t, err)
	gameID := game.ID

	// Set current turn to the test player
	err = gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	require.NoError(t, err)

	// Create a test player with available actions and credits
	player := model.Player{
		ID:               playerID,
		Name:             "Test Player",
		Cards:            []string{"manual-card"},
		Resources:        model.Resources{Credits: 10},
		AvailableActions: 2,
		Actions:          []model.PlayerAction{}, // Start with no actions
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Setup the mock card repository with test cards
	cardWithManual := model.Card{
		ID:          "manual-card",
		Name:        "Test Manual Card",
		Type:        model.CardTypeActive,
		Cost:        5,
		Description: "A card with manual actions",
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerManual,
					},
				},
				Inputs: []model.ResourceCondition{
					{
						Type:   model.ResourceEnergy,
						Amount: 1,
						Target: model.TargetSelfPlayer,
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:   model.ResourceCredits,
						Amount: 2,
						Target: model.TargetSelfPlayer,
					},
				},
			},
		},
	}
	cardRepo.cards["manual-card"] = cardWithManual

	// Play the card
	err = cardService.OnPlayCard(ctx, gameID, playerID, "manual-card", makePayment(ctx, cardRepo, "manual-card"), nil, nil)
	require.NoError(t, err)

	// Verify the player has the action
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)

	assert.Len(t, updatedPlayer.Actions, 1, "Player should have one action after playing card with manual trigger")
	assert.Equal(t, "manual-card", updatedPlayer.Actions[0].CardID)
	assert.Equal(t, "Test Manual Card", updatedPlayer.Actions[0].CardName)
	assert.Equal(t, 0, updatedPlayer.Actions[0].BehaviorIndex)
	assert.Len(t, updatedPlayer.Actions[0].Behavior.Triggers, 1)
	assert.Equal(t, model.ResourceTriggerManual, updatedPlayer.Actions[0].Behavior.Triggers[0].Type)

	// Verify card was removed from hand and added to played cards
	assert.NotContains(t, updatedPlayer.Cards, "manual-card")
	assert.Contains(t, updatedPlayer.PlayedCards, "manual-card")

	// Verify available actions was decremented
	assert.Equal(t, 1, updatedPlayer.AvailableActions)

	// Verify credits were deducted
	assert.Equal(t, 5, updatedPlayer.Resources.Credits)
}

func TestCardService_OnPlayCard_WithoutManualTriggers_NoActions(t *testing.T) {
	// Setup test data
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Setup repositories
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := NewMockCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Create a mock session manager that doesn't actually send messages
	sessionManager := &MockSessionManager{}

	// Create card service
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber, forcedActionManager)

	playerID := "player1"

	// Create a test game
	gameSettings := model.GameSettings{
		MaxPlayers:      4,
		DevelopmentMode: true,
	}
	game, err := gameRepo.Create(ctx, gameSettings)
	require.NoError(t, err)
	gameID := game.ID

	// Set current turn to the test player
	err = gameRepo.SetCurrentTurn(ctx, gameID, &playerID)
	require.NoError(t, err)

	// Create a test player with available actions and credits
	player := model.Player{
		ID:               playerID,
		Name:             "Test Player",
		Cards:            []string{"auto-card"},
		Resources:        model.Resources{Credits: 10},
		AvailableActions: 2,
		Actions:          []model.PlayerAction{}, // Start with no actions
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Create a test card with only automatic triggers
	cardWithAuto := model.Card{
		ID:          "auto-card",
		Name:        "Test Auto Card",
		Type:        model.CardTypeAutomated,
		Cost:        3,
		Description: "A card with automatic effects only",
		Behaviors: []model.CardBehavior{
			{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerAuto,
					},
				},
				Outputs: []model.ResourceCondition{
					{
						Type:   model.ResourceCreditsProduction,
						Amount: 1,
						Target: model.TargetSelfPlayer,
					},
				},
			},
		},
	}
	cardRepo.cards["auto-card"] = cardWithAuto

	// Play the card
	err = cardService.OnPlayCard(ctx, gameID, playerID, "auto-card", makePayment(ctx, cardRepo, "auto-card"), nil, nil)
	require.NoError(t, err)

	// Verify the player has no actions
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)

	assert.Len(t, updatedPlayer.Actions, 0, "Player should have no actions after playing card without manual triggers")

	// Verify card was still processed normally
	assert.NotContains(t, updatedPlayer.Cards, "auto-card")
	assert.Contains(t, updatedPlayer.PlayedCards, "auto-card")
	assert.Equal(t, 1, updatedPlayer.AvailableActions)
	assert.Equal(t, 7, updatedPlayer.Resources.Credits) // 10 - 3 cost
}

func TestPlayerRepository_UpdatePlayerActions(t *testing.T) {
	// Setup
	ctx := context.Background()
	eventBus := events.NewEventBus()
	playerRepo := repository.NewPlayerRepository(eventBus)
	gameID := "test-game"
	playerID := "player1"

	// Create a test player
	player := model.Player{
		ID:      playerID,
		Name:    "Test Player",
		Actions: []model.PlayerAction{},
	}
	err := playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	// Create test actions
	actions := []model.PlayerAction{
		{
			CardID:        "card1",
			CardName:      "Test Card 1",
			BehaviorIndex: 0,
			Behavior: model.CardBehavior{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerManual,
					},
				},
			},
		},
		{
			CardID:        "card2",
			CardName:      "Test Card 2",
			BehaviorIndex: 1,
			Behavior: model.CardBehavior{
				Triggers: []model.Trigger{
					{
						Type: model.ResourceTriggerManual,
					},
				},
			},
		},
	}

	// Update player actions
	err = playerRepo.UpdatePlayerActions(ctx, gameID, playerID, actions)
	require.NoError(t, err)

	// Verify actions were updated
	updatedPlayer, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)

	assert.Len(t, updatedPlayer.Actions, 2)
	assert.Equal(t, "card1", updatedPlayer.Actions[0].CardID)
	assert.Equal(t, "Test Card 1", updatedPlayer.Actions[0].CardName)
	assert.Equal(t, "card2", updatedPlayer.Actions[1].CardID)
	assert.Equal(t, "Test Card 2", updatedPlayer.Actions[1].CardName)
}

// MockSessionManager implements session.SessionManager for testing
type MockSessionManager struct{}

func (m *MockSessionManager) Broadcast(gameID string) error {
	// Mock implementation - do nothing
	return nil
}

func (m *MockSessionManager) Send(gameID, playerID string) error {
	// Mock implementation - do nothing
	return nil
}

// MockCardRepository implements repository.CardRepository for testing
type MockCardRepository struct {
	cards map[string]model.Card
}

func NewMockCardRepository() *MockCardRepository {
	return &MockCardRepository{
		cards: make(map[string]model.Card),
	}
}

func (m *MockCardRepository) LoadCards(ctx context.Context) error {
	return nil
}

func (m *MockCardRepository) GetCardByID(ctx context.Context, cardID string) (*model.Card, error) {
	if card, exists := m.cards[cardID]; exists {
		return &card, nil
	}
	return nil, nil
}

func (m *MockCardRepository) ListCardsByIdMap(ctx context.Context, cardIDs map[string]struct{}) (map[string]model.Card, error) {
	result := make(map[string]model.Card)
	for id := range cardIDs {
		if card, exists := m.cards[id]; exists {
			result[id] = card
		}
	}
	return result, nil
}

func (m *MockCardRepository) GetAllCards(ctx context.Context) ([]model.Card, error) {
	cards := make([]model.Card, 0, len(m.cards))
	for _, card := range m.cards {
		cards = append(cards, card)
	}
	return cards, nil
}

func (m *MockCardRepository) GetProjectCards(ctx context.Context) ([]model.Card, error) {
	return m.GetAllCards(ctx)
}

func (m *MockCardRepository) GetCorporationCards(ctx context.Context) ([]model.Card, error) {
	return []model.Card{}, nil
}

func (m *MockCardRepository) GetPreludeCards(ctx context.Context) ([]model.Card, error) {
	return []model.Card{}, nil
}

func (m *MockCardRepository) GetStartingCardPool(ctx context.Context) ([]model.Card, error) {
	return []model.Card{}, nil
}

func (m *MockCardRepository) FilterCardsByRequirements(ctx context.Context, cards []model.Card, gameState interface{}) ([]model.Card, error) {
	return cards, nil
}

func (m *MockCardRepository) GetCardsByAllTags(ctx context.Context, tags []model.CardTag) ([]model.Card, error) {
	return []model.Card{}, nil
}

func (m *MockCardRepository) GetCardsByType(ctx context.Context, cardType model.CardType) ([]model.Card, error) {
	return []model.Card{}, nil
}

func (m *MockCardRepository) GetCardsByTag(ctx context.Context, tag model.CardTag) ([]model.Card, error) {
	return []model.Card{}, nil
}

func (m *MockCardRepository) GetCardsByCostRange(ctx context.Context, minCost, maxCost int) ([]model.Card, error) {
	return []model.Card{}, nil
}

func (m *MockCardRepository) GetCardsByTags(ctx context.Context, tags []model.CardTag) ([]model.Card, error) {
	return []model.Card{}, nil
}

func (m *MockCardRepository) GetCorporations(ctx context.Context) ([]model.Card, error) {
	return []model.Card{}, nil
}
