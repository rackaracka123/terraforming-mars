package actions_test

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service/actions/play_card"
	"testing"
)

func TestPlayCardHandler_Handle(t *testing.T) {
	handler := &play_card.PlayCardHandler{}
	
	// Get available cards to use in tests
	availableCards := model.GetStartingCards()
	cardMap := make(map[string]model.Card)
	for _, card := range availableCards {
		cardMap[card.ID] = card
	}

	tests := []struct {
		name           string
		gamePhase      model.GamePhase
		playerCredits  int
		playerCards    []string
		selectedCard   string
		wantErr        bool
		wantCredits    int
		wantCardCount  int
		wantPlayedCount int
	}{
		{
			name:           "valid card play - early settlement",
			gamePhase:      model.GamePhaseAction,
			playerCredits:  10,
			playerCards:    []string{"early-settlement"},
			selectedCard:   "early-settlement",
			wantErr:        false,
			wantCredits:    2, // 10 - 8 = 2
			wantCardCount:  0, // card moved from hand
			wantPlayedCount: 1, // card added to played cards
		},
		{
			name:           "valid card play - power plant",
			gamePhase:      model.GamePhaseAction,
			playerCredits:  15,
			playerCards:    []string{"power-plant", "heat-generators"},
			selectedCard:   "power-plant",
			wantErr:        false,
			wantCredits:    9, // 15 - 6 = 9
			wantCardCount:  1, // one card remains in hand
			wantPlayedCount: 1, // card added to played cards
		},
		{
			name:           "insufficient credits",
			gamePhase:      model.GamePhaseAction,
			playerCredits:  5,
			playerCards:    []string{"early-settlement"},
			selectedCard:   "early-settlement",
			wantErr:        true,
			wantCredits:    5, // unchanged
			wantCardCount:  1, // unchanged
			wantPlayedCount: 0, // unchanged
		},
		{
			name:           "wrong game phase",
			gamePhase:      model.GamePhaseCorporationSelection,
			playerCredits:  10,
			playerCards:    []string{"early-settlement"},
			selectedCard:   "early-settlement",
			wantErr:        true,
			wantCredits:    10, // unchanged
			wantCardCount:  1,  // unchanged
			wantPlayedCount: 0,  // unchanged
		},
		{
			name:           "card not in hand",
			gamePhase:      model.GamePhaseAction,
			playerCredits:  10,
			playerCards:    []string{"power-plant"},
			selectedCard:   "early-settlement",
			wantErr:        true,
			wantCredits:    10, // unchanged
			wantCardCount:  1,  // unchanged
			wantPlayedCount: 0,  // unchanged
		},
		{
			name:           "invalid card ID",
			gamePhase:      model.GamePhaseAction,
			playerCredits:  10,
			playerCards:    []string{"non-existent-card"},
			selectedCard:   "non-existent-card",
			wantErr:        true,
			wantCredits:    10, // unchanged
			wantCardCount:  1,  // unchanged
			wantPlayedCount: 0,  // unchanged
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
						Production: model.Production{
							Credits: 1,
						},
						Cards:       tt.playerCards,
						PlayedCards: []string{},
					},
				},
				GlobalParameters: model.GlobalParameters{
					Temperature: -30,
					Oxygen:      0,
					Oceans:      0,
				},
			}

			// Create action request
			actionRequest := dto.ActionPlayCardRequest{
				Type:   dto.ActionTypePlayCard,
				CardID: tt.selectedCard,
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

			// Check player card count in hand
			if len(game.Players[0].Cards) != tt.wantCardCount {
				t.Errorf("Expected player to have %d cards in hand, got %d", tt.wantCardCount, len(game.Players[0].Cards))
			}

			// Check player played card count
			if len(game.Players[0].PlayedCards) != tt.wantPlayedCount {
				t.Errorf("Expected player to have %d played cards, got %d", tt.wantPlayedCount, len(game.Players[0].PlayedCards))
			}

			// If no error and played successfully, check that the played card is correct
			if !tt.wantErr && tt.wantPlayedCount > 0 {
				if game.Players[0].PlayedCards[0] != tt.selectedCard {
					t.Errorf("Expected played card to be %s, got %s", tt.selectedCard, game.Players[0].PlayedCards[0])
				}
			}
		})
	}
}

func TestPlayCardHandler_CardEffects(t *testing.T) {
	handler := &play_card.PlayCardHandler{}
	
	// Get available cards from domain
	availableCards := model.GetStartingCards()
	implementedCards := []string{"early-settlement", "power-plant", "heat-generators", "mining-operation", "space-mirrors", "water-import", "nitrogen-plants", "atmospheric-processors"}
	
	// Create a map of implemented cards with their expected effects
	cardEffects := map[string]struct {
		expectedProduction model.Production
		expectedResources  func(startCredits, cardCost int) model.Resources
		expectedGlobal     model.GlobalParameters
	}{
		"early-settlement": {
			expectedProduction: model.Production{Credits: 2}, // base 1 + 1 from card
			expectedResources: func(startCredits, cardCost int) model.Resources {
				return model.Resources{Credits: startCredits - cardCost}
			},
			expectedGlobal: model.GlobalParameters{Temperature: -30, Oxygen: 0, Oceans: 0},
		},
		"power-plant": {
			expectedProduction: model.Production{Credits: 1, Energy: 1}, // +1 energy production
			expectedResources: func(startCredits, cardCost int) model.Resources {
				return model.Resources{Credits: startCredits - cardCost}
			},
			expectedGlobal: model.GlobalParameters{Temperature: -30, Oxygen: 0, Oceans: 0},
		},
		"heat-generators": {
			expectedProduction: model.Production{Credits: 1, Heat: 1}, // +1 heat production
			expectedResources: func(startCredits, cardCost int) model.Resources {
				return model.Resources{Credits: startCredits - cardCost}
			},
			expectedGlobal: model.GlobalParameters{Temperature: -30, Oxygen: 0, Oceans: 0},
		},
		"mining-operation": {
			expectedProduction: model.Production{Credits: 1}, // no production change
			expectedResources: func(startCredits, cardCost int) model.Resources {
				return model.Resources{Credits: startCredits - cardCost, Steel: 2} // +2 steel
			},
			expectedGlobal: model.GlobalParameters{Temperature: -30, Oxygen: 0, Oceans: 0},
		},
		"space-mirrors": {
			expectedProduction: model.Production{Credits: 1}, // no immediate effect
			expectedResources: func(startCredits, cardCost int) model.Resources {
				return model.Resources{Credits: startCredits - cardCost}
			},
			expectedGlobal: model.GlobalParameters{Temperature: -30, Oxygen: 0, Oceans: 0},
		},
		"water-import": {
			expectedProduction: model.Production{Credits: 1}, // no production change
			expectedResources: func(startCredits, cardCost int) model.Resources {
				return model.Resources{Credits: startCredits - cardCost}
			},
			expectedGlobal: model.GlobalParameters{Temperature: -30, Oxygen: 0, Oceans: 1}, // +1 ocean
		},
		"nitrogen-plants": {
			expectedProduction: model.Production{Credits: 1, Plants: 1}, // +1 plant production
			expectedResources: func(startCredits, cardCost int) model.Resources {
				return model.Resources{Credits: startCredits - cardCost}
			},
			expectedGlobal: model.GlobalParameters{Temperature: -30, Oxygen: 0, Oceans: 0},
		},
		"atmospheric-processors": {
			expectedProduction: model.Production{Credits: 1}, // no production change
			expectedResources: func(startCredits, cardCost int) model.Resources {
				return model.Resources{Credits: startCredits - cardCost}
			},
			expectedGlobal: model.GlobalParameters{Temperature: -30, Oxygen: 1, Oceans: 0}, // +1 oxygen
		},
	}

	for _, cardID := range implementedCards {
		// Find the card in available cards
		var card model.Card
		found := false
		for _, c := range availableCards {
			if c.ID == cardID {
				card = c
				found = true
				break
			}
		}
		
		if !found {
			t.Fatalf("Card %s not found in available cards", cardID)
		}
		
		expectedEffect, hasEffect := cardEffects[cardID]
		if !hasEffect {
			t.Fatalf("No expected effect defined for card %s", cardID)
		}

		t.Run("card_"+cardID+"_effects", func(t *testing.T) {
			startCredits := 20 // Enough for any card
			
			// Create game with action phase and player with enough credits
			game := &model.Game{
				ID:           "test-game",
				CurrentPhase: model.GamePhaseAction,
				Players: []model.Player{
					{
						ID:   "player-1",
						Name: "Test Player",
						Resources: model.Resources{
							Credits: startCredits,
						},
						Production: model.Production{
							Credits: 1, // Base production
						},
						Cards:       []string{cardID},
						PlayedCards: []string{},
					},
				},
				GlobalParameters: model.GlobalParameters{
					Temperature: -30,
					Oxygen:      0,
					Oceans:      0,
				},
			}

			// Create action request
			actionRequest := dto.ActionPlayCardRequest{
				Type:   dto.ActionTypePlayCard,
				CardID: cardID,
			}

			// Apply the action
			err := handler.Handle(game, &game.Players[0], actionRequest)
			if err != nil {
				t.Fatalf("Handle() unexpected error = %v", err)
			}

			// Check resources
			expectedResources := expectedEffect.expectedResources(startCredits, card.Cost)
			if game.Players[0].Resources != expectedResources {
				t.Errorf("Card %s: Expected resources %+v, got %+v", card.Name, expectedResources, game.Players[0].Resources)
			}

			// Check production
			if game.Players[0].Production != expectedEffect.expectedProduction {
				t.Errorf("Card %s: Expected production %+v, got %+v", card.Name, expectedEffect.expectedProduction, game.Players[0].Production)
			}

			// Check global parameters
			if game.GlobalParameters != expectedEffect.expectedGlobal {
				t.Errorf("Card %s: Expected global parameters %+v, got %+v", card.Name, expectedEffect.expectedGlobal, game.GlobalParameters)
			}

			// Check that card was moved to played cards
			if len(game.Players[0].PlayedCards) != 1 || game.Players[0].PlayedCards[0] != cardID {
				t.Errorf("Card %s: Expected card to be in played cards, got %v", card.Name, game.Players[0].PlayedCards)
			}

			// Check that card was removed from hand
			if len(game.Players[0].Cards) != 0 {
				t.Errorf("Card %s: Expected card to be removed from hand, but hand still has %v", card.Name, game.Players[0].Cards)
			}
		})
	}
}

func TestPlayCardHandler_UnimplementedCards(t *testing.T) {
	handler := &play_card.PlayCardHandler{}
	
	// Get all available cards and find unimplemented ones
	availableCards := model.GetStartingCards()
	implementedCards := map[string]bool{
		"early-settlement": true,
		"power-plant": true,
		"heat-generators": true,
		"mining-operation": true,
		"space-mirrors": true,
		"water-import": true,
		"nitrogen-plants": true,
		"atmospheric-processors": true,
	}
	
	var unimplementedCards []model.Card
	for _, card := range availableCards {
		if !implementedCards[card.ID] {
			unimplementedCards = append(unimplementedCards, card)
		}
	}

	for _, card := range unimplementedCards {
		t.Run("card_"+card.ID+"_not_implemented", func(t *testing.T) {
			game := &model.Game{
				ID:           "test-game",
				CurrentPhase: model.GamePhaseAction,
				Players: []model.Player{
					{
						ID:   "player-1",
						Name: "Test Player",
						Resources: model.Resources{
							Credits: 20,
						},
						Cards: []string{card.ID},
					},
				},
			}

			actionRequest := dto.ActionPlayCardRequest{
				Type:   dto.ActionTypePlayCard,
				CardID: card.ID,
			}

			err := handler.Handle(game, &game.Players[0], actionRequest)
			if err == nil {
				t.Errorf("Expected error for unimplemented card %s, but got none", card.Name)
			}

			if err != nil && !containsSubstring(err.Error(), "not yet implemented") {
				t.Errorf("Expected 'not yet implemented' error for card %s, got: %v", card.Name, err)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}