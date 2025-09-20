package store

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/model"
)

// ApplicationState represents the entire application state
type ApplicationState struct {
	games        map[string]GameState   `json:"games" ts:"Record<string, GameState>"`
	players      map[string]PlayerState `json:"players" ts:"Record<string, PlayerState>"`
	cardRegistry *cards.CardRegistry    `json:"cardRegistry" ts:"CardRegistry"`
}

// GameState represents the state of a single game
type GameState struct {
	game    model.Game `json:"game" ts:"Game"`
	players []string   `json:"players" ts:"string[]"` // Player IDs in this game
}

// PlayerState represents the state of a single player
type PlayerState struct {
	player model.Player `json:"player" ts:"Player"`
	gameID string       `json:"gameId" ts:"string"`
}

// NewApplicationState creates a new empty application state
func NewApplicationState() ApplicationState {
	return ApplicationState{
		games:        make(map[string]GameState),
		players:      make(map[string]PlayerState),
		cardRegistry: nil, // Will be set during initialization
	}
}

// NewGameState creates a new game state
func NewGameState(game model.Game) GameState {
	playersCopy := make([]string, len(game.PlayerIDs))
	copy(playersCopy, game.PlayerIDs)
	return GameState{
		game:    game,
		players: playersCopy,
	}
}

// NewPlayerState creates a new player state
func NewPlayerState(player model.Player, gameID string) PlayerState {
	return PlayerState{
		player: player,
		gameID: gameID,
	}
}

// Getter methods return defensive copies
func (s ApplicationState) Games() map[string]GameState {
	gamesCopy := make(map[string]GameState)
	for k, v := range s.games {
		gamesCopy[k] = v
	}
	return gamesCopy
}

func (s ApplicationState) Players() map[string]PlayerState {
	playersCopy := make(map[string]PlayerState)
	for k, v := range s.players {
		playersCopy[k] = v
	}
	return playersCopy
}

func (s ApplicationState) CardRegistry() *cards.CardRegistry {
	return s.cardRegistry
}

// GetGame returns a game by ID (returns copy)
func (s ApplicationState) GetGame(gameID string) (GameState, bool) {
	game, exists := s.games[gameID]
	return game, exists
}

// GetPlayer returns a player by ID (returns copy)
func (s ApplicationState) GetPlayer(playerID string) (PlayerState, bool) {
	player, exists := s.players[playerID]
	return player, exists
}

// GetGamePlayers returns all players in a game (returns copies)
func (s ApplicationState) GetGamePlayers(gameID string) []PlayerState {
	game, exists := s.games[gameID]
	if !exists {
		return nil
	}

	players := make([]PlayerState, 0, len(game.players))
	for _, playerID := range game.players {
		if player, exists := s.players[playerID]; exists {
			players = append(players, player)
		}
	}
	return players
}

// "With" methods for immutable updates

// WithGame returns a new ApplicationState with the specified game added/updated
func (s ApplicationState) WithGame(gameID string, gameState GameState) ApplicationState {
	newGames := make(map[string]GameState)
	for k, v := range s.games {
		newGames[k] = v
	}
	newGames[gameID] = gameState
	return ApplicationState{
		games:        newGames,
		players:      s.players,
		cardRegistry: s.cardRegistry,
	}
}

// WithPlayer returns a new ApplicationState with the specified player added/updated
func (s ApplicationState) WithPlayer(playerID string, playerState PlayerState) ApplicationState {
	newPlayers := make(map[string]PlayerState)
	for k, v := range s.players {
		newPlayers[k] = v
	}
	newPlayers[playerID] = playerState
	return ApplicationState{
		games:        s.games,
		players:      newPlayers,
		cardRegistry: s.cardRegistry,
	}
}

// WithoutGame returns a new ApplicationState with the specified game removed
func (s ApplicationState) WithoutGame(gameID string) ApplicationState {
	newGames := make(map[string]GameState)
	for k, v := range s.games {
		if k != gameID {
			newGames[k] = v
		}
	}

	// Remove all players from this game
	newPlayers := make(map[string]PlayerState)
	for k, v := range s.players {
		if v.gameID != gameID {
			newPlayers[k] = v
		}
	}

	return ApplicationState{
		games:        newGames,
		players:      newPlayers,
		cardRegistry: s.cardRegistry,
	}
}

// WithoutPlayer returns a new ApplicationState with the specified player removed
func (s ApplicationState) WithoutPlayer(playerID string) ApplicationState {
	player, exists := s.players[playerID]
	if !exists {
		return s
	}

	newPlayers := make(map[string]PlayerState)
	for k, v := range s.players {
		if k != playerID {
			newPlayers[k] = v
		}
	}

	// Remove player from game's player list
	newGames := make(map[string]GameState)
	for k, v := range s.games {
		if k == player.gameID {
			newGames[k] = v.WithoutPlayer(playerID)
		} else {
			newGames[k] = v
		}
	}

	return ApplicationState{
		games:        newGames,
		players:      newPlayers,
		cardRegistry: s.cardRegistry,
	}
}

// WithCardRegistry returns a new ApplicationState with the card registry set
func (s ApplicationState) WithCardRegistry(cardRegistry *cards.CardRegistry) ApplicationState {
	return ApplicationState{
		games:        s.games,
		players:      s.players,
		cardRegistry: cardRegistry,
	}
}

// GameState immutable methods

// Game returns the embedded game (copy)
func (gs GameState) Game() model.Game {
	return gs.game
}

// Players returns the player IDs (copy)
func (gs GameState) Players() []string {
	playersCopy := make([]string, len(gs.players))
	copy(playersCopy, gs.players)
	return playersCopy
}

// WithGame returns a new GameState with updated game
func (gs GameState) WithGame(game model.Game) GameState {
	return GameState{
		game:    game,
		players: gs.players,
	}
}

// WithPlayer returns a new GameState with a player added
func (gs GameState) WithPlayer(playerID string) GameState {
	newPlayers := make([]string, len(gs.players)+1)
	copy(newPlayers, gs.players)
	newPlayers[len(gs.players)] = playerID
	return GameState{
		game:    gs.game,
		players: newPlayers,
	}
}

// WithoutPlayer returns a new GameState with a player removed
func (gs GameState) WithoutPlayer(playerID string) GameState {
	newPlayers := make([]string, 0, len(gs.players))
	for _, id := range gs.players {
		if id != playerID {
			newPlayers = append(newPlayers, id)
		}
	}
	return GameState{
		game:    gs.game,
		players: newPlayers,
	}
}

// PlayerState immutable methods

// Player returns the embedded player (copy)
func (ps PlayerState) Player() model.Player {
	return ps.player
}

// GameID returns the game ID
func (ps PlayerState) GameID() string {
	return ps.gameID
}

// WithPlayer returns a new PlayerState with updated player
func (ps PlayerState) WithPlayer(player model.Player) PlayerState {
	return PlayerState{
		player: player,
		gameID: ps.gameID,
	}
}

// WithGameID returns a new PlayerState with updated game ID
func (ps PlayerState) WithGameID(gameID string) PlayerState {
	return PlayerState{
		player: ps.player,
		gameID: gameID,
	}
}
