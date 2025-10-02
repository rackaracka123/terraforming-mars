package service

import (
	"context"
	"fmt"
	"math/rand"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameService handles all game lifecycle operations including creation, player management, and state updates
type GameService interface {
	// Create a new game with specified settings
	CreateGame(ctx context.Context, settings model.GameSettings) (model.Game, error)

	// Get game by ID
	GetGame(ctx context.Context, gameID string) (model.Game, error)

	// List games by status
	ListGames(ctx context.Context, status string) ([]model.Game, error)

	// Start a game (transition from status "lobby" to "active")
	StartGame(ctx context.Context, gameID string, playerID string) error

	// Skip a player's turn (advance to next player)
	SkipPlayerTurn(ctx context.Context, gameID string, playerID string) error

	// Add player to game (join game flow)
	JoinGame(ctx context.Context, gameID string, playerName string) (model.Game, error)
	JoinGameWithPlayerID(ctx context.Context, gameID string, playerName string, playerID string) (model.Game, error)

	// Get global parameters (read-only access)
	GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error)

	// Process production phase ready acknowledgment from client
	ProcessProductionPhaseReady(ctx context.Context, gameID string, playerID string) (*model.Game, error)

	// Handle player reconnection - updates connection status and sends complete game state
	PlayerReconnected(ctx context.Context, gameID string, playerID string) error
}

// GameServiceImpl implements GameService interface
type GameServiceImpl struct {
	gameRepo       repository.GameRepository
	playerRepo     repository.PlayerRepository
	cardRepo       repository.CardRepository
	cardService    CardService
	cardDeckRepo   repository.CardDeckRepository
	boardService   BoardService
	sessionManager session.SessionManager
}

// NewGameService creates a new GameService instance
func NewGameService(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardRepo repository.CardRepository,
	cardService CardService,
	cardDeckRepo repository.CardDeckRepository,
	boardService BoardService,
	sessionManager session.SessionManager,
) GameService {
	return &GameServiceImpl{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		cardRepo:       cardRepo,
		cardService:    cardService,
		cardDeckRepo:   cardDeckRepo,
		boardService:   boardService,
		sessionManager: sessionManager,
	}
}

// CreateGame creates a new game with specified settings
func (s *GameServiceImpl) CreateGame(ctx context.Context, settings model.GameSettings) (model.Game, error) {
	log := logger.WithContext()

	// Validate settings
	if err := s.validateGameSettings(settings); err != nil {
		log.Error("Invalid game settings", zap.Error(err))
		return model.Game{}, fmt.Errorf("invalid game settings: %w", err)
	}

	log.Debug("Creating game via GameService")

	game, err := s.gameRepo.Create(ctx, settings)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to create game: %w", err)
	}

	// Initialize the board with default Mars tiles
	board := s.boardService.GenerateDefaultBoard()
	game.Board = board

	// Update the game with the initialized board
	if err := s.gameRepo.UpdateBoard(ctx, game.ID, board); err != nil {
		log.Error("Failed to update game with board", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to initialize game board: %w", err)
	}

	// Initialize the card deck with all project cards
	projectCards, err := s.cardRepo.GetProjectCards(ctx)
	if err != nil {
		log.Error("Failed to get project cards for deck initialization", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to get project cards: %w", err)
	}

	if err := s.cardDeckRepo.InitializeDeck(ctx, game.ID, projectCards); err != nil {
		log.Error("Failed to initialize card deck", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to initialize card deck: %w", err)
	}

	log.Info("Game created via GameService",
		zap.String("game_id", game.ID),
		zap.Int("board_tiles", len(board.Tiles)),
		zap.Int("deck_size", len(projectCards)))
	return game, nil
}

// GetGame retrieves a game by ID
func (s *GameServiceImpl) GetGame(ctx context.Context, gameID string) (model.Game, error) {
	return s.gameRepo.GetByID(ctx, gameID)
}

// GetGameForPlayer gets a game prepared for a specific player's perspective
func (s *GameServiceImpl) GetGameForPlayer(ctx context.Context, gameID string, playerID string) (model.Game, error) {
	// Get the game data
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return model.Game{}, err
	}

	// Get the players separately (clean architecture pattern)
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return model.Game{}, err
	}

	// Create a copy of the game to modify
	gameCopy := game

	// Note: CurrentPlayer and OtherPlayers are legacy fields
	// In clean architecture, frontend should call player repo directly
	// For backward compatibility, we'll skip these fields for now

	// The frontend should use PlayerIds to fetch players when needed
	otherPlayersCap := len(players) - 1
	if otherPlayersCap < 0 {
		otherPlayersCap = 0
	}
	// Clean architecture: Frontend should fetch players separately using PlayerIds
	// No need for OtherPlayers or CurrentPlayer in this layer

	return gameCopy, nil
}

// ListGames lists games by status
func (s *GameServiceImpl) ListGames(ctx context.Context, status string) ([]model.Game, error) {
	return s.gameRepo.List(ctx, status)
}

func (s *GameServiceImpl) StartGame(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Starting game via GameService")

	// Get current game state to validate
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for start", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Get players separately
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for start", zap.Error(err))
		return fmt.Errorf("failed to get players: %w", err)
	}

	// Ensure player is host
	if game.HostPlayerID != playerID {
		log.Warn("Non-host player attempted to start game", zap.String("player_id", playerID))
		return fmt.Errorf("only the host can start the game")
	}

	// Validate game can be started
	if game.Status != model.GameStatusLobby {
		log.Warn("Attempted to start game not in lobby state", zap.String("current_status", string(game.Status)))
		return fmt.Errorf("game is not in lobby state")
	}
	if len(players) < 1 {
		log.Warn("Attempted to start game with no players")
		return fmt.Errorf("cannot start game with no players")
	}

	// Transition game status to active using granular updates
	if err := s.gameRepo.UpdateStatus(ctx, gameID, model.GameStatusActive); err != nil {
		log.Error("Failed to update game status to active", zap.Error(err))
		return fmt.Errorf("failed to update game status: %w", err)
	}

	if err := s.gameRepo.UpdatePhase(ctx, gameID, model.GamePhaseStartingCardSelection); err != nil {
		log.Error("Failed to update game phase", zap.Error(err))
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	// Set the first player's turn (typically the host)
	if len(players) > 0 {
		firstPlayerID := players[0].ID
		if err := s.gameRepo.SetCurrentTurn(ctx, gameID, &firstPlayerID); err != nil {
			log.Error("Failed to set initial current turn", zap.Error(err))
			return fmt.Errorf("failed to set current turn: %w", err)
		}
		log.Info("Set initial current turn", zap.String("first_player_id", firstPlayerID))
	}

	// Distribute starting cards to all players
	if err := s.distributeStartingCards(ctx, gameID, players); err != nil {
		log.Error("Failed to distribute starting cards", zap.Error(err))
		return fmt.Errorf("failed to distribute starting cards: %w", err)
	}

	// If we are starting a game with 1 player, give it unlimited actions (-1)
	if len(players) == 1 {
		soloPlayerID := players[0].ID
		if err := s.playerRepo.UpdateAvailableActions(ctx, gameID, soloPlayerID, -1); err != nil {
			log.Error("Failed to set unlimited actions for solo player", zap.Error(err))
		}
		log.Info("🏃 Solo player detected at game start - granting unlimited actions",
			zap.String("player_id", soloPlayerID),
			zap.Int("total_players", len(players)))
	}

	log.Info("Game started", zap.String("game_id", gameID))

	// Broadcast game state to all players after starting
	err = s.sessionManager.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast game started", zap.Error(err))
		// Don't fail the start operation, just log the error
	}

	return nil
}

// distributeStartingCards gives each player 10 random cards to choose from for starting selection
func (s *GameServiceImpl) distributeStartingCards(ctx context.Context, gameID string, players []model.Player) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Distributing starting cards to players")

	// Get the starting card pool from card service
	startingCards, err := s.cardService.GetStartingCards(ctx)
	if err != nil {
		return fmt.Errorf("failed to get starting card pool: %w", err)
	}

	if len(startingCards) < 10 {
		return fmt.Errorf("insufficient cards in pool: need at least 10, got %d", len(startingCards))
	}

	// Separate cards into groups: cards with requirements and without requirements
	cardsWithRequirements := make([]model.Card, 0)
	cardsWithoutRequirements := make([]model.Card, 0)

	for _, card := range startingCards {
		hasRequirements := len(card.Requirements) > 0

		if hasRequirements {
			cardsWithRequirements = append(cardsWithRequirements, card)
		} else {
			cardsWithoutRequirements = append(cardsWithoutRequirements, card)
		}
	}

	log.Debug("Card pools separated",
		zap.Int("cards_with_requirements", len(cardsWithRequirements)),
		zap.Int("cards_without_requirements", len(cardsWithoutRequirements)))

	// For each player, select 10 cards: at least 1 with requirements, rest random
	for _, player := range players {
		playerStartingCards := make([]model.Card, 0, 10)

		// Always include at least one card with requirements (if available)
		if len(cardsWithRequirements) > 0 {
			// Shuffle cards with requirements and pick one
			requirementCards := make([]model.Card, len(cardsWithRequirements))
			copy(requirementCards, cardsWithRequirements)
			for i := len(requirementCards) - 1; i > 0; i-- {
				j := rand.Intn(i + 1)
				requirementCards[i], requirementCards[j] = requirementCards[j], requirementCards[i]
			}
			playerStartingCards = append(playerStartingCards, requirementCards[0])
		}

		// Fill remaining slots with random cards from the entire pool
		remainingSlots := 10 - len(playerStartingCards)
		if remainingSlots > 0 {
			// Create a pool of remaining cards (excluding already selected)
			remainingPool := make([]model.Card, 0, len(startingCards)-len(playerStartingCards))
			selectedIDs := make(map[string]bool)
			for _, card := range playerStartingCards {
				selectedIDs[card.ID] = true
			}

			for _, card := range startingCards {
				if !selectedIDs[card.ID] {
					remainingPool = append(remainingPool, card)
				}
			}

			// Shuffle remaining pool and take what we need
			for i := len(remainingPool) - 1; i > 0; i-- {
				j := rand.Intn(i + 1)
				remainingPool[i], remainingPool[j] = remainingPool[j], remainingPool[i]
			}

			// Add remaining cards up to 10 total
			cardsToAdd := remainingSlots
			if cardsToAdd > len(remainingPool) {
				cardsToAdd = len(remainingPool)
			}
			playerStartingCards = append(playerStartingCards, remainingPool[:cardsToAdd]...)
		}

		// Convert cards to card IDs
		playerStartingCardIDs := make([]string, len(playerStartingCards))
		for i, card := range playerStartingCards {
			playerStartingCardIDs[i] = card.ID
		}

		// Update the player with starting card IDs
		startingCardsPhase := model.SelectStartingCardsPhase{
			AvailableCards: playerStartingCardIDs,
		}

		if err := s.playerRepo.UpdateSelectStartingCardsPhase(ctx, gameID, player.ID, &startingCardsPhase); err != nil {
			log.Error("Failed to set starting selection for player",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return fmt.Errorf("failed to set starting selection for player %s: %w", player.ID, err)
		}

		log.Info("Distributed starting cards to player",
			zap.String("player_id", player.ID),
			zap.Int("card_count", len(playerStartingCards)))
	}

	log.Info("Successfully distributed starting cards to all players",
		zap.Int("player_count", len(players)))
	return nil
}

// SkipPlayerTurnResult contains the result of skipping a player's turn
type SkipPlayerTurnResult struct {
	AllPlayersPassed bool
	Game             *model.Game
}

func (s *GameServiceImpl) SkipPlayerTurn(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Skipping player turn via GameService")

	// Get current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for skip turn", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game is active
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted to skip turn in non-active game", zap.String("current_status", string(game.Status)))
		return fmt.Errorf("game is not active")
	}

	if game.CurrentTurn == nil {
		log.Warn("Attempted to skip turn but current turn is not set")
		return fmt.Errorf("current turn is not set")
	}

	// Validate requesting player is the current player
	if game.CurrentTurn != nil && *game.CurrentTurn != playerID {
		log.Warn("Non-current player attempted to skip turn",
			zap.String("current_player", *game.CurrentTurn),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("only the current player can skip their turn")
	}

	// Find current player and determine SKIP vs PASS behavior
	currentPlayerIndex := -1
	for i, id := range game.PlayerIDs {
		if id == playerID {
			currentPlayerIndex = i
			break
		}
	}

	if currentPlayerIndex == -1 {
		log.Error("Current player not found in game", zap.String("player_id", playerID))
		return fmt.Errorf("player not found in game")
	}

	gamePlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players for skip turn", zap.Error(err))
	}

	// PASS vs SKIP logic based on available actions
	var currentPlayer *model.Player = nil
	for i := range gamePlayers {
		if gamePlayers[i].ID == playerID {
			currentPlayer = &gamePlayers[i]
		}
	}
	if currentPlayer == nil {
		log.Error("Current player data not found", zap.String("player_id", playerID))
		return fmt.Errorf("player data not found")
	}

	// Check how many players are still active (not passed and have actions)
	activePlayerCount := 0
	for _, player := range gamePlayers {
		if !player.Passed {
			activePlayerCount++
		}
	}

	isPassing := currentPlayer.AvailableActions == 2 || currentPlayer.AvailableActions == -1
	if isPassing {
		// PASS: Player hasn't done any actions or has unlimited actions, mark as passed for generation end check
		err = s.playerRepo.UpdatePassed(ctx, gameID, playerID, true)
		if err != nil {
			log.Error("Failed to mark player as passed", zap.Error(err))
			return fmt.Errorf("failed to update player passed status: %w", err)
		}

		log.Debug("Player PASSED (marked as passed for generation)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", currentPlayer.AvailableActions))

		// If this means only one active player remains, grant them unlimited actions
		if activePlayerCount == 2 {
			for _, player := range gamePlayers {
				if !player.Passed && player.ID != playerID {
					err = s.playerRepo.UpdateAvailableActions(ctx, gameID, player.ID, -1)
					if err != nil {
						log.Error("Failed to grant unlimited actions to last active player", zap.String("player_id", player.ID), zap.Error(err))
						return fmt.Errorf("failed to update last active player's actions: %w", err)
					} else {
						log.Info("🏃 Last active player granted unlimited actions due to others passing", zap.String("player_id", player.ID))
					}
				}
			}
		}

	} else {
		// SKIP: Player has done some actions, just advance turn without passing
		log.Debug("Player SKIPPED (turn advanced, not passed)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", currentPlayer.AvailableActions))
	}

	// List all players again to reflect any status changes
	gamePlayers, err = s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players after skip processing", zap.Error(err))
	}

	// Check if all players have exhausted their actions or formally passed
	allPlayersFinished := true
	passedCount := 0
	playersWithNoActions := 0
	for _, player := range gamePlayers {
		if player.Passed {
			passedCount++
		} else if player.AvailableActions == 0 {
			playersWithNoActions++
		} else if player.AvailableActions > 0 || player.AvailableActions == -1 {
			allPlayersFinished = false
		}
	}

	log.Debug("Checking generation end condition",
		zap.Int("passed_count", passedCount),
		zap.Int("players_with_no_actions", playersWithNoActions),
		zap.Int("total_players", len(gamePlayers)),
		zap.Bool("all_players_finished", allPlayersFinished))

	if allPlayersFinished {
		// All players have either passed or have no actions left - production phase should start
		log.Info("🏭 All players finished their turns - generation ending",
			zap.String("game_id", gameID),
			zap.Int("generation", game.Generation),
			zap.Int("passed_players", passedCount),
			zap.Int("players_with_no_actions", playersWithNoActions))

		_, err = s.executeProductionPhase(ctx, gameID)
		if err != nil {
			log.Error("Failed to execute production phase", zap.Error(err))
			return fmt.Errorf("failed to execute production phase: %w", err)
		}

		// Broadcast updated game state to all players
		if err := s.sessionManager.Broadcast(gameID); err != nil {
			log.Error("Failed to broadcast game state after production phase start", zap.Error(err))
			// Don't fail the skip operation, just log the error
		}

		return nil
	}

	// Find next player who still has actions available
	nextPlayerIndex := (currentPlayerIndex + 1) % len(gamePlayers)

	// Cycle through players to find the next player with available actions
	for i := 0; i < len(gamePlayers); i++ {
		nextPlayer := &gamePlayers[nextPlayerIndex]
		if !nextPlayer.Passed {
			break
		}
		nextPlayerIndex = (nextPlayerIndex + 1) % len(gamePlayers)
	}

	game.CurrentTurn = &gamePlayers[nextPlayerIndex].ID

	// Update game through repository
	if err := s.gameRepo.UpdateCurrentTurn(ctx, game.ID, game.CurrentTurn); err != nil {
		log.Error("Failed to update game after skip turn", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Player turn skipped, advanced to next player",
		zap.String("previous_player", playerID),
		zap.String("current_player", *game.CurrentTurn))

	// Broadcast updated game state to all players
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after skip turn", zap.Error(err))
		// Don't fail the skip operation, just log the error
	}

	return nil
}

// AddPlayerToGame adds a player to the game (clean architecture pattern)
func (s *GameServiceImpl) AddPlayerToGame(ctx context.Context, gameID string, player model.Player) error {
	// Get current game to check max players
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	// Check if game is full
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return err
	}

	if len(players) >= game.Settings.MaxPlayers {
		return fmt.Errorf("game is full")
	}

	// Add player to player repository
	if err := s.playerRepo.Create(ctx, gameID, player); err != nil {
		return err
	}

	// Add player ID to game
	return s.gameRepo.AddPlayerID(ctx, gameID, player.ID)
}

// GetPlayerFromGame returns a player by ID (clean architecture pattern)
func (s *GameServiceImpl) GetPlayerFromGame(ctx context.Context, gameID, playerID string) (model.Player, error) {
	return s.playerRepo.GetByID(ctx, gameID, playerID)
}

// IsGameFull returns true if the game has reached maximum players (clean architecture pattern)
func (s *GameServiceImpl) IsGameFull(ctx context.Context, gameID string) (bool, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return false, err
	}

	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return false, err
	}

	return len(players) >= game.Settings.MaxPlayers, nil
}

// IsHost returns true if the given player ID is the host of the game (business logic from Game model)
func (s *GameServiceImpl) IsHost(game *model.Game, playerID string) bool {
	return game.HostPlayerID == playerID
}

// JoinGame adds a player to a game using both GameState and Player repositories
func (s *GameServiceImpl) JoinGame(ctx context.Context, gameID string, playerName string) (model.Game, error) {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Player joining game via GameService", zap.String("player_name", playerName))

	// Get the current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for join", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to get game: %w", err)
	}

	// Check if game is joinable
	if game.Status == model.GameStatusCompleted {
		log.Warn("Attempted to join completed game", zap.String("player_name", playerName))
		return model.Game{}, fmt.Errorf("cannot join completed game")
	}

	// Check if game is full
	isFull, err := s.IsGameFull(ctx, gameID)
	if err != nil {
		log.Error("Failed to check if game is full", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to check game capacity: %w", err)
	}
	if isFull {
		log.Warn("Attempted to join full game", zap.String("player_name", playerName))
		return model.Game{}, fmt.Errorf("game is full")
	}

	// Check if player with this name already exists to prevent duplicates
	existingPlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for duplicate check", zap.Error(err))
		// Continue with player creation if we can't check for duplicates
	} else {
		for _, player := range existingPlayers {
			if player.Name == playerName {
				log.Debug("Player with this name already exists, returning existing player",
					zap.String("existing_player_id", player.ID),
					zap.String("player_name", playerName))
				// Return existing game state since player already exists
				return game, nil
			}
		}
	}

	// Create new player
	playerID := uuid.New().String()
	player := model.Player{
		ID:   playerID,
		Name: playerName,
		Resources: model.Resources{
			Credits: 40, // Starting credits for standard projects
		},
		Production: model.Production{
			Credits: 1, // Base production
		},
		TerraformRating:  20, // Starting terraform rating
		PlayedCards:      make([]string, 0),
		Passed:           false, // Player starts active, not passed
		AvailableActions: 2,     // Standard actions per turn in action phase
		VictoryPoints:    0,     // Starting victory points
		IsConnected:      true,
	}

	// Add player using clean architecture method
	if err := s.AddPlayerToGame(ctx, gameID, player); err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to add player: %w", err)
	}

	// Host setting is handled automatically by gameRepo.AddPlayerID
	// Get updated game state to return
	updatedGame, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to get updated game: %w", err)
	}

	// If game is in starting_card_selection phase, distribute starting cards to the new player
	if updatedGame.Status == model.GameStatusActive && updatedGame.CurrentPhase == model.GamePhaseStartingCardSelection {
		log.Debug("Game is in starting card selection phase, distributing cards to new player", zap.String("player_id", playerID))

		// Create a slice with just the new player for card distribution
		newPlayerSlice := []model.Player{player}

		if err := s.distributeStartingCards(ctx, gameID, newPlayerSlice); err != nil {
			log.Error("Failed to distribute starting cards to new player", zap.Error(err), zap.String("player_id", playerID))
			// Don't return error here - player joined successfully, just missing starting cards
			// We could handle this gracefully by allowing them to get cards later
		} else {
			log.Info("🃏 Distributed starting cards to late-joining player", zap.String("player_id", playerID))
		}
	}

	log.Debug("Player joined game", zap.String("player_id", playerID))

	// Broadcast updated game state to all players
	err = s.sessionManager.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast player joined", zap.Error(err))
		// Don't fail the join operation, just log the error
	}

	return updatedGame, nil
}

// JoinGameWithPlayerID allows a player to join with a pre-specified player ID
// This is used by WebSocket connections to ensure consistent player ID before broadcasting
func (s *GameServiceImpl) JoinGameWithPlayerID(ctx context.Context, gameID string, playerName string, playerID string) (model.Game, error) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player joining game via GameService with pre-specified ID", zap.String("player_name", playerName))

	// Get the current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for join", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to get game: %w", err)
	}

	// Check if game is joinable
	if game.Status == model.GameStatusCompleted {
		log.Warn("Attempted to join completed game", zap.String("player_name", playerName))
		return model.Game{}, fmt.Errorf("cannot join completed game")
	}

	// Check if game is full
	isFull, err := s.IsGameFull(ctx, gameID)
	if err != nil {
		log.Error("Failed to check if game is full", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to check game capacity: %w", err)
	}
	if isFull {
		log.Warn("Attempted to join full game", zap.String("player_name", playerName))
		return model.Game{}, fmt.Errorf("game is full")
	}

	// Check if player with this name already exists to prevent duplicates
	existingPlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for duplicate check", zap.Error(err))
		// Continue with player creation if we can't check for duplicates
	} else {
		for _, player := range existingPlayers {
			if player.Name == playerName {
				log.Debug("Player with this name already exists, returning existing player",
					zap.String("existing_player_id", player.ID),
					zap.String("player_name", playerName))
				// Return existing game state since player already exists
				return game, nil
			}
		}
	}

	// Create new player with the provided ID
	player := model.Player{
		ID:   playerID,
		Name: playerName,
		Resources: model.Resources{
			Credits: 40, // Starting credits for standard projects
		},
		Production: model.Production{
			Credits: 1, // Base production
		},
		TerraformRating:  20, // Starting terraform rating
		PlayedCards:      make([]string, 0),
		Passed:           false, // Player starts active, not passed
		AvailableActions: 2,     // Standard actions per turn in action phase
		VictoryPoints:    0,     // Starting victory points
		IsConnected:      true,
	}

	// Add player using clean architecture method (same as in JoinGame)
	if err := s.AddPlayerToGame(ctx, gameID, player); err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to add player: %w", err)
	}

	// Host setting is handled automatically by gameRepo.AddPlayerID

	// Get updated game state
	updatedGame, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game", zap.Error(err))
		return model.Game{}, fmt.Errorf("failed to get updated game: %w", err)
	}

	// If game is in starting_card_selection phase, distribute starting cards to the new player
	if updatedGame.Status == model.GameStatusActive && updatedGame.CurrentPhase == model.GamePhaseStartingCardSelection {
		log.Debug("Game is in starting card selection phase, distributing cards to new player", zap.String("player_id", playerID))

		// Create a slice with just the new player for card distribution
		newPlayerSlice := []model.Player{player}

		if err := s.distributeStartingCards(ctx, gameID, newPlayerSlice); err != nil {
			log.Error("Failed to distribute starting cards to new player", zap.Error(err), zap.String("player_id", playerID))
			// Don't return error here - player joined successfully, just missing starting cards
			// We could handle this gracefully by allowing them to get cards later
		} else {
			log.Info("🃏 Distributed starting cards to late-joining player", zap.String("player_id", playerID))
		}
	}

	log.Debug("Player joined game", zap.String("player_id", playerID))

	// Broadcast updated game state to all players
	err = s.sessionManager.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast player joined", zap.Error(err))
		// Don't fail the join operation, just log the error
	}

	return updatedGame, nil
}

// executeProductionPhase updates all players' resources based on their production and advances generation (internal use only)
func (s *GameServiceImpl) executeProductionPhase(ctx context.Context, gameID string) (*model.Game, error) {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Executing production phase via GameService")

	// Get current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for production phase", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game is in correct phase
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted to execute production phase in non-active game", zap.String("current_status", string(game.Status)))
		return nil, fmt.Errorf("game is not active")
	}

	gamePlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players for production phase", zap.Error(err))
		return nil, fmt.Errorf("failed to list players: %w", err)
	}

	// Convert energy to heat and apply production to all players using repository methods
	isSolo := len(gamePlayers) == 1
	resetTurns := 2
	if isSolo {
		resetTurns = -1 // Unlimited actions for solo play
	}

	for i := range gamePlayers {
		player := &gamePlayers[i]

		energyConverted := player.Resources.Energy
		oldResources := player.Resources.DeepCopy()
		newResources := model.Resources{
			Credits:  player.Resources.Credits + player.Production.Credits + player.TerraformRating, // TR provides 1 credit per point
			Steel:    player.Resources.Steel + player.Production.Steel,
			Titanium: player.Resources.Titanium + player.Production.Titanium,
			Plants:   player.Resources.Plants + player.Production.Plants,
			Energy:   player.Production.Energy,
			Heat:     player.Resources.Heat + energyConverted + player.Production.Heat,
		}

		// Update player resources through repository (follows clean architecture)
		if err := s.playerRepo.UpdateResources(ctx, gameID, player.ID, newResources); err != nil {
			log.Error("Failed to update player resources during production",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to update player resources: %w", err)
		}

		// Reset player state for next generation
		if err := s.playerRepo.UpdatePassed(ctx, gameID, player.ID, false); err != nil {
			log.Error("Failed to reset player passed status",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to reset player passed status: %w", err)
		}

		if err := s.playerRepo.UpdateAvailableActions(ctx, gameID, player.ID, resetTurns); err != nil {
			log.Error("Failed to reset player available actions",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to reset player available actions: %w", err)
		}

		// Draw 4 cards from deck for production phase card selection
		drawnCards := []string{}
		for i := 0; i < 4; i++ {
			cardID, err := s.cardDeckRepo.Pop(ctx, gameID)
			if err != nil {
				log.Warn("Failed to draw card from deck for production phase",
					zap.String("player_id", player.ID),
					zap.Int("drawn_so_far", i),
					zap.Error(err))
				// If we can't draw more cards, continue with whatever we have
				break
			}
			drawnCards = append(drawnCards, cardID)
		}

		log.Debug("Drew cards for production phase",
			zap.String("player_id", player.ID),
			zap.Strings("drawn_cards", drawnCards),
			zap.Int("cards_drawn", len(drawnCards)))

		// Set their production phase data
		productionPhaseData := model.ProductionPhase{
			AvailableCards:    drawnCards,
			SelectionComplete: false,
			BeforeResources:   oldResources,
			AfterResources:    newResources,
			EnergyConverted:   energyConverted,
			CreditsIncome:     player.Production.Credits + player.TerraformRating,
		}
		if err := s.playerRepo.UpdateProductionPhase(ctx, gameID, player.ID, &productionPhaseData); err != nil {
			log.Error("Failed to set player production phase data",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to set player production phase data: %w", err)
		}

		log.Debug("Applied production to player",
			zap.String("player_id", player.ID),
			zap.Int("energy_converted", energyConverted),
			zap.Int("credits_gained", player.Production.Credits+player.TerraformRating))
	}

	// Update game through repository
	if err := s.gameRepo.UpdateGeneration(ctx, game.ID, game.Generation+1); err != nil {
		log.Error("Failed to update game generation", zap.Error(err))
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	// Update the current turn to the first player for the new generation
	if len(game.PlayerIDs) > 0 {
		if err := s.gameRepo.UpdateCurrentTurn(ctx, game.ID, &game.PlayerIDs[0]); err != nil {
			log.Error("Failed to update current turn for new generation", zap.Error(err))
			return nil, fmt.Errorf("failed to update current turn: %w", err)
		}
	}

	if err := s.gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseProductionAndCardDraw); err != nil {
		log.Error("Failed to update game phase to action", zap.Error(err))
		return nil, fmt.Errorf("failed to update game phase: %w", err)
	}

	game, err = s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game after production phase", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated game: %w", err)
	}

	log.Info("Production phase start",
		zap.String("game_id", gameID),
		zap.Int("generation", game.Generation),
		zap.Int("player_count", len(gamePlayers)))

	return &game, nil
}

// ProcessProductionPhaseReady processes a player's ready acknowledgment and transitions phase when all players are ready
func (s *GameServiceImpl) ProcessProductionPhaseReady(ctx context.Context, gameID string, playerID string) (*model.Game, error) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Processing production phase ready via GameService")

	// Get current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for production ready", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game state
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted to mark ready in non-active game", zap.String("current_status", string(game.Status)))
		return nil, fmt.Errorf("game is not active")
	}

	if game.CurrentPhase != model.GamePhaseProductionAndCardDraw {
		log.Warn("Attempted to mark ready in non-production phase", zap.String("current_phase", string(game.CurrentPhase)))
		return nil, fmt.Errorf("game is not in production phase")
	}

	// Check if all players are ready
	updatedPlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players to check readiness", zap.Error(err))
		return nil, fmt.Errorf("failed to list players: %w", err)
	}

	allReady := true
	readyCount := 0
	for _, player := range updatedPlayers {
		if player.ProductionPhase.SelectionComplete {
			readyCount++
		} else {
			allReady = false
		}
	}

	log.Debug("Checking production phase readiness",
		zap.Int("ready_count", readyCount),
		zap.Int("total_players", len(updatedPlayers)),
		zap.Bool("all_ready", allReady))

	if allReady {
		// All players are ready - advance to next phase (action phase)
		log.Info("🎯 All players ready for next phase - advancing from production to action",
			zap.String("game_id", gameID),
			zap.Int("generation", game.Generation))

		// Set first player's turn for new generation
		if len(updatedPlayers) > 0 {
			firstPlayerID := updatedPlayers[0].ID
			if err := s.gameRepo.SetCurrentTurn(ctx, gameID, &firstPlayerID); err != nil {
				log.Error("Failed to set current turn for new generation", zap.Error(err))
				return nil, fmt.Errorf("failed to set current turn: %w", err)
			}
		}

		// Reset all player action play counts for the new generation
		if err := s.resetPlayerActionPlayCounts(ctx, gameID); err != nil {
			log.Error("Failed to reset player action play counts", zap.Error(err))
			return nil, fmt.Errorf("failed to reset action play counts: %w", err)
		}

		// Clear production phase data for all players
		for _, player := range updatedPlayers {
			if err := s.playerRepo.UpdateProductionPhase(ctx, gameID, player.ID, nil); err != nil {
				log.Error("Failed to clear player production phase data",
					zap.String("player_id", player.ID),
					zap.Error(err))
				return nil, fmt.Errorf("failed to clear production phase data for player %s: %w", player.ID, err)
			}
		}

		// Advance to action phase
		if err := s.gameRepo.UpdatePhase(ctx, gameID, model.GamePhaseAction); err != nil {
			log.Error("Failed to advance game phase to action", zap.Error(err))
			return nil, fmt.Errorf("failed to advance game phase: %w", err)
		}

		log.Info("Production phase completed, advanced to action phase",
			zap.String("game_id", gameID),
			zap.Int("generation", game.Generation))
	}

	game, err = s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game after production ready", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated game: %w", err)
	}

	err = s.sessionManager.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast game state after production ready", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return &game, nil
}

// validateGameSettings validates game creation settings
func (s *GameServiceImpl) validateGameSettings(settings model.GameSettings) error {
	if settings.MaxPlayers < 1 || settings.MaxPlayers > 5 {
		return fmt.Errorf("max players must be between 1 and 5, got %d", settings.MaxPlayers)
	}

	if settings.Oxygen != nil {
		if *settings.Oxygen < 0 || *settings.Oxygen > 14 {
			return fmt.Errorf("oxygen level must be between 0 and 14, got %d", settings.Oxygen)
		}
	}

	if settings.Temperature != nil {
		if *settings.Temperature < -30 || *settings.Temperature > 8 {
			return fmt.Errorf("temperature must be between -30 and 8, got %d", settings.Temperature)
		}
	}

	if settings.Oceans != nil {
		if *settings.Oceans < 0 || *settings.Oceans > 9 {
			return fmt.Errorf("oceans must be between 0 and 9, got %d", settings.Oceans)
		}
	}
	return nil
}

// GetGlobalParameters gets current global parameters
func (s *GameServiceImpl) GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return model.GlobalParameters{}, err
	}
	return game.GlobalParameters, nil
}

// PlayerReconnected handles player reconnection by updating connection status and sending complete game state
func (s *GameServiceImpl) PlayerReconnected(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithContext().With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
	)

	log.Info("🔄 Processing player reconnection")

	// Update player connection status
	err := s.playerRepo.UpdateConnectionStatus(ctx, gameID, playerID, true)
	if err != nil {
		log.Error("Failed to update player connection status", zap.Error(err))
		// Continue with reconnection even if status update fails
	}

	// Use the session manager to broadcast the complete game state to all players
	err = s.sessionManager.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast game state for reconnection", zap.Error(err))
		return fmt.Errorf("failed to broadcast game state: %w", err)
	}

	log.Info("✅ Player reconnection completed successfully")
	return nil
}

// getPlayerName is a helper method to find player name by ID
func (s *GameServiceImpl) getPlayerName(players []model.Player, playerID string) string {
	for _, player := range players {
		if player.ID == playerID {
			return player.Name
		}
	}
	return "Unknown" // Fallback if player not found
}

// resetPlayerActionPlayCounts resets all player action play counts to 0 for a new generation
func (s *GameServiceImpl) resetPlayerActionPlayCounts(ctx context.Context, gameID string) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("🔄 Resetting player action play counts for new generation")

	// Get all players in the game
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to list players: %w", err)
	}

	// Reset play counts for each player
	for _, player := range players {
		if len(player.Actions) == 0 {
			// Player has no actions, skip
			continue
		}

		// Create a copy of the actions with reset play counts
		resetActions := make([]model.PlayerAction, len(player.Actions))
		for i, action := range player.Actions {
			resetActions[i] = *action.DeepCopy()
			resetActions[i].PlayCount = 0
		}

		// Update the player's actions
		if err := s.playerRepo.UpdatePlayerActions(ctx, gameID, player.ID, resetActions); err != nil {
			log.Error("Failed to reset action play counts for player",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return fmt.Errorf("failed to reset action play counts for player %s: %w", player.ID, err)
		}

		log.Debug("✅ Reset action play counts for player",
			zap.String("player_id", player.ID),
			zap.Int("actions_count", len(resetActions)))
	}

	log.Info("🔄 Successfully reset all player action play counts for new generation",
		zap.Int("players_updated", len(players)))

	return nil
}

// CalculateAvailableOceanHexes returns a list of hex coordinate strings that are available for ocean placement
func (s *GameServiceImpl) CalculateAvailableOceanHexes(ctx context.Context, gameID string) ([]string, error) {
	log := logger.WithGameContext(gameID, "")

	// Get the current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Count actual oceans placed on board (board is source of truth)
	oceansPlaced := 0
	availableHexes := make([]string, 0)

	for _, tile := range game.Board.Tiles {
		if tile.Type == model.ResourceOceanTile {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type == model.ResourceOceanTile {
				// This ocean space is occupied
				oceansPlaced++
			} else {
				// This ocean space is available for placement
				hexKey := fmt.Sprintf("%d,%d,%d", tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S)
				availableHexes = append(availableHexes, hexKey)
			}
		}
	}

	// Check if we've reached maximum oceans based on actual board state
	if oceansPlaced >= model.MaxOceans {
		log.Debug("🌊 Maximum oceans reached, no more ocean placement allowed",
			zap.Int("oceans_placed", oceansPlaced),
			zap.Int("max_oceans", model.MaxOceans))
		return []string{}, nil
	}

	log.Debug("🌊 Ocean hex availability calculated",
		zap.Int("available_hexes", len(availableHexes)),
		zap.Int("oceans_placed", oceansPlaced),
		zap.Int("max_oceans", model.MaxOceans))

	return availableHexes, nil
}
