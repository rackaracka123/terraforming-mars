package game

import (
	"context"
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// RepositoryImpl implements the core.Repository interface with in-memory storage
type RepositoryImpl struct {
	mu       sync.RWMutex
	games    map[string]*Game // gameID -> Game
	eventBus *events.EventBusImpl
}

// NewRepository creates a new game repository
func NewRepository(eventBus *events.EventBusImpl) *RepositoryImpl {
	return &RepositoryImpl{
		games:    make(map[string]*Game),
		eventBus: eventBus,
	}
}

// Create creates a new game
func (r *RepositoryImpl) Create(ctx context.Context, g *Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.games[g.ID]; exists {
		return fmt.Errorf("game %s already exists", g.ID)
	}

	r.games[g.ID] = g
	return nil
}

// GetByID retrieves a game by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID string) (*Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	g, exists := r.games[gameID]
	if !exists {
		return nil, &types.NotFoundError{Resource: "game", ID: gameID}
	}

	return g, nil
}

// List retrieves all games, optionally filtered by status
func (r *RepositoryImpl) List(ctx context.Context, status types.GameStatus) ([]*Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var games []*Game
	for _, g := range r.games {
		// If status is empty, include all games
		if status == "" || g.Status == status {
			games = append(games, g)
		}
	}

	return games, nil
}

// UpdateTemperature updates the temperature for a game
func (r *RepositoryImpl) UpdateTemperature(ctx context.Context, gameID string, newTemp int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, exists := r.games[gameID]
	if !exists {
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldTemp := g.GlobalParameters.Temperature
	g.GlobalParameters.Temperature = newTemp

	// Publish event
	events.Publish(r.eventBus, events.TemperatureChangedEvent{
		GameID:   gameID,
		OldValue: oldTemp,
		NewValue: newTemp,
	})

	logger.Get().Debug("Temperature updated",
		zap.String("game_id", gameID),
		zap.Int("old_temp", oldTemp),
		zap.Int("new_temp", newTemp))

	return nil
}

// UpdateOxygen updates the oxygen level for a game
func (r *RepositoryImpl) UpdateOxygen(ctx context.Context, gameID string, newOxygen int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, exists := r.games[gameID]
	if !exists {
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldOxygen := g.GlobalParameters.Oxygen
	g.GlobalParameters.Oxygen = newOxygen

	// Publish event
	events.Publish(r.eventBus, events.OxygenChangedEvent{
		GameID:   gameID,
		OldValue: oldOxygen,
		NewValue: newOxygen,
	})

	logger.Get().Debug("Oxygen updated",
		zap.String("game_id", gameID),
		zap.Int("old_oxygen", oldOxygen),
		zap.Int("new_oxygen", newOxygen))

	return nil
}

// UpdateOceans updates the ocean count for a game
func (r *RepositoryImpl) UpdateOceans(ctx context.Context, gameID string, newOceans int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, exists := r.games[gameID]
	if !exists {
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldOceans := g.GlobalParameters.Oceans
	g.GlobalParameters.Oceans = newOceans

	// Publish event
	events.Publish(r.eventBus, events.OceansChangedEvent{
		GameID:   gameID,
		OldValue: oldOceans,
		NewValue: newOceans,
	})

	logger.Get().Debug("Oceans updated",
		zap.String("game_id", gameID),
		zap.Int("old_oceans", oldOceans),
		zap.Int("new_oceans", newOceans))

	return nil
}

// UpdateGeneration updates the game generation
func (r *RepositoryImpl) UpdateGeneration(ctx context.Context, gameID string, newGeneration int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, exists := r.games[gameID]
	if !exists {
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldGeneration := g.Generation
	g.Generation = newGeneration

	// Publish event
	events.Publish(r.eventBus, events.GenerationAdvancedEvent{
		GameID:        gameID,
		OldGeneration: oldGeneration,
		NewGeneration: newGeneration,
	})

	logger.Get().Debug("Generation updated",
		zap.String("game_id", gameID),
		zap.Int("old_generation", oldGeneration),
		zap.Int("new_generation", newGeneration))

	return nil
}

// UpdatePhase updates the game phase
func (r *RepositoryImpl) UpdatePhase(ctx context.Context, gameID string, newPhase types.GamePhase) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, exists := r.games[gameID]
	if !exists {
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldPhase := g.CurrentPhase
	g.CurrentPhase = newPhase

	// Publish event
	events.Publish(r.eventBus, events.GamePhaseChangedEvent{
		GameID:   gameID,
		OldPhase: string(oldPhase),
		NewPhase: string(newPhase),
	})

	logger.Get().Debug("Game phase updated",
		zap.String("game_id", gameID),
		zap.String("old_phase", string(oldPhase)),
		zap.String("new_phase", string(newPhase)))

	return nil
}

// UpdateStatus updates the game status
func (r *RepositoryImpl) UpdateStatus(ctx context.Context, gameID string, newStatus types.GameStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, exists := r.games[gameID]
	if !exists {
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	oldStatus := g.Status
	g.Status = newStatus

	// Publish event
	events.Publish(r.eventBus, events.GameStatusChangedEvent{
		GameID:    gameID,
		OldStatus: string(oldStatus),
		NewStatus: string(newStatus),
	})

	logger.Get().Debug("Game status updated",
		zap.String("game_id", gameID),
		zap.String("old_status", string(oldStatus)),
		zap.String("new_status", string(newStatus)))

	return nil
}

// SetCurrentTurn sets which player's turn it is
func (r *RepositoryImpl) SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, exists := r.games[gameID]
	if !exists {
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	g.CurrentTurn = playerID

	logger.Get().Debug("Current turn updated",
		zap.String("game_id", gameID),
		zap.Stringp("player_id", playerID))

	return nil
}

// AddPlayer adds a player ID reference to the game
// Note: Player is already added to Game.Players via Session.CreateAndAddPlayer
// This method just publishes the event for event-driven broadcasting
func (r *RepositoryImpl) AddPlayer(ctx context.Context, gameID string, playerID string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	g, exists := r.games[gameID]
	if !exists {
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	// Verify player exists in game (already added by Session)
	if _, exists := g.Players[playerID]; !exists {
		return fmt.Errorf("player %s not found in game %s (should be added via Session first)", playerID, gameID)
	}

	// Publish event (event-driven broadcasting)
	events.Publish(r.eventBus, events.PlayerJoinedEvent{
		GameID:   gameID,
		PlayerID: playerID,
	})

	logger.Get().Debug("Player join event published",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID))

	return nil
}

// SetHostPlayer sets the host player for the game
func (r *RepositoryImpl) SetHostPlayer(ctx context.Context, gameID string, playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, exists := r.games[gameID]
	if !exists {
		return &types.NotFoundError{Resource: "game", ID: gameID}
	}

	g.HostPlayerID = playerID

	logger.Get().Debug("Host player set",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID))

	return nil
}
