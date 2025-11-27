package game

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
