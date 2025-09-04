package cards

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/model"
)



// ValidateCardRequirements checks if a card's requirements are met using PlayerService
func ValidateCardRequirements(ctx context.Context, game *model.Game, playerID string, playerService PlayerService, requirements model.CardRequirements) error {
	// Check temperature requirements
	if requirements.MinTemperature != nil && game.GlobalParameters.Temperature < *requirements.MinTemperature {
		return fmt.Errorf("temperature too low: need %d°C, current %d°C", *requirements.MinTemperature, game.GlobalParameters.Temperature)
	}
	
	if requirements.MaxTemperature != nil && game.GlobalParameters.Temperature > *requirements.MaxTemperature {
		return fmt.Errorf("temperature too high: max %d°C, current %d°C", *requirements.MaxTemperature, game.GlobalParameters.Temperature)
	}
	
	// Check oxygen requirements
	if requirements.MinOxygen != nil && game.GlobalParameters.Oxygen < *requirements.MinOxygen {
		return fmt.Errorf("oxygen too low: need %d%%, current %d%%", *requirements.MinOxygen, game.GlobalParameters.Oxygen)
	}
	
	if requirements.MaxOxygen != nil && game.GlobalParameters.Oxygen > *requirements.MaxOxygen {
		return fmt.Errorf("oxygen too high: max %d%%, current %d%%", *requirements.MaxOxygen, game.GlobalParameters.Oxygen)
	}
	
	// Check ocean requirements
	if requirements.MinOceans != nil && game.GlobalParameters.Oceans < *requirements.MinOceans {
		return fmt.Errorf("not enough oceans: need %d, current %d", *requirements.MinOceans, game.GlobalParameters.Oceans)
	}
	
	if requirements.MaxOceans != nil && game.GlobalParameters.Oceans > *requirements.MaxOceans {
		return fmt.Errorf("too many oceans: max %d, current %d", *requirements.MaxOceans, game.GlobalParameters.Oceans)
	}
	
	// Check tag requirements
	if len(requirements.RequiredTags) > 0 {
		playerTags, err := GetPlayerTags(ctx, game, playerID, playerService)
		if err != nil {
			return fmt.Errorf("failed to get player tags: %w", err)
		}
		for _, requiredTag := range requirements.RequiredTags {
			if !hasTag(playerTags, requiredTag) {
				return fmt.Errorf("missing required tag: %s", string(requiredTag))
			}
		}
	}
	
	// Check production requirements
	if requirements.RequiredProduction != nil {
		if err := playerService.ValidateProductionRequirement(ctx, game.ID, playerID, *requirements.RequiredProduction); err != nil {
			return fmt.Errorf("production requirement not met: %w", err)
		}
	}
	
	return nil
}


// GetPlayerTags returns all tags from cards the player has played
func GetPlayerTags(ctx context.Context, game *model.Game, playerID string, playerService PlayerService) ([]model.CardTag, error) {
	// Get the player from the game
	player, found := game.GetPlayer(playerID)
	if !found {
		return nil, fmt.Errorf("player not found in game")
	}
	tagMap := make(map[model.CardTag]bool)
	cards := model.GetStartingCards()
	
	// Create a map for quick card lookup
	cardMap := make(map[string]model.Card)
	for _, card := range cards {
		cardMap[card.ID] = card
	}
	
	// Collect unique tags from all played cards
	for _, playedCardID := range player.PlayedCards {
		if card, exists := cardMap[playedCardID]; exists {
			for _, tag := range card.Tags {
				tagMap[tag] = true
			}
		}
	}
	
	// Convert map to slice
	tags := make([]model.CardTag, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	
	return tags, nil
}

// hasTag checks if a tag exists in a slice of tags
func hasTag(tags []model.CardTag, tag model.CardTag) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}