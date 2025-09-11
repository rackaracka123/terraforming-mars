package broadcast

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Broadcaster handles sending messages to WebSocket connections
type Broadcaster struct {
	manager       *core.Manager
	gameService   service.GameService
	playerService service.PlayerService
	logger        *zap.Logger

	// Specialized broadcasters
	gameUpdates    *GameUpdates
	playerEvents   *PlayerEvents
	systemMessages *SystemMessages
}

// NewBroadcaster creates a new message broadcaster
func NewBroadcaster(manager *core.Manager, gameService service.GameService, playerService service.PlayerService) *Broadcaster {
	logger := logger.Get()

	b := &Broadcaster{
		manager:       manager,
		gameService:   gameService,
		playerService: playerService,
		logger:        logger,
	}

	// Initialize specialized broadcasters
	b.gameUpdates = NewGameUpdates(manager, gameService, playerService, logger)
	b.playerEvents = NewPlayerEvents(manager, gameService, playerService, logger)
	b.systemMessages = NewSystemMessages(manager, logger)

	return b
}

// Core broadcasting methods (delegate to basic broadcaster)
func (b *Broadcaster) BroadcastToGame(gameID string, message dto.WebSocketMessage) {
	b.systemMessages.BroadcastToGame(gameID, message)
}

func (b *Broadcaster) BroadcastToGameExcept(gameID string, message dto.WebSocketMessage, excludeConnection *core.Connection) {
	b.systemMessages.BroadcastToGameExcept(gameID, message, excludeConnection)
}

func (b *Broadcaster) SendToConnection(connection *core.Connection, message dto.WebSocketMessage) {
	b.systemMessages.SendToConnection(connection, message)
}

// Game update methods (delegate to game updates)
func (b *Broadcaster) SendPersonalizedGameUpdates(ctx context.Context, gameID string) {
	b.gameUpdates.SendPersonalizedGameUpdates(ctx, gameID)
}

// Player event methods (delegate to player events)
func (b *Broadcaster) BroadcastPlayerDisconnection(ctx context.Context, playerID, gameID string, connection *core.Connection) {
	b.playerEvents.BroadcastPlayerDisconnection(ctx, playerID, gameID, connection)
}

// Card-specific methods (delegate to game updates for now)
func (b *Broadcaster) SendAvailableCardsToPlayer(ctx context.Context, gameID, playerID string, cards []dto.CardDto) {
	b.gameUpdates.SendAvailableCardsToPlayer(ctx, gameID, playerID, cards)
}

func (b *Broadcaster) BroadcastProductionPhaseStarted(ctx context.Context, gameID string, playersData []dto.PlayerProductionData) {
	b.gameUpdates.BroadcastProductionPhaseStarted(ctx, gameID, playersData)
}
