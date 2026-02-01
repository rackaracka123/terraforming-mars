// Package contracts defines the interface for the Game State Repository.
// This is a CONTRACT file - it defines the interface, not the implementation.
package contracts

import (
	"context"
)

// GameStateRepository manages game state with diff tracking.
// It provides three core operations: Write, Get, and GetDiff.
type GameStateRepository interface {
	// Write stores the current game state and computes a diff from the previous state.
	// Returns the computed diff representing what changed.
	// For the first write to a game, computes diff from empty state.
	// Thread-safe for concurrent access.
	Write(ctx context.Context, gameID string, state GameState) (*StateDiff, error)

	// Get retrieves the current game state for the specified game.
	// Returns error if game does not exist.
	// Thread-safe for concurrent access.
	Get(ctx context.Context, gameID string) (GameState, error)

	// GetDiff retrieves all diffs for the specified game in chronological order.
	// Returns empty slice if game exists but has no diffs yet.
	// Returns error if game does not exist.
	// Thread-safe for concurrent access.
	GetDiff(ctx context.Context, gameID string) ([]StateDiff, error)
}

// GameState represents the complete state of a game at a point in time.
// This is an alias for the serializable game representation (GameDto).
type GameState interface{}

// StateDiff represents the difference between two consecutive game states.
type StateDiff struct {
	SequenceNumber int64        `json:"sequenceNumber" ts:"number"`
	Timestamp      string       `json:"timestamp" ts:"string"`
	GameID         string       `json:"gameId" ts:"string"`
	Changes        *GameChanges `json:"changes" ts:"GameChanges"`
}

// GameChanges contains all changes in a single state transition.
type GameChanges struct {
	Status              *DiffValueString            `json:"status,omitempty" ts:"DiffValueString | undefined"`
	Phase               *DiffValueString            `json:"phase,omitempty" ts:"DiffValueString | undefined"`
	Generation          *DiffValueInt               `json:"generation,omitempty" ts:"DiffValueInt | undefined"`
	CurrentTurnPlayerID *DiffValueString            `json:"currentTurnPlayerId,omitempty" ts:"DiffValueString | undefined"`
	Temperature         *DiffValueInt               `json:"temperature,omitempty" ts:"DiffValueInt | undefined"`
	Oxygen              *DiffValueInt               `json:"oxygen,omitempty" ts:"DiffValueInt | undefined"`
	Oceans              *DiffValueInt               `json:"oceans,omitempty" ts:"DiffValueInt | undefined"`
	PlayerChanges       map[string]*PlayerChanges   `json:"playerChanges,omitempty" ts:"Record<string, PlayerChanges> | undefined"`
	BoardChanges        *BoardChanges               `json:"boardChanges,omitempty" ts:"BoardChanges | undefined"`
}

// DiffValueString represents old/new values for string fields.
type DiffValueString struct {
	Old string `json:"old" ts:"string"`
	New string `json:"new" ts:"string"`
}

// DiffValueInt represents old/new values for integer fields.
type DiffValueInt struct {
	Old int `json:"old" ts:"number"`
	New int `json:"new" ts:"number"`
}

// DiffValueBool represents old/new values for boolean fields.
type DiffValueBool struct {
	Old bool `json:"old" ts:"boolean"`
	New bool `json:"new" ts:"boolean"`
}

// PlayerChanges contains all changes to a single player's state.
type PlayerChanges struct {
	Credits            *DiffValueInt    `json:"credits,omitempty" ts:"DiffValueInt | undefined"`
	Steel              *DiffValueInt    `json:"steel,omitempty" ts:"DiffValueInt | undefined"`
	Titanium           *DiffValueInt    `json:"titanium,omitempty" ts:"DiffValueInt | undefined"`
	Plants             *DiffValueInt    `json:"plants,omitempty" ts:"DiffValueInt | undefined"`
	Energy             *DiffValueInt    `json:"energy,omitempty" ts:"DiffValueInt | undefined"`
	Heat               *DiffValueInt    `json:"heat,omitempty" ts:"DiffValueInt | undefined"`
	TerraformRating    *DiffValueInt    `json:"terraformRating,omitempty" ts:"DiffValueInt | undefined"`
	CreditsProduction  *DiffValueInt    `json:"creditsProduction,omitempty" ts:"DiffValueInt | undefined"`
	SteelProduction    *DiffValueInt    `json:"steelProduction,omitempty" ts:"DiffValueInt | undefined"`
	TitaniumProduction *DiffValueInt    `json:"titaniumProduction,omitempty" ts:"DiffValueInt | undefined"`
	PlantsProduction   *DiffValueInt    `json:"plantsProduction,omitempty" ts:"DiffValueInt | undefined"`
	EnergyProduction   *DiffValueInt    `json:"energyProduction,omitempty" ts:"DiffValueInt | undefined"`
	HeatProduction     *DiffValueInt    `json:"heatProduction,omitempty" ts:"DiffValueInt | undefined"`
	CardsAdded         []string         `json:"cardsAdded,omitempty" ts:"string[] | undefined"`
	CardsRemoved       []string         `json:"cardsRemoved,omitempty" ts:"string[] | undefined"`
	CardsPlayed        []string         `json:"cardsPlayed,omitempty" ts:"string[] | undefined"`
	Corporation        *DiffValueString `json:"corporation,omitempty" ts:"DiffValueString | undefined"`
	Passed             *DiffValueBool   `json:"passed,omitempty" ts:"DiffValueBool | undefined"`
}

// BoardChanges contains all changes to the game board.
type BoardChanges struct {
	TilesPlaced []TilePlacement `json:"tilesPlaced,omitempty" ts:"TilePlacement[] | undefined"`
}

// TilePlacement records a single tile placement on the board.
type TilePlacement struct {
	HexID    string `json:"hexId" ts:"string"`
	TileType string `json:"tileType" ts:"string"`
	OwnerID  string `json:"ownerId,omitempty" ts:"string | undefined"`
}

// DiffLog contains the complete history of state changes for a game.
type DiffLog struct {
	GameID          string      `json:"gameId" ts:"string"`
	Diffs           []StateDiff `json:"diffs" ts:"StateDiff[]"`
	CurrentSequence int64       `json:"currentSequence" ts:"number"`
}
