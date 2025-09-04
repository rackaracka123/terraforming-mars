package actions_test

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service/actions/select_starting_card"
	"testing"
)

func TestSelectStartingCardsHandler_Handle(t *testing.T) {
	handler := &select_starting_card.SelectStartingCardsHandler{}

	tests := []struct {
		name           string
		gamePhase      model.GamePhase
		playerCredits  int
		playerCards    []string
		selectedCards  []string
		wantErr        bool
		wantCredits    int
		wantCardCount  int
	}{
		{
			name:           "valid selection - single card costs 3 MC",
			gamePhase:      model.GamePhaseStartingCardSelection,
			playerCredits:  10,
			playerCards:    []string{},
			selectedCards:  []string{"investment"},
			wantErr:        false,
			wantCredits:    7, // 10 - 3 = 7
			wantCardCount:  1,
		},
		{
			name:           "valid selection - multiple cards cost 3 MC each",
			gamePhase:      model.GamePhaseStartingCardSelection,
			playerCredits:  20,
			playerCards:    []string{},
			selectedCards:  []string{"investment", "early-settlement", "research-grant"},
			wantErr:        false,
			wantCredits:    11, // 20 - (3 * 3) = 11
			wantCardCount:  3,
		},
		{
			name:           "insufficient credits",
			gamePhase:      model.GamePhaseStartingCardSelection,
			playerCredits:  2,
			playerCards:    []string{},
			selectedCards:  []string{"investment"},
			wantErr:        true,
			wantCredits:    2, // unchanged
			wantCardCount:  0,
		},
		{
			name:           "wrong game phase",
			gamePhase:      model.GamePhaseCorporationSelection,
			playerCredits:  10,
			playerCards:    []string{},
			selectedCards:  []string{"investment"},
			wantErr:        true,
			wantCredits:    10, // unchanged
			wantCardCount:  0,
		},
		{
			name:           "cards already selected",
			gamePhase:      model.GamePhaseStartingCardSelection,
			playerCredits:  10,
			playerCards:    []string{"some-card"},
			selectedCards:  []string{"investment"},
			wantErr:        true,
			wantCredits:    10, // unchanged
			wantCardCount:  1,  // existing card remains
		},
		{
			name:           "invalid card ID",
			gamePhase:      model.GamePhaseStartingCardSelection,
			playerCredits:  10,
			playerCards:    []string{},
			selectedCards:  []string{"non-existent-card"},
			wantErr:        true,
			wantCredits:    10, // unchanged
			wantCardCount:  0,
		},
		{
			name:           "expensive card still costs 3 MC (not original cost)",
			gamePhase:      model.GamePhaseStartingCardSelection,
			playerCredits:  15,
			playerCards:    []string{},
			selectedCards:  []string{"water-import"}, // This card costs 12 MC in game, but should cost 3 MC for starting selection
			wantErr:        false,
			wantCredits:    12, // 15 - 3 = 12 (NOT 15 - 12 = 3)
			wantCardCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create game with the specified phase
			game := &model.Game{
				ID:           "test-game",
				CurrentPhase: tt.gamePhase,
				Players: []model.Player{
					{
						ID:   "player-1",
						Name: "Test Player",
						Resources: model.Resources{
							Credits: tt.playerCredits,
						},
						Cards: tt.playerCards,
					},
				},
			}

			// Create action request
			actionRequest := dto.ActionSelectStartingCardRequest{
				Type:    dto.ActionTypeSelectStartingCard,
				CardIDs: tt.selectedCards,
			}

			// Apply the action
			err := handler.Handle(game, &game.Players[0], actionRequest)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check player credits after action
			if game.Players[0].Resources.Credits != tt.wantCredits {
				t.Errorf("Expected player credits to be %d, got %d", tt.wantCredits, game.Players[0].Resources.Credits)
			}

			// Check player card count
			if len(game.Players[0].Cards) != tt.wantCardCount {
				t.Errorf("Expected player to have %d cards, got %d", tt.wantCardCount, len(game.Players[0].Cards))
			}

			// If no error, check that selected cards are added correctly
			if !tt.wantErr && len(tt.selectedCards) > 0 {
				for i, expectedCardID := range tt.selectedCards {
					if i >= len(game.Players[0].Cards) || game.Players[0].Cards[i] != expectedCardID {
						t.Errorf("Expected card %s at position %d, got %v", expectedCardID, i, game.Players[0].Cards)
					}
				}
			}
		})
	}
}

func TestSelectStartingCardsHandler_AllPlayersSelected(t *testing.T) {
	handler := &select_starting_card.SelectStartingCardsHandler{}

	// Create game with 2 players, one has already selected cards
	game := &model.Game{
		ID:           "test-game",
		CurrentPhase: model.GamePhaseStartingCardSelection,
		Players: []model.Player{
			{
				ID:   "player-1",
				Name: "Player 1",
				Resources: model.Resources{
					Credits: 10,
				},
				Cards: []string{"investment"}, // already selected
			},
			{
				ID:   "player-2",
				Name: "Player 2",
				Resources: model.Resources{
					Credits: 10,
				},
				Cards: []string{}, // hasn't selected yet
			},
		},
	}

	// Player 2 selects cards
	actionRequest := dto.ActionSelectStartingCardRequest{
		Type:    dto.ActionTypeSelectStartingCard,
		CardIDs: []string{"research-grant"},
	}

	err := handler.Handle(game, &game.Players[1], actionRequest)
	if err != nil {
		t.Fatalf("Handle() unexpected error = %v", err)
	}

	// Check that game phase advanced to corporation selection
	if game.CurrentPhase != model.GamePhaseCorporationSelection {
		t.Errorf("Expected game phase to advance to %s, got %s", model.GamePhaseCorporationSelection, game.CurrentPhase)
	}

	// Check that player 2 paid correctly
	if game.Players[1].Resources.Credits != 7 {
		t.Errorf("Expected player 2 credits to be 7, got %d", game.Players[1].Resources.Credits)
	}

	// Check that player 2's cards were set correctly
	if len(game.Players[1].Cards) != 1 || game.Players[1].Cards[0] != "research-grant" {
		t.Errorf("Expected player 2 cards to be [research-grant], got %v", game.Players[1].Cards)
	}
}

func TestSelectStartingCards_PaymentLogic(t *testing.T) {
	// This test specifically verifies that all cards cost 3 MC regardless of their actual cost
	handler := &select_starting_card.SelectStartingCardsHandler{}

	// Get all available starting cards to test
	availableCards := model.GetStartingCards()
	if len(availableCards) == 0 {
		t.Fatal("Expected at least one starting card")
	}

	for _, card := range availableCards {
		t.Run("card_"+card.ID+"_costs_3_MC", func(t *testing.T) {
			game := &model.Game{
				ID:           "test-game",
				CurrentPhase: model.GamePhaseStartingCardSelection,
				Players: []model.Player{
					{
						ID:   "player-1",
						Name: "Test Player",
						Resources: model.Resources{
							Credits: 10,
						},
						Cards: []string{},
					},
				},
			}

			actionRequest := dto.ActionSelectStartingCardRequest{
				Type:    dto.ActionTypeSelectStartingCard,
				CardIDs: []string{card.ID},
			}

			err := handler.Handle(game, &game.Players[0], actionRequest)
			if err != nil {
				t.Fatalf("Handle() unexpected error for card %s: %v", card.Name, err)
			}

			// Every card should cost exactly 3 MC, regardless of its Cost field
			expectedCredits := 7 // 10 - 3
			if game.Players[0].Resources.Credits != expectedCredits {
				t.Errorf("Card %s (original cost %d MC): expected remaining credits %d, got %d", 
					card.Name, card.Cost, expectedCredits, game.Players[0].Resources.Credits)
			}
		})
	}
}