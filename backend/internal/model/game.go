package model

import (
	"time"

	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/features/turn"
)

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
