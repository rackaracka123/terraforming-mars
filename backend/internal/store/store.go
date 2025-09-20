package store

import (
	"context"
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Reducer is a function that takes the current state and an action and returns new state
type Reducer func(ApplicationState, Action) (ApplicationState, error)

// Middleware is a function that wraps the reducer with additional functionality
type Middleware func(Reducer) Reducer

// Store manages the application state and handles action dispatching
type Store struct {
	state       ApplicationState
	reducer     Reducer
	eventBus    events.EventBus
	subscribers []Subscriber
	mutex       sync.RWMutex
}

// Subscriber is a function that gets called when state changes
type Subscriber func(ApplicationState, Action)

// NewStore creates a new store with the given reducer and middleware
func NewStore(rootReducer Reducer, eventBus events.EventBus, middleware ...Middleware) *Store {
	// Apply middleware to the reducer
	finalReducer := rootReducer
	for i := len(middleware) - 1; i >= 0; i-- {
		finalReducer = middleware[i](finalReducer)
	}

	return &Store{
		state:       NewApplicationState(),
		reducer:     finalReducer,
		eventBus:    eventBus,
		subscribers: make([]Subscriber, 0),
	}
}

// GetState returns a copy of the current state
func (s *Store) GetState() ApplicationState {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state
}

// Dispatch executes an action and updates the state
func (s *Store) Dispatch(ctx context.Context, action Action) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger.Debug("🎯 Dispatching action",
		zap.String("type", string(action.Type)),
		zap.String("game_id", action.Meta.GameID),
		zap.String("player_id", action.Meta.PlayerID),
		zap.String("source", action.Meta.Source))

	// Apply the action through the reducer
	newState, err := s.reducer(s.state, action)
	if err != nil {
		logger.Error("❌ Action failed",
			zap.String("type", string(action.Type)),
			zap.Error(err))
		return err
	}

	// Update state
	s.state = newState

	logger.Debug("✅ Action completed successfully",
		zap.String("type", string(action.Type)),
		zap.String("game_id", action.Meta.GameID))

	// Notify subscribers
	for _, subscriber := range s.subscribers {
		subscriber(s.state, action)
	}

	return nil
}

// Subscribe adds a subscriber that will be called when state changes
func (s *Store) Subscribe(subscriber Subscriber) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.subscribers = append(s.subscribers, subscriber)
}

// GetGame returns a game by ID from current state
func (s *Store) GetGame(gameID string) (GameState, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state.GetGame(gameID)
}

// GetPlayer returns a player by ID from current state
func (s *Store) GetPlayer(playerID string) (PlayerState, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state.GetPlayer(playerID)
}

// GetGamePlayers returns all players in a game from current state
func (s *Store) GetGamePlayers(gameID string) []PlayerState {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state.GetGamePlayers(gameID)
}

// ListGames returns all games with optional status filter
func (s *Store) ListGames(status string) []GameState {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	games := make([]GameState, 0)
	for _, game := range s.state.Games() {
		if status == "" || string(game.Game().Status) == status {
			games = append(games, game)
		}
	}
	return games
}
