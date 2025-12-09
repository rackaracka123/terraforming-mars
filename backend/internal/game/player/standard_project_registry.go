package player

import (
	"terraforming-mars-backend/internal/events"
)

// GlobalParametersProvider defines what we need from a game to provide global parameters
type GlobalParametersProvider interface {
	Temperature() int
	Oxygen() int
	Oceans() int
}

// RegisterStandardProjects registers all 6 standard projects for a player
// Each project subscribes to relevant events to track its own availability
// The globalParams should be game.GlobalParameters()
func RegisterStandardProjects(globalParams GlobalParametersProvider, player *Player, eventBus *events.EventBusImpl) {
	// Create all 6 standard projects
	projects := []*StandardProject{
		NewSellPatentsProject(globalParams, player),
		NewPowerPlantProject(globalParams, player),
		NewAsteroidProject(globalParams, player),
		NewAquiferProject(globalParams, player),
		NewGreeneryProject(globalParams, player),
		NewCityProject(globalParams, player),
	}

	// Register each project and subscribe to events
	for _, project := range projects {
		player.StandardProjects().Register(project)
		project.Subscribe(eventBus, player.ID())
	}
}
