package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// ActionHandler handles game action requests
type ActionHandler struct {
	actionRegistry *core.ActionRegistry
	parser         *utils.MessageParser
	errorHandler   *utils.ErrorHandler
	logger         *zap.Logger
}

// NewActionHandler creates a new action handler
func NewActionHandler(
	gameService service.GameService,
	playerService service.PlayerService,
	standardProjectService service.StandardProjectService,
	cardService service.CardService,
	broadcaster *core.Broadcaster,
) *ActionHandler {
	parser := utils.NewMessageParser()
	actionRegistry := SetupActionRegistry(gameService, playerService, standardProjectService, cardService, broadcaster)

	return &ActionHandler{
		actionRegistry: actionRegistry,
		parser:         parser,
		errorHandler:   utils.NewErrorHandler(),
		logger:         logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (ah *ActionHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	switch message.Type {
	case dto.MessageTypePlayAction:
		ah.handlePlayAction(ctx, connection, message)
	}
}

// handlePlayAction handles game action requests
func (ah *ActionHandler) handlePlayAction(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		ah.logger.Warn("Action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		ah.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	// Parse action payload
	var payload dto.PlayActionPayload
	if err := ah.parser.ParsePayload(message.Payload, &payload); err != nil {
		ah.logger.Error("Failed to parse play action payload", zap.Error(err))
		ah.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	// Parse action type
	actionType, _, err := ah.parser.ParseAction(payload.ActionRequest)
	if err != nil {
		ah.logger.Error("Failed to parse action type", zap.Error(err))
		ah.errorHandler.SendError(connection, utils.ErrInvalidActionType)
		return
	}

	ah.logger.Debug("🎮 Processing action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.String("action_type", actionType))

	// Route to appropriate action handler
	if err := ah.routeAction(ctx, gameID, playerID, actionType, payload.ActionRequest); err != nil {
		ah.logger.Error("Failed to process action",
			zap.Error(err),
			zap.String("action_type", actionType),
			zap.String("player_id", playerID))
		ah.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	ah.logger.Info("✅ Action processed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("action_type", actionType))
}

// routeAction routes actions to the appropriate handler using the registry
func (ah *ActionHandler) routeAction(ctx context.Context, gameID, playerID, actionType string, actionRequest interface{}) error {
	handler, err := ah.actionRegistry.GetHandler(dto.ActionType(actionType))
	if err != nil {
		ah.logger.Warn("Unsupported action type", zap.String("action_type", actionType))
		return fmt.Errorf("unsupported action type: %s", actionType)
	}

	return handler.Handle(ctx, gameID, playerID, actionRequest)
}
