package player

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Repository provides clean CRUD operations and granular updates for players
type Repository interface {
	Create(ctx context.Context, gameID string, player Player) error
	GetByID(ctx context.Context, gameID, playerID string) (Player, error)
	Delete(ctx context.Context, gameID, playerID string) error
	ListByGameID(ctx context.Context, gameID string) ([]Player, error)

	UpdateResources(ctx context.Context, gameID, playerID string, resources Resources) error
	UpdateProduction(ctx context.Context, gameID, playerID string, production Production) error
	UpdateTerraformRating(ctx context.Context, gameID, playerID string, rating int) error
	UpdateCorporation(ctx context.Context, gameID, playerID string, corporation Card) error
	UpdateConnectionStatus(ctx context.Context, gameID, playerID string, isConnected bool) error
	UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error
	UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error
	UpdateVictoryPoints(ctx context.Context, gameID, playerID string, points int) error
	UpdatePlayerActions(ctx context.Context, gameID, playerID string, actions []PlayerAction) error
	UpdatePlayerEffects(ctx context.Context, gameID, playerID string, effects []PlayerEffect) error

	// Card management methods (Living Card Instance Pattern)
	AddCard(ctx context.Context, gameID, playerID string, card Card) error
	AddCards(ctx context.Context, gameID, playerID string, cards []Card) error
	RemoveCard(ctx context.Context, gameID, playerID string, cardID string) error
	RemoveCardFromHand(ctx context.Context, gameID, playerID string, cardID, cardName string, cardType CardType) error
	AddPlayedCard(ctx context.Context, gameID, playerID string, card Card) error
	UpdateCards(ctx context.Context, gameID, playerID string, cards []Card) error
	UpdatePlayedCards(ctx context.Context, gameID, playerID string, cards []Card) error

	// Modifier application methods
	ApplyModifiersToPlayerCards(ctx context.Context, gameID, playerID string, sourceCardID string, tag card.CardTag, amount int) error

	UpdateSelectStartingCardsPhase(ctx context.Context, gameID, playerID string, selectStartingCardPhase *SelectStartingCardsPhase) error
	SetStartingCardsSelectionComplete(ctx context.Context, gameID, playerID string) error
	ClearStartingCardsPhase(ctx context.Context, gameID, playerID string) error

	UpdateProductionPhase(ctx context.Context, gameID, playerID string, productionPhase *ProductionPhase) error
	SetProductionCardsSelectionComplete(ctx context.Context, gameID, playerID string) error

	// Tile selection methods
	UpdatePendingTileSelection(ctx context.Context, gameID, playerID string, selection *PendingTileSelection) error
	GetPendingTileSelection(ctx context.Context, gameID, playerID string) (*PendingTileSelection, error)
	ClearPendingTileSelection(ctx context.Context, gameID, playerID string) error

	// Tile queue methods
	UpdatePendingTileSelectionQueue(ctx context.Context, gameID, playerID string, queue *PendingTileSelectionQueue) error
	GetPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) (*PendingTileSelectionQueue, error)
	ClearPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) error

	// Tile queue operations
	CreateTileQueue(ctx context.Context, gameID, playerID, cardID string, tilePlacements []string) error
	ProcessNextTileInQueue(ctx context.Context, gameID, playerID string) (string, error)

	// Card selection methods
	UpdatePendingCardSelection(ctx context.Context, gameID, playerID string, selection *PendingCardSelection) error
	GetPendingCardSelection(ctx context.Context, gameID, playerID string) (*PendingCardSelection, error)
	ClearPendingCardSelection(ctx context.Context, gameID, playerID string) error

	// Card draw/peek selection methods
	UpdatePendingCardDrawSelection(ctx context.Context, gameID, playerID string, selection *PendingCardDrawSelection) error
	GetPendingCardDrawSelection(ctx context.Context, gameID, playerID string) (*PendingCardDrawSelection, error)
	ClearPendingCardDrawSelection(ctx context.Context, gameID, playerID string) error

	// Forced first action methods
	UpdateForcedFirstAction(ctx context.Context, gameID, playerID string, action *ForcedFirstAction) error
	GetForcedFirstAction(ctx context.Context, gameID, playerID string) (*ForcedFirstAction, error)
	MarkForcedFirstActionComplete(ctx context.Context, gameID, playerID string) error
	ClearForcedFirstAction(ctx context.Context, gameID, playerID string) error

	// Resource storage methods
	UpdateResourceStorage(ctx context.Context, gameID, playerID string, resourceStorage map[string]int) error

	// Payment substitute methods
	UpdatePaymentSubstitutes(ctx context.Context, gameID, playerID string, substitutes []PaymentSubstitute) error

	// Resource operation methods
	DeductResources(ctx context.Context, gameID, playerID string, cost domain.ResourceSet) error
	AddResources(ctx context.Context, gameID, playerID string, resources domain.ResourceSet) error
	AddProduction(ctx context.Context, gameID, playerID string, production domain.ResourceSet) error
	CanAfford(ctx context.Context, gameID, playerID string, cost domain.ResourceSet) (bool, error)
	GetAvailableActions(ctx context.Context, gameID, playerID string) (int, error)
}

// RepositoryImpl implements Repository with in-memory storage
type RepositoryImpl struct {
	// Map of gameID -> map of playerID -> Player
	players  map[string]map[string]*Player
	mutex    sync.RWMutex
	eventBus *events.EventBusImpl

	// Feature repositories (scoped by gameID_playerID key)
	// TODO: Refactor resources management after feature cleanup
	// resourcesRepos  map[string]resources.Repository
	// productionRepos map[string]production.Repository
	tileQueueRepos  map[string]tiles.TileQueueRepository
	playerTurnRepos map[string]turn.PlayerTurnRepository
}

// NewRepository creates a new player repository
func NewRepository(eventBus *events.EventBusImpl) Repository {
	return &RepositoryImpl{
		players:  make(map[string]map[string]*Player),
		eventBus: eventBus,
		// resourcesRepos:  make(map[string]resources.Repository),
		// productionRepos: make(map[string]production.Repository),
		tileQueueRepos:  make(map[string]tiles.TileQueueRepository),
		playerTurnRepos: make(map[string]turn.PlayerTurnRepository),
	}
}

// getPlayerKey returns a unique key for player-scoped repositories
func (r *RepositoryImpl) getPlayerKey(gameID, playerID string) string {
	return gameID + "_" + playerID
}

// Create adds a player to a game
func (r *RepositoryImpl) Create(ctx context.Context, gameID string, player Player) error {
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
		r.players[gameID] = make(map[string]*Player)
	}

	// Check if player already exists
	if _, exists := r.players[gameID][player.ID]; exists {
		log.Error("Player already exists in game")
		return fmt.Errorf("player with ID %s already exists in game %s", player.ID, gameID)
	}

	// Create feature repositories with initial state
	playerKey := r.getPlayerKey(gameID, player.ID)

	// Tile queue repository (empty initially)
	tileQueueRepo := tiles.NewTileQueueRepository(nil, nil)
	r.tileQueueRepos[playerKey] = tileQueueRepo

	// Player turn repository (not passed, 0 actions initially)
	playerTurnRepo := turn.NewPlayerTurnRepository(false, 0)
	r.playerTurnRepos[playerKey] = playerTurnRepo

	// Store a copy to prevent external mutation
	playerCopy := player.DeepCopy()
	r.players[gameID][player.ID] = playerCopy

	log.Debug("Player added to game with feature services", zap.String("player_name", player.Name))

	return nil
}

// GetByID retrieves a player by game and player ID
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID, playerID string) (Player, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return Player{}, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	if gameID == "" {
		return Player{}, fmt.Errorf("game ID cannot be empty")
	}

	if playerID == "" {
		return Player{}, fmt.Errorf("player ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return Player{}, fmt.Errorf("no players found for game %s", gameID)
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return Player{}, &game.NotFoundError{Resource: "player", ID: playerID}
	}

	// Return a copy to prevent external mutation
	return *player.DeepCopy(), nil
}

// Delete removes a player from a game
func (r *RepositoryImpl) Delete(ctx context.Context, gameID, playerID string) error {
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
func (r *RepositoryImpl) ListByGameID(ctx context.Context, gameID string) ([]Player, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	gamePlayers, exists := r.players[gameID]
	if !exists {
		return make([]Player, 0), nil
	}

	players := make([]Player, 0, len(gamePlayers))
	for _, player := range gamePlayers {
		players = append(players, *player.DeepCopy())
	}

	return players, nil
}

// UpdateResources updates a player's resources
func (r *RepositoryImpl) UpdateResources(ctx context.Context, gameID, playerID string, res Resources) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	// Verify player exists
	_, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Update resources directly on player
	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.Resources = res

	log.Info("Player resources updated")

	// Publish ResourcesChangedEvent
	events.Publish(r.eventBus, events.ResourcesChangedEvent{
		GameID:   gameID,
		PlayerID: playerID,
	})

	return nil
}

// UpdateProduction updates a player's production
func (r *RepositoryImpl) UpdateProduction(ctx context.Context, gameID, playerID string, prod Production) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	// Update production directly on player
	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.Production = prod

	log.Info("Player production updated")

	// Publish ProductionChangedEvent
	events.Publish(r.eventBus, events.ProductionChangedEvent{
		GameID:   gameID,
		PlayerID: playerID,
	})

	return nil
}

// UpdateTerraformRating updates a player's terraform rating
func (r *RepositoryImpl) UpdateTerraformRating(ctx context.Context, gameID, playerID string, rating int) error {
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

	// Publish terraform rating changed event
	if r.eventBus != nil && oldTR != rating {
		events.Publish(r.eventBus, events.TerraformRatingChangedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			OldRating: oldTR,
			NewRating: rating,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateCorporation updates a player's corporation
func (r *RepositoryImpl) UpdateCorporation(ctx context.Context, gameID, playerID string, corporation Card) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	var oldCorporationName string
	if player.Corporation != nil {
		oldCorporationName = player.Corporation.Name
	}
	player.Corporation = &corporation

	log.Info("Player corporation updated", zap.String("old_corp", oldCorporationName), zap.String("new_corp", corporation.Name))

	// Publish CorporationSelectedEvent
	events.Publish(r.eventBus, events.CorporationSelectedEvent{
		GameID:        gameID,
		PlayerID:      playerID,
		CorporationID: corporation.ID,
	})

	return nil
}

// UpdateConnectionStatus updates a player's connection status
func (r *RepositoryImpl) UpdateConnectionStatus(ctx context.Context, gameID, playerID string, isConnected bool) error {
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
func (r *RepositoryImpl) UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	// Verify player exists
	_, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Update via feature repository
	playerKey := r.getPlayerKey(gameID, playerID)
	playerTurnRepo, exists := r.playerTurnRepos[playerKey]
	if !exists {
		return fmt.Errorf("player turn repository not found for player %s in game %s", playerID, gameID)
	}

	if err := playerTurnRepo.SetPassed(ctx, passed); err != nil {
		return fmt.Errorf("failed to update passed status: %w", err)
	}

	log.Info("Player passed status updated", zap.Bool("passed", passed))

	return nil
}

// UpdateAvailableActions updates a player's available actions count
func (r *RepositoryImpl) UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	// Verify player exists
	_, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Update via feature repository
	playerKey := r.getPlayerKey(gameID, playerID)
	playerTurnRepo, exists := r.playerTurnRepos[playerKey]
	if !exists {
		return fmt.Errorf("player turn repository not found for player %s in game %s", playerID, gameID)
	}

	if err := playerTurnRepo.SetAvailableActions(ctx, actions); err != nil {
		return fmt.Errorf("failed to update available actions: %w", err)
	}

	log.Debug("Player available actions updated", zap.Int("actions", actions))

	return nil
}

// UpdateVictoryPoints updates a player's victory points
func (r *RepositoryImpl) UpdateVictoryPoints(ctx context.Context, gameID, playerID string, points int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.VictoryPoints = points

	log.Info("Player victory points updated", zap.Int("points", points))

	// Publish VictoryPointsChangedEvent
	events.Publish(r.eventBus, events.VictoryPointsChangedEvent{
		GameID:    gameID,
		PlayerID:  playerID,
		NewPoints: points,
		Source:    "direct_update",
	})

	return nil
}

// UpdatePlayerActions updates a player's available actions
func (r *RepositoryImpl) UpdatePlayerActions(ctx context.Context, gameID, playerID string, actions []PlayerAction) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Deep copy the actions to prevent external mutation
	actionsCopy := make([]PlayerAction, len(actions))
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

// UpdatePlayerEffects updates a player's active passive effects
func (r *RepositoryImpl) UpdatePlayerEffects(ctx context.Context, gameID, playerID string, effects []PlayerEffect) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Deep copy the effects to prevent external mutation
	effectsCopy := make([]PlayerEffect, len(effects))
	for i, effect := range effects {
		effectsCopy[i] = *effect.DeepCopy()
	}

	oldEffectsCount := len(player.Effects)
	player.Effects = effectsCopy

	log.Info("✨ Player effects updated",
		zap.Int("old_effects_count", oldEffectsCount),
		zap.Int("new_effects_count", len(effects)))

	// Publish PlayerEffectAddedEvent (if effects were added)
	if len(effects) > oldEffectsCount {
		events.Publish(r.eventBus, events.PlayerEffectAddedEvent{
			GameID:   gameID,
			PlayerID: playerID,
		})
	}

	return nil
}

// AddCard adds a single card instance to player's hand
func (r *RepositoryImpl) AddCard(ctx context.Context, gameID, playerID string, card Card) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Check if card already exists in hand
	for _, existingCard := range player.Cards {
		if existingCard.ID == card.ID {
			return fmt.Errorf("card %s already exists in player %s hand", card.ID, playerID)
		}
	}

	player.Cards = append(player.Cards, card)

	log.Info("Card added to player hand", zap.String("card_id", card.ID))

	// Publish CardAddedToHandEvent
	events.Publish(r.eventBus, events.CardAddedToHandEvent{
		GameID:   gameID,
		PlayerID: playerID,
		CardID:   card.ID,
	})

	return nil
}

// AddCards adds multiple card instances to player's hand
func (r *RepositoryImpl) AddCards(ctx context.Context, gameID, playerID string, cards []Card) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.Cards = append(player.Cards, cards...)

	log.Info("Cards added to player hand", zap.Int("count", len(cards)))

	return nil
}

// UpdateCards replaces all cards in player's hand
func (r *RepositoryImpl) UpdateCards(ctx context.Context, gameID, playerID string, cards []Card) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.Cards = cards

	log.Info("Player hand cards updated", zap.Int("count", len(cards)))

	return nil
}

// UpdatePlayedCards replaces all played cards for a player
func (r *RepositoryImpl) UpdatePlayedCards(ctx context.Context, gameID, playerID string, cards []Card) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.PlayedCards = cards

	log.Info("Player played cards updated", zap.Int("count", len(cards)))

	return nil
}

// ApplyModifiersToPlayerCards applies cost modifiers to all matching cards in player's hand
// This is used when a discount card is played
func (r *RepositoryImpl) ApplyModifiersToPlayerCards(ctx context.Context, gameID, playerID string, sourceCardID string, tag card.CardTag, amount int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Apply modifier to all cards in hand that have the matching tag
	modifiedCount := 0
	for i := range player.Cards {
		// Check if card has the target tag
		hasTag := false
		for _, cardTag := range player.Cards[i].Tags {
			if cardTag == tag {
				hasTag = true
				break
			}
		}

		if hasTag {
			// Add the modifier to this card
			modifier := card.CostModifier{
				SourceCardID: sourceCardID,
				Amount:       amount,
				Tag:          string(tag),
			}
			player.Cards[i].Modifiers = append(player.Cards[i].Modifiers, modifier)
			modifiedCount++
		}
	}

	log.Info("Applied cost modifiers to player cards",
		zap.String("source_card_id", sourceCardID),
		zap.String("tag", string(tag)),
		zap.Int("amount", amount),
		zap.Int("cards_modified", modifiedCount))

	return nil
}

// RemoveCard removes a card from a player's hand
func (r *RepositoryImpl) RemoveCard(ctx context.Context, gameID, playerID string, cardID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Find and remove card
	for i, existingCard := range player.Cards {
		if existingCard.ID == cardID {
			player.Cards = append(player.Cards[:i], player.Cards[i+1:]...)
			log.Info("Card removed from player hand", zap.String("card_id", cardID))
			return nil
		}
	}

	return fmt.Errorf("card %s not found in player %s hand", cardID, playerID)
}

// PlayCard moves a card from hand to played cards
func (r *RepositoryImpl) RemoveCardFromHand(ctx context.Context, gameID, playerID string, cardID, cardName string, cardType CardType) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Find and remove card from hand
	var removedCard *Card
	for i, existingCard := range player.Cards {
		if existingCard.ID == cardID {
			removedCard = &player.Cards[i]
			player.Cards = append(player.Cards[:i], player.Cards[i+1:]...)
			break
		}
	}

	if removedCard == nil {
		return fmt.Errorf("card %s not found in player %s hand", cardID, playerID)
	}

	// Add to played cards
	player.PlayedCards = append(player.PlayedCards, *removedCard)

	// Publish CardPlayedEvent
	if r.eventBus != nil {
		events.Publish(r.eventBus, events.CardPlayedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			CardID:    cardID,
			CardName:  cardName,
			CardType:  string(cardType),
			Timestamp: time.Now(),
		})
	}

	log.Info("🃏 Card played", zap.String("card_id", cardID), zap.String("card_name", cardName), zap.String("card_type", string(cardType)))

	return nil
}

// AddPlayedCard adds a card to a player's played cards
func (r *RepositoryImpl) AddPlayedCard(ctx context.Context, gameID, playerID string, card Card) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Add to played cards
	player.PlayedCards = append(player.PlayedCards, card)

	log.Info("🃏 Card added to played cards", zap.String("card_id", card.ID))

	return nil
}

// getPlayerUnsafe retrieves a player without acquiring locks (assumes lock is held)
func (r *RepositoryImpl) getPlayerUnsafe(gameID, playerID string) (*Player, error) {
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
func (r *RepositoryImpl) UpdateSelectStartingCardsPhase(ctx context.Context, gameID, playerID string, selectStartingCardPhase *SelectStartingCardsPhase) error {
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

		corpsCopy := make([]string, len(selectStartingCardPhase.AvailableCorporations))
		copy(corpsCopy, selectStartingCardPhase.AvailableCorporations)

		player.SelectStartingCardsPhase = &SelectStartingCardsPhase{
			AvailableCards:        cardsCopy,
			AvailableCorporations: corpsCopy,
		}

		log.Info("Starting card selection phase updated for player",
			zap.Int("card_count", len(cardsCopy)),
			zap.Int("corporation_count", len(corpsCopy)))
	} else {
		player.SelectStartingCardsPhase = nil
		log.Info("Starting card selection phase cleared for player")
	}

	return nil
}

// SetStartingCardsSelectionComplete marks the starting card selection phase as complete for a player
func (r *RepositoryImpl) SetStartingCardsSelectionComplete(ctx context.Context, gameID, playerID string) error {
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

// ClearStartingCardsPhase removes the starting cards selection phase from a player
func (r *RepositoryImpl) ClearStartingCardsPhase(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.SelectStartingCardsPhase = nil

	log.Info("✅ Cleared starting card selection phase for player")

	return nil
}

// UpdateProductionPhase updates the production phase for a player
func (r *RepositoryImpl) UpdateProductionPhase(ctx context.Context, gameID, playerID string, productionPhase *ProductionPhase) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	_, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		log.Error("Failed to update production phase: "+err.Error(), zap.String("game_id", gameID), zap.String("player_id", playerID))
		return err
	}

	// TODO: Production phase state management needs refactoring after feature cleanup
	// This was previously handled by the deleted production feature
	log.Warn("UpdateProductionPhaseState not yet implemented after production feature removal")
	return fmt.Errorf("production phase state management not implemented")
}

// SetProductionCardsSelectionComplete marks the production phase as complete for a player
func (r *RepositoryImpl) SetProductionCardsSelectionComplete(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	_, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// TODO: Production phase state management needs refactoring after feature cleanup
	// This was previously handled by the deleted production feature
	log.Warn("SetProductionCardsSelectionComplete not yet implemented after production feature removal")
	return fmt.Errorf("production phase state management not implemented")
}

// UpdatePendingTileSelection updates the pending tile selection for a player
func (r *RepositoryImpl) UpdatePendingTileSelection(ctx context.Context, gameID, playerID string, selection *PendingTileSelection) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	_, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	playerKey := r.getPlayerKey(gameID, playerID)
	tileQueueRepo, exists := r.tileQueueRepos[playerKey]
	if !exists {
		return fmt.Errorf("tile queue repository not found for player %s in game %s", playerID, gameID)
	}

	// Handle nil input properly - clear selection
	if selection == nil {
		if err := tileQueueRepo.ClearPendingSelection(ctx); err != nil {
			return fmt.Errorf("failed to clear pending selection: %w", err)
		}
		log.Info("🎯 Pending tile selection cleared for player")
	} else {
		// Convert PendingTileSelection to tiles.PendingTileSelection
		tileSelection := &tiles.PendingTileSelection{
			TileType:       selection.TileType,
			Source:         selection.Source,
			AvailableHexes: make([]string, len(selection.AvailableHexes)),
		}
		copy(tileSelection.AvailableHexes, selection.AvailableHexes)

		if err := tileQueueRepo.SetPendingSelection(ctx, tileSelection); err != nil {
			return fmt.Errorf("failed to set pending selection: %w", err)
		}

		log.Info("🎯 Pending tile selection updated for player",
			zap.String("tile_type", selection.TileType),
			zap.String("source", selection.Source),
			zap.Int("available_hexes", len(selection.AvailableHexes)))
	}

	return nil
}

// GetPendingTileSelection retrieves the pending tile selection for a player
func (r *RepositoryImpl) GetPendingTileSelection(ctx context.Context, gameID, playerID string) (*PendingTileSelection, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return nil, err
	}

	playerKey := r.getPlayerKey(gameID, playerID)
	tileQueueRepo, exists := r.tileQueueRepos[playerKey]
	if !exists {
		return nil, fmt.Errorf("tile queue repository not found for player %s in game %s", playerID, gameID)
	}

	tileSelection, err := tileQueueRepo.GetPendingSelection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending selection: %w", err)
	}

	if tileSelection == nil {
		return nil, nil
	}

	// Convert tiles.PendingTileSelection to PendingTileSelection
	return &PendingTileSelection{
		TileType:       tileSelection.TileType,
		Source:         tileSelection.Source,
		AvailableHexes: append([]string{}, tileSelection.AvailableHexes...),
	}, nil
}

// ClearPendingTileSelection clears the pending tile selection for a player
func (r *RepositoryImpl) ClearPendingTileSelection(ctx context.Context, gameID, playerID string) error {
	return r.UpdatePendingTileSelection(ctx, gameID, playerID, nil)
}

// UpdatePendingTileSelectionQueue updates the pending tile selection queue for a player
func (r *RepositoryImpl) UpdatePendingTileSelectionQueue(ctx context.Context, gameID, playerID string, queue *PendingTileSelectionQueue) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	_, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	playerKey := r.getPlayerKey(gameID, playerID)
	tileQueueRepo, exists := r.tileQueueRepos[playerKey]
	if !exists {
		return fmt.Errorf("tile queue repository not found for player %s in game %s", playerID, gameID)
	}

	// Handle nil input properly - clear queue
	if queue == nil {
		if err := tileQueueRepo.ClearQueue(ctx); err != nil {
			return fmt.Errorf("failed to clear queue: %w", err)
		}
		log.Info("🎯 Pending tile selection queue cleared for player")
	} else {
		// Convert PendingTileSelectionQueue to tiles.PendingTileSelectionQueue
		tileQueue := &tiles.PendingTileSelectionQueue{
			Items:  make([]string, len(queue.Items)),
			Source: queue.Source,
		}
		copy(tileQueue.Items, queue.Items)

		if err := tileQueueRepo.SetQueue(ctx, tileQueue); err != nil {
			return fmt.Errorf("failed to set queue: %w", err)
		}

		log.Info("🎯 Pending tile selection queue updated for player",
			zap.String("source", queue.Source),
			zap.Int("queue_length", len(queue.Items)),
			zap.Strings("queue_items", queue.Items))
	}

	return nil
}

// GetPendingTileSelectionQueue retrieves the pending tile selection queue for a player
func (r *RepositoryImpl) GetPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) (*PendingTileSelectionQueue, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return nil, err
	}

	playerKey := r.getPlayerKey(gameID, playerID)
	tileQueueRepo, exists := r.tileQueueRepos[playerKey]
	if !exists {
		return nil, fmt.Errorf("tile queue repository not found for player %s in game %s", playerID, gameID)
	}

	tileQueue, err := tileQueueRepo.GetQueue(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue: %w", err)
	}

	if tileQueue == nil {
		return nil, nil
	}

	// Convert tiles.PendingTileSelectionQueue to PendingTileSelectionQueue
	return &PendingTileSelectionQueue{
		Items:  append([]string{}, tileQueue.Items...),
		Source: tileQueue.Source,
	}, nil
}

// ClearPendingTileSelectionQueue clears the pending tile selection queue for a player
func (r *RepositoryImpl) ClearPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) error {
	return r.UpdatePendingTileSelectionQueue(ctx, gameID, playerID, nil)
}

// CreateTileQueue creates and stores a tile placement queue for a player
// Note: This is a pure data operation - processing and validation is handled by the service layer
func (r *RepositoryImpl) CreateTileQueue(ctx context.Context, gameID, playerID, cardID string, tilePlacements []string) error {
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
	queue := &PendingTileSelectionQueue{
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
func (r *RepositoryImpl) ProcessNextTileInQueue(ctx context.Context, gameID, playerID string) (string, error) {
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
		updatedQueue := &PendingTileSelectionQueue{
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

// UpdateResourceStorage updates a player's resource storage map
func (r *RepositoryImpl) UpdateResourceStorage(ctx context.Context, gameID, playerID string, resourceStorage map[string]int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Deep copy the map to prevent external mutation
	if resourceStorage == nil {
		player.ResourceStorage = make(map[string]int)
	} else {
		player.ResourceStorage = make(map[string]int)
		for cardID, count := range resourceStorage {
			player.ResourceStorage[cardID] = count
		}
	}

	log.Debug("Player resource storage updated", zap.Int("storage_entries", len(player.ResourceStorage)))

	// Publish ResourceStorageChangedEvent
	events.Publish(r.eventBus, events.ResourceStorageChangedEvent{
		GameID:   gameID,
		PlayerID: playerID,
	})

	return nil
}

// UpdatePaymentSubstitutes updates a player's payment substitutes list
func (r *RepositoryImpl) UpdatePaymentSubstitutes(ctx context.Context, gameID, playerID string, substitutes []PaymentSubstitute) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Deep copy the substitutes slice to prevent external mutation
	if substitutes == nil {
		player.PaymentSubstitutes = []PaymentSubstitute{}
	} else {
		player.PaymentSubstitutes = make([]PaymentSubstitute, len(substitutes))
		copy(player.PaymentSubstitutes, substitutes)
	}

	log.Debug("Player payment substitutes updated", zap.Int("substitutes_count", len(player.PaymentSubstitutes)))

	return nil
}

// UpdatePendingCardSelection updates a player's pending card selection
func (r *RepositoryImpl) UpdatePendingCardSelection(ctx context.Context, gameID, playerID string, selection *PendingCardSelection) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	if selection == nil {
		player.PendingCardSelection = nil
		log.Debug("Cleared pending card selection")
		return nil
	}

	// Deep copy the selection to prevent external mutation
	availableCardsCopy := make([]string, len(selection.AvailableCards))
	copy(availableCardsCopy, selection.AvailableCards)

	cardCostsCopy := make(map[string]int)
	for cardID, cost := range selection.CardCosts {
		cardCostsCopy[cardID] = cost
	}

	cardRewardsCopy := make(map[string]int)
	for cardID, reward := range selection.CardRewards {
		cardRewardsCopy[cardID] = reward
	}

	player.PendingCardSelection = &PendingCardSelection{
		AvailableCards: availableCardsCopy,
		CardCosts:      cardCostsCopy,
		CardRewards:    cardRewardsCopy,
		Source:         selection.Source,
		MinCards:       selection.MinCards,
		MaxCards:       selection.MaxCards,
	}

	log.Debug("🃏 Pending card selection updated",
		zap.String("source", selection.Source),
		zap.Int("available_cards", len(selection.AvailableCards)),
		zap.Int("min_cards", selection.MinCards),
		zap.Int("max_cards", selection.MaxCards))

	return nil
}

// GetPendingCardSelection retrieves a player's pending card selection
func (r *RepositoryImpl) GetPendingCardSelection(ctx context.Context, gameID, playerID string) (*PendingCardSelection, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return nil, err
	}

	if player.PendingCardSelection == nil {
		return nil, nil
	}

	// Return a deep copy to maintain immutability
	availableCardsCopy := make([]string, len(player.PendingCardSelection.AvailableCards))
	copy(availableCardsCopy, player.PendingCardSelection.AvailableCards)

	cardCostsCopy := make(map[string]int)
	for cardID, cost := range player.PendingCardSelection.CardCosts {
		cardCostsCopy[cardID] = cost
	}

	cardRewardsCopy := make(map[string]int)
	for cardID, reward := range player.PendingCardSelection.CardRewards {
		cardRewardsCopy[cardID] = reward
	}

	return &PendingCardSelection{
		AvailableCards: availableCardsCopy,
		CardCosts:      cardCostsCopy,
		CardRewards:    cardRewardsCopy,
		Source:         player.PendingCardSelection.Source,
		MinCards:       player.PendingCardSelection.MinCards,
		MaxCards:       player.PendingCardSelection.MaxCards,
	}, nil
}

// ClearPendingCardSelection clears a player's pending card selection
func (r *RepositoryImpl) ClearPendingCardSelection(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.PendingCardSelection = nil
	log.Debug("🗑️ Pending card selection cleared")

	return nil
}

// UpdatePendingCardDrawSelection updates a player's pending card draw/peek/take/buy selection
func (r *RepositoryImpl) UpdatePendingCardDrawSelection(ctx context.Context, gameID, playerID string, selection *PendingCardDrawSelection) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	if selection == nil {
		player.PendingCardDrawSelection = nil
		log.Debug("Cleared pending card draw selection")
		return nil
	}

	// Deep copy the selection to prevent external mutation
	availableCardsCopy := make([]string, len(selection.AvailableCards))
	copy(availableCardsCopy, selection.AvailableCards)

	player.PendingCardDrawSelection = &PendingCardDrawSelection{
		AvailableCards: availableCardsCopy,
		FreeTakeCount:  selection.FreeTakeCount,
		MaxBuyCount:    selection.MaxBuyCount,
		CardBuyCost:    selection.CardBuyCost,
		Source:         selection.Source,
	}

	log.Debug("🃏 Pending card draw selection updated",
		zap.String("source", selection.Source),
		zap.Int("available_cards", len(selection.AvailableCards)),
		zap.Int("free_take_count", selection.FreeTakeCount),
		zap.Int("max_buy_count", selection.MaxBuyCount),
		zap.Int("card_buy_cost", selection.CardBuyCost))

	return nil
}

// GetPendingCardDrawSelection retrieves a player's pending card draw/peek/take/buy selection
func (r *RepositoryImpl) GetPendingCardDrawSelection(ctx context.Context, gameID, playerID string) (*PendingCardDrawSelection, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return nil, err
	}

	if player.PendingCardDrawSelection == nil {
		return nil, nil
	}

	// Return a deep copy to maintain immutability
	availableCardsCopy := make([]string, len(player.PendingCardDrawSelection.AvailableCards))
	copy(availableCardsCopy, player.PendingCardDrawSelection.AvailableCards)

	return &PendingCardDrawSelection{
		AvailableCards: availableCardsCopy,
		FreeTakeCount:  player.PendingCardDrawSelection.FreeTakeCount,
		MaxBuyCount:    player.PendingCardDrawSelection.MaxBuyCount,
		CardBuyCost:    player.PendingCardDrawSelection.CardBuyCost,
		Source:         player.PendingCardDrawSelection.Source,
	}, nil
}

// ClearPendingCardDrawSelection clears a player's pending card draw/peek selection
func (r *RepositoryImpl) ClearPendingCardDrawSelection(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.PendingCardDrawSelection = nil
	log.Debug("🗑️ Pending card draw selection cleared")

	return nil
}

// UpdateForcedFirstAction sets a player's forced first action (corporation-specific first turn requirement)
func (r *RepositoryImpl) UpdateForcedFirstAction(ctx context.Context, gameID, playerID string, action *ForcedFirstAction) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.ForcedFirstAction = action
	if action != nil {
		log.Info("🎯 Forced first action set",
			zap.String("action_type", action.ActionType),
			zap.String("corporation_id", action.CorporationID),
			zap.String("description", action.Description))
	}

	return nil
}

// GetForcedFirstAction retrieves a player's forced first action
func (r *RepositoryImpl) GetForcedFirstAction(ctx context.Context, gameID, playerID string) (*ForcedFirstAction, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return nil, err
	}

	return player.ForcedFirstAction, nil
}

// MarkForcedFirstActionComplete marks a player's forced first action as completed
func (r *RepositoryImpl) MarkForcedFirstActionComplete(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	if player.ForcedFirstAction != nil {
		player.ForcedFirstAction.Completed = true
		log.Info("✅ Forced first action marked as completed",
			zap.String("action_type", player.ForcedFirstAction.ActionType))
	}

	return nil
}

// ClearForcedFirstAction clears a player's forced first action
func (r *RepositoryImpl) ClearForcedFirstAction(ctx context.Context, gameID, playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	player.ForcedFirstAction = nil
	log.Debug("🗑️ Forced first action cleared")

	return nil
}

// Clear removes all players from the repository
func (r *RepositoryImpl) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.players = make(map[string]map[string]*Player)
	// r.resourcesRepos = make(map[string]resources.Repository)
	// r.productionRepos = make(map[string]production.Repository)
	r.tileQueueRepos = make(map[string]tiles.TileQueueRepository)
	r.playerTurnRepos = make(map[string]turn.PlayerTurnRepository)
}

// ============================================================================
// RESOURCE OPERATION METHODS
// ============================================================================

// DeductResources atomically deducts resources from a player
// Returns error if player cannot afford the cost
func (r *RepositoryImpl) DeductResources(ctx context.Context, gameID, playerID string, cost domain.ResourceSet) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Check affordability
	if !player.Resources.CanAfford(cost) {
		return fmt.Errorf("player cannot afford cost: need %+v, have %+v", cost, player.Resources)
	}

	// Deduct resources
	player.Resources.Subtract(cost)

	log.Info("💰 Resources deducted",
		zap.Int("credits_spent", cost.Credits),
		zap.Int("steel_spent", cost.Steel),
		zap.Int("titanium_spent", cost.Titanium),
		zap.Int("plants_spent", cost.Plants),
		zap.Int("energy_spent", cost.Energy),
		zap.Int("heat_spent", cost.Heat))

	return nil
}

// AddResources adds resources to a player
func (r *RepositoryImpl) AddResources(ctx context.Context, gameID, playerID string, resources domain.ResourceSet) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Add resources
	player.Resources.Add(resources)

	log.Info("💰 Resources added",
		zap.Int("credits_gained", resources.Credits),
		zap.Int("steel_gained", resources.Steel),
		zap.Int("titanium_gained", resources.Titanium),
		zap.Int("plants_gained", resources.Plants),
		zap.Int("energy_gained", resources.Energy),
		zap.Int("heat_gained", resources.Heat))

	return nil
}

// AddProduction adds production to a player
func (r *RepositoryImpl) AddProduction(ctx context.Context, gameID, playerID string, production domain.ResourceSet) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, playerID)

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return err
	}

	// Add production
	player.Production.Add(production)

	log.Info("🏭 Production added",
		zap.Int("credits_production", production.Credits),
		zap.Int("steel_production", production.Steel),
		zap.Int("titanium_production", production.Titanium),
		zap.Int("plants_production", production.Plants),
		zap.Int("energy_production", production.Energy),
		zap.Int("heat_production", production.Heat))

	return nil
}

// CanAfford checks if a player can afford a cost
func (r *RepositoryImpl) CanAfford(ctx context.Context, gameID, playerID string, cost domain.ResourceSet) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	player, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return false, err
	}

	return player.Resources.CanAfford(cost), nil
}

// GetAvailableActions retrieves a player's available action count
func (r *RepositoryImpl) GetAvailableActions(ctx context.Context, gameID, playerID string) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Verify player exists
	_, err := r.getPlayerUnsafe(gameID, playerID)
	if err != nil {
		return 0, err
	}

	// Get from feature repository
	playerKey := r.getPlayerKey(gameID, playerID)
	playerTurnRepo, exists := r.playerTurnRepos[playerKey]
	if !exists {
		return 0, fmt.Errorf("player turn repository not found for player %s in game %s", playerID, gameID)
	}

	actions, err := playerTurnRepo.GetAvailableActions(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get available actions: %w", err)
	}

	return actions, nil
}
