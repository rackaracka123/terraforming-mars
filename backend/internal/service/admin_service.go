package service

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// AdminService handles administrative operations for development mode
type AdminService interface {
	// OnAdminGiveCard gives a specific card to a player
	OnAdminGiveCard(ctx context.Context, gameID, playerID, cardID string) error

	// OnAdminSetPhase sets the game phase
	OnAdminSetPhase(ctx context.Context, gameID string, phase model.GamePhase) error

	// OnAdminSetResources sets a player's resources
	OnAdminSetResources(ctx context.Context, gameID, playerID string, resources model.Resources) error

	// OnAdminSetProduction sets a player's production
	OnAdminSetProduction(ctx context.Context, gameID, playerID string, production model.Production) error

	// OnAdminSetGlobalParameters sets the global terraforming parameters
	OnAdminSetGlobalParameters(ctx context.Context, gameID string, params model.GlobalParameters) error

	// OnAdminStartTileSelection starts tile selection for testing
	OnAdminStartTileSelection(ctx context.Context, gameID, playerID, tileType string) error

	// OnAdminSetCurrentTurn sets the current player turn
	OnAdminSetCurrentTurn(ctx context.Context, gameID, playerID string) error
}

// AdminServiceImpl implements AdminService interface
type AdminServiceImpl struct {
	gameRepo       repository.GameRepository
	playerRepo     repository.PlayerRepository
	cardRepo       repository.CardRepository
	sessionManager session.SessionManager
}

// NewAdminService creates a new AdminService instance
func NewAdminService(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardRepo repository.CardRepository,
	sessionManager session.SessionManager,
) AdminService {
	return &AdminServiceImpl{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		cardRepo:       cardRepo,
		sessionManager: sessionManager,
	}
}

// OnAdminGiveCard gives a specific card to a player
func (s *AdminServiceImpl) OnAdminGiveCard(ctx context.Context, gameID, playerID, cardID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🎴 Admin giving card to player", zap.String("card_id", cardID))

	// Verify card exists
	card, err := s.cardRepo.GetCardByID(ctx, cardID)
	if err != nil || card == nil {
		log.Error("Card not found", zap.String("card_id", cardID), zap.Error(err))
		return fmt.Errorf("card not found: %s", cardID)
	}

	// Verify player exists
	_, err = s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// Add card to player's hand using repository method
	if err := s.playerRepo.AddCard(ctx, gameID, playerID, cardID); err != nil {
		log.Error("Failed to add card to player hand", zap.Error(err))
		return fmt.Errorf("failed to add card to player hand: %w", err)
	}

	log.Info("✅ Card given to player successfully",
		zap.String("card_id", cardID),
		zap.String("card_name", card.Name))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after giving card", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminSetPhase sets the game phase
func (s *AdminServiceImpl) OnAdminSetPhase(ctx context.Context, gameID string, phase model.GamePhase) error {
	log := logger.WithGameContext(gameID, "")
	log.Info("🔄 Admin setting game phase", zap.String("phase", string(phase)))

	// Verify game exists
	_, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %w", err)
	}

	// Update game phase
	if err := s.gameRepo.UpdatePhase(ctx, gameID, phase); err != nil {
		log.Error("Failed to update game phase", zap.Error(err))
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	log.Info("✅ Game phase set successfully", zap.String("phase", string(phase)))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after phase change", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminSetResources sets a player's resources
func (s *AdminServiceImpl) OnAdminSetResources(ctx context.Context, gameID, playerID string, resources model.Resources) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("💰 Admin setting player resources", zap.Any("resources", resources))

	// Verify player exists
	_, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// Update player resources
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, resources); err != nil {
		log.Error("Failed to update player resources", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	log.Info("✅ Player resources set successfully",
		zap.Int("credits", resources.Credits),
		zap.Int("steel", resources.Steel),
		zap.Int("titanium", resources.Titanium),
		zap.Int("plants", resources.Plants),
		zap.Int("energy", resources.Energy),
		zap.Int("heat", resources.Heat))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after resources change", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminSetProduction sets a player's production
func (s *AdminServiceImpl) OnAdminSetProduction(ctx context.Context, gameID, playerID string, production model.Production) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🏭 Admin setting player production", zap.Any("production", production))

	// Verify player exists
	_, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// Update player production
	if err := s.playerRepo.UpdateProduction(ctx, gameID, playerID, production); err != nil {
		log.Error("Failed to update player production", zap.Error(err))
		return fmt.Errorf("failed to update player production: %w", err)
	}

	log.Info("✅ Player production set successfully",
		zap.Int("credits", production.Credits),
		zap.Int("steel", production.Steel),
		zap.Int("titanium", production.Titanium),
		zap.Int("plants", production.Plants),
		zap.Int("energy", production.Energy),
		zap.Int("heat", production.Heat))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after production change", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminSetGlobalParameters sets the global terraforming parameters
func (s *AdminServiceImpl) OnAdminSetGlobalParameters(ctx context.Context, gameID string, params model.GlobalParameters) error {
	log := logger.WithGameContext(gameID, "")
	log.Info("🌍 Admin setting global parameters",
		zap.Int("temperature", params.Temperature),
		zap.Int("oxygen", params.Oxygen),
		zap.Int("oceans", params.Oceans))

	// Verify game exists
	_, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %w", err)
	}

	// Validate parameters are within game bounds
	if params.Temperature < -30 || params.Temperature > 8 {
		return fmt.Errorf("temperature must be between -30 and 8, got %d", params.Temperature)
	}
	if params.Oxygen < 0 || params.Oxygen > 14 {
		return fmt.Errorf("oxygen must be between 0 and 14, got %d", params.Oxygen)
	}
	if params.Oceans < 0 || params.Oceans > 9 {
		return fmt.Errorf("oceans must be between 0 and 9, got %d", params.Oceans)
	}

	// Update global parameters
	if err := s.gameRepo.UpdateGlobalParameters(ctx, gameID, params); err != nil {
		log.Error("Failed to update global parameters", zap.Error(err))
		return fmt.Errorf("failed to update global parameters: %w", err)
	}

	log.Info("✅ Global parameters set successfully",
		zap.Int("temperature", params.Temperature),
		zap.Int("oxygen", params.Oxygen),
		zap.Int("oceans", params.Oceans))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after global parameters change", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminStartTileSelection starts tile selection for testing
func (s *AdminServiceImpl) OnAdminStartTileSelection(ctx context.Context, gameID, playerID, tileType string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🎯 Admin starting tile selection", zap.String("tile_type", tileType))

	// Verify player exists
	_, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// Get game to access the board
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %w", err)
	}

	// Calculate available hexes based on tile type (demo logic)
	availableHexes := s.calculateDemoAvailableHexes(game, tileType)

	if len(availableHexes) == 0 {
		log.Warn("No valid positions available for tile type", zap.String("tile_type", tileType))
		return fmt.Errorf("no valid positions available for %s placement", tileType)
	}

	// Set pending tile selection
	pendingSelection := &model.PendingTileSelection{
		TileType:       tileType,
		AvailableHexes: availableHexes,
		Source:         "admin_demo",
	}

	if err := s.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, pendingSelection); err != nil {
		log.Error("Failed to set pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to set pending tile selection: %w", err)
	}

	log.Info("✅ Tile selection started successfully",
		zap.String("tile_type", tileType),
		zap.Int("available_positions", len(availableHexes)))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after starting tile selection", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminSetCurrentTurn sets the current player turn
func (s *AdminServiceImpl) OnAdminSetCurrentTurn(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🔄 Admin setting current turn")

	// Verify player exists
	_, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// Set current turn
	if err := s.gameRepo.SetCurrentTurn(ctx, gameID, &playerID); err != nil {
		log.Error("Failed to set current turn", zap.Error(err))
		return fmt.Errorf("failed to set current turn: %w", err)
	}

	log.Info("✅ Current turn set successfully")

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after setting current turn", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// calculateDemoAvailableHexes calculates available positions for demo tile placement
func (s *AdminServiceImpl) calculateDemoAvailableHexes(game model.Game, tileType string) []string {
	var availableHexes []string

	// Demo logic: just find empty tiles based on type
	for _, tile := range game.Board.Tiles {
		// Tile must be empty
		if tile.OccupiedBy != nil {
			continue
		}

		switch tileType {
		case "ocean":
			// Ocean tiles can only be placed on ocean-designated spaces
			if tile.Type == model.ResourceOceanTile {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}

		case "city", "greenery":
			// Cities and greenery can be placed on any empty land space
			if tile.Type != model.ResourceOceanTile {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}
		}
	}

	return availableHexes
}
