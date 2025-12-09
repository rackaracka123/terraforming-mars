package playability

import (
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
)

// Standard project costs
const (
	CostSellPatents = 0 // No upfront cost, sells cards
	CostPowerPlant  = 11
	CostAsteroid    = 14
	CostAquifer     = 18
	CostGreenery    = 23
	CostCity        = 25
)

// GetAllStandardProjects returns all standard projects with their availability for a player
func GetAllStandardProjects(g *game.Game, p *player.Player) []StandardProject {
	return []StandardProject{
		CheckSellPatents(g, p),
		CheckPowerPlant(g, p),
		CheckAsteroid(g, p),
		CheckAquifer(g, p),
		CheckGreenery(g, p),
		CheckCity(g, p),
	}
}

// CheckSellPatents checks if the Sell Patents standard project is available
func CheckSellPatents(g *game.Game, p *player.Player) StandardProject {
	project := StandardProject{
		ID:          "sell-patents",
		Name:        "Sell Patents",
		Type:        StandardProjectSellPatents,
		Cost:        CostSellPatents,
		Description: "Sell any number of cards from hand for 1 M€ each",
		IsAvailable: true,
		Errors:      []ValidationError{},
	}

	// Check if player has cards in hand to sell
	if p.Hand().CardCount() == 0 {
		project.IsAvailable = false
		project.Errors = append(project.Errors, ValidationError{
			Type:          ValidationErrorTypeResource,
			Message:       "No cards in hand to sell",
			RequiredValue: 1,
			CurrentValue:  0,
		})
	}

	return project
}

// CheckPowerPlant checks if the Power Plant standard project is available
func CheckPowerPlant(g *game.Game, p *player.Player) StandardProject {
	project := StandardProject{
		ID:          "power-plant",
		Name:        "Power Plant",
		Type:        StandardProjectPowerPlant,
		Cost:        CostPowerPlant,
		Description: "Spend 11 M€ to increase your energy production 1 step",
		IsAvailable: true,
		Errors:      []ValidationError{},
	}

	resources := p.Resources().Get()
	if resources.Credits < CostPowerPlant {
		project.IsAvailable = false
		project.Errors = append(project.Errors, ValidationError{
			Type:          ValidationErrorTypeCost,
			Message:       "Insufficient credits",
			RequiredValue: CostPowerPlant,
			CurrentValue:  resources.Credits,
		})
	}

	return project
}

// CheckAsteroid checks if the Asteroid standard project is available
func CheckAsteroid(g *game.Game, p *player.Player) StandardProject {
	project := StandardProject{
		ID:          "asteroid",
		Name:        "Asteroid",
		Type:        StandardProjectAsteroid,
		Cost:        CostAsteroid,
		Description: "Spend 14 M€ to raise temperature 1 step",
		IsAvailable: true,
		Errors:      []ValidationError{},
	}

	// Check if temperature is already at maximum
	currentTemp := g.GlobalParameters().Temperature()
	if currentTemp >= 8 { // MaxTemperature
		project.IsAvailable = false
		project.Errors = append(project.Errors, ValidationError{
			Type:          ValidationErrorTypeGlobalParam,
			Message:       "Temperature already at maximum",
			RequiredValue: 8,
			CurrentValue:  currentTemp,
		})
	}

	// Check affordability
	resources := p.Resources().Get()
	if resources.Credits < CostAsteroid {
		project.IsAvailable = false
		project.Errors = append(project.Errors, ValidationError{
			Type:          ValidationErrorTypeCost,
			Message:       "Insufficient credits",
			RequiredValue: CostAsteroid,
			CurrentValue:  resources.Credits,
		})
	}

	return project
}

// CheckAquifer checks if the Aquifer standard project is available
func CheckAquifer(g *game.Game, p *player.Player) StandardProject {
	project := StandardProject{
		ID:          "aquifer",
		Name:        "Aquifer",
		Type:        StandardProjectAquifer,
		Cost:        CostAquifer,
		Description: "Spend 18 M€ to place an ocean tile",
		IsAvailable: true,
		Errors:      []ValidationError{},
	}

	// Check if oceans are already at maximum
	currentOceans := g.GlobalParameters().Oceans()
	if currentOceans >= 9 { // MaxOceans
		project.IsAvailable = false
		project.Errors = append(project.Errors, ValidationError{
			Type:          ValidationErrorTypeGlobalParam,
			Message:       "Oceans already at maximum",
			RequiredValue: 9,
			CurrentValue:  currentOceans,
		})
	}

	// Check affordability
	resources := p.Resources().Get()
	if resources.Credits < CostAquifer {
		project.IsAvailable = false
		project.Errors = append(project.Errors, ValidationError{
			Type:          ValidationErrorTypeCost,
			Message:       "Insufficient credits",
			RequiredValue: CostAquifer,
			CurrentValue:  resources.Credits,
		})
	}

	return project
}

// CheckGreenery checks if the Greenery standard project is available
func CheckGreenery(g *game.Game, p *player.Player) StandardProject {
	project := StandardProject{
		ID:          "greenery",
		Name:        "Greenery",
		Type:        StandardProjectGreenery,
		Cost:        CostGreenery,
		Description: "Spend 23 M€ to place a greenery tile and raise oxygen 1 step",
		IsAvailable: true,
		Errors:      []ValidationError{},
	}

	// Check if oxygen is already at maximum
	currentOxygen := g.GlobalParameters().Oxygen()
	if currentOxygen >= 14 { // MaxOxygen
		project.IsAvailable = false
		project.Errors = append(project.Errors, ValidationError{
			Type:          ValidationErrorTypeGlobalParam,
			Message:       "Oxygen already at maximum",
			RequiredValue: 14,
			CurrentValue:  currentOxygen,
		})
	}

	// Check affordability
	resources := p.Resources().Get()
	if resources.Credits < CostGreenery {
		project.IsAvailable = false
		project.Errors = append(project.Errors, ValidationError{
			Type:          ValidationErrorTypeCost,
			Message:       "Insufficient credits",
			RequiredValue: CostGreenery,
			CurrentValue:  resources.Credits,
		})
	}

	return project
}

// CheckCity checks if the City standard project is available
func CheckCity(g *game.Game, p *player.Player) StandardProject {
	project := StandardProject{
		ID:          "city",
		Name:        "City",
		Type:        StandardProjectCity,
		Cost:        CostCity,
		Description: "Spend 25 M€ to place a city tile and increase your M€ production 1 step",
		IsAvailable: true,
		Errors:      []ValidationError{},
	}

	// Check affordability
	resources := p.Resources().Get()
	if resources.Credits < CostCity {
		project.IsAvailable = false
		project.Errors = append(project.Errors, ValidationError{
			Type:          ValidationErrorTypeCost,
			Message:       "Insufficient credits",
			RequiredValue: CostCity,
			CurrentValue:  resources.Credits,
		})
	}

	return project
}
