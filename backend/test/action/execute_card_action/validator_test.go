package execute_card_action

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action/execute_card_action"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/types"
)

// testSessionFactory is a minimal SessionFactory implementation for testing
type testSessionFactory struct {
	sessions map[string]*session.Session
}

func newTestSessionFactory() *testSessionFactory {
	return &testSessionFactory{
		sessions: make(map[string]*session.Session),
	}
}

func (f *testSessionFactory) Get(gameID string) *session.Session {
	return f.sessions[gameID]
}

func (f *testSessionFactory) GetOrCreate(gameID string) *session.Session {
	if sess, exists := f.sessions[gameID]; exists {
		return sess
	}
	sess := session.NewSession(gameID, events.NewEventBus())
	sess.Game = types.NewGame(gameID, types.GameSettings{})
	f.sessions[gameID] = sess
	return sess
}

func (f *testSessionFactory) Remove(gameID string) {
	delete(f.sessions, gameID)
}

func (f *testSessionFactory) WireGameRepositories(g *types.Game) {
	// No-op for tests
}

func TestValidator_ValidateActionInputs_SufficientResources(t *testing.T) {
	// Setup test session factory
	sessionFactory := newTestSessionFactory()
	validator := execute_card_action.NewValidator(sessionFactory)

	// Create session and game
	sess := sessionFactory.GetOrCreate("game1")

	// Create player with resources
	playerData := &types.Player{
		ID: "player1",
		Resources: types.Resources{
			Credits:  20,
			Steel:    5,
			Titanium: 3,
			Plants:   10,
			Energy:   5,
			Heat:     8,
		},
		GameID: "game1",
	}
	sess.AddPlayer(playerData)

	// Create action with inputs
	action := &types.CardAction{
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
	sessionFactory := newTestSessionFactory()
	validator := execute_card_action.NewValidator(sessionFactory)

	// Create session and game
	sess := sessionFactory.GetOrCreate("game1")

	// Create player with limited resources
	playerData := &types.Player{
		ID: "player1",
		Resources: types.Resources{
			Credits: 5, // Not enough
			Steel:   1, // Not enough
		},
		GameID: "game1",
	}
	sess.AddPlayer(playerData)

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
			action := &types.CardAction{
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
	sessionFactory := newTestSessionFactory()
	validator := execute_card_action.NewValidator(sessionFactory)

	// Create session and game
	sess := sessionFactory.GetOrCreate("game1")

	// Create player with resources
	playerData := &types.Player{
		ID: "player1",
		Resources: types.Resources{
			Credits:  20,
			Titanium: 5,
		},
		GameID: "game1",
	}
	sess.AddPlayer(playerData)

	// Create action with base inputs and choice-specific inputs
	choiceIndex := 0
	action := &types.CardAction{
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
	sessionFactory := newTestSessionFactory()
	validator := execute_card_action.NewValidator(sessionFactory)

	// Create session and game
	sess := sessionFactory.GetOrCreate("game1")

	// Create player with card storage
	playerData := &types.Player{
		ID: "player1",
		ResourceStorage: map[string]int{
			"card1": 5, // 5 animals on this card
		},
		GameID: "game1",
	}
	sess.AddPlayer(playerData)

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
			action := &types.CardAction{
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
	sessionFactory := newTestSessionFactory()
	validator := execute_card_action.NewValidator(sessionFactory)

	action := &types.CardAction{
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
