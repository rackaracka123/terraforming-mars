package player

import (
	"fmt"

	"terraforming-mars-backend/internal/domain"
)

// ============================================================================
// RESOURCES & PRODUCTION MODELS
// ============================================================================

// Resources represents a player's resources
type Resources struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// IsZero returns true if all resource values are zero
func (r Resources) IsZero() bool {
	return r.Credits == 0 && r.Steel == 0 && r.Titanium == 0 &&
		r.Plants == 0 && r.Energy == 0 && r.Heat == 0
}

// DeepCopy creates a deep copy of the Resources struct
func (r Resources) DeepCopy() Resources {
	return Resources{
		Credits:  r.Credits,
		Steel:    r.Steel,
		Titanium: r.Titanium,
		Plants:   r.Plants,
		Energy:   r.Energy,
		Heat:     r.Heat,
	}
}

// Production represents a player's production values
type Production struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// DeepCopy creates a deep copy of the Production struct
func (p Production) DeepCopy() Production {
	return Production{
		Credits:  p.Credits,
		Steel:    p.Steel,
		Titanium: p.Titanium,
		Plants:   p.Plants,
		Energy:   p.Energy,
		Heat:     p.Heat,
	}
}

// ============================================================================
// PAYMENT & SUBSTITUTES
// ============================================================================

// PaymentSubstitute represents an alternative resource that can be used as payment for credits
// Example: Helion allows using heat as M€ with 1:1 conversion
type PaymentSubstitute struct {
	ResourceType   domain.ResourceType `json:"resourceType" ts:"ResourceType"` // The resource that can be used (e.g., heat)
	ConversionRate int                 `json:"conversionRate" ts:"number"`     // How many credits each resource is worth (1 = 1:1)
}

// CardPayment represents how a player is paying for a card
type CardPayment struct {
	Credits     int                         `json:"credits" ts:"number"`                                           // MC spent
	Steel       int                         `json:"steel" ts:"number"`                                             // Steel resources used (2 MC value each)
	Titanium    int                         `json:"titanium" ts:"number"`                                          // Titanium resources used (3 MC value each)
	Substitutes map[domain.ResourceType]int `json:"substitutes,omitempty" ts:"Record<string, number> | undefined"` // Payment substitutes (e.g., heat for Helion) with conversion rates from player
}

// PaymentMethod defines conversion rates for alternative payment resources
const (
	SteelValue    = 2 // Each steel is worth 2 MC
	TitaniumValue = 3 // Each titanium is worth 3 MC
)

// TotalValue calculates the total MC value of this payment
// playerSubstitutes is needed to get conversion rates for substitute payments
func (p CardPayment) TotalValue(playerSubstitutes []PaymentSubstitute) int {
	total := p.Credits + (p.Steel * SteelValue) + (p.Titanium * TitaniumValue)

	// Add value from payment substitutes
	if p.Substitutes != nil {
		for resourceType, amount := range p.Substitutes {
			// Find the conversion rate for this resource type
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

	// Validate substitutes (if any)
	if p.Substitutes != nil {
		for resourceType, amount := range p.Substitutes {
			if amount < 0 {
				return fmt.Errorf("payment substitute %s cannot be negative: %d", resourceType, amount)
			}
		}
	}

	return nil
}

// CoversCardCost checks if this payment covers the card cost, considering allowed payment methods
// playerSubstitutes are the player's available payment substitutes with conversion rates
func (p CardPayment) CoversCardCost(cardCost int, allowSteel, allowTitanium bool, playerSubstitutes []PaymentSubstitute) error {
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

	// Validate that substitutes used are in player's allowed substitutes
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

	// Check if payment total covers card cost (overpayment is allowed, excess is wasted)
	totalValue := p.TotalValue(playerSubstitutes)
	if totalValue < cardCost {
		return fmt.Errorf("payment insufficient: card costs %d MC, payment provides %d MC", cardCost, totalValue)
	}

	// Note: Overpayment is allowed in Terraforming Mars - excess resources are simply wasted
	// This matches the board game rules where you can't always make exact change with steel/titanium

	return nil
}

// ============================================================================
// STANDARD PROJECTS
// ============================================================================

// StandardProject represents the different types of standard projects available to players
type StandardProject string

const (
	// Standard Projects (M€-based)
	StandardProjectSellPatents StandardProject = "sell-patents"
	StandardProjectPowerPlant  StandardProject = "power-plant"
	StandardProjectAsteroid    StandardProject = "asteroid"
	StandardProjectAquifer     StandardProject = "aquifer"
	StandardProjectGreenery    StandardProject = "greenery"
	StandardProjectCity        StandardProject = "city"
	// Resource Conversion Actions (resource-based, not M€)
	StandardProjectConvertPlantsToGreenery  StandardProject = "convert-plants-to-greenery"
	StandardProjectConvertHeatToTemperature StandardProject = "convert-heat-to-temperature"
)

// StandardProjectCost represents the credit cost for each standard project
var StandardProjectCost = map[StandardProject]int{
	StandardProjectSellPatents: 0,  // No cost - player gains M€ instead
	StandardProjectPowerPlant:  11, // 11 M€
	StandardProjectAsteroid:    14, // 14 M€
	StandardProjectAquifer:     18, // 18 M€
	StandardProjectGreenery:    23, // 23 M€
	StandardProjectCity:        25, // 25 M€
}
