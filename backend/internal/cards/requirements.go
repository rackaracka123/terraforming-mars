package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// RequirementsValidator handles enhanced card requirement validation
type RequirementsValidator struct {
	cardRepo repository.CardRepository
}

// NewRequirementsValidator creates a new enhanced requirements validator
func NewRequirementsValidator(cardRepo repository.CardRepository) *RequirementsValidator {
	return &RequirementsValidator{
		cardRepo: cardRepo,
	}
}

// ValidateCardRequirements validates that a card's requirements are met with full context
func (rv *RequirementsValidator) ValidateCardRequirements(ctx context.Context, gameID, playerID string, card *model.Card, game *model.Game, player *model.Player) error {
	// Check if card has any requirements
	hasRequirements := card.Requirements.MinTemperature != nil ||
		card.Requirements.MaxTemperature != nil ||
		card.Requirements.MinOxygen != nil ||
		card.Requirements.MaxOxygen != nil ||
		card.Requirements.MinOceans != nil ||
		card.Requirements.MaxOceans != nil ||
		len(card.Requirements.RequiredTags) > 0 ||
		card.Requirements.RequiredProduction != nil

	if !hasRequirements {
		return nil
	}

	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🚨 Validating card requirements - card has requirements to check")

	// Validate global parameter requirements
	if err := rv.validateGlobalParameterRequirements(card.Requirements, game.GlobalParameters); err != nil {
		return err
	}

	// Validate tag requirements
	if len(card.Requirements.RequiredTags) > 0 {
		if err := rv.validateTagRequirements(ctx, card.Requirements.RequiredTags, player); err != nil {
			return err
		}
	}

	// Validate production requirements
	if card.Requirements.RequiredProduction != nil {
		if err := rv.validateProductionRequirements(*card.Requirements.RequiredProduction, player.Production); err != nil {
			return err
		}
	}

	log.Debug("✅ Card requirements validation passed")
	return nil
}

// validateGlobalParameterRequirements checks temperature, oxygen, and ocean requirements
func (rv *RequirementsValidator) validateGlobalParameterRequirements(requirements model.CardRequirements, globalParams model.GlobalParameters) error {
	log := logger.Get()

	// Log current global parameters and requirements for debugging
	log.Debug("🌍 Validating global parameter requirements",
		zap.Int("current_temperature", globalParams.Temperature),
		zap.Int("current_oxygen", globalParams.Oxygen),
		zap.Int("current_oceans", globalParams.Oceans))

	if requirements.MinTemperature != nil {
		log.Debug("❄️ Checking min temperature", zap.Int("required", *requirements.MinTemperature), zap.Int("current", globalParams.Temperature))
	}
	if requirements.MinOxygen != nil {
		log.Debug("💨 Checking min oxygen", zap.Int("required", *requirements.MinOxygen), zap.Int("current", globalParams.Oxygen))
	}
	if requirements.MinOceans != nil {
		log.Debug("🌊 Checking min oceans", zap.Int("required", *requirements.MinOceans), zap.Int("current", globalParams.Oceans))
	}

	// Check temperature requirements
	if requirements.MinTemperature != nil && globalParams.Temperature < *requirements.MinTemperature {
		return fmt.Errorf("minimum temperature requirement not met: need %d°C, current %d°C", *requirements.MinTemperature, globalParams.Temperature)
	}
	if requirements.MaxTemperature != nil && globalParams.Temperature > *requirements.MaxTemperature {
		return fmt.Errorf("maximum temperature requirement exceeded: limit %d°C, current %d°C", *requirements.MaxTemperature, globalParams.Temperature)
	}

	// Check oxygen requirements
	if requirements.MinOxygen != nil && globalParams.Oxygen < *requirements.MinOxygen {
		return fmt.Errorf("minimum oxygen requirement not met: need %d%%, current %d%%", *requirements.MinOxygen, globalParams.Oxygen)
	}
	if requirements.MaxOxygen != nil && globalParams.Oxygen > *requirements.MaxOxygen {
		return fmt.Errorf("maximum oxygen requirement exceeded: limit %d%%, current %d%%", *requirements.MaxOxygen, globalParams.Oxygen)
	}

	// Check ocean requirements
	if requirements.MinOceans != nil && globalParams.Oceans < *requirements.MinOceans {
		return fmt.Errorf("minimum ocean requirement not met: need %d, current %d", *requirements.MinOceans, globalParams.Oceans)
	}
	if requirements.MaxOceans != nil && globalParams.Oceans > *requirements.MaxOceans {
		return fmt.Errorf("maximum ocean requirement exceeded: limit %d, current %d", *requirements.MaxOceans, globalParams.Oceans)
	}

	return nil
}

// validateTagRequirements checks if player has required card tags
func (rv *RequirementsValidator) validateTagRequirements(ctx context.Context, requiredTags []model.CardTag, player *model.Player) error {
	playerTagCounts := rv.countPlayerTags(ctx, player)

	for _, requiredTag := range requiredTags {
		if playerTagCounts[requiredTag] == 0 {
			return fmt.Errorf("required tag not found: %s", requiredTag)
		}
	}

	return nil
}

// validateProductionRequirements checks if player has sufficient production levels
func (rv *RequirementsValidator) validateProductionRequirements(requiredProduction model.ResourceSet, playerProduction model.Production) error {
	if requiredProduction.Credits > 0 && playerProduction.Credits < requiredProduction.Credits {
		return fmt.Errorf("insufficient credit production: need %d, have %d", requiredProduction.Credits, playerProduction.Credits)
	}
	if requiredProduction.Steel > 0 && playerProduction.Steel < requiredProduction.Steel {
		return fmt.Errorf("insufficient steel production: need %d, have %d", requiredProduction.Steel, playerProduction.Steel)
	}
	if requiredProduction.Titanium > 0 && playerProduction.Titanium < requiredProduction.Titanium {
		return fmt.Errorf("insufficient titanium production: need %d, have %d", requiredProduction.Titanium, playerProduction.Titanium)
	}
	if requiredProduction.Plants > 0 && playerProduction.Plants < requiredProduction.Plants {
		return fmt.Errorf("insufficient plant production: need %d, have %d", requiredProduction.Plants, playerProduction.Plants)
	}
	if requiredProduction.Energy > 0 && playerProduction.Energy < requiredProduction.Energy {
		return fmt.Errorf("insufficient energy production: need %d, have %d", requiredProduction.Energy, playerProduction.Energy)
	}
	if requiredProduction.Heat > 0 && playerProduction.Heat < requiredProduction.Heat {
		return fmt.Errorf("insufficient heat production: need %d, have %d", requiredProduction.Heat, playerProduction.Heat)
	}

	return nil
}

// countPlayerTags counts the occurrence of each tag in player's played cards and corporation
func (rv *RequirementsValidator) countPlayerTags(ctx context.Context, player *model.Player) map[model.CardTag]int {
	tagCounts := make(map[model.CardTag]int)

	// Count tags from played cards
	for _, cardID := range player.PlayedCards {
		card, err := rv.cardRepo.GetCardByID(ctx, cardID)
		if err != nil || card == nil {
			continue // Skip if card not found
		}

		for _, tag := range card.Tags {
			tagCounts[tag]++
		}
	}

	// Add corporation tags if player has a corporation
	if player.Corporation != nil && *player.Corporation != "" {
		corporationCard, err := rv.cardRepo.GetCardByID(ctx, *player.Corporation)
		if err == nil && corporationCard != nil {
			for _, tag := range corporationCard.Tags {
				tagCounts[tag]++
			}
		}
	}

	return tagCounts
}

// HasRequirements checks if a card has any requirements to validate
func (rv *RequirementsValidator) HasRequirements(card *model.Card) bool {
	return card.Requirements.MinTemperature != nil ||
		card.Requirements.MaxTemperature != nil ||
		card.Requirements.MinOxygen != nil ||
		card.Requirements.MaxOxygen != nil ||
		card.Requirements.MinOceans != nil ||
		card.Requirements.MaxOceans != nil ||
		len(card.Requirements.RequiredTags) > 0 ||
		card.Requirements.RequiredProduction != nil
}

// GetPlayerTagCounts returns the tag counts for a player (public method for external use)
func (rv *RequirementsValidator) GetPlayerTagCounts(ctx context.Context, player *model.Player) map[model.CardTag]int {
	return rv.countPlayerTags(ctx, player)
}
