package player

import (
	"github.com/google/uuid"
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
	hand        *Hand
	playedCards *PlayedCards
	resources   *PlayerResources
	selection   *Selection
}

// NewPlayer creates a new player with initialized components
func NewPlayer(eventBus *events.EventBusImpl, gameID, playerID, name string) *Player {
	if playerID == "" {
		playerID = uuid.New().String()
	}

	return &Player{
		id:            playerID,
		name:          name,
		gameID:        gameID,
		connected:     true, // Players start as connected when created
		eventBus:      eventBus,
		corporationID: "",
		hasPassed:     false, // Players start not having passed
		hand:          newHand(eventBus, gameID, playerID),
		playedCards:   newPlayedCards(eventBus, gameID, playerID),
		resources:     newResources(eventBus, gameID, playerID),
		selection:     newSelection(),
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

	// Publish broadcast event to notify clients of connection status change
	if p.eventBus != nil {
		events.Publish(p.eventBus, events.BroadcastEvent{
			GameID:    p.gameID,
			PlayerIDs: []string{p.id},
		})
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

// ==================== Turn State ====================

func (p *Player) HasPassed() bool {
	return p.hasPassed
}

func (p *Player) SetPassed(passed bool) {
	p.hasPassed = passed

	// Publish broadcast event to notify clients of pass state change
	if p.eventBus != nil {
		events.Publish(p.eventBus, events.BroadcastEvent{
			GameID:    p.gameID,
			PlayerIDs: []string{p.id},
		})
	}
}
