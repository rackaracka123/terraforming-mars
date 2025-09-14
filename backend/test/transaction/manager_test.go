package transaction_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/transaction"
)

// MockPlayerRepository implements transaction.PlayerRepository for testing
type MockPlayerRepository struct {
	players map[string]map[string]*model.Player // gameID -> playerID -> player
	updates []string                            // Track update calls for verification
}

func NewMockPlayerRepository() *MockPlayerRepository {
	return &MockPlayerRepository{
		players: make(map[string]map[string]*model.Player),
		updates: make([]string, 0),
	}
}

func (m *MockPlayerRepository) GetByID(ctx context.Context, gameID, playerID string) (model.Player, error) {
	if game, exists := m.players[gameID]; exists {
		if player, exists := game[playerID]; exists {
			return *player, nil
		}
	}
	return model.Player{}, &model.ErrPlayerNotFound{PlayerID: playerID}
}

func (m *MockPlayerRepository) UpdateResources(ctx context.Context, gameID, playerID string, resources model.Resources) error {
	m.updates = append(m.updates, "UpdateResources:"+gameID+":"+playerID)
	if game, exists := m.players[gameID]; exists {
		if player, exists := game[playerID]; exists {
			player.Resources = resources
			return nil
		}
	}
	return &model.ErrPlayerNotFound{PlayerID: playerID}
}

func (m *MockPlayerRepository) UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error {
	m.updates = append(m.updates, "UpdateAvailableActions:"+gameID+":"+playerID)
	if game, exists := m.players[gameID]; exists {
		if player, exists := game[playerID]; exists {
			player.AvailableActions = actions
			return nil
		}
	}
	return &model.ErrPlayerNotFound{PlayerID: playerID}
}

func (m *MockPlayerRepository) SetPlayer(gameID, playerID string, player *model.Player) {
	if m.players[gameID] == nil {
		m.players[gameID] = make(map[string]*model.Player)
	}
	m.players[gameID][playerID] = player
}

// Stub implementations for full repository.PlayerRepository interface
func (m *MockPlayerRepository) Create(ctx context.Context, gameID string, player model.Player) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) Delete(ctx context.Context, gameID, playerID string) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) ListByGameID(ctx context.Context, gameID string) ([]model.Player, error) {
	return nil, nil // Stub implementation
}
func (m *MockPlayerRepository) UpdateProduction(ctx context.Context, gameID, playerID string, production model.Production) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) UpdateTerraformRating(ctx context.Context, gameID, playerID string, rating int) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) UpdateCorporation(ctx context.Context, gameID, playerID string, corporation string) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) UpdateConnectionStatus(ctx context.Context, gameID, playerID string, status model.ConnectionStatus) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) UpdateIsActive(ctx context.Context, gameID, playerID string, isActive bool) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) UpdateIsReady(ctx context.Context, gameID, playerID string, isReady bool) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) UpdateVictoryPoints(ctx context.Context, gameID, playerID string, points int) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) AddCard(ctx context.Context, gameID, playerID string, cardID string) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) RemoveCard(ctx context.Context, gameID, playerID string, cardID string) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) PlayCard(ctx context.Context, gameID, playerID string, cardID string) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) SetCardSelection(ctx context.Context, gameID, playerID string, selection *model.ProductionPhase) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) SetCardSelectionComplete(ctx context.Context, gameID, playerID string) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) GetCardSelection(ctx context.Context, gameID, playerID string) (*model.ProductionPhase, error) {
	return nil, nil // Stub implementation
}
func (m *MockPlayerRepository) ClearCardSelection(ctx context.Context, gameID, playerID string) error {
	return nil // Stub implementation
}
func (m *MockPlayerRepository) SetStartingSelection(ctx context.Context, gameID, playerID string, cards []model.Card) error {
	return nil // Stub implementation
}

// MockGameRepository implements transaction.GameRepository for testing
type MockGameRepository struct {
	games map[string]*model.Game
}

func NewMockGameRepository() *MockGameRepository {
	return &MockGameRepository{
		games: make(map[string]*model.Game),
	}
}

func (m *MockGameRepository) GetByID(ctx context.Context, gameID string) (model.Game, error) {
	if game, exists := m.games[gameID]; exists {
		return *game, nil
	}
	return model.Game{}, &model.ErrGameNotFound{GameID: gameID}
}

func (m *MockGameRepository) SetGame(gameID string, game *model.Game) {
	m.games[gameID] = game
}

// Stub implementations for full repository.GameRepository interface
func (m *MockGameRepository) Create(ctx context.Context, settings model.GameSettings) (model.Game, error) {
	return model.Game{}, nil // Stub implementation
}
func (m *MockGameRepository) Delete(ctx context.Context, gameID string) error {
	return nil // Stub implementation
}
func (m *MockGameRepository) List(ctx context.Context, status string) ([]model.Game, error) {
	return nil, nil // Stub implementation
}
func (m *MockGameRepository) UpdateStatus(ctx context.Context, gameID string, status model.GameStatus) error {
	return nil // Stub implementation
}
func (m *MockGameRepository) UpdatePhase(ctx context.Context, gameID string, phase model.GamePhase) error {
	return nil // Stub implementation
}
func (m *MockGameRepository) UpdateGlobalParameters(ctx context.Context, gameID string, params model.GlobalParameters) error {
	return nil // Stub implementation
}
func (m *MockGameRepository) UpdateCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	return nil // Stub implementation
}
func (m *MockGameRepository) SetCurrentPlayer(ctx context.Context, gameID string, playerID string) error {
	return nil // Stub implementation
}
func (m *MockGameRepository) SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	return nil // Stub implementation
}
func (m *MockGameRepository) AddPlayerID(ctx context.Context, gameID string, playerID string) error {
	return nil // Stub implementation
}
func (m *MockGameRepository) RemovePlayerID(ctx context.Context, gameID string, playerID string) error {
	return nil // Stub implementation
}
func (m *MockGameRepository) SetHostPlayer(ctx context.Context, gameID string, playerID string) error {
	return nil // Stub implementation
}
func (m *MockGameRepository) UpdateGeneration(ctx context.Context, gameID string, generation int) error {
	return nil // Stub implementation
}
func (m *MockGameRepository) UpdateRemainingActions(ctx context.Context, gameID string, actions int) error {
	return nil // Stub implementation
}

func TestManagerCreation(t *testing.T) {
	playerRepo := NewMockPlayerRepository()
	gameRepo := NewMockGameRepository()

	manager := transaction.NewManager(playerRepo, gameRepo)

	if manager == nil {
		t.Fatal("Expected manager to be created")
	}

	if manager.GetPlayerRepo() != playerRepo {
		t.Error("Expected player repository to match")
	}

	if manager.GetGameRepo() != gameRepo {
		t.Error("Expected game repository to match")
	}
}

func TestExecuteAtomicSuccess(t *testing.T) {
	playerRepo := NewMockPlayerRepository()
	gameRepo := NewMockGameRepository()
	manager := transaction.NewManager(playerRepo, gameRepo)

	// Setup test data
	gameID := "test-game"
	playerID := "test-player"

	player := &model.Player{
		ID:               playerID,
		Resources:        model.Resources{Credits: 100},
		AvailableActions: 2,
	}
	playerRepo.SetPlayer(gameID, playerID, player)

	game := &model.Game{
		ID:           gameID,
		CurrentPhase: model.GamePhaseAction,
		CurrentTurn:  &playerID,
	}
	gameRepo.SetGame(gameID, game)

	ctx := context.Background()
	cost := model.Resources{Credits: 10}

	// Execute atomic transaction
	err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
		return tx.ProcessTurnAction(ctx, gameID, playerID, cost)
	})

	if err != nil {
		t.Fatalf("Expected transaction to succeed, got error: %v", err)
	}

	// Verify state changes
	updatedPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
	if updatedPlayer.Resources.Credits != 90 {
		t.Errorf("Expected credits to be 90, got %d", updatedPlayer.Resources.Credits)
	}
	if updatedPlayer.AvailableActions != 1 {
		t.Errorf("Expected available actions to be 1, got %d", updatedPlayer.AvailableActions)
	}

	// Verify update calls were made
	expectedUpdates := []string{
		"UpdateResources:" + gameID + ":" + playerID,
		"UpdateAvailableActions:" + gameID + ":" + playerID,
	}
	if len(playerRepo.updates) != len(expectedUpdates) {
		t.Errorf("Expected %d updates, got %d", len(expectedUpdates), len(playerRepo.updates))
	}
}

func TestExecuteAtomicFailureAndRollback(t *testing.T) {
	playerRepo := NewMockPlayerRepository()
	gameRepo := NewMockGameRepository()
	manager := transaction.NewManager(playerRepo, gameRepo)

	// Setup test data
	gameID := "test-game"
	playerID := "test-player"

	player := &model.Player{
		ID:               playerID,
		Resources:        model.Resources{Credits: 5}, // Insufficient credits
		AvailableActions: 2,
	}
	playerRepo.SetPlayer(gameID, playerID, player)

	game := &model.Game{
		ID:           gameID,
		CurrentPhase: model.GamePhaseAction,
		CurrentTurn:  &playerID,
	}
	gameRepo.SetGame(gameID, game)

	ctx := context.Background()
	cost := model.Resources{Credits: 10} // More than player has

	// Execute atomic transaction (should fail)
	err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
		return tx.ProcessTurnAction(ctx, gameID, playerID, cost)
	})

	if err == nil {
		t.Fatal("Expected transaction to fail due to insufficient credits")
	}

	// Verify no state changes occurred (rollback)
	updatedPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
	if updatedPlayer.Resources.Credits != 5 {
		t.Errorf("Expected credits to remain 5 after rollback, got %d", updatedPlayer.Resources.Credits)
	}
	if updatedPlayer.AvailableActions != 2 {
		t.Errorf("Expected available actions to remain 2 after rollback, got %d", updatedPlayer.AvailableActions)
	}
}
