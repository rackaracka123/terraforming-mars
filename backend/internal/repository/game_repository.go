package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameRepository provides clean CRUD operations and granular updates for games
type GameRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, settings model.GameSettings) (model.Game, error)
	GetByID(ctx context.Context, gameID string) (model.Game, error)
	Delete(ctx context.Context, gameID string) error
	List(ctx context.Context, status string) ([]model.Game, error)

	// Granular update methods for specific fields
	UpdateStatus(ctx context.Context, gameID string, status model.GameStatus) error
	UpdatePhase(ctx context.Context, gameID string, phase model.GamePhase) error
	UpdateGlobalParameters(ctx context.Context, gameID string, params model.GlobalParameters) error
	SetCurrentPlayer(ctx context.Context, gameID string, playerID string) error
	SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error
	AddPlayerID(ctx context.Context, gameID string, playerID string) error
	RemovePlayerID(ctx context.Context, gameID string, playerID string) error
	SetHostPlayer(ctx context.Context, gameID string, playerID string) error
	UpdateGeneration(ctx context.Context, gameID string, generation int) error
	UpdateRemainingActions(ctx context.Context, gameID string, actions int) error
}

// GameRepositoryImpl implements GameRepository with in-memory storage
type GameRepositoryImpl struct {
	games    map[string]*model.Game
	mutex    sync.RWMutex
	eventBus events.EventBus
}

// NewGameRepository creates a new game repository
func NewGameRepository(eventBus events.EventBus) GameRepository {
	return &GameRepositoryImpl{
		games:    make(map[string]*model.Game),
		eventBus: eventBus,
	}
}

// Create creates a new game with the given settings
func (r *GameRepositoryImpl) Create(ctx context.Context, settings model.GameSettings) (model.Game, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.Get()
	log.Debug("Creating new game")

	// Generate unique game ID
	gameID := uuid.New().String()

	// Create the game
	game := model.NewGame(gameID, settings)

	// Store in repository
	r.games[gameID] = game

	log.Debug("Game created", zap.String("game_id", gameID))

	// Publish game created event
	if r.eventBus != nil {
		gameCreatedEvent := events.NewGameCreatedEvent(gameID, settings)
		if err := r.eventBus.Publish(ctx, gameCreatedEvent); err != nil {
			log.Warn("Failed to publish game created event", zap.Error(err))
		}
	}

	return *game, nil
}

// GetByID retrieves a game by ID
func (r *GameRepositoryImpl) GetByID(ctx context.Context, gameID string) (model.Game, error) {
	if gameID == "" {
		return model.Game{}, fmt.Errorf("game ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	game, exists := r.games[gameID]
	if !exists {
		return model.Game{}, fmt.Errorf("game with ID %s not found", gameID)
	}

	// Return a copy to prevent external mutation
	gameCopy := *game
	gameCopy.PlayerIDs = make([]string, len(game.PlayerIDs))
	copy(gameCopy.PlayerIDs, game.PlayerIDs)

	return gameCopy, nil
}

// Delete removes a game from the repository
func (r *GameRepositoryImpl) Delete(ctx context.Context, gameID string) error {
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

	// Publish game deleted event
	if r.eventBus != nil {
		gameDeletedEvent := events.NewGameDeletedEvent(gameID)
		if err := r.eventBus.Publish(ctx, gameDeletedEvent); err != nil {
			log.Warn("Failed to publish game deleted event", zap.Error(err))
		}
	}

	return nil
}

// List returns all games, optionally filtered by status
func (r *GameRepositoryImpl) List(ctx context.Context, status string) ([]model.Game, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	games := make([]model.Game, 0)

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
func (r *GameRepositoryImpl) UpdateStatus(ctx context.Context, gameID string, status model.GameStatus) error {
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

	// Publish event if status changed
	if r.eventBus != nil && oldStatus != status {
		gameUpdatedEvent := events.NewGameUpdatedEvent(gameID)
		if err := r.eventBus.Publish(ctx, gameUpdatedEvent); err != nil {
			log.Warn("Failed to publish game updated event", zap.Error(err))
		}
	}

	return nil
}

// UpdatePhase updates a game's current phase
func (r *GameRepositoryImpl) UpdatePhase(ctx context.Context, gameID string, phase model.GamePhase) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	oldPhase := game.CurrentPhase
	game.CurrentPhase = phase
	game.UpdatedAt = time.Now()

	log.Info("Game phase updated", zap.String("old_phase", string(oldPhase)), zap.String("new_phase", string(phase)))

	// Publish event if phase changed
	if r.eventBus != nil && oldPhase != phase {
		gameUpdatedEvent := events.NewGameUpdatedEvent(gameID)
		if err := r.eventBus.Publish(ctx, gameUpdatedEvent); err != nil {
			log.Warn("Failed to publish game updated event", zap.Error(err))
		}
	}

	return nil
}

// UpdateGlobalParameters updates global parameters for a game
func (r *GameRepositoryImpl) UpdateGlobalParameters(ctx context.Context, gameID string, params model.GlobalParameters) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	if err := r.validateGlobalParameters(&params); err != nil {
		log.Error("Invalid global parameters", zap.Error(err))
		return fmt.Errorf("invalid parameters: %w", err)
	}

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	oldParams := game.GlobalParameters
	game.GlobalParameters = params
	game.UpdatedAt = time.Now()

	log.Info("Global parameters updated",
		zap.Int("temperature", params.Temperature),
		zap.Int("oxygen", params.Oxygen),
		zap.Int("oceans", params.Oceans),
	)

	// Publish specific events for each parameter that changed
	if r.eventBus != nil {
		if oldParams.Temperature != params.Temperature {
			tempEvent := events.NewTemperatureChangedEvent(gameID, oldParams.Temperature, params.Temperature)
			if err := r.eventBus.Publish(ctx, tempEvent); err != nil {
				log.Warn("Failed to publish temperature changed event", zap.Error(err))
			}
		}

		if oldParams.Oxygen != params.Oxygen {
			oxygenEvent := events.NewOxygenChangedEvent(gameID, oldParams.Oxygen, params.Oxygen)
			if err := r.eventBus.Publish(ctx, oxygenEvent); err != nil {
				log.Warn("Failed to publish oxygen changed event", zap.Error(err))
			}
		}

		if oldParams.Oceans != params.Oceans {
			oceansEvent := events.NewOceansChangedEvent(gameID, oldParams.Oceans, params.Oceans)
			if err := r.eventBus.Publish(ctx, oceansEvent); err != nil {
				log.Warn("Failed to publish oceans changed event", zap.Error(err))
			}
		}

		// General parameters changed event
		parametersChangedEvent := events.NewGlobalParametersChangedEvent(gameID, oldParams, params)
		if err := r.eventBus.Publish(ctx, parametersChangedEvent); err != nil {
			log.Warn("Failed to publish global parameters changed event", zap.Error(err))
		}
	}

	return nil
}

// SetCurrentPlayer sets the current active player
func (r *GameRepositoryImpl) SetCurrentPlayer(ctx context.Context, gameID string, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	oldPlayerID := game.ViewingPlayerID
	game.ViewingPlayerID = playerID
	game.UpdatedAt = time.Now()

	log.Info("Current player updated", zap.String("old_player", oldPlayerID), zap.String("new_player", playerID))

	return nil
}

// SetCurrentTurn sets the current turn (whose turn it is to play)
func (r *GameRepositoryImpl) SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	var oldTurnPlayer string
	if game.CurrentTurn != nil {
		oldTurnPlayer = *game.CurrentTurn
	} else {
		oldTurnPlayer = "none"
	}

	game.CurrentTurn = playerID
	game.UpdatedAt = time.Now()

	var newTurnPlayer string
	if playerID != nil {
		newTurnPlayer = *playerID
		log.Info("Current turn updated", zap.String("old_turn", oldTurnPlayer), zap.String("new_turn", newTurnPlayer))
	} else {
		newTurnPlayer = "none"
		log.Info("Current turn cleared", zap.String("old_turn", oldTurnPlayer))
	}

	// Publish game updated event if turn changed
	if r.eventBus != nil {
		gameUpdatedEvent := events.NewGameUpdatedEvent(gameID)
		if err := r.eventBus.Publish(ctx, gameUpdatedEvent); err != nil {
			log.Warn("Failed to publish game updated event", zap.Error(err))
		}
	}

	return nil
}

// AddPlayerID adds a player ID to the game
func (r *GameRepositoryImpl) AddPlayerID(ctx context.Context, gameID string, playerID string) error {
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
func (r *GameRepositoryImpl) RemovePlayerID(ctx context.Context, gameID string, playerID string) error {
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
func (r *GameRepositoryImpl) SetHostPlayer(ctx context.Context, gameID string, playerID string) error {
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
func (r *GameRepositoryImpl) UpdateGeneration(ctx context.Context, gameID string, generation int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	game.Generation = generation
	game.UpdatedAt = time.Now()

	log.Info("Generation updated", zap.Int("generation", generation))

	return nil
}

// UpdateRemainingActions updates the remaining actions count
func (r *GameRepositoryImpl) UpdateRemainingActions(ctx context.Context, gameID string, actions int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	game, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	game.RemainingActions = actions
	game.UpdatedAt = time.Now()

	log.Debug("Remaining actions updated", zap.Int("actions", actions))

	return nil
}

// validateGlobalParameters ensures parameters are within valid game ranges
func (r *GameRepositoryImpl) validateGlobalParameters(params *model.GlobalParameters) error {
	if params.Temperature < -30 || params.Temperature > 8 {
		return fmt.Errorf("temperature must be between -30 and 8, got %d", params.Temperature)
	}

	if params.Oxygen < 0 || params.Oxygen > 14 {
		return fmt.Errorf("oxygen must be between 0 and 14, got %d", params.Oxygen)
	}

	if params.Oceans < 0 || params.Oceans > 9 {
		return fmt.Errorf("oceans must be between 0 and 9, got %d", params.Oceans)
	}

	return nil
}
