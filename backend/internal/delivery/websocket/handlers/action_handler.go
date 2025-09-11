package handlers

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/actions"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/core/broadcast"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// ModularActionHandler handles game action requests using specialized action handlers
type ModularActionHandler struct {
	// Specialized action handlers
	lifecycleActions *actions.LifecycleActions
	skipActions      *actions.SkipActions
	cardActions      *actions.CardActions
	resourceActions  *actions.ResourceActions
	placementActions *actions.PlacementActions

	// Utilities
	parser       *utils.MessageParser
	errorHandler *utils.ErrorHandler
	logger       *zap.Logger
}

// NewModularActionHandler creates a new action handler with specialized modules
func NewModularActionHandler(
	gameService service.GameService,
	playerService service.PlayerService,
	standardProjectService service.StandardProjectService,
	cardService service.CardService,
	broadcaster *broadcast.Broadcaster,
) *ModularActionHandler {
	parser := utils.NewMessageParser()
	errorHandler := utils.NewErrorHandler()
	logger := logger.Get()

	return &ModularActionHandler{
		lifecycleActions: actions.NewLifecycleActions(gameService),
		skipActions:      actions.NewSkipActions(gameService, playerService, broadcaster),
		cardActions:      actions.NewCardActions(cardService, gameService, parser),
		resourceActions:  actions.NewResourceActions(standardProjectService, parser),
		placementActions: actions.NewPlacementActions(standardProjectService, parser),
		parser:           parser,
		errorHandler:     errorHandler,
		logger:           logger,
	}
}

// HandleMessage implements the MessageHandler interface
func (ah *ModularActionHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	switch message.Type {
	case dto.MessageTypePlayAction:
		ah.handlePlayAction(ctx, connection, message)
	}
}

// handlePlayAction handles game action requests
func (ah *ModularActionHandler) handlePlayAction(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
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

	// Route to appropriate specialized action handler
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

// routeAction routes actions to the appropriate specialized handler
func (ah *ModularActionHandler) routeAction(ctx context.Context, gameID, playerID, actionType string, actionRequest interface{}) error {
	switch dto.ActionType(actionType) {
	// Lifecycle actions
	case dto.ActionTypeStartGame:
		return ah.lifecycleActions.StartGame(ctx, gameID, playerID)
	case dto.ActionTypeSkipAction:
		return ah.skipActions.SkipAction(ctx, gameID, playerID)

	// Card actions
	case dto.ActionTypeSelectStartingCard:
		return ah.cardActions.SelectStartingCards(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypeSelectCards:
		return ah.cardActions.SelectProductionCards(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypePlayCard:
		return ah.cardActions.PlayCard(ctx, gameID, playerID, actionRequest)

	// Resource actions
	case dto.ActionTypeSellPatents:
		return ah.resourceActions.SellPatents(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypeBuildPowerPlant:
		return ah.resourceActions.BuildPowerPlant(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypeLaunchAsteroid:
		return ah.resourceActions.LaunchAsteroid(ctx, gameID, playerID, actionRequest)

	// Placement actions
	case dto.ActionTypeBuildAquifer:
		return ah.placementActions.BuildAquifer(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypePlantGreenery:
		return ah.placementActions.PlantGreenery(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypeBuildCity:
		return ah.placementActions.BuildCity(ctx, gameID, playerID, actionRequest)

	default:
		ah.logger.Warn("Unsupported action type", zap.String("action_type", actionType))
		return ErrModularUnsupportedActionType
	}
}

// Custom errors
var (
	ErrModularUnsupportedActionType = fmt.Errorf("unsupported action type")
)
