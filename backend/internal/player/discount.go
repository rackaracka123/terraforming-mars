package player

import (
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/features/card"
)

// ============================================================================
// DISCOUNT CALCULATION UTILITIES
// ============================================================================

// CalculateCardDiscounts calculates the total discount for a card based on player's played cards
// This examines all played cards for discount effects that apply to the target card's tags
func CalculateCardDiscounts(player *Player, targetCard card.Card) int {
	if player == nil {
		return 0
	}

	totalDiscount := 0

	// Check all played cards for discount effects
	for _, playedCard := range player.PlayedCards {
		// Check if this played card has any behaviors that provide discounts
		for _, behavior := range playedCard.Behaviors {
			// Check if behavior has discount effects
			for _, effect := range behavior.Effects {
				if effect.Type == string(card.EffectDiscountByTag) {
					// Check if discount applies to the target card
					if discountApplies(effect, targetCard) {
						totalDiscount += effect.Amount
					}
				}
			}
		}
	}

	// Also check for corporation-specific discounts
	if player.Corporation != nil {
		for _, behavior := range player.Corporation.Behaviors {
			for _, effect := range behavior.Effects {
				if effect.Type == string(card.EffectDiscountByTag) {
					if discountApplies(effect, targetCard) {
						totalDiscount += effect.Amount
					}
				}
			}
		}
	}

	return totalDiscount
}

// discountApplies checks if a discount effect applies to the target card
func discountApplies(effect card.Effect, targetCard card.Card) bool {
	// Check if target card has the required tag specified by the effect
	if effect.Tag == "" {
		return false
	}

	// Check if target card has the required tag
	for _, cardTag := range targetCard.Tags {
		if string(cardTag) == effect.Tag {
			return true
		}
	}

	return false
}

// CalculateStandardProjectDiscount calculates discount for standard projects
// Some cards provide discounts to specific standard projects
// TODO: Implement when standard project discount effects are defined in card data
func CalculateStandardProjectDiscount(player *Player, projectType StandardProject) int {
	if player == nil {
		return 0
	}

	// Standard project discounts are rare in the base game
	// This would need to check card effects similar to CalculateCardDiscounts
	// but matching against standard project types rather than card tags
	return 0
}

// ApplyCardCostModifiers applies all cost modifiers to a card's base cost
// This includes both discounts and increases
func ApplyCardCostModifiers(baseCardCost int, modifiers []card.CostModifier) int {
	finalCost := baseCardCost

	for _, modifier := range modifiers {
		finalCost += modifier.Amount // Amount is negative for discounts
	}

	// Card cost cannot go below 0
	if finalCost < 0 {
		finalCost = 0
	}

	return finalCost
}

// CalculateResourceDiscount calculates how much discount applies to resource-based payments
// For example, Helion corporation allows using heat as credits for standard projects
func CalculateResourceDiscount(player *Player, resourceType domain.ResourceType) int {
	if player == nil || player.PaymentSubstitutes == nil {
		return 0
	}

	// Check if player has a payment substitute for this resource type
	for _, substitute := range player.PaymentSubstitutes {
		if substitute.ResourceType == resourceType {
			return substitute.ConversionRate
		}
	}

	return 0
}
