package websocket

import (
	"sync"

	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game/board"
	"terraforming-mars-backend/internal/session/game/card"
	game "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

// BroadcasterFactory creates and manages session-aware broadcasters
// Each game gets its own broadcaster instance that is bound to that specific gameID
type BroadcasterFactory struct {
	gameRepo       game.Repository
	sessionFactory session.SessionFactory
	cardRepo       card.Repository
	boardRepo      board.Repository
	hub            *core.Hub
	eventBus       *events.EventBusImpl

	broadcasters map[string]session.SessionManager
	mu           sync.RWMutex
	logger       *zap.Logger
}

// NewBroadcasterFactory creates a new factory for session-aware broadcasters
func NewBroadcasterFactory(
	gameRepo game.Repository,
	sessionFactory session.SessionFactory,
	cardRepo card.Repository,
	boardRepo board.Repository,
	hub *core.Hub,
	eventBus *events.EventBusImpl,
) session.SessionManagerFactory {
	factory := &BroadcasterFactory{
		gameRepo:       gameRepo,
		sessionFactory: sessionFactory,
		cardRepo:       cardRepo,
		boardRepo:      boardRepo,
		hub:            hub,
		eventBus:       eventBus,
		broadcasters:   make(map[string]session.SessionManager),
		logger:         logger.Get(),
	}

	// Subscribe to domain events once at factory level
	// When events fire, we find the appropriate game-specific broadcaster
	factory.subscribeToEvents()

	return factory
}

// GetOrCreate returns the SessionManager for a specific game, creating it if needed
func (f *BroadcasterFactory) GetOrCreate(gameID string) session.SessionManager {
	// Try read lock first (fast path for existing broadcasters)
	f.mu.RLock()
	if broadcaster, exists := f.broadcasters[gameID]; exists {
		f.mu.RUnlock()
		return broadcaster
	}
	f.mu.RUnlock()

	// Need to create new broadcaster (write lock)
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock
	if broadcaster, exists := f.broadcasters[gameID]; exists {
		return broadcaster
	}

	// Create new session-aware broadcaster
	broadcaster := NewSessionBroadcaster(
		gameID,
		f.gameRepo,
		f.sessionFactory,
		f.cardRepo,
		f.boardRepo,
		f.hub,
	)

	f.broadcasters[gameID] = broadcaster
	f.logger.Info("📡 Created session-aware broadcaster for game", zap.String("game_id", gameID))

	return broadcaster
}

// Remove destroys the SessionManager for a game (when game ends)
func (f *BroadcasterFactory) Remove(gameID string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.broadcasters, gameID)
	f.logger.Info("🗑️  Removed broadcaster for game", zap.String("game_id", gameID))
}

// subscribeToEvents subscribes to domain events and routes to game-specific broadcasters
func (f *BroadcasterFactory) subscribeToEvents() {
	log := f.logger

	// Game state change events
	events.Subscribe(f.eventBus, func(e events.GameStatusChangedEvent) {
		log.Debug("🎮 GameStatusChanged event received, broadcasting",
			zap.String("game_id", e.GameID),
			zap.String("new_status", e.NewStatus))
		broadcaster := f.GetOrCreate(e.GameID)
		if err := broadcaster.Broadcast(); err != nil {
			log.Error("Failed to broadcast after GameStatusChanged", zap.Error(err))
		}
	})

	events.Subscribe(f.eventBus, func(e events.GamePhaseChangedEvent) {
		log.Debug("🎮 GamePhaseChanged event received, broadcasting",
			zap.String("game_id", e.GameID),
			zap.String("new_phase", e.NewPhase))
		broadcaster := f.GetOrCreate(e.GameID)
		if err := broadcaster.Broadcast(); err != nil {
			log.Error("Failed to broadcast after GamePhaseChanged", zap.Error(err))
		}
	})

	events.Subscribe(f.eventBus, func(e events.GenerationAdvancedEvent) {
		log.Debug("🎮 GenerationAdvanced event received, broadcasting",
			zap.String("game_id", e.GameID),
			zap.Int("new_generation", e.NewGeneration))
		broadcaster := f.GetOrCreate(e.GameID)
		if err := broadcaster.Broadcast(); err != nil {
			log.Error("Failed to broadcast after GenerationAdvanced", zap.Error(err))
		}
	})

	// Global parameter change events
	events.Subscribe(f.eventBus, func(e events.TemperatureChangedEvent) {
		log.Debug("🌡️  TemperatureChanged event received, broadcasting",
			zap.String("game_id", e.GameID),
			zap.Int("new_value", e.NewValue))
		broadcaster := f.GetOrCreate(e.GameID)
		if err := broadcaster.Broadcast(); err != nil {
			log.Error("Failed to broadcast after TemperatureChanged", zap.Error(err))
		}
	})

	events.Subscribe(f.eventBus, func(e events.OxygenChangedEvent) {
		log.Debug("💨 OxygenChanged event received, broadcasting",
			zap.String("game_id", e.GameID),
			zap.Int("new_value", e.NewValue))
		broadcaster := f.GetOrCreate(e.GameID)
		if err := broadcaster.Broadcast(); err != nil {
			log.Error("Failed to broadcast after OxygenChanged", zap.Error(err))
		}
	})

	events.Subscribe(f.eventBus, func(e events.OceansChangedEvent) {
		log.Debug("🌊 OceansChanged event received, broadcasting",
			zap.String("game_id", e.GameID),
			zap.Int("new_value", e.NewValue))
		broadcaster := f.GetOrCreate(e.GameID)
		if err := broadcaster.Broadcast(); err != nil {
			log.Error("Failed to broadcast after OceansChanged", zap.Error(err))
		}
	})

	// Player state change events
	events.Subscribe(f.eventBus, func(e events.PlayerJoinedEvent) {
		log.Debug("👤 PlayerJoined event received, broadcasting",
			zap.String("game_id", e.GameID),
			zap.String("player_id", e.PlayerID))
		broadcaster := f.GetOrCreate(e.GameID)
		if err := broadcaster.Broadcast(); err != nil {
			log.Error("Failed to broadcast after PlayerJoined", zap.Error(err))
		}
	})

	events.Subscribe(f.eventBus, func(e events.ResourcesChangedEvent) {
		log.Debug("💰 ResourcesChanged event received, broadcasting",
			zap.String("game_id", e.GameID),
			zap.String("player_id", e.PlayerID),
			zap.String("resource_type", e.ResourceType))
		broadcaster := f.GetOrCreate(e.GameID)
		if err := broadcaster.Broadcast(); err != nil {
			log.Error("Failed to broadcast after ResourcesChanged", zap.Error(err))
		}
	})

	events.Subscribe(f.eventBus, func(e events.TerraformRatingChangedEvent) {
		log.Debug("⭐ TerraformRatingChanged event received, broadcasting",
			zap.String("game_id", e.GameID),
			zap.String("player_id", e.PlayerID),
			zap.Int("new_rating", e.NewRating))
		broadcaster := f.GetOrCreate(e.GameID)
		if err := broadcaster.Broadcast(); err != nil {
			log.Error("Failed to broadcast after TerraformRatingChanged", zap.Error(err))
		}
	})

	events.Subscribe(f.eventBus, func(e events.CardHandUpdatedEvent) {
		log.Debug("🃏 CardHandUpdated event received, broadcasting",
			zap.String("game_id", e.GameID),
			zap.String("player_id", e.PlayerID))
		broadcaster := f.GetOrCreate(e.GameID)
		if err := broadcaster.Broadcast(); err != nil {
			log.Error("Failed to broadcast after CardHandUpdated", zap.Error(err))
		}
	})

	events.Subscribe(f.eventBus, func(e events.TilePlacedEvent) {
		log.Debug("🗺️  TilePlaced event received, broadcasting",
			zap.String("game_id", e.GameID),
			zap.String("player_id", e.PlayerID),
			zap.String("tile_type", e.TileType))
		broadcaster := f.GetOrCreate(e.GameID)
		if err := broadcaster.Broadcast(); err != nil {
			log.Error("Failed to broadcast after TilePlaced", zap.Error(err))
		}
	})

	log.Info("📡 BroadcasterFactory subscribed to all domain events for automatic broadcasting")
}
