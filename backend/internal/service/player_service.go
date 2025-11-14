package service

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// PlayerService handles player-specific operations
type PlayerService interface {

	// Get player information
	GetPlayer(ctx context.Context, gameID, playerID string) (model.Player, error)
	GetPlayerByName(ctx context.Context, gameID, playerName string) (model.Player, error)
	GetPlayersForGame(ctx context.Context, gameID string) ([]model.Player, error)

	// Handle player disconnection - updates connection status and broadcasts game state
	PlayerDisconnected(ctx context.Context, gameID, playerID string) error

	// Standard project utility methods
	CanAffordStandardProject(player *model.Player, project model.StandardProject) bool
	HasCardsToSell(player *model.Player, count int) bool
	GetMaxCardsToSell(player *model.Player) int

	// Tile selection methods
	OnTileSelected(ctx context.Context, gameID, playerID string, coordinate model.HexPosition) error
}

// PlayerServiceImpl implements PlayerService interface
type PlayerServiceImpl struct {
	gameRepo            game.Repository
	playerRepo          player.Repository
	sessionManager      session.SessionManager
	tilesMech           tiles.Service
	parametersMech      parameters.Service
	forcedActionManager cards.ForcedActionManager
}

// NewPlayerService creates a new PlayerService instance
func NewPlayerService(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionManager session.SessionManager,
	tilesMech tiles.Service,
	parametersMech parameters.Service,
	forcedActionManager cards.ForcedActionManager,
) PlayerService {
	return &PlayerServiceImpl{
		gameRepo:            gameRepo,
		playerRepo:          playerRepo,
		sessionManager:      sessionManager,
		tilesMech:           tilesMech,
		parametersMech:      parametersMech,
		forcedActionManager: forcedActionManager,
	}
}

// GetPlayer retrieves player information
func (s *PlayerServiceImpl) GetPlayer(ctx context.Context, gameID, playerID string) (model.Player, error) {
	return s.playerRepo.GetByID(ctx, gameID, playerID)
}

// CanAffordStandardProject checks if the player has enough credits for a standard project (business logic from Player model)
func (s *PlayerServiceImpl) CanAffordStandardProject(player *model.Player, project model.StandardProject) bool {
	cost, exists := model.StandardProjectCost[project]
	if !exists {
		return false
	}
	return player.Resources.Credits >= cost
}

// HasCardsToSell checks if the player has enough cards in hand to sell (business logic from Player model)
func (s *PlayerServiceImpl) HasCardsToSell(player *model.Player, count int) bool {
	return len(player.Cards) >= count && count > 0
}

// GetMaxCardsToSell returns the maximum number of cards the player can sell (business logic from Player model)
func (s *PlayerServiceImpl) GetMaxCardsToSell(player *model.Player) int {
	return len(player.Cards)
}

// UpdatePlayerConnectionStatus updates a player's connection status
func (s *PlayerServiceImpl) updatePlayerConnectionStatus(ctx context.Context, gameID, playerID string, isConnected bool) error {
	log := logger.WithGameContext(gameID, playerID)

	// Update connection status using granular method
	err := s.playerRepo.UpdateConnectionStatus(ctx, gameID, playerID, isConnected)
	if err != nil {
		log.Error("Failed to update player connection status", zap.Error(err))
		return fmt.Errorf("failed to update player: %w", err)
	}

	log.Info("Updated player connection status",
		zap.Bool("is_connected", isConnected))

	return nil
}

// PlayerDisconnected handles player disconnection by updating connection status and broadcasting game state
func (s *PlayerServiceImpl) PlayerDisconnected(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🔌 Processing player disconnection")

	// Update connection status to false
	err := s.updatePlayerConnectionStatus(ctx, gameID, playerID, false)
	if err != nil {
		log.Error("Failed to update connection status during disconnection", zap.Error(err))
		return fmt.Errorf("failed to update connection status: %w", err)
	}

	// Broadcast updated game state to other players (if SessionManager is available)
	if s.sessionManager != nil {
		err = s.sessionManager.Broadcast(gameID)
		if err != nil {
			log.Error("Failed to broadcast game state after player disconnection", zap.Error(err))
			return fmt.Errorf("failed to broadcast game state: %w", err)
		}
	} else {
		log.Warn("SessionManager not available, skipping broadcast")
	}

	log.Info("✅ Player disconnection processed successfully")
	return nil
}

// GetPlayerByName finds a player by name in a specific game
func (s *PlayerServiceImpl) GetPlayerByName(ctx context.Context, gameID, playerName string) (model.Player, error) {
	log := logger.WithGameContext(gameID, playerName)

	// Get all players from the player repository
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for name lookup", zap.Error(err))
		return model.Player{}, fmt.Errorf("failed to get players: %w", err)
	}

	// Search for player by name
	for _, player := range players {
		if player.Name == playerName {
			log.Debug("Found player by name",
				zap.String("player_id", player.ID),
				zap.String("player_name", player.Name))
			return player, nil
		}
	}

	log.Warn("Player not found by name", zap.String("player_name", playerName))
	return model.Player{}, fmt.Errorf("player with name %s not found in game %s", playerName, gameID)
}

// GetPlayersForGame returns all players in a specific game
func (s *PlayerServiceImpl) GetPlayersForGame(ctx context.Context, gameID string) ([]model.Player, error) {
	return s.playerRepo.ListByGameID(ctx, gameID)
}

// OnTileSelected handles player tile selection and placement
func (s *PlayerServiceImpl) OnTileSelected(ctx context.Context, gameID, playerID string, coordinate model.HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🎯 Processing tile selection",
		zap.Int("q", coordinate.Q),
		zap.Int("r", coordinate.R),
		zap.Int("s", coordinate.S))

	// Get player's pending tile selection to determine tile type
	pendingSelection, err := s.playerRepo.GetPendingTileSelection(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to get pending tile selection: %w", err)
	}

	if pendingSelection == nil {
		log.Warn("No pending tile selection found for player")
		return fmt.Errorf("player has no pending tile selection")
	}

	// Convert coordinate to string for validation
	coordinateKey := coordinate.String()

	// Basic validation that the clicked tile is in the available hexes
	validTile := false
	for _, hexID := range pendingSelection.AvailableHexes {
		if hexID == coordinateKey {
			validTile = true
			break
		}
	}

	if !validTile {
		log.Error("Invalid tile selection",
			zap.String("coordinate", coordinateKey),
			zap.Strings("available", pendingSelection.AvailableHexes))
		return fmt.Errorf("selected coordinate %s is not in available positions", coordinateKey)
	}

	// Check if this is a plant conversion (special handling for raising oxygen and TR)
	isPlantConversion := pendingSelection.Source == "convert-plants-to-greenery"

	// Convert model.HexPosition to tiles.HexPosition
	tileCoordinate := tiles.HexPosition{
		Q: coordinate.Q,
		R: coordinate.R,
		S: coordinate.S,
	}

	// Place the tile using tiles mechanic
	if err := s.tilesMech.PlaceTile(ctx, gameID, playerID, pendingSelection.TileType, tileCoordinate); err != nil {
		log.Error("Failed to place tile", zap.Error(err))
		return fmt.Errorf("failed to place tile: %w", err)
	}

	// Check if this tile placement was triggered by a forced action
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for forced action check", zap.Error(err))
	} else {
		isForcedAction := player.ForcedFirstAction != nil &&
			player.ForcedFirstAction.CorporationID == pendingSelection.Source

		if isForcedAction {
			if err := s.forcedActionManager.MarkComplete(ctx, gameID, playerID); err != nil {
				log.Error("Failed to mark forced action complete", zap.Error(err))
				// Don't fail the operation, just log the error
			} else {
				log.Info("🎯 Forced action marked as complete", zap.String("source", pendingSelection.Source))
			}
		}
	}

	// Handle plant conversion completion (raise oxygen and TR)
	if isPlantConversion {
		log.Info("🌱 Completing plant conversion - raising oxygen")

		// Raise oxygen using parameters mechanic (automatically awards TR if raised)
		newOxygen, err := s.parametersMech.RaiseOxygen(ctx, gameID, playerID, 1)
		if err != nil {
			log.Error("Failed to raise oxygen", zap.Error(err))
			return fmt.Errorf("failed to raise oxygen: %w", err)
		}

		log.Info("🌍 Oxygen raised", zap.Int("new_oxygen", newOxygen))
		log.Info("✅ Plant conversion completed successfully")
	}

	// Clear the current pending tile selection
	if err := s.playerRepo.ClearPendingTileSelection(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to clear pending tile selection: %w", err)
	}

	// Process the next tile in the queue using tiles mechanic
	if err := s.tilesMech.ProcessTileQueue(ctx, gameID, playerID); err != nil {
		log.Error("Failed to process next tile in queue", zap.Error(err))
		return fmt.Errorf("failed to process next tile in queue: %w", err)
	}

	log.Info("🎯 Tile placed and queue processed")

	// Broadcast updated game state
	if s.sessionManager != nil {
		if err := s.sessionManager.Broadcast(gameID); err != nil {
			log.Error("Failed to broadcast game state after tile selection", zap.Error(err))
			return fmt.Errorf("failed to broadcast game state: %w", err)
		}
	}

	log.Info("✅ Tile selection processed successfully",
		zap.String("coordinate", coordinateKey),
		zap.String("tile_type", pendingSelection.TileType))

	return nil
}
