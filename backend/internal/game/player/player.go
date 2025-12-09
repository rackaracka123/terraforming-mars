package player

import (
	"terraforming-mars-backend/internal/events"
)

// Player represents a player in the game with delegated component management
type Player struct {
	// Identity (immutable)
	id     string
	name   string
	gameID string

	// Connection status
	connected bool

	// Infrastructure
	eventBus *events.EventBusImpl

	// Corporation reference (quick lookup in playedCards)
	corporationID string

	// Turn state
	hasPassed bool

	// Delegated Components (private, exposed via accessors)
	hand             *Hand
	playedCards      *PlayedCards
	resources        *PlayerResources
	selection        *Selection
	actions          *Actions
	effects          *Effects
	standardProjects *StandardProjects
}

// NewPlayer creates a new player with initialized components
// playerID must be provided (generated at handler level for session persistence)
func NewPlayer(eventBus *events.EventBusImpl, gameID, playerID, name string) *Player {
	// Initialize actions component
	actions := NewActions()
	actions.SetInfrastructure(eventBus, playerID)

	return &Player{
		id:               playerID,
		name:             name,
		gameID:           gameID,
		connected:        true, // Players start as connected when created
		eventBus:         eventBus,
		corporationID:    "",
		hasPassed:        false, // Players start not having passed
		hand:             newHand(eventBus, gameID, playerID),
		playedCards:      newPlayedCards(eventBus, gameID, playerID),
		resources:        newResources(eventBus, gameID, playerID),
		selection:        newSelection(eventBus, gameID, playerID),
		actions:          actions,
		effects:          NewEffects(),
		standardProjects: newStandardProjects(),
	}
}

// ==================== Identity ====================

func (p *Player) ID() string {
	return p.id
}

func (p *Player) Name() string {
	return p.name
}

func (p *Player) GameID() string {
	return p.gameID
}

// ==================== Connection Status ====================

func (p *Player) IsConnected() bool {
	return p.connected
}

func (p *Player) SetConnected(connected bool) {
	p.connected = connected

	if p.eventBus != nil {
	}
}

// ==================== Component Accessors ====================

func (p *Player) CorporationID() string {
	return p.corporationID
}

func (p *Player) SetCorporationID(corporationID string) {
	p.corporationID = corporationID
}

func (p *Player) HasCorporation() bool {
	return p.corporationID != ""
}

func (p *Player) Hand() *Hand {
	return p.hand
}

func (p *Player) PlayedCards() *PlayedCards {
	return p.playedCards
}

func (p *Player) Resources() *PlayerResources {
	return p.resources
}

func (p *Player) Selection() *Selection {
	return p.selection
}

func (p *Player) Actions() *Actions {
	return p.actions
}

func (p *Player) Effects() *Effects {
	return p.effects
}

func (p *Player) StandardProjects() *StandardProjects {
	return p.standardProjects
}

// ==================== Turn State ====================

func (p *Player) HasPassed() bool {
	return p.hasPassed
}

func (p *Player) SetPassed(passed bool) {
	p.hasPassed = passed

	if p.eventBus != nil {
	}
}
