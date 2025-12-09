package validator

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/playability"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CanPlayCard checks if a player can play a card and returns detailed validation results.
// This orchestrates validation by checking game phase, turn state, hand state, requirements, and cost.
func CanPlayCard(card *gamecards.Card, g *game.Game, p *player.Player, cardRegistry cards.CardRegistry) playability.PlayabilityResult {
	result := playability.NewPlayabilityResult(true, nil)

	// Check if game is in action phase
	if g.CurrentPhase() != game.GamePhaseAction {
		result.AddError(playability.ValidationError{
			Type:          playability.ValidationErrorTypePhase,
			Message:       "Not in action phase",
			RequiredValue: game.GamePhaseAction,
			CurrentValue:  g.CurrentPhase(),
		})
	}

	// Check if it's player's turn
	currentTurn := g.CurrentTurn()
	if currentTurn == nil || currentTurn.PlayerID() != p.ID() {
		result.AddError(playability.ValidationError{
			Type:    playability.ValidationErrorTypeTurn,
			Message: "Not player's turn",
		})
	}

	// Check if card is in player's hand
	if !p.Hand().HasCard(card.ID) {
		result.AddError(playability.ValidationError{
			Type:    playability.ValidationErrorTypeGameState,
			Message: "Card not in player's hand",
		})
		// If card is not in hand, no point checking further
		return result
	}

	// Validate card requirements using shared validator
	requirementErrors := gamecards.ValidateCardRequirements(card, g, p, cardRegistry)
	for _, err := range requirementErrors {
		result.AddError(playability.ValidationError{
			Type:          playability.ValidationErrorTypeRequirement,
			Message:       err.Message,
			RequiredValue: err.RequiredValue,
			CurrentValue:  err.CurrentValue,
		})
	}

	// Validate payment affordability (basic cost check)
	validateCardCost(card, p, &result)

	return result
}

// validateCardCost validates that the player can afford the card's base cost
func validateCardCost(card *gamecards.Card, p *player.Player, result *playability.PlayabilityResult) {
	resources := p.Resources().Get()
	cost := card.Cost

	// Basic check: does player have enough credits to cover full cost?
	// (We don't check steel/titanium discounts here - that's more complex payment validation)
	if resources.Credits < cost {
		// Check if steel/titanium could help (for building/space cards)
		canUseSteel := cardHasTag(card, shared.TagBuilding) && resources.Steel > 0
		canUseTitanium := cardHasTag(card, shared.TagSpace) && resources.Titanium > 0

		if !canUseSteel && !canUseTitanium {
			result.AddError(playability.ValidationError{
				Type:          playability.ValidationErrorTypeCost,
				Message:       "Insufficient credits",
				RequiredValue: cost,
				CurrentValue:  resources.Credits,
			})
		}
		// If steel/titanium could help, consider it playable (actual payment validation happens during play)
	}
}

// cardHasTag checks if a card has a specific tag
func cardHasTag(card *gamecards.Card, tag shared.CardTag) bool {
	for _, cardTag := range card.Tags {
		if cardTag == tag {
			return true
		}
	}
	return false
}
