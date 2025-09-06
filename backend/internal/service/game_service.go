package service

import (
	"context"
	"fmt"

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

	// Update game state (internal use for synchronization)
	UpdateGame(ctx context.Context, game *model.Game) error

	// Start a game (transition from status "lobby" to "active")
	StartGame(ctx context.Context, gameID string, playerID string) error

	// Add player to game (join game flow)
	JoinGame(ctx context.Context, gameID string, playerName string) (*model.Game, error)
}

// GameServiceImpl implements GameService interface
type GameServiceImpl struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

// NewGameService creates a new GameService instance
func NewGameService(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
) GameService {
	return &GameServiceImpl{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
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

	log.Info("Creating game via GameService", zap.Int("max_players", settings.MaxPlayers))

	// Create game through repository
	game, err := s.gameRepo.Create(ctx, settings)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Initialize global parameters
	game.GlobalParameters = model.GlobalParameters{
		Temperature: -30, // Mars starting temperature
		Oxygen:      0,   // Starting oxygen level
		Oceans:      0,   // Starting ocean tiles
	}

	// Update game with initial parameters
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game with initial parameters", zap.Error(err))
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Game created successfully via GameService", zap.String("game_id", game.ID))
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

// UpdateGame updates game state
func (s *GameServiceImpl) UpdateGame(ctx context.Context, game *model.Game) error {
	return s.gameRepo.Update(ctx, game)
}

func (s *GameServiceImpl) StartGame(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, "")
	log.Info("Starting game via GameService")

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

	// Update game through repository
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game status to active", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Game started successfully", zap.String("game_id", gameID))
	return nil
}

// JoinGame adds a player to a game using both GameState and Player repositories
func (s *GameServiceImpl) JoinGame(ctx context.Context, gameID string, playerName string) (*model.Game, error) {
	log := logger.WithGameContext(gameID, "")
	log.Info("Player joining game via GameService", zap.String("player_name", playerName))

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

	if game.IsGameFull() {
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
		TerraformRating: 20, // Starting terraform rating
		IsActive:        true,
		PlayedCards:     make([]string, 0),
	}

	// Add player through PlayerRepository
	if err := s.playerRepo.AddPlayer(ctx, gameID, player); err != nil {
		log.Error("Failed to add player", zap.Error(err))
		return nil, fmt.Errorf("failed to add player: %w", err)
	}

	// Update game state to include the new player
	if !game.AddPlayer(player) {
		log.Error("Failed to add player to game state")
		return nil, fmt.Errorf("failed to add player to game")
	}

	// Set the first player as host if no host is set
	if game.HostPlayerID == "" {
		game.HostPlayerID = player.ID
		log.Info("Player set as host", zap.String("player_id", playerID))
	}

	// Update game through GameStateRepository
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game after player join", zap.Error(err))
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Player joined game successfully",
		zap.String("player_id", playerID),
		zap.Int("total_players", len(game.Players)),
	)

	return game, nil
}

// validateGameSettings validates game creation settings
func (s *GameServiceImpl) validateGameSettings(settings model.GameSettings) error {
	if settings.MaxPlayers < 1 || settings.MaxPlayers > 5 {
		return fmt.Errorf("max players must be between 1 and 5, got %d", settings.MaxPlayers)
	}
	return nil
}
