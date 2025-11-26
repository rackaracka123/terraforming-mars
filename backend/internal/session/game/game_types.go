package game

import "terraforming-mars-backend/internal/session/types"

// Type aliases for backward compatibility during migration
// These will be removed once all code uses the game package directly
type (
	GamePhase        = types.GamePhase
	GameStatus       = types.GameStatus
	GameSettings     = types.GameSettings
	GlobalParameters = types.GlobalParameters
)

// Re-export constants for backward compatibility
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

	MinTemperature = types.MinTemperature
	MaxTemperature = types.MaxTemperature
	MinOxygen      = types.MinOxygen
	MaxOxygen      = types.MaxOxygen
	MinOceans      = types.MinOceans
	MaxOceans      = types.MaxOceans
)

// Re-export functions for backward compatibility
var DefaultCardPacks = types.DefaultCardPacks
