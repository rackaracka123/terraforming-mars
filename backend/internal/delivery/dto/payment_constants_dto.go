package dto

import "terraforming-mars-backend/internal/session/game/card"

// PaymentConstantsDto contains the conversion rates for alternative payment methods
// These values are sent to the frontend so it knows how much each resource is worth
type PaymentConstantsDto struct {
	SteelValue    int `json:"steelValue" ts:"number"`    // How many MC each steel is worth (2)
	TitaniumValue int `json:"titaniumValue" ts:"number"` // How many MC each titanium is worth (3)
}

// GetPaymentConstants returns the payment constants from the domain model
// This ensures DRY - the values are defined once in types.payment.go
func GetPaymentConstants() PaymentConstantsDto {
	return PaymentConstantsDto{
		SteelValue:    card.SteelValue,
		TitaniumValue: card.TitaniumValue,
	}
}
