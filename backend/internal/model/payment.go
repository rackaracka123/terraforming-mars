package model

// PaymentMethod represents different ways to pay for cards and actions
type PaymentMethod string

const (
	PaymentMethodCredits  PaymentMethod = "credits"
	PaymentMethodSteel    PaymentMethod = "steel"
	PaymentMethodTitanium PaymentMethod = "titanium"
)

// Payment represents how a player wants to pay for something
type Payment struct {
	Credits  int `json:"credits" ts:"number"`  // MegaCredits to spend
	Steel    int `json:"steel" ts:"number"`    // Steel to spend (2 MC discount per steel for buildings)
	Titanium int `json:"titanium" ts:"number"` // Titanium to spend (3 MC discount per titanium for space projects)
}

// PaymentCost represents the different ways something can be paid for
type PaymentCost struct {
	BaseCost       int  `json:"baseCost" ts:"number"`       // Base MegaCredit cost
	CanUseSteel    bool `json:"canUseSteel" ts:"boolean"`   // Can use steel for discount (building cards)
	CanUseTitanium bool `json:"canUseTitanium" ts:"boolean"` // Can use titanium for discount (space cards)
}