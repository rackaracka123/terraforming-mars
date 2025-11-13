package play_card_action

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/game/actions"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles play card action requests
type Handler struct {
	playCardActionAction *actions.PlayCardActionAction
	parser               *utils.MessageParser
	errorHandler         *utils.ErrorHandler
	logger               *zap.Logger
}

// NewHandler creates a new play card action handler
func NewHandler(playCardActionAction *actions.PlayCardActionAction, parser *utils.MessageParser) *Handler {
	return &Handler{
		playCardActionAction: playCardActionAction,
		parser:               parser,
		errorHandler:         utils.NewErrorHandler(),
		logger:               logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Play card action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🎯 Processing play card action request",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Parse the action payload
	var request dto.ActionPlayCardActionRequest
	if err := h.parser.ParsePayload(message.Payload, &request); err != nil {
		h.logger.Warn("Failed to parse play card action request payload",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID),
			zap.Error(err))
		h.errorHandler.SendError(connection, "Invalid play card action request format")
		return
	}

	h.logger.Debug("📋 Play card action request parsed",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("card_id", request.CardID),
		zap.Int("behavior_index", request.BehaviorIndex))

	// Validate required fields
	if request.CardID == "" {
		h.logger.Warn("Play card action request missing card ID",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID))
		h.errorHandler.SendError(connection, "Card ID is required")
		return
	}

	if request.BehaviorIndex < 0 {
		h.logger.Warn("Play card action request has invalid behavior index",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID),
			zap.Int("behavior_index", request.BehaviorIndex))
		h.errorHandler.SendError(connection, "Behavior index must be non-negative")
		return
	}

	// Execute the play card action via PlayCardActionAction
	err := h.playCardActionAction.Execute(ctx, gameID, playerID, request.CardID, request.BehaviorIndex, request.ChoiceIndex, request.CardStorageTarget)
	if err != nil {
		h.logger.Warn("Failed to play card action via PlayCardActionAction",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID),
			zap.String("card_id", request.CardID),
			zap.Int("behavior_index", request.BehaviorIndex),
			zap.Error(err))
		h.errorHandler.SendError(connection, "Failed to play card action: "+err.Error())
		return
	}

	h.logger.Info("✅ Card action played successfully via PlayCardActionAction",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("card_id", request.CardID),
		zap.Int("behavior_index", request.BehaviorIndex))

	// The PlayCardActionAction will broadcast game state, so we don't need to send a response here
	// The client will receive the updated game state via the game-updated event
}
