package player

import (
	"github.com/google/uuid"
	"terraforming-mars-backend/internal/events"
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

// CreatePlayer creates a new player with initialized state and wired repositories
func (f *Factory) CreatePlayer(gameID, playerID, name string) *Player {
	// Generate ID if not provided
	if playerID == "" {
		playerID = uuid.New().String()
	}

	// Create player with initial state
	p := &types.Player{
		ID:     playerID,
		Name:   name,
		GameID: gameID,
		Resources: types.Resources{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		Production: types.Production{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		TerraformRating:  20,
		Cards:            []string{},
		PlayedCards:      []string{},
		Passed:           false,
		AvailableActions: 0,
		VictoryPoints:    0,
		IsConnected:      false,
		Effects:          []types.PlayerEffect{},
		Actions:          []types.PlayerAction{},
		ResourceStorage:  make(map[string]int),
	}

	// Wrap with repositories and return
	return NewPlayer(p, f.eventBus)
}
