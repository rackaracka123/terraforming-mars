package model

// Corporation represents a corporation card with special abilities
type Corporation struct {
	ID                 string      `json:"id" ts:"string"`
	Name               string      `json:"name" ts:"string"`
	Description        string      `json:"description" ts:"string"`
	StartingCredits    int         `json:"startingCredits" ts:"number"`
	StartingResources  ResourceSet `json:"startingResources" ts:"ResourceSet"`
	StartingProduction ResourceSet `json:"startingProduction" ts:"ResourceSet"`
	Tags               []CardTag   `json:"tags" ts:"CardTag[]"`
	SpecialEffects     []string    `json:"specialEffects" ts:"string[]"` // Descriptions of special abilities
}

// ConvertCardToCorporation converts a corporation Card to a Corporation struct
func ConvertCardToCorporation(card Card) Corporation {
	corp := Corporation{
		ID:                 card.ID,
		Name:               card.Name,
		Description:        card.Description,
		StartingCredits:    0,
		StartingResources:  ResourceSet{},
		StartingProduction: ResourceSet{},
		Tags:               card.Tags,
		SpecialEffects:     []string{card.Description},
	}

	// Parse starting bonuses from the first auto-trigger behavior
	// Corporation cards have their starting bonuses in the outputs of auto behaviors
	for _, behavior := range card.Behaviors {
		// Look for auto-trigger behaviors
		hasAutoTrigger := false
		for _, trigger := range behavior.Triggers {
			if trigger.Type == "auto" {
				hasAutoTrigger = true
				break
			}
		}

		if !hasAutoTrigger {
			continue
		}

		// Parse outputs to extract starting resources and production
		for _, output := range behavior.Outputs {
			switch output.Type {
			// Starting resources
			case ResourceCredits:
				corp.StartingCredits = output.Amount
				corp.StartingResources.Credits = output.Amount
			case ResourceSteel:
				corp.StartingResources.Steel = output.Amount
			case ResourceTitanium:
				corp.StartingResources.Titanium = output.Amount
			case ResourcePlants:
				corp.StartingResources.Plants = output.Amount
			case ResourceEnergy:
				corp.StartingResources.Energy = output.Amount
			case ResourceHeat:
				corp.StartingResources.Heat = output.Amount

			// Starting production
			case ResourceCreditsProduction:
				corp.StartingProduction.Credits = output.Amount
			case ResourceSteelProduction:
				corp.StartingProduction.Steel = output.Amount
			case ResourceTitaniumProduction:
				corp.StartingProduction.Titanium = output.Amount
			case ResourcePlantsProduction:
				corp.StartingProduction.Plants = output.Amount
			case ResourceEnergyProduction:
				corp.StartingProduction.Energy = output.Amount
			case ResourceHeatProduction:
				corp.StartingProduction.Heat = output.Amount
			}
		}

		// Only process the first auto behavior (starting bonuses)
		break
	}

	return corp
}
