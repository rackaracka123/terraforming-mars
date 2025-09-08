package service

import (
	"context"
	"fmt"
	"time"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameService handles all game lifecycle operations including creation, player management, and state updates
type GameService interface {
	// Create a new game with specified settings
	CreateGame(ctx context.Context, settings model.GameSettings) (*model.Game, error)

	// Get game by ID
	GetGame(ctx context.Context, gameID string) (*model.Game, error)

	// List games by status
	ListGames(ctx context.Context, status string) ([]*model.Game, error)

	// Start a game (transition from status "lobby" to "active")
	StartGame(ctx context.Context, gameID string, playerID string) error

	// Skip a player's turn (advance to next player)
	SkipPlayerTurn(ctx context.Context, gameID string, playerID string) error

	// Execute production phase (convert energy to heat, generate resources, increment generation)
	ExecuteProductionPhase(ctx context.Context, gameID string) (*model.Game, error)

	// Process production phase ready acknowledgment from clients and resume game when all ready
	ProcessProductionPhaseReady(ctx context.Context, gameID string, playerID string) (*model.Game, error)

	// Add player to game (join game flow)
	JoinGame(ctx context.Context, gameID string, playerName string) (*model.Game, error)
}

// GameServiceImpl implements GameService interface
type GameServiceImpl struct {
	gameRepo       repository.GameRepository
	playerRepo     repository.PlayerRepository
	parametersRepo repository.GlobalParametersRepository
}

// NewGameService creates a new GameService instance
func NewGameService(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	parametersRepo repository.GlobalParametersRepository,
) GameService {
	return &GameServiceImpl{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		parametersRepo: parametersRepo,
	}
}

// CreateGame creates a new game with specified settings
func (s *GameServiceImpl) CreateGame(ctx context.Context, settings model.GameSettings) (*model.Game, error) {
	log := logger.WithContext()

	// Validate settings
	if err := s.validateGameSettings(settings); err != nil {
		log.Error("Invalid game settings", zap.Error(err))
		return nil, fmt.Errorf("invalid game settings: %w", err)
	}

	log.Debug("Creating game via GameService")

	game, err := s.gameRepo.Create(ctx, settings)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	log.Info("Game created via GameService", zap.String("game_id", game.ID))
	return game, nil
}

// GetGame retrieves a game by ID
func (s *GameServiceImpl) GetGame(ctx context.Context, gameID string) (*model.Game, error) {
	return s.gameRepo.Get(ctx, gameID)
}

// ListGames lists games by status
func (s *GameServiceImpl) ListGames(ctx context.Context, status string) ([]*model.Game, error) {
	return s.gameRepo.List(ctx, status)
}

func (s *GameServiceImpl) StartGame(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Starting game via GameService")

	// Get current game state
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for start", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
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
	if len(game.Players) < 1 {
		log.Warn("Attempted to start game with no players")
		return fmt.Errorf("cannot start game with no players")
	}

	// Transition game status to active
	game.Status = model.GameStatusActive

	// Set the first player as the starting current player for testing purposes
	if len(game.Players) > 0 {
		game.CurrentPlayerID = game.Players[0].ID
		// TODO: Implement corporation selection phase
		// For now, skip directly to action phase to test production phase functionality
		game.CurrentPhase = model.GamePhaseAction
		// Set initial action count for testing (should be set by corporation selection)
		game.RemainingActions = 2
		log.Debug("Set starting player", zap.String("current_player_id", game.CurrentPlayerID))
	}

	// Update game through repository
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game status to active", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Game started", zap.String("game_id", gameID))
	return nil
}

func (s *GameServiceImpl) SkipPlayerTurn(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🔄 Skipping player turn via GameService")

	// Get current game state
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for skip turn", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game is active
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted to skip turn in non-active game", zap.String("current_status", string(game.Status)))
		return fmt.Errorf("game is not active")
	}

	// Validate requesting player is the current player
	if game.CurrentPlayerID != playerID {
		log.Warn("Non-current player attempted to skip turn",
			zap.String("current_player", game.CurrentPlayerID),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("only the current player can skip their turn")
	}

	// Find current player and mark them as passed
	currentPlayerIndex := -1
	for i := range game.Players {
		if game.Players[i].ID == playerID {
			currentPlayerIndex = i
			game.Players[i].Passed = true
			log.Debug("👋 Player marked as passed", zap.String("player_name", game.Players[i].Name))
			break
		}
	}

	if currentPlayerIndex == -1 {
		log.Error("Current player not found in game", zap.String("player_id", playerID))
		return fmt.Errorf("player not found in game")
	}

	// Check if all players have passed
	allPassed := true
	passedCount := 0
	for _, player := range game.Players {
		if player.Passed {
			passedCount++
		} else {
			allPassed = false
		}
	}

	log.Debug("📊 Pass status check",
		zap.Int("passed_players", passedCount),
		zap.Int("total_players", len(game.Players)),
		zap.Bool("all_passed", allPassed))

	if allPassed {
		// All players have passed - trigger production phase
		log.Info("🏭 All players passed, triggering production phase",
			zap.Int("generation", game.Generation))

		// Set phase to production temporarily
		game.CurrentPhase = model.GamePhaseProduction

		// Update game state first
		if err := s.gameRepo.Update(ctx, game); err != nil {
			log.Error("Failed to update game before production phase", zap.Error(err))
			return fmt.Errorf("failed to update game: %w", err)
		}

		// Execute production phase
		_, err := s.ExecuteProductionPhase(ctx, gameID)
		if err != nil {
			log.Error("Failed to execute production phase", zap.Error(err))
			return fmt.Errorf("failed to execute production phase: %w", err)
		}

		log.Info("✅ Production phase completed, new generation started")
		return nil
	} else {
		// Not all players have passed - find next active player
		nextPlayerIndex := (currentPlayerIndex + 1) % len(game.Players)

		// Find next player who hasn't passed yet
		attempts := 0
		for game.Players[nextPlayerIndex].Passed && attempts < len(game.Players) {
			nextPlayerIndex = (nextPlayerIndex + 1) % len(game.Players)
			attempts++
		}

		if attempts >= len(game.Players) {
			// This shouldn't happen if allPassed check is correct, but safety check
			log.Error("Could not find next active player - all players appear to have passed")
			return fmt.Errorf("no active players found")
		}

		game.CurrentPlayerID = game.Players[nextPlayerIndex].ID
		game.RemainingActions = 2 // Reset actions for next player

		// Update game through repository
		if err := s.gameRepo.Update(ctx, game); err != nil {
			log.Error("Failed to update game after skip turn", zap.Error(err))
			return fmt.Errorf("failed to update game: %w", err)
		}

		log.Info("➡️ Player turn skipped, advanced to next active player",
			zap.String("previous_player", playerID),
			zap.String("current_player", game.CurrentPlayerID),
			zap.String("next_player_name", game.Players[nextPlayerIndex].Name))
		return nil
	}
}

// AddPlayerToGame adds a player to the game (business logic from Game model)
func (s *GameServiceImpl) AddPlayerToGame(game *model.Game, player model.Player) bool {
	if len(game.Players) >= game.Settings.MaxPlayers {
		return false
	}

	game.Players = append(game.Players, player)
	game.UpdatedAt = time.Now()

	return true
}

// GetPlayerFromGame returns a player by ID (business logic from Game model)
func (s *GameServiceImpl) GetPlayerFromGame(game *model.Game, playerID string) (*model.Player, bool) {
	for i := range game.Players {
		if game.Players[i].ID == playerID {
			return &game.Players[i], true
		}
	}
	return nil, false
}

// IsGameFull returns true if the game has reached maximum players (business logic from Game model)
func (s *GameServiceImpl) IsGameFull(game *model.Game) bool {
	return len(game.Players) >= game.Settings.MaxPlayers
}

// IsHost returns true if the given player ID is the host of the game (business logic from Game model)
func (s *GameServiceImpl) IsHost(game *model.Game, playerID string) bool {
	return game.HostPlayerID == playerID
}

// JoinGame adds a player to a game using both GameState and Player repositories
func (s *GameServiceImpl) JoinGame(ctx context.Context, gameID string, playerName string) (*model.Game, error) {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Player joining game via GameService", zap.String("player_name", playerName))

	// Get the current game state
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for join", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Check if game is joinable
	if game.Status == model.GameStatusCompleted {
		log.Warn("Attempted to join completed game", zap.String("player_name", playerName))
		return nil, fmt.Errorf("cannot join completed game")
	}

	if s.IsGameFull(game) {
		log.Warn("Attempted to join full game",
			zap.String("player_name", playerName),
			zap.Int("current_players", len(game.Players)),
		)
		return nil, fmt.Errorf("game is full")
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

	// Add player through PlayerRepository
	if err := s.playerRepo.AddPlayer(ctx, gameID, player); err != nil {
		log.Error("Failed to add player", zap.Error(err))
		return nil, fmt.Errorf("failed to add player: %w", err)
	}

	// Update game state to include the new player
	if !s.AddPlayerToGame(game, player) {
		log.Error("Failed to add player to game state")
		return nil, fmt.Errorf("failed to add player to game")
	}

	// Set the first player as host if no host is set
	if game.HostPlayerID == "" {
		game.HostPlayerID = player.ID
		log.Debug("Player set as host", zap.String("player_id", playerID))
	}

	// Update game through GameStateRepository
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game after player join", zap.Error(err))
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	log.Debug("Player joined game",
		zap.String("player_id", playerID),
		zap.Int("total_players", len(game.Players)),
	)

	return game, nil
}

// ExecuteProductionPhase executes the production phase for all players
func (s *GameServiceImpl) ExecuteProductionPhase(ctx context.Context, gameID string) (*model.Game, error) {
	log := logger.WithGameContext(gameID, "")
	log.Debug("🏭 Executing production phase via GameService")

	// Get current game state
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for production phase", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game is active
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted to execute production phase in non-active game", zap.String("current_status", string(game.Status)))
		return nil, fmt.Errorf("game is not active")
	}

	log.Info("🎯 Starting production phase",
		zap.Int("generation", game.Generation),
		zap.Int("player_count", len(game.Players)))

	// Step 1: Convert all energy to heat for all players
	for i := range game.Players {
		player := &game.Players[i]
		energyToConvert := player.Resources.Energy

		if energyToConvert > 0 {
			player.Resources.Heat += energyToConvert
			player.Resources.Energy = 0
			log.Debug("⚡→🔥 Converted energy to heat",
				zap.String("player_name", player.Name),
				zap.Int("energy_converted", energyToConvert),
				zap.Int("new_heat_total", player.Resources.Heat))
		}
	}

	// Step 2: Generate resources based on production for all players
	for i := range game.Players {
		player := &game.Players[i]

		// Calculate M€ income = TR + M€ production
		creditsIncome := player.TerraformRating + player.Production.Credits
		player.Resources.Credits += creditsIncome

		// Add other resource production
		player.Resources.Steel += player.Production.Steel
		player.Resources.Titanium += player.Production.Titanium
		player.Resources.Plants += player.Production.Plants
		player.Resources.Energy += player.Production.Energy
		player.Resources.Heat += player.Production.Heat

		log.Debug("💰 Generated resources for player",
			zap.String("player_name", player.Name),
			zap.Int("credits_income", creditsIncome),
			zap.Int("tr", player.TerraformRating),
			zap.Int("credits_production", player.Production.Credits),
			zap.Int("steel_production", player.Production.Steel),
			zap.Int("titanium_production", player.Production.Titanium),
			zap.Int("plants_production", player.Production.Plants),
			zap.Int("energy_production", player.Production.Energy),
			zap.Int("heat_production", player.Production.Heat))
	}

	// Step 3: Reset player states for next generation
	for i := range game.Players {
		player := &game.Players[i]
		player.Passed = false
		player.AvailableActions = 2 // Reset to standard actions per turn
	}

	// Step 4: Advance generation but keep in production phase until clients are ready
	game.Generation++
	// Keep phase as GamePhaseProduction - will be changed to GamePhaseAction when all clients are ready

	// Set first player as current player for new generation (but don't start actions yet)
	if len(game.Players) > 0 {
		game.CurrentPlayerID = game.Players[0].ID
		game.RemainingActions = 2
	}

	// Update game through repository
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game after production phase", zap.Error(err))
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("✅ Production phase executed, waiting for all clients to be ready",
		zap.Int("new_generation", game.Generation),
		zap.String("current_player", game.CurrentPlayerID))

	return game, nil
}

// ProcessProductionPhaseReady handles client ready acknowledgments and resumes game when all ready
func (s *GameServiceImpl) ProcessProductionPhaseReady(ctx context.Context, gameID, playerID string) (*model.Game, error) {
	log := logger.Get().With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
	)
	log.Debug("🏁 Processing production phase ready acknowledgment")

	// Get current game state
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for production phase ready", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game is in production phase
	if game.CurrentPhase != model.GamePhaseProduction {
		log.Warn("Received production phase ready for game not in production phase",
			zap.String("current_phase", string(game.CurrentPhase)))
		return game, nil // Not an error, just ignore
	}

	// Add player to ready list if not already there
	playerAlreadyReady := false
	for _, readyPlayerID := range game.ProductionPhaseReadyPlayers {
		if readyPlayerID == playerID {
			playerAlreadyReady = true
			break
		}
	}

	if !playerAlreadyReady {
		game.ProductionPhaseReadyPlayers = append(game.ProductionPhaseReadyPlayers, playerID)
		log.Info("🏁 Player marked as production phase ready",
			zap.String("player_id", playerID),
			zap.Int("ready_count", len(game.ProductionPhaseReadyPlayers)),
			zap.Int("total_players", len(game.Players)))
	}

	// Check if all players are ready
	allReady := len(game.ProductionPhaseReadyPlayers) >= len(game.Players)

	if allReady {
		log.Info("🏁 All players ready, resuming game action phase")

		// Reset ready tracking for next production phase
		game.ProductionPhaseReadyPlayers = make([]string, 0)

		// Resume game - set phase back to action
		game.CurrentPhase = model.GamePhaseAction
	} else {
		log.Info("🏁 Waiting for more players to be ready",
			zap.Int("ready_count", len(game.ProductionPhaseReadyPlayers)),
			zap.Int("total_players", len(game.Players)))
	}

	// Update game through repository
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game after production phase ready", zap.Error(err))
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	if allReady {
		log.Info("✅ Game resumed from production phase",
			zap.Int("generation", game.Generation),
			zap.String("current_player", game.CurrentPlayerID))
	} else {
		log.Info("⏳ Production phase ready acknowledgment processed, waiting for other players",
			zap.String("player_id", playerID))
	}

	return game, nil
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
