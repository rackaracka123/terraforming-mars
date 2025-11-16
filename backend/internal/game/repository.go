package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameRepository provides clean CRUD operations and granular updates for games
type Repository interface {
	// Basic CRUD operations
	Create(ctx context.Context, settings GameSettings) (Game, error)
	GetByID(ctx context.Context, gameID string) (Game, error)
	Delete(ctx context.Context, gameID string) error
	List(ctx context.Context, status string) ([]Game, error)

	// Granular update methods for specific fields
	UpdateStatus(ctx context.Context, gameID string, status GameStatus) error
	UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error
	SetCurrentPlayer(ctx context.Context, gameID string, playerID string) error
	AddPlayerID(ctx context.Context, gameID string, playerID string) error
	RemovePlayerID(ctx context.Context, gameID string, playerID string) error
	SetHostPlayer(ctx context.Context, gameID string, playerID string) error
	UpdateGeneration(ctx context.Context, gameID string, generation int) error
}

// GameRepositoryImpl implements GameRepository with in-memory storage
type RepositoryImpl struct {
	games    map[string]*Game
	mutex    sync.RWMutex
	eventBus *events.EventBusImpl
}

// NewRepository creates a new game repository
func NewRepository(eventBus *events.EventBusImpl) Repository {
	return &RepositoryImpl{
		games:    make(map[string]*Game),
		eventBus: eventBus,
	}
}

// Create creates a new game with the given settings
func (r *RepositoryImpl) Create(ctx context.Context, settings GameSettings) (Game, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.Get()
	log.Debug("Creating new game")

	// Generate unique game ID
	gameID := uuid.New().String()

	// Create the game metadata
	game := NewGame(gameID, settings)

	// Store game in repository
	r.games[gameID] = game

	log.Debug("Game created", zap.String("game_id", gameID))

	// Return a copy to prevent external mutation
	gameCopy := *game
	gameCopy.PlayerIDs = make([]string, len(game.PlayerIDs))
	copy(gameCopy.PlayerIDs, game.PlayerIDs)
	return gameCopy, nil
}

// GetByID retrieves a game by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID string) (Game, error) {
	if gameID == "" {
		return Game{}, fmt.Errorf("game ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	game, exists := r.games[gameID]
	if !exists {
		return Game{}, &NotFoundError{Resource: "game", ID: gameID}
	}

	// Return a copy to prevent external mutation
	gameCopy := *game
	gameCopy.PlayerIDs = make([]string, len(game.PlayerIDs))
	copy(gameCopy.PlayerIDs, game.PlayerIDs)
	return gameCopy, nil
}

// Delete removes a game from the repository
func (r *RepositoryImpl) Delete(ctx context.Context, gameID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	if gameID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	_, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	delete(r.games, gameID)

	log.Info("Game deleted")

	return nil
}

// List returns all games, optionally filtered by status
func (r *RepositoryImpl) List(ctx context.Context, status string) ([]Game, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	games := make([]Game, 0)

	for _, game := range r.games {
		if status == "" || string(game.Status) == status {
			// Return a copy to prevent external mutation
			gameCopy := *game
			gameCopy.PlayerIDs = make([]string, len(game.PlayerIDs))
			copy(gameCopy.PlayerIDs, game.PlayerIDs)
			games = append(games, gameCopy)
		}
	}

	return games, nil
}

// UpdateStatus updates a game's status
func (r *RepositoryImpl) UpdateStatus(ctx context.Context, gameID string, status GameStatus) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	oldStatus := game.Status
	game.Status = status
	game.UpdatedAt = time.Now()

	log.Info("Game status updated", zap.String("old_status", string(oldStatus)), zap.String("new_status", string(status)))

	// Publish GameStatusChangedEvent
	events.Publish(r.eventBus, events.GameStatusChangedEvent{
		GameID:    gameID,
		OldStatus: string(oldStatus),
		NewStatus: string(status),
	})

	return nil
}

// UpdatePhase updates a game's current phase
func (r *RepositoryImpl) UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error {
	// Store old phase value before locking for event publishing
	var oldPhase GamePhase
	var shouldPublishEvent bool

	func() {
		r.mutex.Lock()
		defer r.mutex.Unlock()

		log := logger.WithGameContext(gameID, "")

		game, exists := r.games[gameID]
		if !exists {
			return
		}

		oldPhase = game.CurrentPhase
		game.CurrentPhase = phase
		game.UpdatedAt = time.Now()

		log.Info("Game phase updated", zap.String("old_phase", string(oldPhase)), zap.String("new_phase", string(phase)))

		shouldPublishEvent = r.eventBus != nil && oldPhase != phase
	}()

	// Publish event AFTER releasing the lock to avoid deadlock
	if shouldPublishEvent {
		events.Publish(r.eventBus, events.GamePhaseChangedEvent{
			GameID:    gameID,
			OldPhase:  string(oldPhase),
			NewPhase:  string(phase),
			Timestamp: time.Now(),
		})
	}

	return nil
}

// SetCurrentPlayer sets the current active player
func (r *RepositoryImpl) SetCurrentPlayer(ctx context.Context, gameID string, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	oldPlayerID := game.CurrentPlayerID
	game.CurrentPlayerID = playerID
	game.UpdatedAt = time.Now()

	log.Info("Current player updated", zap.String("old_player", oldPlayerID), zap.String("new_player", playerID))

	return nil
}

// AddPlayerID adds a player ID to the game
func (r *RepositoryImpl) AddPlayerID(ctx context.Context, gameID string, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Check if player already exists
	for _, existingPlayerID := range game.PlayerIDs {
		if existingPlayerID == playerID {
			return fmt.Errorf("player %s already exists in game %s", playerID, gameID)
		}
	}

	game.PlayerIDs = append(game.PlayerIDs, playerID)
	game.UpdatedAt = time.Now()

	// Set as host if first player
	if len(game.PlayerIDs) == 1 {
		game.HostPlayerID = playerID
		log.Info("Player added and set as host")
	} else {
		log.Info("Player added to game")
	}

	return nil
}

// RemovePlayerID removes a player ID from the game
func (r *RepositoryImpl) RemovePlayerID(ctx context.Context, gameID string, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Find and remove player ID
	for i, existingPlayerID := range game.PlayerIDs {
		if existingPlayerID == playerID {
			game.PlayerIDs = append(game.PlayerIDs[:i], game.PlayerIDs[i+1:]...)
			game.UpdatedAt = time.Now()

			// Clear host if they were the host
			if game.HostPlayerID == playerID {
				if len(game.PlayerIDs) > 0 {
					game.HostPlayerID = game.PlayerIDs[0] // Set first remaining player as host
					log.Info("Player removed and host transferred", zap.String("new_host", game.HostPlayerID))
				} else {
					game.HostPlayerID = ""
					log.Info("Player removed and no host remaining")
				}
			} else {
				log.Info("Player removed from game")
			}

			return nil
		}
	}

	return fmt.Errorf("player %s not found in game %s", playerID, gameID)
}

// SetHostPlayer sets the host player for the game
func (r *RepositoryImpl) SetHostPlayer(ctx context.Context, gameID string, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Verify player exists in game
	playerExists := false
	for _, existingPlayerID := range game.PlayerIDs {
		if existingPlayerID == playerID {
			playerExists = true
			break
		}
	}

	if !playerExists {
		return fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}

	game.HostPlayerID = playerID
	game.UpdatedAt = time.Now()

	log.Info("Host player updated")

	return nil
}

// UpdateGeneration updates the game generation
func (r *RepositoryImpl) UpdateGeneration(ctx context.Context, gameID string, generation int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	oldGeneration := game.Generation
	game.Generation = generation
	game.UpdatedAt = time.Now()

	log.Info("Generation updated", zap.Int("generation", generation))

	// Publish GenerationAdvancedEvent
	events.Publish(r.eventBus, events.GenerationAdvancedEvent{
		GameID:        gameID,
		OldGeneration: oldGeneration,
		NewGeneration: generation,
	})

	return nil
}

// Clear removes all games from the repository
func (r *RepositoryImpl) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.games = make(map[string]*Game)
}
