package repository

import (
	"context"
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"

	"go.uber.org/zap"
)

// PlayerRepository provides clean CRUD operations and granular updates for players
type PlayerRepository interface {
	Create(ctx context.Context, gameID string, player model.Player) error
	GetByID(ctx context.Context, gameID, playerID string) (model.Player, error)
	Delete(ctx context.Context, gameID, playerID string) error
	ListByGameID(ctx context.Context, gameID string) ([]model.Player, error)

	UpdateResources(ctx context.Context, gameID, playerID string, resources model.Resources) error
	UpdateProduction(ctx context.Context, gameID, playerID string, production model.Production) error
	UpdateTerraformRating(ctx context.Context, gameID, playerID string, rating int) error
	UpdateCorporation(ctx context.Context, gameID, playerID string, corporation string) error
	UpdateConnectionStatus(ctx context.Context, gameID, playerID string, isConnected bool) error
	UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error
	UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error
	UpdateVictoryPoints(ctx context.Context, gameID, playerID string, points int) error
	UpdateEffects(ctx context.Context, gameID, playerID string, effects []model.PlayerEffect) error
	UpdatePlayerActions(ctx context.Context, gameID, playerID string, actions []model.PlayerAction) error
	AddCard(ctx context.Context, gameID, playerID string, cardID string) error
	RemoveCard(ctx context.Context, gameID, playerID string, cardID string) error
	RemoveCardFromHand(ctx context.Context, gameID, playerID string, cardID string) error

	UpdateSelectStartingCardsPhase(ctx context.Context, gameID, playerID string, selectStartingCardPhase *model.SelectStartingCardsPhase) error
	SetStartingCardsSelectionComplete(ctx context.Context, gameID, playerID string) error

	UpdateProductionPhase(ctx context.Context, gameID, playerID string, productionPhase *model.ProductionPhase) error
	SetProductionCardsSelectionComplete(ctx context.Context, gameID, playerID string) error

	// Tile selection methods
	UpdatePendingTileSelection(ctx context.Context, gameID, playerID string, selection *model.PendingTileSelection) error
	GetPendingTileSelection(ctx context.Context, gameID, playerID string) (*model.PendingTileSelection, error)
	ClearPendingTileSelection(ctx context.Context, gameID, playerID string) error

	// Tile queue methods
	UpdatePendingTileSelectionQueue(ctx context.Context, gameID, playerID string, queue *model.PendingTileSelectionQueue) error
	GetPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) (*model.PendingTileSelectionQueue, error)
	ClearPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) error

	// Tile queue operations
	CreateTileQueue(ctx context.Context, gameID, playerID, cardID string, tilePlacements []string) error
	ProcessNextTileInQueue(ctx context.Context, gameID, playerID string) (string, error)
}

// PlayerRepositoryImpl implements PlayerRepository with in-memory storage
type PlayerRepositoryImpl struct {
	// Map of gameID -> map of playerID -> Player
	players map[string]map[string]*model.Player
	mutex   sync.RWMutex
}

// NewPlayerRepository creates a new player repository
func NewPlayerRepository() PlayerRepository {
	return &PlayerRepositoryImpl{
		players: make(map[string]map[string]*model.Player),
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
		return model.Player{}, &model.NotFoundError{Resource: "player", ID: playerID}
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

	player.Resources = resources

	log.Info("Player resources updated")

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

	player.Production = production

	log.Info("Player production updated")

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

	var oldCorporation string
	if player.Corporation != nil {
		oldCorporation = *player.Corporation
	}
	player.Corporation = &corporation

	log.Info("Player corporation updated", zap.String("old_corp", oldCorporation), zap.String("new_corp", corporation))

	return nil
}

// UpdateConnectionStatus updates a player's connection status
func (r *PlayerRepositoryImpl) UpdateConnectionStatus(ctx context.Context, gameID, playerID string, isConnected bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	oldStatus := player.IsConnected
	player.IsConnected = isConnected

	log.Info("Player connection status updated", zap.Bool("old_status", oldStatus), zap.Bool("new_status", isConnected))

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

// UpdateEffects updates a player's active effects list
func (r *PlayerRepositoryImpl) UpdateEffects(ctx context.Context, gameID, playerID string, effects []model.PlayerEffect) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Deep copy the effects to prevent external mutation
	effectsCopy := make([]model.PlayerEffect, len(effects))
	for i, effect := range effects {
		effectsCopy[i] = *effect.DeepCopy()
	}

	oldEffectsCount := len(player.Effects)
	player.Effects = effectsCopy

	log.Info("Player effects updated",
		zap.Int("old_effects_count", oldEffectsCount),
		zap.Int("new_effects_count", len(effects)))

	return nil
}

// UpdatePlayerActions updates a player's available actions
func (r *PlayerRepositoryImpl) UpdatePlayerActions(ctx context.Context, gameID, playerID string, actions []model.PlayerAction) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Deep copy the actions to prevent external mutation
	actionsCopy := make([]model.PlayerAction, len(actions))
	for i, action := range actions {
		actionsCopy[i] = *action.DeepCopy()
	}

	oldActionsCount := len(player.Actions)
	player.Actions = actionsCopy

	log.Info("⚡ Player actions updated",
		zap.Int("old_actions_count", oldActionsCount),
		zap.Int("new_actions_count", len(actions)))
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
func (r *PlayerRepositoryImpl) RemoveCardFromHand(ctx context.Context, gameID, playerID string, cardID string) error {
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

// UpdateSelectStartingCardsPhase updates the starting card selection phase for a player
func (r *PlayerRepositoryImpl) UpdateSelectStartingCardsPhase(ctx context.Context, gameID, playerID string, selectStartingCardPhase *model.SelectStartingCardsPhase) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Deep copy the starting card phase to prevent external mutation
	if selectStartingCardPhase != nil {
		cardsCopy := make([]string, len(selectStartingCardPhase.AvailableCards))
		copy(cardsCopy, selectStartingCardPhase.AvailableCards)

		player.SelectStartingCardsPhase = &model.SelectStartingCardsPhase{
			AvailableCards: cardsCopy,
		}

		log.Info("Starting card selection phase updated for player", zap.Int("card_count", len(cardsCopy)))
	} else {
		player.SelectStartingCardsPhase = nil
		log.Info("Starting card selection phase cleared for player")
	}

	return nil
}

// SetStartingCardsSelectionComplete marks the starting card selection phase as complete for a player
func (r *PlayerRepositoryImpl) SetStartingCardsSelectionComplete(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	if player.SelectStartingCardsPhase == nil {
		return fmt.Errorf("select starting cards phase not initialized for player %s in game %s", playerID, gameID)
	}

	// Mark selection as complete
	player.SelectStartingCardsPhase.SelectionComplete = true

	log.Info("Player completed starting card selection phase")

	return nil
}

// UpdateProductionPhase updates the production phase for a player
func (r *PlayerRepositoryImpl) UpdateProductionPhase(ctx context.Context, gameID, playerID string, productionPhase *model.ProductionPhase) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		log.Error("Failed to update production phase: "+err.Error(), zap.String("game_id", gameID), zap.String("player_id", playerID))
		return err
	}

	// Deep copy the production phase to prevent external mutation
	if productionPhase != nil {
		cardsCopy := make([]string, len(productionPhase.AvailableCards))
		copy(cardsCopy, productionPhase.AvailableCards)

		player.ProductionPhase = productionPhase.DeepCopy()

		log.Info("Production phase state updated for player", zap.Int("card_count", len(cardsCopy)), zap.Bool("complete", productionPhase.SelectionComplete))
	} else {
		player.ProductionPhase = nil
		log.Info("Production phase state cleared for player")
	}

	return nil
}

// SetProductionCardsSelectionComplete marks the production phase as complete for a player
func (r *PlayerRepositoryImpl) SetProductionCardsSelectionComplete(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	if player.ProductionPhase == nil {
		return fmt.Errorf("production phase not initialized for player %s in game %s", playerID, gameID)
	}

	// Mark selection as complete
	player.ProductionPhase.SelectionComplete = true

	log.Info("Player completed production phase selection")

	return nil
}

// UpdatePendingTileSelection updates the pending tile selection for a player
func (r *PlayerRepositoryImpl) UpdatePendingTileSelection(ctx context.Context, gameID, playerID string, selection *model.PendingTileSelection) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Handle nil input properly - set to nil to clear selection
	if selection == nil {
		player.PendingTileSelection = nil
		log.Info("🎯 Pending tile selection cleared for player")
	} else {
		// Create a deep copy to prevent external mutation
		availableHexesCopy := make([]string, len(selection.AvailableHexes))
		copy(availableHexesCopy, selection.AvailableHexes)

		player.PendingTileSelection = &model.PendingTileSelection{
			TileType:       selection.TileType,
			AvailableHexes: availableHexesCopy,
			Source:         selection.Source,
		}

		log.Info("🎯 Pending tile selection updated for player",
			zap.String("tile_type", selection.TileType),
			zap.String("source", selection.Source),
			zap.Int("available_hexes", len(selection.AvailableHexes)))
	}

	return nil
}

// GetPendingTileSelection retrieves the pending tile selection for a player
func (r *PlayerRepositoryImpl) GetPendingTileSelection(ctx context.Context, gameID, playerID string) (*model.PendingTileSelection, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return nil, err
	}

	if player.PendingTileSelection == nil {
		return nil, nil
	}

	// Return a deep copy to prevent external mutation
	availableHexesCopy := make([]string, len(player.PendingTileSelection.AvailableHexes))
	copy(availableHexesCopy, player.PendingTileSelection.AvailableHexes)

	return &model.PendingTileSelection{
		TileType:       player.PendingTileSelection.TileType,
		AvailableHexes: availableHexesCopy,
		Source:         player.PendingTileSelection.Source,
	}, nil
}

// ClearPendingTileSelection clears the pending tile selection for a player
func (r *PlayerRepositoryImpl) ClearPendingTileSelection(ctx context.Context, gameID, playerID string) error {
	return r.UpdatePendingTileSelection(ctx, gameID, playerID, nil)
}

// UpdatePendingTileSelectionQueue updates the pending tile selection queue for a player
func (r *PlayerRepositoryImpl) UpdatePendingTileSelectionQueue(ctx context.Context, gameID, playerID string, queue *model.PendingTileSelectionQueue) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Handle nil input properly - set to nil to clear queue
	if queue == nil {
		player.PendingTileSelectionQueue = nil
		log.Info("🎯 Pending tile selection queue cleared for player")
	} else {
		// Create a deep copy to prevent external mutation
		itemsCopy := make([]string, len(queue.Items))
		copy(itemsCopy, queue.Items)

		player.PendingTileSelectionQueue = &model.PendingTileSelectionQueue{
			Items:  itemsCopy,
			Source: queue.Source,
		}

		log.Info("🎯 Pending tile selection queue updated for player",
			zap.String("source", queue.Source),
			zap.Int("queue_length", len(queue.Items)),
			zap.Strings("queue_items", queue.Items))
	}

	return nil
}

// GetPendingTileSelectionQueue retrieves the pending tile selection queue for a player
func (r *PlayerRepositoryImpl) GetPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) (*model.PendingTileSelectionQueue, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return nil, err
	}

	if player.PendingTileSelectionQueue == nil {
		return nil, nil
	}

	// Return a deep copy to prevent external mutation
	itemsCopy := make([]string, len(player.PendingTileSelectionQueue.Items))
	copy(itemsCopy, player.PendingTileSelectionQueue.Items)

	return &model.PendingTileSelectionQueue{
		Items:  itemsCopy,
		Source: player.PendingTileSelectionQueue.Source,
	}, nil
}

// ClearPendingTileSelectionQueue clears the pending tile selection queue for a player
func (r *PlayerRepositoryImpl) ClearPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) error {
	return r.UpdatePendingTileSelectionQueue(ctx, gameID, playerID, nil)
}

// CreateTileQueue creates and stores a tile placement queue for a player
// Note: This is a pure data operation - processing and validation is handled by the service layer
func (r *PlayerRepositoryImpl) CreateTileQueue(ctx context.Context, gameID, playerID, cardID string, tilePlacements []string) error {
	log := logger.WithGameContext(gameID, playerID)

	if len(tilePlacements) == 0 {
		log.Debug("No tile placements to queue")
		return nil
	}

	log.Info("🎯 Creating tile placement queue",
		zap.String("card_id", cardID),
		zap.Int("total_placements", len(tilePlacements)),
		zap.Strings("placement_queue", tilePlacements))

	// Create the tile placement queue
	queue := &model.PendingTileSelectionQueue{
		Items:  tilePlacements,
		Source: cardID,
	}

	if err := r.UpdatePendingTileSelectionQueue(ctx, gameID, playerID, queue); err != nil {
		return fmt.Errorf("failed to create tile placement queue: %w", err)
	}

	log.Debug("✅ Tile queue created (service layer will handle processing)")
	return nil
}

// ProcessNextTileInQueue pops the next tile type from the queue and returns it
// Returns empty string if queue is empty or doesn't exist
// This is a pure data operation - validation is handled by the service layer
func (r *PlayerRepositoryImpl) ProcessNextTileInQueue(ctx context.Context, gameID, playerID string) (string, error) {
	log := logger.WithGameContext(gameID, playerID)

	// Get current queue
	queue, err := r.GetPendingTileSelectionQueue(ctx, gameID, playerID)
	if err != nil {
		return "", fmt.Errorf("failed to get tile placement queue: %w", err)
	}

	// If no queue exists or queue is empty, nothing to process
	if queue == nil || len(queue.Items) == 0 {
		log.Debug("No tile placements in queue")
		return "", nil
	}

	// Pop the first item from the queue
	nextTileType := queue.Items[0]
	remainingItems := queue.Items[1:]

	log.Info("🎯 Popping next tile from queue",
		zap.String("tile_type", nextTileType),
		zap.String("source", queue.Source),
		zap.Int("remaining_in_queue", len(remainingItems)))

	// Update queue with remaining items or clear if empty
	if len(remainingItems) > 0 {
		updatedQueue := &model.PendingTileSelectionQueue{
			Items:  remainingItems,
			Source: queue.Source,
		}
		if err := r.UpdatePendingTileSelectionQueue(ctx, gameID, playerID, updatedQueue); err != nil {
			return "", fmt.Errorf("failed to update tile placement queue: %w", err)
		}
	} else {
		// This is the last item, clear the queue
		if err := r.ClearPendingTileSelectionQueue(ctx, gameID, playerID); err != nil {
			return "", fmt.Errorf("failed to clear tile placement queue: %w", err)
		}
	}

	log.Debug("✅ Tile popped from queue", zap.String("tile_type", nextTileType))
	return nextTileType, nil
}
