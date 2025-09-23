package handlers

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// PlayerHandler handles all player actions within the game
type PlayerHandler struct {
	gameService            service.GameService
	playerService          service.PlayerService
	standardProjectService service.StandardProjectService
	cardService            service.CardService
	broadcaster            *core.Broadcaster
	parser                 *utils.MessageParser
	errorHandler           *utils.ErrorHandler
	logger                 *zap.Logger
}

// NewPlayerHandler creates a new player handler
func NewPlayerHandler(gameService service.GameService, playerService service.PlayerService, standardProjectService service.StandardProjectService, cardService service.CardService, broadcaster *core.Broadcaster, parser *utils.MessageParser) *PlayerHandler {
	return &PlayerHandler{
		gameService:            gameService,
		playerService:          playerService,
		standardProjectService: standardProjectService,
		cardService:            cardService,
		broadcaster:            broadcaster,
		parser:                 parser,
		errorHandler:           utils.NewErrorHandler(),
		logger:                 logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface with player action routing
func (h *PlayerHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Player action received from unassigned connection",
			zap.String("connection_id", connection.ID),
			zap.String("message_type", string(message.Type)))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	// Route to specific player action handler
	switch message.Type {
	// Standard Project Actions
	case dto.MessageTypeActionLaunchAsteroid:
		h.handleLaunchAsteroid(ctx, gameID, playerID, connection)
	case dto.MessageTypeActionSellPatents:
		h.handleSellPatents(ctx, gameID, playerID, connection, message)
	case dto.MessageTypeActionBuildPowerPlant:
		h.handleBuildPowerPlant(ctx, gameID, playerID, connection)
	case dto.MessageTypeActionBuildAquifer:
		h.handleBuildAquifer(ctx, gameID, playerID, connection, message)
	case dto.MessageTypeActionPlantGreenery:
		h.handlePlantGreenery(ctx, gameID, playerID, connection, message)
	case dto.MessageTypeActionBuildCity:
		h.handleBuildCity(ctx, gameID, playerID, connection, message)

	// Card Actions
	case dto.MessageTypeActionPlayCard:
		h.handlePlayCard(ctx, gameID, playerID, connection, message)
	case dto.MessageTypeActionSelectStartingCard:
		h.handleSelectStartingCards(ctx, gameID, playerID, connection, message)
	case dto.MessageTypeActionSelectCards:
		h.handleSelectProductionCards(ctx, gameID, playerID, connection, message)

	// Player Turn Actions
	case dto.MessageTypeActionSkipAction:
		h.handleSkipAction(ctx, gameID, playerID, connection)

	default:
		h.logger.Warn("Unknown player action type", zap.String("type", string(message.Type)))
		h.errorHandler.SendError(connection, "Unknown player action type")
	}
}

// Helper methods for consistent error handling and logging
func (h *PlayerHandler) logAndHandleError(connection *core.Connection, playerID, gameID, action string, err error) {
	h.logger.Error("Player action failed",
		zap.Error(err),
		zap.String("action", action),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
	h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
}

func (h *PlayerHandler) logSuccess(connection *core.Connection, playerID, gameID, action string, extra ...zap.Field) {
	fields := []zap.Field{
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.String("action", action),
	}
	fields = append(fields, extra...)
	h.logger.Info("✅ Player action completed successfully", fields...)
}

func (h *PlayerHandler) parsePayload(connection *core.Connection, playerID string, message dto.WebSocketMessage, target any) bool {
	if err := h.parser.ParsePayload(message.Payload, target); err != nil {
		h.logger.Error("Failed to parse payload",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("message_type", string(message.Type)))
		h.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return false
	}
	return true
}

// Standard Project Actions (no payload required)
func (h *PlayerHandler) handleLaunchAsteroid(ctx context.Context, gameID, playerID string, connection *core.Connection) {
	h.logger.Debug("🚀 Processing launch asteroid action", zap.String("player_id", playerID))

	if err := h.standardProjectService.LaunchAsteroid(ctx, gameID, playerID); err != nil {
		h.logAndHandleError(connection, playerID, gameID, "launch_asteroid", err)
		return
	}
	h.logSuccess(connection, playerID, gameID, "launch_asteroid")
}

func (h *PlayerHandler) handleBuildPowerPlant(ctx context.Context, gameID, playerID string, connection *core.Connection) {
	h.logger.Debug("⚡ Processing build power plant action", zap.String("player_id", playerID))

	if err := h.standardProjectService.BuildPowerPlant(ctx, gameID, playerID); err != nil {
		h.logAndHandleError(connection, playerID, gameID, "build_power_plant", err)
		return
	}
	h.logSuccess(connection, playerID, gameID, "build_power_plant")
}

// Standard Project Actions (with payload)
func (h *PlayerHandler) handleSellPatents(ctx context.Context, gameID, playerID string, connection *core.Connection, message dto.WebSocketMessage) {
	h.logger.Debug("🏛️ Processing sell patents action", zap.String("player_id", playerID))

	var request dto.ActionSellPatentsRequest
	if !h.parsePayload(connection, playerID, message, &request) {
		return
	}

	if err := h.standardProjectService.SellPatents(ctx, gameID, playerID, request.CardCount); err != nil {
		h.logAndHandleError(connection, playerID, gameID, "sell_patents", err)
		return
	}
	h.logSuccess(connection, playerID, gameID, "sell_patents", zap.Int("card_count", request.CardCount))
}

func (h *PlayerHandler) handleBuildAquifer(ctx context.Context, gameID, playerID string, connection *core.Connection, message dto.WebSocketMessage) {
	h.logger.Debug("🌊 Processing build aquifer action", zap.String("player_id", playerID))

	var request dto.ActionBuildAquiferRequest
	if !h.parsePayload(connection, playerID, message, &request) {
		return
	}

	hexPosition := dto.ConvertHexPosition(request.HexPosition)
	if err := h.standardProjectService.BuildAquifer(ctx, gameID, playerID, hexPosition); err != nil {
		h.logAndHandleError(connection, playerID, gameID, "build_aquifer", err)
		return
	}
	h.logSuccess(connection, playerID, gameID, "build_aquifer", zap.Any("hex_position", request.HexPosition))
}

func (h *PlayerHandler) handlePlantGreenery(ctx context.Context, gameID, playerID string, connection *core.Connection, message dto.WebSocketMessage) {
	h.logger.Debug("🌱 Processing plant greenery action", zap.String("player_id", playerID))

	var request dto.ActionPlantGreeneryRequest
	if !h.parsePayload(connection, playerID, message, &request) {
		return
	}

	hexPosition := dto.ConvertHexPosition(request.HexPosition)
	if err := h.standardProjectService.PlantGreenery(ctx, gameID, playerID, hexPosition); err != nil {
		h.logAndHandleError(connection, playerID, gameID, "plant_greenery", err)
		return
	}
	h.logSuccess(connection, playerID, gameID, "plant_greenery", zap.Any("hex_position", request.HexPosition))
}

func (h *PlayerHandler) handleBuildCity(ctx context.Context, gameID, playerID string, connection *core.Connection, message dto.WebSocketMessage) {
	h.logger.Debug("🏙️ Processing build city action", zap.String("player_id", playerID))

	var request dto.ActionBuildCityRequest
	if !h.parsePayload(connection, playerID, message, &request) {
		return
	}

	hexPosition := dto.ConvertHexPosition(request.HexPosition)
	if err := h.standardProjectService.BuildCity(ctx, gameID, playerID, hexPosition); err != nil {
		h.logAndHandleError(connection, playerID, gameID, "build_city", err)
		return
	}
	h.logSuccess(connection, playerID, gameID, "build_city", zap.Any("hex_position", request.HexPosition))
}

// Card Actions
func (h *PlayerHandler) handlePlayCard(ctx context.Context, gameID, playerID string, connection *core.Connection, message dto.WebSocketMessage) {
	h.logger.Debug("🎯 Processing play card action", zap.String("player_id", playerID))

	var request dto.ActionPlayCardRequest
	if !h.parsePayload(connection, playerID, message, &request) {
		return
	}

	if request.CardID == "" {
		h.logger.Warn("Play card request missing card ID", zap.String("player_id", playerID))
		h.errorHandler.SendError(connection, "Card ID is required")
		return
	}

	if err := h.cardService.PlayCard(ctx, gameID, playerID, request.CardID); err != nil {
		h.logAndHandleError(connection, playerID, gameID, "play_card", err)
		return
	}
	h.logSuccess(connection, playerID, gameID, "play_card", zap.String("card_id", request.CardID))
}

func (h *PlayerHandler) handleSelectStartingCards(ctx context.Context, gameID, playerID string, connection *core.Connection, message dto.WebSocketMessage) {
	h.logger.Debug("🃏 Processing select starting cards action", zap.String("player_id", playerID))

	var request dto.ActionSelectStartingCardRequest
	if !h.parsePayload(connection, playerID, message, &request) {
		return
	}

	if err := h.processStartingCardSelection(ctx, gameID, playerID, request.CardIDs); err != nil {
		h.logAndHandleError(connection, playerID, gameID, "select_starting_cards", err)
		return
	}

	h.logSuccess(connection, playerID, gameID, "select_starting_cards",
		zap.Strings("card_ids", request.CardIDs),
		zap.Int("count", len(request.CardIDs)))
}

func (h *PlayerHandler) handleSelectProductionCards(ctx context.Context, gameID, playerID string, connection *core.Connection, message dto.WebSocketMessage) {
	h.logger.Debug("🃏 Processing select production cards action", zap.String("player_id", playerID))

	var request dto.ActionSelectProductionCardsRequest
	if !h.parsePayload(connection, playerID, message, &request) {
		return
	}

	if err := h.processProductionCardSelection(ctx, gameID, playerID, request.CardIDs); err != nil {
		h.logAndHandleError(connection, playerID, gameID, "select_production_cards", err)
		return
	}

	h.logSuccess(connection, playerID, gameID, "select_production_cards",
		zap.Strings("card_ids", request.CardIDs),
		zap.Int("count", len(request.CardIDs)))
}

// Turn Actions
func (h *PlayerHandler) handleSkipAction(ctx context.Context, gameID, playerID string, connection *core.Connection) {
	h.logger.Debug("⏭️ Processing skip action", zap.String("player_id", playerID))

	if err := h.processSkipAction(ctx, gameID, playerID); err != nil {
		h.logAndHandleError(connection, playerID, gameID, "skip_action", err)
		return
	}
	h.logSuccess(connection, playerID, gameID, "skip_action")
}

// Card selection logic
func (h *PlayerHandler) processStartingCardSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player selecting starting cards",
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)))

	if err := h.cardService.SelectStartingCards(ctx, gameID, playerID, cardIDs); err != nil {
		log.Error("Failed to select starting cards", zap.Error(err))
		return fmt.Errorf("card selection failed: %w", err)
	}

	if h.cardService.IsAllPlayersCardSelectionComplete(ctx, gameID) {
		log.Info("All players completed starting card selection, advancing game phase")

		if err := h.gameService.AdvanceFromCardSelectionPhase(ctx, gameID); err != nil {
			log.Error("Failed to advance game phase", zap.Error(err))
			return fmt.Errorf("failed to advance game phase: %w", err)
		}

		log.Info("Game phase advanced to Action phase")
	}

	log.Info("Player completed starting card selection", zap.Strings("selected_cards", cardIDs))
	return nil
}

func (h *PlayerHandler) processProductionCardSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Player selecting production cards",
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)))

	if err := h.cardService.SelectProductionCards(ctx, gameID, playerID, cardIDs); err != nil {
		log.Error("Failed to select production cards", zap.Error(err))
		return fmt.Errorf("card selection failed: %w", err)
	}

	updatedGame, err := h.gameService.ProcessProductionPhaseReady(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to process production phase ready", zap.Error(err))
		return fmt.Errorf("failed to process production phase ready: %w", err)
	}

	log.Info("Player completed production card selection and marked as ready",
		zap.Strings("selected_cards", cardIDs),
		zap.String("game_phase", string(updatedGame.CurrentPhase)))

	return nil
}

// Skip action logic
func (h *PlayerHandler) processSkipAction(ctx context.Context, gameID, playerID string) error {
	if err := h.validateGamePhase(ctx, gameID); err != nil {
		return err
	}

	result, err := h.gameService.SkipPlayerTurn(ctx, gameID, playerID)
	if err != nil {
		return err
	}

	if result.AllPlayersPassed {
		return h.handleProductionPhase(ctx, gameID)
	}

	return nil
}

func (h *PlayerHandler) validateGamePhase(ctx context.Context, gameID string) error {
	game, err := h.gameService.GetGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state for phase validation: %w", err)
	}

	if game.CurrentPhase != model.GamePhaseAction {
		return fmt.Errorf("skip action not allowed in phase %s, must be in action phase", game.CurrentPhase)
	}

	return nil
}

func (h *PlayerHandler) handleProductionPhase(ctx context.Context, gameID string) error {
	gameAfterProduction, err := h.gameService.ExecuteProductionPhase(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to execute production phase: %w", err)
	}

	playersData, err := h.generateProductionData(ctx, gameID, gameAfterProduction.PlayerIDs)
	if err != nil {
		return err
	}

	h.broadcaster.BroadcastProductionPhaseStarted(ctx, gameID, playersData)
	return nil
}

func (h *PlayerHandler) generateProductionData(ctx context.Context, gameID string, playerIDs []string) ([]dto.PlayerProductionData, error) {
	playersData := make([]dto.PlayerProductionData, len(playerIDs))

	for idx, playerID := range playerIDs {
		data, err := h.computePlayerProduction(ctx, gameID, playerID)
		if err != nil {
			return nil, fmt.Errorf("failed to compute production for player %s: %w", playerID, err)
		}
		playersData[idx] = data
	}

	return playersData, nil
}

func (h *PlayerHandler) computePlayerProduction(ctx context.Context, gameID, playerID string) (dto.PlayerProductionData, error) {
	player, err := h.playerService.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return dto.PlayerProductionData{}, fmt.Errorf("failed to get player state for production computation: %w", err)
	}

	beforeResources := h.calculateBeforeResources(player)
	afterResources := h.convertResources(player.Resources)
	resourceDelta := h.calculateResourceDelta(beforeResources, afterResources)
	production := h.convertProduction(player.Production)
	creditsIncome := player.Production.Credits + player.TerraformRating

	return dto.PlayerProductionData{
		PlayerID:        player.ID,
		PlayerName:      player.Name,
		BeforeResources: beforeResources,
		AfterResources:  afterResources,
		ResourceDelta:   resourceDelta,
		Production:      production,
		TerraformRating: player.TerraformRating,
		EnergyConverted: player.Production.Energy,
		CreditsIncome:   creditsIncome,
	}, nil
}

func (h *PlayerHandler) calculateBeforeResources(player model.Player) dto.ResourcesDto {
	return dto.ResourcesDto{
		Credits:  player.Resources.Credits - player.Production.Credits - player.TerraformRating,
		Steel:    player.Resources.Steel - player.Production.Steel,
		Titanium: player.Resources.Titanium - player.Production.Titanium,
		Plants:   player.Resources.Plants - player.Production.Plants,
		Energy:   player.Production.Energy, // Energy before was the old energy that got converted to heat
		Heat:     player.Resources.Heat - player.Production.Heat - player.Production.Energy,
	}
}

func (h *PlayerHandler) convertResources(resources model.Resources) dto.ResourcesDto {
	return dto.ResourcesDto{
		Credits:  resources.Credits,
		Steel:    resources.Steel,
		Titanium: resources.Titanium,
		Plants:   resources.Plants,
		Energy:   resources.Energy,
		Heat:     resources.Heat,
	}
}

func (h *PlayerHandler) convertProduction(production model.Production) dto.ProductionDto {
	return dto.ProductionDto{
		Credits:  production.Credits,
		Steel:    production.Steel,
		Titanium: production.Titanium,
		Plants:   production.Plants,
		Energy:   production.Energy,
		Heat:     production.Heat,
	}
}

func (h *PlayerHandler) calculateResourceDelta(before, after dto.ResourcesDto) dto.ResourceDelta {
	return dto.ResourceDelta{
		Credits:  after.Credits - before.Credits,
		Steel:    after.Steel - before.Steel,
		Titanium: after.Titanium - before.Titanium,
		Plants:   after.Plants - before.Plants,
		Energy:   after.Energy - before.Energy,
		Heat:     after.Heat - before.Heat,
	}
}