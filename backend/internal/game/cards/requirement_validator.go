package cards

import (
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// RequirementValidationError represents a single card requirement validation failure
type RequirementValidationError struct {
	RequirementType string
	Message         string
	RequiredValue   interface{}
	CurrentValue    interface{}
}

// Error implements the error interface
func (e RequirementValidationError) Error() string {
	return e.Message
}

// ValidateCardRequirements validates all requirements for a card and returns all validation errors.
// This is the single source of truth for card requirement validation logic.
// Callers can use all errors (for UI display) or just the first error (for fail-fast validation).
func ValidateCardRequirements(
	card *Card,
	g *game.Game,
	p *player.Player,
	cardLookup CardLookup,
) []RequirementValidationError {
	if len(card.Requirements) == 0 {
		return nil
	}

	errors := []RequirementValidationError{}

	for _, req := range card.Requirements {
		switch req.Type {
		case RequirementTemperature:
			temp := g.GlobalParameters().Temperature()
			if req.Min != nil && temp < *req.Min {
				errors = append(errors, RequirementValidationError{
					RequirementType: "temperature",
					Message:         "Temperature requirement not met",
					RequiredValue:   *req.Min,
					CurrentValue:    temp,
				})
			}
			if req.Max != nil && temp > *req.Max {
				errors = append(errors, RequirementValidationError{
					RequirementType: "temperature",
					Message:         "Temperature exceeds maximum",
					RequiredValue:   *req.Max,
					CurrentValue:    temp,
				})
			}

		case RequirementOxygen:
			oxygen := g.GlobalParameters().Oxygen()
			if req.Min != nil && oxygen < *req.Min {
				errors = append(errors, RequirementValidationError{
					RequirementType: "oxygen",
					Message:         "Oxygen requirement not met",
					RequiredValue:   *req.Min,
					CurrentValue:    oxygen,
				})
			}
			if req.Max != nil && oxygen > *req.Max {
				errors = append(errors, RequirementValidationError{
					RequirementType: "oxygen",
					Message:         "Oxygen exceeds maximum",
					RequiredValue:   *req.Max,
					CurrentValue:    oxygen,
				})
			}

		case RequirementOceans:
			oceans := g.GlobalParameters().Oceans()
			if req.Min != nil && oceans < *req.Min {
				errors = append(errors, RequirementValidationError{
					RequirementType: "oceans",
					Message:         "Oceans requirement not met",
					RequiredValue:   *req.Min,
					CurrentValue:    oceans,
				})
			}
			if req.Max != nil && oceans > *req.Max {
				errors = append(errors, RequirementValidationError{
					RequirementType: "oceans",
					Message:         "Oceans exceeds maximum",
					RequiredValue:   *req.Max,
					CurrentValue:    oceans,
				})
			}

		case RequirementTR:
			tr := p.Resources().TerraformRating()
			if req.Min != nil && tr < *req.Min {
				errors = append(errors, RequirementValidationError{
					RequirementType: "terraform-rating",
					Message:         "Terraform rating requirement not met",
					RequiredValue:   *req.Min,
					CurrentValue:    tr,
				})
			}
			if req.Max != nil && tr > *req.Max {
				errors = append(errors, RequirementValidationError{
					RequirementType: "terraform-rating",
					Message:         "Terraform rating exceeds maximum",
					RequiredValue:   *req.Max,
					CurrentValue:    tr,
				})
			}

		case RequirementTags:
			if req.Tag == nil {
				continue
			}

			// Count tags across all played cards (including corporation)
			tagCount := 0
			for _, playedCardID := range p.PlayedCards().Cards() {
				playedCard, err := cardLookup.GetByID(playedCardID)
				if err != nil {
					continue
				}
				if hasTag(playedCard, *req.Tag) {
					tagCount++
				}
			}

			// Also count corporation tags
			if corpID := p.CorporationID(); corpID != "" {
				corpCard, err := cardLookup.GetByID(corpID)
				if err == nil && hasTag(corpCard, *req.Tag) {
					tagCount++
				}
			}

			if req.Min != nil && tagCount < *req.Min {
				errors = append(errors, RequirementValidationError{
					RequirementType: "tags",
					Message:         "Tag requirement not met",
					RequiredValue:   *req.Min,
					CurrentValue:    tagCount,
				})
			}
			if req.Max != nil && tagCount > *req.Max {
				errors = append(errors, RequirementValidationError{
					RequirementType: "tags",
					Message:         "Tag count exceeds maximum",
					RequiredValue:   *req.Max,
					CurrentValue:    tagCount,
				})
			}

		case RequirementProduction:
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
				errors = append(errors, RequirementValidationError{
					RequirementType: "production",
					Message:         "Production requirement not met",
					RequiredValue:   *req.Min,
					CurrentValue:    currentProduction,
				})
			}
			if req.Max != nil && currentProduction > *req.Max {
				errors = append(errors, RequirementValidationError{
					RequirementType: "production",
					Message:         "Production exceeds maximum",
					RequiredValue:   *req.Max,
					CurrentValue:    currentProduction,
				})
			}

		case RequirementResource:
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
				errors = append(errors, RequirementValidationError{
					RequirementType: "resource",
					Message:         "Resource requirement not met",
					RequiredValue:   *req.Min,
					CurrentValue:    currentAmount,
				})
			}
			if req.Max != nil && currentAmount > *req.Max {
				errors = append(errors, RequirementValidationError{
					RequirementType: "resource",
					Message:         "Resource exceeds maximum",
					RequiredValue:   *req.Max,
					CurrentValue:    currentAmount,
				})
			}

		case RequirementCities:
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
				errors = append(errors, RequirementValidationError{
					RequirementType: "cities",
					Message:         "City requirement not met",
					RequiredValue:   *req.Min,
					CurrentValue:    cityCount,
				})
			}
			if req.Max != nil && cityCount > *req.Max {
				errors = append(errors, RequirementValidationError{
					RequirementType: "cities",
					Message:         "City count exceeds maximum",
					RequiredValue:   *req.Max,
					CurrentValue:    cityCount,
				})
			}

		case RequirementGreeneries:
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
				errors = append(errors, RequirementValidationError{
					RequirementType: "greeneries",
					Message:         "Greenery requirement not met",
					RequiredValue:   *req.Min,
					CurrentValue:    greeneryCount,
				})
			}
			if req.Max != nil && greeneryCount > *req.Max {
				errors = append(errors, RequirementValidationError{
					RequirementType: "greeneries",
					Message:         "Greenery count exceeds maximum",
					RequiredValue:   *req.Max,
					CurrentValue:    greeneryCount,
				})
			}
		}
	}

	return errors
}

// hasTag checks if a card has a specific tag
func hasTag(card *Card, tag shared.CardTag) bool {
	for _, cardTag := range card.Tags {
		if cardTag == tag {
			return true
		}
	}
	return false
}
