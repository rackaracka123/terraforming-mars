package game

import (
	"terraforming-mars-backend/internal/session/game/card"
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/types"
)

// RequirementsValidator handles card requirement validation in session-scoped architecture
type RequirementsValidator struct {
	cardRepo card.Repository
}

// NewRequirementsValidator creates a new requirements validator
func NewRequirementsValidator(cardRepo card.Repository) *RequirementsValidator {
	return &RequirementsValidator{
		cardRepo: cardRepo,
	}
}

// ValidateCardRequirements validates that a card's requirements are met with full context
func (rv *RequirementsValidator) ValidateCardRequirements(ctx context.Context, game *Game, p *player.Player, card *card.Card) error {
	// Check if card has any requirements
	if len(card.Requirements) == 0 {
		return nil
	}

	log := logger.WithGameContext(p.GameID(), p.ID())
	log.Debug("🚨 Validating card requirements - card has requirements to check")

	// Validate each requirement
	for _, requirement := range card.Requirements {
		if err := rv.validateSingleRequirement(ctx, requirement, game, p); err != nil {
			return err
		}
	}

	log.Debug("✅ Card requirements validation passed")
	return nil
}

// validateSingleRequirement validates a single requirement
func (rv *RequirementsValidator) validateSingleRequirement(ctx context.Context, requirement card.Requirement, game *Game, p *player.Player) error {
	switch requirement.Type {
	case card.RequirementTemperature:
		return rv.validateParameterRequirement(requirement, game.GlobalParameters.Temperature, "temperature")
	case card.RequirementOxygen:
		return rv.validateParameterRequirement(requirement, game.GlobalParameters.Oxygen, "oxygen")
	case card.RequirementOceans:
		return rv.validateParameterRequirement(requirement, game.GlobalParameters.Oceans, "oceans")
	case card.RequirementTR:
		return rv.validateParameterRequirement(requirement, p.TerraformRating(), "terraform rating")
	case card.RequirementTags:
		return rv.validateTagRequirement(ctx, requirement, game, p)
	case card.RequirementProduction:
		return rv.validateProductionRequirement(ctx, requirement, game, p)
	case card.RequirementResource:
		return rv.validateResourceRequirement(ctx, requirement, game, p)
	default:
		// Unknown requirement type, pass for now
		return nil
	}
}

// validateParameterRequirement validates min/max parameter requirements
func (rv *RequirementsValidator) validateParameterRequirement(requirement card.Requirement, currentValue int, paramName string) error {
	if requirement.Min != nil && currentValue < *requirement.Min {
		return fmt.Errorf("%s requirement not met: need at least %d, current is %d", paramName, *requirement.Min, currentValue)
	}
	if requirement.Max != nil && currentValue > *requirement.Max {
		return fmt.Errorf("%s requirement not met: need at most %d, current is %d", paramName, *requirement.Max, currentValue)
	}
	return nil
}

// validateTagRequirement validates tag-based requirements
func (rv *RequirementsValidator) validateTagRequirement(ctx context.Context, requirement card.Requirement, game *Game, p *player.Player) error {
	if requirement.Tag == nil {
		return fmt.Errorf("tag requirement missing tag specification")
	}

	playerTagCounts := rv.countPlayerTags(ctx, p)
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

// validateProductionRequirement validates production-based requirements
func (rv *RequirementsValidator) validateProductionRequirement(ctx context.Context, requirement card.Requirement, game *Game, p *player.Player) error {
	if requirement.Resource == nil {
		return fmt.Errorf("production requirement missing resource specification")
	}

	var currentProduction int
	resourceName := string(*requirement.Resource)

	// Get current production value based on resource type
	switch *requirement.Resource {
	case types.ResourceCredits:
		currentProduction = p.Production().Credits
	case types.ResourceSteel:
		currentProduction = p.Production().Steel
	case types.ResourceTitanium:
		currentProduction = p.Production().Titanium
	case types.ResourcePlants:
		currentProduction = p.Production().Plants
	case types.ResourceEnergy:
		currentProduction = p.Production().Energy
	case types.ResourceHeat:
		currentProduction = p.Production().Heat
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

// validateResourceRequirement validates resource-based requirements
func (rv *RequirementsValidator) validateResourceRequirement(ctx context.Context, requirement card.Requirement, game *Game, p *player.Player) error {
	if requirement.Resource == nil {
		return fmt.Errorf("resource requirement missing resource specification")
	}

	var currentResources int
	resourceName := string(*requirement.Resource)

	// Get current resource value based on resource type
	switch *requirement.Resource {
	case types.ResourceCredits:
		currentResources = p.Resources().Credits
	case types.ResourceSteel:
		currentResources = p.Resources().Steel
	case types.ResourceTitanium:
		currentResources = p.Resources().Titanium
	case types.ResourcePlants:
		currentResources = p.Resources().Plants
	case types.ResourceEnergy:
		currentResources = p.Resources().Energy
	case types.ResourceHeat:
		currentResources = p.Resources().Heat
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

// countPlayerTags counts the occurrence of each tag in player's played cards and corporation
func (rv *RequirementsValidator) countPlayerTags(ctx context.Context, p *player.Player) map[types.CardTag]int {
	tagCounts := make(map[types.CardTag]int)

	// Count tags from played cards
	for _, cardID := range p.PlayedCards() {
		card, err := rv.cardRepo.GetCardByID(ctx, cardID)
		if err != nil || card == nil {
			continue // Skip if card not found
		}

		for _, tag := range card.Tags {
			tagCounts[tag]++
		}
	}

	// Add corporation tags if player has a corporation
	if p.Corporation() != nil {
		for _, tag := range p.Corporation().Tags {
			tagCounts[tag]++
		}
	}

	return tagCounts
}

// HasRequirements checks if a card has any requirements to validate
func (rv *RequirementsValidator) HasRequirements(card *card.Card) bool {
	return len(card.Requirements) > 0
}

// GetPlayerTagCounts returns the tag counts for a player (public method for external use)
func (rv *RequirementsValidator) GetPlayerTagCounts(ctx context.Context, p *player.Player) map[types.CardTag]int {
	return rv.countPlayerTags(ctx, p)
}

// calculateEffectiveCost calculates the card cost after applying requirement modifiers (discounts)
func (rv *RequirementsValidator) calculateEffectiveCost(card *card.Card, p *player.Player) int {
	effectiveCost := card.Cost

	// Apply discounts from requirement modifiers
	for _, modifier := range p.RequirementModifiers() {
		// Check if this modifier applies to this card
		if modifier.CardTarget != nil && *modifier.CardTarget == card.ID {
			// Check if this modifier affects credits (cost discounts)
			for _, affectedResource := range modifier.AffectedResources {
				if affectedResource == types.ResourceCredits {
					// Discount modifiers have negative amounts (e.g., -2 for 2 MC discount)
					// But the stored amount can be positive representing the discount value
					// So we subtract the absolute value
					if modifier.Amount < 0 {
						effectiveCost += modifier.Amount // Already negative, so this reduces cost
					} else {
						effectiveCost -= modifier.Amount // Positive amount means discount, so subtract
					}
					break
				}
			}
		}
	}

	// Ensure cost doesn't go below 0
	if effectiveCost < 0 {
		effectiveCost = 0
	}

	return effectiveCost
}

// ValidateCardAffordability validates that the player can afford to play a card including all resource deductions
// choiceIndex is optional and used when the card has choices between different effects
// payment is the proposed payment method (credits, steel, titanium) for the card cost
func (rv *RequirementsValidator) ValidateCardAffordability(ctx context.Context, p *player.Player, card *card.Card, payment *card.CardPayment, choiceIndex *int) error {
	log := logger.WithGameContext(p.GameID(), p.ID())

	// Calculate total costs from card behaviors (excluding card cost which is paid separately)
	totalCosts := rv.calculateTotalCardCosts(ctx, card, p, choiceIndex)

	// Validate payment for card cost
	if card.Cost > 0 {
		// Get player resources (needed for multiple checks)
		resources := p.Resources()

		// Calculate effective cost after applying discounts
		effectiveCost := rv.calculateEffectiveCost(card, p)

		// Check if card allows steel (has building tag) or titanium (has space tag)
		allowSteel := rv.cardHasTag(card, types.TagBuilding)
		allowTitanium := rv.cardHasTag(card, types.TagSpace)

		// Validate payment format and coverage (including payment substitutes)
		if err := payment.CoversCardCost(effectiveCost, allowSteel, allowTitanium, p.PaymentSubstitutes()); err != nil {
			// Calculate minimum alternative resources for better error messages
			minSteel, minTitanium := types.CalculateMinimumAlternativeResources(card.Cost, resources, allowSteel, allowTitanium)

			if minSteel > 0 && minTitanium > 0 {
				return fmt.Errorf("cannot afford card: %w (hint: need minSteel:%d or minTitanium:%d)", err, minSteel, minTitanium)
			} else if minSteel > 0 {
				return fmt.Errorf("cannot afford card: %w (hint: need minSteel:%d)", err, minSteel)
			} else if minTitanium > 0 {
				return fmt.Errorf("cannot afford card: %w (hint: need minTitanium:%d)", err, minTitanium)
			}
			return fmt.Errorf("cannot afford card: %w", err)
		}

		// Validate player has the resources for this payment
		if err := payment.CanAfford(resources); err != nil {
			return fmt.Errorf("insufficient resources for payment: %w", err)
		}
	}

	// Validate resource deductions from card behaviors (separate from card cost payment)
	resources := p.Resources()
	if totalCosts.Steel > 0 && resources.Steel-payment.Steel < totalCosts.Steel {
		return fmt.Errorf("insufficient steel for card effects: need %d after payment, have %d total (payment uses %d)",
			totalCosts.Steel, resources.Steel, payment.Steel)
	}
	if totalCosts.Titanium > 0 && resources.Titanium-payment.Titanium < totalCosts.Titanium {
		return fmt.Errorf("insufficient titanium for card effects: need %d after payment, have %d total (payment uses %d)",
			totalCosts.Titanium, resources.Titanium, payment.Titanium)
	}
	if totalCosts.Plants > 0 && resources.Plants < totalCosts.Plants {
		return fmt.Errorf("insufficient plants for card effects: need %d, have %d", totalCosts.Plants, resources.Plants)
	}
	if totalCosts.Energy > 0 && resources.Energy < totalCosts.Energy {
		return fmt.Errorf("insufficient energy for card effects: need %d, have %d", totalCosts.Energy, resources.Energy)
	}

	// Check payment substitutes don't interfere with behavioral costs
	if payment.Substitutes != nil {
		for resourceType, paymentAmount := range payment.Substitutes {
			var totalUsed int
			var available int

			switch resourceType {
			case types.ResourceHeat:
				totalUsed = paymentAmount + totalCosts.Heat
				available = resources.Heat
			case types.ResourceEnergy:
				totalUsed = paymentAmount + totalCosts.Energy
				available = resources.Energy
			case types.ResourcePlants:
				totalUsed = paymentAmount + totalCosts.Plants
				available = resources.Plants
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
		if resources.Heat-heatUsedForPayment < totalCosts.Heat {
			return fmt.Errorf("insufficient heat for card effects: need %d after payment, have %d total (payment uses %d)",
				totalCosts.Heat, resources.Heat, heatUsedForPayment)
		}
	}

	// Note: Credits check includes both payment and behavioral costs
	if totalCosts.Credits > 0 && resources.Credits-payment.Credits < totalCosts.Credits {
		return fmt.Errorf("insufficient credits for card effects: need %d after payment, have %d total (payment uses %d)",
			totalCosts.Credits, resources.Credits, payment.Credits)
	}

	// Validate production deductions (negative production effects)
	if err := rv.validateProductionDeductions(ctx, card, p, choiceIndex); err != nil {
		return fmt.Errorf("production requirement validation failed: %w", err)
	}

	// Calculate effective cost for logging purposes
	effectiveCost := card.Cost
	if card.Cost > 0 {
		effectiveCost = rv.calculateEffectiveCost(card, p)
	}

	log.Debug("✅ Card affordability validation passed",
		zap.Int("card_cost", card.Cost),
		zap.Int("effective_cost", effectiveCost),
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
func (rv *RequirementsValidator) cardHasTag(card *card.Card, tag types.CardTag) bool {
	for _, cardTag := range card.Tags {
		if cardTag == tag {
			return true
		}
	}
	return false
}

// calculateTotalCardCosts analyzes card behaviors and calculates all resource costs
// choiceIndex is optional and used when the card has choices between different effects
func (rv *RequirementsValidator) calculateTotalCardCosts(ctx context.Context, card *card.Card, p *player.Player, choiceIndex *int) types.Resources {
	totalCosts := types.Resources{}

	// Process all behaviors to find immediate resource costs (auto triggers)
	for _, behavior := range card.Behaviors {
		// Only process auto triggers (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == card.ResourceTriggerAuto {
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
func (rv *RequirementsValidator) validateProductionDeductions(ctx context.Context, card *card.Card, p *player.Player, choiceIndex *int) error {
	// Calculate total production changes from card behaviors
	productionChanges := types.Production{}

	// Process all behaviors to find production effects
	for _, behavior := range card.Behaviors {
		// Only process auto triggers (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == card.ResourceTriggerAuto {
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
	production := p.Production()
	if productionChanges.Credits < 0 && production.Credits+productionChanges.Credits < -5 {
		return fmt.Errorf("insufficient credit production: card would reduce production to %d (below minimum of -5)",
			production.Credits+productionChanges.Credits)
	}
	if productionChanges.Steel < 0 && production.Steel+productionChanges.Steel < 0 {
		return fmt.Errorf("insufficient steel production: card would reduce production to %d (below 0)",
			production.Steel+productionChanges.Steel)
	}
	if productionChanges.Titanium < 0 && production.Titanium+productionChanges.Titanium < 0 {
		return fmt.Errorf("insufficient titanium production: card would reduce production to %d (below 0)",
			production.Titanium+productionChanges.Titanium)
	}
	if productionChanges.Plants < 0 && production.Plants+productionChanges.Plants < 0 {
		return fmt.Errorf("insufficient plant production: card would reduce production to %d (below 0)",
			production.Plants+productionChanges.Plants)
	}
	if productionChanges.Energy < 0 && production.Energy+productionChanges.Energy < 0 {
		return fmt.Errorf("insufficient energy production: card would reduce production to %d (below 0)",
			production.Energy+productionChanges.Energy)
	}
	if productionChanges.Heat < 0 && production.Heat+productionChanges.Heat < 0 {
		return fmt.Errorf("insufficient heat production: card would reduce production to %d (below 0)",
			production.Heat+productionChanges.Heat)
	}

	return nil
}
