package repository

import (
	"context"
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"

	"go.uber.org/zap"
)

// PlayerRepository manages player entities including resources and production
type PlayerRepository interface {
	// Add player to game
	AddPlayer(ctx context.Context, gameID string, player model.Player) error

	// Get player by game and player ID
	GetPlayer(ctx context.Context, gameID, playerID string) (*model.Player, error)

	// Update player
	UpdatePlayer(ctx context.Context, gameID string, player *model.Player) error

	// Get all players in a game
	ListPlayers(ctx context.Context, gameID string) ([]model.Player, error)

	// Remove player from game
	RemovePlayer(ctx context.Context, gameID, playerID string) error
}

// PlayerRepositoryImpl implements PlayerRepository interface
type PlayerRepositoryImpl struct {
	// Map of gameID -> map of playerID -> Player
	players  map[string]map[string]*model.Player
	mutex    sync.RWMutex
	eventBus events.EventBus
}

// NewPlayerRepository creates a new player repository
func NewPlayerRepository(eventBus events.EventBus) PlayerRepository {
	return &PlayerRepositoryImpl{
		players:  make(map[string]map[string]*model.Player),
		eventBus: eventBus,
	}
}

// AddPlayer adds a player to a game
func (r *PlayerRepositoryImpl) AddPlayer(ctx context.Context, gameID string, player model.Player) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, player.ID)

	if gameID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	if player.ID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	// Initialize game players map if it doesn't exist
	if r.players[gameID] == nil {
		r.players[gameID] = make(map[string]*model.Player)
	}

	// Check if player already exists
	if _, exists := r.players[gameID][player.ID]; exists {
		log.Error("Player already exists in game")
		return fmt.Errorf("player with ID %s already exists in game %s", player.ID, gameID)
	}

	// Add player
	r.players[gameID][player.ID] = &player

	log.Debug("Player added to game",
		zap.String("player_name", player.Name),
	)

	// Publish player added event
	if r.eventBus != nil {
		playerAddedEvent := events.NewPlayerJoinedEvent(gameID, player.ID, player.Name)
		if err := r.eventBus.Publish(ctx, playerAddedEvent); err != nil {
			log.Warn("Failed to publish player added event", zap.Error(err))
		}
	}

	return nil
}

// GetPlayer retrieves a player by game and player ID
func (r *PlayerRepositoryImpl) GetPlayer(ctx context.Context, gameID, playerID string) (*model.Player, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	if playerID == "" {
		return nil, fmt.Errorf("player ID cannot be empty")
	}

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return nil, fmt.Errorf("no players found for game %s", gameID)
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return nil, fmt.Errorf("player with ID %s not found in game %s", playerID, gameID)
	}

	return player.DeepCopy(), nil
}

// UpdatePlayer updates a player
func (r *PlayerRepositoryImpl) UpdatePlayer(ctx context.Context, gameID string, player *model.Player) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, player.ID)

	if gameID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	if player == nil {
		return fmt.Errorf("player cannot be nil")
	}

	if player.ID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return fmt.Errorf("no players found for game %s", gameID)
	}

	oldPlayer, exists := gamePlayers[player.ID]
	if !exists {
		return fmt.Errorf("player with ID %s not found in game %s", player.ID, gameID)
	}

	// Capture old state for events (make copies to avoid pointer issues)
	oldResources := oldPlayer.Resources
	oldProduction := oldPlayer.Production
	oldTR := oldPlayer.TerraformRating

	// Update player (store a copy to avoid pointer issues)
	playerCopy := *player
	r.players[gameID][player.ID] = &playerCopy

	log.Info("Player updated")

	// Publish events if resources, production, or TR changed
	if r.eventBus != nil {
		if oldResources != player.Resources {
			resourcesChangedEvent := events.NewPlayerResourcesChangedEvent(gameID, player.ID, oldResources, player.Resources)
			if err := r.eventBus.Publish(ctx, resourcesChangedEvent); err != nil {
				log.Warn("Failed to publish player resources changed event", zap.Error(err))
			}
		}

		if oldProduction != player.Production {
			productionChangedEvent := events.NewPlayerProductionChangedEvent(gameID, player.ID, oldProduction, player.Production)
			if err := r.eventBus.Publish(ctx, productionChangedEvent); err != nil {
				log.Warn("Failed to publish player production changed event", zap.Error(err))
			}
		}

		if oldTR != player.TerraformRating {
			trChangedEvent := events.NewPlayerTRChangedEvent(gameID, player.ID, oldTR, player.TerraformRating)
			if err := r.eventBus.Publish(ctx, trChangedEvent); err != nil {
				log.Warn("Failed to publish player TR changed event", zap.Error(err))
			}
		}
	}

	return nil
}

// ListPlayers returns all players in a game
func (r *PlayerRepositoryImpl) ListPlayers(ctx context.Context, gameID string) ([]model.Player, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return make([]model.Player, 0), nil
	}

	players := make([]model.Player, 0, len(gamePlayers))
	for _, player := range gamePlayers {
		players = append(players, *player.DeepCopy())
	}

	return players, nil
}

// RemovePlayer removes a player from a game
func (r *PlayerRepositoryImpl) RemovePlayer(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	if gameID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	if playerID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return fmt.Errorf("no players found for game %s", gameID)
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return fmt.Errorf("player with ID %s not found in game %s", playerID, gameID)
	}

	delete(gamePlayers, playerID)

	// Clean up empty game
	if len(gamePlayers) == 0 {
		delete(r.players, gameID)
	}

	log.Info("Player removed from game",
		zap.String("player_name", player.Name),
	)

	// Publish player removed event
	if r.eventBus != nil {
		playerRemovedEvent := events.NewPlayerLeftEvent(gameID, playerID, player.Name)
		if err := r.eventBus.Publish(ctx, playerRemovedEvent); err != nil {
			log.Warn("Failed to publish player removed event", zap.Error(err))
		}
	}

	return nil
}
