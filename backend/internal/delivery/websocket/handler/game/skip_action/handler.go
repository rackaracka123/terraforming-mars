package skip_action

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

// Handler handles skip action requests
type Handler struct {
	gameService   service.GameService
	playerService service.PlayerService
	broadcaster   *core.Broadcaster
	errorHandler  *utils.ErrorHandler
	logger        *zap.Logger
}

// NewHandler creates a new skip action handler
func NewHandler(gameService service.GameService, playerService service.PlayerService, broadcaster *core.Broadcaster) *Handler {
	return &Handler{
		gameService:   gameService,
		playerService: playerService,
		broadcaster:   broadcaster,
		errorHandler:  utils.NewErrorHandler(),
		logger:        logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Skip action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("⏭️ Processing skip action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Skip action doesn't need payload parsing - it's a simple action
	if err := h.handle(ctx, gameID, playerID); err != nil {
		h.logger.Error("Failed to skip action",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	h.logger.Info("✅ Skip action completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))
}

// handle processes the skip action (internal method)
func (h *Handler) handle(ctx context.Context, gameID, playerID string) error {
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

// validateGamePhase ensures the game is in the correct phase for skipping
func (h *Handler) validateGamePhase(ctx context.Context, gameID string) error {
	game, err := h.gameService.GetGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state for phase validation: %w", err)
	}

	if game.CurrentPhase != model.GamePhaseAction {
		return fmt.Errorf("skip action not allowed in phase %s, must be in action phase", game.CurrentPhase)
	}

	return nil
}

// handleProductionPhase executes production phase when all players have passed
func (h *Handler) handleProductionPhase(ctx context.Context, gameID string) error {
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

// generateProductionData computes production data for all players
func (h *Handler) generateProductionData(ctx context.Context, gameID string, playerIDs []string) ([]dto.PlayerProductionData, error) {
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

// computePlayerProduction calculates production data for a single player
func (h *Handler) computePlayerProduction(ctx context.Context, gameID, playerID string) (dto.PlayerProductionData, error) {
	player, err := h.playerService.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return dto.PlayerProductionData{}, fmt.Errorf("failed to get player state for production computation: %w", err)
	}

	beforeResources := h.calculateBeforeResources(player)
	afterResources := h.convertResources(player.Resources)
	production := h.convertProduction(player.Production)
	creditsIncome := player.Production.Credits + player.TerraformRating

	return dto.PlayerProductionData{
		PlayerID:        player.ID,
		PlayerName:      player.Name,
		BeforeResources: beforeResources,
		AfterResources:  afterResources,
		Production:      production,
		TerraformRating: player.TerraformRating,
		EnergyConverted: player.Production.Energy,
		CreditsIncome:   creditsIncome,
	}, nil
}

// calculateBeforeResources computes resources before production was applied
func (h *Handler) calculateBeforeResources(player model.Player) dto.ResourcesDto {
	return dto.ResourcesDto{
		Credits:  player.Resources.Credits - player.Production.Credits - player.TerraformRating,
		Steel:    player.Resources.Steel - player.Production.Steel,
		Titanium: player.Resources.Titanium - player.Production.Titanium,
		Plants:   player.Resources.Plants - player.Production.Plants,
		Energy:   player.Production.Energy, // Energy before was the old energy that got converted to heat
		Heat:     player.Resources.Heat - player.Production.Heat - player.Production.Energy,
	}
}

// convertResources converts model resources to DTO
func (h *Handler) convertResources(resources model.Resources) dto.ResourcesDto {
	return dto.ResourcesDto{
		Credits:  resources.Credits,
		Steel:    resources.Steel,
		Titanium: resources.Titanium,
		Plants:   resources.Plants,
		Energy:   resources.Energy,
		Heat:     resources.Heat,
	}
}

// convertProduction converts model production to DTO
func (h *Handler) convertProduction(production model.Production) dto.ProductionDto {
	return dto.ProductionDto{
		Credits:  production.Credits,
		Steel:    production.Steel,
		Titanium: production.Titanium,
		Plants:   production.Plants,
		Energy:   production.Energy,
		Heat:     production.Heat,
	}
}
