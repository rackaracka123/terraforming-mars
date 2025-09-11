package service

import (
	"fmt"
	"terraforming-mars-backend/internal/model"
)

// PaymentService handles payment validation and processing logic
type PaymentService interface {
	// CanAfford checks if a player can afford a payment with their current resources
	CanAfford(payment *model.Payment, playerResources *model.Resources) bool

	// IsValidPayment checks if a payment is valid for the given cost requirements
	IsValidPayment(payment *model.Payment, cost *model.PaymentCost) bool

	// GetEffectiveCost calculates the actual MegaCredits needed after applying discounts
	GetEffectiveCost(payment *model.Payment, cost *model.PaymentCost) int

	// ProcessPayment deducts resources from player for a payment
	ProcessPayment(payment *model.Payment, playerResources *model.Resources) (*model.Resources, error)

	// GetCardPaymentCost determines the payment options for a specific card based on its tags
	GetCardPaymentCost(card *model.Card) *model.PaymentCost
}

// PaymentServiceImpl implements PaymentService interface
type PaymentServiceImpl struct{}

// NewPaymentService creates a new PaymentService instance
func NewPaymentService() PaymentService {
	return &PaymentServiceImpl{}
}

// GetCardPaymentCost determines the payment options for a specific card based on its tags
func (ps *PaymentServiceImpl) GetCardPaymentCost(card *model.Card) *model.PaymentCost {
	cost := &model.PaymentCost{
		BaseCost:       card.Cost,
		CanUseSteel:    false,
		CanUseTitanium: false,
	}

	// Check card tags to determine payment methods
	for _, tag := range card.Tags {
		switch tag {
		case model.TagBuilding:
			cost.CanUseSteel = true
		case model.TagSpace:
			cost.CanUseTitanium = true
		}
	}

	return cost
}

// CanAfford checks if a player can afford the payment with their current resources
func (ps *PaymentServiceImpl) CanAfford(payment *model.Payment, playerResources *model.Resources) bool {
	return playerResources.Credits >= payment.Credits &&
		playerResources.Steel >= payment.Steel &&
		playerResources.Titanium >= payment.Titanium
}

// IsValidPayment checks if this payment is valid for the given cost requirements
func (ps *PaymentServiceImpl) IsValidPayment(payment *model.Payment, cost *model.PaymentCost) bool {
	// Check if trying to use steel when not allowed
	if payment.Steel > 0 && !cost.CanUseSteel {
		return false
	}

	// Check if trying to use titanium when not allowed
	if payment.Titanium > 0 && !cost.CanUseTitanium {
		return false
	}

	// Check if the effective cost is covered
	effectiveCost := ps.GetEffectiveCost(payment, cost)
	return payment.Credits >= effectiveCost
}

// GetEffectiveCost calculates the actual MegaCredits needed after applying discounts
func (ps *PaymentServiceImpl) GetEffectiveCost(payment *model.Payment, cost *model.PaymentCost) int {
	effectiveCost := cost.BaseCost

	if cost.CanUseSteel {
		effectiveCost -= payment.Steel * 2 // Steel provides 2 MC discount per unit
	}

	if cost.CanUseTitanium {
		effectiveCost -= payment.Titanium * 3 // Titanium provides 3 MC discount per unit
	}

	if effectiveCost < 0 {
		effectiveCost = 0
	}

	return effectiveCost
}

// ProcessPayment deducts resources from player for a payment
func (ps *PaymentServiceImpl) ProcessPayment(payment *model.Payment, playerResources *model.Resources) (*model.Resources, error) {
	// Validate the player has enough resources
	if !ps.CanAfford(payment, playerResources) {
		return nil, fmt.Errorf("insufficient resources for payment")
	}

	// Create a copy of resources and deduct payment amounts
	newResources := *playerResources
	newResources.Credits -= payment.Credits
	newResources.Steel -= payment.Steel
	newResources.Titanium -= payment.Titanium

	return &newResources, nil
}
