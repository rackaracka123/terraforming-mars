package service

import (
	"context"
	"fmt"
	"time"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	gamePkg "terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

const (
	// BasePlantsForGreenery is the base cost in plants to convert to greenery (before discounts)
	BasePlantsForGreenery = 8

	// BaseHeatForTemperature is the base cost in heat to raise temperature (before discounts)
	BaseHeatForTemperature = 8
)

// ResourceConversionService handles resource conversion operations (plants→greenery, heat→temperature)
type ResourceConversionService interface {
	// InitiatePlantConversion initiates plants-to-greenery conversion (deducts plants, creates tile selection)
	InitiatePlantConversion(ctx context.Context, gameID, playerID string) error

	// CompletePlantConversion completes the plant conversion by placing the greenery tile
	CompletePlantConversion(ctx context.Context, gameID, playerID string, coordinate model.HexPosition) error

	// ConvertHeatToTemperature converts heat to raise temperature (awards TR)
	ConvertHeatToTemperature(ctx context.Context, gameID, playerID string) error
}

// ResourceConversionServiceImpl implements ResourceConversionService interface
type ResourceConversionServiceImpl struct {
	gameRepo       gamePkg.Repository
	playerRepo     playerPkg.Repository
	boardService   BoardService
	sessionManager session.SessionManager
	eventBus       *events.EventBusImpl
}

// NewResourceConversionService creates a new ResourceConversionService instance
func NewResourceConversionService(
	gameRepo gamePkg.Repository,
	playerRepo playerPkg.Repository,
	boardService BoardService,
	sessionManager session.SessionManager,
	eventBus *events.EventBusImpl,
) ResourceConversionService {
	return &ResourceConversionServiceImpl{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		boardService:   boardService,
		sessionManager: sessionManager,
		eventBus:       eventBus,
	}
}

// InitiatePlantConversion initiates the plant conversion process (deducts plants, creates pending tile selection)
func (s *ResourceConversionServiceImpl) InitiatePlantConversion(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🌱 Initiating plant conversion")

	// Get player
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Calculate required plants (considering discounts)
	requiredPlants := cards.CalculateResourceConversionCost(&player, model.StandardProjectConvertPlantsToGreenery, BasePlantsForGreenery)

	// Validate player has enough plants
	if player.Resources.Plants < requiredPlants {
		log.Warn("Player cannot afford plants conversion",
			zap.Int("required", requiredPlants),
			zap.Int("available", player.Resources.Plants))
		return fmt.Errorf("insufficient plants: need %d, have %d", requiredPlants, player.Resources.Plants)
	}

	// Deduct plants
	updatedResources := player.Resources
	updatedResources.Plants -= requiredPlants

	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, updatedResources); err != nil {
		log.Error("Failed to deduct plants", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	// Get game to access the board for calculating available hexes
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Calculate available hexes using board service (same logic as standard project greenery)
	availableHexes, err := s.boardService.CalculateAvailableHexesForTileTypeWithPlayer(game, "greenery", playerID)
	if err != nil {
		log.Error("Failed to calculate available hexes", zap.Error(err))
		return fmt.Errorf("failed to calculate available hexes: %w", err)
	}

	// Create pending tile selection
	pendingSelection := &model.PendingTileSelection{
		TileType:       "greenery",
		AvailableHexes: availableHexes,
		Source:         "convert-plants-to-greenery",
	}

	if err := s.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, pendingSelection); err != nil {
		log.Error("Failed to create pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to create pending tile selection: %w", err)
	}

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
	}

	log.Info("✅ Plant conversion initiated, waiting for tile selection",
		zap.Int("plants_spent", requiredPlants),
		zap.Int("available_hexes", len(availableHexes)))

	return nil
}

// CompletePlantConversion completes the plant conversion by placing the greenery tile and raising oxygen
func (s *ResourceConversionServiceImpl) CompletePlantConversion(ctx context.Context, gameID, playerID string, coordinate model.HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🌱 Completing plant conversion",
		zap.Int("q", coordinate.Q),
		zap.Int("r", coordinate.R),
		zap.Int("s", coordinate.S))

	// Place greenery tile
	if err := s.placeTile(ctx, gameID, playerID, coordinate); err != nil {
		log.Error("Failed to place greenery tile", zap.Error(err))
		return fmt.Errorf("failed to place greenery tile: %w", err)
	}

	// Raise oxygen by 1 (if not already maxed)
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	oxygenRaised := false
	if game.GlobalParameters.Oxygen < model.MaxOxygen {
		newParams := game.GlobalParameters
		newParams.Oxygen++

		if err := s.gameRepo.UpdateGlobalParameters(ctx, gameID, newParams); err != nil {
			log.Error("Failed to raise oxygen", zap.Error(err))
			return fmt.Errorf("failed to raise oxygen: %w", err)
		}

		oxygenRaised = true
		log.Info("🌍 Oxygen raised",
			zap.Int("new_oxygen", newParams.Oxygen))
	} else {
		log.Info("🌍 Oxygen already at maximum, no TR awarded")
	}

	// Award TR if oxygen was raised
	if oxygenRaised {
		player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			log.Error("Failed to get player for TR update", zap.Error(err))
			return fmt.Errorf("failed to get player: %w", err)
		}

		newTR := player.TerraformRating + 1
		if err := s.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
			log.Error("Failed to update terraform rating", zap.Error(err))
			return fmt.Errorf("failed to update terraform rating: %w", err)
		}

		log.Info("⭐ Terraform rating increased",
			zap.Int("new_tr", newTR))
	}

	// Clear pending tile selection
	if err := s.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, nil); err != nil {
		log.Error("Failed to clear pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to clear pending tile selection: %w", err)
	}

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
	}

	log.Info("✅ Plant conversion completed successfully")

	return nil
}

// ConvertHeatToTemperature converts 8 heat (or less with discounts) to raise temperature by 1 step
func (s *ResourceConversionServiceImpl) ConvertHeatToTemperature(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🔥 Converting heat to temperature")

	// Get player
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Calculate required heat (considering discounts)
	requiredHeat := cards.CalculateResourceConversionCost(&player, model.StandardProjectConvertHeatToTemperature, BaseHeatForTemperature)

	// Validate player has enough heat
	if player.Resources.Heat < requiredHeat {
		log.Warn("Player cannot afford heat conversion",
			zap.Int("required", requiredHeat),
			zap.Int("available", player.Resources.Heat))
		return fmt.Errorf("insufficient heat: need %d, have %d", requiredHeat, player.Resources.Heat)
	}

	// Deduct heat
	updatedResources := player.Resources
	updatedResources.Heat -= requiredHeat

	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, updatedResources); err != nil {
		log.Error("Failed to deduct heat", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	// Raise temperature by 1 step (+2°C) if not already maxed
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	temperatureRaised := false
	if game.GlobalParameters.Temperature < model.MaxTemperature {
		newParams := game.GlobalParameters
		newParams.Temperature += 2 // Each step is 2°C

		// Ensure we don't exceed max temperature
		if newParams.Temperature > model.MaxTemperature {
			newParams.Temperature = model.MaxTemperature
		}

		if err := s.gameRepo.UpdateGlobalParameters(ctx, gameID, newParams); err != nil {
			log.Error("Failed to raise temperature", zap.Error(err))
			return fmt.Errorf("failed to raise temperature: %w", err)
		}

		temperatureRaised = true
		log.Info("🌡️ Temperature raised",
			zap.Int("new_temperature", newParams.Temperature))
	} else {
		log.Info("🌡️ Temperature already at maximum, no TR awarded")
	}

	// Award TR if temperature was raised
	if temperatureRaised {
		player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			log.Error("Failed to get player for TR update", zap.Error(err))
			return fmt.Errorf("failed to get player: %w", err)
		}

		newTR := player.TerraformRating + 1
		if err := s.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
			log.Error("Failed to update terraform rating", zap.Error(err))
			return fmt.Errorf("failed to update terraform rating: %w", err)
		}

		log.Info("⭐ Terraform rating increased",
			zap.Int("new_tr", newTR))
	}

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
	}

	log.Info("✅ Heat converted to temperature successfully",
		zap.Int("heat_spent", requiredHeat))

	return nil
}

// placeTile places a greenery tile at the specified coordinate and awards placement bonuses
func (s *ResourceConversionServiceImpl) placeTile(ctx context.Context, gameID, playerID string, coordinate model.HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🔧 Placing greenery tile",
		zap.Int("q", coordinate.Q),
		zap.Int("r", coordinate.R),
		zap.Int("s", coordinate.S))

	// Create the greenery tile occupant
	occupant := &model.TileOccupant{
		Type: model.ResourceGreeneryTile,
		Tags: []string{},
	}

	// Award placement bonuses to the player
	if err := s.awardTilePlacementBonuses(ctx, gameID, playerID, coordinate); err != nil {
		log.Error("Failed to award tile placement bonuses", zap.Error(err))
		return fmt.Errorf("failed to award tile placement bonuses: %w", err)
	}

	// Update the tile occupancy in the game board
	if err := s.gameRepo.UpdateTileOccupancy(ctx, gameID, coordinate, occupant, &playerID); err != nil {
		log.Error("Failed to update tile occupancy", zap.Error(err))
		return fmt.Errorf("failed to update tile occupancy: %w", err)
	}

	// Passive effects are triggered automatically via TilePlacedEvent from the repository
	// No manual triggering needed here - event system handles it

	log.Info("✅ Greenery tile placed successfully",
		zap.String("coordinate", coordinate.String()))

	return nil
}

// awardTilePlacementBonuses awards bonuses to the player for placing a tile
func (s *ResourceConversionServiceImpl) awardTilePlacementBonuses(ctx context.Context, gameID, playerID string, coordinate model.HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get the game state to access the board
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Find the placed tile in the board
	var placedTile *model.Tile
	for i, tile := range game.Board.Tiles {
		if tile.Coordinates.Equals(coordinate) {
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

	// Collect all placement bonuses for event publishing
	placementBonuses := make(map[string]int)

	// Award tile bonuses (steel, titanium, plants, card draw)
	for _, bonus := range placedTile.Bonuses {
		description, applied, resourceType, amount := s.applyTileBonus(ctx, gameID, playerID, coordinate, &newResources, bonus, log)
		if applied {
			bonusesAwarded = append(bonusesAwarded, description)
			// Collect bonus for event
			if resourceType != "" {
				placementBonuses[resourceType] += amount
			}
		}
	}

	// Award ocean adjacency bonus (+2 MC per adjacent ocean)
	oceanAdjacencyBonus := s.calculateOceanAdjacencyBonus(game, player, coordinate)
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

	// Publish PlacementBonusGainedEvent with all resources if any bonuses were gained
	if len(placementBonuses) > 0 && s.eventBus != nil {
		events.Publish(s.eventBus, gamePkg.PlacementBonusGainedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			Resources: placementBonuses,
			Q:         coordinate.Q,
			R:         coordinate.R,
			S:         coordinate.S,
			Timestamp: time.Now(),
		})

		log.Debug("📢 Published PlacementBonusGainedEvent",
			zap.Any("resources", placementBonuses))
	}

	return nil
}

// calculateOceanAdjacencyBonus calculates the bonus from adjacent ocean tiles
// Base bonus is 2 MC per ocean
func (s *ResourceConversionServiceImpl) calculateOceanAdjacencyBonus(game model.Game, player model.Player, coordinate model.HexPosition) int {
	adjacentOceanCount := 0

	// Check each adjacent position for ocean tiles
	for _, adjacentCoord := range coordinate.GetNeighbors() {
		// Find the adjacent tile in the board
		for _, tile := range game.Board.Tiles {
			if tile.Coordinates.Equals(adjacentCoord) {
				// Check if this tile has an ocean
				if tile.OccupiedBy != nil && tile.OccupiedBy.Type == model.ResourceOceanTile {
					adjacentOceanCount++
				}
				break
			}
		}
	}

	// Base ocean adjacency bonus is 2 MC per ocean
	baseBonus := 2

	// Calculate total bonus: adjacent oceans * base bonus
	return adjacentOceanCount * baseBonus
}

// applyTileBonus applies a single tile bonus to player resources
// Returns: description, applied (bool), resourceType, amount
func (s *ResourceConversionServiceImpl) applyTileBonus(ctx context.Context, gameID, playerID string, coordinate model.HexPosition, resources *model.Resources, bonus model.TileBonus, log *zap.Logger) (string, bool, string, int) {
	var resourceName string

	switch bonus.Type {
	case model.ResourceSteel:
		resources.Steel += bonus.Amount
		resourceName = "steel"

	case model.ResourceTitanium:
		resources.Titanium += bonus.Amount
		resourceName = "titanium"

	case model.ResourcePlants:
		resources.Plants += bonus.Amount
		resourceName = "plants"

	case model.ResourceCardDraw:
		// TODO: Implement card drawing bonus
		log.Info("🎁 Tile bonus awarded (card draw not yet implemented)",
			zap.String("type", "card-draw"),
			zap.Int("amount", bonus.Amount))
		return fmt.Sprintf("+%d cards", bonus.Amount), true, string(model.ResourceCardDraw), bonus.Amount

	default:
		// Unknown bonus type, skip it
		return "", false, "", 0
	}

	log.Info("🎁 Tile bonus awarded",
		zap.String("type", resourceName),
		zap.Int("amount", bonus.Amount))

	return fmt.Sprintf("+%d %s", bonus.Amount, resourceName), true, string(bonus.Type), bonus.Amount
}
