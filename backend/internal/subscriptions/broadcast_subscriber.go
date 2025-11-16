package subscriptions

import (
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// BroadcastSubscriber listens to all domain events and automatically broadcasts
// game state updates to connected players via WebSocket
type BroadcastSubscriber struct {
	sessionManager session.SessionManager
	eventBus       *events.EventBusImpl
	subscriptions  []events.SubscriptionID
}

// NewBroadcastSubscriber creates a new broadcast subscriber and registers all event handlers
func NewBroadcastSubscriber(
	sessionManager session.SessionManager,
	eventBus *events.EventBusImpl,
) *BroadcastSubscriber {
	subscriber := &BroadcastSubscriber{
		sessionManager: sessionManager,
		eventBus:       eventBus,
		subscriptions:  make([]events.SubscriptionID, 0),
	}
	subscriber.registerSubscriptions()
	return subscriber
}

// registerSubscriptions subscribes to all domain events that should trigger broadcasts
func (s *BroadcastSubscriber) registerSubscriptions() {
	// Game events
	s.subscriptions = append(s.subscriptions,
		events.Subscribe(s.eventBus, s.handleGamePhaseChanged),
		events.Subscribe(s.eventBus, s.handleGameStatusChanged),
		events.Subscribe(s.eventBus, s.handleGenerationAdvanced),
	)

	// Player events
	s.subscriptions = append(s.subscriptions,
		events.Subscribe(s.eventBus, s.handleTerraformRatingChanged),
		events.Subscribe(s.eventBus, s.handleCardPlayed),
		events.Subscribe(s.eventBus, s.handleCardAddedToHand),
		events.Subscribe(s.eventBus, s.handleCorporationSelected),
		events.Subscribe(s.eventBus, s.handleResourcesChanged),
		events.Subscribe(s.eventBus, s.handleProductionChanged),
		events.Subscribe(s.eventBus, s.handleVictoryPointsChanged),
		events.Subscribe(s.eventBus, s.handlePlayerEffectAdded),
		events.Subscribe(s.eventBus, s.handleResourceStorageChanged),
	)

	// Global parameter events
	s.subscriptions = append(s.subscriptions,
		events.Subscribe(s.eventBus, s.handleTemperatureChanged),
		events.Subscribe(s.eventBus, s.handleOxygenChanged),
		events.Subscribe(s.eventBus, s.handleOceansChanged),
	)

	// Tile events
	s.subscriptions = append(s.subscriptions,
		events.Subscribe(s.eventBus, s.handleTilePlaced),
		events.Subscribe(s.eventBus, s.handlePlacementBonusGained),
	)

	log := logger.Get()
	log.Info("📢 BroadcastSubscriber registered event subscriptions", zap.Int("count", len(s.subscriptions)))
}

// Unsubscribe removes all event subscriptions (useful for cleanup/testing)
func (s *BroadcastSubscriber) Unsubscribe() {
	for _, subID := range s.subscriptions {
		s.eventBus.Unsubscribe(subID)
	}
	s.subscriptions = nil
	log := logger.Get()
	log.Info("BroadcastSubscriber unsubscribed from all events")
}

// Game event handlers

func (s *BroadcastSubscriber) handleGamePhaseChanged(event events.GamePhaseChangedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: phase changed", zap.String("game_id", event.GameID), zap.String("phase", string(event.NewPhase)))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast game phase change", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handleGameStatusChanged(event events.GameStatusChangedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: status changed", zap.String("game_id", event.GameID), zap.String("status", event.NewStatus))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast game status change", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handleGenerationAdvanced(event events.GenerationAdvancedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: generation advanced", zap.String("game_id", event.GameID), zap.Int("generation", event.NewGeneration))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast generation advance", zap.Error(err))
	}
}

// Player event handlers

func (s *BroadcastSubscriber) handleTerraformRatingChanged(event events.TerraformRatingChangedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: TR changed", zap.String("game_id", event.GameID), zap.String("player_id", event.PlayerID))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast TR change", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handleCardPlayed(event events.CardPlayedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: card played", zap.String("game_id", event.GameID), zap.String("player_id", event.PlayerID), zap.String("card_id", event.CardID))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast card played", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handleCardAddedToHand(event events.CardAddedToHandEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: card added to hand", zap.String("game_id", event.GameID), zap.String("player_id", event.PlayerID))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast card added to hand", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handleCorporationSelected(event events.CorporationSelectedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: corporation selected", zap.String("game_id", event.GameID), zap.String("player_id", event.PlayerID), zap.String("corp_id", event.CorporationID))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast corporation selection", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handleResourcesChanged(event events.ResourcesChangedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: resources changed", zap.String("game_id", event.GameID), zap.String("player_id", event.PlayerID))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast resources change", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handleProductionChanged(event events.ProductionChangedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: production changed", zap.String("game_id", event.GameID), zap.String("player_id", event.PlayerID))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast production change", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handleVictoryPointsChanged(event events.VictoryPointsChangedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: victory points changed", zap.String("game_id", event.GameID), zap.String("player_id", event.PlayerID))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast victory points change", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handlePlayerEffectAdded(event events.PlayerEffectAddedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: effect added", zap.String("game_id", event.GameID), zap.String("player_id", event.PlayerID))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast player effect added", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handleResourceStorageChanged(event events.ResourceStorageChangedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: resource storage changed", zap.String("game_id", event.GameID), zap.String("player_id", event.PlayerID))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast resource storage change", zap.Error(err))
	}
}

// Global parameter event handlers

func (s *BroadcastSubscriber) handleTemperatureChanged(event events.TemperatureChangedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: temperature changed", zap.String("game_id", event.GameID), zap.Int("new_value", event.NewValue))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast temperature change", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handleOxygenChanged(event events.OxygenChangedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: oxygen changed", zap.String("game_id", event.GameID), zap.Int("new_value", event.NewValue))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast oxygen change", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handleOceansChanged(event events.OceansChangedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: oceans changed", zap.String("game_id", event.GameID), zap.Int("new_value", event.NewValue))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast oceans change", zap.Error(err))
	}
}

// Tile event handlers

func (s *BroadcastSubscriber) handleTilePlaced(event events.TilePlacedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: tile placed", zap.String("game_id", event.GameID), zap.Int("q", event.Q), zap.Int("r", event.R))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast tile placement", zap.Error(err))
	}
}

func (s *BroadcastSubscriber) handlePlacementBonusGained(event events.PlacementBonusGainedEvent) {
	log := logger.Get()
	log.Debug("📢 Broadcasting game update: placement bonus gained", zap.String("game_id", event.GameID), zap.String("player_id", event.PlayerID))
	if err := s.sessionManager.Broadcast(event.GameID); err != nil {
		log.Error("Failed to broadcast placement bonus", zap.Error(err))
	}
}
