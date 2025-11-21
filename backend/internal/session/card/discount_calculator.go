package card

import (
	"terraforming-mars-backend/internal/session/types"
)

const (
	// ResourceConversionDiscountType is the discount type for resource conversion actions
	ResourceConversionDiscountType = "discount"
)

// CalculateResourceConversionCost calculates the cost for resource conversions (plants→greenery, heat→temp)
// considering discount effects from the player's active effects.
//
// Parameters:
//   - player: The player attempting the conversion
//   - conversionType: The standard project type ("convert-plants-to-greenery" or "convert-heat-to-temperature")
//   - baseCost: The base cost before discounts (typically 8)
//
// Returns the final cost after applying all applicable discounts (minimum 1).
func CalculateResourceConversionCost(player *types.Player, conversionType types.StandardProject, baseCost int) int {
	if player == nil || player.Effects == nil {
		return baseCost
	}

	totalDiscount := 0

	// Loop through all player effects
	for _, effect := range player.Effects {
		// Check each output in the effect's behavior
		for _, output := range effect.Behavior.Outputs {
			// Check if this is a discount effect
			if output.Type != types.ResourceDiscount {
				continue
			}

			// Check if this discount applies to our conversion type
			if !containsStandardProject(output.AffectedStandardProjects, conversionType) {
				continue
			}

			// Add the discount amount
			totalDiscount += output.Amount
		}
	}

	// Calculate final cost (never below 1)
	finalCost := baseCost - totalDiscount
	if finalCost < 1 {
		finalCost = 1
	}

	return finalCost
}

// containsStandardProject checks if a slice contains a specific standard project
func containsStandardProject(projects []types.StandardProject, target types.StandardProject) bool {
	for _, project := range projects {
		if project == target {
			return true
		}
	}
	return false
}
