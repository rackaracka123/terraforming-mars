package service

import (
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service/actions"
	"time"

	"github.com/google/uuid"
)

// GameService handles game business logic
type GameService struct {
	gameRepo       *repository.GameRepository
	actionHandlers *actions.ActionHandlers
}

// NewGameService creates a new game service
func NewGameService(gameRepo *repository.GameRepository) *GameService {
	return &GameService{
		gameRepo:       gameRepo,
		actionHandlers: actions.NewActionHandlers(),
	}
}

// CreateGame creates a new game with the given settings
func (s *GameService) CreateGame(settings domain.GameSettings) (*domain.Game, error) {
	// Validate settings
	if err := s.validateGameSettings(settings); err != nil {
		return nil, fmt.Errorf("invalid game settings: %w", err)
	}

	// Create game through repository
	game, err := s.gameRepo.CreateGame(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

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
	// Get the game
	game, err := s.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Check if game is joinable
	if game.Status == domain.GameStatusCompleted {
		return nil, fmt.Errorf("cannot join completed game")
	}

	if game.IsGameFull() {
		return nil, fmt.Errorf("game is full")
	}

	// Create new player
	player := domain.Player{
		ID:   uuid.New().String(),
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
		return nil, fmt.Errorf("failed to add player to game")
	}

	// Set the first player as host if no host is set
	if game.HostPlayerID == "" {
		game.HostPlayerID = player.ID
	}

	// Update game in repository
	if err := s.gameRepo.UpdateGame(game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

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
func (s *GameService) ApplyAction(gameID, playerID string, actionPayload dto.ActionPayload) (*domain.Game, error) {
	// Get the game
	game, err := s.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Find the player
	player, found := game.GetPlayer(playerID)
	if !found {
		return nil, fmt.Errorf("player not found in game")
	}

	// Validate that it's the player's turn (except for start game action)
	if actionPayload.Type != dto.ActionTypeStartGame && game.CurrentPlayerID != "" && game.CurrentPlayerID != playerID {
		return nil, fmt.Errorf("not your turn")
	}

	// Apply the action based on DTO type
	switch actionPayload.Type {
	case dto.ActionTypeStandardProjectAsteroid:
		err = s.actionHandlers.StandardProjectAsteroid.Handle(game, player, actionPayload)
	case dto.ActionTypeRaiseTemperature:
		err = s.actionHandlers.RaiseTemperature.Handle(game, player, actionPayload)
	case dto.ActionTypeSelectCorporation:
		err = s.actionHandlers.SelectCorporation.Handle(game, player, actionPayload)
	case dto.ActionTypeSkipAction:
		err = s.actionHandlers.SkipAction.Handle(game, player, actionPayload)
	case dto.ActionTypeStartGame:
		err = s.actionHandlers.StartGame.Handle(game, player, actionPayload)
	default:
		return nil, fmt.Errorf("unknown action type: %s", actionPayload.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to apply action %s: %w", actionPayload.Type, err)
	}

	// Update the game in repository
	if err := s.gameRepo.UpdateGame(game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return game, nil
}


// validateGameSettings validates game settings
func (s *GameService) validateGameSettings(settings domain.GameSettings) error {
	if settings.MaxPlayers < 1 || settings.MaxPlayers > 5 {
		return fmt.Errorf("max players must be between 1 and 5")
	}

	return nil
}
