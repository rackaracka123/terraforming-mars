package service

import (
	"context"
	"fmt"

	cardpkg "terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/production"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/internal/shared/types"

	"go.uber.org/zap"
)

// RequirementsValidator handles enhanced card requirement validation
type RequirementsValidator struct {
	cardRepo cardpkg.CardRepository
}

// NewRequirementsValidator creates a new enhanced requirements validator
func NewRequirementsValidator(cardRepo cardpkg.CardRepository) *RequirementsValidator {
	return &RequirementsValidator{
		cardRepo: cardRepo,
	}
}

// ValidateCardRequirements validates that a card's requirements are met with full context
func (rv *RequirementsValidator) ValidateCardRequirements(ctx context.Context, gameID, playerID string, card *cardpkg.Card, game *game.Game, player *player.Player) error {
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
func (rv *RequirementsValidator) validateSingleRequirement(ctx context.Context, requirement cardpkg.Requirement, game *game.Game, player *player.Player) error {
	globalParams, err := game.GetGlobalParameters()
	if err != nil {
		return fmt.Errorf("failed to get global parameters: %w", err)
	}

	switch requirement.Type {
	case cardpkg.RequirementTemperature:
		return rv.validateParameterRequirement(requirement, globalParams.Temperature, "temperature")
	case cardpkg.RequirementOxygen:
		return rv.validateParameterRequirement(requirement, globalParams.Oxygen, "oxygen")
	case cardpkg.RequirementOceans:
		return rv.validateParameterRequirement(requirement, globalParams.Oceans, "oceans")
	case cardpkg.RequirementTR:
		return rv.validateParameterRequirement(requirement, player.TerraformRating, "terraform rating")
	case cardpkg.RequirementTags:
		return rv.validateTagRequirement(ctx, requirement, game, player)
	case cardpkg.RequirementProduction:
		return rv.validateProductionRequirement(ctx, requirement, game, player)
	case cardpkg.RequirementResource:
		return rv.validateResourceRequirement(ctx, requirement, game, player)
	default:
		// Unknown requirement type, pass for now
		return nil
	}
}

// validateParameterRequirement validates min/max parameter requirements
func (rv *RequirementsValidator) validateParameterRequirement(requirement cardpkg.Requirement, currentValue int, paramName string) error {
	if requirement.Min != nil && currentValue < *requirement.Min {
		return fmt.Errorf("%s requirement not met: need at least %d, current is %d", paramName, *requirement.Min, currentValue)
	}
	if requirement.Max != nil && currentValue > *requirement.Max {
		return fmt.Errorf("%s requirement not met: need at most %d, current is %d", paramName, *requirement.Max, currentValue)
	}
	return nil
}

// validateGlobalParameterRequirements checks temperature, oxygen, and ocean requirements
func (rv *RequirementsValidator) validateGlobalParameterRequirements(requirements cardpkg.CardRequirements, globalParams parameters.GlobalParameters) error {
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
func (rv *RequirementsValidator) validateTagRequirements(ctx context.Context, requiredTags []cardpkg.CardTag, player *player.Player) error {
	playerTagCounts := rv.countPlayerTags(ctx, player)

	for _, requiredTag := range requiredTags {
		if playerTagCounts[requiredTag] == 0 {
			return fmt.Errorf("required tag not found: %s", requiredTag)
		}
	}

	return nil
}

// validateProductionRequirements checks if player has sufficient production levels
func (rv *RequirementsValidator) validateProductionRequirements(requiredProduction resources.Production, playerProduction resources.Production) error {
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
func (rv *RequirementsValidator) countPlayerTags(ctx context.Context, player *player.Player) map[cardpkg.CardTag]int {
	tagCounts := make(map[cardpkg.CardTag]int)

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
	if player.Corporation != nil {
		for _, tag := range player.Corporation.Tags {
			tagCounts[tag]++
		}
	}

	return tagCounts
}

// HasRequirements checks if a card has any requirements to validate
func (rv *RequirementsValidator) HasRequirements(card *cardpkg.Card) bool {
	return len(card.Requirements) > 0
}

// GetPlayerTagCounts returns the tag counts for a player (public method for external use)
func (rv *RequirementsValidator) GetPlayerTagCounts(ctx context.Context, player *player.Player) map[cardpkg.CardTag]int {
	return rv.countPlayerTags(ctx, player)
}

// validateTagRequirement validates tag-based requirements using the new Requirement structure
func (rv *RequirementsValidator) validateTagRequirement(ctx context.Context, requirement cardpkg.Requirement, game *game.Game, player *player.Player) error {
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
func (rv *RequirementsValidator) validateProductionRequirement(ctx context.Context, requirement cardpkg.Requirement, game *game.Game, player *player.Player) error {
	if requirement.Resource == nil {
		return fmt.Errorf("production requirement missing resource specification")
	}

	playerProduction, err := player.GetProduction()
	if err != nil {
		return fmt.Errorf("failed to get player production: %w", err)
	}

	var currentProduction int
	resourceName := string(*requirement.Resource)

	// Get current production value based on resource type
	switch *requirement.Resource {
	case types.ResourceCredits:
		currentProduction = playerProduction.Credits
	case types.ResourceSteel:
		currentProduction = playerProduction.Steel
	case types.ResourceTitanium:
		currentProduction = playerProduction.Titanium
	case types.ResourcePlants:
		currentProduction = playerProduction.Plants
	case types.ResourceEnergy:
		currentProduction = playerProduction.Energy
	case types.ResourceHeat:
		currentProduction = playerProduction.Heat
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
func (rv *RequirementsValidator) validateResourceRequirement(ctx context.Context, requirement cardpkg.Requirement, game *game.Game, player *player.Player) error {
	if requirement.Resource == nil {
		return fmt.Errorf("resource requirement missing resource specification")
	}

	playerResources, err := player.GetResources()
	if err != nil {
		return fmt.Errorf("failed to get player resources: %w", err)
	}

	var currentResources int
	resourceName := string(*requirement.Resource)

	// Get current resource value based on resource type
	switch *requirement.Resource {
	case types.ResourceCredits:
		currentResources = playerResources.Credits
	case types.ResourceSteel:
		currentResources = playerResources.Steel
	case types.ResourceTitanium:
		currentResources = playerResources.Titanium
	case types.ResourcePlants:
		currentResources = playerResources.Plants
	case types.ResourceEnergy:
		currentResources = playerResources.Energy
	case types.ResourceHeat:
		currentResources = playerResources.Heat
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
// choiceIndex is optional and used when the card has choices between different effects
// payment is the proposed payment method (credits, steel, titanium) for the card cost
func (rv *RequirementsValidator) ValidateCardAffordability(ctx context.Context, gameID, playerID string, card *cardpkg.Card, player *player.Player, payment *types.CardPayment, choiceIndex *int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get player's current resources
	playerResources, err := player.GetResources()
	if err != nil {
		log.Error("Failed to get player resources", zap.Error(err))
		return fmt.Errorf("failed to get player resources: %w", err)
	}

	// Calculate total costs from card behaviors (excluding card cost which is paid separately)
	totalCosts := rv.calculateTotalCardCosts(ctx, card, player, choiceIndex)

	// Validate payment for card cost
	if card.Cost > 0 {
		// Check if card allows steel (has building tag) or titanium (has space tag)
		allowSteel := rv.cardHasTag(card, cardpkg.TagBuilding)
		allowTitanium := rv.cardHasTag(card, cardpkg.TagSpace)

		// Validate payment format and coverage (including payment substitutes)
		if err := payment.CoversCardCost(card.Cost, allowSteel, allowTitanium, player.PaymentSubstitutes); err != nil {
			return fmt.Errorf("cannot afford card: %w", err)
		}

		// Validate player has enough resources for the proposed payment
		if payment.Credits > playerResources.Credits {
			return fmt.Errorf("insufficient credits for payment: need %d, have %d", payment.Credits, playerResources.Credits)
		}
		if payment.Steel > playerResources.Steel {
			return fmt.Errorf("insufficient steel for payment: need %d, have %d", payment.Steel, playerResources.Steel)
		}
		if payment.Titanium > playerResources.Titanium {
			return fmt.Errorf("insufficient titanium for payment: need %d, have %d", payment.Titanium, playerResources.Titanium)
		}
		// Validate payment substitutes
		if payment.Substitutes != nil {
			for resourceType, amount := range payment.Substitutes {
				switch resourceType {
				case types.ResourceHeat:
					if amount > playerResources.Heat {
						return fmt.Errorf("insufficient heat for payment substitute: need %d, have %d", amount, playerResources.Heat)
					}
				case types.ResourceEnergy:
					if amount > playerResources.Energy {
						return fmt.Errorf("insufficient energy for payment substitute: need %d, have %d", amount, playerResources.Energy)
					}
				case types.ResourcePlants:
					if amount > playerResources.Plants {
						return fmt.Errorf("insufficient plants for payment substitute: need %d, have %d", amount, playerResources.Plants)
					}
				}
			}
		}
	}

	// Validate resource deductions from card behaviors (separate from card cost payment)
	if totalCosts.Steel > 0 && playerResources.Steel-payment.Steel < totalCosts.Steel {
		return fmt.Errorf("insufficient steel for card effects: need %d after payment, have %d total (payment uses %d)",
			totalCosts.Steel, playerResources.Steel, payment.Steel)
	}
	if totalCosts.Titanium > 0 && playerResources.Titanium-payment.Titanium < totalCosts.Titanium {
		return fmt.Errorf("insufficient titanium for card effects: need %d after payment, have %d total (payment uses %d)",
			totalCosts.Titanium, playerResources.Titanium, payment.Titanium)
	}
	if totalCosts.Plants > 0 && playerResources.Plants < totalCosts.Plants {
		return fmt.Errorf("insufficient plants for card effects: need %d, have %d", totalCosts.Plants, playerResources.Plants)
	}
	if totalCosts.Energy > 0 && playerResources.Energy < totalCosts.Energy {
		return fmt.Errorf("insufficient energy for card effects: need %d, have %d", totalCosts.Energy, playerResources.Energy)
	}

	// Check payment substitutes don't interfere with behavioral costs
	if payment.Substitutes != nil {
		for resourceType, paymentAmount := range payment.Substitutes {
			var totalUsed int
			var available int

			switch resourceType {
			case types.ResourceHeat:
				totalUsed = paymentAmount + totalCosts.Heat
				available = playerResources.Heat
			case types.ResourceEnergy:
				totalUsed = paymentAmount + totalCosts.Energy
				available = playerResources.Energy
			case types.ResourcePlants:
				totalUsed = paymentAmount + totalCosts.Plants
				available = playerResources.Plants
			}

			if totalUsed > available {
				return fmt.Errorf("insufficient %s: need %d for payment + %d for card effects = %d total, have %d",
					resourceType, paymentAmount, totalUsed-paymentAmount, totalUsed, available)
			}
		}
	}

	if totalCosts.Heat > 0 {
		heatUsedForPayment := 0
		if payment.Substitutes != nil {
			heatUsedForPayment = payment.Substitutes[types.ResourceHeat]
		}
		if playerResources.Heat-heatUsedForPayment < totalCosts.Heat {
			return fmt.Errorf("insufficient heat for card effects: need %d after payment, have %d total (payment uses %d)",
				totalCosts.Heat, playerResources.Heat, heatUsedForPayment)
		}
	}

	// Note: Credits check includes both payment and behavioral costs
	if totalCosts.Credits > 0 && playerResources.Credits-payment.Credits < totalCosts.Credits {
		return fmt.Errorf("insufficient credits for card effects: need %d after payment, have %d total (payment uses %d)",
			totalCosts.Credits, playerResources.Credits, payment.Credits)
	}

	// Validate production deductions (negative production effects)
	if err := rv.validateProductionDeductions(ctx, card, player, choiceIndex); err != nil {
		return fmt.Errorf("production requirement validation failed: %w", err)
	}

	log.Debug("✅ Card affordability validation passed",
		zap.Int("card_cost", card.Cost),
		zap.Int("payment_credits", payment.Credits),
		zap.Int("payment_steel", payment.Steel),
		zap.Int("payment_titanium", payment.Titanium),
		zap.Int("behavioral_costs_credits", totalCosts.Credits),
		zap.Int("behavioral_costs_steel", totalCosts.Steel),
		zap.Int("behavioral_costs_titanium", totalCosts.Titanium),
		zap.Int("behavioral_costs_plants", totalCosts.Plants),
		zap.Int("behavioral_costs_energy", totalCosts.Energy),
		zap.Int("behavioral_costs_heat", totalCosts.Heat))

	return nil
}

// cardHasTag checks if a card has a specific tag
func (rv *RequirementsValidator) cardHasTag(card *cardpkg.Card, tag cardpkg.CardTag) bool {
	for _, cardTag := range card.Tags {
		if cardTag == tag {
			return true
		}
	}
	return false
}

// calculateTotalCardCosts analyzes card behaviors and calculates all resource costs
// choiceIndex is optional and used when the card has choices between different effects
func (rv *RequirementsValidator) calculateTotalCardCosts(ctx context.Context, card *cardpkg.Card, player *player.Player, choiceIndex *int) resources.Resources {
	totalCosts := resources.Resources{}

	// Process all behaviors to find immediate resource costs (auto triggers)
	for _, behavior := range card.Behaviors {
		// Only process auto triggers (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == cardpkg.ResourceTriggerAuto {
			// Aggregate all inputs: behavior.Inputs + choice[choiceIndex].Inputs
			allInputs := behavior.Inputs

			// If choiceIndex is provided and this behavior has choices, add choice inputs
			if choiceIndex != nil && len(behavior.Choices) > 0 && *choiceIndex < len(behavior.Choices) {
				selectedChoice := behavior.Choices[*choiceIndex]
				allInputs = append(allInputs, selectedChoice.Inputs...)
			}

			// Check all aggregated inputs (explicit costs)
			for _, input := range allInputs {
				switch input.Type {
				case types.ResourceCredits:
					totalCosts.Credits += input.Amount
				case types.ResourceSteel:
					totalCosts.Steel += input.Amount
				case types.ResourceTitanium:
					totalCosts.Titanium += input.Amount
				case types.ResourcePlants:
					totalCosts.Plants += input.Amount
				case types.ResourceEnergy:
					totalCosts.Energy += input.Amount
				case types.ResourceHeat:
					totalCosts.Heat += input.Amount
				}
			}

			// Aggregate all outputs: behavior.Outputs + choice[choiceIndex].Outputs
			allOutputs := behavior.Outputs

			// If choiceIndex is provided and this behavior has choices, add choice outputs
			if choiceIndex != nil && len(behavior.Choices) > 0 && *choiceIndex < len(behavior.Choices) {
				selectedChoice := behavior.Choices[*choiceIndex]
				allOutputs = append(allOutputs, selectedChoice.Outputs...)
			}

			// Check outputs with negative amounts (resource deductions)
			for _, output := range allOutputs {
				if output.Amount < 0 {
					switch output.Type {
					case types.ResourceCredits:
						totalCosts.Credits += -output.Amount
					case types.ResourceSteel:
						totalCosts.Steel += -output.Amount
					case types.ResourceTitanium:
						totalCosts.Titanium += -output.Amount
					case types.ResourcePlants:
						totalCosts.Plants += -output.Amount
					case types.ResourceEnergy:
						totalCosts.Energy += -output.Amount
					case types.ResourceHeat:
						totalCosts.Heat += -output.Amount
					}
				}
			}
		}
	}

	return totalCosts
}

// validateProductionDeductions validates that the player can afford any negative production effects
func (rv *RequirementsValidator) validateProductionDeductions(ctx context.Context, card *cardpkg.Card, player *player.Player, choiceIndex *int) error {
	// Get player's current production
	playerProduction, err := player.GetProduction()
	if err != nil {
		return fmt.Errorf("failed to get player production: %w", err)
	}

	// Calculate total production changes from card behaviors
	productionChanges := resources.Production{}

	// Process all behaviors to find production effects
	for _, behavior := range card.Behaviors {
		// Only process auto triggers (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == cardpkg.ResourceTriggerAuto {
			// Aggregate all outputs: behavior.Outputs + choice[choiceIndex].Outputs
			allOutputs := behavior.Outputs

			// If choiceIndex is provided and this behavior has choices, add choice outputs
			if choiceIndex != nil && len(behavior.Choices) > 0 && *choiceIndex < len(behavior.Choices) {
				selectedChoice := behavior.Choices[*choiceIndex]
				allOutputs = append(allOutputs, selectedChoice.Outputs...)
			}

			// Process all aggregated outputs
			for _, output := range allOutputs {
				switch output.Type {
				case types.ResourceCreditsProduction:
					productionChanges.Credits += output.Amount
				case types.ResourceSteelProduction:
					productionChanges.Steel += output.Amount
				case types.ResourceTitaniumProduction:
					productionChanges.Titanium += output.Amount
				case types.ResourcePlantsProduction:
					productionChanges.Plants += output.Amount
				case types.ResourceEnergyProduction:
					productionChanges.Energy += output.Amount
				case types.ResourceHeatProduction:
					productionChanges.Heat += output.Amount
				}
			}
		}
	}

	// Check that negative production changes don't exceed minimum allowed production
	// Credits production can go down to -5, other productions can't go below 0
	if productionChanges.Credits < 0 && playerProduction.Credits+productionChanges.Credits < -5 {
		return fmt.Errorf("insufficient credit production: card would reduce production to %d (below minimum of -5)",
			playerProduction.Credits+productionChanges.Credits)
	}
	if productionChanges.Steel < 0 && playerProduction.Steel+productionChanges.Steel < 0 {
		return fmt.Errorf("insufficient steel production: card would reduce production to %d (below 0)",
			playerProduction.Steel+productionChanges.Steel)
	}
	if productionChanges.Titanium < 0 && playerProduction.Titanium+productionChanges.Titanium < 0 {
		return fmt.Errorf("insufficient titanium production: card would reduce production to %d (below 0)",
			playerProduction.Titanium+productionChanges.Titanium)
	}
	if productionChanges.Plants < 0 && playerProduction.Plants+productionChanges.Plants < 0 {
		return fmt.Errorf("insufficient plant production: card would reduce production to %d (below 0)",
			playerProduction.Plants+productionChanges.Plants)
	}
	if productionChanges.Energy < 0 && playerProduction.Energy+productionChanges.Energy < 0 {
		return fmt.Errorf("insufficient energy production: card would reduce production to %d (below 0)",
			playerProduction.Energy+productionChanges.Energy)
	}
	if productionChanges.Heat < 0 && playerProduction.Heat+productionChanges.Heat < 0 {
		return fmt.Errorf("insufficient heat production: card would reduce production to %d (below 0)",
			playerProduction.Heat+productionChanges.Heat)
	}

	return nil
}
