package cards

import (
	"fmt"

	"terraforming-mars-backend/internal/game/shared"
)

// ==================== Card Payment ====================

// CardPayment represents how a player is paying for a card
type CardPayment struct {
	Credits     int
	Steel       int
	Titanium    int
	Substitutes map[shared.ResourceType]int
}

// Payment method constants
const (
	SteelValue    = 2
	TitaniumValue = 3
)

// TotalValue calculates the total MC value of this payment
func (p CardPayment) TotalValue(playerSubstitutes []shared.PaymentSubstitute) int {
	total := p.Credits + (p.Steel * SteelValue) + (p.Titanium * TitaniumValue)

	if p.Substitutes != nil {
		for resourceType, amount := range p.Substitutes {
			for _, sub := range playerSubstitutes {
				if sub.ResourceType == resourceType {
					total += amount * sub.ConversionRate
					break
				}
			}
		}
	}

	return total
}

// Validate checks if the payment is valid
func (p CardPayment) Validate() error {
	if p.Credits < 0 {
		return fmt.Errorf("payment credits cannot be negative: %d", p.Credits)
	}
	if p.Steel < 0 {
		return fmt.Errorf("payment steel cannot be negative: %d", p.Steel)
	}
	if p.Titanium < 0 {
		return fmt.Errorf("payment titanium cannot be negative: %d", p.Titanium)
	}

	if p.Substitutes != nil {
		for resourceType, amount := range p.Substitutes {
			if amount < 0 {
				return fmt.Errorf("payment substitute %s cannot be negative: %d", resourceType, amount)
			}
		}
	}

	return nil
}

// CanAfford checks if a player has sufficient resources for this payment
func (p CardPayment) CanAfford(playerResources shared.Resources) error {
	if playerResources.Credits < p.Credits {
		return fmt.Errorf("insufficient credits: need %d, have %d", p.Credits, playerResources.Credits)
	}
	if playerResources.Steel < p.Steel {
		return fmt.Errorf("insufficient steel: need %d, have %d", p.Steel, playerResources.Steel)
	}
	if playerResources.Titanium < p.Titanium {
		return fmt.Errorf("insufficient titanium: need %d, have %d", p.Titanium, playerResources.Titanium)
	}

	if p.Substitutes != nil {
		for resourceType, amount := range p.Substitutes {
			var available int
			switch resourceType {
			case shared.ResourceHeat:
				available = playerResources.Heat
			case shared.ResourceEnergy:
				available = playerResources.Energy
			case shared.ResourcePlant:
				available = playerResources.Plants
			default:
				return fmt.Errorf("unsupported payment substitute resource type: %s", resourceType)
			}

			if available < amount {
				return fmt.Errorf("insufficient %s: need %d, have %d", resourceType, amount, available)
			}
		}
	}

	return nil
}

// CoversCardCost checks if this payment covers the card cost
func (p CardPayment) CoversCardCost(cardCost int, allowSteel, allowTitanium bool, playerSubstitutes []shared.PaymentSubstitute) error {
	if err := p.Validate(); err != nil {
		return err
	}

	if p.Steel > 0 && !allowSteel {
		return fmt.Errorf("card does not have building tag, cannot use steel")
	}
	if p.Titanium > 0 && !allowTitanium {
		return fmt.Errorf("card does not have space tag, cannot use titanium")
	}

	if p.Substitutes != nil {
		for resourceType := range p.Substitutes {
			found := false
			for _, sub := range playerSubstitutes {
				if sub.ResourceType == resourceType {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("player cannot use %s as payment substitute", resourceType)
			}
		}
	}

	totalValue := p.TotalValue(playerSubstitutes)
	if totalValue < cardCost {
		return fmt.Errorf("payment insufficient: card costs %d MC, payment provides %d MC", cardCost, totalValue)
	}

	return nil
}
