package service

import (
	"fmt"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
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

	// Validate that it's the player's turn (basic validation)
	if game.CurrentPlayerID != "" && game.CurrentPlayerID != playerID {
		return nil, fmt.Errorf("not your turn")
	}

	// Apply the action based on DTO type
	switch actionPayload.Type {
	case dto.ActionTypeStandardProjectAsteroid:
		err = s.applyStandardProjectAsteroid(game, player, dto.StandardProjectAsteroidAction{Type: actionPayload.Type})
	case dto.ActionTypeRaiseTemperature:
		if actionPayload.HeatAmount == nil {
			return nil, fmt.Errorf("heat amount is required for raise temperature action")
		}
		err = s.applyRaiseTemperature(game, player, dto.RaiseTemperatureAction{
			Type:       actionPayload.Type,
			HeatAmount: *actionPayload.HeatAmount,
		})
	case dto.ActionTypeSelectCorporation:
		if actionPayload.CorporationName == nil {
			return nil, fmt.Errorf("corporation name is required for select corporation action")
		}
		err = s.applySelectCorporation(game, player, dto.SelectCorporationAction{
			Type:            actionPayload.Type,
			CorporationName: *actionPayload.CorporationName,
		})
	case dto.ActionTypeSkipAction:
		err = s.applySkipAction(game, player, dto.SkipActionAction{Type: actionPayload.Type})
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

// applyStandardProjectAsteroid applies the standard project asteroid action
func (s *GameService) applyStandardProjectAsteroid(game *domain.Game, player *domain.Player, action dto.StandardProjectAsteroidAction) error {
	// Cost: 14 MC, Effect: Raise temperature 1 step
	if player.Resources.Credits < 14 {
		return fmt.Errorf("insufficient credits (need 14, have %d)", player.Resources.Credits)
	}

	if game.GlobalParameters.Temperature >= 8 {
		return fmt.Errorf("temperature already at maximum")
	}

	// Deduct cost
	player.Resources.Credits -= 14

	// Apply effect
	game.GlobalParameters.Temperature += 2 // Each step is 2 degrees

	// Player gains terraform rating
	player.TerraformRating += 1

	return nil
}

// applyRaiseTemperature applies heat to raise temperature
func (s *GameService) applyRaiseTemperature(game *domain.Game, player *domain.Player, action dto.RaiseTemperatureAction) error {
	if action.HeatAmount < 8 {
		return fmt.Errorf("need at least 8 heat to raise temperature")
	}

	if player.Resources.Heat < action.HeatAmount {
		return fmt.Errorf("insufficient heat (need %d, have %d)", action.HeatAmount, player.Resources.Heat)
	}

	if game.GlobalParameters.Temperature >= 8 {
		return fmt.Errorf("temperature already at maximum")
	}

	// Spend 8 heat to raise temperature 1 step
	steps := action.HeatAmount / 8
	player.Resources.Heat -= steps * 8
	game.GlobalParameters.Temperature += steps * 2

	// Cap at maximum
	if game.GlobalParameters.Temperature > 8 {
		game.GlobalParameters.Temperature = 8
	}

	// Player gains terraform rating for each step
	player.TerraformRating += steps

	return nil
}

// applySelectCorporation applies corporation selection
func (s *GameService) applySelectCorporation(game *domain.Game, player *domain.Player, action dto.SelectCorporationAction) error {
	if player.Corporation != "" {
		return fmt.Errorf("player already has a corporation")
	}

	if action.CorporationName == "" {
		return fmt.Errorf("corporation name cannot be empty")
	}

	// TODO: Validate corporation exists and apply starting resources/production
	player.Corporation = action.CorporationName

	return nil
}

// applySkipAction applies skip action
func (s *GameService) applySkipAction(game *domain.Game, player *domain.Player, action dto.SkipActionAction) error {
	// TODO: Implement turn system and move to next player
	return nil
}

// validateGameSettings validates game settings
func (s *GameService) validateGameSettings(settings domain.GameSettings) error {
	if settings.MaxPlayers < 1 || settings.MaxPlayers > 5 {
		return fmt.Errorf("max players must be between 1 and 5")
	}

	return nil
}
