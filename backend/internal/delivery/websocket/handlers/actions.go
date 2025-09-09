package handlers

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/actions"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// ActionHandler handles game action requests
type ActionHandler struct {
	gameActions      *actions.GameActions
	standardProjects *actions.StandardProjects
	cardActions      *actions.CardActions
	parser           *utils.MessageParser
	errorHandler     *utils.ErrorHandler
	broadcaster      *core.Broadcaster
	logger           *zap.Logger
}

// NewActionHandler creates a new action handler
func NewActionHandler(
	gameService service.GameService,
	playerService service.PlayerService,
	standardProjectService service.StandardProjectService,
	cardService service.CardService,
	broadcaster *core.Broadcaster,
) *ActionHandler {
	return &ActionHandler{
		gameActions:      actions.NewGameActions(gameService, playerService, broadcaster),
		standardProjects: actions.NewStandardProjects(standardProjectService),
		cardActions:      actions.NewCardActions(cardService, gameService),
		parser:           utils.NewMessageParser(),
		errorHandler:     utils.NewErrorHandler(),
		broadcaster:      broadcaster,
		logger:           logger.Get(),
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

// routeAction routes actions to the appropriate handler
func (ah *ActionHandler) routeAction(ctx context.Context, gameID, playerID, actionType string, actionRequest interface{}) error {
	switch dto.ActionType(actionType) {
	// Game actions
	case dto.ActionTypeStartGame:
		return ah.gameActions.StartGame(ctx, gameID, playerID)
	case dto.ActionTypeSkipAction:
		return ah.gameActions.SkipAction(ctx, gameID, playerID)

	// Card actions
	case dto.ActionTypeSelectStartingCard:
		return ah.cardActions.SelectStartingCards(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypeSelectCards:
		return ah.cardActions.SelectProductionCards(ctx, gameID, playerID, actionRequest)

	// Standard projects
	case dto.ActionTypeSellPatents:
		return ah.standardProjects.SellPatents(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypeBuildPowerPlant:
		return ah.standardProjects.BuildPowerPlant(ctx, gameID, playerID)
	case dto.ActionTypeLaunchAsteroid:
		return ah.standardProjects.LaunchAsteroid(ctx, gameID, playerID)
	case dto.ActionTypeBuildAquifer:
		return ah.standardProjects.BuildAquifer(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypePlantGreenery:
		return ah.standardProjects.PlantGreenery(ctx, gameID, playerID, actionRequest)
	case dto.ActionTypeBuildCity:
		return ah.standardProjects.BuildCity(ctx, gameID, playerID, actionRequest)

	default:
		ah.logger.Warn("Unsupported action type", zap.String("action_type", actionType))
		return ErrUnsupportedActionType
	}
}

// Custom errors
var (
	ErrUnsupportedActionType = fmt.Errorf("unsupported action type")
)
