package deck

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game/card"

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
	var cards []card.Card
	if err := json.Unmarshal(data, &cards); err != nil {
		return nil, fmt.Errorf("failed to parse card data: %w", err)
	}

	// Organize cards by type
	defs := &CardDefinitions{
		AllCards:         make(map[string]card.Card),
		ProjectCards:     make([]card.Card, 0),
		CorporationCards: make([]card.Card, 0),
		PreludeCards:     make([]card.Card, 0),
		StartingCards:    make([]card.Card, 0),
	}

	// Categorize cards
	for _, c := range cards {
		// Add to all cards map
		defs.AllCards[c.ID] = c

		// Categorize by type
		switch c.Type {
		case card.CardTypeCorporation:
			defs.CorporationCards = append(defs.CorporationCards, c)
		case card.CardTypePrelude:
			defs.PreludeCards = append(defs.PreludeCards, c)
		case card.CardTypeAutomated, card.CardTypeActive, card.CardTypeEvent:
			defs.ProjectCards = append(defs.ProjectCards, c)

			// Check if it's a starting card (cost <= 10 and in base-game pack)
			if c.Cost <= 10 && c.Pack == "base-game" {
				defs.StartingCards = append(defs.StartingCards, c)
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
func extractCardIDs(cards []card.Card) []string {
	ids := make([]string, len(cards))
	for i, c := range cards {
		ids[i] = c.ID
	}
	return ids
}
