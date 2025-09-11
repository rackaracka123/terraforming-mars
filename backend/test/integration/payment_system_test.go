package integration

import (
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentSystem_Integration(t *testing.T) {
	// Create payment service
	paymentService := service.NewPaymentService()

	// Test player resources
	initialResources := model.Resources{
		Credits:  50,
		Steel:    10,
		Titanium: 5,
		Plants:   8,
		Energy:   6,
		Heat:     12,
	}

	t.Run("PaymentService Basic Operations", func(t *testing.T) {
		// Test card payment cost calculation
		buildingCard := createTestCard("building-card", 20, []model.CardTag{model.TagBuilding})
		spaceCard := createTestCard("space-card", 15, []model.CardTag{model.TagSpace})
		regularCard := createTestCard("regular-card", 10, []model.CardTag{})

		// Test building card (can use steel)
		buildingCost := paymentService.GetCardPaymentCost(&buildingCard)
		assert.Equal(t, 20, buildingCost.BaseCost)
		assert.True(t, buildingCost.CanUseSteel)
		assert.False(t, buildingCost.CanUseTitanium)

		// Test space card (can use titanium)
		spaceCost := paymentService.GetCardPaymentCost(&spaceCard)
		assert.Equal(t, 15, spaceCost.BaseCost)
		assert.False(t, spaceCost.CanUseSteel)
		assert.True(t, spaceCost.CanUseTitanium)

		// Test regular card (credits only)
		regularCost := paymentService.GetCardPaymentCost(&regularCard)
		assert.Equal(t, 10, regularCost.BaseCost)
		assert.False(t, regularCost.CanUseSteel)
		assert.False(t, regularCost.CanUseTitanium)
	})

	t.Run("Payment Validation", func(t *testing.T) {
		buildingCard := createTestCard("building-card", 20, []model.CardTag{model.TagBuilding})
		cost := paymentService.GetCardPaymentCost(&buildingCard)

		// Test valid payment with credits only
		creditPayment := &model.Payment{Credits: 20, Steel: 0, Titanium: 0}
		assert.True(t, paymentService.CanAfford(creditPayment, &initialResources))
		assert.True(t, paymentService.IsValidPayment(creditPayment, cost))

		// Test valid payment with steel discount
		steelPayment := &model.Payment{Credits: 10, Steel: 5, Titanium: 0} // 10 credits + 5 steel (10 MC discount) = 20 MC total
		assert.True(t, paymentService.CanAfford(steelPayment, &initialResources))
		assert.True(t, paymentService.IsValidPayment(steelPayment, cost))

		// Test invalid payment - trying to use titanium for building card
		titaniumPayment := &model.Payment{Credits: 5, Steel: 0, Titanium: 5}
		assert.True(t, paymentService.CanAfford(titaniumPayment, &initialResources))
		assert.False(t, paymentService.IsValidPayment(titaniumPayment, cost))

		// Test insufficient resources
		expensivePayment := &model.Payment{Credits: 100, Steel: 0, Titanium: 0}
		assert.False(t, paymentService.CanAfford(expensivePayment, &initialResources))

		// Test insufficient payment
		lowPayment := &model.Payment{Credits: 10, Steel: 0, Titanium: 0}
		assert.True(t, paymentService.CanAfford(lowPayment, &initialResources))
		assert.False(t, paymentService.IsValidPayment(lowPayment, cost))
	})

	t.Run("Payment Processing", func(t *testing.T) {
		// Test processing a valid payment
		payment := &model.Payment{Credits: 15, Steel: 2, Titanium: 1}
		newResources, err := paymentService.ProcessPayment(payment, &initialResources)

		require.NoError(t, err)
		assert.Equal(t, 35, newResources.Credits) // 50 - 15 = 35
		assert.Equal(t, 8, newResources.Steel)    // 10 - 2 = 8
		assert.Equal(t, 4, newResources.Titanium) // 5 - 1 = 4
		assert.Equal(t, initialResources.Plants, newResources.Plants)
		assert.Equal(t, initialResources.Energy, newResources.Energy)
		assert.Equal(t, initialResources.Heat, newResources.Heat)

		// Test processing invalid payment (insufficient resources)
		invalidPayment := &model.Payment{Credits: 100, Steel: 0, Titanium: 0}
		_, err = paymentService.ProcessPayment(invalidPayment, &initialResources)
		assert.Error(t, err)
	})

	t.Run("Card Play with Payments", func(t *testing.T) {
		// Create a test card and add it to player's hand
		buildingCard := createTestCard("test-building", 16, []model.CardTag{model.TagBuilding})

		// Add card to card data service (mock implementation would be needed for full test)
		// For now, test the service interface

		// Test different payment methods for the same card
		testCases := []struct {
			name    string
			payment *model.Payment
			valid   bool
		}{
			{
				name:    "Credits only",
				payment: &model.Payment{Credits: 16, Steel: 0, Titanium: 0},
				valid:   true,
			},
			{
				name:    "Mixed payment with steel",
				payment: &model.Payment{Credits: 10, Steel: 3, Titanium: 0}, // 10 + (3*2) = 16
				valid:   true,
			},
			{
				name:    "Optimal steel usage",
				payment: &model.Payment{Credits: 0, Steel: 8, Titanium: 0}, // 8*2 = 16
				valid:   true,
			},
			{
				name:    "Invalid titanium for building",
				payment: &model.Payment{Credits: 10, Steel: 0, Titanium: 2},
				valid:   false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cost := paymentService.GetCardPaymentCost(&buildingCard)
				canAfford := paymentService.CanAfford(tc.payment, &initialResources)
				isValid := paymentService.IsValidPayment(tc.payment, cost)

				if tc.valid {
					assert.True(t, canAfford, "Should be able to afford payment")
					assert.True(t, isValid, "Payment should be valid")

					// Test effective cost calculation
					effectiveCost := paymentService.GetEffectiveCost(tc.payment, cost)
					assert.Equal(t, tc.payment.Credits, effectiveCost, "Effective cost should match credits paid")
				} else {
					assert.False(t, isValid, "Payment should be invalid")
				}
			})
		}
	})

	t.Run("Standard Project Payments", func(t *testing.T) {
		// Test standard project cost structure (credits only)
		standardCost := &model.PaymentCost{
			BaseCost:       11,
			CanUseSteel:    false,
			CanUseTitanium: false,
		}

		// Test payment validation for standard projects
		validPayment := &model.Payment{Credits: 11, Steel: 0, Titanium: 0}
		assert.True(t, paymentService.IsValidPayment(validPayment, standardCost))

		invalidPayment := &model.Payment{Credits: 5, Steel: 3, Titanium: 0} // Trying to use steel
		assert.False(t, paymentService.IsValidPayment(invalidPayment, standardCost))
	})

	t.Run("Space Card with Titanium", func(t *testing.T) {
		spaceCard := createTestCard("space-station", 21, []model.CardTag{model.TagSpace})
		cost := paymentService.GetCardPaymentCost(&spaceCard)

		// Test various titanium payment combinations
		testCases := []struct {
			name    string
			payment *model.Payment
			valid   bool
		}{
			{
				name:    "Credits only",
				payment: &model.Payment{Credits: 21, Steel: 0, Titanium: 0},
				valid:   true,
			},
			{
				name:    "Mixed payment with titanium",
				payment: &model.Payment{Credits: 12, Steel: 0, Titanium: 3}, // 12 + (3*3) = 21
				valid:   true,
			},
			{
				name:    "All titanium",
				payment: &model.Payment{Credits: 0, Steel: 0, Titanium: 7}, // 7*3 = 21
				valid:   false,                                             // Player only has 5 titanium
			},
			{
				name:    "Invalid steel for space card",
				payment: &model.Payment{Credits: 15, Steel: 3, Titanium: 0},
				valid:   false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				canAfford := paymentService.CanAfford(tc.payment, &initialResources)
				isValid := paymentService.IsValidPayment(tc.payment, cost)

				if tc.valid {
					assert.True(t, canAfford, "Should be able to afford payment")
					assert.True(t, isValid, "Payment should be valid")
				} else {
					// Could be invalid due to insufficient resources or invalid method
					if canAfford {
						assert.False(t, isValid, "Payment should be invalid due to wrong payment method")
					} else {
						// Can't afford, so don't test validity
					}
				}
			})
		}
	})
}

// Helper functions

func createTestCard(id string, cost int, tags []model.CardTag) model.Card {
	return model.Card{
		ID:   id,
		Name: id,
		Cost: cost,
		Tags: tags,
		Type: model.CardTypeAutomated,
	}
}

// Helper functions can be added here as needed
