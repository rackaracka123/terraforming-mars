package cards

import (
	"fmt"
	"terraforming-mars-backend/internal/model"
)


// ValidateCardRequirements checks if a card's requirements are met
func ValidateCardRequirements(game *model.Game, player *model.Player, requirements model.CardRequirements) error {
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
		playerTags := GetPlayerTags(player)
		for _, requiredTag := range requirements.RequiredTags {
			if !hasTag(playerTags, requiredTag) {
				return fmt.Errorf("missing required tag: %s", string(requiredTag))
			}
		}
	}
	
	// Check production requirements
	if requirements.RequiredProduction != nil {
		if err := ValidateProductionRequirement(player, *requirements.RequiredProduction); err != nil {
			return fmt.Errorf("production requirement not met: %w", err)
		}
	}
	
	return nil
}

// ValidateResourceCost checks if a player has enough resources to pay a cost
func ValidateResourceCost(player *model.Player, cost model.ResourceSet) error {
	if player.Resources.Credits < cost.Credits {
		return fmt.Errorf("insufficient credits: need %d, have %d", cost.Credits, player.Resources.Credits)
	}
	if player.Resources.Steel < cost.Steel {
		return fmt.Errorf("insufficient steel: need %d, have %d", cost.Steel, player.Resources.Steel)
	}
	if player.Resources.Titanium < cost.Titanium {
		return fmt.Errorf("insufficient titanium: need %d, have %d", cost.Titanium, player.Resources.Titanium)
	}
	if player.Resources.Plants < cost.Plants {
		return fmt.Errorf("insufficient plants: need %d, have %d", cost.Plants, player.Resources.Plants)
	}
	if player.Resources.Energy < cost.Energy {
		return fmt.Errorf("insufficient energy: need %d, have %d", cost.Energy, player.Resources.Energy)
	}
	if player.Resources.Heat < cost.Heat {
		return fmt.Errorf("insufficient heat: need %d, have %d", cost.Heat, player.Resources.Heat)
	}
	return nil
}

// ValidateProductionRequirement checks if a player has minimum production levels
func ValidateProductionRequirement(player *model.Player, requirement model.ResourceSet) error {
	if player.Production.Credits < requirement.Credits {
		return fmt.Errorf("insufficient credit production: need %d, have %d", requirement.Credits, player.Production.Credits)
	}
	if player.Production.Steel < requirement.Steel {
		return fmt.Errorf("insufficient steel production: need %d, have %d", requirement.Steel, player.Production.Steel)
	}
	if player.Production.Titanium < requirement.Titanium {
		return fmt.Errorf("insufficient titanium production: need %d, have %d", requirement.Titanium, player.Production.Titanium)
	}
	if player.Production.Plants < requirement.Plants {
		return fmt.Errorf("insufficient plant production: need %d, have %d", requirement.Plants, player.Production.Plants)
	}
	if player.Production.Energy < requirement.Energy {
		return fmt.Errorf("insufficient energy production: need %d, have %d", requirement.Energy, player.Production.Energy)
	}
	if player.Production.Heat < requirement.Heat {
		return fmt.Errorf("insufficient heat production: need %d, have %d", requirement.Heat, player.Production.Heat)
	}
	return nil
}

// PayResourceCost deducts the cost from player's resources
func PayResourceCost(player *model.Player, cost model.ResourceSet) {
	player.Resources.Credits -= cost.Credits
	player.Resources.Steel -= cost.Steel
	player.Resources.Titanium -= cost.Titanium
	player.Resources.Plants -= cost.Plants
	player.Resources.Energy -= cost.Energy
	player.Resources.Heat -= cost.Heat
}

// AddResources adds resources to a player
func AddResources(player *model.Player, resources model.ResourceSet) {
	player.Resources.Credits += resources.Credits
	player.Resources.Steel += resources.Steel
	player.Resources.Titanium += resources.Titanium
	player.Resources.Plants += resources.Plants
	player.Resources.Energy += resources.Energy
	player.Resources.Heat += resources.Heat
}

// AddProduction increases a player's production
func AddProduction(player *model.Player, production model.ResourceSet) {
	player.Production.Credits += production.Credits
	player.Production.Steel += production.Steel
	player.Production.Titanium += production.Titanium
	player.Production.Plants += production.Plants
	player.Production.Energy += production.Energy
	player.Production.Heat += production.Heat
}

// GetPlayerTags returns all tags from cards the player has played
func GetPlayerTags(player *model.Player) []model.CardTag {
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
	
	return tags
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