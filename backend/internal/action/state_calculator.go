package action

import (
	"fmt"
	"time"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// ========================================
// PlayerCard State Calculation
// ========================================

// CalculatePlayerCardState computes playability state for a card.
// This function can access both Game and Player without circular dependencies.
// card parameter must be *gamecards.Card
func CalculatePlayerCardState(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) player.EntityState {
	errors := []player.StateError{}
	metadata := make(map[string]interface{})

	// 1. Phase check
	errors = append(errors, validatePhase(g)...)

	// 2. Requirements check (extracted from PlayCardAction lines 209-321)
	errors = append(errors, validateRequirements(card, p, g, cardRegistry)...)

	// 3. Cost calculation with discounts (uses existing RequirementModifierCalculator)
	effectiveCost, discounts := calculateEffectiveCost(card, p)
	if len(discounts) > 0 {
		metadata["discounts"] = discounts
	}

	// 4. Affordability check
	errors = append(errors, validateAffordability(p, effectiveCost)...)

	return player.EntityState{
		Errors:         errors,
		Cost:           &effectiveCost,
		Metadata:       metadata,
		LastCalculated: time.Now(),
	}
}

// ========================================
// PlayerCardAction State Calculation
// ========================================

// CalculatePlayerCardActionState computes usability state for a card action.
func CalculatePlayerCardActionState(
	cardID string,
	behavior shared.CardBehavior,
	pca *player.PlayerCardAction,
	p *player.Player,
	g *game.Game,
) player.EntityState {
	errors := []player.StateError{}

	// 1. Check if it's the player's turn
	currentTurn := g.CurrentTurn()
	if currentTurn != nil && currentTurn.PlayerID() != p.ID() {
		errors = append(errors, player.StateError{
			Code:     "NOT_YOUR_TURN",
			Category: "turn",
			Message:  "Not your turn",
		})
	}

	// 2. Check input resource availability
	resources := p.Resources().Get()
	for _, input := range behavior.Inputs {
		available := getResourceAmount(resources, input.ResourceType)
		if available < input.Amount {
			errors = append(errors, player.StateError{
				Code:     "INSUFFICIENT_RESOURCES",
				Category: "input",
				Message:  fmt.Sprintf("Need %d %s, have %d", input.Amount, input.ResourceType, available),
			})
		}
	}

	// 3. TODO: Check max usage limits when CardBehavior supports MaxUsesPerTurn/MaxUsesPerGeneration fields
	// For now, usage limits are not enforced in state calculation
	_ = pca // Silence unused warning until usage limit fields are added

	return player.EntityState{
		Errors:         errors,
		Cost:           nil, // Actions typically don't have credit costs
		Metadata:       make(map[string]interface{}),
		LastCalculated: time.Now(),
	}
}

// ========================================
// PlayerStandardProject State Calculation
// ========================================

// CalculatePlayerStandardProjectState computes availability state for a standard project.
func CalculatePlayerStandardProjectState(
	projectType shared.StandardProject,
	p *player.Player,
	g *game.Game,
) player.EntityState {
	errors := []player.StateError{}
	metadata := make(map[string]interface{})

	// Get base cost
	cost, exists := shared.StandardProjectCost[projectType]
	if !exists {
		errors = append(errors, player.StateError{
			Code:     "INVALID_PROJECT_TYPE",
			Category: "configuration",
			Message:  fmt.Sprintf("Unknown standard project type: %s", projectType),
		})
		return player.EntityState{
			Errors:         errors,
			Cost:           &cost,
			Metadata:       metadata,
			LastCalculated: time.Now(),
		}
	}

	// 1. Check affordability (reuse shared helper)
	errors = append(errors, validateAffordability(p, cost)...)

	// 2. Check project-specific availability
	switch projectType {
	case shared.StandardProjectAquifer:
		// Check if max oceans reached (max is 9 in base game)
		currentOceans := g.GlobalParameters().Oceans()
		oceansRemaining := 9 - currentOceans
		metadata["oceansRemaining"] = oceansRemaining
		if oceansRemaining <= 0 {
			errors = append(errors, player.StateError{
				Code:     "NO_OCEAN_TILES",
				Category: "availability",
				Message:  "No ocean tiles remaining",
			})
		}

	case shared.StandardProjectAsteroid:
		// No special prerequisites

	case shared.StandardProjectCity:
		// TODO: Check if board has available city placement locations
		// For now, assume always available if affordable

	case shared.StandardProjectGreenery:
		// TODO: Check if board has available greenery placement locations
		// For now, assume always available if affordable

	case shared.StandardProjectPowerPlant:
		// No special prerequisites

	default:
		// Other standard projects have no special prerequisites
	}

	return player.EntityState{
		Errors:         errors,
		Cost:           &cost,
		Metadata:       metadata,
		LastCalculated: time.Now(),
	}
}

// ========================================
// Shared Validation Helpers
// ========================================

// validatePhase checks if action is allowed in current phase.
func validatePhase(g *game.Game) []player.StateError {
	if g.CurrentPhase() != game.GamePhaseAction {
		return []player.StateError{{
			Code:     "WRONG_PHASE",
			Category: "phase",
			Message:  fmt.Sprintf("Can only play cards during action phase, current phase: %s", g.CurrentPhase()),
		}}
	}
	return nil
}

// validateRequirements checks all card requirements.
// Extracted from PlayCardAction.validateCardRequirements() lines 209-321.
func validateRequirements(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) []player.StateError {
	if len(card.Requirements) == 0 {
		return nil
	}

	errors := []player.StateError{}

	for _, req := range card.Requirements {
		err := checkRequirement(req, p, g, cardRegistry)
		if err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// checkRequirement validates a single requirement.
// Extracted from PlayCardAction - contains the switch statement for all requirement types.
func checkRequirement(
	req gamecards.Requirement,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) *player.StateError {
	switch req.Type {
	case gamecards.RequirementTemperature:
		temp := g.GlobalParameters().Temperature()
		if req.Min != nil && temp < *req.Min {
			return &player.StateError{
				Code:     "TEMPERATURE_TOO_LOW",
				Category: "requirement",
				Message:  fmt.Sprintf("Temperature requirement not met: need %d°C, current %d°C", *req.Min, temp),
			}
		}
		if req.Max != nil && temp > *req.Max {
			return &player.StateError{
				Code:     "TEMPERATURE_TOO_HIGH",
				Category: "requirement",
				Message:  fmt.Sprintf("Temperature requirement not met: max %d°C, current %d°C", *req.Max, temp),
			}
		}

	case gamecards.RequirementOxygen:
		oxygen := g.GlobalParameters().Oxygen()
		if req.Min != nil && oxygen < *req.Min {
			return &player.StateError{
				Code:     "OXYGEN_TOO_LOW",
				Category: "requirement",
				Message:  fmt.Sprintf("Oxygen requirement not met: need %d%%, current %d%%", *req.Min, oxygen),
			}
		}
		if req.Max != nil && oxygen > *req.Max {
			return &player.StateError{
				Code:     "OXYGEN_TOO_HIGH",
				Category: "requirement",
				Message:  fmt.Sprintf("Oxygen requirement not met: max %d%%, current %d%%", *req.Max, oxygen),
			}
		}

	case gamecards.RequirementOceans:
		oceans := g.GlobalParameters().Oceans()
		if req.Min != nil && oceans < *req.Min {
			return &player.StateError{
				Code:     "OCEANS_TOO_LOW",
				Category: "requirement",
				Message:  fmt.Sprintf("Ocean requirement not met: need %d, current %d", *req.Min, oceans),
			}
		}
		if req.Max != nil && oceans > *req.Max {
			return &player.StateError{
				Code:     "OCEANS_TOO_HIGH",
				Category: "requirement",
				Message:  fmt.Sprintf("Ocean requirement not met: max %d, current %d", *req.Max, oceans),
			}
		}

	case gamecards.RequirementTR:
		tr := p.Resources().TerraformRating()
		if req.Min != nil && tr < *req.Min {
			return &player.StateError{
				Code:     "TR_TOO_LOW",
				Category: "requirement",
				Message:  fmt.Sprintf("Terraform rating requirement not met: need %d, current %d", *req.Min, tr),
			}
		}
		if req.Max != nil && tr > *req.Max {
			return &player.StateError{
				Code:     "TR_TOO_HIGH",
				Category: "requirement",
				Message:  fmt.Sprintf("Terraform rating requirement not met: max %d, current %d", *req.Max, tr),
			}
		}

	case gamecards.RequirementTags:
		if req.Tag == nil {
			return &player.StateError{
				Code:     "INVALID_REQUIREMENT",
				Category: "requirement",
				Message:  "Tag requirement missing tag specification",
			}
		}

		// Count tags across all played cards (including corporation)
		tagCount := 0
		for _, playedCardID := range p.PlayedCards().Cards() {
			// TODO: Get card from registry and check if it has the tag
			// This requires fully integrating with CardRegistry
			// For now, skip tag validation (same as PlayCardAction line 260)
			_ = playedCardID
		}

		if req.Min != nil && tagCount < *req.Min {
			return &player.StateError{
				Code:     "INSUFFICIENT_TAGS",
				Category: "requirement",
				Message:  fmt.Sprintf("Tag requirement not met: need %d %s tags, have %d", *req.Min, *req.Tag, tagCount),
			}
		}
		if req.Max != nil && tagCount > *req.Max {
			return &player.StateError{
				Code:     "TOO_MANY_TAGS",
				Category: "requirement",
				Message:  fmt.Sprintf("Tag requirement not met: max %d %s tags, have %d", *req.Max, *req.Tag, tagCount),
			}
		}

	case gamecards.RequirementProduction:
		if req.Resource == nil {
			return &player.StateError{
				Code:     "INVALID_REQUIREMENT",
				Category: "requirement",
				Message:  "Production requirement missing resource specification",
			}
		}
		// TODO: Implement production requirement validation
		// This requires checking player's production values
		// For now, skip production validation (same as PlayCardAction line 277-279)

	case gamecards.RequirementResource:
		if req.Resource == nil {
			return &player.StateError{
				Code:     "INVALID_REQUIREMENT",
				Category: "requirement",
				Message:  "Resource requirement missing resource specification",
			}
		}
		resources := p.Resources().Get()
		currentAmount := getResourceAmount(resources, *req.Resource)

		if req.Min != nil && currentAmount < *req.Min {
			return &player.StateError{
				Code:     "INSUFFICIENT_RESOURCES",
				Category: "requirement",
				Message:  fmt.Sprintf("Resource requirement not met: need %d %s, have %d", *req.Min, *req.Resource, currentAmount),
			}
		}
		if req.Max != nil && currentAmount > *req.Max {
			return &player.StateError{
				Code:     "TOO_MANY_RESOURCES",
				Category: "requirement",
				Message:  fmt.Sprintf("Resource requirement not met: max %d %s, have %d", *req.Max, *req.Resource, currentAmount),
			}
		}

	case gamecards.RequirementCities, gamecards.RequirementGreeneries:
		// TODO: Implement tile-based requirements when Board tile counting is ready
		// For now, skip these validations (same as PlayCardAction line 310-312)

	case gamecards.RequirementVenus:
		// TODO: Implement Venus track when expansion is supported
		// For now, skip Venus validation (same as PlayCardAction line 314-316)
	}

	return nil
}

// calculateEffectiveCost computes cost with discounts.
// TODO: Integrate with actual RequirementModifier system for card cost discounts
// For now, returns base cost without discounts since RequirementModifier structure
// differs from initial assumptions (has Amount, AffectedResources, CardTarget, StandardProjectTarget)
func calculateEffectiveCost(card *gamecards.Card, p *player.Player) (int, map[shared.CardTag]int) {
	baseCost := card.Cost
	discounts := make(map[shared.CardTag]int)

	// TODO: Implement discount calculation when RequirementModifier integration is clarified
	// Current RequirementModifier doesn't have Discounts or CardTypes fields
	// May need to check if AffectedResources contains credits and apply Amount as discount
	_ = p // Silence unused warning until discount logic is implemented

	return baseCost, discounts
}

// validateAffordability checks if player can afford the cost.
func validateAffordability(p *player.Player, cost int) []player.StateError {
	credits := p.Resources().Get().Credits
	if credits < cost {
		return []player.StateError{{
			Code:     "INSUFFICIENT_CREDITS",
			Category: "cost",
			Message:  fmt.Sprintf("Need %d credits, have %d", cost, credits),
		}}
	}
	return nil
}

// getResourceAmount extracts the amount of a specific resource from Resources.
func getResourceAmount(resources shared.Resources, resourceType shared.ResourceType) int {
	switch resourceType {
	case shared.ResourceCredits:
		return resources.Credits
	case shared.ResourceSteel:
		return resources.Steel
	case shared.ResourceTitanium:
		return resources.Titanium
	case shared.ResourcePlants:
		return resources.Plants
	case shared.ResourceEnergy:
		return resources.Energy
	case shared.ResourceHeat:
		return resources.Heat
	default:
		return 0
	}
}
