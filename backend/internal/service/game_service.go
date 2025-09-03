package service

import (
	"fmt"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/repository"
	"time"

	"github.com/google/uuid"
)

// GameService handles game business logic
type GameService struct {
	gameRepo *repository.GameRepository
}

// NewGameService creates a new game service
func NewGameService(gameRepo *repository.GameRepository) *GameService {
	return &GameService{
		gameRepo: gameRepo,
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

// GetAvailableGames returns games that can be joined
func (s *GameService) GetAvailableGames() ([]*domain.Game, error) {
	allGames, err := s.gameRepo.ListGames()
	if err != nil {
		return nil, err
	}

	availableGames := make([]*domain.Game, 0)
	for _, game := range allGames {
		if game.Status == domain.GameStatusWaiting && !game.IsGameFull() {
			availableGames = append(availableGames, game)
		}
	}

	return availableGames, nil
}

// validateGameSettings validates game settings
func (s *GameService) validateGameSettings(settings domain.GameSettings) error {
	if settings.MaxPlayers < 1 || settings.MaxPlayers > 5 {
		return fmt.Errorf("max players must be between 1 and 5")
	}

	return nil
}
