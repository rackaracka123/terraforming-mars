package service

import (
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestPaymentService(t *testing.T) {
	paymentService := service.NewPaymentService()

	t.Run("GetCardPaymentCost", func(t *testing.T) {
		// Test building card
		buildingCard := &model.Card{
			ID:   "building-1",
			Cost: 15,
			Tags: []model.CardTag{model.TagBuilding},
		}

		buildingCost := paymentService.GetCardPaymentCost(buildingCard)
		assert.Equal(t, 15, buildingCost.BaseCost)
		assert.True(t, buildingCost.CanUseSteel)
		assert.False(t, buildingCost.CanUseTitanium)

		// Test space card
		spaceCard := &model.Card{
			ID:   "space-1",
			Cost: 20,
			Tags: []model.CardTag{model.TagSpace},
		}

		spaceCost := paymentService.GetCardPaymentCost(spaceCard)
		assert.Equal(t, 20, spaceCost.BaseCost)
		assert.False(t, spaceCost.CanUseSteel)
		assert.True(t, spaceCost.CanUseTitanium)

		// Test card with both tags
		hybridCard := &model.Card{
			ID:   "hybrid-1",
			Cost: 25,
			Tags: []model.CardTag{model.TagBuilding, model.TagSpace},
		}

		hybridCost := paymentService.GetCardPaymentCost(hybridCard)
		assert.Equal(t, 25, hybridCost.BaseCost)
		assert.True(t, hybridCost.CanUseSteel)
		assert.True(t, hybridCost.CanUseTitanium)

		// Test regular card (no special tags)
		regularCard := &model.Card{
			ID:   "regular-1",
			Cost: 12,
			Tags: []model.CardTag{model.TagScience},
		}

		regularCost := paymentService.GetCardPaymentCost(regularCard)
		assert.Equal(t, 12, regularCost.BaseCost)
		assert.False(t, regularCost.CanUseSteel)
		assert.False(t, regularCost.CanUseTitanium)
	})

	t.Run("CanAfford", func(t *testing.T) {
		playerResources := &model.Resources{
			Credits:  30,
			Steel:    5,
			Titanium: 3,
			Plants:   10,
			Energy:   8,
			Heat:     12,
		}

		// Test affordable payment
		affordablePayment := &model.Payment{
			Credits:  20,
			Steel:    3,
			Titanium: 2,
		}
		assert.True(t, paymentService.CanAfford(affordablePayment, playerResources))

		// Test unaffordable payment (too many credits)
		expensivePayment := &model.Payment{
			Credits:  50,
			Steel:    0,
			Titanium: 0,
		}
		assert.False(t, paymentService.CanAfford(expensivePayment, playerResources))

		// Test unaffordable payment (too much steel)
		steelPayment := &model.Payment{
			Credits:  10,
			Steel:    10,
			Titanium: 0,
		}
		assert.False(t, paymentService.CanAfford(steelPayment, playerResources))

		// Test unaffordable payment (too much titanium)
		titaniumPayment := &model.Payment{
			Credits:  10,
			Steel:    0,
			Titanium: 5,
		}
		assert.False(t, paymentService.CanAfford(titaniumPayment, playerResources))

		// Test edge case - exact resources
		exactPayment := &model.Payment{
			Credits:  30,
			Steel:    5,
			Titanium: 3,
		}
		assert.True(t, paymentService.CanAfford(exactPayment, playerResources))
	})

	t.Run("GetEffectiveCost", func(t *testing.T) {
		// Test building card cost with steel
		buildingCost := &model.PaymentCost{
			BaseCost:       20,
			CanUseSteel:    true,
			CanUseTitanium: false,
		}

		// Test credits only
		creditsPayment := &model.Payment{Credits: 20, Steel: 0, Titanium: 0}
		assert.Equal(t, 20, paymentService.GetEffectiveCost(creditsPayment, buildingCost))

		// Test steel discount
		steelPayment := &model.Payment{Credits: 10, Steel: 5, Titanium: 0} // 5 steel = 10 MC discount
		assert.Equal(t, 10, paymentService.GetEffectiveCost(steelPayment, buildingCost))

		// Test overpayment with steel (should cap at 0)
		overPayment := &model.Payment{Credits: 5, Steel: 10, Titanium: 0} // 10 steel = 20 MC discount
		assert.Equal(t, 0, paymentService.GetEffectiveCost(overPayment, buildingCost))

		// Test space card cost with titanium
		spaceCost := &model.PaymentCost{
			BaseCost:       21,
			CanUseSteel:    false,
			CanUseTitanium: true,
		}

		// Test titanium discount
		titaniumPayment := &model.Payment{Credits: 12, Steel: 0, Titanium: 3} // 3 titanium = 9 MC discount
		assert.Equal(t, 12, paymentService.GetEffectiveCost(titaniumPayment, spaceCost))

		// Test card that allows both steel and titanium
		hybridCost := &model.PaymentCost{
			BaseCost:       30,
			CanUseSteel:    true,
			CanUseTitanium: true,
		}

		hybridPayment := &model.Payment{Credits: 18, Steel: 3, Titanium: 2} // 3 steel (6 MC) + 2 titanium (6 MC) = 12 MC discount
		assert.Equal(t, 18, paymentService.GetEffectiveCost(hybridPayment, hybridCost))
	})

	t.Run("IsValidPayment", func(t *testing.T) {
		// Building card that allows steel
		buildingCost := &model.PaymentCost{
			BaseCost:       16,
			CanUseSteel:    true,
			CanUseTitanium: false,
		}

		// Valid payment with credits only
		validCredits := &model.Payment{Credits: 16, Steel: 0, Titanium: 0}
		assert.True(t, paymentService.IsValidPayment(validCredits, buildingCost))

		// Valid payment with steel
		validSteel := &model.Payment{Credits: 10, Steel: 3, Titanium: 0} // 10 + (3*2) = 16
		assert.True(t, paymentService.IsValidPayment(validSteel, buildingCost))

		// Invalid payment - trying to use titanium
		invalidTitanium := &model.Payment{Credits: 10, Steel: 0, Titanium: 2}
		assert.False(t, paymentService.IsValidPayment(invalidTitanium, buildingCost))

		// Invalid payment - insufficient credits after discount
		insufficient := &model.Payment{Credits: 5, Steel: 3, Titanium: 0} // 5 + (3*2) = 11 < 16
		assert.False(t, paymentService.IsValidPayment(insufficient, buildingCost))

		// Space card that allows titanium
		spaceCost := &model.PaymentCost{
			BaseCost:       18,
			CanUseSteel:    false,
			CanUseTitanium: true,
		}

		// Valid payment with titanium
		validTitaniumSpace := &model.Payment{Credits: 9, Steel: 0, Titanium: 3} // 9 + (3*3) = 18
		assert.True(t, paymentService.IsValidPayment(validTitaniumSpace, spaceCost))

		// Invalid payment - trying to use steel for space card
		invalidSteelSpace := &model.Payment{Credits: 10, Steel: 4, Titanium: 0}
		assert.False(t, paymentService.IsValidPayment(invalidSteelSpace, spaceCost))
	})

	t.Run("ProcessPayment", func(t *testing.T) {
		initialResources := &model.Resources{
			Credits:  40,
			Steel:    8,
			Titanium: 6,
			Plants:   15,
			Energy:   10,
			Heat:     20,
		}

		// Test successful payment processing
		payment := &model.Payment{
			Credits:  15,
			Steel:    3,
			Titanium: 2,
		}

		newResources, err := paymentService.ProcessPayment(payment, initialResources)
		assert.NoError(t, err)
		assert.Equal(t, 25, newResources.Credits)                     // 40 - 15 = 25
		assert.Equal(t, 5, newResources.Steel)                        // 8 - 3 = 5
		assert.Equal(t, 4, newResources.Titanium)                     // 6 - 2 = 4
		assert.Equal(t, initialResources.Plants, newResources.Plants) // Unchanged
		assert.Equal(t, initialResources.Energy, newResources.Energy) // Unchanged
		assert.Equal(t, initialResources.Heat, newResources.Heat)     // Unchanged

		// Test payment with insufficient resources
		expensivePayment := &model.Payment{
			Credits:  50,
			Steel:    0,
			Titanium: 0,
		}

		_, err = paymentService.ProcessPayment(expensivePayment, initialResources)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient resources")

		// Test payment with insufficient steel
		steelPayment := &model.Payment{
			Credits:  10,
			Steel:    15,
			Titanium: 0,
		}

		_, err = paymentService.ProcessPayment(steelPayment, initialResources)
		assert.Error(t, err)

		// Test payment with insufficient titanium
		titaniumPayment := &model.Payment{
			Credits:  10,
			Steel:    0,
			Titanium: 10,
		}

		_, err = paymentService.ProcessPayment(titaniumPayment, initialResources)
		assert.Error(t, err)
	})
}
