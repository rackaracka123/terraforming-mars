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

	// ProcessTileQueue processes the tile queue, validating and setting up the first valid tile selection
	// This should be called after any operation that creates a tile queue (e.g., card play)
	// Returns nil if queue is empty or doesn't exist
	ProcessTileQueue(ctx context.Context, gameID, playerID string) error
}

// PlayerServiceImpl implements PlayerService interface
type PlayerServiceImpl struct {
	gameRepo       repository.GameRepository
	playerRepo     repository.PlayerRepository
	sessionManager session.SessionManager
	boardService   BoardService
}

// NewPlayerService creates a new PlayerService instance
func NewPlayerService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, sessionManager session.SessionManager, boardService BoardService) PlayerService {
	return &PlayerServiceImpl{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		sessionManager: sessionManager,
		boardService:   boardService,
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

// validateTilePlacement checks if a tile type can still be placed in the game
func (s *PlayerServiceImpl) validateTilePlacement(ctx context.Context, gameID, tileType string) (bool, error) {
	log := logger.WithGameContext(gameID, "")

	// Get game state to check global parameters
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return false, fmt.Errorf("failed to get game state: %w", err)
	}

	switch tileType {
	case "ocean":
		// Check if we've reached the maximum of 9 ocean tiles
		// Count existing ocean tiles on the board
		oceanCount := 0
		for _, tile := range game.Board.Tiles {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type == model.ResourceOceanTile {
				oceanCount++
			}
		}

		canPlace := oceanCount < 9
		log.Debug("Ocean tile validation",
			zap.String("tile_type", tileType),
			zap.Int("current_oceans", oceanCount),
			zap.Bool("can_place", canPlace))
		return canPlace, nil

	case "city", "greenery":
		// Cities and greenery can generally always be placed if there are available spaces
		// More complex validation (adjacency rules, etc.) should be handled elsewhere
		return true, nil

	default:
		// Unknown tile types are considered valid for now
		log.Warn("Unknown tile type for validation", zap.String("tile_type", tileType))
		return true, nil
	}
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

	// Validate that this tile type can still be placed (business logic validation)
	canPlace, err := s.validateTilePlacement(ctx, gameID, pendingSelection.TileType)
	if err != nil {
		log.Error("Failed to validate tile placement", zap.Error(err))
		return fmt.Errorf("failed to validate tile placement: %w", err)
	}

	if !canPlace {
		log.Error("Tile placement no longer valid",
			zap.String("tile_type", pendingSelection.TileType),
			zap.String("coordinate", coordinateKey))
		return fmt.Errorf("cannot place %s tile - game constraints not met", pendingSelection.TileType)
	}

	// Place the tile using the private method
	if err := s.placeTile(ctx, gameID, playerID, pendingSelection.TileType, coordinate); err != nil {
		log.Error("Failed to place tile", zap.Error(err))
		return fmt.Errorf("failed to place tile: %w", err)
	}

	// Clear the current pending tile selection
	if err := s.playerRepo.ClearPendingTileSelection(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to clear pending tile selection: %w", err)
	}

	// Process the next tile in the queue with validation
	if err := s.processNextTileInQueueWithValidation(ctx, gameID, playerID); err != nil {
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

	// Award placement bonuses to the player
	if err := s.awardTilePlacementBonuses(ctx, gameID, playerID, coordinate); err != nil {
		log.Error("Failed to award tile placement bonuses", zap.Error(err))
		return fmt.Errorf("failed to award tile placement bonuses: %w", err)
	}

	log.Info("✅ Tile placed successfully",
		zap.String("tile_type", tileType),
		zap.String("coordinate", fmt.Sprintf("%d,%d,%d", coordinate.Q, coordinate.R, coordinate.S)))

	return nil
}

// processNextTileInQueueWithValidation processes the next tile in queue with business logic validation
func (s *PlayerServiceImpl) processNextTileInQueueWithValidation(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get the queue to extract the source (card ID)
	queue, err := s.playerRepo.GetPendingTileSelectionQueue(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get tile queue: %w", err)
	}

	// If no queue, we're done
	if queue == nil {
		log.Debug("No tile queue exists")
		return nil
	}

	source := queue.Source // Store the source card ID

	for {
		// Pop the next tile type from repository (pure data operation)
		nextTileType, err := s.playerRepo.ProcessNextTileInQueue(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to pop next tile from queue: %w", err)
		}

		// If no tile type returned, we're done
		if nextTileType == "" {
			log.Debug("No more tiles in queue")
			return nil
		}

		log.Info("🎯 Validating next tile from queue",
			zap.String("tile_type", nextTileType),
			zap.String("source", source))

		// Validate this tile placement is still possible (especially important for oceans)
		canPlace, err := s.validateTilePlacement(ctx, gameID, nextTileType)
		if err != nil {
			return fmt.Errorf("failed to validate tile placement: %w", err)
		}

		log.Info("🏙️ Tile placement validation result",
			zap.String("tile_type", nextTileType),
			zap.Bool("can_place", canPlace))

		if canPlace {
			// Tile is valid, now calculate available hexes for this tile type
			// For greenery, use player-specific calculation to enforce adjacency rule
			availableHexes, err := s.calculateAvailableHexesForTileTypeWithPlayer(ctx, gameID, playerID, nextTileType)
			if err != nil {
				return fmt.Errorf("failed to calculate available hexes: %w", err)
			}

			// Create and set the pending tile selection with available hexes
			selection := &model.PendingTileSelection{
				TileType:       nextTileType,
				AvailableHexes: availableHexes,
				Source:         source,
			}

			if err := s.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, selection); err != nil {
				return fmt.Errorf("failed to set pending tile selection: %w", err)
			}

			log.Info("🎯 Tile validation successful and available hexes calculated",
				zap.String("tile_type", nextTileType),
				zap.Int("available_hexes", len(availableHexes)))
			return nil
		}

		// Tile is no longer valid, skip it and try next
		log.Info("⚠️ Tile placement no longer possible, skipping and checking next",
			zap.String("tile_type", nextTileType))

		// Continue loop to pop and process next tile
	}
}

// ProcessTileQueue processes the tile queue, validating and setting up the first valid tile selection
// This is the public API method that should be called after card effects that create tile queues
func (s *PlayerServiceImpl) ProcessTileQueue(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🎯 Processing tile queue")

	// Process the queue through the private validation method
	return s.processNextTileInQueueWithValidation(ctx, gameID, playerID)
}

// calculateAvailableHexesForTileType returns available hexes for a specific tile type
func (s *PlayerServiceImpl) calculateAvailableHexesForTileType(ctx context.Context, gameID, tileType string) ([]string, error) {
	log := logger.WithGameContext(gameID, "")

	log.Info("🏙️ Starting available hexes calculation",
		zap.String("tile_type", tileType))

	// Get the current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for hex calculation", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	log.Info("🏙️ Got game state, delegating to BoardService",
		zap.String("tile_type", tileType))

	// Delegate to BoardService for hex calculation
	availableHexes, err := s.boardService.CalculateAvailableHexesForTileType(game, tileType)
	if err != nil {
		log.Error("🏙️ BoardService calculation failed", zap.Error(err))
		return nil, err
	}

	log.Info("🏙️ BoardService calculation completed",
		zap.String("tile_type", tileType),
		zap.Int("available_count", len(availableHexes)))

	return availableHexes, nil
}

// calculateAvailableHexesForTileTypeWithPlayer returns available hexes with player context
// Used for greenery placement which requires adjacency to player's tiles
func (s *PlayerServiceImpl) calculateAvailableHexesForTileTypeWithPlayer(ctx context.Context, gameID, playerID, tileType string) ([]string, error) {
	log := logger.WithGameContext(gameID, playerID)

	log.Info("🏙️ Starting available hexes calculation with player context",
		zap.String("tile_type", tileType))

	// Get the current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for hex calculation", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Delegate to BoardService with player context
	availableHexes, err := s.boardService.CalculateAvailableHexesForTileTypeWithPlayer(game, tileType, playerID)
	if err != nil {
		log.Error("🏙️ BoardService calculation failed", zap.Error(err))
		return nil, err
	}

	log.Info("🏙️ BoardService calculation completed",
		zap.String("tile_type", tileType),
		zap.Int("available_count", len(availableHexes)))

	return availableHexes, nil
}

// awardTilePlacementBonuses awards bonuses to the player for placing a tile
func (s *PlayerServiceImpl) awardTilePlacementBonuses(ctx context.Context, gameID, playerID string, coordinate model.HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get the game state to access the board
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Find the placed tile in the board
	var placedTile *model.Tile
	for i, tile := range game.Board.Tiles {
		if tile.Coordinates.Q == coordinate.Q && tile.Coordinates.R == coordinate.R && tile.Coordinates.S == coordinate.S {
			placedTile = &game.Board.Tiles[i]
			break
		}
	}

	if placedTile == nil {
		log.Warn("Placed tile not found in board", zap.Any("coordinate", coordinate))
		return nil // Not critical, just skip bonuses
	}

	// Get current player resources
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	newResources := player.Resources
	var totalCreditsBonus int
	var bonusesAwarded []string

	// Award tile bonuses (steel, titanium, plants, card draw)
	for _, bonus := range placedTile.Bonuses {
		switch bonus.Type {
		case model.ResourceSteel:
			newResources.Steel += bonus.Amount
			bonusesAwarded = append(bonusesAwarded, fmt.Sprintf("+%d steel", bonus.Amount))
			log.Info("🎁 Tile bonus awarded",
				zap.String("type", "steel"),
				zap.Int("amount", bonus.Amount))

		case model.ResourceTitanium:
			newResources.Titanium += bonus.Amount
			bonusesAwarded = append(bonusesAwarded, fmt.Sprintf("+%d titanium", bonus.Amount))
			log.Info("🎁 Tile bonus awarded",
				zap.String("type", "titanium"),
				zap.Int("amount", bonus.Amount))

		case model.ResourcePlants:
			newResources.Plants += bonus.Amount
			bonusesAwarded = append(bonusesAwarded, fmt.Sprintf("+%d plants", bonus.Amount))
			log.Info("🎁 Tile bonus awarded",
				zap.String("type", "plants"),
				zap.Int("amount", bonus.Amount))

		case model.ResourceCardDraw:
			// TODO: Implement card drawing bonus
			// For now, just log it
			bonusesAwarded = append(bonusesAwarded, fmt.Sprintf("+%d cards", bonus.Amount))
			log.Info("🎁 Tile bonus awarded (card draw not yet implemented)",
				zap.String("type", "card-draw"),
				zap.Int("amount", bonus.Amount))
		}
	}

	// Award ocean adjacency bonus (+2 MC per adjacent ocean)
	oceanAdjacencyBonus := s.calculateOceanAdjacencyBonus(game, coordinate)
	if oceanAdjacencyBonus > 0 {
		totalCreditsBonus += oceanAdjacencyBonus
		bonusesAwarded = append(bonusesAwarded, fmt.Sprintf("+%d MC (ocean adjacency)", oceanAdjacencyBonus))
		log.Info("🌊 Ocean adjacency bonus awarded",
			zap.Int("bonus", oceanAdjacencyBonus))
	}

	// Apply credits bonus if any
	if totalCreditsBonus > 0 {
		newResources.Credits += totalCreditsBonus
	}

	// Update player resources if any bonuses were awarded
	if len(bonusesAwarded) > 0 {
		if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
			return fmt.Errorf("failed to update player resources: %w", err)
		}

		log.Info("✅ All tile placement bonuses awarded",
			zap.Strings("bonuses", bonusesAwarded))
	}

	return nil
}

// calculateOceanAdjacencyBonus calculates the bonus from adjacent ocean tiles
func (s *PlayerServiceImpl) calculateOceanAdjacencyBonus(game model.Game, coordinate model.HexPosition) int {
	// Define the 6 adjacent hex directions (cube coordinates)
	directions := []model.HexPosition{
		{Q: 1, R: -1, S: 0}, // East
		{Q: 1, R: 0, S: -1}, // Southeast
		{Q: 0, R: 1, S: -1}, // Southwest
		{Q: -1, R: 1, S: 0}, // West
		{Q: -1, R: 0, S: 1}, // Northwest
		{Q: 0, R: -1, S: 1}, // Northeast
	}

	adjacentOceanCount := 0

	// Check each adjacent position for ocean tiles
	for _, dir := range directions {
		adjacentCoord := model.HexPosition{
			Q: coordinate.Q + dir.Q,
			R: coordinate.R + dir.R,
			S: coordinate.S + dir.S,
		}

		// Find the adjacent tile in the board
		for _, tile := range game.Board.Tiles {
			if tile.Coordinates.Q == adjacentCoord.Q &&
				tile.Coordinates.R == adjacentCoord.R &&
				tile.Coordinates.S == adjacentCoord.S {
				// Check if this tile has an ocean
				if tile.OccupiedBy != nil && tile.OccupiedBy.Type == model.ResourceOceanTile {
					adjacentOceanCount++
				}
				break
			}
		}
	}

	// Each adjacent ocean provides +2 MC
	return adjacentOceanCount * 2
}
