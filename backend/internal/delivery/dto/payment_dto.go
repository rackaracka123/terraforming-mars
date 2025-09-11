package dto

// PaymentMethodDto represents different ways to pay for cards and actions
type PaymentMethodDto string

const (
	PaymentMethodCreditsDto  PaymentMethodDto = "credits"
	PaymentMethodSteelDto    PaymentMethodDto = "steel"
	PaymentMethodTitaniumDto PaymentMethodDto = "titanium"
)

// PaymentDto represents how a player wants to pay for something
type PaymentDto struct {
	Credits  int `json:"credits" ts:"number"`  // MegaCredits to spend
	Steel    int `json:"steel" ts:"number"`    // Steel to spend (2 MC discount per steel for buildings)
	Titanium int `json:"titanium" ts:"number"` // Titanium to spend (3 MC discount per titanium for space projects)
}

// PaymentCostDto represents the different ways something can be paid for
type PaymentCostDto struct {
	BaseCost       int  `json:"baseCost" ts:"number"`        // Base MegaCredit cost
	CanUseSteel    bool `json:"canUseSteel" ts:"boolean"`    // Can use steel for discount (building cards)
	CanUseTitanium bool `json:"canUseTitanium" ts:"boolean"` // Can use titanium for discount (space cards)
}
