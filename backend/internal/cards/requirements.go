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
	if len(card.Requirements) == 0 {
		return nil
	}

	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🚨 Validating card requirements - card has requirements to check")

	// Validate each requirement
	for _, requirement := range card.Requirements {
		if err := rv.validateSingleRequirement(ctx, requirement, game, player); err != nil {
			return err
		}
	}

	log.Debug("✅ Card requirements validation passed")
	return nil
}

// validateSingleRequirement validates a single requirement
func (rv *RequirementsValidator) validateSingleRequirement(ctx context.Context, requirement model.Requirement, game *model.Game, player *model.Player) error {
	switch requirement.Type {
	case model.RequirementTemperature:
		return rv.validateParameterRequirement(requirement, game.GlobalParameters.Temperature, "temperature")
	case model.RequirementOxygen:
		return rv.validateParameterRequirement(requirement, game.GlobalParameters.Oxygen, "oxygen")
	case model.RequirementOceans:
		return rv.validateParameterRequirement(requirement, game.GlobalParameters.Oceans, "oceans")
	case model.RequirementTR:
		return rv.validateParameterRequirement(requirement, player.TerraformRating, "terraform rating")
	case model.RequirementTags:
		return rv.validateTagRequirement(ctx, requirement, game, player)
	case model.RequirementProduction:
		return rv.validateProductionRequirement(ctx, requirement, game, player)
	case model.RequirementResource:
		return rv.validateResourceRequirement(ctx, requirement, game, player)
	default:
		// Unknown requirement type, pass for now
		return nil
	}
}

// validateParameterRequirement validates min/max parameter requirements
func (rv *RequirementsValidator) validateParameterRequirement(requirement model.Requirement, currentValue int, paramName string) error {
	if requirement.Min != nil && currentValue < *requirement.Min {
		return fmt.Errorf("%s requirement not met: need at least %d, current is %d", paramName, *requirement.Min, currentValue)
	}
	if requirement.Max != nil && currentValue > *requirement.Max {
		return fmt.Errorf("%s requirement not met: need at most %d, current is %d", paramName, *requirement.Max, currentValue)
	}
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
	return len(card.Requirements) > 0
}

// GetPlayerTagCounts returns the tag counts for a player (public method for external use)
func (rv *RequirementsValidator) GetPlayerTagCounts(ctx context.Context, player *model.Player) map[model.CardTag]int {
	return rv.countPlayerTags(ctx, player)
}

// validateTagRequirement validates tag-based requirements using the new Requirement structure
func (rv *RequirementsValidator) validateTagRequirement(ctx context.Context, requirement model.Requirement, game *model.Game, player *model.Player) error {
	if requirement.Tag == nil {
		return fmt.Errorf("tag requirement missing tag specification")
	}

	playerTagCounts := rv.countPlayerTags(ctx, player)
	currentCount := playerTagCounts[*requirement.Tag]

	// Check minimum tag count requirement
	if requirement.Min != nil && currentCount < *requirement.Min {
		return fmt.Errorf("insufficient %s tags: need at least %d, have %d", *requirement.Tag, *requirement.Min, currentCount)
	}

	// Check maximum tag count requirement (rare but possible)
	if requirement.Max != nil && currentCount > *requirement.Max {
		return fmt.Errorf("too many %s tags: need at most %d, have %d", *requirement.Tag, *requirement.Max, currentCount)
	}

	return nil
}

// validateProductionRequirement validates production-based requirements using the new Requirement structure
func (rv *RequirementsValidator) validateProductionRequirement(ctx context.Context, requirement model.Requirement, game *model.Game, player *model.Player) error {
	if requirement.Resource == nil {
		return fmt.Errorf("production requirement missing resource specification")
	}

	var currentProduction int
	resourceName := string(*requirement.Resource)

	// Get current production value based on resource type
	switch *requirement.Resource {
	case model.ResourceCredits:
		currentProduction = player.Production.Credits
	case model.ResourceSteel:
		currentProduction = player.Production.Steel
	case model.ResourceTitanium:
		currentProduction = player.Production.Titanium
	case model.ResourcePlants:
		currentProduction = player.Production.Plants
	case model.ResourceEnergy:
		currentProduction = player.Production.Energy
	case model.ResourceHeat:
		currentProduction = player.Production.Heat
	default:
		return fmt.Errorf("unsupported production resource type: %s", *requirement.Resource)
	}

	// Check minimum production requirement
	if requirement.Min != nil && currentProduction < *requirement.Min {
		return fmt.Errorf("insufficient %s production: need at least %d, have %d", resourceName, *requirement.Min, currentProduction)
	}

	// Check maximum production requirement (for cards that require low production)
	if requirement.Max != nil && currentProduction > *requirement.Max {
		return fmt.Errorf("too much %s production: need at most %d, have %d", resourceName, *requirement.Max, currentProduction)
	}

	return nil
}

// validateResourceRequirement validates resource-based requirements using the new Requirement structure
func (rv *RequirementsValidator) validateResourceRequirement(ctx context.Context, requirement model.Requirement, game *model.Game, player *model.Player) error {
	if requirement.Resource == nil {
		return fmt.Errorf("resource requirement missing resource specification")
	}

	var currentResources int
	resourceName := string(*requirement.Resource)

	// Get current resource value based on resource type
	switch *requirement.Resource {
	case model.ResourceCredits:
		currentResources = player.Resources.Credits
	case model.ResourceSteel:
		currentResources = player.Resources.Steel
	case model.ResourceTitanium:
		currentResources = player.Resources.Titanium
	case model.ResourcePlants:
		currentResources = player.Resources.Plants
	case model.ResourceEnergy:
		currentResources = player.Resources.Energy
	case model.ResourceHeat:
		currentResources = player.Resources.Heat
	default:
		return fmt.Errorf("unsupported resource type: %s", *requirement.Resource)
	}

	// Check minimum resource requirement
	if requirement.Min != nil && currentResources < *requirement.Min {
		return fmt.Errorf("insufficient %s: need at least %d, have %d", resourceName, *requirement.Min, currentResources)
	}

	// Check maximum resource requirement (rare but could exist)
	if requirement.Max != nil && currentResources > *requirement.Max {
		return fmt.Errorf("too many %s: need at most %d, have %d", resourceName, *requirement.Max, currentResources)
	}

	return nil
}

// ValidateCardAffordability validates that the player can afford to play a card including all resource deductions
func (rv *RequirementsValidator) ValidateCardAffordability(ctx context.Context, gameID, playerID string, card *model.Card, player *model.Player) error {
	log := logger.WithGameContext(gameID, playerID)

	// Calculate total costs including card cost and behavioral resource deductions
	totalCosts := rv.calculateTotalCardCosts(ctx, card, player)

	// Validate card cost
	if card.Cost > 0 {
		if player.Resources.Credits < card.Cost+totalCosts.Credits {
			return fmt.Errorf("insufficient credits: need %d (card cost %d + behavioral costs %d), have %d",
				card.Cost+totalCosts.Credits, card.Cost, totalCosts.Credits, player.Resources.Credits)
		}
	}

	// Validate resource deductions from card behaviors
	if totalCosts.Steel > 0 && player.Resources.Steel < totalCosts.Steel {
		return fmt.Errorf("insufficient steel for card effects: need %d, have %d", totalCosts.Steel, player.Resources.Steel)
	}
	if totalCosts.Titanium > 0 && player.Resources.Titanium < totalCosts.Titanium {
		return fmt.Errorf("insufficient titanium for card effects: need %d, have %d", totalCosts.Titanium, player.Resources.Titanium)
	}
	if totalCosts.Plants > 0 && player.Resources.Plants < totalCosts.Plants {
		return fmt.Errorf("insufficient plants for card effects: need %d, have %d", totalCosts.Plants, player.Resources.Plants)
	}
	if totalCosts.Energy > 0 && player.Resources.Energy < totalCosts.Energy {
		return fmt.Errorf("insufficient energy for card effects: need %d, have %d", totalCosts.Energy, player.Resources.Energy)
	}
	if totalCosts.Heat > 0 && player.Resources.Heat < totalCosts.Heat {
		return fmt.Errorf("insufficient heat for card effects: need %d, have %d", totalCosts.Heat, player.Resources.Heat)
	}

	// Validate production deductions (negative production effects)
	if err := rv.validateProductionDeductions(ctx, card, player); err != nil {
		return fmt.Errorf("production requirement validation failed: %w", err)
	}

	log.Debug("✅ Card affordability validation passed",
		zap.Int("total_credit_cost", card.Cost+totalCosts.Credits),
		zap.Int("card_cost", card.Cost),
		zap.Int("behavioral_costs_credits", totalCosts.Credits),
		zap.Int("behavioral_costs_steel", totalCosts.Steel),
		zap.Int("behavioral_costs_titanium", totalCosts.Titanium),
		zap.Int("behavioral_costs_plants", totalCosts.Plants),
		zap.Int("behavioral_costs_energy", totalCosts.Energy),
		zap.Int("behavioral_costs_heat", totalCosts.Heat))

	return nil
}

// calculateTotalCardCosts analyzes card behaviors and calculates all resource costs
func (rv *RequirementsValidator) calculateTotalCardCosts(ctx context.Context, card *model.Card, player *model.Player) model.Resources {
	totalCosts := model.Resources{}

	// Process all behaviors to find immediate resource costs (auto triggers)
	for _, behavior := range card.Behaviors {
		// Only process auto triggers (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == model.ResourceTriggerAuto {
			// Check inputs (explicit costs)
			for _, input := range behavior.Inputs {
				switch input.Type {
				case model.ResourceCredits:
					totalCosts.Credits += input.Amount
				case model.ResourceSteel:
					totalCosts.Steel += input.Amount
				case model.ResourceTitanium:
					totalCosts.Titanium += input.Amount
				case model.ResourcePlants:
					totalCosts.Plants += input.Amount
				case model.ResourceEnergy:
					totalCosts.Energy += input.Amount
				case model.ResourceHeat:
					totalCosts.Heat += input.Amount
				}
			}

			// Check outputs with negative amounts (resource deductions)
			for _, output := range behavior.Outputs {
				if output.Amount < 0 {
					switch output.Type {
					case model.ResourceCredits:
						totalCosts.Credits += -output.Amount
					case model.ResourceSteel:
						totalCosts.Steel += -output.Amount
					case model.ResourceTitanium:
						totalCosts.Titanium += -output.Amount
					case model.ResourcePlants:
						totalCosts.Plants += -output.Amount
					case model.ResourceEnergy:
						totalCosts.Energy += -output.Amount
					case model.ResourceHeat:
						totalCosts.Heat += -output.Amount
					}
				}
			}
		}
	}

	return totalCosts
}

// validateProductionDeductions validates that the player can afford any negative production effects
func (rv *RequirementsValidator) validateProductionDeductions(ctx context.Context, card *model.Card, player *model.Player) error {
	// Calculate total production changes from card behaviors
	productionChanges := model.Production{}

	// Process all behaviors to find production effects
	for _, behavior := range card.Behaviors {
		// Only process auto triggers (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == model.ResourceTriggerAuto {
			for _, output := range behavior.Outputs {
				switch output.Type {
				case model.ResourceCreditsProduction:
					productionChanges.Credits += output.Amount
				case model.ResourceSteelProduction:
					productionChanges.Steel += output.Amount
				case model.ResourceTitaniumProduction:
					productionChanges.Titanium += output.Amount
				case model.ResourcePlantsProduction:
					productionChanges.Plants += output.Amount
				case model.ResourceEnergyProduction:
					productionChanges.Energy += output.Amount
				case model.ResourceHeatProduction:
					productionChanges.Heat += output.Amount
				}
			}
		}
	}

	// Check that negative production changes don't exceed minimum allowed production
	// Credits production can go down to -5, other productions can't go below 0
	if productionChanges.Credits < 0 && player.Production.Credits+productionChanges.Credits < -5 {
		return fmt.Errorf("insufficient credit production: card would reduce production to %d (below minimum of -5)",
			player.Production.Credits+productionChanges.Credits)
	}
	if productionChanges.Steel < 0 && player.Production.Steel+productionChanges.Steel < 0 {
		return fmt.Errorf("insufficient steel production: card would reduce production to %d (below 0)",
			player.Production.Steel+productionChanges.Steel)
	}
	if productionChanges.Titanium < 0 && player.Production.Titanium+productionChanges.Titanium < 0 {
		return fmt.Errorf("insufficient titanium production: card would reduce production to %d (below 0)",
			player.Production.Titanium+productionChanges.Titanium)
	}
	if productionChanges.Plants < 0 && player.Production.Plants+productionChanges.Plants < 0 {
		return fmt.Errorf("insufficient plant production: card would reduce production to %d (below 0)",
			player.Production.Plants+productionChanges.Plants)
	}
	if productionChanges.Energy < 0 && player.Production.Energy+productionChanges.Energy < 0 {
		return fmt.Errorf("insufficient energy production: card would reduce production to %d (below 0)",
			player.Production.Energy+productionChanges.Energy)
	}
	if productionChanges.Heat < 0 && player.Production.Heat+productionChanges.Heat < 0 {
		return fmt.Errorf("insufficient heat production: card would reduce production to %d (below 0)",
			player.Production.Heat+productionChanges.Heat)
	}

	return nil
}
