package transaction_test

import (
	"context"
	"sync"
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/transaction"
)

func TestCompleteTransactionScenarios(t *testing.T) {
	t.Run("successful standard project action", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		// Setup: Player with resources and actions in action phase
		player := &model.Player{
			ID:               playerID,
			Resources:        model.Resources{Credits: 50, Steel: 5},
			AvailableActions: 2,
		}
		playerRepo.SetPlayer(gameID, playerID, player)

		game := &model.Game{
			ID:           gameID,
			CurrentPhase: model.GamePhaseAction,
			CurrentTurn:  &playerID,
		}
		gameRepo.SetGame(gameID, game)

		// Execute standard project (costs 23 credits)
		ctx := context.Background()
		cost := model.Resources{Credits: 23}

		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessTurnAction(ctx, gameID, playerID, cost)
		})

		if err != nil {
			t.Fatalf("Transaction failed: %v", err)
		}

		// Verify final state
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		if finalPlayer.Resources.Credits != 27 {
			t.Errorf("Expected 27 credits, got %d", finalPlayer.Resources.Credits)
		}
		if finalPlayer.AvailableActions != 1 {
			t.Errorf("Expected 1 action remaining, got %d", finalPlayer.AvailableActions)
		}
	})

	t.Run("card purchase (no action consumed)", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		// Setup: Player buying cards
		player := &model.Player{
			ID:               playerID,
			Resources:        model.Resources{Credits: 30},
			AvailableActions: 2,
		}
		playerRepo.SetPlayer(gameID, playerID, player)

		// Execute card purchase (costs 3 credits per card, buying 4 cards = 12 credits)
		ctx := context.Background()
		cardCost := model.Resources{Credits: 12}

		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessPurchase(ctx, gameID, playerID, cardCost)
		})

		if err != nil {
			t.Fatalf("Transaction failed: %v", err)
		}

		// Verify: Resources deducted but no action consumed
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		if finalPlayer.Resources.Credits != 18 {
			t.Errorf("Expected 18 credits after purchase, got %d", finalPlayer.Resources.Credits)
		}
		if finalPlayer.AvailableActions != 2 {
			t.Errorf("Expected actions to remain 2 (no action consumed), got %d", finalPlayer.AvailableActions)
		}
	})

	t.Run("transaction failure with proper rollback", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		// Setup: Player with limited resources
		initialResources := model.Resources{Credits: 10, Steel: 3}
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

		// Try complex transaction that should fail on second operation
		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			// First: deduct small amount (should succeed)
			if err := tx.DeductResources(ctx, gameID, playerID, model.Resources{Credits: 5}); err != nil {
				return err
			}

			// Second: deduct more than remaining (should fail)
			if err := tx.DeductResources(ctx, gameID, playerID, model.Resources{Credits: 20}); err != nil {
				return err
			}

			// Third: consume action (should not execute due to previous failure)
			return tx.ConsumePlayerAction(ctx, gameID, playerID)
		})

		if err == nil {
			t.Fatal("Expected transaction to fail")
		}

		// Verify complete rollback - no changes should have occurred
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		if !resourcesEqual(finalPlayer.Resources, initialResources) {
			t.Errorf("Expected resources to be rolled back to %+v, got %+v",
				initialResources, finalPlayer.Resources)
		}
		if finalPlayer.AvailableActions != 2 {
			t.Errorf("Expected actions to remain 2 after rollback, got %d", finalPlayer.AvailableActions)
		}
	})

	t.Run("multiple resource types transaction", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		// Setup: Player with various resources
		player := &model.Player{
			ID:               playerID,
			Resources:        model.Resources{Credits: 100, Steel: 10, Titanium: 5, Plants: 8, Energy: 6, Heat: 12},
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

		// Execute action with complex cost
		complexCost := model.Resources{Credits: 20, Steel: 3, Titanium: 2, Energy: 4}

		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessTurnAction(ctx, gameID, playerID, complexCost)
		})

		if err != nil {
			t.Fatalf("Transaction failed: %v", err)
		}

		// Verify all resources were properly deducted
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		expected := model.Resources{Credits: 80, Steel: 7, Titanium: 3, Plants: 8, Energy: 2, Heat: 12}
		if !resourcesEqual(finalPlayer.Resources, expected) {
			t.Errorf("Expected resources %+v, got %+v", expected, finalPlayer.Resources)
		}
		if finalPlayer.AvailableActions != 1 {
			t.Errorf("Expected 1 action remaining, got %d", finalPlayer.AvailableActions)
		}
	})
}

func TestConcurrentTransactions(t *testing.T) {
	t.Run("concurrent resource access", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		// Setup: Player with resources
		player := &model.Player{
			ID:        playerID,
			Resources: model.Resources{Credits: 1000}, // Plenty of resources
		}
		playerRepo.SetPlayer(gameID, playerID, player)

		ctx := context.Background()
		var wg sync.WaitGroup
		errors := make(chan error, 10)

		// Launch multiple concurrent transactions
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(cost int) {
				defer wg.Done()

				err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
					return tx.ProcessPurchase(ctx, gameID, playerID, model.Resources{Credits: cost})
				})

				if err != nil {
					errors <- err
				}
			}(10) // Each transaction costs 10 credits
		}

		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Errorf("Concurrent transaction failed: %v", err)
		}

		// Verify final state is consistent
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		expectedCredits := 1000 - (10 * 10) // 1000 - 100 = 900
		if finalPlayer.Resources.Credits != expectedCredits {
			t.Errorf("Expected %d credits after concurrent transactions, got %d",
				expectedCredits, finalPlayer.Resources.Credits)
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("zero cost transaction", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		player := &model.Player{
			ID:               playerID,
			Resources:        model.Resources{Credits: 50},
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
		zeroCost := model.Resources{} // All zeros

		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessTurnAction(ctx, gameID, playerID, zeroCost)
		})

		if err != nil {
			t.Fatalf("Zero cost transaction failed: %v", err)
		}

		// Verify: No resource changes, but action still consumed
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		if finalPlayer.Resources.Credits != 50 {
			t.Errorf("Expected credits to remain 50, got %d", finalPlayer.Resources.Credits)
		}
		if finalPlayer.AvailableActions != 1 {
			t.Errorf("Expected action to be consumed, got %d actions", finalPlayer.AvailableActions)
		}
	})

	t.Run("transaction with empty operations", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		ctx := context.Background()

		// Execute transaction with no operations
		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			// Do nothing
			return nil
		})

		if err != nil {
			t.Errorf("Empty transaction should succeed, got error: %v", err)
		}
	})

	t.Run("exact resource amounts", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameRepo := NewMockGameRepository()
		manager := transaction.NewManager(playerRepo, gameRepo)

		gameID := "test-game"
		playerID := "player1"

		// Setup: Player with exact amount needed
		exactAmount := model.Resources{Credits: 25, Steel: 3}
		player := &model.Player{
			ID:               playerID,
			Resources:        exactAmount,
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

		// Execute transaction using exact amount
		err := manager.ExecuteAtomic(ctx, func(tx *transaction.Transaction) error {
			return tx.ProcessTurnAction(ctx, gameID, playerID, exactAmount)
		})

		if err != nil {
			t.Fatalf("Exact amount transaction failed: %v", err)
		}

		// Verify: All resources consumed, action consumed
		finalPlayer, _ := playerRepo.GetByID(ctx, gameID, playerID)
		zeroResources := model.Resources{}
		if !resourcesEqual(finalPlayer.Resources, zeroResources) {
			t.Errorf("Expected all resources consumed, got %+v", finalPlayer.Resources)
		}
		if finalPlayer.AvailableActions != 1 {
			t.Errorf("Expected 1 action remaining, got %d", finalPlayer.AvailableActions)
		}
	})
}
