package core

// DEPRECATED: This file provides backward compatibility aliases
// Game type moved to parent game package but we can't import it (import cycle)
// TODO: Remove core package entirely and move repositories to game package

import (
	"terraforming-mars-backend/internal/session/types"
)

// Game - REMOVED: Can't alias parent package due to import cycle
// Use session/game.Game directly instead of core.Game
// type Game = game.Game

// GamePhase is an alias (still in types for now)
type GamePhase = types.GamePhase

// GameStatus is an alias (still in types for now)
type GameStatus = types.GameStatus

// GameSettings is an alias (still in types for now)
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

// NewGame - REMOVED: Can't alias parent package due to import cycle
// Use game.NewGame directly instead
// var NewGame = game.NewGame

// DefaultCardPacks is an alias to the unified DefaultCardPacks function
var DefaultCardPacks = types.DefaultCardPacks
