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
	// Update player resources
	UpdatePlayerResources(ctx context.Context, gameID, playerID string, newResources model.Resources) error

	// Update player production
	UpdatePlayerProduction(ctx context.Context, gameID, playerID string, newProduction model.Production) error

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

	// Terraform rating methods
	UpdatePlayerTR(ctx context.Context, gameID, playerID string, newTR int) error
	AddPlayerTR(ctx context.Context, gameID, playerID string, trIncrease int) error

	// Standard project utility methods
	CanAffordStandardProject(player *model.Player, project model.StandardProject) bool
	HasCardsToSell(player *model.Player, count int) bool
	GetMaxCardsToSell(player *model.Player) int

	// Admin methods (development mode only)
	AddCardToHand(ctx context.Context, gameID, playerID, cardID string) error
	SetResources(ctx context.Context, gameID, playerID string, resources model.Resources) error
	SetProduction(ctx context.Context, gameID, playerID string, production model.Production) error
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

// UpdatePlayerResources updates a player's resources
func (s *PlayerServiceImpl) UpdatePlayerResources(ctx context.Context, gameID, playerID string, newResources model.Resources) error {
	log := logger.WithGameContext(gameID, playerID)

	// Update through PlayerRepository using granular update
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
		log.Error("Failed to update player resources", zap.Error(err))
		return fmt.Errorf("failed to update player: %w", err)
	}

	log.Info("Player resources updated")
	return nil
}

// UpdatePlayerProduction updates a player's production
func (s *PlayerServiceImpl) UpdatePlayerProduction(ctx context.Context, gameID, playerID string, newProduction model.Production) error {
	log := logger.WithGameContext(gameID, playerID)

	// Update through PlayerRepository using granular update
	if err := s.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
		log.Error("Failed to update player production", zap.Error(err))
		return fmt.Errorf("failed to update player: %w", err)
	}

	log.Info("Player production updated")
	return nil
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

	return s.UpdatePlayerProduction(ctx, gameID, playerID, newProduction)
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

	return s.UpdatePlayerResources(ctx, gameID, playerID, newResources)
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

	return s.UpdatePlayerResources(ctx, gameID, playerID, newResources)
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

// UpdatePlayerTR updates a player's terraform rating
func (s *PlayerServiceImpl) UpdatePlayerTR(ctx context.Context, gameID, playerID string, newTR int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for TR update", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Update terraform rating using granular repository method
	if err := s.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
		log.Error("Failed to update player terraform rating", zap.Error(err))
		return fmt.Errorf("failed to update player terraform rating: %w", err)
	}

	log.Info("Player terraform rating updated",
		zap.Int("old_tr", player.TerraformRating),
		zap.Int("new_tr", newTR))
	return nil
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

// AddPlayerTR increases a player's terraform rating by the specified amount
func (s *PlayerServiceImpl) AddPlayerTR(ctx context.Context, gameID, playerID string, trIncrease int) error {
	player, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	newTR := player.TerraformRating + trIncrease
	return s.UpdatePlayerTR(ctx, gameID, playerID, newTR)
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

// Admin methods (development mode only)

// AddCardToHand adds a card to a player's hand (admin command)
func (s *PlayerServiceImpl) AddCardToHand(ctx context.Context, gameID, playerID, cardID string) error {
	log := logger.WithContext()

	log.Info("🎴 Admin adding card to player's hand",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("card_id", cardID))

	// Verify player exists before adding card
	_, err := s.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("player not found: %w", err)
	}

	// Add card to player's hand using repository method
	if err := s.playerRepo.AddCard(ctx, gameID, playerID, cardID); err != nil {
		log.Error("Failed to add card to player's hand", zap.Error(err))
		return fmt.Errorf("failed to add card to player's hand: %w", err)
	}

	// Broadcast game state to all players
	if broadcastErr := s.sessionManager.Broadcast(gameID); broadcastErr != nil {
		log.Error("Failed to broadcast game state after adding card", zap.Error(broadcastErr))
		// Don't fail the operation, just log the error
	}

	log.Info("✅ Card added to player's hand successfully",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("card_id", cardID))

	return nil
}

// SetResources sets a player's resources directly (admin command)
func (s *PlayerServiceImpl) SetResources(ctx context.Context, gameID, playerID string, resources model.Resources) error {
	log := logger.WithContext()

	log.Info("💰 Admin setting player resources",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Any("resources", resources))

	// Use resources directly since they're already model.Resources
	modelResources := resources

	// Validate that resources are non-negative
	if modelResources.Credits < 0 || modelResources.Steel < 0 || modelResources.Titanium < 0 ||
		modelResources.Plants < 0 || modelResources.Energy < 0 || modelResources.Heat < 0 {
		return fmt.Errorf("resource values cannot be negative")
	}

	// Update player resources
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, modelResources); err != nil {
		log.Error("Failed to set player resources", zap.Error(err))
		return fmt.Errorf("failed to set player resources: %w", err)
	}

	// Broadcast game state to all players
	if broadcastErr := s.sessionManager.Broadcast(gameID); broadcastErr != nil {
		log.Error("Failed to broadcast game state after setting resources", zap.Error(broadcastErr))
		// Don't fail the operation, just log the error
	}

	log.Info("✅ Player resources set successfully",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Any("new_resources", modelResources))

	return nil
}

// SetProduction sets a player's production directly (admin command)
func (s *PlayerServiceImpl) SetProduction(ctx context.Context, gameID, playerID string, production model.Production) error {
	log := logger.WithContext()

	log.Info("🏭 Admin setting player production",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Any("production", production))

	// Use production directly since it's already model.Production
	modelProduction := production

	// Validate that production values are reasonable (can be negative but not too extreme)
	const maxProduction = 50
	const minProduction = -20

	if modelProduction.Credits < minProduction || modelProduction.Credits > maxProduction ||
		modelProduction.Steel < minProduction || modelProduction.Steel > maxProduction ||
		modelProduction.Titanium < minProduction || modelProduction.Titanium > maxProduction ||
		modelProduction.Plants < minProduction || modelProduction.Plants > maxProduction ||
		modelProduction.Energy < minProduction || modelProduction.Energy > maxProduction ||
		modelProduction.Heat < minProduction || modelProduction.Heat > maxProduction {
		return fmt.Errorf("production values must be between %d and %d", minProduction, maxProduction)
	}

	// Update player production
	if err := s.playerRepo.UpdateProduction(ctx, gameID, playerID, modelProduction); err != nil {
		log.Error("Failed to set player production", zap.Error(err))
		return fmt.Errorf("failed to set player production: %w", err)
	}

	// Broadcast game state to all players
	if broadcastErr := s.sessionManager.Broadcast(gameID); broadcastErr != nil {
		log.Error("Failed to broadcast game state after setting production", zap.Error(broadcastErr))
		// Don't fail the operation, just log the error
	}

	log.Info("✅ Player production set successfully",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Any("new_production", modelProduction))

	return nil
}
