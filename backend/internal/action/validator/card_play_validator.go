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

	// Validate card requirements
	validateCardRequirements(card, g, p, cardRegistry, &result)

	// Validate payment affordability (basic cost check)
	validateCardCost(card, p, &result)

	return result
}

// validateCardRequirements validates all card requirements using domain primitives
func validateCardRequirements(card *gamecards.Card, g *game.Game, p *player.Player, cardRegistry cards.CardRegistry, result *playability.PlayabilityResult) {
	if len(card.Requirements) == 0 {
		return
	}

	for _, req := range card.Requirements {
		switch req.Type {
		case gamecards.RequirementTemperature:
			temp := g.GlobalParameters().Temperature()
			if req.Min != nil && temp < *req.Min {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Temperature requirement not met",
					RequiredValue: *req.Min,
					CurrentValue:  temp,
				})
			}
			if req.Max != nil && temp > *req.Max {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Temperature exceeds maximum",
					RequiredValue: *req.Max,
					CurrentValue:  temp,
				})
			}

		case gamecards.RequirementOxygen:
			oxygen := g.GlobalParameters().Oxygen()
			if req.Min != nil && oxygen < *req.Min {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Oxygen requirement not met",
					RequiredValue: *req.Min,
					CurrentValue:  oxygen,
				})
			}
			if req.Max != nil && oxygen > *req.Max {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Oxygen exceeds maximum",
					RequiredValue: *req.Max,
					CurrentValue:  oxygen,
				})
			}

		case gamecards.RequirementOceans:
			oceans := g.GlobalParameters().Oceans()
			if req.Min != nil && oceans < *req.Min {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Oceans requirement not met",
					RequiredValue: *req.Min,
					CurrentValue:  oceans,
				})
			}
			if req.Max != nil && oceans > *req.Max {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Oceans exceeds maximum",
					RequiredValue: *req.Max,
					CurrentValue:  oceans,
				})
			}

		case gamecards.RequirementTR:
			tr := p.Resources().TerraformRating()
			if req.Min != nil && tr < *req.Min {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Terraform rating requirement not met",
					RequiredValue: *req.Min,
					CurrentValue:  tr,
				})
			}
			if req.Max != nil && tr > *req.Max {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Terraform rating exceeds maximum",
					RequiredValue: *req.Max,
					CurrentValue:  tr,
				})
			}

		case gamecards.RequirementTags:
			if req.Tag == nil {
				continue
			}

			// Count tags across all played cards (including corporation)
			tagCount := 0
			for _, playedCardID := range p.PlayedCards().Cards() {
				playedCard, err := cardRegistry.GetByID(playedCardID)
				if err != nil {
					continue
				}
				if hasTag(playedCard, *req.Tag) {
					tagCount++
				}
			}

			// Also count corporation tags
			if corpID := p.CorporationID(); corpID != "" {
				corpCard, err := cardRegistry.GetByID(corpID)
				if err == nil && hasTag(corpCard, *req.Tag) {
					tagCount++
				}
			}

			if req.Min != nil && tagCount < *req.Min {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Tag requirement not met",
					RequiredValue: *req.Min,
					CurrentValue:  tagCount,
				})
			}
			if req.Max != nil && tagCount > *req.Max {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Tag count exceeds maximum",
					RequiredValue: *req.Max,
					CurrentValue:  tagCount,
				})
			}

		case gamecards.RequirementProduction:
			if req.Resource == nil {
				continue
			}
			production := p.Resources().Production()
			var currentProduction int

			switch *req.Resource {
			case shared.ResourceCreditsProduction:
				currentProduction = production.Credits
			case shared.ResourceSteelProduction:
				currentProduction = production.Steel
			case shared.ResourceTitaniumProduction:
				currentProduction = production.Titanium
			case shared.ResourcePlantsProduction:
				currentProduction = production.Plants
			case shared.ResourceEnergyProduction:
				currentProduction = production.Energy
			case shared.ResourceHeatProduction:
				currentProduction = production.Heat
			}

			if req.Min != nil && currentProduction < *req.Min {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Production requirement not met",
					RequiredValue: *req.Min,
					CurrentValue:  currentProduction,
				})
			}
			if req.Max != nil && currentProduction > *req.Max {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Production exceeds maximum",
					RequiredValue: *req.Max,
					CurrentValue:  currentProduction,
				})
			}

		case gamecards.RequirementResource:
			if req.Resource == nil {
				continue
			}
			resources := p.Resources().Get()
			var currentAmount int

			switch *req.Resource {
			case shared.ResourceCredits:
				currentAmount = resources.Credits
			case shared.ResourceSteel:
				currentAmount = resources.Steel
			case shared.ResourceTitanium:
				currentAmount = resources.Titanium
			case shared.ResourcePlants:
				currentAmount = resources.Plants
			case shared.ResourceEnergy:
				currentAmount = resources.Energy
			case shared.ResourceHeat:
				currentAmount = resources.Heat
			}

			if req.Min != nil && currentAmount < *req.Min {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Resource requirement not met",
					RequiredValue: *req.Min,
					CurrentValue:  currentAmount,
				})
			}
			if req.Max != nil && currentAmount > *req.Max {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Resource exceeds maximum",
					RequiredValue: *req.Max,
					CurrentValue:  currentAmount,
				})
			}

		case gamecards.RequirementCities:
			// Count cities owned by the player on the board
			cityCount := 0
			for _, tile := range g.Board().Tiles() {
				if tile.OccupiedBy != nil && tile.OccupiedBy.Type == shared.ResourceCityTile {
					if tile.OwnerID != nil && *tile.OwnerID == p.ID() {
						cityCount++
					}
				}
			}

			if req.Min != nil && cityCount < *req.Min {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "City requirement not met",
					RequiredValue: *req.Min,
					CurrentValue:  cityCount,
				})
			}
			if req.Max != nil && cityCount > *req.Max {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "City count exceeds maximum",
					RequiredValue: *req.Max,
					CurrentValue:  cityCount,
				})
			}

		case gamecards.RequirementGreeneries:
			// Count greeneries owned by the player on the board
			greeneryCount := 0
			for _, tile := range g.Board().Tiles() {
				if tile.OccupiedBy != nil && tile.OccupiedBy.Type == shared.ResourceGreeneryTile {
					if tile.OwnerID != nil && *tile.OwnerID == p.ID() {
						greeneryCount++
					}
				}
			}

			if req.Min != nil && greeneryCount < *req.Min {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Greenery requirement not met",
					RequiredValue: *req.Min,
					CurrentValue:  greeneryCount,
				})
			}
			if req.Max != nil && greeneryCount > *req.Max {
				result.AddError(playability.ValidationError{
					Type:          playability.ValidationErrorTypeRequirement,
					Message:       "Greenery count exceeds maximum",
					RequiredValue: *req.Max,
					CurrentValue:  greeneryCount,
				})
			}
		}
	}
}

// validateCardCost validates that the player can afford the card's base cost
func validateCardCost(card *gamecards.Card, p *player.Player, result *playability.PlayabilityResult) {
	resources := p.Resources().Get()
	cost := card.Cost

	// Basic check: does player have enough credits to cover full cost?
	// (We don't check steel/titanium discounts here - that's more complex payment validation)
	if resources.Credits < cost {
		// Check if steel/titanium could help (for building/space cards)
		canUseSteel := hasTag(card, shared.TagBuilding) && resources.Steel > 0
		canUseTitanium := hasTag(card, shared.TagSpace) && resources.Titanium > 0

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

// hasTag checks if a card has a specific tag
func hasTag(card *gamecards.Card, tag shared.CardTag) bool {
	for _, cardTag := range card.Tags {
		if cardTag == tag {
			return true
		}
	}
	return false
}
