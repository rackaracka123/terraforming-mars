package cards

import (
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
)

// Validator handles card requirement validation
type Validator struct{}

// NewValidator creates a new card validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRequirements validates that a card's requirements are met
func (v *Validator) ValidateRequirements(card *model.Card, game *model.Game, player *model.Player) error {
	// Check if card has any requirements
	if !v.hasRequirements(card.Requirements) {
		return nil
	}

	log := logger.Get()
	log.Debug("🔍 Validating card requirements",
		zap.String("card_name", card.Name),
		zap.Any("requirements", card.Requirements))

	// Validate global parameter requirements
	if err := v.validateGlobalParameters(card.Requirements, game.GlobalParameters); err != nil {
		return err
	}

	// Validate tag requirements
	if len(card.Requirements.RequiredTags) > 0 {
		if err := v.validateTags(card.Requirements.RequiredTags, player); err != nil {
			return err
		}
	}

	// Validate production requirements
	if card.Requirements.RequiredProduction != nil {
		if err := v.validateProduction(*card.Requirements.RequiredProduction, player.Production); err != nil {
			return err
		}
	}

	return nil
}

// hasRequirements checks if a card has any requirements to validate
func (v *Validator) hasRequirements(requirements model.CardRequirements) bool {
	return requirements.MinTemperature != nil ||
		requirements.MaxTemperature != nil ||
		requirements.MinOxygen != nil ||
		requirements.MaxOxygen != nil ||
		requirements.MinOceans != nil ||
		requirements.MaxOceans != nil ||
		len(requirements.RequiredTags) > 0 ||
		requirements.RequiredProduction != nil
}

// validateGlobalParameters checks temperature, oxygen, and ocean requirements
func (v *Validator) validateGlobalParameters(requirements model.CardRequirements, globalParams model.GlobalParameters) error {
	// Check temperature requirements
	if requirements.MinTemperature != nil && globalParams.Temperature < *requirements.MinTemperature {
		return fmt.Errorf("minimum temperature requirement not met: need %d°C, current %d°C",
			*requirements.MinTemperature, globalParams.Temperature)
	}
	if requirements.MaxTemperature != nil && globalParams.Temperature > *requirements.MaxTemperature {
		return fmt.Errorf("maximum temperature requirement exceeded: limit %d°C, current %d°C",
			*requirements.MaxTemperature, globalParams.Temperature)
	}

	// Check oxygen requirements
	if requirements.MinOxygen != nil && globalParams.Oxygen < *requirements.MinOxygen {
		return fmt.Errorf("minimum oxygen requirement not met: need %d%%, current %d%%",
			*requirements.MinOxygen, globalParams.Oxygen)
	}
	if requirements.MaxOxygen != nil && globalParams.Oxygen > *requirements.MaxOxygen {
		return fmt.Errorf("maximum oxygen requirement exceeded: limit %d%%, current %d%%",
			*requirements.MaxOxygen, globalParams.Oxygen)
	}

	// Check ocean requirements
	if requirements.MinOceans != nil && globalParams.Oceans < *requirements.MinOceans {
		return fmt.Errorf("minimum ocean requirement not met: need %d, current %d",
			*requirements.MinOceans, globalParams.Oceans)
	}
	if requirements.MaxOceans != nil && globalParams.Oceans > *requirements.MaxOceans {
		return fmt.Errorf("maximum ocean requirement exceeded: limit %d, current %d",
			*requirements.MaxOceans, globalParams.Oceans)
	}

	return nil
}

// validateTags checks if player has required card tags
func (v *Validator) validateTags(requiredTags []model.CardTag, player *model.Player) error {
	playerTagCounts := v.countPlayerTags(player)

	for _, requiredTag := range requiredTags {
		if playerTagCounts[requiredTag] == 0 {
			return fmt.Errorf("required tag not met: need %s tag", requiredTag)
		}
	}

	return nil
}

// validateProduction checks if player meets production requirements
func (v *Validator) validateProduction(required model.ResourceSet, playerProduction model.Production) error {
	// Convert production to ResourceSet-like structure for comparison
	// This is a simplified approach - actual implementation would depend on model structure
	if required.Credits > 0 && playerProduction.Credits < required.Credits {
		return fmt.Errorf("insufficient credit production: need %d, have %d", required.Credits, playerProduction.Credits)
	}
	if required.Steel > 0 && playerProduction.Steel < required.Steel {
		return fmt.Errorf("insufficient steel production: need %d, have %d", required.Steel, playerProduction.Steel)
	}
	if required.Titanium > 0 && playerProduction.Titanium < required.Titanium {
		return fmt.Errorf("insufficient titanium production: need %d, have %d", required.Titanium, playerProduction.Titanium)
	}
	if required.Plants > 0 && playerProduction.Plants < required.Plants {
		return fmt.Errorf("insufficient plant production: need %d, have %d", required.Plants, playerProduction.Plants)
	}
	if required.Energy > 0 && playerProduction.Energy < required.Energy {
		return fmt.Errorf("insufficient energy production: need %d, have %d", required.Energy, playerProduction.Energy)
	}
	if required.Heat > 0 && playerProduction.Heat < required.Heat {
		return fmt.Errorf("insufficient heat production: need %d, have %d", required.Heat, playerProduction.Heat)
	}

	return nil
}

// countPlayerTags counts the occurrences of each tag in the player's played cards
func (v *Validator) countPlayerTags(player *model.Player) map[model.CardTag]int {
	tagCounts := make(map[model.CardTag]int)

	// This would need to be implemented based on how card data is accessed
	// For now, return empty map to avoid compilation errors
	// TODO: Implement tag counting from player's played cards and corporation

	return tagCounts
}
