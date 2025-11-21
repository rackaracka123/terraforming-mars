package cards

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/session/card"

	"github.com/stretchr/testify/assert"
)

func TestRequirementsValidator_ProductionRequirements(t *testing.T) {
	// Setup
	cardRepo := repository.NewCardRepository()
	validator := card.NewRequirementsValidator(cardRepo)
	ctx := context.Background()

	// Create test player with specific production levels
	player := model.Player{
		ID: "player1",
		Production: model.Production{
			Credits:  3,
			Steel:    2,
			Titanium: 0,
			Plants:   1,
			Energy:   4,
			Heat:     2,
		},
	}

	game := model.Game{}

	tests := []struct {
		name          string
		requirement   model.Requirement
		expectedError bool
		errorContains string
	}{
		{
			name: "Should pass when player has sufficient credit production",
			requirement: model.Requirement{
				Type:     model.RequirementProduction,
				Min:      &[]int{2}[0],
				Resource: &[]model.ResourceType{model.ResourceCredits}[0],
			},
			expectedError: false,
		},
		{
			name: "Should fail when player lacks sufficient steel production",
			requirement: model.Requirement{
				Type:     model.RequirementProduction,
				Min:      &[]int{3}[0],
				Resource: &[]model.ResourceType{model.ResourceSteel}[0],
			},
			expectedError: true,
			errorContains: "insufficient steel production",
		},
		{
			name: "Should fail when player has zero titanium production",
			requirement: model.Requirement{
				Type:     model.RequirementProduction,
				Min:      &[]int{1}[0],
				Resource: &[]model.ResourceType{model.ResourceTitanium}[0],
			},
			expectedError: true,
			errorContains: "insufficient titanium production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCardRequirements(ctx, "gameID", "playerID", &model.Card{
				Requirements: []model.Requirement{tt.requirement},
			}, &game, &player)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequirementsValidator_ResourceRequirements(t *testing.T) {
	// Setup
	cardRepo := repository.NewCardRepository()
	validator := card.NewRequirementsValidator(cardRepo)
	ctx := context.Background()

	// Create test player with specific resource levels
	player := model.Player{
		ID: "player1",
		Resources: model.Resources{
			Credits:  25,
			Steel:    8,
			Titanium: 3,
			Plants:   12,
			Energy:   0,
			Heat:     15,
		},
	}

	game := model.Game{}

	tests := []struct {
		name          string
		requirement   model.Requirement
		expectedError bool
		errorContains string
	}{
		{
			name: "Should pass when player has sufficient credits",
			requirement: model.Requirement{
				Type:     model.RequirementResource,
				Min:      &[]int{20}[0],
				Resource: &[]model.ResourceType{model.ResourceCredits}[0],
			},
			expectedError: false,
		},
		{
			name: "Should fail when player lacks sufficient plants",
			requirement: model.Requirement{
				Type:     model.RequirementResource,
				Min:      &[]int{15}[0],
				Resource: &[]model.ResourceType{model.ResourcePlants}[0],
			},
			expectedError: true,
			errorContains: "insufficient plants",
		},
		{
			name: "Should fail when player has zero energy",
			requirement: model.Requirement{
				Type:     model.RequirementResource,
				Min:      &[]int{1}[0],
				Resource: &[]model.ResourceType{model.ResourceEnergy}[0],
			},
			expectedError: true,
			errorContains: "insufficient energy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCardRequirements(ctx, "gameID", "playerID", &model.Card{
				Requirements: []model.Requirement{tt.requirement},
			}, &game, &player)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequirementsValidator_AffordabilityValidation(t *testing.T) {
	// Setup
	cardRepo := repository.NewCardRepository()
	validator := card.NewRequirementsValidator(cardRepo)
	ctx := context.Background()

	// Create test player with limited resources
	player := model.Player{
		ID: "player1",
		Resources: model.Resources{
			Credits:  30,
			Steel:    5,
			Titanium: 2,
			Plants:   8,
			Energy:   3,
			Heat:     4,
		},
		Production: model.Production{
			Credits:  2,
			Steel:    1,
			Titanium: 1,
			Plants:   1,
			Energy:   3,
			Heat:     1,
		},
	}

	tests := []struct {
		name          string
		card          model.Card
		expectedError bool
		errorContains string
	}{
		{
			name: "Should pass for affordable card",
			card: model.Card{
				ID:   "affordable-card",
				Cost: 20,
				Behaviors: []model.CardBehavior{
					{
						Triggers: []model.Trigger{{Type: model.ResourceTriggerAuto}},
						Inputs: []model.ResourceCondition{
							{Type: model.ResourceSteel, Amount: 3, Target: model.TargetSelfPlayer},
						},
						Outputs: []model.ResourceCondition{
							{Type: model.ResourceCreditsProduction, Amount: 2, Target: model.TargetSelfPlayer},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "Should fail for card requiring too many credits",
			card: model.Card{
				ID:        "expensive-card",
				Cost:      40, // Player only has 30
				Behaviors: []model.CardBehavior{},
			},
			expectedError: true,
			errorContains: "insufficient credits",
		},
		{
			name: "Should fail for card requiring too much steel for behaviors",
			card: model.Card{
				ID:   "steel-heavy-card",
				Cost: 15,
				Behaviors: []model.CardBehavior{
					{
						Triggers: []model.Trigger{{Type: model.ResourceTriggerAuto}},
						Inputs: []model.ResourceCondition{
							{Type: model.ResourceSteel, Amount: 8, Target: model.TargetSelfPlayer}, // Player only has 5
						},
						Outputs: []model.ResourceCondition{
							{Type: model.ResourceCreditsProduction, Amount: 3, Target: model.TargetSelfPlayer},
						},
					},
				},
			},
			expectedError: true,
			errorContains: "insufficient steel for card effects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a payment that covers the card cost with credits
			payment := &model.CardPayment{
				Credits:  tt.card.Cost,
				Steel:    0,
				Titanium: 0,
			}
			err := validator.ValidateCardAffordability(ctx, "gameID", "playerID", &tt.card, &player, payment, nil)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
