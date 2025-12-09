package player

import (
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/playability"
)

// StandardProject represents a single standard project that can be executed by a player
// Each standard project subscribes to relevant events and maintains its own availability state
type StandardProject struct {
	id            string
	projectType   playability.StandardProjectType
	name          string
	description   string
	cost          int
	availability  playability.StandardProject
	availabilityMu sync.RWMutex
	subscriptionIDs []events.SubscriptionID

	// Validation function to check availability
	checkAvailability func() playability.StandardProject
}

// NewStandardProject creates a new standard project
func NewStandardProject(
	id string,
	projectType playability.StandardProjectType,
	name string,
	description string,
	cost int,
	checkFunc func() playability.StandardProject,
) *StandardProject {
	return &StandardProject{
		id:            id,
		projectType:   projectType,
		name:          name,
		description:   description,
		cost:          cost,
		availability:  playability.StandardProject{},
		subscriptionIDs: []events.SubscriptionID{},
		checkAvailability: checkFunc,
	}
}

// Subscribe registers event handlers for this standard project
func (sp *StandardProject) Subscribe(eventBus *events.EventBusImpl, playerID string) {
	// Each project type subscribes to different events based on what affects its availability
	switch sp.projectType {
	case playability.StandardProjectSellPatents:
		sp.subscribeToHandChanges(eventBus, playerID)

	case playability.StandardProjectPowerPlant:
		sp.subscribeToResourceChanges(eventBus, playerID)

	case playability.StandardProjectAsteroid:
		sp.subscribeToResourceChanges(eventBus, playerID)
		sp.subscribeToTemperatureChanges(eventBus)

	case playability.StandardProjectAquifer:
		sp.subscribeToResourceChanges(eventBus, playerID)
		sp.subscribeToOceanChanges(eventBus)

	case playability.StandardProjectGreenery:
		sp.subscribeToResourceChanges(eventBus, playerID)
		sp.subscribeToOxygenChanges(eventBus)

	case playability.StandardProjectCity:
		sp.subscribeToResourceChanges(eventBus, playerID)
	}

	// Initial calculation
	sp.recalculate()
}

// GetAvailability returns the current availability state
func (sp *StandardProject) GetAvailability() playability.StandardProject {
	sp.availabilityMu.RLock()
	defer sp.availabilityMu.RUnlock()
	return sp.availability
}

// recalculate updates the availability state by calling the check function
func (sp *StandardProject) recalculate() {
	if sp.checkAvailability == nil {
		return
	}

	newAvailability := sp.checkAvailability()

	sp.availabilityMu.Lock()
	sp.availability = newAvailability
	sp.availabilityMu.Unlock()
}

// Event subscription helpers

func (sp *StandardProject) subscribeToHandChanges(eventBus *events.EventBusImpl, playerID string) {
	subID := events.Subscribe(eventBus, func(event events.CardHandUpdatedEvent) {
		if event.PlayerID == playerID {
			sp.recalculate()
		}
	})
	sp.subscriptionIDs = append(sp.subscriptionIDs, subID)
}

func (sp *StandardProject) subscribeToResourceChanges(eventBus *events.EventBusImpl, playerID string) {
	subID := events.Subscribe(eventBus, func(event events.ResourcesChangedEvent) {
		if event.PlayerID == playerID {
			sp.recalculate()
		}
	})
	sp.subscriptionIDs = append(sp.subscriptionIDs, subID)
}

func (sp *StandardProject) subscribeToTemperatureChanges(eventBus *events.EventBusImpl) {
	subID := events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		sp.recalculate()
	})
	sp.subscriptionIDs = append(sp.subscriptionIDs, subID)
}

func (sp *StandardProject) subscribeToOxygenChanges(eventBus *events.EventBusImpl) {
	subID := events.Subscribe(eventBus, func(event events.OxygenChangedEvent) {
		sp.recalculate()
	})
	sp.subscriptionIDs = append(sp.subscriptionIDs, subID)
}

func (sp *StandardProject) subscribeToOceanChanges(eventBus *events.EventBusImpl) {
	subID := events.Subscribe(eventBus, func(event events.OceansChangedEvent) {
		sp.recalculate()
	})
	sp.subscriptionIDs = append(sp.subscriptionIDs, subID)
}

// Factory functions for creating specific standard projects

// NewSellPatentsProject creates the Sell Patents standard project
func NewSellPatentsProject(game GameInterface, player *Player) *StandardProject {
	return NewStandardProject(
		"sell-patents",
		playability.StandardProjectSellPatents,
		"Sell Patents",
		"Sell any number of cards from hand for 1 M€ each",
		0, // Free action
		func() playability.StandardProject {
			result := playability.StandardProject{
				ID:          "sell-patents",
				Name:        "Sell Patents",
				Type:        playability.StandardProjectSellPatents,
				Cost:        0,
				Description: "Sell any number of cards from hand for 1 M€ each",
				IsAvailable: true,
				Errors:      []playability.ValidationError{},
			}

			if player.Hand().CardCount() == 0 {
				result.IsAvailable = false
				result.Errors = append(result.Errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeResource,
					Message:       "No cards in hand to sell",
					RequiredValue: 1,
					CurrentValue:  0,
				})
			}

			return result
		},
	)
}

// NewPowerPlantProject creates the Power Plant standard project
func NewPowerPlantProject(game GameInterface, player *Player) *StandardProject {
	return NewStandardProject(
		"power-plant",
		playability.StandardProjectPowerPlant,
		"Power Plant",
		"Spend 11 M€ to increase your energy production 1 step",
		11,
		func() playability.StandardProject {
			result := playability.StandardProject{
				ID:          "power-plant",
				Name:        "Power Plant",
				Type:        playability.StandardProjectPowerPlant,
				Cost:        11,
				Description: "Spend 11 M€ to increase your energy production 1 step",
				IsAvailable: true,
				Errors:      []playability.ValidationError{},
			}

			resources := player.Resources().Get()
			if resources.Credits < 11 {
				result.IsAvailable = false
				result.Errors = append(result.Errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeCost,
					Message:       "Insufficient credits",
					RequiredValue: 11,
					CurrentValue:  resources.Credits,
				})
			}

			return result
		},
	)
}

// NewAsteroidProject creates the Asteroid standard project
func NewAsteroidProject(game GameInterface, player *Player) *StandardProject {
	return NewStandardProject(
		"asteroid",
		playability.StandardProjectAsteroid,
		"Asteroid",
		"Spend 14 M€ to raise temperature 1 step",
		14,
		func() playability.StandardProject {
			result := playability.StandardProject{
				ID:          "asteroid",
				Name:        "Asteroid",
				Type:        playability.StandardProjectAsteroid,
				Cost:        14,
				Description: "Spend 14 M€ to raise temperature 1 step",
				IsAvailable: true,
				Errors:      []playability.ValidationError{},
			}

			// Check temperature limit
			currentTemp := game.Temperature()
			if currentTemp >= 8 {
				result.IsAvailable = false
				result.Errors = append(result.Errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeGlobalParam,
					Message:       "Temperature already at maximum",
					RequiredValue: 8,
					CurrentValue:  currentTemp,
				})
			}

			// Check affordability
			resources := player.Resources().Get()
			if resources.Credits < 14 {
				result.IsAvailable = false
				result.Errors = append(result.Errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeCost,
					Message:       "Insufficient credits",
					RequiredValue: 14,
					CurrentValue:  resources.Credits,
				})
			}

			return result
		},
	)
}

// NewAquiferProject creates the Aquifer standard project
func NewAquiferProject(game GameInterface, player *Player) *StandardProject {
	return NewStandardProject(
		"aquifer",
		playability.StandardProjectAquifer,
		"Aquifer",
		"Spend 18 M€ to place an ocean tile",
		18,
		func() playability.StandardProject {
			result := playability.StandardProject{
				ID:          "aquifer",
				Name:        "Aquifer",
				Type:        playability.StandardProjectAquifer,
				Cost:        18,
				Description: "Spend 18 M€ to place an ocean tile",
				IsAvailable: true,
				Errors:      []playability.ValidationError{},
			}

			// Check oceans limit
			currentOceans := game.Oceans()
			if currentOceans >= 9 {
				result.IsAvailable = false
				result.Errors = append(result.Errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeGlobalParam,
					Message:       "Oceans already at maximum",
					RequiredValue: 9,
					CurrentValue:  currentOceans,
				})
			}

			// Check affordability
			resources := player.Resources().Get()
			if resources.Credits < 18 {
				result.IsAvailable = false
				result.Errors = append(result.Errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeCost,
					Message:       "Insufficient credits",
					RequiredValue: 18,
					CurrentValue:  resources.Credits,
				})
			}

			return result
		},
	)
}

// NewGreeneryProject creates the Greenery standard project
func NewGreeneryProject(game GameInterface, player *Player) *StandardProject {
	return NewStandardProject(
		"greenery",
		playability.StandardProjectGreenery,
		"Greenery",
		"Spend 23 M€ to place a greenery tile and raise oxygen 1 step",
		23,
		func() playability.StandardProject {
			result := playability.StandardProject{
				ID:          "greenery",
				Name:        "Greenery",
				Type:        playability.StandardProjectGreenery,
				Cost:        23,
				Description: "Spend 23 M€ to place a greenery tile and raise oxygen 1 step",
				IsAvailable: true,
				Errors:      []playability.ValidationError{},
			}

			// Check oxygen limit
			currentOxygen := game.Oxygen()
			if currentOxygen >= 14 {
				result.IsAvailable = false
				result.Errors = append(result.Errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeGlobalParam,
					Message:       "Oxygen already at maximum",
					RequiredValue: 14,
					CurrentValue:  currentOxygen,
				})
			}

			// Check affordability
			resources := player.Resources().Get()
			if resources.Credits < 23 {
				result.IsAvailable = false
				result.Errors = append(result.Errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeCost,
					Message:       "Insufficient credits",
					RequiredValue: 23,
					CurrentValue:  resources.Credits,
				})
			}

			return result
		},
	)
}

// NewCityProject creates the City standard project
func NewCityProject(game GameInterface, player *Player) *StandardProject {
	return NewStandardProject(
		"city",
		playability.StandardProjectCity,
		"City",
		"Spend 25 M€ to place a city tile and increase your M€ production 1 step",
		25,
		func() playability.StandardProject {
			result := playability.StandardProject{
				ID:          "city",
				Name:        "City",
				Type:        playability.StandardProjectCity,
				Cost:        25,
				Description: "Spend 25 M€ to place a city tile and increase your M€ production 1 step",
				IsAvailable: true,
				Errors:      []playability.ValidationError{},
			}

			// Check affordability
			resources := player.Resources().Get()
			if resources.Credits < 25 {
				result.IsAvailable = false
				result.Errors = append(result.Errors, playability.ValidationError{
					Type:          playability.ValidationErrorTypeCost,
					Message:       "Insufficient credits",
					RequiredValue: 25,
					CurrentValue:  resources.Credits,
				})
			}

			return result
		},
	)
}

// GameInterface minimal interface needed by standard projects
// This avoids circular dependencies while keeping standard projects flexible
type GameInterface interface {
	Temperature() int
	Oxygen() int
	Oceans() int
}
