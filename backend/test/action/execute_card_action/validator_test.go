package execute_card_action

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action/execute_card_action"
	"terraforming-mars-backend/internal/session/types"
	"terraforming-mars-backend/test/mocks"
)

func TestValidator_ValidateActionInputs_SufficientResources(t *testing.T) {
	// Setup mock repository
	mockPlayerRepo := mocks.NewMockPlayerRepository()
	validator := execute_card_action.NewValidator(mockPlayerRepo)

	// Create player with resources
	player := types.Player{
		ID: "player1",
		Resources: types.Resources{
			Credits:  20,
			Steel:    5,
			Titanium: 3,
			Plants:   10,
			Energy:   5,
			Heat:     8,
		},
	}
	mockPlayerRepo.SetPlayer("game1", "player1", player)

	// Create action with inputs
	action := &types.PlayerAction{
		CardID: "card1",
		Behavior: types.CardBehavior{
			Inputs: []types.ResourceCondition{
				{Type: types.ResourceCredits, Amount: 10},
				{Type: types.ResourceSteel, Amount: 2},
			},
		},
	}

	ctx := context.Background()
	err := validator.ValidateActionInputs(ctx, "game1", "player1", action, nil)

	if err != nil {
		t.Errorf("ValidateActionInputs() should not error with sufficient resources, got: %v", err)
	}
}

func TestValidator_ValidateActionInputs_InsufficientResources(t *testing.T) {
	mockPlayerRepo := mocks.NewMockPlayerRepository()
	validator := execute_card_action.NewValidator(mockPlayerRepo)

	// Create player with limited resources
	player := types.Player{
		ID: "player1",
		Resources: types.Resources{
			Credits: 5, // Not enough
			Steel:   1, // Not enough
		},
	}
	mockPlayerRepo.SetPlayer("game1", "player1", player)

	tests := []struct {
		name   string
		inputs []types.ResourceCondition
	}{
		{
			name: "insufficient credits",
			inputs: []types.ResourceCondition{
				{Type: types.ResourceCredits, Amount: 10},
			},
		},
		{
			name: "insufficient steel",
			inputs: []types.ResourceCondition{
				{Type: types.ResourceSteel, Amount: 5},
			},
		},
		{
			name: "insufficient multiple resources",
			inputs: []types.ResourceCondition{
				{Type: types.ResourceCredits, Amount: 10},
				{Type: types.ResourceSteel, Amount: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &types.PlayerAction{
				CardID: "card1",
				Behavior: types.CardBehavior{
					Inputs: tt.inputs,
				},
			}

			ctx := context.Background()
			err := validator.ValidateActionInputs(ctx, "game1", "player1", action, nil)

			if err == nil {
				t.Errorf("ValidateActionInputs() should error with insufficient resources")
			}
		})
	}
}

func TestValidator_ValidateActionInputs_WithChoiceIndex(t *testing.T) {
	mockPlayerRepo := mocks.NewMockPlayerRepository()
	validator := execute_card_action.NewValidator(mockPlayerRepo)

	// Create player with resources
	player := types.Player{
		ID: "player1",
		Resources: types.Resources{
			Credits:  20,
			Titanium: 5,
		},
	}
	mockPlayerRepo.SetPlayer("game1", "player1", player)

	// Create action with base inputs and choice-specific inputs
	choiceIndex := 0
	action := &types.PlayerAction{
		CardID: "card1",
		Behavior: types.CardBehavior{
			Inputs: []types.ResourceCondition{
				{Type: types.ResourceCredits, Amount: 5},
			},
			Choices: []types.Choice{
				{
					Inputs: []types.ResourceCondition{
						{Type: types.ResourceTitanium, Amount: 2},
					},
				},
			},
		},
	}

	ctx := context.Background()
	err := validator.ValidateActionInputs(ctx, "game1", "player1", action, &choiceIndex)

	if err != nil {
		t.Errorf("ValidateActionInputs() should validate both base and choice inputs, got: %v", err)
	}
}

func TestValidator_ValidateActionInputs_CardStorage(t *testing.T) {
	mockPlayerRepo := mocks.NewMockPlayerRepository()
	validator := execute_card_action.NewValidator(mockPlayerRepo)

	// Create player with card storage
	player := types.Player{
		ID: "player1",
		ResourceStorage: map[string]int{
			"card1": 5, // 5 animals on this card
		},
	}
	mockPlayerRepo.SetPlayer("game1", "player1", player)

	tests := []struct {
		name       string
		storage    int
		wantErr    bool
		errMessage string
	}{
		{
			name:    "sufficient card storage",
			storage: 3,
			wantErr: false,
		},
		{
			name:       "insufficient card storage",
			storage:    10,
			wantErr:    true,
			errMessage: "insufficient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &types.PlayerAction{
				CardID: "card1",
				Behavior: types.CardBehavior{
					Inputs: []types.ResourceCondition{
						{
							Type:   types.ResourceAnimals,
							Amount: tt.storage,
							Target: types.TargetSelfCard,
						},
					},
				},
			}

			ctx := context.Background()
			err := validator.ValidateActionInputs(ctx, "game1", "player1", action, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateActionInputs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateActionInputs_PlayerNotFound(t *testing.T) {
	mockPlayerRepo := mocks.NewMockPlayerRepository()
	validator := execute_card_action.NewValidator(mockPlayerRepo)

	action := &types.PlayerAction{
		CardID: "card1",
		Behavior: types.CardBehavior{
			Inputs: []types.ResourceCondition{},
		},
	}

	ctx := context.Background()
	err := validator.ValidateActionInputs(ctx, "game1", "nonexistent", action, nil)

	if err == nil {
		t.Errorf("ValidateActionInputs() should error when player not found")
	}
}
