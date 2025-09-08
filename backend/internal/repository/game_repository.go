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

// GameRepository manages game metadata and state with integrated global parameters
type GameRepository interface {
	// Create a new game
	Create(ctx context.Context, settings model.GameSettings) (model.Game, error)

	// Get game by ID
	Get(ctx context.Context, gameID string) (model.Game, error)

	// Update game state
	Update(ctx context.Context, game *model.Game) error

	// List games with optional status filter
	List(ctx context.Context, status string) ([]model.Game, error)

	// Delete game
	Delete(ctx context.Context, gameID string) error

	// Global parameters methods (merged from GlobalParametersRepository)
	GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error)
	UpdateGlobalParameters(ctx context.Context, gameID string, params *model.GlobalParameters) error
}

// GameRepositoryImpl implements GameRepository interface
type GameRepositoryImpl struct {
	gameEntities map[string]*GameEntity // Use GameEntity for storage
	mutex        sync.RWMutex
	eventBus     events.EventBus
	playerRepo   PlayerRepository
}

// NewGameRepository creates a new game repository
func NewGameRepository(eventBus events.EventBus, playerRepo PlayerRepository) GameRepository {
	return &GameRepositoryImpl{
		gameEntities: make(map[string]*GameEntity),
		eventBus:     eventBus,
		playerRepo:   playerRepo,
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

	// Create the game entity
	gameEntity := NewGameEntity(gameID, settings)

	// Store in repository
	r.gameEntities[gameID] = gameEntity

	log.Debug("Game created",
		zap.String("game_id", gameID),
	)

	// Publish game created event
	if r.eventBus != nil {
		gameCreatedEvent := events.NewGameCreatedEvent(gameID, settings)
		if err := r.eventBus.Publish(ctx, gameCreatedEvent); err != nil {
			log.Warn("Failed to publish game created event", zap.Error(err))
		}
	}

	// Convert GameEntity to Game with empty players (will be populated by Get method)
	game := gameEntity.ToGame([]model.Player{})
	return *game, nil
}

// Get retrieves a game by ID with fresh player data
func (r *GameRepositoryImpl) Get(ctx context.Context, gameID string) (model.Game, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if gameID == "" {
		return model.Game{}, fmt.Errorf("game ID cannot be empty")
	}

	gameEntity, exists := r.gameEntities[gameID]
	if !exists {
		return model.Game{}, fmt.Errorf("game with ID %s not found", gameID)
	}

	// Fetch fresh player data from PlayerRepository to ensure synchronization
	players := make([]model.Player, 0, len(gameEntity.PlayerIDs))
	for _, playerID := range gameEntity.PlayerIDs {
		freshPlayer, err := r.playerRepo.GetPlayer(ctx, gameID, playerID)
		if err != nil {
			// Log warning but don't fail - use stale data as fallback
			logger.Get().Warn("Failed to fetch fresh player data, using stale data",
				zap.String("game_id", gameID),
				zap.String("player_id", playerID),
				zap.Error(err))
			continue
		}
		players = append(players, freshPlayer)
	}

	// Convert GameEntity to Game with populated players
	return *gameEntity.ToGame(players), nil
}

// Update updates a game in the repository
func (r *GameRepositoryImpl) Update(ctx context.Context, game *model.Game) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(game.ID, "")

	if game == nil {
		return fmt.Errorf("game cannot be nil")
	}

	oldGameEntity, exists := r.gameEntities[game.ID]
	if !exists {
		log.Error("Attempted to update non-existent game")
		return fmt.Errorf("game with ID %s not found", game.ID)
	}

	// Capture old state for event
	oldGame := oldGameEntity.ToGame(game.Players) // Use current players for old state

	// Convert Game to GameEntity (extract player IDs)
	playerIDs := make([]string, len(game.Players))
	for i, player := range game.Players {
		playerIDs[i] = player.ID
	}

	// Create new GameEntity
	newGameEntity := &GameEntity{
		ID:               game.ID,
		CreatedAt:        game.CreatedAt,
		UpdatedAt:        time.Now(),
		Status:           game.Status,
		Settings:         game.Settings,
		PlayerIDs:        playerIDs,
		HostPlayerID:     game.HostPlayerID,
		CurrentPhase:     game.CurrentPhase,
		GlobalParameters: game.GlobalParameters,
		CurrentPlayerID:  game.CurrentPlayerID,
		Generation:       game.Generation,
		RemainingActions: game.RemainingActions,
	}

	// Store updated game entity
	r.gameEntities[game.ID] = newGameEntity

	log.Debug("Game updated")

	// Publish game updated event
	if r.eventBus != nil {
		gameUpdatedEvent := events.NewGameStateChangedEvent(game.ID, oldGame, game)
		if err := r.eventBus.Publish(ctx, gameUpdatedEvent); err != nil {
			log.Warn("Failed to publish game updated event", zap.Error(err))
		}
	}

	return nil
}

// List returns all games, optionally filtered by status
func (r *GameRepositoryImpl) List(ctx context.Context, status string) ([]model.Game, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	games := make([]model.Game, 0)

	for _, gameEntity := range r.gameEntities {
		if status == "" || string(gameEntity.Status) == status {
			// For list operations, we can return games with empty player arrays for performance
			game := gameEntity.ToGame([]model.Player{})
			games = append(games, *game)
		}
	}

	return games, nil
}

// Delete removes a game from the repository
func (r *GameRepositoryImpl) Delete(ctx context.Context, gameID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	if gameID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	_, exists := r.gameEntities[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	delete(r.gameEntities, gameID)

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

// GetGlobalParameters retrieves global parameters for a game
func (r *GameRepositoryImpl) GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if gameID == "" {
		return model.GlobalParameters{}, fmt.Errorf("game ID cannot be empty")
	}

	gameEntity, exists := r.gameEntities[gameID]
	if !exists {
		// Return default parameters if game doesn't exist
		defaultParams := model.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		}
		return defaultParams, nil
	}

	paramsCopy := gameEntity.GlobalParameters
	return paramsCopy, nil
}

// UpdateGlobalParameters updates global parameters for a game
func (r *GameRepositoryImpl) UpdateGlobalParameters(ctx context.Context, gameID string, params *model.GlobalParameters) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	if gameID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	if params == nil {
		return fmt.Errorf("global parameters cannot be nil")
	}

	// Validate parameter ranges
	if err := r.validateGlobalParameters(params); err != nil {
		log.Error("Invalid global parameters", zap.Error(err))
		return fmt.Errorf("invalid parameters: %w", err)
	}

	gameEntity, exists := r.gameEntities[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	// Capture old parameters for events
	oldParams := gameEntity.GlobalParameters

	// Update parameters in game entity
	gameEntity.GlobalParameters = *params
	gameEntity.UpdatedAt = time.Now()

	log.Info("Global parameters updated",
		zap.Int("temperature", params.Temperature),
		zap.Int("oxygen", params.Oxygen),
		zap.Int("oceans", params.Oceans),
	)

	// Publish events for each parameter that changed
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

		// Also publish a general parameters changed event
		parametersChangedEvent := events.NewGlobalParametersChangedEvent(gameID, oldParams, *params)
		if err := r.eventBus.Publish(ctx, parametersChangedEvent); err != nil {
			log.Warn("Failed to publish global parameters changed event", zap.Error(err))
		}
	}

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
