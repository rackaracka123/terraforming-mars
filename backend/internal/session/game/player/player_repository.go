package player

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
)

var log = logger.Get()

// Repository manages player data with event-driven updates
type Repository interface {
	// Create creates a new player in a game
	Create(ctx context.Context, gameID string, player *Player) error

	// GetByID retrieves a player by ID from a specific game
	GetByID(ctx context.Context, gameID string, playerID string) (*Player, error)

	// ListByGameID retrieves all players in a game
	ListByGameID(ctx context.Context, gameID string) ([]*Player, error)

	// UpdateResources updates player resources (event-driven)
	UpdateResources(ctx context.Context, gameID string, playerID string, resources model.Resources) error

	// UpdateConnectionStatus updates player connection status
	UpdateConnectionStatus(ctx context.Context, gameID string, playerID string, isConnected bool) error

	// SetStartingCardsSelection sets the starting cards selection phase for a player
	SetStartingCardsSelection(ctx context.Context, gameID string, playerID string, cardIDs []string, corpIDs []string) error

	// AddCard adds a card to player's hand
	AddCard(ctx context.Context, gameID string, playerID string, cardID string) error

	// SetCorporation sets the player's corporation
	SetCorporation(ctx context.Context, gameID string, playerID string, corporationID string) error

	// CompleteStartingSelection marks the starting selection as complete
	CompleteStartingSelection(ctx context.Context, gameID string, playerID string) error

	// CompleteProductionSelection marks the production selection as complete
	CompleteProductionSelection(ctx context.Context, gameID string, playerID string) error

	// UpdateProduction updates player production
	UpdateProduction(ctx context.Context, gameID string, playerID string, production model.Production) error

	// UpdateSelectStartingCardsPhase updates the starting cards selection phase
	UpdateSelectStartingCardsPhase(ctx context.Context, gameID string, playerID string, phase *SelectStartingCardsPhase) error

	// UpdateProductionPhase updates the production phase
	UpdateProductionPhase(ctx context.Context, gameID string, playerID string, phase *model.ProductionPhase) error

	// UpdateCorporation updates the player's corporation with full card data
	UpdateCorporation(ctx context.Context, gameID string, playerID string, corporation model.Card) error

	// UpdatePaymentSubstitutes updates player payment substitutes
	UpdatePaymentSubstitutes(ctx context.Context, gameID string, playerID string, substitutes []model.PaymentSubstitute) error

	// UpdatePlayerActions updates player actions
	UpdatePlayerActions(ctx context.Context, gameID string, playerID string, actions []model.PlayerAction) error

	// UpdateForcedFirstAction updates player forced first action
	UpdateForcedFirstAction(ctx context.Context, gameID string, playerID string, action *model.ForcedFirstAction) error

	// UpdateRequirementModifiers updates player requirement modifiers
	UpdateRequirementModifiers(ctx context.Context, gameID string, playerID string, modifiers []model.RequirementModifier) error

	// UpdatePlayerEffects updates player active effects
	UpdatePlayerEffects(ctx context.Context, gameID string, playerID string, effects []model.PlayerEffect) error

	// UpdateTerraformRating updates player terraform rating
	UpdateTerraformRating(ctx context.Context, gameID string, playerID string, rating int) error

	// UpdateVictoryPoints updates player victory points
	UpdateVictoryPoints(ctx context.Context, gameID string, playerID string, victoryPoints int) error

	// CreateTileQueue creates a tile placement queue for the player
	CreateTileQueue(ctx context.Context, gameID string, playerID string, cardID string, tileTypes []string) error

	// GetPendingTileSelectionQueue retrieves the pending tile selection queue for a player
	GetPendingTileSelectionQueue(ctx context.Context, gameID string, playerID string) (*model.PendingTileSelectionQueue, error)

	// ProcessNextTileInQueue pops the next tile type from the queue and returns it
	ProcessNextTileInQueue(ctx context.Context, gameID string, playerID string) (string, error)

	// UpdatePendingTileSelection updates the pending tile selection for a player
	UpdatePendingTileSelection(ctx context.Context, gameID string, playerID string, selection *model.PendingTileSelection) error

	// ClearPendingTileSelection clears the pending tile selection for a player
	ClearPendingTileSelection(ctx context.Context, gameID string, playerID string) error

	// UpdatePendingTileSelectionQueue updates the pending tile selection queue
	UpdatePendingTileSelectionQueue(ctx context.Context, gameID string, playerID string, queue *model.PendingTileSelectionQueue) error

	// ClearPendingTileSelectionQueue clears the pending tile selection queue
	ClearPendingTileSelectionQueue(ctx context.Context, gameID string, playerID string) error

	// UpdatePendingCardDrawSelection updates player pending card draw selection
	UpdatePendingCardDrawSelection(ctx context.Context, gameID string, playerID string, selection *model.PendingCardDrawSelection) error

	// UpdateResourceStorage updates player resource storage
	UpdateResourceStorage(ctx context.Context, gameID string, playerID string, storage map[string]int) error

	// RemoveCardFromHand removes a card from the player's hand
	RemoveCardFromHand(ctx context.Context, gameID string, playerID string, cardID string) error

	// UpdatePassed updates player passed status for generation
	UpdatePassed(ctx context.Context, gameID string, playerID string, passed bool) error

	// UpdateAvailableActions updates player available actions count
	UpdateAvailableActions(ctx context.Context, gameID string, playerID string, actions int) error
}

// RepositoryImpl implements the Repository interface with in-memory storage
type RepositoryImpl struct {
	mu       sync.RWMutex
	players  map[string]map[string]*Player // gameID -> playerID -> Player
	eventBus *events.EventBusImpl
}

// NewRepository creates a new player repository
func NewRepository(eventBus *events.EventBusImpl) Repository {
	return &RepositoryImpl{
		players:  make(map[string]map[string]*Player),
		eventBus: eventBus,
	}
}

// Create creates a new player in a game
func (r *RepositoryImpl) Create(ctx context.Context, gameID string, player *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.players[gameID]; !exists {
		r.players[gameID] = make(map[string]*Player)
	}

	if _, exists := r.players[gameID][player.ID]; exists {
		return fmt.Errorf("player %s already exists in game %s", player.ID, gameID)
	}

	r.players[gameID][player.ID] = player

	// Event publishing can be added here if needed

	return nil
}

// GetByID retrieves a player by ID from a specific game
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID string, playerID string) (*Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return nil, &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return nil, &model.NotFoundError{Resource: "player", ID: playerID}
	}

	return player, nil
}

// ListByGameID retrieves all players in a game
func (r *RepositoryImpl) ListByGameID(ctx context.Context, gameID string) ([]*Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return []*Player{}, nil
	}

	result := make([]*Player, 0, len(gamePlayers))
	for _, player := range gamePlayers {
		result = append(result, player)
	}

	return result, nil
}

// UpdateResources updates player resources (event-driven)
func (r *RepositoryImpl) UpdateResources(ctx context.Context, gameID string, playerID string, resources model.Resources) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.Resources = resources

	// Event publishing can be added here if needed
	// For now, simplified for proof of concept

	return nil
}

// UpdateConnectionStatus updates player connection status
func (r *RepositoryImpl) UpdateConnectionStatus(ctx context.Context, gameID string, playerID string, isConnected bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.IsConnected = isConnected

	return nil
}

// SetStartingCardsSelection sets the starting cards selection phase for a player
func (r *RepositoryImpl) SetStartingCardsSelection(ctx context.Context, gameID string, playerID string, cardIDs []string, corpIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.SelectStartingCardsPhase = &SelectStartingCardsPhase{
		AvailableCards:        cardIDs,
		AvailableCorporations: corpIDs,
	}

	return nil
}

// AddCard adds a card to player's hand
func (r *RepositoryImpl) AddCard(ctx context.Context, gameID string, playerID string, cardID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	log.Debug("🃏 BEFORE AddCard",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("card_to_add", cardID),
		zap.Strings("current_cards", player.Cards),
		zap.Int("card_count", len(player.Cards)))

	player.Cards = append(player.Cards, cardID)

	log.Debug("🃏 AFTER AddCard",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("card_added", cardID),
		zap.Strings("current_cards", player.Cards),
		zap.Int("card_count", len(player.Cards)))

	return nil
}

// SetCorporation sets the player's corporation
func (r *RepositoryImpl) SetCorporation(ctx context.Context, gameID string, playerID string, corporationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.CorporationID = corporationID

	return nil
}

// CompleteStartingSelection marks the starting selection as complete and clears the phase
func (r *RepositoryImpl) CompleteStartingSelection(ctx context.Context, gameID string, playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	// Clear the phase entirely - selection is complete and modal should close
	player.SelectStartingCardsPhase = nil

	return nil
}

// CompleteProductionSelection marks the production selection as complete
func (r *RepositoryImpl) CompleteProductionSelection(ctx context.Context, gameID string, playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	if player.ProductionPhase != nil {
		player.ProductionPhase.SelectionComplete = true
	}

	return nil
}

// UpdateProduction updates player production
func (r *RepositoryImpl) UpdateProduction(ctx context.Context, gameID string, playerID string, production model.Production) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.Production = production

	return nil
}

// UpdateSelectStartingCardsPhase updates the starting cards selection phase
func (r *RepositoryImpl) UpdateSelectStartingCardsPhase(ctx context.Context, gameID string, playerID string, phase *SelectStartingCardsPhase) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.SelectStartingCardsPhase = phase

	return nil
}

// UpdateProductionPhase updates the production phase
func (r *RepositoryImpl) UpdateProductionPhase(ctx context.Context, gameID string, playerID string, phase *model.ProductionPhase) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.ProductionPhase = phase

	return nil
}

// UpdateCorporation updates the player's corporation with full card data
func (r *RepositoryImpl) UpdateCorporation(ctx context.Context, gameID string, playerID string, corporation model.Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.Corporation = &corporation
	player.CorporationID = corporation.ID

	return nil
}

// UpdatePaymentSubstitutes updates player payment substitutes
func (r *RepositoryImpl) UpdatePaymentSubstitutes(ctx context.Context, gameID string, playerID string, substitutes []model.PaymentSubstitute) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.PaymentSubstitutes = substitutes

	return nil
}

// UpdatePlayerActions updates player actions
func (r *RepositoryImpl) UpdatePlayerActions(ctx context.Context, gameID string, playerID string, actions []model.PlayerAction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.Actions = actions

	return nil
}

// UpdateForcedFirstAction updates player forced first action
func (r *RepositoryImpl) UpdateForcedFirstAction(ctx context.Context, gameID string, playerID string, action *model.ForcedFirstAction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.ForcedFirstAction = action

	return nil
}

// UpdateRequirementModifiers updates player requirement modifiers
func (r *RepositoryImpl) UpdateRequirementModifiers(ctx context.Context, gameID string, playerID string, modifiers []model.RequirementModifier) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.RequirementModifiers = modifiers

	return nil
}

// UpdatePlayerEffects updates player active effects
func (r *RepositoryImpl) UpdatePlayerEffects(ctx context.Context, gameID string, playerID string, effects []model.PlayerEffect) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.Effects = effects

	return nil
}

// UpdateTerraformRating updates player terraform rating
func (r *RepositoryImpl) UpdateTerraformRating(ctx context.Context, gameID string, playerID string, rating int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.TerraformRating = rating

	return nil
}

// UpdateVictoryPoints updates player victory points
func (r *RepositoryImpl) UpdateVictoryPoints(ctx context.Context, gameID string, playerID string, victoryPoints int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.VictoryPoints = victoryPoints

	return nil
}

// CreateTileQueue creates a tile placement queue for the player
func (r *RepositoryImpl) CreateTileQueue(ctx context.Context, gameID string, playerID string, cardID string, tileTypes []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	// Create tile queue
	if len(tileTypes) > 0 {
		player.PendingTileSelectionQueue = &model.PendingTileSelectionQueue{
			Items:  tileTypes,
			Source: cardID,
		}
	}

	return nil
}

// GetPendingTileSelectionQueue retrieves the pending tile selection queue for a player
func (r *RepositoryImpl) GetPendingTileSelectionQueue(ctx context.Context, gameID string, playerID string) (*model.PendingTileSelectionQueue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return nil, &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return nil, &model.NotFoundError{Resource: "player", ID: playerID}
	}

	if player.PendingTileSelectionQueue == nil {
		return nil, nil
	}

	// Return a copy to prevent external mutation
	itemsCopy := make([]string, len(player.PendingTileSelectionQueue.Items))
	copy(itemsCopy, player.PendingTileSelectionQueue.Items)

	return &model.PendingTileSelectionQueue{
		Items:  itemsCopy,
		Source: player.PendingTileSelectionQueue.Source,
	}, nil
}

// ProcessNextTileInQueue pops the next tile type from the queue and returns it
func (r *RepositoryImpl) ProcessNextTileInQueue(ctx context.Context, gameID string, playerID string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return "", &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return "", &model.NotFoundError{Resource: "player", ID: playerID}
	}

	// If no queue exists or queue is empty, nothing to process
	if player.PendingTileSelectionQueue == nil || len(player.PendingTileSelectionQueue.Items) == 0 {
		log.Debug("No tile placements in queue",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		return "", nil
	}

	// Pop the first item from the queue
	nextTileType := player.PendingTileSelectionQueue.Items[0]
	remainingItems := player.PendingTileSelectionQueue.Items[1:]

	log.Info("🎯 Popping next tile from queue",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("tile_type", nextTileType),
		zap.String("source", player.PendingTileSelectionQueue.Source),
		zap.Int("remaining_in_queue", len(remainingItems)))

	// Update queue with remaining items or clear if empty
	if len(remainingItems) > 0 {
		player.PendingTileSelectionQueue = &model.PendingTileSelectionQueue{
			Items:  remainingItems,
			Source: player.PendingTileSelectionQueue.Source,
		}
	} else {
		// This is the last item, clear the queue
		player.PendingTileSelectionQueue = nil
	}

	log.Debug("✅ Tile popped from queue",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("tile_type", nextTileType))

	return nextTileType, nil
}

// UpdatePendingTileSelection updates the pending tile selection for a player
func (r *RepositoryImpl) UpdatePendingTileSelection(ctx context.Context, gameID string, playerID string, selection *model.PendingTileSelection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.PendingTileSelection = selection

	return nil
}

// ClearPendingTileSelection clears the pending tile selection for a player
func (r *RepositoryImpl) ClearPendingTileSelection(ctx context.Context, gameID string, playerID string) error {
	return r.UpdatePendingTileSelection(ctx, gameID, playerID, nil)
}

// UpdatePendingTileSelectionQueue updates the pending tile selection queue
func (r *RepositoryImpl) UpdatePendingTileSelectionQueue(ctx context.Context, gameID string, playerID string, queue *model.PendingTileSelectionQueue) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.PendingTileSelectionQueue = queue

	return nil
}

// ClearPendingTileSelectionQueue clears the pending tile selection queue
func (r *RepositoryImpl) ClearPendingTileSelectionQueue(ctx context.Context, gameID string, playerID string) error {
	return r.UpdatePendingTileSelectionQueue(ctx, gameID, playerID, nil)
}

// UpdatePendingCardDrawSelection updates player pending card draw selection
func (r *RepositoryImpl) UpdatePendingCardDrawSelection(ctx context.Context, gameID string, playerID string, selection *model.PendingCardDrawSelection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.PendingCardDrawSelection = selection

	return nil
}

// UpdateResourceStorage updates player resource storage
func (r *RepositoryImpl) UpdateResourceStorage(ctx context.Context, gameID string, playerID string, storage map[string]int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.ResourceStorage = storage

	return nil
}

// RemoveCardFromHand removes a card from the player's hand
func (r *RepositoryImpl) RemoveCardFromHand(ctx context.Context, gameID string, playerID string, cardID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	// Find and remove the card from the player's hand
	for i, id := range player.Cards {
		if id == cardID {
			player.Cards = append(player.Cards[:i], player.Cards[i+1:]...)
			// Add to played cards
			player.PlayedCards = append(player.PlayedCards, cardID)
			return nil
		}
	}

	return fmt.Errorf("card %s not found in player's hand", cardID)
}

// UpdatePassed updates player passed status for generation
func (r *RepositoryImpl) UpdatePassed(ctx context.Context, gameID string, playerID string, passed bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.Passed = passed

	log.Debug("✅ Player passed status updated",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Bool("passed", passed))

	return nil
}

// UpdateAvailableActions updates player available actions count
func (r *RepositoryImpl) UpdateAvailableActions(ctx context.Context, gameID string, playerID string, actions int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return &model.NotFoundError{Resource: "player", ID: playerID}
	}

	player.AvailableActions = actions

	log.Debug("✅ Player available actions updated",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Int("available_actions", actions))

	return nil
}
