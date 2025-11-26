package turn_test

import (
	"testing"

	"terraforming-mars-backend/internal/session/game/player/turn"
)

func TestNewTurn(t *testing.T) {
	tr := turn.NewTurn()

	if tr.Passed() {
		t.Error("Expected new turn to not be passed")
	}
	if tr.AvailableActions() != 0 {
		t.Errorf("Expected 0 available actions, got %d", tr.AvailableActions())
	}
	if tr.IsConnected() {
		t.Error("Expected new turn to be disconnected")
	}
	if tr.HasActions() {
		t.Error("Expected no actions available")
	}
}

func TestTurn_SetPassed(t *testing.T) {
	tr := turn.NewTurn()

	tr.SetPassed(true)
	if !tr.Passed() {
		t.Error("Expected turn to be passed after SetPassed(true)")
	}

	tr.SetPassed(false)
	if tr.Passed() {
		t.Error("Expected turn to not be passed after SetPassed(false)")
	}
}

func TestTurn_SetAvailableActions(t *testing.T) {
	tr := turn.NewTurn()

	tr.SetAvailableActions(5)
	if tr.AvailableActions() != 5 {
		t.Errorf("Expected 5 available actions, got %d", tr.AvailableActions())
	}
	if !tr.HasActions() {
		t.Error("Expected HasActions() to return true when actions > 0")
	}

	tr.SetAvailableActions(0)
	if tr.AvailableActions() != 0 {
		t.Errorf("Expected 0 available actions, got %d", tr.AvailableActions())
	}
	if tr.HasActions() {
		t.Error("Expected HasActions() to return false when actions = 0")
	}
}

func TestTurn_SetConnectionStatus(t *testing.T) {
	tr := turn.NewTurn()

	tr.SetConnectionStatus(true)
	if !tr.IsConnected() {
		t.Error("Expected turn to be connected after SetConnectionStatus(true)")
	}

	tr.SetConnectionStatus(false)
	if tr.IsConnected() {
		t.Error("Expected turn to be disconnected after SetConnectionStatus(false)")
	}
}

func TestTurn_ConsumeAction(t *testing.T) {
	tests := []struct {
		name           string
		initialActions int
		expectedResult bool
		expectedFinal  int
	}{
		{
			name:           "Consume with actions available",
			initialActions: 3,
			expectedResult: true,
			expectedFinal:  2,
		},
		{
			name:           "Consume with one action available",
			initialActions: 1,
			expectedResult: true,
			expectedFinal:  0,
		},
		{
			name:           "Consume with no actions available",
			initialActions: 0,
			expectedResult: false,
			expectedFinal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := turn.NewTurn()
			tr.SetAvailableActions(tt.initialActions)

			result := tr.ConsumeAction()

			if result != tt.expectedResult {
				t.Errorf("Expected ConsumeAction() to return %v, got %v", tt.expectedResult, result)
			}
			if tr.AvailableActions() != tt.expectedFinal {
				t.Errorf("Expected %d actions remaining, got %d", tt.expectedFinal, tr.AvailableActions())
			}
		})
	}
}

func TestTurn_ResetForGeneration(t *testing.T) {
	tr := turn.NewTurn()

	// Set some state
	tr.SetPassed(true)
	tr.SetAvailableActions(5)
	tr.SetConnectionStatus(true)

	// Reset for generation
	tr.ResetForGeneration()

	// Verify passed and actions are reset
	if tr.Passed() {
		t.Error("Expected passed to be reset to false")
	}
	if tr.AvailableActions() != 0 {
		t.Errorf("Expected actions to be reset to 0, got %d", tr.AvailableActions())
	}
	// Connection status should NOT be reset
	if !tr.IsConnected() {
		t.Error("Expected connection status to remain unchanged")
	}
}

func TestTurn_DeepCopy(t *testing.T) {
	original := turn.NewTurn()
	original.SetPassed(true)
	original.SetAvailableActions(7)
	original.SetConnectionStatus(true)

	copy := original.DeepCopy()

	// Verify copy has same values
	if copy.Passed() != original.Passed() {
		t.Error("Expected copy to have same passed value")
	}
	if copy.AvailableActions() != original.AvailableActions() {
		t.Error("Expected copy to have same available actions")
	}
	if copy.IsConnected() != original.IsConnected() {
		t.Error("Expected copy to have same connection status")
	}

	// Verify modifying copy doesn't affect original
	copy.SetPassed(false)
	copy.SetAvailableActions(0)
	copy.SetConnectionStatus(false)

	if !original.Passed() {
		t.Error("Expected original to still be passed")
	}
	if original.AvailableActions() != 7 {
		t.Error("Expected original to still have 7 actions")
	}
	if !original.IsConnected() {
		t.Error("Expected original to still be connected")
	}
}

func TestTurn_DeepCopy_Nil(t *testing.T) {
	var tr *turn.Turn = nil
	copy := tr.DeepCopy()

	if copy != nil {
		t.Error("Expected DeepCopy of nil to return nil")
	}
}
