package cards

import (
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CardLookup is a minimal interface for looking up cards by ID
// This avoids import cycles with internal/cards package
type CardLookup interface {
	GetByID(id string) (*Card, error)
}

// RequirementModifierCalculator computes requirement modifiers from player effects and hand
type RequirementModifierCalculator struct {
	cardLookup CardLookup
}

// NewRequirementModifierCalculator creates a new calculator
func NewRequirementModifierCalculator(cardLookup CardLookup) *RequirementModifierCalculator {
	return &RequirementModifierCalculator{
		cardLookup: cardLookup,
	}
}

// Calculate computes all requirement modifiers for a player based on their effects and hand
func (c *RequirementModifierCalculator) Calculate(p *player.Player) []shared.RequirementModifier {
	if p == nil {
		return []shared.RequirementModifier{}
	}

	modifiers := []shared.RequirementModifier{}

	effects := p.Effects().List()
	handCardIDs := p.Hand().Cards()

	for _, effect := range effects {
		for _, output := range effect.Behavior.Outputs {
			if output.ResourceType != shared.ResourceDiscount {
				continue
			}

			// Case 1: Standard project discount (e.g., Ecoline's plant discount)
			if len(output.AffectedStandardProjects) > 0 {
				for _, project := range output.AffectedStandardProjects {
					projectCopy := project
					modifier := shared.RequirementModifier{
						Amount:                output.Amount,
						AffectedResources:     c.convertAffectedResources(output.AffectedResources),
						StandardProjectTarget: &projectCopy,
					}
					modifiers = append(modifiers, modifier)
				}
				continue
			}

			if len(output.AffectedTags) > 0 {
				for _, cardID := range handCardIDs {
					card, err := c.cardLookup.GetByID(cardID)
					if err != nil {
						continue
					}

					if c.cardHasMatchingTag(card, output.AffectedTags) {
						cardIDCopy := cardID
						modifier := shared.RequirementModifier{
							Amount:            output.Amount,
							AffectedResources: []shared.ResourceType{shared.ResourceCredit},
							CardTarget:        &cardIDCopy,
						}
						modifiers = append(modifiers, modifier)
					}
				}
				continue
			}

			// Case 3: Card type discount
			if len(output.AffectedCardTypes) > 0 {
				for _, cardID := range handCardIDs {
					card, err := c.cardLookup.GetByID(cardID)
					if err != nil {
						continue
					}

					if c.cardHasMatchingType(card, output.AffectedCardTypes) {
						cardIDCopy := cardID
						modifier := shared.RequirementModifier{
							Amount:            output.Amount,
							AffectedResources: []shared.ResourceType{shared.ResourceCredit},
							CardTarget:        &cardIDCopy,
						}
						modifiers = append(modifiers, modifier)
					}
				}
				continue
			}

			// Case 4: Global discount (applies to all cards in hand)
			for _, cardID := range handCardIDs {
				cardIDCopy := cardID
				modifier := shared.RequirementModifier{
					Amount:            output.Amount,
					AffectedResources: []shared.ResourceType{shared.ResourceCredit},
					CardTarget:        &cardIDCopy,
				}
				modifiers = append(modifiers, modifier)
			}
		}
	}

	return c.mergeModifiers(modifiers)
}

// convertAffectedResources converts string slice to ResourceType slice
func (c *RequirementModifierCalculator) convertAffectedResources(resources []string) []shared.ResourceType {
	if len(resources) == 0 {
		return []shared.ResourceType{shared.ResourceCredit} // Default to credits discount
	}
	result := make([]shared.ResourceType, len(resources))
	for i, r := range resources {
		result[i] = shared.ResourceType(r)
	}
	return result
}

// cardHasMatchingTag checks if a card has any of the specified tags
func (c *RequirementModifierCalculator) cardHasMatchingTag(card *Card, tags []shared.CardTag) bool {
	for _, cardTag := range card.Tags {
		for _, targetTag := range tags {
			if cardTag == targetTag {
				return true
			}
		}
	}
	return false
}

// cardHasMatchingType checks if a card matches any of the specified types
func (c *RequirementModifierCalculator) cardHasMatchingType(card *Card, types []string) bool {
	cardType := string(card.Type)
	for _, t := range types {
		if cardType == t {
			return true
		}
	}
	return false
}

// mergeModifiers combines modifiers targeting the same card/project by summing amounts
func (c *RequirementModifierCalculator) mergeModifiers(modifiers []shared.RequirementModifier) []shared.RequirementModifier {
	cardModifiers := make(map[string]*shared.RequirementModifier)
	projectModifiers := make(map[shared.StandardProject]*shared.RequirementModifier)

	for _, mod := range modifiers {
		if mod.CardTarget != nil {
			key := *mod.CardTarget
			if existing, ok := cardModifiers[key]; ok {
				existing.Amount += mod.Amount
			} else {
				modCopy := mod
				cardModifiers[key] = &modCopy
			}
		} else if mod.StandardProjectTarget != nil {
			key := *mod.StandardProjectTarget
			if existing, ok := projectModifiers[key]; ok {
				existing.Amount += mod.Amount
			} else {
				modCopy := mod
				projectModifiers[key] = &modCopy
			}
		}
	}

	result := make([]shared.RequirementModifier, 0, len(cardModifiers)+len(projectModifiers))
	for _, mod := range cardModifiers {
		result = append(result, *mod)
	}
	for _, mod := range projectModifiers {
		result = append(result, *mod)
	}
	return result
}

// CalculateCardDiscounts computes the total credit discount for a specific card.
// This is used during EntityState calculation instead of pre-computing all modifiers.
// Returns the total discount amount in credits that applies to this card.
func (c *RequirementModifierCalculator) CalculateCardDiscounts(p *player.Player, card *Card) int {
	if p == nil || card == nil {
		return 0
	}

	totalDiscount := 0
	effects := p.Effects().List()

	for _, effect := range effects {
		for _, output := range effect.Behavior.Outputs {
			if output.ResourceType != shared.ResourceDiscount {
				continue
			}

			hasCardTargets := len(output.AffectedTags) > 0 || len(output.AffectedCardTypes) > 0
			hasOnlyStandardProjectTargets := len(output.AffectedStandardProjects) > 0 && !hasCardTargets

			// Skip discounts that only target standard projects (not cards)
			if hasOnlyStandardProjectTargets {
				continue
			}

			// Check card-specific targeting (OR logic: tag match OR card type match)
			if hasCardTargets {
				// Check if card matches any of the targeting criteria
				matchesTags := len(output.AffectedTags) > 0 && c.cardHasMatchingTag(card, output.AffectedTags)
				matchesCardType := len(output.AffectedCardTypes) > 0 && c.cardHasMatchingType(card, output.AffectedCardTypes)

				if matchesTags || matchesCardType {
					totalDiscount += output.Amount
				}
				continue
			}

			// Global discount (no specific targets - applies to all cards)
			totalDiscount += output.Amount
		}
	}

	return totalDiscount
}

// CalculateStandardProjectDiscounts computes discounts for a specific standard project.
// Returns a map of resource type to discount amount.
// For example, Ecoline's discount on PlantGreenery returns {"plants": 1}.
func (c *RequirementModifierCalculator) CalculateStandardProjectDiscounts(
	p *player.Player,
	projectType shared.StandardProject,
) map[shared.ResourceType]int {
	discounts := make(map[shared.ResourceType]int)

	if p == nil {
		return discounts
	}

	effects := p.Effects().List()

	for _, effect := range effects {
		for _, output := range effect.Behavior.Outputs {
			if output.ResourceType != shared.ResourceDiscount {
				continue
			}

			// Only process if this discount affects this standard project
			projectMatches := false
			for _, proj := range output.AffectedStandardProjects {
				if proj == projectType {
					projectMatches = true
					break
				}
			}

			if !projectMatches {
				continue
			}

			// Determine which resources are discounted
			affectedResources := c.convertAffectedResources(output.AffectedResources)
			for _, resourceType := range affectedResources {
				discounts[resourceType] += output.Amount
			}
		}
	}

	return discounts
}
