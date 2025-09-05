package service

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// PlayerService handles player-specific operations
type PlayerService interface {
	// Update player resources
	UpdatePlayerResources(ctx context.Context, gameID, playerID string, newResources model.Resources) error

	// Update player production
	UpdatePlayerProduction(ctx context.Context, gameID, playerID string, newProduction model.Production) error

	// Get player information
	GetPlayer(ctx context.Context, gameID, playerID string) (*model.Player, error)
}

// PlayerServiceImpl implements PlayerService interface
type PlayerServiceImpl struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

// NewPlayerService creates a new PlayerService instance
func NewPlayerService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) PlayerService {
	return &PlayerServiceImpl{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
	}
}

// UpdatePlayerResources updates a player's resources
func (s *PlayerServiceImpl) UpdatePlayerResources(ctx context.Context, gameID, playerID string, newResources model.Resources) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player
	player, err := s.playerRepo.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Create a copy of the player to avoid modifying the stored one
	updatedPlayer := *player
	updatedPlayer.Resources = newResources

	// Update through PlayerRepository (this will publish events)
	if err := s.playerRepo.UpdatePlayer(ctx, gameID, &updatedPlayer); err != nil {
		log.Error("Failed to update player resources", zap.Error(err))
		return fmt.Errorf("failed to update player: %w", err)
	}

	// Also need to update the game state to keep the main Game entity in sync
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for player update", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Find and update player in game
	for i, p := range game.Players {
		if p.ID == playerID {
			game.Players[i] = updatedPlayer
			break
		}
	}

	// Update game state
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game after player resource change", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Player resources updated successfully")
	return nil
}

// UpdatePlayerProduction updates a player's production
func (s *PlayerServiceImpl) UpdatePlayerProduction(ctx context.Context, gameID, playerID string, newProduction model.Production) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player
	player, err := s.playerRepo.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Create a copy of the player to avoid modifying the stored one
	updatedPlayer := *player
	updatedPlayer.Production = newProduction

	// Update through PlayerRepository (this will publish events)
	if err := s.playerRepo.UpdatePlayer(ctx, gameID, &updatedPlayer); err != nil {
		log.Error("Failed to update player production", zap.Error(err))
		return fmt.Errorf("failed to update player: %w", err)
	}

	// Also need to update the game state to keep the main Game entity in sync
	game, err := s.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for player update", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Find and update player in game
	for i, p := range game.Players {
		if p.ID == playerID {
			game.Players[i] = updatedPlayer
			break
		}
	}

	// Update game state
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game after player production change", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Player production updated successfully")
	return nil
}

// GetPlayer retrieves player information
func (s *PlayerServiceImpl) GetPlayer(ctx context.Context, gameID, playerID string) (*model.Player, error) {
	return s.playerRepo.GetPlayer(ctx, gameID, playerID)
}