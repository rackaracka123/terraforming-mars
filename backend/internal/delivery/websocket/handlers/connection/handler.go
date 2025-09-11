package connection

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Handler handles player connection and reconnection logic
type Handler struct {
	gameService     service.GameService
	playerService   service.PlayerService
	broadcaster     *core.Broadcaster
	manager         *core.Manager
	parser          *utils.MessageParser
	errorHandler    *utils.ErrorHandler
	logger          *zap.Logger
	newPlayerFlow   *NewPlayerFlow
	reconnectFlow   *ReconnectFlow
	validator       *Validator
}

// NewHandler creates a new connection handler
func NewHandler(
	gameService service.GameService,
	playerService service.PlayerService,
	broadcaster *core.Broadcaster,
	manager *core.Manager,
) *Handler {
	parser := utils.NewMessageParser()
	errorHandler := utils.NewErrorHandler()
	logger := logger.Get()
	validator := NewValidator(gameService, playerService, parser, errorHandler, logger)

	return &Handler{
		gameService:   gameService,
		playerService: playerService,
		broadcaster:   broadcaster,
		manager:       manager,
		parser:        parser,
		errorHandler:  errorHandler,
		logger:        logger,
		newPlayerFlow: NewNewPlayerFlow(gameService, playerService, broadcaster, parser, errorHandler, logger),
		reconnectFlow: NewReconnectFlow(gameService, playerService, broadcaster, manager, parser, errorHandler, logger),
		validator:     validator,
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	switch message.Type {
	case dto.MessageTypePlayerConnect:
		h.handlePlayerConnect(ctx, connection, message)
	case dto.MessageTypePlayerReconnect:
		h.handlePlayerReconnect(ctx, connection, message)
	}
}

// handlePlayerConnect handles player connection requests
func (h *Handler) handlePlayerConnect(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	h.logger.Debug("🚪 Starting player connect handler", zap.String("connection_id", connection.ID))

	// Parse and validate payload
	var payload dto.PlayerConnectPayload
	if !h.validator.ParseAndValidateConnectPayload(connection, message, &payload) {
		return
	}

	// Validate game exists
	if !h.validator.ValidateGameExists(ctx, connection, payload.GameID) {
		return
	}

	// Check if player already exists
	existingPlayerID := h.validator.FindExistingPlayer(ctx, payload.GameID, payload.PlayerName)
	isNewPlayer := existingPlayerID == ""

	if isNewPlayer {
		h.newPlayerFlow.HandleNewPlayer(ctx, connection, payload)
	} else {
		h.reconnectFlow.HandleExistingPlayerConnect(ctx, connection, payload, existingPlayerID)
	}
}

// handlePlayerReconnect handles player reconnection requests
func (h *Handler) handlePlayerReconnect(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	var payload dto.PlayerReconnectPayload
	if err := h.parser.ParsePayload(message.Payload, &payload); err != nil {
		h.logger.Error("Failed to parse player reconnect payload", zap.Error(err))
		h.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	h.reconnectFlow.HandleReconnect(ctx, connection, payload)
}

