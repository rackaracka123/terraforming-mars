package admin_command

import (
	"context"
	"encoding/json"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Handler handles admin command requests (development mode only)
type Handler struct {
	gameService   service.GameService
	playerService service.PlayerService
	cardService   service.CardService
	errorHandler  *utils.ErrorHandler
	logger        *zap.Logger
}

// NewHandler creates a new admin command handler
func NewHandler(gameService service.GameService, playerService service.PlayerService, cardService service.CardService) *Handler {
	return &Handler{
		gameService:   gameService,
		playerService: playerService,
		cardService:   cardService,
		errorHandler:  utils.NewErrorHandler(),
		logger:        logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Admin command received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🔧 Processing admin command",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// First check if the game is in development mode
	if err := h.validateDevelopmentMode(ctx, gameID); err != nil {
		h.logger.Warn("Admin command rejected - not in development mode",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, "Admin commands are only available in development mode")
		return
	}

	// Parse the admin command request
	var adminRequest dto.AdminCommandRequest
	payloadBytes, err := json.Marshal(message.Payload)
	if err != nil {
		h.logger.Error("Failed to marshal message payload",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, "Invalid message payload")
		return
	}

	if err := json.Unmarshal(payloadBytes, &adminRequest); err != nil {
		h.logger.Error("Failed to parse admin command request",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, "Invalid admin command format")
		return
	}

	// Handle the specific admin command
	if err := h.handleAdminCommand(ctx, gameID, playerID, adminRequest); err != nil {
		h.logger.Error("Failed to execute admin command",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID),
			zap.String("command_type", string(adminRequest.CommandType)))
		h.errorHandler.SendError(connection, "Admin command failed: "+err.Error())
		return
	}

	h.logger.Info("✅ Admin command executed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.String("command_type", string(adminRequest.CommandType)))
}

// validateDevelopmentMode ensures the game is in development mode
func (h *Handler) validateDevelopmentMode(ctx context.Context, gameID string) error {
	game, err := h.gameService.GetGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	if !game.Settings.DevelopmentMode {
		return fmt.Errorf("admin commands are only available in development mode")
	}

	return nil
}

// handleAdminCommand routes and executes the specific admin command
func (h *Handler) handleAdminCommand(ctx context.Context, gameID, playerID string, request dto.AdminCommandRequest) error {
	switch request.CommandType {
	case dto.AdminCommandTypeGiveCard:
		return h.handleGiveCard(ctx, gameID, request.Payload)
	case dto.AdminCommandTypeSetPhase:
		return h.handleSetPhase(ctx, gameID, request.Payload)
	case dto.AdminCommandTypeSetResources:
		return h.handleSetResources(ctx, gameID, request.Payload)
	case dto.AdminCommandTypeSetProduction:
		return h.handleSetProduction(ctx, gameID, request.Payload)
	case dto.AdminCommandTypeSetGlobalParams:
		return h.handleSetGlobalParams(ctx, gameID, request.Payload)
	default:
		return fmt.Errorf("unknown admin command type: %s", request.CommandType)
	}
}

// handleGiveCard gives a specific card to a specific player
func (h *Handler) handleGiveCard(ctx context.Context, gameID string, payload interface{}) error {
	var command dto.GiveCardAdminCommand
	if err := h.parsePayload(payload, &command); err != nil {
		return fmt.Errorf("invalid give card payload: %w", err)
	}

	h.logger.Info("🎴 Admin giving card to player",
		zap.String("game_id", gameID),
		zap.String("player_id", command.PlayerID),
		zap.String("card_id", command.CardID))

	// Verify card exists
	card, err := h.cardService.GetCardByID(ctx, command.CardID)
	if err != nil || card == nil {
		return fmt.Errorf("card not found: %s", command.CardID)
	}

	// Add card to player's hand
	if err := h.playerService.AddCardToHand(ctx, gameID, command.PlayerID, command.CardID); err != nil {
		return fmt.Errorf("failed to add card to player's hand: %w", err)
	}

	return nil
}

// handleSetPhase sets the game phase
func (h *Handler) handleSetPhase(ctx context.Context, gameID string, payload interface{}) error {
	var command dto.SetPhaseAdminCommand
	if err := h.parsePayload(payload, &command); err != nil {
		return fmt.Errorf("invalid set phase payload: %w", err)
	}

	h.logger.Info("🔄 Admin setting game phase",
		zap.String("game_id", gameID),
		zap.String("phase", command.Phase))

	if err := h.gameService.SetGamePhase(ctx, gameID, command.Phase); err != nil {
		return fmt.Errorf("failed to set game phase: %w", err)
	}

	return nil
}

// handleSetResources sets a player's resources
func (h *Handler) handleSetResources(ctx context.Context, gameID string, payload interface{}) error {
	var command dto.SetResourcesAdminCommand
	if err := h.parsePayload(payload, &command); err != nil {
		return fmt.Errorf("invalid set resources payload: %w", err)
	}

	h.logger.Info("💰 Admin setting player resources",
		zap.String("game_id", gameID),
		zap.String("player_id", command.PlayerID),
		zap.Any("resources", command.Resources))

	// Convert DTO to model
	resources := model.Resources{
		Credits:  command.Resources.Credits,
		Steel:    command.Resources.Steel,
		Titanium: command.Resources.Titanium,
		Plants:   command.Resources.Plants,
		Energy:   command.Resources.Energy,
		Heat:     command.Resources.Heat,
	}

	if err := h.playerService.SetResources(ctx, gameID, command.PlayerID, resources); err != nil {
		return fmt.Errorf("failed to set player resources: %w", err)
	}

	return nil
}

// handleSetProduction sets a player's production
func (h *Handler) handleSetProduction(ctx context.Context, gameID string, payload interface{}) error {
	var command dto.SetProductionAdminCommand
	if err := h.parsePayload(payload, &command); err != nil {
		return fmt.Errorf("invalid set production payload: %w", err)
	}

	h.logger.Info("🏭 Admin setting player production",
		zap.String("game_id", gameID),
		zap.String("player_id", command.PlayerID),
		zap.Any("production", command.Production))

	// Convert DTO to model
	production := model.Production{
		Credits:  command.Production.Credits,
		Steel:    command.Production.Steel,
		Titanium: command.Production.Titanium,
		Plants:   command.Production.Plants,
		Energy:   command.Production.Energy,
		Heat:     command.Production.Heat,
	}

	if err := h.playerService.SetProduction(ctx, gameID, command.PlayerID, production); err != nil {
		return fmt.Errorf("failed to set player production: %w", err)
	}

	return nil
}

// handleSetGlobalParams sets the global parameters
func (h *Handler) handleSetGlobalParams(ctx context.Context, gameID string, payload interface{}) error {
	var command dto.SetGlobalParamsAdminCommand
	if err := h.parsePayload(payload, &command); err != nil {
		return fmt.Errorf("invalid set global params payload: %w", err)
	}

	h.logger.Info("🌍 Admin setting global parameters",
		zap.String("game_id", gameID),
		zap.Any("global_parameters", command.GlobalParameters))

	// Convert DTO to model
	globalParams := model.GlobalParameters{
		Temperature: command.GlobalParameters.Temperature,
		Oxygen:      command.GlobalParameters.Oxygen,
		Oceans:      command.GlobalParameters.Oceans,
	}

	if err := h.gameService.SetGlobalParameters(ctx, gameID, globalParams); err != nil {
		return fmt.Errorf("failed to set global parameters: %w", err)
	}

	return nil
}

// parsePayload parses the payload interface{} into the target struct
func (h *Handler) parsePayload(payload interface{}, target interface{}) error {
	// Convert payload to JSON bytes and then unmarshal into target
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}
