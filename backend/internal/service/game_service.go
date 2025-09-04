package service

import (
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service/actions"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameService handles game business logic
type GameService struct {
	gameRepo       *repository.GameRepository
	actionHandlers *actions.ActionHandlers
	eventBus       events.EventBus
}

// NewGameService creates a new game service
func NewGameService(gameRepo *repository.GameRepository, eventBus events.EventBus) *GameService {
	return &GameService{
		gameRepo:       gameRepo,
		actionHandlers: actions.NewActionHandlers(eventBus),
		eventBus:       eventBus,
	}
}

// CreateGame creates a new game with the given settings
func (s *GameService) CreateGame(settings domain.GameSettings) (*domain.Game, error) {
	log := logger.Get()
	
	log.Info("Creating game", zap.Int("max_players", settings.MaxPlayers))
	
	// Validate settings
	if err := s.validateGameSettings(settings); err != nil {
		log.Error("Invalid game settings", zap.Error(err), zap.Int("max_players", settings.MaxPlayers))
		return nil, fmt.Errorf("invalid game settings: %w", err)
	}

	// Create game through repository
	game, err := s.gameRepo.CreateGame(settings)
	if err != nil {
		log.Error("Failed to create game in repository", zap.Error(err))
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	log.Info("Game created successfully", 
		zap.String("game_id", game.ID),
		zap.Int("max_players", settings.MaxPlayers),
	)

	return game, nil
}

// GetGame retrieves a game by ID
func (s *GameService) GetGame(gameID string) (*domain.Game, error) {
	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	game, err := s.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return game, nil
}

// JoinGame adds a player to a game
func (s *GameService) JoinGame(gameID string, playerName string) (*domain.Game, error) {
	log := logger.WithGameContext(gameID, "")
	
	log.Info("Player attempting to join game", zap.String("player_name", playerName))
	
	// Get the game
	game, err := s.gameRepo.GetGame(gameID)
	if err != nil {
		log.Error("Failed to get game for join", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Check if game is joinable
	if game.Status == domain.GameStatusCompleted {
		log.Warn("Player attempted to join completed game", zap.String("player_name", playerName))
		return nil, fmt.Errorf("cannot join completed game")
	}

	if game.IsGameFull() {
		log.Warn("Player attempted to join full game", 
			zap.String("player_name", playerName),
			zap.Int("current_players", len(game.Players)),
		)
		return nil, fmt.Errorf("game is full")
	}

	// Create new player
	playerID := uuid.New().String()
	player := domain.Player{
		ID:   playerID,
		Name: playerName,
		Resources: domain.Resources{
			Credits: 0,
		},
		Production: domain.Production{
			Credits: 1, // Base production
		},
		TerraformRating: 20, // Starting terraform rating
		IsActive:        true,
		PlayedCards:     make([]string, 0),
	}

	// Add player to game
	if !game.AddPlayer(player) {
		log.Error("Failed to add player to game", 
			zap.String("player_name", playerName),
			zap.String("player_id", playerID),
		)
		return nil, fmt.Errorf("failed to add player to game")
	}

	// Set the first player as host if no host is set
	if game.HostPlayerID == "" {
		game.HostPlayerID = player.ID
		log.Info("Player set as host", 
			zap.String("player_name", playerName),
			zap.String("player_id", playerID),
		)
	}

	// Update game in repository
	if err := s.gameRepo.UpdateGame(game); err != nil {
		log.Error("Failed to update game after player join", zap.Error(err))
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Player joined game successfully", 
		zap.String("player_name", playerName),
		zap.String("player_id", playerID),
		zap.Int("total_players", len(game.Players)),
	)

	return game, nil
}

// ListGames returns all games, optionally filtered by status
func (s *GameService) ListGames(status string) ([]*domain.Game, error) {
	if status == "" {
		return s.gameRepo.ListGames()
	}

	return s.gameRepo.GetGamesByStatus(status)
}

// UpdateGame updates a game
func (s *GameService) UpdateGame(game *domain.Game) error {
	if game == nil {
		return fmt.Errorf("game cannot be nil")
	}

	game.UpdatedAt = time.Now()

	return s.gameRepo.UpdateGame(game)
}

// ApplyAction validates and applies a game action using DTO types
func (s *GameService) ApplyAction(gameID, playerID string, actionRequest interface{}) (*domain.Game, error) {
	log := logger.WithGameContext(gameID, playerID)
	
	// Get the game
	game, err := s.gameRepo.GetGame(gameID)
	if err != nil {
		log.Error("Failed to get game for action", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Find the player
	player, found := game.GetPlayer(playerID)
	if !found {
		log.Error("Player not found in game")
		return nil, fmt.Errorf("player not found in game")
	}

	// Apply the action based on the request type
	switch request := actionRequest.(type) {
	case dto.ActionSelectStartingCardRequest:
		log.Info("Applying select starting card action")
		err = s.actionHandlers.SelectStartingCards.Handle(game, player, request)
	case dto.ActionStartGameRequest:
		log.Info("Applying start game action")
		// Validate that it's the host for start game action
		if !game.IsHost(playerID) {
			return nil, fmt.Errorf("only the host can start the game")
		}
		err = s.actionHandlers.StartGame.Handle(game, player, request)
	case dto.ActionPlayCardRequest:
		log.Info("Applying play card action")
		// Validate that it's the player's turn for play card action
		if game.CurrentPlayerID != "" && game.CurrentPlayerID != playerID {
			log.Warn("Player attempted action out of turn", 
				zap.String("current_player", game.CurrentPlayerID),
			)
			return nil, fmt.Errorf("not your turn")
		}
		err = s.actionHandlers.PlayCard.Handle(game, player, request)
	default:
		log.Error("Unknown action request type", zap.Any("request_type", request))
		return nil, fmt.Errorf("unknown action request type")
	}

	if err != nil {
		log.Error("Failed to apply action", zap.Error(err))
		return nil, fmt.Errorf("failed to apply action: %w", err)
	}

	// Update the game in repository
	if err := s.gameRepo.UpdateGame(game); err != nil {
		log.Error("Failed to update game after action", zap.Error(err))
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Game action applied successfully")
	return game, nil
}


// validateGameSettings validates game settings
func (s *GameService) validateGameSettings(settings domain.GameSettings) error {
	if settings.MaxPlayers < 1 || settings.MaxPlayers > 5 {
		return fmt.Errorf("max players must be between 1 and 5")
	}

	return nil
}
