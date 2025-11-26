package player

import (
	"github.com/google/uuid"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/types"
)

// Factory creates new players with wired repositories
type Factory struct {
	eventBus *events.EventBusImpl
}

// NewFactory creates a new player factory
func NewFactory(eventBus *events.EventBusImpl) *Factory {
	return &Factory{
		eventBus: eventBus,
	}
}

// CreatePlayer creates a new player with initialized state and injected eventBus
func (f *Factory) CreatePlayer(gameID, playerID, name string) *Player {
	// Generate ID if not provided
	if playerID == "" {
		playerID = uuid.New().String()
	}

	// Create player with initial state (using private fields)
	p := &Player{
		eventBus: f.eventBus,
		id:       playerID,
		name:     name,
		gameID:   gameID,
		resources: types.Resources{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		production: types.Production{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		terraformRating:  20,
		cards:            []string{},
		playedCards:      []string{},
		passed:           false,
		availableActions: 0,
		victoryPoints:    0,
		isConnected:      false,
		effects:          []card.PlayerEffect{},
		actions:          []PlayerAction{},
		resourceStorage:  make(map[string]int),
	}

	return p
}
