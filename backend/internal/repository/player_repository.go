package repository

import (
	"context"
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"

	"go.uber.org/zap"
)

// PlayerRepository provides clean CRUD operations and granular updates for players
type PlayerRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, gameID string, player model.Player) error
	GetByID(ctx context.Context, gameID, playerID string) (model.Player, error)
	Delete(ctx context.Context, gameID, playerID string) error
	ListByGameID(ctx context.Context, gameID string) ([]model.Player, error)

	// Granular update methods for specific fields
	UpdateResources(ctx context.Context, gameID, playerID string, resources model.Resources) error
	UpdateProduction(ctx context.Context, gameID, playerID string, production model.Production) error
	UpdateTerraformRating(ctx context.Context, gameID, playerID string, rating int) error
	UpdateCorporation(ctx context.Context, gameID, playerID string, corporation string) error
	UpdateConnectionStatus(ctx context.Context, gameID, playerID string, status model.ConnectionStatus) error
	UpdateIsActive(ctx context.Context, gameID, playerID string, isActive bool) error
	UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error
	UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error
	UpdateVictoryPoints(ctx context.Context, gameID, playerID string, points int) error
	AddCard(ctx context.Context, gameID, playerID string, cardID string) error
	RemoveCard(ctx context.Context, gameID, playerID string, cardID string) error
	PlayCard(ctx context.Context, gameID, playerID string, cardID string) error
}

// PlayerRepositoryImpl implements PlayerRepository with in-memory storage
type PlayerRepositoryImpl struct {
	// Map of gameID -> map of playerID -> Player
	players  map[string]map[string]*model.Player
	mutex    sync.RWMutex
	eventBus events.EventBus
}

// NewPlayerRepository creates a new player repository
func NewPlayerRepository(eventBus events.EventBus) PlayerRepository {
	return &PlayerRepositoryImpl{
		players:  make(map[string]map[string]*model.Player),
		eventBus: eventBus,
	}
}

// Create adds a player to a game
func (r *PlayerRepositoryImpl) Create(ctx context.Context, gameID string, player model.Player) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, player.ID)

	if gameID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	if player.ID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	// Initialize game players map if it doesn't exist
	if r.players[gameID] == nil {
		r.players[gameID] = make(map[string]*model.Player)
	}

	// Check if player already exists
	if _, exists := r.players[gameID][player.ID]; exists {
		log.Error("Player already exists in game")
		return fmt.Errorf("player with ID %s already exists in game %s", player.ID, gameID)
	}

	// Store a copy to prevent external mutation
	playerCopy := player.DeepCopy()
	r.players[gameID][player.ID] = playerCopy

	log.Debug("Player added to game", zap.String("player_name", player.Name))

	// Publish player added event
	if r.eventBus != nil {
		playerAddedEvent := events.NewPlayerJoinedEvent(gameID, player.ID, player.Name)
		if err := r.eventBus.Publish(ctx, playerAddedEvent); err != nil {
			log.Warn("Failed to publish player added event", zap.Error(err))
		}
	}

	return nil
}

// GetByID retrieves a player by game and player ID
func (r *PlayerRepositoryImpl) GetByID(ctx context.Context, gameID, playerID string) (model.Player, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return model.Player{}, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	if gameID == "" {
		return model.Player{}, fmt.Errorf("game ID cannot be empty")
	}

	if playerID == "" {
		return model.Player{}, fmt.Errorf("player ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return model.Player{}, fmt.Errorf("no players found for game %s", gameID)
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return model.Player{}, fmt.Errorf("player with ID %s not found in game %s", playerID, gameID)
	}

	// Return a copy to prevent external mutation
	return *player.DeepCopy(), nil
}

// Delete removes a player from a game
func (r *PlayerRepositoryImpl) Delete(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	if gameID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	if playerID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return fmt.Errorf("no players found for game %s", gameID)
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return fmt.Errorf("player with ID %s not found in game %s", playerID, gameID)
	}

	delete(gamePlayers, playerID)

	// Clean up empty game
	if len(gamePlayers) == 0 {
		delete(r.players, gameID)
	}

	log.Info("Player removed from game", zap.String("player_name", player.Name))

	// Publish player removed event
	if r.eventBus != nil {
		playerRemovedEvent := events.NewPlayerLeftEvent(gameID, playerID, player.Name)
		if err := r.eventBus.Publish(ctx, playerRemovedEvent); err != nil {
			log.Warn("Failed to publish player removed event", zap.Error(err))
		}
	}

	return nil
}

// ListByGameID returns all players in a game
func (r *PlayerRepositoryImpl) ListByGameID(ctx context.Context, gameID string) ([]model.Player, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return make([]model.Player, 0), nil
	}

	players := make([]model.Player, 0, len(gamePlayers))
	for _, player := range gamePlayers {
		players = append(players, *player.DeepCopy())
	}

	return players, nil
}

// UpdateResources updates a player's resources
func (r *PlayerRepositoryImpl) UpdateResources(ctx context.Context, gameID, playerID string, resources model.Resources) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	oldResources := player.Resources
	player.Resources = resources

	log.Info("Player resources updated")

	// Publish consolidated player changed event for resources
	if r.eventBus != nil && oldResources != resources {
		resourcesChangedEvent := events.NewPlayerResourcesChangedEvent(gameID, playerID)
		if err := r.eventBus.Publish(ctx, resourcesChangedEvent); err != nil {
			log.Warn("Failed to publish player resources changed event", zap.Error(err))
		}
	}

	return nil
}

// UpdateProduction updates a player's production
func (r *PlayerRepositoryImpl) UpdateProduction(ctx context.Context, gameID, playerID string, production model.Production) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	oldProduction := player.Production
	player.Production = production

	log.Info("Player production updated")

	// Publish consolidated player changed event for production
	if r.eventBus != nil && oldProduction != production {
		productionChangedEvent := events.NewPlayerProductionChangedEvent(gameID, playerID)
		if err := r.eventBus.Publish(ctx, productionChangedEvent); err != nil {
			log.Warn("Failed to publish player production changed event", zap.Error(err))
		}
	}

	return nil
}

// UpdateTerraformRating updates a player's terraform rating
func (r *PlayerRepositoryImpl) UpdateTerraformRating(ctx context.Context, gameID, playerID string, rating int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	oldTR := player.TerraformRating
	player.TerraformRating = rating

	log.Info("Player terraform rating updated", zap.Int("old_tr", oldTR), zap.Int("new_tr", rating))

	// Publish consolidated player changed event for terraform rating
	if r.eventBus != nil && oldTR != rating {
		trChangedEvent := events.NewPlayerTRChangedEvent(gameID, playerID)
		if err := r.eventBus.Publish(ctx, trChangedEvent); err != nil {
			log.Warn("Failed to publish player TR changed event", zap.Error(err))
		}
	}

	return nil
}

// UpdateCorporation updates a player's corporation
func (r *PlayerRepositoryImpl) UpdateCorporation(ctx context.Context, gameID, playerID string, corporation string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	oldCorporation := player.Corporation
	player.Corporation = corporation

	log.Info("Player corporation updated", zap.String("old_corp", oldCorporation), zap.String("new_corp", corporation))

	return nil
}

// UpdateConnectionStatus updates a player's connection status
func (r *PlayerRepositoryImpl) UpdateConnectionStatus(ctx context.Context, gameID, playerID string, status model.ConnectionStatus) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	oldStatus := player.ConnectionStatus
	player.ConnectionStatus = status

	log.Info("Player connection status updated", zap.String("old_status", string(oldStatus)), zap.String("new_status", string(status)))

	return nil
}

// UpdateIsActive updates a player's active status
func (r *PlayerRepositoryImpl) UpdateIsActive(ctx context.Context, gameID, playerID string, isActive bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.IsActive = isActive

	log.Info("Player active status updated", zap.Bool("is_active", isActive))

	return nil
}

// UpdatePassed updates a player's passed status
func (r *PlayerRepositoryImpl) UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.Passed = passed

	log.Info("Player passed status updated", zap.Bool("passed", passed))

	return nil
}

// UpdateAvailableActions updates a player's available actions count
func (r *PlayerRepositoryImpl) UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.AvailableActions = actions

	log.Debug("Player available actions updated", zap.Int("actions", actions))

	return nil
}

// UpdateVictoryPoints updates a player's victory points
func (r *PlayerRepositoryImpl) UpdateVictoryPoints(ctx context.Context, gameID, playerID string, points int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.VictoryPoints = points

	log.Info("Player victory points updated", zap.Int("points", points))

	return nil
}

// AddCard adds a card to a player's hand
func (r *PlayerRepositoryImpl) AddCard(ctx context.Context, gameID, playerID string, cardID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Check if card already exists in hand
	for _, existingCard := range player.Cards {
		if existingCard == cardID {
			return fmt.Errorf("card %s already exists in player %s hand", cardID, playerID)
		}
	}

	player.Cards = append(player.Cards, cardID)

	log.Info("Card added to player hand", zap.String("card_id", cardID))

	return nil
}

// RemoveCard removes a card from a player's hand
func (r *PlayerRepositoryImpl) RemoveCard(ctx context.Context, gameID, playerID string, cardID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Find and remove card
	for i, existingCard := range player.Cards {
		if existingCard == cardID {
			player.Cards = append(player.Cards[:i], player.Cards[i+1:]...)
			log.Info("Card removed from player hand", zap.String("card_id", cardID))
			return nil
		}
	}

	return fmt.Errorf("card %s not found in player %s hand", cardID, playerID)
}

// PlayCard moves a card from hand to played cards
func (r *PlayerRepositoryImpl) PlayCard(ctx context.Context, gameID, playerID string, cardID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Find and remove card from hand
	cardFound := false
	for i, existingCard := range player.Cards {
		if existingCard == cardID {
			player.Cards = append(player.Cards[:i], player.Cards[i+1:]...)
			cardFound = true
			break
		}
	}

	if !cardFound {
		return fmt.Errorf("card %s not found in player %s hand", cardID, playerID)
	}

	// Add to played cards
	player.PlayedCards = append(player.PlayedCards, cardID)

	log.Info("Card played", zap.String("card_id", cardID))

	return nil
}

// getPlayerUnsafe retrieves a player without acquiring locks (assumes lock is held)
func (r *PlayerRepositoryImpl) getPlayerUnsafe(gameID, playerID string) (*model.Player, error) {
	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	if playerID == "" {
		return nil, fmt.Errorf("player ID cannot be empty")
	}

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return nil, fmt.Errorf("no players found for game %s", gameID)
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return nil, fmt.Errorf("player with ID %s not found in game %s", playerID, gameID)
	}

	return player, nil
}
