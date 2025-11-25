package core

import (
	"terraforming-mars-backend/internal/session/types"
)

// Game is an alias to the unified Game type
type Game = types.Game

// GamePhase is an alias to the unified GamePhase type
type GamePhase = types.GamePhase

// GameStatus is an alias to the unified GameStatus type
type GameStatus = types.GameStatus

// GameSettings is an alias to the unified GameSettings type
type GameSettings = types.GameSettings

// Re-export constants from types package
const (
	GamePhaseWaitingForGameStart   = types.GamePhaseWaitingForGameStart
	GamePhaseStartingCardSelection = types.GamePhaseStartingCardSelection
	GamePhaseStartGameSelection    = types.GamePhaseStartGameSelection
	GamePhaseAction                = types.GamePhaseAction
	GamePhaseProductionAndCardDraw = types.GamePhaseProductionAndCardDraw
	GamePhaseComplete              = types.GamePhaseComplete

	GameStatusLobby     = types.GameStatusLobby
	GameStatusActive    = types.GameStatusActive
	GameStatusCompleted = types.GameStatusCompleted

	PackBaseGame = types.PackBaseGame
	PackFuture   = types.PackFuture

	DefaultMaxPlayers  = types.DefaultMaxPlayers
	DefaultTemperature = types.DefaultTemperature
	DefaultOxygen      = types.DefaultOxygen
	DefaultOceans      = types.DefaultOceans
)

// NewGame is an alias to the unified NewGame function
var NewGame = types.NewGame

// DefaultCardPacks is an alias to the unified DefaultCardPacks function
var DefaultCardPacks = types.DefaultCardPacks
