package transaction_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/transaction/operations"
)

func TestValidateTurnOperation(t *testing.T) {
	tests := []struct {
		name          string
		gamePhase     model.GamePhase
		currentTurn   *string
		playerActions int
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "valid turn",
			gamePhase:     model.GamePhaseAction,
			currentTurn:   stringPtr("player1"),
			playerActions: 2,
			expectError:   false,
		},
		{
			name:          "wrong game phase",
			gamePhase:     model.GamePhaseStartingCardSelection,
			currentTurn:   stringPtr("player1"),
			playerActions: 2,
			expectError:   true,
			errorMsg:      "actions not allowed in phase",
		},
		{
			name:          "not player's turn",
			gamePhase:     model.GamePhaseAction,
			currentTurn:   stringPtr("player2"),
			playerActions: 2,
			expectError:   true,
			errorMsg:      "not your turn",
		},
		{
			name:          "no current turn",
			gamePhase:     model.GamePhaseAction,
			currentTurn:   nil,
			playerActions: 2,
			expectError:   true,
			errorMsg:      "not your turn",
		},
		{
			name:          "no actions remaining",
			gamePhase:     model.GamePhaseAction,
			currentTurn:   stringPtr("player1"),
			playerActions: 0,
			expectError:   true,
			errorMsg:      "no actions remaining",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playerRepo := NewMockPlayerRepository()
			gameRepo := NewMockGameRepository()
			gameID := "test-game"
			playerID := "player1"

			// Setup test data
			player := &model.Player{
				ID:               playerID,
				AvailableActions: tt.playerActions,
			}
			playerRepo.SetPlayer(gameID, playerID, player)

			game := &model.Game{
				ID:           gameID,
				CurrentPhase: tt.gamePhase,
				CurrentTurn:  tt.currentTurn,
			}
			gameRepo.SetGame(gameID, game)

			// Create and execute operation
			op := operations.NewValidateTurnOperation(gameRepo, playerRepo, gameID, playerID)
			err := op.Execute(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestDeductResourcesOperation(t *testing.T) {
	tests := []struct {
		name             string
		initialResources model.Resources
		cost             model.Resources
		expectError      bool
		expectedFinal    model.Resources
	}{
		{
			name:             "sufficient credits",
			initialResources: model.Resources{Credits: 100},
			cost:             model.Resources{Credits: 30},
			expectError:      false,
			expectedFinal:    model.Resources{Credits: 70},
		},
		{
			name:             "insufficient credits",
			initialResources: model.Resources{Credits: 10},
			cost:             model.Resources{Credits: 20},
			expectError:      true,
		},
		{
			name:             "multiple resource types",
			initialResources: model.Resources{Credits: 50, Steel: 10, Titanium: 5},
			cost:             model.Resources{Credits: 20, Steel: 3, Titanium: 2},
			expectError:      false,
			expectedFinal:    model.Resources{Credits: 30, Steel: 7, Titanium: 3},
		},
		{
			name:             "insufficient steel",
			initialResources: model.Resources{Credits: 100, Steel: 2},
			cost:             model.Resources{Credits: 10, Steel: 5},
			expectError:      true,
		},
		{
			name:             "zero cost",
			initialResources: model.Resources{Credits: 50},
			cost:             model.Resources{},
			expectError:      false,
			expectedFinal:    model.Resources{Credits: 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playerRepo := NewMockPlayerRepository()
			gameID := "test-game"
			playerID := "player1"

			// Setup test data
			player := &model.Player{
				ID:        playerID,
				Resources: tt.initialResources,
			}
			playerRepo.SetPlayer(gameID, playerID, player)

			// Create and execute operation
			op, err := operations.NewDeductResourcesOperation(context.Background(), playerRepo, gameID, playerID, tt.cost)
			require.NoError(t, err, "Failed to create operation")
			err = op.Execute(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				// Verify no changes on error
				updatedPlayer, _ := playerRepo.GetByID(context.Background(), gameID, playerID)
				if !resourcesEqual(updatedPlayer.Resources, tt.initialResources) {
					t.Error("Resources should not change on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				// Verify correct deduction
				updatedPlayer, _ := playerRepo.GetByID(context.Background(), gameID, playerID)
				if !resourcesEqual(updatedPlayer.Resources, tt.expectedFinal) {
					t.Errorf("Expected resources %+v, got %+v", tt.expectedFinal, updatedPlayer.Resources)
				}
			}
		})
	}
}

func TestConsumeActionOperation(t *testing.T) {
	tests := []struct {
		name            string
		initialActions  int
		expectError     bool
		expectedActions int
	}{
		{
			name:            "consume from 2 actions",
			initialActions:  2,
			expectError:     false,
			expectedActions: 1,
		},
		{
			name:            "consume from 1 action",
			initialActions:  1,
			expectError:     false,
			expectedActions: 0,
		},
		{
			name:           "cannot consume with 0 actions",
			initialActions: 0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playerRepo := NewMockPlayerRepository()
			gameID := "test-game"
			playerID := "player1"

			// Setup test data
			player := &model.Player{
				ID:               playerID,
				AvailableActions: tt.initialActions,
			}
			playerRepo.SetPlayer(gameID, playerID, player)

			// Create and execute operation
			op, err := operations.NewConsumeActionOperation(context.Background(), playerRepo, gameID, playerID)
			require.NoError(t, err, "Failed to create operation")
			err = op.Execute(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				// Verify action consumption
				updatedPlayer, _ := playerRepo.GetByID(context.Background(), gameID, playerID)
				if updatedPlayer.AvailableActions != tt.expectedActions {
					t.Errorf("Expected %d actions, got %d", tt.expectedActions, updatedPlayer.AvailableActions)
				}
			}
		})
	}
}

func TestOperationRollback(t *testing.T) {
	t.Run("deduct resources rollback", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameID := "test-game"
		playerID := "player1"
		initialResources := model.Resources{Credits: 100, Steel: 10}

		// Setup test data
		player := &model.Player{
			ID:        playerID,
			Resources: initialResources,
		}
		playerRepo.SetPlayer(gameID, playerID, player)

		// Create and execute operation
		cost := model.Resources{Credits: 30, Steel: 5}
		op, err := operations.NewDeductResourcesOperation(context.Background(), playerRepo, gameID, playerID, cost)
		require.NoError(t, err, "Failed to create operation")

		// Execute then rollback
		err = op.Execute(context.Background())
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// Verify resources were deducted
		updatedPlayer, _ := playerRepo.GetByID(context.Background(), gameID, playerID)
		if updatedPlayer.Resources.Credits != 70 {
			t.Error("Resources were not deducted during execute")
		}

		// Rollback
		err = op.Rollback(context.Background())
		if err != nil {
			t.Fatalf("Rollback failed: %v", err)
		}

		// Verify resources were restored
		rolledBackPlayer, _ := playerRepo.GetByID(context.Background(), gameID, playerID)
		if !resourcesEqual(rolledBackPlayer.Resources, initialResources) {
			t.Errorf("Expected resources to be restored to %+v, got %+v",
				initialResources, rolledBackPlayer.Resources)
		}
	})

	t.Run("consume action rollback", func(t *testing.T) {
		playerRepo := NewMockPlayerRepository()
		gameID := "test-game"
		playerID := "player1"
		initialActions := 2

		// Setup test data
		player := &model.Player{
			ID:               playerID,
			AvailableActions: initialActions,
		}
		playerRepo.SetPlayer(gameID, playerID, player)

		// Create and execute operation
		op, err := operations.NewConsumeActionOperation(context.Background(), playerRepo, gameID, playerID)
		require.NoError(t, err, "Failed to create operation")

		// Execute then rollback
		err = op.Execute(context.Background())
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// Verify action was consumed
		updatedPlayer, _ := playerRepo.GetByID(context.Background(), gameID, playerID)
		if updatedPlayer.AvailableActions != 1 {
			t.Error("Action was not consumed during execute")
		}

		// Rollback
		err = op.Rollback(context.Background())
		if err != nil {
			t.Fatalf("Rollback failed: %v", err)
		}

		// Verify actions were restored
		rolledBackPlayer, _ := playerRepo.GetByID(context.Background(), gameID, playerID)
		if rolledBackPlayer.AvailableActions != initialActions {
			t.Errorf("Expected actions to be restored to %d, got %d",
				initialActions, rolledBackPlayer.AvailableActions)
		}
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && len(substr) > 0 &&
				func() bool {
					for i := 1; i <= len(s)-len(substr); i++ {
						if s[i:i+len(substr)] == substr {
							return true
						}
					}
					return false
				}())))
}

func resourcesEqual(a, b model.Resources) bool {
	return a.Credits == b.Credits && a.Steel == b.Steel && a.Titanium == b.Titanium &&
		a.Plants == b.Plants && a.Energy == b.Energy && a.Heat == b.Heat
}
