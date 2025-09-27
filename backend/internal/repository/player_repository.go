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
	UpdateConnectionStatus(ctx context.Context, gameID, playerID string, isConnected bool) error
	UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error
	UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error
	UpdateVictoryPoints(ctx context.Context, gameID, playerID string, points int) error
	UpdateEffects(ctx context.Context, gameID, playerID string, effects []model.PlayerEffect) error
	UpdatePlayerActions(ctx context.Context, gameID, playerID string, actions []model.PlayerAction) error
	AddCard(ctx context.Context, gameID, playerID string, cardID string) error
	RemoveCard(ctx context.Context, gameID, playerID string, cardID string) error
	PlayCard(ctx context.Context, gameID, playerID string, cardID string) error

	// Card selection methods - using nullable ProductionPhase
	SetCardSelection(ctx context.Context, gameID, playerID string, productionPhase *model.ProductionPhase) error
	SetCardSelectionComplete(ctx context.Context, gameID, playerID string) error
	GetCardSelection(ctx context.Context, gameID, playerID string) (*model.ProductionPhase, error)
	ClearCardSelection(ctx context.Context, gameID, playerID string) error

	// Starting card selection methods
	SetStartingSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error
	SetHasSelectedStartingCards(ctx context.Context, gameID, playerID string, value bool) error
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
		// Copy affected tags slice
		affectedTagsCopy := make([]model.CardTag, len(effect.AffectedTags))
		copy(affectedTagsCopy, effect.AffectedTags)

		effectsCopy[i] = model.PlayerEffect{
			Type:         effect.Type,
			Amount:       effect.Amount,
			AffectedTags: affectedTagsCopy,
		}
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

// SetCardSelection sets the card selection state for a player
func (r *PlayerRepositoryImpl) SetCardSelection(ctx context.Context, gameID, playerID string, productionPhase *model.ProductionPhase) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Deep copy the production phase to prevent external mutation
	if productionPhase != nil {
		cardsCopy := make([]model.Card, len(productionPhase.AvailableCards))
		copy(cardsCopy, productionPhase.AvailableCards)

		player.ProductionSelection = &model.ProductionPhase{
			AvailableCards:    cardsCopy,
			SelectionComplete: productionPhase.SelectionComplete,
		}

		log.Info("Production phase state set for player", zap.Int("card_count", len(cardsCopy)), zap.Bool("complete", productionPhase.SelectionComplete))
	} else {
		player.ProductionSelection = nil
		log.Info("Production phase state cleared for player")
	}

	return nil
}

// SetCardSelectionComplete marks the card selection as complete for a player
func (r *PlayerRepositoryImpl) SetCardSelectionComplete(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	if player.ProductionSelection == nil {
		return fmt.Errorf("player has no card selection data to complete")
	}

	// Mark selection as complete but keep the data
	player.ProductionSelection.SelectionComplete = true
	log.Info("Card selection marked as complete for player")

	return nil
}

// GetCardSelection returns the card selection state for a player
func (r *PlayerRepositoryImpl) GetCardSelection(ctx context.Context, gameID, playerID string) (*model.ProductionPhase, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return nil, err
	}

	if player.ProductionSelection == nil {
		return nil, nil
	}

	// Return a deep copy to prevent external mutation
	cardsCopy := make([]model.Card, len(player.ProductionSelection.AvailableCards))
	copy(cardsCopy, player.ProductionSelection.AvailableCards)

	return &model.ProductionPhase{
		AvailableCards:    cardsCopy,
		SelectionComplete: player.ProductionSelection.SelectionComplete,
	}, nil
}

// ClearCardSelection clears the card selection data for a player
func (r *PlayerRepositoryImpl) ClearCardSelection(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.ProductionSelection = nil

	log.Info("Card selection data cleared for player")

	return nil
}

// SetStartingSelection sets the starting card IDs for a player
func (r *PlayerRepositoryImpl) SetStartingSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Handle nil input properly - set to nil instead of empty slice
	if cardIDs == nil {
		player.StartingSelection = nil
	} else {
		// Create a copy of the card IDs to prevent external mutation
		cardIDsCopy := make([]string, len(cardIDs))
		copy(cardIDsCopy, cardIDs)
		player.StartingSelection = cardIDsCopy
	}

	cardCount := 0
	if cardIDs != nil {
		cardCount = len(cardIDs)
	}
	log.Info("🃏 Starting cards set for player", zap.Int("card_count", cardCount))

	return nil
}

// SetHasSelectedStartingCards sets whether the player has completed starting card selection
func (r *PlayerRepositoryImpl) SetHasSelectedStartingCards(ctx context.Context, gameID, playerID string, value bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.HasSelectedStartingCards = value

	log.Info("🃏 Starting card selection completion flag set for player", zap.Bool("has_selected", value))

	return nil
}
