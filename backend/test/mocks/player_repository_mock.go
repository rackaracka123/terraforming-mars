package mocks

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session/types"
)

// MockPlayerRepository is a mock implementation of player.Repository for testing
type MockPlayerRepository struct {
	players map[string]map[string]types.Player // gameID -> playerID -> Player
}

// NewMockPlayerRepository creates a new mock player repository
func NewMockPlayerRepository() *MockPlayerRepository {
	return &MockPlayerRepository{
		players: make(map[string]map[string]types.Player),
	}
}

// SetPlayer sets a player in the mock repository
func (m *MockPlayerRepository) SetPlayer(gameID, playerID string, player types.Player) {
	if m.players[gameID] == nil {
		m.players[gameID] = make(map[string]types.Player)
	}
	m.players[gameID][playerID] = player
}

// GetByID retrieves a player by ID
func (m *MockPlayerRepository) GetByID(ctx context.Context, gameID, playerID string) (*types.Player, error) {
	if gamePlayers, ok := m.players[gameID]; ok {
		if player, ok := gamePlayers[playerID]; ok {
			playerCopy := player
			return &playerCopy, nil
		}
	}
	return nil, fmt.Errorf("player not found: %s", playerID)
}

// ListByGameID retrieves all players for a game
func (m *MockPlayerRepository) ListByGameID(ctx context.Context, gameID string) ([]*types.Player, error) {
	if gamePlayers, ok := m.players[gameID]; ok {
		players := make([]*types.Player, 0, len(gamePlayers))
		for _, player := range gamePlayers {
			playerCopy := player
			players = append(players, &playerCopy)
		}
		return players, nil
	}
	return []*types.Player{}, nil
}

// UpdateResources updates player resources
func (m *MockPlayerRepository) UpdateResources(ctx context.Context, gameID, playerID string, resources types.Resources) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.Resources = resources
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

// UpdateProduction updates player production
func (m *MockPlayerRepository) UpdateProduction(ctx context.Context, gameID, playerID string, production types.Production) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.Production = production
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

// UpdateResourceStorage updates player resource storage
func (m *MockPlayerRepository) UpdateResourceStorage(ctx context.Context, gameID, playerID string, storage map[string]int) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.ResourceStorage = storage
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

// UpdateAvailableActions updates player available actions
func (m *MockPlayerRepository) UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.AvailableActions = actions
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

// UpdatePlayerActions updates player actions
func (m *MockPlayerRepository) UpdatePlayerActions(ctx context.Context, gameID, playerID string, actions []types.PlayerAction) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.Actions = actions
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

// UpdateTerraformRating updates terraform rating
func (m *MockPlayerRepository) UpdateTerraformRating(ctx context.Context, gameID, playerID string, rating int) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.TerraformRating = rating
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

// UpdatePendingCardDrawSelection updates pending card draw selection
func (m *MockPlayerRepository) UpdatePendingCardDrawSelection(ctx context.Context, gameID, playerID string, selection *types.PendingCardDrawSelection) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.PendingCardDrawSelection = selection
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

// Stub implementations for other required methods
func (m *MockPlayerRepository) Create(ctx context.Context, gameID string, player *types.Player) error {
	m.SetPlayer(gameID, player.ID, *player)
	return nil
}

func (m *MockPlayerRepository) Delete(ctx context.Context, gameID, playerID string) error {
	if m.players[gameID] != nil {
		delete(m.players[gameID], playerID)
	}
	return nil
}

func (m *MockPlayerRepository) AddCard(ctx context.Context, gameID, playerID, cardID string) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.Cards = append(player.Cards, cardID)
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdateCorporation(ctx context.Context, gameID, playerID string, corporation types.Card) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.Corporation = &corporation
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdateCards(ctx context.Context, gameID, playerID string, cards []string) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.Cards = cards
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdatePlayedCards(ctx context.Context, gameID, playerID string, playedCards []string) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.PlayedCards = playedCards
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdateEffects(ctx context.Context, gameID, playerID string, effects []types.PlayerEffect) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.Effects = effects
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdatePendingTileSelection(ctx context.Context, gameID, playerID string, selection *types.PendingTileSelection) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.PendingTileSelection = selection
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdatePendingTileSelectionQueue(ctx context.Context, gameID, playerID string, queue *types.PendingTileSelectionQueue) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.PendingTileSelectionQueue = queue
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdatePendingCardSelection(ctx context.Context, gameID, playerID string, selection *types.PendingCardSelection) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.PendingCardSelection = selection
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdateForcedFirstAction(ctx context.Context, gameID, playerID string, action *types.ForcedFirstAction) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.ForcedFirstAction = action
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) ClearPendingTileSelection(ctx context.Context, gameID, playerID string) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.PendingTileSelection = nil
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) ClearPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.PendingTileSelectionQueue = nil
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) ClearPendingCardSelection(ctx context.Context, gameID, playerID string) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.PendingCardSelection = nil
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) CompleteStartingSelection(ctx context.Context, gameID, playerID string) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.SelectStartingCardsPhase = nil
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) CompleteProductionSelection(ctx context.Context, gameID, playerID string) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	if player.ProductionPhase != nil {
		player.ProductionPhase.SelectionComplete = true
	}
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

// Additional required interface methods
func (m *MockPlayerRepository) CreateTileQueue(ctx context.Context, gameID, playerID, cardID string, tileTypes []string) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.PendingTileSelectionQueue = &types.PendingTileSelectionQueue{
		Items:  tileTypes,
		Source: cardID,
	}
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) GetPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) (*types.PendingTileSelectionQueue, error) {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return nil, err
	}
	return player.PendingTileSelectionQueue, nil
}

func (m *MockPlayerRepository) ProcessNextTileInQueue(ctx context.Context, gameID, playerID string) (string, error) {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return "", err
	}
	if player.PendingTileSelectionQueue == nil || len(player.PendingTileSelectionQueue.Items) == 0 {
		return "", fmt.Errorf("no tiles in queue")
	}

	tileType := player.PendingTileSelectionQueue.Items[0]
	player.PendingTileSelectionQueue.Items = player.PendingTileSelectionQueue.Items[1:]

	if len(player.PendingTileSelectionQueue.Items) == 0 {
		player.PendingTileSelectionQueue = nil
	}

	m.SetPlayer(gameID, playerID, *player)
	return tileType, nil
}

func (m *MockPlayerRepository) RemoveCardFromHand(ctx context.Context, gameID, playerID, cardID string) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}

	for i, id := range player.Cards {
		if id == cardID {
			player.Cards = append(player.Cards[:i], player.Cards[i+1:]...)
			break
		}
	}

	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) SetCorporation(ctx context.Context, gameID, playerID, corporationID string) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	// For mock purposes, just set a simple corporation reference
	player.Corporation = &types.Card{ID: corporationID}
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) SetStartingCardsSelection(ctx context.Context, gameID, playerID string, cardIDs []string, corpIDs []string) error {
	// Mock implementation - simplified
	return nil
}

func (m *MockPlayerRepository) UpdateConnectionStatus(ctx context.Context, gameID, playerID string, isConnected bool) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.IsConnected = isConnected
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.Passed = passed
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdatePaymentSubstitutes(ctx context.Context, gameID, playerID string, substitutes []types.PaymentSubstitute) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.PaymentSubstitutes = substitutes
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdatePlayerEffects(ctx context.Context, gameID, playerID string, effects []types.PlayerEffect) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.Effects = effects
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdateProductionPhase(ctx context.Context, gameID, playerID string, phase *types.ProductionPhase) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.ProductionPhase = phase
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdateRequirementModifiers(ctx context.Context, gameID, playerID string, modifiers []types.RequirementModifier) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.RequirementModifiers = modifiers
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdateSelectStartingCardsPhase(ctx context.Context, gameID, playerID string, phase *types.SelectStartingCardsPhase) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.SelectStartingCardsPhase = phase
	m.SetPlayer(gameID, playerID, *player)
	return nil
}

func (m *MockPlayerRepository) UpdateVictoryPoints(ctx context.Context, gameID, playerID string, victoryPoints int) error {
	player, err := m.GetByID(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	player.VictoryPoints = victoryPoints
	m.SetPlayer(gameID, playerID, *player)
	return nil
}
