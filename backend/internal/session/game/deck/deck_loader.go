package deck

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

const cardDataPath = "assets/terraforming_mars_cards.json"

// LoadCardsFromJSON loads all card definitions from the JSON file
func LoadCardsFromJSON(ctx context.Context) (*CardDefinitions, error) {
	log := logger.Get()

	// Read the JSON file
	data, err := os.ReadFile(cardDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read card data file: %w", err)
	}

	// Parse JSON into card array
	var cards []types.Card
	if err := json.Unmarshal(data, &cards); err != nil {
		return nil, fmt.Errorf("failed to parse card data: %w", err)
	}

	// Organize cards by type
	defs := &CardDefinitions{
		AllCards:         make(map[string]types.Card),
		ProjectCards:     make([]types.Card, 0),
		CorporationCards: make([]types.Card, 0),
		PreludeCards:     make([]types.Card, 0),
		StartingCards:    make([]types.Card, 0),
	}

	// Categorize cards
	for _, card := range cards {
		// Add to all cards map
		defs.AllCards[card.ID] = card

		// Categorize by type
		switch card.Type {
		case types.CardTypeCorporation:
			defs.CorporationCards = append(defs.CorporationCards, card)
		case types.CardTypePrelude:
			defs.PreludeCards = append(defs.PreludeCards, card)
		case types.CardTypeAutomated, types.CardTypeActive, types.CardTypeEvent:
			defs.ProjectCards = append(defs.ProjectCards, card)

			// Check if it's a starting card (cost <= 10 and in base-game pack)
			if card.Cost <= 10 && card.Pack == "base-game" {
				defs.StartingCards = append(defs.StartingCards, card)
			}
		}
	}

	log.Info("📚 Card definitions loaded successfully",
		zap.Int("total_cards", len(defs.AllCards)),
		zap.Int("project_cards", len(defs.ProjectCards)),
		zap.Int("corporation_cards", len(defs.CorporationCards)),
		zap.Int("prelude_cards", len(defs.PreludeCards)),
		zap.Int("starting_cards", len(defs.StartingCards)))

	return defs, nil
}

// extractCardIDs extracts card IDs from a slice of cards
func extractCardIDs(cards []types.Card) []string {
	ids := make([]string, len(cards))
	for i, card := range cards {
		ids[i] = card.ID
	}
	return ids
}
