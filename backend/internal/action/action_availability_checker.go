package action

import (
	"fmt"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/playability"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CreateActionAvailabilityChecker creates a closure that checks if a card action is available
// The closure captures game and player references for availability checking
// This is shared logic used by PlayCardAction, SelectStartingCardsAction, and admin actions
func CreateActionAvailabilityChecker(
	g *game.Game,
	p *player.Player,
	card *gamecards.Card,
	behaviorIndex int,
	cardRegistry cards.CardRegistry,
) player.ActionAvailabilityChecker {
	return func() playability.ActionPlayabilityResult {
		result := playability.NewActionPlayabilityResult()

		// Get the specific behavior
		if behaviorIndex < 0 || behaviorIndex >= len(card.Behaviors) {
			result.Errors = append(result.Errors, playability.ValidationError{
				Type:    playability.ValidationErrorTypeGameState,
				Message: fmt.Sprintf("Behavior index %d out of range", behaviorIndex),
			})
			return result
		}

		behavior := card.Behaviors[behaviorIndex]

		// Check if player can afford the inputs
		resources := p.Resources().Get()
		canAfford := true

		for _, input := range behavior.Inputs {
			inputType := input.ResourceType
			amount := input.Amount

			var available int
			switch inputType {
			case shared.ResourceCredits:
				available = resources.Credits
			case shared.ResourceSteel:
				available = resources.Steel
			case shared.ResourceTitanium:
				available = resources.Titanium
			case shared.ResourcePlants:
				available = resources.Plants
			case shared.ResourceEnergy:
				available = resources.Energy
			case shared.ResourceHeat:
				available = resources.Heat
			default:
				// Unknown resource type
				canAfford = false
				result.Errors = append(result.Errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeResource,
					Message:       fmt.Sprintf("Unknown resource type: %s", inputType),
					RequiredValue: amount,
					CurrentValue:  0,
				})
				continue
			}

			if available < amount {
				canAfford = false
				result.Errors = append(result.Errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeCost,
					Message:       fmt.Sprintf("Insufficient %s", inputType),
					RequiredValue: amount,
					CurrentValue:  available,
				})
			}
		}

		result.IsAffordable = canAfford
		if canAfford {
			result.PlayableChoices = []int{0} // Simple actions have one choice
		}

		return result
	}
}
