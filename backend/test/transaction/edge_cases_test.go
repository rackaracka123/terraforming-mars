package transaction_test

import (
	"context"
	"fmt"
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/transaction"
)

// FailingMockPlayerRepository simulates repository failures
type FailingMockPlayerRepository struct {
	*MockPlayerRepository
	failOnUpdate    bool
	updateCallCount int
}

func NewFailingMockPlayerRepository() *FailingMockPlayerRepository {
	return &FailingMockPlayerRepository{
		MockPlayerRepository: NewMockPlayerRepository(),
		failOnUpdate:         false,
	}
}

func (f *FailingMockPlayerRepository) UpdateResources(ctx context.Context, gameID, playerID string, resources model.Resources) error {
	f.updateCallCount++
	if f.failOnUpdate {
		return fmt.Errorf("simulated repository failure")
	}
	return f.MockPlayerRepository.UpdateResources(ctx, gameID, playerID, resources)
}

func (f *FailingMockPlayerRepository) UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error {
	f.updateCallCount++
	if f.failOnUpdate {
		return fmt.Errorf("simulated repository failure")
	}
	return f.MockPlayerRepository.UpdateAvailableActions(ctx, gameID, playerID, actions)
}

func TestRepositoryFailureScenarios(t *testing.T) {
	t.Run("resource update failure during execution", func(t *testing.T) {
		playerRepo := NewFailingMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		// Setup test data
		initialResources := model.Resources{Credits: 100}
		player := &model.Player{
			ID:               playerID,
			Resources:        initialResources,
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

		// Enable failure after successful validation
		playerRepo.failOnUpdate = true

		// Execute transaction that should fail during resource update
		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessTurnAction(gameID, playerID, model.Resources{Credits: 10})
		})

		if err == nil {
			t.Fatal("Expected transaction to fail due to repository error")
		}

		// Verify no partial updates occurred
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		if !resourcesEqual(finalPlayer.Resources, initialResources) {
			t.Error("Resources should not be modified after repository failure")
		}

		if finalPlayer.AvailableActions != 2 {
			t.Error("Actions should not be modified after repository failure")
		}
	})

	t.Run("partial failure with rollback", func(t *testing.T) {
		playerRepo := NewFailingMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		// Setup test data
		initialResources := model.Resources{Credits: 100}
		player := &model.Player{
			ID:               playerID,
			Resources:        initialResources,
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

		// Execute transaction that fails on second operation
		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			// First operation: deduct resources (should succeed)
			if err := tx.DeductResources(gameID, playerID, model.Resources{Credits: 10}); err != nil {
				return err
			}

			// Enable failure for second operation
			playerRepo.failOnUpdate = true

			// Second operation: consume action (should fail and trigger rollback)
			return tx.ConsumePlayerAction(gameID, playerID)
		})

		if err == nil {
			t.Fatal("Expected transaction to fail")
		}

		// Verify rollback occurred - resources should be restored
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		if !resourcesEqual(finalPlayer.Resources, initialResources) {
			t.Errorf("Expected resources to be rolled back to %+v, got %+v",
				initialResources, finalPlayer.Resources)
		}

		if finalPlayer.AvailableActions != 2 {
			t.Error("Actions should be restored after rollback")
		}
	})
}

func TestBoundaryConditions(t *testing.T) {
	t.Run("exactly zero resources", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		// Player with exactly zero of everything
		player := &model.Player{
			ID:        playerID,
			Resources: model.Resources{}, // All zeros
		}
		playerRepo.SetPlayer(gameID, playerID, player)

		ctx := context.Background()

		// Try to spend zero (should succeed)
		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessPurchase(gameID, playerID, model.Resources{})
		})

		if err != nil {
			t.Errorf("Zero cost purchase should succeed, got: %v", err)
		}

		// Try to spend one credit (should fail)
		err = manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessPurchase(gameID, playerID, model.Resources{Credits: 1})
		})

		if err == nil {
			t.Error("Expected purchase to fail with insufficient credits")
		}
	})

	t.Run("maximum resource values", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		// Player with very large resource amounts
		maxResources := model.Resources{
			Credits:  999999,
			Steel:    999999,
			Titanium: 999999,
			Plants:   999999,
			Energy:   999999,
			Heat:     999999,
		}
		player := &model.Player{
			ID:        playerID,
			Resources: maxResources,
		}
		playerRepo.SetPlayer(gameID, playerID, player)

		ctx := context.Background()
		largeCost := model.Resources{Credits: 500000, Steel: 250000}

		// Execute large transaction
		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessPurchase(gameID, playerID, largeCost)
		})

		if err != nil {
			t.Errorf("Large transaction failed: %v", err)
		}

		// Verify correct arithmetic
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		expectedCredits := 999999 - 500000
		expectedSteel := 999999 - 250000

		if finalPlayer.Resources.Credits != expectedCredits {
			t.Errorf("Expected %d credits, got %d", expectedCredits, finalPlayer.Resources.Credits)
		}
		if finalPlayer.Resources.Steel != expectedSteel {
			t.Errorf("Expected %d steel, got %d", expectedSteel, finalPlayer.Resources.Steel)
		}
	})

	t.Run("negative resource costs (invalid)", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		player := &model.Player{
			ID:        playerID,
			Resources: model.Resources{Credits: 100},
		}
		playerRepo.SetPlayer(gameID, playerID, player)

		ctx := context.Background()

		// Try transaction with negative cost (should be treated as zero or cause issues)
		negativeCost := model.Resources{Credits: -10}

		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessPurchase(gameID, playerID, negativeCost)
		})

		// The behavior with negative costs should be well-defined
		// In this case, we expect it to succeed since negative costs don't make sense
		// but the validation should handle it gracefully
		if err != nil {
			t.Logf("Negative cost transaction failed as expected: %v", err)
		}

		// Verify no unexpected changes
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		if finalPlayer.Resources.Credits < 0 {
			t.Error("Resources should never go negative")
		}
	})

	t.Run("actions at boundary values", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		// Test with exactly 1 action
		player := &model.Player{
			ID:               playerID,
			AvailableActions: 1,
		}
		playerRepo.SetPlayer(gameID, playerID, player)

		game := &model.Game{
			ID:           gameID,
			CurrentPhase: model.GamePhaseAction,
			CurrentTurn:  &playerID,
		}
		gameRepo.SetGame(gameID, game)

		ctx := context.Background()

		// First consumption should succeed
		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ConsumePlayerAction(gameID, playerID)
		})

		if err != nil {
			t.Errorf("First action consumption failed: %v", err)
		}

		// Verify player now has 0 actions
		updatedPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		if updatedPlayer.AvailableActions != 0 {
			t.Errorf("Expected 0 actions, got %d", updatedPlayer.AvailableActions)
		}

		// Second consumption should fail
		err = manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ConsumePlayerAction(gameID, playerID)
		})

		if err == nil {
			t.Error("Expected second action consumption to fail")
		}
	})
}

func TestErrorRecovery(t *testing.T) {
	t.Run("multiple transaction attempts", func(t *testing.T) {
		playerRepo := NewFailingMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		player := &model.Player{
			ID:        playerID,
			Resources: model.Resources{Credits: 100},
		}
		playerRepo.SetPlayer(gameID, playerID, player)

		ctx := context.Background()
		cost := model.Resources{Credits: 10}

		// First attempt fails
		playerRepo.failOnUpdate = true
		err1 := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessPurchase(gameID, playerID, cost)
		})

		if err1 == nil {
			t.Fatal("First attempt should have failed")
		}

		// Second attempt succeeds
		playerRepo.failOnUpdate = false
		err2 := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessPurchase(gameID, playerID, cost)
		})

		if err2 != nil {
			t.Fatalf("Second attempt should have succeeded: %v", err2)
		}

		// Verify final state
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		if finalPlayer.Resources.Credits != 90 {
			t.Errorf("Expected 90 credits after successful transaction, got %d", finalPlayer.Resources.Credits)
		}
	})

	t.Run("transaction state isolation", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		player := &model.Player{
			ID:        playerID,
			Resources: model.Resources{Credits: 100},
		}
		playerRepo.SetPlayer(gameID, playerID, player)

		ctx := context.Background()

		// Create two separate transactions
		tx1 := manager.NewTransaction()
		tx2 := manager.NewTransaction()

		// Add different operations to each
		tx1.DeductResources(gameID, playerID, model.Resources{Credits: 10})
		tx2.DeductResources(gameID, playerID, model.Resources{Credits: 20})

		// Execute tx1
		err1 := tx1.Execute(ctx)
		if err1 != nil {
			t.Fatalf("Transaction 1 failed: %v", err1)
		}

		// Verify state after tx1
		player1State, _ := playerRepo.GetByID(ctx, gameID, playerID)
		if player1State.Resources.Credits != 90 {
			t.Errorf("Expected 90 credits after tx1, got %d", player1State.Resources.Credits)
		}

		// Execute tx2 (should use updated state)
		err2 := tx2.Execute(ctx)
		if err2 != nil {
			t.Fatalf("Transaction 2 failed: %v", err2)
		}

		// Verify final state
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		if finalPlayer.Resources.Credits != 70 {
			t.Errorf("Expected 70 credits after both transactions, got %d", finalPlayer.Resources.Credits)
		}
	})
}
