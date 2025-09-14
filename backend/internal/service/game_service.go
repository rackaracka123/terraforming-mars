package service

import (
	"context"
	"fmt"
	"math/rand"

	"terraforming-mars-backend/internal/events"
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

	// Advance game phase after all players complete starting card selection
	AdvanceFromCardSelectionPhase(ctx context.Context, gameID string) error

	// Skip a player's turn (advance to next player)
	SkipPlayerTurn(ctx context.Context, gameID string, playerID string) (*SkipPlayerTurnResult, error)

	// Add player to game (join game flow)
	JoinGame(ctx context.Context, gameID string, playerName string) (model.Game, error)

	// Global parameters methods (merged from GlobalParametersService)
	UpdateGlobalParameters(ctx context.Context, gameID string, newParams model.GlobalParameters) error
	GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error)
	IncreaseTemperature(ctx context.Context, gameID string, steps int) error
	IncreaseOxygen(ctx context.Context, gameID string, steps int) error
	PlaceOcean(ctx context.Context, gameID string, count int) error

	// Execute production phase (update resources but keep in production phase)
	ExecuteProductionPhase(ctx context.Context, gameID string) (*model.Game, error)

	// Process production phase ready acknowledgment from client
	ProcessProductionPhaseReady(ctx context.Context, gameID string, playerID string) (*model.Game, error)
}

// GameServiceImpl implements GameService interface
type GameServiceImpl struct {
	gameRepo    repository.GameRepository
	playerRepo  repository.PlayerRepository
	cardService *CardServiceImpl // Use concrete type to access StorePlayerCardOptions
	eventBus    events.EventBus
}

// NewGameService creates a new GameService instance
func NewGameService(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardService *CardServiceImpl,
	eventBus events.EventBus,
) GameService {
	return &GameServiceImpl{
		gameRepo:    gameRepo,
		playerRepo:  playerRepo,
		cardService: cardService,
		eventBus:    eventBus,
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

	log.Info("Game created via GameService", zap.String("game_id", game.ID))
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

	log.Info("Game started", zap.String("game_id", gameID))
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

	// For each player, select 10 random cards for starting selection
	for _, player := range players {
		// Create a copy of the starting cards pool for randomization
		cardPool := make([]model.Card, len(startingCards))
		copy(cardPool, startingCards)

		// Shuffle the card pool using Fisher-Yates algorithm
		for i := len(cardPool) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			cardPool[i], cardPool[j] = cardPool[j], cardPool[i]
		}

		// Take the first 10 cards as the player's starting selection
		playerStartingCards := cardPool[:10]

		// Update the player with starting cards
		if err := s.playerRepo.SetStartingSelection(ctx, gameID, player.ID, playerStartingCards); err != nil {
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

func (s *GameServiceImpl) SkipPlayerTurn(ctx context.Context, gameID string, playerID string) (*SkipPlayerTurnResult, error) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Skipping player turn via GameService")

	// Get current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for skip turn", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game is active
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted to skip turn in non-active game", zap.String("current_status", string(game.Status)))
		return nil, fmt.Errorf("game is not active")
	}

	if game.CurrentTurn == nil {
		log.Warn("Attempted to skip turn but current turn is not set")
		return nil, fmt.Errorf("current turn is not set")
	}

	// Validate requesting player is the current player
	if game.CurrentTurn != nil && *game.CurrentTurn != playerID {
		log.Warn("Non-current player attempted to skip turn",
			zap.String("current_player", *game.CurrentTurn),
			zap.String("requesting_player", playerID))
		return nil, fmt.Errorf("only the current player can skip their turn")
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
		return nil, fmt.Errorf("player not found in game")
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
		return nil, fmt.Errorf("player data not found")
	}

	isPassing := currentPlayer.AvailableActions == 2
	if isPassing {
		// PASS: Player hasn't done any actions, mark as passed for generation end check
		err = s.playerRepo.UpdatePassed(ctx, gameID, playerID, true)
		if err != nil {
			log.Error("Failed to mark player as passed", zap.Error(err))
			return nil, fmt.Errorf("failed to update player passed status: %w", err)
		}

		// List all players again to reflect the passed status
		gamePlayers, err = s.playerRepo.ListByGameID(ctx, gameID)
		if err != nil {
			log.Error("Failed to list players after passing", zap.Error(err))
		}

		log.Debug("Player PASSED (marked as passed for generation)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", currentPlayer.AvailableActions))
	} else {
		// SKIP: Player has done some actions, just advance turn without passing
		log.Debug("Player SKIPPED (turn advanced, not passed)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", currentPlayer.AvailableActions))
	}

	// Check if all players have passed
	allPassed := true
	passedCount := 0
	for _, player := range gamePlayers {
		if player.Passed {
			passedCount++
		} else {
			allPassed = false
		}
	}

	log.Debug("Checking passed status",
		zap.Int("passed_count", passedCount),
		zap.Int("total_players", len(gamePlayers)),
		zap.Bool("all_passed", allPassed))

	if allPassed {
		// All players have passed - production phase should start
		log.Info("🏭 All players have passed - generation ending",
			zap.String("game_id", gameID),
			zap.Int("generation", game.Generation))

		return &SkipPlayerTurnResult{
			AllPlayersPassed: true,
			Game:             &game,
		}, nil
	}

	// Find next non-passed player in turn order
	nextPlayerIndex := (currentPlayerIndex + 1) % len(gamePlayers)

	// Cycle through players to find the next non-passed player
	for i := 0; i < len(gamePlayers); i++ {
		if !gamePlayers[nextPlayerIndex].Passed {
			break
		}
		nextPlayerIndex = (nextPlayerIndex + 1) % len(gamePlayers)
	}

	game.CurrentTurn = &gamePlayers[nextPlayerIndex].ID

	// Update game through repository
	if err := s.gameRepo.UpdateCurrentTurn(ctx, game.ID, game.CurrentTurn); err != nil {
		log.Error("Failed to update game after skip turn", zap.Error(err))
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Player turn skipped, advanced to next player",
		zap.String("previous_player", playerID),
		zap.String("current_player", *game.CurrentTurn))

	return &SkipPlayerTurnResult{
		AllPlayersPassed: false,
		Game:             &game,
	}, nil
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
		IsActive:         true,
		PlayedCards:      make([]string, 0),
		Passed:           false, // Player starts active, not passed
		AvailableActions: 2,     // Standard actions per turn in action phase
		VictoryPoints:    0,     // Starting victory points
		MilestoneIcon:    "",    // No milestone achieved initially
		ConnectionStatus: model.ConnectionStatusConnected,
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

	return updatedGame, nil
}

// ExecuteProductionPhase updates all players' resources based on their production and advances generation
func (s *GameServiceImpl) ExecuteProductionPhase(ctx context.Context, gameID string) (*model.Game, error) {
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
	for i := range gamePlayers {
		player := &gamePlayers[i]

		// Get current resources to calculate energy conversion
		energyConverted := player.Resources.Energy

		// Calculate new resources after production
		newResources := model.Resources{
			Credits:  player.Resources.Credits + player.Production.Credits + player.TerraformRating, // TR provides 1 credit per point
			Steel:    player.Resources.Steel + player.Production.Steel,
			Titanium: player.Resources.Titanium + player.Production.Titanium,
			Plants:   player.Resources.Plants + player.Production.Plants,
			Energy:   player.Production.Energy,                                         // Energy resets to production value (old energy converted to heat)
			Heat:     player.Resources.Heat + energyConverted + player.Production.Heat, // Add converted energy + heat production
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

		if err := s.playerRepo.UpdateAvailableActions(ctx, gameID, player.ID, 2); err != nil {
			log.Error("Failed to reset player available actions",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to reset player available actions: %w", err)
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

	if err := s.gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseProductionAndCardDraw); err != nil {
		log.Error("Failed to update game phase to production", zap.Error(err))
		return nil, fmt.Errorf("failed to update game phase: %w", err)
	}

	game, err = s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game after production phase", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated game: %w", err)
	}

	log.Info("Production phase executed",
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

	gamePlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players for production ready", zap.Error(err))
		return nil, fmt.Errorf("failed to list players: %w", err)
	}

	// Validate player exists in game
	playerFound := false
	for _, player := range gamePlayers {
		if player.ID == playerID {
			playerFound = true
			break
		}
	}

	if !playerFound {
		log.Error("Player not found in game", zap.String("player_id", playerID))
		return nil, fmt.Errorf("player not found in game")
	}

	// Check if player is already marked as ready
	var currentPlayer *model.Player
	for _, player := range gamePlayers {
		if player.ID == playerID {
			currentPlayer = &player
			break
		}
	}

	if currentPlayer != nil && currentPlayer.IsReady {
		log.Debug("Player already marked as ready", zap.String("player_id", playerID))
		// Player already ready, just return current game state
		game, err = s.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			log.Error("Failed to get game state", zap.Error(err))
			return nil, fmt.Errorf("failed to get game: %w", err)
		}
		return &game, nil
	}

	// Mark player as ready
	if err := s.playerRepo.UpdateIsReady(ctx, gameID, playerID, true); err != nil {
		log.Error("Failed to mark player as ready", zap.Error(err))
		return nil, fmt.Errorf("failed to mark player as ready: %w", err)
	}

	log.Debug("Player marked as ready for production phase", zap.String("player_id", playerID))

	// Check if all players are ready
	updatedPlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players to check readiness", zap.Error(err))
		return nil, fmt.Errorf("failed to list players: %w", err)
	}

	allReady := true
	readyCount := 0
	for _, player := range updatedPlayers {
		if player.IsReady {
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

		// Reset all players ready status and advance phase
		for _, player := range updatedPlayers {
			if err := s.playerRepo.UpdateIsReady(ctx, gameID, player.ID, false); err != nil {
				log.Error("Failed to reset player ready status",
					zap.String("player_id", player.ID),
					zap.Error(err))
				return nil, fmt.Errorf("failed to reset player ready status: %w", err)
			}
		}

		// Set first player's turn for new generation
		if len(updatedPlayers) > 0 {
			firstPlayerID := updatedPlayers[0].ID
			if err := s.gameRepo.SetCurrentTurn(ctx, gameID, &firstPlayerID); err != nil {
				log.Error("Failed to set current turn for new generation", zap.Error(err))
				return nil, fmt.Errorf("failed to set current turn: %w", err)
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


// AdvanceFromCardSelectionPhase advances the game from starting card selection to action phase
func (s *GameServiceImpl) AdvanceFromCardSelectionPhase(ctx context.Context, gameID string) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Advancing game phase from card selection")

	// Get current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for phase advancement", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Validate current phase
	if game.CurrentPhase != model.GamePhaseStartingCardSelection {
		log.Warn("Attempted to advance from card selection phase but game is not in that phase",
			zap.String("current_phase", string(game.CurrentPhase)))
		return fmt.Errorf("game is not in starting card selection phase")
	}

	// Validate game is active
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted to advance phase but game is not active",
			zap.String("current_status", string(game.Status)))
		return fmt.Errorf("game is not active")
	}

	// Advance to action phase using granular update
	if err := s.gameRepo.UpdatePhase(ctx, gameID, model.GamePhaseAction); err != nil {
		log.Error("Failed to update game phase", zap.Error(err))
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	// Clear temporary card selection data
	if s.cardService != nil {
		s.cardService.ClearGameSelectionData(gameID)
	}

	log.Info("Game phase advanced successfully",
		zap.String("previous_phase", string(model.GamePhaseStartingCardSelection)),
		zap.String("new_phase", string(model.GamePhaseAction)))

	return nil
}

// UpdateGlobalParameters updates global terraforming parameters
func (s *GameServiceImpl) UpdateGlobalParameters(ctx context.Context, gameID string, newParams model.GlobalParameters) error {
	log := logger.WithGameContext(gameID, "")

	log.Info("Updating global parameters via GameService",
		zap.Int("temperature", newParams.Temperature),
		zap.Int("oxygen", newParams.Oxygen),
		zap.Int("oceans", newParams.Oceans))

	// Update through GameRepository using granular update
	return s.gameRepo.UpdateGlobalParameters(ctx, gameID, newParams)
}

// GetGlobalParameters gets current global parameters
func (s *GameServiceImpl) GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return model.GlobalParameters{}, err
	}
	return game.GlobalParameters, nil
}

// IncreaseTemperature increases temperature by specified steps
func (s *GameServiceImpl) IncreaseTemperature(ctx context.Context, gameID string, steps int) error {
	log := logger.WithGameContext(gameID, "")

	// Get current parameters
	params, err := s.GetGlobalParameters(ctx, gameID)
	if err != nil {
		log.Error("Failed to get global parameters", zap.Error(err))
		return fmt.Errorf("failed to get parameters: %w", err)
	}

	// Calculate new temperature (max +8°C)
	newTemp := params.Temperature + (steps * 2) // Each step = 2°C
	if newTemp > 8 {
		newTemp = 8
	}

	// Update parameters
	updatedParams := params
	updatedParams.Temperature = newTemp

	return s.UpdateGlobalParameters(ctx, gameID, updatedParams)
}

// IncreaseOxygen increases oxygen by specified steps
func (s *GameServiceImpl) IncreaseOxygen(ctx context.Context, gameID string, steps int) error {
	log := logger.WithGameContext(gameID, "")

	// Get current parameters
	params, err := s.GetGlobalParameters(ctx, gameID)
	if err != nil {
		log.Error("Failed to get global parameters", zap.Error(err))
		return fmt.Errorf("failed to get parameters: %w", err)
	}

	// Calculate new oxygen level (max 14%)
	newOxygen := params.Oxygen + steps
	if newOxygen > 14 {
		newOxygen = 14
	}

	// Update parameters
	updatedParams := params
	updatedParams.Oxygen = newOxygen

	return s.UpdateGlobalParameters(ctx, gameID, updatedParams)
}

// PlaceOcean places ocean tiles
func (s *GameServiceImpl) PlaceOcean(ctx context.Context, gameID string, count int) error {
	log := logger.WithGameContext(gameID, "")

	// Get current parameters
	params, err := s.GetGlobalParameters(ctx, gameID)
	if err != nil {
		log.Error("Failed to get global parameters", zap.Error(err))
		return fmt.Errorf("failed to get parameters: %w", err)
	}

	// Calculate new ocean count (max 9 oceans)
	newOceans := params.Oceans + count
	if newOceans > 9 {
		newOceans = 9
	}

	// Update parameters
	updatedParams := params
	updatedParams.Oceans = newOceans

	return s.UpdateGlobalParameters(ctx, gameID, updatedParams)
}
