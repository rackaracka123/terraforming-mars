package game

import (
	"time"

	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/shared/types"
)

// Type aliases to avoid circular imports
type HexPosition = types.HexPosition
type TileOccupant = tiles.TileOccupant
type Board = tiles.Board
type GlobalParameters = parameters.GlobalParameters

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseWaitingForGameStart   GamePhase = "waiting_for_game_start"
	GamePhaseStartingCardSelection GamePhase = "starting_card_selection"
	GamePhaseStartGameSelection    GamePhase = "start_game_selection"
	GamePhaseAction                GamePhase = "action"
	GamePhaseProductionAndCardDraw GamePhase = "production_and_card_draw"
	GamePhaseComplete              GamePhase = "complete"
)

// GameStatus represents the current status of the game
type GameStatus string

const (
	GameStatusLobby     GameStatus = "lobby"
	GameStatusActive    GameStatus = "active"
	GameStatusCompleted GameStatus = "completed"
)

// GameSettings contains configurable game parameters (all optional)
type GameSettings struct {
	MaxPlayers      int      `json:"maxPlayers,omitempty" ts:"number"`              // Default: 5
	Temperature     *int     `json:"temperature,omitempty" ts:"number | undefined"` // Default: -30°C
	Oxygen          *int     `json:"oxygen,omitempty" ts:"number | undefined"`      // Default: 0%
	Oceans          *int     `json:"oceans,omitempty" ts:"number | undefined"`      // Default: 0
	DevelopmentMode bool     `json:"developmentMode,omitempty" ts:"boolean"`        // Default: false
	CardPacks       []string `json:"cardPacks,omitempty" ts:"string[] | undefined"` // Default: ["base-game"]
}

// Card pack constants
const (
	PackBaseGame = "base-game" // Tested simple cards only
	PackFuture   = "future"    // Untested/complex cards for future implementation
)

// Default values for game settings
const (
	DefaultMaxPlayers  = 5
	DefaultTemperature = parameters.MinTemperature // -30°C
	DefaultOxygen      = parameters.MinOxygen      // 0%
	DefaultOceans      = parameters.MinOceans      // 0
)

// DefaultCardPacks returns the default card packs
func DefaultCardPacks() []string {
	return []string{PackBaseGame}
}

// Game represents a unified game entity with service references to features
type Game struct {
	// Metadata
	ID           string       `json:"id" ts:"string"`
	CreatedAt    time.Time    `json:"createdAt" ts:"string"`
	UpdatedAt    time.Time    `json:"updatedAt" ts:"string"`
	Status       GameStatus   `json:"status" ts:"GameStatus"`
	Settings     GameSettings `json:"settings" ts:"GameSettings"`
	PlayerIDs    []string     `json:"playerIds" ts:"string[]"`
	HostPlayerID string       `json:"hostPlayerId" ts:"string"`
	CurrentPhase GamePhase    `json:"currentPhase" ts:"GamePhase"`
	Generation   int          `json:"generation" ts:"number"`

	// View context (not persisted, set per request)
	ViewingPlayerID string `json:"viewingPlayerId" ts:"string"` // The player viewing this game state

	// Feature Services (game-level state management)
	ParametersService parameters.Service    `json:"-"` // Global parameters (temperature, oxygen, oceans)
	BoardService      tiles.BoardService    `json:"-"` // Board with tile placement
	TurnOrderService  turn.TurnOrderService `json:"-"` // Current turn and player rotation
}

// NewGame creates a new game with the given settings
// Feature services must be injected after creation
func NewGame(id string, settings GameSettings) *Game {
	now := time.Now()

	return &Game{
		ID:           id,
		CreatedAt:    now,
		UpdatedAt:    now,
		Status:       GameStatusLobby,
		Settings:     settings,
		PlayerIDs:    make([]string, 0),
		CurrentPhase: GamePhaseWaitingForGameStart,
		Generation:   1,
		// Services will be injected during initialization
	}
}

// GetGlobalParameters retrieves global parameters via service
func (g *Game) GetGlobalParameters() (parameters.GlobalParameters, error) {
	if g.ParametersService == nil {
		return parameters.GlobalParameters{}, nil
	}
	return g.ParametersService.GetGlobalParameters(nil)
}

// GetBoard retrieves board via service
func (g *Game) GetBoard() (tiles.Board, error) {
	if g.BoardService == nil {
		return tiles.Board{}, nil
	}
	return g.BoardService.GetBoard(nil)
}

// GetCurrentTurn retrieves current turn via service
func (g *Game) GetCurrentTurn() (*string, error) {
	if g.TurnOrderService == nil {
		return nil, nil
	}
	return g.TurnOrderService.GetCurrentTurn(nil)
}

// Next returns the next player ID in turn order
func (g *Game) Next() *string {
	if g.TurnOrderService == nil || len(g.PlayerIDs) == 0 {
		return nil
	}

	// Use service to advance turn
	nextPlayerID, err := g.TurnOrderService.AdvanceTurn(nil)
	if err != nil {
		return nil
	}

	return nextPlayerID
}
