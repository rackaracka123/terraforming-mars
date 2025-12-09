package validator

import (
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/playability"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CanUseCardAction checks if a player can use a card action and returns which choices are affordable.
// This orchestrates validation by checking game phase, turn state, and resource affordability.
func CanUseCardAction(action *player.CardAction, p *player.Player, g *game.Game) playability.ActionPlayabilityResult {
	result := playability.NewActionPlayabilityResult()

	// Check if game is in action phase
	if g.CurrentPhase() != game.GamePhaseAction {
		result.Errors = append(result.Errors, playability.ValidationError{
			Type:          playability.ValidationErrorTypePhase,
			Message:       "Not in action phase",
			RequiredValue: game.GamePhaseAction,
			CurrentValue:  g.CurrentPhase(),
		})
		return result
	}

	// Check if it's player's turn
	currentTurn := g.CurrentTurn()
	if currentTurn == nil || currentTurn.PlayerID() != p.ID() {
		result.Errors = append(result.Errors, playability.ValidationError{
			Type:    playability.ValidationErrorTypeTurn,
			Message: "Not player's turn",
		})
		return result
	}

	// Note: Play count per generation check is handled by the backend action execution
	// We don't check it here for playability as it's a runtime constraint

	// Check if action has choices
	if len(action.Behavior.Choices) > 0 {
		// For choice-based actions, check each choice
		for choiceIndex, choice := range action.Behavior.Choices {
			choiceErrors := validateActionInputs(choice.Inputs, p)
			if len(choiceErrors) == 0 {
				result.AddPlayableChoice(choiceIndex)
			} else {
				result.AddUnaffordableChoice(choiceIndex, choiceErrors)
			}
		}
	} else {
		// For non-choice actions, check the main inputs
		inputErrors := validateActionInputs(action.Behavior.Inputs, p)
		if len(inputErrors) == 0 {
			result.IsAffordable = true
		} else {
			result.Errors = inputErrors
		}
	}

	return result
}

// validateActionInputs validates that a player has sufficient resources for the inputs.
// This is a primitive helper that checks resource affordability.
func validateActionInputs(inputs []shared.ResourceCondition, p *player.Player) []playability.ValidationError {
	errors := []playability.ValidationError{}
	resources := p.Resources().Get()
	storage := p.Resources().Storage()

	for _, input := range inputs {
		switch input.ResourceType {
		case shared.ResourceCredits:
			if resources.Credits < input.Amount {
				errors = append(errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeResource,
					Message:       "Insufficient credits",
					RequiredValue: input.Amount,
					CurrentValue:  resources.Credits,
				})
			}

		case shared.ResourceSteel:
			if resources.Steel < input.Amount {
				errors = append(errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeResource,
					Message:       "Insufficient steel",
					RequiredValue: input.Amount,
					CurrentValue:  resources.Steel,
				})
			}

		case shared.ResourceTitanium:
			if resources.Titanium < input.Amount {
				errors = append(errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeResource,
					Message:       "Insufficient titanium",
					RequiredValue: input.Amount,
					CurrentValue:  resources.Titanium,
				})
			}

		case shared.ResourcePlants:
			if resources.Plants < input.Amount {
				errors = append(errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeResource,
					Message:       "Insufficient plants",
					RequiredValue: input.Amount,
					CurrentValue:  resources.Plants,
				})
			}

		case shared.ResourceEnergy:
			if resources.Energy < input.Amount {
				errors = append(errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeResource,
					Message:       "Insufficient energy",
					RequiredValue: input.Amount,
					CurrentValue:  resources.Energy,
				})
			}

		case shared.ResourceHeat:
			if resources.Heat < input.Amount {
				errors = append(errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeResource,
					Message:       "Insufficient heat",
					RequiredValue: input.Amount,
					CurrentValue:  resources.Heat,
				})
			}

		default:
			// Check card storage resources (e.g., microbes, animals, floaters)
			storageKey := string(input.ResourceType)
			storedAmount, exists := storage[storageKey]
			if !exists || storedAmount < input.Amount {
				currentAmount := 0
				if exists {
					currentAmount = storedAmount
				}
				errors = append(errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeResource,
					Message:       "Insufficient " + storageKey,
					RequiredValue: input.Amount,
					CurrentValue:  currentAmount,
				})
			}
		}
	}

	return errors
}
