package game

import (
	"time"

	"terraforming-mars-backend/internal/model"
)

// Game represents a unified game entity containing both metadata and state
type Game struct {
	ID               string                 `json:"id"`
	CreatedAt        time.Time              `json:"createdAt"`
	UpdatedAt        time.Time              `json:"updatedAt"`
	Status           GameStatus             `json:"status"`
	Settings         GameSettings           `json:"settings"`
	PlayerIDs        []string               `json:"playerIds"` // Player IDs in this game
	HostPlayerID     string                 `json:"hostPlayerId"`
	CurrentPhase     GamePhase              `json:"currentPhase"`
	GlobalParameters model.GlobalParameters `json:"globalParameters"`
	ViewingPlayerID  string                 `json:"viewingPlayerId"` // The player viewing this game state
	CurrentTurn      *string                `json:"currentTurn"`     // Whose turn it is (nullable)
	Generation       int                    `json:"generation"`
	Board            model.Board            `json:"board"` // Game board with tiles and occupancy state
}

// NewGame creates a new game with the given settings
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
		GlobalParameters: model.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
		Generation: 1,
		Board:      model.Board{Tiles: []model.Tile{}}, // Initialize with empty board
	}
}

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
	MaxPlayers      int      `json:"maxPlayers,omitempty"`
	Temperature     *int     `json:"temperature,omitempty"`
	Oxygen          *int     `json:"oxygen,omitempty"`
	Oceans          *int     `json:"oceans,omitempty"`
	DevelopmentMode bool     `json:"developmentMode,omitempty"`
	CardPacks       []string `json:"cardPacks,omitempty"`
}

// Card pack constants
const (
	PackBaseGame = "base-game"
	PackFuture   = "future"
)

// Default values for game settings
const (
	DefaultMaxPlayers  = 5
	DefaultTemperature = -30
	DefaultOxygen      = 0
	DefaultOceans      = 0
)

// DefaultCardPacks returns the default card packs
func DefaultCardPacks() []string {
	return []string{PackBaseGame}
}
