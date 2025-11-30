package cards

import (
	"encoding/json"
	"fmt"
	"os"

	"terraforming-mars-backend/internal/game/cards"
)

// LoadCardsFromJSON loads cards from a JSON file
func LoadCardsFromJSON(filepath string) ([]cards.Card, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read card file: %w", err)
	}

	var cards []cards.Card
	if err := json.Unmarshal(data, &cards); err != nil {
		return nil, fmt.Errorf("failed to parse card JSON: %w", err)
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("no cards found in file: %s", filepath)
	}

	return cards, nil
}
