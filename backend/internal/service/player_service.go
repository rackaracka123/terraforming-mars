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

// PlayerService handles player-specific operations
type PlayerService interface {

	// Get player information
	GetPlayer(ctx context.Context, gameID, playerID string) (model.Player, error)
	GetPlayerByName(ctx context.Context, gameID, playerName string) (model.Player, error)
	GetPlayersForGame(ctx context.Context, gameID string) ([]model.Player, error)

	// Handle player disconnection - updates connection status and broadcasts game state
	PlayerDisconnected(ctx context.Context, gameID, playerID string) error

	// Validation methods for card system
	ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement model.ResourceSet) error
	ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error

	// Card effect methods
	AddProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error
	PayResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error
	AddResources(ctx context.Context, gameID, playerID string, resources model.ResourceSet) error

	// Standard project utility methods
	CanAffordStandardProject(player *model.Player, project model.StandardProject) bool
	HasCardsToSell(player *model.Player, count int) bool
	GetMaxCardsToSell(player *model.Player) int

	// Tile selection methods
	OnTileSelected(ctx context.Context, gameID, playerID string, coordinate model.HexPosition) error
}

// PlayerServiceImpl implements PlayerService interface
type PlayerServiceImpl struct {
	gameRepo       repository.GameRepository
	playerRepo     repository.PlayerRepository
	sessionManager session.SessionManager
}

// NewPlayerService creates a new PlayerService instance
func NewPlayerService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, sessionManager session.SessionManager) PlayerService {
	return &PlayerServiceImpl{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		sessionManager: sessionManager,
	}
}

// GetPlayer retrieves player information
func (s *PlayerServiceImpl) GetPlayer(ctx context.Context, gameID, playerID string) (model.Player, error) {
	return s.playerRepo.GetByID(ctx, gameID, playerID)
}

// ValidateProductionRequirement validates if player meets production requirements
func (s *PlayerServiceImpl) ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement model.ResourceSet) error {
	player, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player has sufficient production
	if player.Production.Credits < requirement.Credits ||
		player.Production.Steel < requirement.Steel ||
		player.Production.Titanium < requirement.Titanium ||
		player.Production.Plants < requirement.Plants ||
		player.Production.Energy < requirement.Energy ||
		player.Production.Heat < requirement.Heat {
		return fmt.Errorf("insufficient production to meet requirement")
	}

	return nil
}

// ValidateResourceCost validates if player can afford the resource cost
func (s *PlayerServiceImpl) ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	player, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player has sufficient resources
	if player.Resources.Credits < cost.Credits ||
		player.Resources.Steel < cost.Steel ||
		player.Resources.Titanium < cost.Titanium ||
		player.Resources.Plants < cost.Plants ||
		player.Resources.Energy < cost.Energy ||
		player.Resources.Heat < cost.Heat {
		return fmt.Errorf("insufficient resources to pay cost")
	}

	return nil
}

// AddProduction adds production to a player
func (s *PlayerServiceImpl) AddProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error {
	player, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Add production
	newProduction := model.Production{
		Credits:  player.Production.Credits + production.Credits,
		Steel:    player.Production.Steel + production.Steel,
		Titanium: player.Production.Titanium + production.Titanium,
		Plants:   player.Production.Plants + production.Plants,
		Energy:   player.Production.Energy + production.Energy,
		Heat:     player.Production.Heat + production.Heat,
	}

	// Update production directly via repository
	if err := s.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
		return fmt.Errorf("failed to update player production: %w", err)
	}
	return nil
}

// PayResourceCost deducts resource cost from player
func (s *PlayerServiceImpl) PayResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error {
	player, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Validate player can afford the cost
	if err := s.ValidateResourceCost(ctx, gameID, playerID, cost); err != nil {
		return err
	}

	// Deduct resources
	newResources := model.Resources{
		Credits:  player.Resources.Credits - cost.Credits,
		Steel:    player.Resources.Steel - cost.Steel,
		Titanium: player.Resources.Titanium - cost.Titanium,
		Plants:   player.Resources.Plants - cost.Plants,
		Energy:   player.Resources.Energy - cost.Energy,
		Heat:     player.Resources.Heat - cost.Heat,
	}

	// Update resources directly via repository
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
		return fmt.Errorf("failed to update player resources: %w", err)
	}
	return nil
}

// AddResources adds resources to a player
func (s *PlayerServiceImpl) AddResources(ctx context.Context, gameID, playerID string, resources model.ResourceSet) error {
	player, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Add resources
	newResources := model.Resources{
		Credits:  player.Resources.Credits + resources.Credits,
		Steel:    player.Resources.Steel + resources.Steel,
		Titanium: player.Resources.Titanium + resources.Titanium,
		Plants:   player.Resources.Plants + resources.Plants,
		Energy:   player.Resources.Energy + resources.Energy,
		Heat:     player.Resources.Heat + resources.Heat,
	}

	// Update resources directly via repository
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
		return fmt.Errorf("failed to update player resources: %w", err)
	}
	return nil
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

	// Convert coordinate to string for validation (temporary until we update the validation system)
	coordinateKey := fmt.Sprintf("%d,%d,%d", coordinate.Q, coordinate.R, coordinate.S)

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

	// Place the tile using the private method
	if err := s.placeTile(ctx, gameID, playerID, pendingSelection.TileType, coordinate); err != nil {
		log.Error("Failed to place tile", zap.Error(err))
		return fmt.Errorf("failed to place tile: %w", err)
	}

	// For admin demo selections, don't clear the pending selection to allow continuous testing
	// For real game selections, clear the pending tile selection to complete the basic flow
	if err := s.playerRepo.ClearPendingTileSelection(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to clear pending tile selection: %w", err)
	}
	log.Info("🎯 Real game tile selection - cleared pending selection")

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

// placeTile places a tile of the specified type at the given coordinate
func (s *PlayerServiceImpl) placeTile(ctx context.Context, gameID, playerID, tileType string, coordinate model.HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🔧 Placing tile",
		zap.String("tile_type", tileType),
		zap.Int("q", coordinate.Q),
		zap.Int("r", coordinate.R),
		zap.Int("s", coordinate.S))

	// Create the tile occupant based on the tile type
	var occupant *model.TileOccupant
	switch tileType {
	case "city":
		occupant = &model.TileOccupant{
			Type: model.ResourceCityTile,
			Tags: []string{},
		}
	case "greenery":
		occupant = &model.TileOccupant{
			Type: model.ResourceGreeneryTile,
			Tags: []string{},
		}
	case "ocean":
		occupant = &model.TileOccupant{
			Type: model.ResourceOceanTile,
			Tags: []string{},
		}
	default:
		log.Error("Unknown tile type", zap.String("tile_type", tileType))
		return fmt.Errorf("unknown tile type: %s", tileType)
	}

	// Update the tile occupancy in the game board
	if err := s.gameRepo.UpdateTileOccupancy(ctx, gameID, coordinate, occupant, &playerID); err != nil {
		log.Error("Failed to update tile occupancy", zap.Error(err))
		return fmt.Errorf("failed to update tile occupancy: %w", err)
	}

	log.Info("✅ Tile placed successfully",
		zap.String("tile_type", tileType),
		zap.String("coordinate", fmt.Sprintf("%d,%d,%d", coordinate.Q, coordinate.R, coordinate.S)))

	return nil
}
