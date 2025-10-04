package select_corporation

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

// Handler handles select corporation action requests
type Handler struct {
	cardService  service.CardService
	gameService  service.GameService
	parser       *utils.MessageParser
	errorHandler *utils.ErrorHandler
	logger       *zap.Logger
}

// NewHandler creates a new select corporation handler
func NewHandler(cardService service.CardService, gameService service.GameService, parser *utils.MessageParser) *Handler {
	return &Handler{
		cardService:  cardService,
		gameService:  gameService,
		parser:       parser,
		errorHandler: utils.NewErrorHandler(),
		logger:       logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Select corporation action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🏢 Processing select corporation action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Parse the action payload
	var request dto.ActionSelectCorporationRequest
	if err := h.parser.ParsePayload(message.Payload, &request); err != nil {
		h.logger.Error("Failed to parse select corporation payload",
			zap.Error(err),
			zap.String("player_id", playerID))
		h.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	// Execute the action
	if err := h.handle(ctx, gameID, playerID, request.CorporationID); err != nil {
		h.logger.Error("Failed to select corporation",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	// Game state broadcasting is handled by the CardService

	h.logger.Info("✅ Select corporation action completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.String("corporation_id", request.CorporationID))
}

// handle processes the select corporation action (internal method)
func (h *Handler) handle(ctx context.Context, gameID, playerID string, corporationID string) error {
	h.logCorporationSelection(gameID, playerID, corporationID)

	return h.selectCorporation(ctx, gameID, playerID, corporationID)
}

// logCorporationSelection logs the corporation selection attempt
func (h *Handler) logCorporationSelection(gameID, playerID string, corporationID string) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player selecting corporation",
		zap.String("corporation_id", corporationID))
}

// selectCorporation processes the corporation selection through the service
func (h *Handler) selectCorporation(ctx context.Context, gameID, playerID string, corporationID string) error {
	log := logger.WithGameContext(gameID, playerID)

	if err := h.cardService.OnSelectCorporation(ctx, gameID, playerID, corporationID); err != nil {
		log.Error("Failed to select corporation", zap.Error(err))
		return fmt.Errorf("corporation selection failed: %w", err)
	}

	log.Info("Player selected corporation",
		zap.String("selected_corporation", corporationID))

	return nil
}
