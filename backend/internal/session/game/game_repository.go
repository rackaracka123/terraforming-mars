package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// Repository manages game data with event-driven updates
type Repository interface {
	// Create creates a new game
	Create(ctx context.Context, game *Game) error

	// GetByID retrieves a game by ID
	GetByID(ctx context.Context, gameID string) (*Game, error)

	// List retrieves all games, optionally filtered by status
	List(ctx context.Context, status string) ([]*Game, error)

	// AddPlayer adds a player to a game (event-driven)
	AddPlayer(ctx context.Context, gameID string, playerID string) error

	// UpdateStatus updates game status (event-driven)
	UpdateStatus(ctx context.Context, gameID string, status GameStatus) error

	// UpdatePhase updates game phase (event-driven)
	UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error

	// SetHostPlayer sets the host player for a game
	SetHostPlayer(ctx context.Context, gameID string, playerID string) error

	// SetCurrentTurn sets the current turn player
	SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error

	// UpdateTemperature updates the game temperature
	UpdateTemperature(ctx context.Context, gameID string, temperature int) error

	// UpdateOxygen updates the game oxygen level
	UpdateOxygen(ctx context.Context, gameID string, oxygen int) error

	// UpdateOceans updates the game ocean count
	UpdateOceans(ctx context.Context, gameID string, oceans int) error

	// UpdateGeneration updates the game generation counter
	UpdateGeneration(ctx context.Context, gameID string, generation int) error
}

// RepositoryImpl implements the Repository interface with in-memory storage
type RepositoryImpl struct {
	mu       sync.RWMutex
	games    map[string]*Game // gameID -> Game
	eventBus *events.EventBusImpl
}

// NewRepository creates a new game repository
func NewRepository(eventBus *events.EventBusImpl) Repository {
	return &RepositoryImpl{
		games:    make(map[string]*Game),
		eventBus: eventBus,
	}
}

// Create creates a new game
func (r *RepositoryImpl) Create(ctx context.Context, game *Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.games[game.ID]; exists {
		return fmt.Errorf("game %s already exists", game.ID)
	}

	r.games[game.ID] = game

	// Event publishing can be added here if needed
	// For now, game creation doesn't need an event

	return nil
}

// GetByID retrieves a game by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID string) (*Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	game, exists := r.games[gameID]
	if !exists {
		return nil, &types.NotFoundError{Resource: "game", ID: gameID}
	}

	return game, nil
}

// List retrieves all games, optionally filtered by status
func (r *RepositoryImpl) List(ctx context.Context, status string) ([]*Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	games := make([]*Game, 0, len(r.games))
	for _, game := range r.games {
		// If status filter is empty or matches, include the game
		if status == "" || string(game.Status) == status {
			games = append(games, game)
		}
	}

	return games, nil
}

// AddPlayer adds a player to a game (event-driven)
func (r *RepositoryImpl) AddPlayer(ctx context.Context, gameID string, playerID string) error {
	r.mu.Lock()

	game, exists := r.games[gameID]
	if !exists {
		r.mu.Unlock()
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	// Check if player already in game
	for _, pid := range game.PlayerIDs {
		if pid == playerID {
			r.mu.Unlock()
			return fmt.Errorf("player %s already in game %s", playerID, gameID)
		}
	}

	game.PlayerIDs = append(game.PlayerIDs, playerID)

	// Release lock BEFORE publishing event to avoid deadlock
	// Event subscribers may need to read from this repository
	r.mu.Unlock()

	// Publish PlayerJoinedEvent for event-driven broadcasting
	// This is safe to do after unlocking because the event only contains IDs (not object references)
	log := logger.Get().With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID))
	log.Info("📡 Publishing PlayerJoinedEvent")

	events.Publish(r.eventBus, events.PlayerJoinedEvent{
		GameID:     gameID,
		PlayerID:   playerID,
		PlayerName: "", // Name not available at repository level, subscribers can look up if needed
		Timestamp:  time.Now(),
	})

	log.Info("✅ PlayerJoinedEvent published")

	return nil
}

// UpdateStatus updates game status (event-driven)
func (r *RepositoryImpl) UpdateStatus(ctx context.Context, gameID string, status GameStatus) error {
	r.mu.Lock()

	game, exists := r.games[gameID]
	if !exists {
		r.mu.Unlock()
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldStatus := game.Status
	game.Status = status

	// Release lock before publishing event to avoid deadlock
	r.mu.Unlock()

	// Publish GameStatusChangedEvent
	if oldStatus != status {
		log := logger.Get().With(
			zap.String("game_id", gameID),
			zap.String("old_status", string(oldStatus)),
			zap.String("new_status", string(status)))
		log.Debug("📡 Publishing GameStatusChangedEvent")

		events.Publish(r.eventBus, events.GameStatusChangedEvent{
			GameID:    gameID,
			OldStatus: string(oldStatus),
			NewStatus: string(status),
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdatePhase updates game phase (event-driven)
func (r *RepositoryImpl) UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error {
	r.mu.Lock()

	game, exists := r.games[gameID]
	if !exists {
		r.mu.Unlock()
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldPhase := game.CurrentPhase
	game.CurrentPhase = phase

	// Release lock before publishing event to avoid deadlock
	r.mu.Unlock()

	// Publish GamePhaseChangedEvent
	if oldPhase != phase {
		log := logger.Get().With(
			zap.String("game_id", gameID),
			zap.String("old_phase", string(oldPhase)),
			zap.String("new_phase", string(phase)))
		log.Debug("📡 Publishing GamePhaseChangedEvent")

		events.Publish(r.eventBus, events.GamePhaseChangedEvent{
			GameID:    gameID,
			OldPhase:  string(oldPhase),
			NewPhase:  string(phase),
			Timestamp: time.Now(),
		})
	}

	return nil
}

// SetHostPlayer sets the host player for a game
func (r *RepositoryImpl) SetHostPlayer(ctx context.Context, gameID string, playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	game, exists := r.games[gameID]
	if !exists {
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	game.HostPlayerID = playerID

	return nil
}

// SetCurrentTurn sets the current turn player
func (r *RepositoryImpl) SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	game, exists := r.games[gameID]
	if !exists {
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	game.CurrentTurn = playerID

	return nil
}

// UpdateTemperature updates the game temperature
func (r *RepositoryImpl) UpdateTemperature(ctx context.Context, gameID string, temperature int) error {
	r.mu.Lock()

	game, exists := r.games[gameID]
	if !exists {
		r.mu.Unlock()
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldTemp := game.GlobalParameters.Temperature
	game.GlobalParameters.Temperature = temperature

	// Release lock before publishing event to avoid deadlock
	r.mu.Unlock()

	// Publish TemperatureChangedEvent
	if oldTemp != temperature {
		log := logger.Get().With(
			zap.String("game_id", gameID),
			zap.Int("old_temperature", oldTemp),
			zap.Int("new_temperature", temperature))
		log.Debug("📡 Publishing TemperatureChangedEvent")

		events.Publish(r.eventBus, events.TemperatureChangedEvent{
			GameID:    gameID,
			OldValue:  oldTemp,
			NewValue:  temperature,
			ChangedBy: "", // Can be enhanced to track player who triggered this
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateOxygen updates the game oxygen level
func (r *RepositoryImpl) UpdateOxygen(ctx context.Context, gameID string, oxygen int) error {
	r.mu.Lock()

	game, exists := r.games[gameID]
	if !exists {
		r.mu.Unlock()
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldOxygen := game.GlobalParameters.Oxygen
	game.GlobalParameters.Oxygen = oxygen

	// Release lock before publishing event to avoid deadlock
	r.mu.Unlock()

	// Publish OxygenChangedEvent
	if oldOxygen != oxygen {
		log := logger.Get().With(
			zap.String("game_id", gameID),
			zap.Int("old_oxygen", oldOxygen),
			zap.Int("new_oxygen", oxygen))
		log.Debug("📡 Publishing OxygenChangedEvent")

		events.Publish(r.eventBus, events.OxygenChangedEvent{
			GameID:    gameID,
			OldValue:  oldOxygen,
			NewValue:  oxygen,
			ChangedBy: "", // Can be enhanced to track player who triggered this
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateOceans updates the game ocean count
func (r *RepositoryImpl) UpdateOceans(ctx context.Context, gameID string, oceans int) error {
	r.mu.Lock()

	game, exists := r.games[gameID]
	if !exists {
		r.mu.Unlock()
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldOceans := game.GlobalParameters.Oceans
	game.GlobalParameters.Oceans = oceans

	// Release lock before publishing event to avoid deadlock
	r.mu.Unlock()

	// Publish OceansChangedEvent
	if oldOceans != oceans {
		log := logger.Get().With(
			zap.String("game_id", gameID),
			zap.Int("old_oceans", oldOceans),
			zap.Int("new_oceans", oceans))
		log.Debug("📡 Publishing OceansChangedEvent")

		events.Publish(r.eventBus, events.OceansChangedEvent{
			GameID:    gameID,
			OldValue:  oldOceans,
			NewValue:  oceans,
			ChangedBy: "", // Can be enhanced to track player who triggered this
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateGeneration updates the game generation counter
func (r *RepositoryImpl) UpdateGeneration(ctx context.Context, gameID string, generation int) error {
	r.mu.Lock()

	game, exists := r.games[gameID]
	if !exists {
		r.mu.Unlock()
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldGeneration := game.Generation
	game.Generation = generation

	// Release lock before publishing event to avoid deadlock
	r.mu.Unlock()

	// Publish GenerationAdvancedEvent
	if oldGeneration != generation {
		log := logger.Get().With(
			zap.String("game_id", gameID),
			zap.Int("old_generation", oldGeneration),
			zap.Int("new_generation", generation))
		log.Debug("📡 Publishing GenerationAdvancedEvent")

		events.Publish(r.eventBus, events.GenerationAdvancedEvent{
			GameID:        gameID,
			OldGeneration: oldGeneration,
			NewGeneration: generation,
			Timestamp:     time.Now(),
		})
	}

	return nil
}
