package service

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// PlayerService handles player-specific business logic
type PlayerService struct {
	gameRepo        *repository.GameRepository
	eventBus        events.EventBus
	eventRepository *events.EventRepository
}

// NewPlayerService creates a new player service
func NewPlayerService(gameRepo *repository.GameRepository, eventBus events.EventBus, eventRepository *events.EventRepository) *PlayerService {
	return &PlayerService{
		gameRepo:        gameRepo,
		eventBus:        eventBus,
		eventRepository: eventRepository,
	}
}

// PayResourceCost deducts the cost from player's resources
func (s *PlayerService) PayResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	log := logger.WithGameContext(gameID, playerID)
	
	// Get the game and player
	game, player, err := s.getGameAndPlayer(gameID, playerID)
	if err != nil {
		return err
	}

	// Capture before state for event
	beforeResources := player.Resources

	// Validate resources are sufficient
	if err := s.validateResourceCost(player, cost); err != nil {
		log.Error("Insufficient resources for cost", zap.Error(err))
		return err
	}

	// Apply the cost
	player.Resources.Credits -= cost.Credits
	player.Resources.Steel -= cost.Steel
	player.Resources.Titanium -= cost.Titanium
	player.Resources.Plants -= cost.Plants
	player.Resources.Energy -= cost.Energy
	player.Resources.Heat -= cost.Heat

	// Update game in repository
	if err := s.gameRepo.UpdateGame(game); err != nil {
		log.Error("Failed to update game after paying resource cost", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	// Publish resource change event
	if s.eventRepository != nil {
		event := events.NewPlayerResourcesChangedEvent(gameID, playerID, beforeResources, player.Resources)
		if err := s.eventRepository.Publish(ctx, event); err != nil {
			log.Warn("Failed to publish player resources changed event", zap.Error(err))
		}
	}

	log.Info("Resource cost paid successfully", 
		zap.Int("credits", cost.Credits),
		zap.Int("steel", cost.Steel),
		zap.Int("titanium", cost.Titanium),
		zap.Int("plants", cost.Plants),
		zap.Int("energy", cost.Energy),
		zap.Int("heat", cost.Heat),
	)

	return nil
}

// AddResources adds resources to a player
func (s *PlayerService) AddResources(ctx context.Context, gameID, playerID string, resources model.ResourceSet) error {
	log := logger.WithGameContext(gameID, playerID)
	
	// Get the game and player
	game, player, err := s.getGameAndPlayer(gameID, playerID)
	if err != nil {
		return err
	}

	// Capture before state for event
	beforeResources := player.Resources

	// Apply the resources
	player.Resources.Credits += resources.Credits
	player.Resources.Steel += resources.Steel
	player.Resources.Titanium += resources.Titanium
	player.Resources.Plants += resources.Plants
	player.Resources.Energy += resources.Energy
	player.Resources.Heat += resources.Heat

	// Update game in repository
	if err := s.gameRepo.UpdateGame(game); err != nil {
		log.Error("Failed to update game after adding resources", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	// Publish resource change event
	if s.eventRepository != nil {
		event := events.NewPlayerResourcesChangedEvent(gameID, playerID, beforeResources, player.Resources)
		if err := s.eventRepository.Publish(ctx, event); err != nil {
			log.Warn("Failed to publish player resources changed event", zap.Error(err))
		}
	}

	log.Info("Resources added successfully", 
		zap.Int("credits", resources.Credits),
		zap.Int("steel", resources.Steel),
		zap.Int("titanium", resources.Titanium),
		zap.Int("plants", resources.Plants),
		zap.Int("energy", resources.Energy),
		zap.Int("heat", resources.Heat),
	)

	return nil
}

// AddProduction increases a player's production
func (s *PlayerService) AddProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error {
	log := logger.WithGameContext(gameID, playerID)
	
	// Get the game and player
	game, player, err := s.getGameAndPlayer(gameID, playerID)
	if err != nil {
		return err
	}

	// Capture before state for event
	beforeProduction := player.Production

	// Apply the production
	player.Production.Credits += production.Credits
	player.Production.Steel += production.Steel
	player.Production.Titanium += production.Titanium
	player.Production.Plants += production.Plants
	player.Production.Energy += production.Energy
	player.Production.Heat += production.Heat

	// Update game in repository
	if err := s.gameRepo.UpdateGame(game); err != nil {
		log.Error("Failed to update game after adding production", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	// Publish production change event
	if s.eventRepository != nil {
		event := events.NewPlayerProductionChangedEvent(gameID, playerID, beforeProduction, player.Production)
		if err := s.eventRepository.Publish(ctx, event); err != nil {
			log.Warn("Failed to publish player production changed event", zap.Error(err))
		}
	}

	log.Info("Production added successfully", 
		zap.Int("credits", production.Credits),
		zap.Int("steel", production.Steel),
		zap.Int("titanium", production.Titanium),
		zap.Int("plants", production.Plants),
		zap.Int("energy", production.Energy),
		zap.Int("heat", production.Heat),
	)

	return nil
}

// RemoveProduction decreases a player's production
func (s *PlayerService) RemoveProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error {
	log := logger.WithGameContext(gameID, playerID)
	
	// Get the game and player
	game, player, err := s.getGameAndPlayer(gameID, playerID)
	if err != nil {
		return err
	}

	// Capture before state for event
	beforeProduction := player.Production

	// Validate production can be removed (don't allow negative)
	if err := s.validateProductionRemoval(player, production); err != nil {
		log.Error("Cannot remove production", zap.Error(err))
		return err
	}

	// Apply the production removal
	player.Production.Credits -= production.Credits
	player.Production.Steel -= production.Steel
	player.Production.Titanium -= production.Titanium
	player.Production.Plants -= production.Plants
	player.Production.Energy -= production.Energy
	player.Production.Heat -= production.Heat

	// Update game in repository
	if err := s.gameRepo.UpdateGame(game); err != nil {
		log.Error("Failed to update game after removing production", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	// Publish production change event
	if s.eventRepository != nil {
		event := events.NewPlayerProductionChangedEvent(gameID, playerID, beforeProduction, player.Production)
		if err := s.eventRepository.Publish(ctx, event); err != nil {
			log.Warn("Failed to publish player production changed event", zap.Error(err))
		}
	}

	log.Info("Production removed successfully", 
		zap.Int("credits", production.Credits),
		zap.Int("steel", production.Steel),
		zap.Int("titanium", production.Titanium),
		zap.Int("plants", production.Plants),
		zap.Int("energy", production.Energy),
		zap.Int("heat", production.Heat),
	)

	return nil
}

// ValidateResourceCost checks if a player has enough resources to pay a cost without actually paying it
func (s *PlayerService) ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	// Get the game and player
	_, player, err := s.getGameAndPlayer(gameID, playerID)
	if err != nil {
		return err
	}

	return s.validateResourceCost(player, cost)
}

// ValidateProductionRequirement checks if a player has minimum production levels
func (s *PlayerService) ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement model.ResourceSet) error {
	// Get the game and player
	_, player, err := s.getGameAndPlayer(gameID, playerID)
	if err != nil {
		return err
	}

	if player.Production.Credits < requirement.Credits {
		return fmt.Errorf("insufficient credit production: need %d, have %d", requirement.Credits, player.Production.Credits)
	}
	if player.Production.Steel < requirement.Steel {
		return fmt.Errorf("insufficient steel production: need %d, have %d", requirement.Steel, player.Production.Steel)
	}
	if player.Production.Titanium < requirement.Titanium {
		return fmt.Errorf("insufficient titanium production: need %d, have %d", requirement.Titanium, player.Production.Titanium)
	}
	if player.Production.Plants < requirement.Plants {
		return fmt.Errorf("insufficient plant production: need %d, have %d", requirement.Plants, player.Production.Plants)
	}
	if player.Production.Energy < requirement.Energy {
		return fmt.Errorf("insufficient energy production: need %d, have %d", requirement.Energy, player.Production.Energy)
	}
	if player.Production.Heat < requirement.Heat {
		return fmt.Errorf("insufficient heat production: need %d, have %d", requirement.Heat, player.Production.Heat)
	}
	return nil
}

// Helper functions

// getGameAndPlayer retrieves a game and player, returning an error if either is not found
func (s *PlayerService) getGameAndPlayer(gameID, playerID string) (*model.Game, *model.Player, error) {
	if gameID == "" {
		return nil, nil, fmt.Errorf("game ID cannot be empty")
	}
	if playerID == "" {
		return nil, nil, fmt.Errorf("player ID cannot be empty")
	}

	game, err := s.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get game: %w", err)
	}

	player, found := game.GetPlayer(playerID)
	if !found {
		return nil, nil, fmt.Errorf("player not found in game")
	}

	return game, player, nil
}

// validateResourceCost checks if a player has enough resources to pay a cost
func (s *PlayerService) validateResourceCost(player *model.Player, cost model.ResourceSet) error {
	if player.Resources.Credits < cost.Credits {
		return fmt.Errorf("insufficient credits: need %d, have %d", cost.Credits, player.Resources.Credits)
	}
	if player.Resources.Steel < cost.Steel {
		return fmt.Errorf("insufficient steel: need %d, have %d", cost.Steel, player.Resources.Steel)
	}
	if player.Resources.Titanium < cost.Titanium {
		return fmt.Errorf("insufficient titanium: need %d, have %d", cost.Titanium, player.Resources.Titanium)
	}
	if player.Resources.Plants < cost.Plants {
		return fmt.Errorf("insufficient plants: need %d, have %d", cost.Plants, player.Resources.Plants)
	}
	if player.Resources.Energy < cost.Energy {
		return fmt.Errorf("insufficient energy: need %d, have %d", cost.Energy, player.Resources.Energy)
	}
	if player.Resources.Heat < cost.Heat {
		return fmt.Errorf("insufficient heat: need %d, have %d", cost.Heat, player.Resources.Heat)
	}
	return nil
}

// validateProductionRemoval checks if production can be safely removed without going negative
func (s *PlayerService) validateProductionRemoval(player *model.Player, production model.ResourceSet) error {
	if player.Production.Credits < production.Credits {
		return fmt.Errorf("cannot remove more credit production than available: trying to remove %d, have %d", production.Credits, player.Production.Credits)
	}
	if player.Production.Steel < production.Steel {
		return fmt.Errorf("cannot remove more steel production than available: trying to remove %d, have %d", production.Steel, player.Production.Steel)
	}
	if player.Production.Titanium < production.Titanium {
		return fmt.Errorf("cannot remove more titanium production than available: trying to remove %d, have %d", production.Titanium, player.Production.Titanium)
	}
	if player.Production.Plants < production.Plants {
		return fmt.Errorf("cannot remove more plant production than available: trying to remove %d, have %d", production.Plants, player.Production.Plants)
	}
	if player.Production.Energy < production.Energy {
		return fmt.Errorf("cannot remove more energy production than available: trying to remove %d, have %d", production.Energy, player.Production.Energy)
	}
	if player.Production.Heat < production.Heat {
		return fmt.Errorf("cannot remove more heat production than available: trying to remove %d, have %d", production.Heat, player.Production.Heat)
	}
	return nil
}