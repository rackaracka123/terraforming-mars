package player

import (
	"github.com/google/uuid"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/game/player/actions"
	"terraforming-mars-backend/internal/session/game/player/corporation"
	"terraforming-mars-backend/internal/session/game/player/effects"
	"terraforming-mars-backend/internal/session/game/player/hand"
	"terraforming-mars-backend/internal/session/game/player/selection"
	"terraforming-mars-backend/internal/session/game/player/resources"
	"terraforming-mars-backend/internal/session/game/player/turn"
	"terraforming-mars-backend/internal/session/types"
)

// RequirementModifier represents a discount or lenience that modifies card/standard project requirements
// These are calculated from player effects and automatically updated when card hand or effects change
type RequirementModifier struct {
	Amount                int                    // Modifier amount (discount/lenience value)
	AffectedResources     []types.ResourceType   // types.Resources affected (e.g., ["credits"] for price discount, ["temperature"] for global param)
	CardTarget            *string                // Optional: specific card ID this applies to
	StandardProjectTarget *types.StandardProject // Optional: specific standard project this applies to
}

// PendingTileSelection represents a pending tile placement action
type PendingTileSelection struct {
	TileType       string   // "city", "greenery", "ocean"
	AvailableHexes []string // Backend-calculated valid hex coordinates
	Source         string   // What triggered this selection (card ID, standard project, etc.)
}

// PendingTileSelectionQueue represents a queue of tile placements to be made
type PendingTileSelectionQueue struct {
	Items  []string // Queue of tile types: ["city", "city", "ocean"]
	Source string   // card.Card ID that triggered all placements
}

// ForcedFirstAction represents an action that must be completed as the player's first turn action
// Examples: Tharsis Republic must place a city as their first action
type ForcedFirstAction struct {
	ActionType    string // Type of action: "city_placement", "card_draw", etc.
	CorporationID string // Corporation that requires this action
	Source        string // Source to match for completion (corporation ID)
	Completed     bool   // Whether the forced action has been completed
	Description   string // Human-readable description for UI
}

// Player represents a player in the game with delegated component management.
// Components are independently thread-safe with their own mutexes.
// Player exposes components for direct access by callers.
type Player struct {
	// Identity (immutable after creation)
	id     string
	name   string
	gameID string

	// Infrastructure (private)
	eventBus *events.EventBusImpl

	// Delegated Components (private, exposed via accessors)
	corp      *corporation.Corporation
	hand      *hand.Hand
	resources *resources.Resources
	turn      *turn.Turn
	effects   *effects.Effects
	actions   *actions.Actions
	selection *selection.Selection
}

// ================== Constructor ==================

// NewPlayer creates a new player with initialized components and starting values.
// If playerID is empty, a new UUID will be generated.
func NewPlayer(eventBus *events.EventBusImpl, gameID, playerID, name string) *Player {
	// Generate ID if not provided
	if playerID == "" {
		playerID = uuid.New().String()
	}

	return &Player{
		id:        playerID,
		name:      name,
		gameID:    gameID,
		eventBus:  eventBus,
		corp:      corporation.NewCorporation(),
		hand:      hand.NewHand(),
		resources: resources.NewResources(eventBus, gameID, playerID),
		turn:      turn.NewTurn(),
		effects:   effects.NewEffects(),
		actions:   actions.NewActions(),
		selection: selection.NewSelection(),
	}
}

// ================== Identity Getters (immutable, no locking needed) ==================

func (p *Player) ID() string {
	return p.id
}

func (p *Player) Name() string {
	return p.name
}

func (p *Player) GameID() string {
	return p.gameID
}

// ================== Component Accessors ==================
// Components are thread-safe with their own mutexes.
// Callers interact with components directly (e.g., player.Hand().Cards()).

// Corp returns the corporation component for direct access.
func (p *Player) Corp() *corporation.Corporation {
	return p.corp
}

// Hand returns the hand component for direct access.
func (p *Player) Hand() *hand.Hand {
	return p.hand
}

// Resources returns the resources component for direct access.
func (p *Player) Resources() *resources.Resources {
	return p.resources
}

// Turn returns the turn state component for direct access.
func (p *Player) Turn() *turn.Turn {
	return p.turn
}

// Effects returns the effects component for direct access.
func (p *Player) Effects() *effects.Effects {
	return p.effects
}

// Actions returns the actions component for direct access.
func (p *Player) Actions() *actions.Actions {
	return p.actions
}

// Selection returns the selection component for direct access.
func (p *Player) Selection() *selection.Selection {
	return p.selection
}

// ================== Utility ==================

// Utility methods for phase cards removed - access via Game methods
// Use game.GetProductionPhase(playerID).AvailableCards instead

// DeepCopy creates a deep copy of the Player
func (p *Player) DeepCopy() *Player {
	if p == nil {
		return nil
	}

	// Components handle their own defensive copying, no locking needed
	return &Player{
		id:        p.id,
		name:      p.name,
		gameID:    p.gameID,
		eventBus:  p.eventBus,
		corp:      p.corp.DeepCopy(),
		hand:      p.hand.DeepCopy(),
		resources: p.resources.DeepCopy(),
		turn:      p.turn.DeepCopy(),
		effects:   p.effects.DeepCopy(),
		actions:   p.actions.DeepCopy(),
		selection: p.selection.DeepCopy(),
	}
}
