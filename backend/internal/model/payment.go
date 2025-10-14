package model

import "fmt"

// CardPayment represents how a player is paying for a card
type CardPayment struct {
	Credits  int `json:"credits" ts:"number"`  // MC spent
	Steel    int `json:"steel" ts:"number"`    // Steel resources used (2 MC value each)
	Titanium int `json:"titanium" ts:"number"` // Titanium resources used (3 MC value each)
}

// PaymentMethod defines conversion rates for alternative payment resources
const (
	SteelValue    = 2 // Each steel is worth 2 MC
	TitaniumValue = 3 // Each titanium is worth 3 MC
)

// TotalValue calculates the total MC value of this payment
func (p CardPayment) TotalValue() int {
	return p.Credits + (p.Steel * SteelValue) + (p.Titanium * TitaniumValue)
}

// Validate checks if the payment is valid (no negative values)
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
	return nil
}

// CanAfford checks if a player has sufficient resources for this payment
func (p CardPayment) CanAfford(playerResources Resources) error {
	if playerResources.Credits < p.Credits {
		return fmt.Errorf("insufficient credits: need %d, have %d", p.Credits, playerResources.Credits)
	}
	if playerResources.Steel < p.Steel {
		return fmt.Errorf("insufficient steel: need %d, have %d", p.Steel, playerResources.Steel)
	}
	if playerResources.Titanium < p.Titanium {
		return fmt.Errorf("insufficient titanium: need %d, have %d", p.Titanium, playerResources.Titanium)
	}
	return nil
}

// CoversCardCost checks if this payment covers the card cost, considering allowed payment methods
func (p CardPayment) CoversCardCost(cardCost int, allowSteel, allowTitanium bool) error {
	// Validate payment format
	if err := p.Validate(); err != nil {
		return err
	}

	// Check if steel/titanium are used when not allowed
	if p.Steel > 0 && !allowSteel {
		return fmt.Errorf("card does not have building tag, cannot use steel")
	}
	if p.Titanium > 0 && !allowTitanium {
		return fmt.Errorf("card does not have space tag, cannot use titanium")
	}

	// Check if payment total covers card cost
	totalValue := p.TotalValue()
	if totalValue < cardCost {
		return fmt.Errorf("payment insufficient: card costs %d MC, payment provides %d MC", cardCost, totalValue)
	}

	// Check for overpayment (payment should be exact)
	if totalValue > cardCost {
		return fmt.Errorf("payment exceeds card cost: card costs %d MC, payment provides %d MC", cardCost, totalValue)
	}

	return nil
}

// CalculateMinimumAlternativeResources calculates the minimum steel or titanium needed when MC is insufficient
// Returns (minSteel, minTitanium) - at least one will be 0 if the card doesn't support that payment type
func CalculateMinimumAlternativeResources(cardCost int, playerResources Resources, allowSteel, allowTitanium bool) (minSteel int, minTitanium int) {
	if cardCost <= 0 {
		return 0, 0
	}

	// If player has enough credits, no alternative payment needed
	if playerResources.Credits >= cardCost {
		return 0, 0
	}

	shortfall := cardCost - playerResources.Credits

	// Calculate minimum steel needed (if building tag)
	if allowSteel {
		// Each steel is worth 2 MC, so divide by 2 and round up
		minSteel = (shortfall + SteelValue - 1) / SteelValue
	}

	// Calculate minimum titanium needed (if space tag)
	if allowTitanium {
		// Each titanium is worth 3 MC, so divide by 3 and round up
		minTitanium = (shortfall + TitaniumValue - 1) / TitaniumValue
	}

	return minSteel, minTitanium
}
