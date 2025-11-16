package admin_command

import (
	"context"
	"encoding/json"
	"fmt"
	"terraforming-mars-backend/internal/features/card"

	"terraforming-mars-backend/internal/admin"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// Handler handles admin command requests (development mode only)
type Handler struct {
	gameRepo     game.Repository
	cardService  card.CardRepository
	adminService admin.AdminService
	errorHandler *utils.ErrorHandler
	logger       *zap.Logger
}

// NewHandler creates a new admin command handler
func NewHandler(gameRepo game.Repository, cardService card.CardRepository, adminService admin.AdminService) *Handler {
	return &Handler{
		gameRepo:     gameRepo,
		cardService:  cardService,
		adminService: adminService,
		errorHandler: utils.NewErrorHandler(),
		logger:       logger.Get(),
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
	game, err := h.gameRepo.GetByID(ctx, gameID)
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
	case dto.AdminCommandTypeStartTileSelection:
		return h.handleStartTileSelection(ctx, gameID, request.Payload)
	case dto.AdminCommandTypeSetCurrentTurn:
		return h.handleSetCurrentTurn(ctx, gameID, request.Payload)
	case dto.AdminCommandTypeSetCorporation:
		return h.handleSetCorporation(ctx, gameID, request.Payload)
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

	// Use AdminService to give the card
	return h.adminService.OnAdminGiveCard(ctx, gameID, command.PlayerID, command.CardID)
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

	// Use AdminService to set the phase
	return h.adminService.OnAdminSetPhase(ctx, gameID, game.GamePhase(command.Phase))
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
	resourcesData := player.Resources{
		Credits:  command.Resources.Credits,
		Steel:    command.Resources.Steel,
		Titanium: command.Resources.Titanium,
		Plants:   command.Resources.Plants,
		Energy:   command.Resources.Energy,
		Heat:     command.Resources.Heat,
	}

	// Use AdminService to set resources
	return h.adminService.OnAdminSetResources(ctx, gameID, command.PlayerID, resourcesData)
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
	productionData := player.Production{
		Credits:  command.Production.Credits,
		Steel:    command.Production.Steel,
		Titanium: command.Production.Titanium,
		Plants:   command.Production.Plants,
		Energy:   command.Production.Energy,
		Heat:     command.Production.Heat,
	}

	// Use AdminService to set production
	return h.adminService.OnAdminSetProduction(ctx, gameID, command.PlayerID, productionData)
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
	globalParams := parameters.GlobalParameters{
		Temperature: command.GlobalParameters.Temperature,
		Oxygen:      command.GlobalParameters.Oxygen,
		Oceans:      command.GlobalParameters.Oceans,
	}

	// Use AdminService to set global parameters
	return h.adminService.OnAdminSetGlobalParameters(ctx, gameID, globalParams)
}

// handleStartTileSelection starts tile selection for testing
func (h *Handler) handleStartTileSelection(ctx context.Context, gameID string, payload interface{}) error {
	var command dto.StartTileSelectionAdminCommand
	if err := h.parsePayload(payload, &command); err != nil {
		return fmt.Errorf("invalid start tile selection payload: %w", err)
	}

	h.logger.Info("🎯 Admin starting tile selection",
		zap.String("game_id", gameID),
		zap.String("player_id", command.PlayerID),
		zap.String("tile_type", command.TileType))

	// Use AdminService to start tile selection
	return h.adminService.OnAdminStartTileSelection(ctx, gameID, command.PlayerID, command.TileType)
}

// handleSetCurrentTurn sets the current player turn
func (h *Handler) handleSetCurrentTurn(ctx context.Context, gameID string, payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payload type: expected map[string]interface{}, got %T", payload)
	}

	playerID, ok := payloadMap["playerId"].(string)
	if !ok {
		return fmt.Errorf("playerId is required and must be a string")
	}

	h.logger.Info("🔄 Admin setting current turn",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID))

	// Use AdminService to set current turn
	return h.adminService.OnAdminSetCurrentTurn(ctx, gameID, playerID)
}

// handleSetCorporation sets a player's corporation
func (h *Handler) handleSetCorporation(ctx context.Context, gameID string, payload interface{}) error {
	var command dto.SetCorporationAdminCommand
	if err := h.parsePayload(payload, &command); err != nil {
		return fmt.Errorf("invalid set corporation payload: %w", err)
	}

	h.logger.Info("🏢 Admin setting player corporation",
		zap.String("game_id", gameID),
		zap.String("player_id", command.PlayerID),
		zap.String("corporation_id", command.CorporationID))

	// Use AdminService to set corporation
	return h.adminService.OnAdminSetCorporation(ctx, gameID, command.PlayerID, command.CorporationID)
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
